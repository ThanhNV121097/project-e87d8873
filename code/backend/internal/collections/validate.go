package collections

import (
	"bytes"
	"encoding/json"
	"mime"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// The limits of a write body (services.md section 3.11).
const (
	maxTextChars  = 5000
	maxListItems  = 50
	maxNumberText = 40
	maxWhole      = 9007199254740991 // 2^53 - 1: a price or a stock no float loses
)

var (
	idShape    = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)
	wholeShape = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
)

// closedSets are the values schema.json allows for the fields that take one
// of a few: the database's CHECK constraints say the same.
var closedSets = map[string][]string{
	"brand":  {"Giant", "Liv"},
	"status": {"draft", "live"},
	"wheel":  {"29", "mixed"},
}

// wholeNumbers are the number fields that hold a whole amount: a price in
// đồng, a price before a reduction, a count in stock. The others (year,
// rating, review count) are kept exactly as written.
var wholeNumbers = map[string]bool{"price": true, "was": true, "stock": true}

// singular is the noun the sentences use for one record of a collection.
var singular = map[string]string{"products": "bike", "categories": "riding type"}

// Refusal is a write body refused: the status, the code, the sentence a
// person reads and, when one field is at fault, its key.
type Refusal struct {
	Status   int
	Code     string
	Sentence string
	Field    string
}

// Record is a body that passed every rule. Values holds one entry per field of
// the collection in field-table order, ready to bind: nil for "no value", a
// string, an int64 for a whole number, the text of any other number (cast in
// SQL, so it is kept exactly as written), a bool, or a []string (non-nil, even
// when empty).
type Record struct {
	ID       string // "" when the body carries none: the server makes one
	Values   []any
	Category *string // the riding type the body names, when it names one
}

func malformed(field, rest string) *Refusal {
	return &Refusal{400, httpx.CodeMalformedRequest, "The field “" + field + "” " + rest, field}
}

func invalid(field, rest string) *Refusal {
	return &Refusal{400, httpx.CodeValidationFailed, "The field “" + field + "” " + rest, field}
}

func noCategory() *Refusal {
	return invalid("category", "must be the id of a riding type that exists.")
}

type jsonKind int

const (
	jsonNull jsonKind = iota
	jsonString
	jsonNumber
	jsonBool
	jsonArray
	jsonObject
)

// kindOf names the JSON type of one raw value by its first byte; the raw
// values come from json.Unmarshal, so they are well formed.
func kindOf(raw json.RawMessage) jsonKind {
	if len(raw) == 0 {
		return jsonNull
	}
	switch raw[0] {
	case '"':
		return jsonString
	case '{':
		return jsonObject
	case '[':
		return jsonArray
	case 't', 'f':
		return jsonBool
	case 'n':
		return jsonNull
	}
	return jsonNumber
}

// Decode reads a write body in the contract's order and returns the record or
// the first refusal. It needs no database. addressID is the id in the address
// of a change ("" for an add): a body id must equal it.
//
// The order: content type, a JSON object, no undefined key, then every field
// in field-table order with its JSON type before its value (the first failure
// answers), then the one rule that spans two fields: a live bike needs a
// price. The size limit is the handler's, before anything is parsed.
func Decode(s Spec, contentType string, body []byte, addressID string, change bool) (Record, *Refusal) {
	if mt, _, err := mime.ParseMediaType(contentType); err != nil || mt != "application/json" {
		return Record{}, &Refusal{400, httpx.CodeMalformedRequest, "The request must be sent as JSON.", ""}
	}
	var fields map[string]json.RawMessage
	if t := bytes.TrimSpace(body); len(t) == 0 || t[0] != '{' || json.Unmarshal(body, &fields) != nil {
		return Record{}, &Refusal{400, httpx.CodeMalformedRequest, "The request must be a JSON object.", ""}
	}

	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if _, ok := s.Field(k); !ok {
			return Record{}, &Refusal{400, httpx.CodeMalformedRequest, "The field “" + k + "” is not part of a " + singular[s.Name] + ".", k}
		}
	}

	var rec Record
	if raw := fields["id"]; len(raw) > 0 && kindOf(raw) != jsonNull {
		if kindOf(raw) != jsonString {
			return Record{}, malformed("id", "must be text.")
		}
		var id string
		_ = json.Unmarshal(raw, &id)
		if !idShape.MatchString(id) {
			return Record{}, invalid("id", "must be 1 to 64 lowercase letters, digits or hyphens, and not start with a hyphen.")
		}
		if change && id != addressID {
			return Record{}, invalid("id", "must equal the address's id.")
		}
		rec.ID = id
	}

	byKey := make(map[string]any, len(s.Fields))
	rec.Values = make([]any, len(s.Fields))
	for i, f := range s.Fields {
		v, ref := decodeField(f, fields[f.Key])
		if ref != nil {
			return Record{}, ref
		}
		rec.Values[i], byKey[f.Key] = v, v
		if cat, ok := v.(string); ok && f.Key == "category" {
			rec.Category = &cat
		}
	}

	if status, _ := byKey["status"].(string); status == "live" {
		if price, _ := byKey["price"].(int64); price <= 0 {
			return Record{}, invalid("price", "")
		}
	}
	return rec, nil
}

// liveNeedsPrice is the one sentence that is not "The field … ": it is fixed.
const liveNeedsPrice = "A live bike needs a price."

// decodeField checks one field. raw is empty when the key is absent.
func decodeField(f Field, raw json.RawMessage) (any, *Refusal) {
	required := f.Key == "name"
	if kindOf(raw) == jsonNull { // absent and null alike
		if required {
			return nil, invalid(f.Key, "is needed.")
		}
		return nil, nil
	}
	switch f.Kind {
	case KindNumber:
		return decodeNumber(f.Key, raw)
	case KindBool:
		if kindOf(raw) != jsonBool {
			return nil, malformed(f.Key, "must be true or false.")
		}
		return string(raw) == "true", nil
	case KindList:
		return decodeList(f.Key, raw)
	}
	return decodeText(f.Key, raw, required)
}

func decodeText(key string, raw json.RawMessage, required bool) (any, *Refusal) {
	if kindOf(raw) != jsonString {
		return nil, malformed(key, "must be text.")
	}
	var text string
	_ = json.Unmarshal(raw, &text)
	if text == "" && !required {
		return nil, nil // an empty text is no value
	}
	if required && strings.TrimSpace(text) == "" {
		return nil, invalid(key, "is needed.")
	}
	if ref := checkText(key, text, "must be at most 5000 characters."); ref != nil {
		return nil, ref
	}
	if set, ok := closedSets[key]; ok {
		found := false
		for _, v := range set {
			found = found || v == text
		}
		if !found {
			return nil, invalid(key, "must be one of: "+strings.Join(set, ", ")+".")
		}
	}
	return text, nil
}

// checkText refuses a text Postgres cannot store (U+0000) and one over the
// limit, counted in characters (code points).
func checkText(key, text, tooLong string) *Refusal {
	if strings.ContainsRune(text, 0) {
		return &Refusal{400, httpx.CodeValidationFailed, "The text in “" + key + "” contains a character that cannot be stored.", key}
	}
	if utf8.RuneCountInString(text) > maxTextChars {
		return invalid(key, tooLong)
	}
	return nil
}

func decodeNumber(key string, raw json.RawMessage) (any, *Refusal) {
	switch kindOf(raw) {
	case jsonString:
		if string(raw) == `""` {
			return nil, nil // an empty text is no value
		}
		return nil, malformed(key, "must be a number.")
	case jsonNumber:
	default:
		return nil, malformed(key, "must be a number.")
	}
	token := string(raw)
	if !wholeNumbers[key] {
		if _, err := strconv.ParseFloat(token, 64); err != nil || len(token) > maxNumberText {
			return nil, invalid(key, "must be a number.")
		}
		return token, nil
	}
	rest := "must be a whole number of đồng, not below 0."
	if key == "stock" {
		rest = "must be a whole number, not below 0."
	}
	if !wholeShape.MatchString(token) {
		return nil, invalid(key, rest)
	}
	n, err := strconv.ParseInt(token, 10, 64)
	if err != nil || n < 0 || n > maxWhole {
		return nil, invalid(key, rest)
	}
	return n, nil
}

func decodeList(key string, raw json.RawMessage) (any, *Refusal) {
	const notList = "must be a list of text."
	if kindOf(raw) != jsonArray {
		return nil, malformed(key, notList)
	}
	var entries []json.RawMessage
	if json.Unmarshal(raw, &entries) != nil {
		return nil, malformed(key, notList)
	}
	items := make([]string, 0, len(entries))
	for _, e := range entries {
		var item string
		if kindOf(e) != jsonString || json.Unmarshal(e, &item) != nil {
			return nil, malformed(key, notList)
		}
		items = append(items, item)
	}
	if len(items) > maxListItems {
		return nil, invalid(key, "must have at most 50 entries.")
	}
	for _, item := range items {
		if ref := checkText(key, item, "must have entries of at most 5000 characters."); ref != nil {
			return nil, ref
		}
	}
	return items, nil
}
