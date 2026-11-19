// Package orders places an order: it reads the request at the boundary, prices
// it from the shop's own records, refuses a bike that cannot be had, holds the
// bikes and stores the order with its number, all in one transaction.
package orders

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/ThanhNV121097/project-e87d8873/backend/internal/httpx"
)

// MaxBody is the largest request the endpoint reads: 32 KiB.
const MaxBody = 32 << 10

const (
	maxLines   = 20
	maxQty     = 10
	maxName    = 100
	maxPhone   = 30
	maxAddress = 300
	maxNote    = 1000
	maxVariant = 20
	maxID      = 200
)

const (
	deliveryLocal   = "local"
	deliveryCourier = "courier"
	deliveryPickup  = "pickup"
	paymentCOD      = "cod"
)

const (
	sentenceNotJSON   = "The request must be sent as JSON."
	sentenceNotObject = "The request must be a JSON object."
	sentenceTooLong   = "A text is too long: name 100, phone 30, address 300, note 1000, size 20 characters."
)

// Line is one bike of the order: its id, the chosen frame size and the count.
type Line struct {
	ID      string
	Variant string
	Qty     int64
}

// Request is a request that passed every rule of the boundary. Texts are
// trimmed; the money the browser may have sent is not in it.
type Request struct {
	Name     string
	Phone    string
	Delivery string
	Address  string
	Payment  string
	Note     string
	Lines    []Line
}

// Problem is a refusal of the request, ready for the error envelope.
type Problem struct {
	Status   int
	Code     string
	Sentence string
	Field    string
}

func malformed(field, format string, a ...any) *Problem {
	return &Problem{Status: http.StatusBadRequest, Code: httpx.CodeMalformedRequest, Sentence: fmt.Sprintf(format, a...), Field: field}
}

func invalid(field, format string, a ...any) *Problem {
	return &Problem{Status: http.StatusBadRequest, Code: httpx.CodeValidationFailed, Sentence: fmt.Sprintf(format, a...), Field: field}
}

// The keys the contract defines, and the advisory keys it accepts and never
// reads, so a browser that sends its own figures cannot change a price.
var (
	topKeys      = set("name", "phone", "delivery", "address", "payment", "note", "lines", "website")
	topAdvisory  = set("number", "placed", "subtotal", "delivery_fee", "total")
	lineKeys     = set("id", "variant", "qty")
	lineAdvisory = set("name", "price", "image")
	textKeys     = []string{"name", "phone", "delivery", "address", "payment", "note", "website"}
	qtyToken     = regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`)
)

func set(keys ...string) map[string]bool {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return m
}

// Parse reads one request body. It runs the contract's checks in order and the
// first failure answers: the content type and the JSON object, then the bot
// (a filled hidden field, which stops everything else), then the shape (keys
// and JSON types), then the values. The size limit is the caller's, because it
// must be met before the body is read.
func Parse(contentType string, body []byte) (req Request, bot bool, problem *Problem) {
	if media, _, err := mime.ParseMediaType(contentType); err != nil || media != "application/json" {
		return Request{}, false, malformed("", sentenceNotJSON)
	}
	obj, p := decodeObject(body)
	if p != nil {
		return Request{}, false, p
	}
	if isBot(obj) {
		return Request{}, true, nil
	}
	lines, p := checkShape(obj)
	if p != nil {
		return Request{}, false, p
	}
	req, p = checkValues(obj, lines)
	if p != nil {
		return Request{}, false, p
	}
	return req, false, nil
}

// decodeObject accepts exactly one JSON object and nothing after it. The first
// byte is tested because decoding null into a map succeeds silently.
func decodeObject(body []byte) (map[string]json.RawMessage, *Problem) {
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return nil, malformed("", sentenceNotObject)
	}
	dec := json.NewDecoder(bytes.NewReader(trimmed))
	var obj map[string]json.RawMessage
	if err := dec.Decode(&obj); err != nil {
		return nil, malformed("", sentenceNotObject)
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, malformed("", sentenceNotObject)
	}
	return obj, nil
}

// isBot reports a hidden field `website` holding any text at all, a single
// space included.
func isBot(obj map[string]json.RawMessage) bool {
	if !isString(obj["website"]) {
		return false
	}
	var s string
	return json.Unmarshal(obj["website"], &s) == nil && s != ""
}

func first(raw json.RawMessage) byte {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 {
		return 0
	}
	return raw[0]
}

func isString(raw json.RawMessage) bool { return first(raw) == '"' }
func isObject(raw json.RawMessage) bool { return first(raw) == '{' }
func isArray(raw json.RawMessage) bool  { return first(raw) == '[' }
func isNumber(raw json.RawMessage) bool {
	c := first(raw)
	return c == '-' || (c >= '0' && c <= '9')
}

// firstUndefined returns the alphabetically first key that is neither defined
// nor advisory.
func firstUndefined(obj map[string]json.RawMessage, defined, advisory map[string]bool) (string, bool) {
	var keys []string
	for k := range obj {
		if !defined[k] && !advisory[k] {
			keys = append(keys, k)
		}
	}
	if len(keys) == 0 {
		return "", false
	}
	sort.Strings(keys)
	return keys[0], true
}

// checkShape refuses a key the contract does not define and a value of the
// wrong JSON type, and returns the lines as objects. An absent key is never a
// shape error: the value rules catch it. The advisory keys are not looked at.
func checkShape(obj map[string]json.RawMessage) ([]map[string]json.RawMessage, *Problem) {
	if key, ok := firstUndefined(obj, topKeys, topAdvisory); ok {
		return nil, malformed(key, "The field “%s” is not part of an order.", key)
	}
	for _, k := range textKeys {
		if raw, ok := obj[k]; ok && !isString(raw) {
			return nil, malformed(k, "The field “%s” must be text.", k)
		}
	}
	rawLines, ok := obj["lines"]
	if !ok {
		return nil, nil
	}
	var items []json.RawMessage
	if !isArray(rawLines) || json.Unmarshal(rawLines, &items) != nil {
		return nil, malformed("lines", "The field “lines” must be a list.")
	}
	lines := make([]map[string]json.RawMessage, 0, len(items))
	for i, item := range items {
		if !isObject(item) {
			return nil, malformed(fmt.Sprintf("lines[%d]", i), "Line %d must be an object.", i+1)
		}
		var line map[string]json.RawMessage
		if err := json.Unmarshal(item, &line); err != nil {
			return nil, malformed(fmt.Sprintf("lines[%d]", i), "Line %d must be an object.", i+1)
		}
		if key, ok := firstUndefined(line, lineKeys, lineAdvisory); ok {
			return nil, malformed(fmt.Sprintf("lines[%d].%s", i, key), "The field “%s” is not part of an order line.", key)
		}
		for _, k := range []string{"id", "variant"} {
			if raw, ok := line[k]; ok && !isString(raw) {
				return nil, malformed(fmt.Sprintf("lines[%d].%s", i, k), "The field “%s” of line %d must be text.", k, i+1)
			}
		}
		if raw, ok := line["qty"]; ok && !isNumber(raw) {
			return nil, malformed(fmt.Sprintf("lines[%d].qty", i), "The quantity of line %d must be a number.", i+1)
		}
		lines = append(lines, line)
	}
	return lines, nil
}

func text(obj map[string]json.RawMessage, key string) string {
	var s string
	if raw, ok := obj[key]; ok {
		// The shape pass already proved it is a JSON string.
		_ = json.Unmarshal(raw, &s)
	}
	return s
}

func length(s string) int { return utf8.RuneCountInString(s) }

// storable refuses a text Postgres cannot hold (a NUL) or that is over its
// limit, naming the field.
func storable(field, s string, limit int) *Problem {
	if strings.ContainsRune(s, 0) {
		return invalid(field, "The text in “%s” contains a character that cannot be stored.", field)
	}
	if length(s) > limit {
		return invalid(field, sentenceTooLong)
	}
	return nil
}

// checkValues applies the value rules in the contract's order: name, phone,
// delivery, payment, cash on delivery, address, note, lines, then each line's
// id, size and quantity. Every text is trimmed and counted in characters.
func checkValues(obj map[string]json.RawMessage, lines []map[string]json.RawMessage) (Request, *Problem) {
	var r Request

	r.Name = strings.TrimSpace(text(obj, "name"))
	if r.Name == "" {
		return r, invalid("name", "Your name is needed.")
	}
	if p := storable("name", r.Name, maxName); p != nil {
		return r, p
	}

	r.Phone = strings.TrimSpace(text(obj, "phone"))
	if r.Phone == "" {
		return r, invalid("phone", "Your phone number is needed.")
	}
	if p := storable("phone", r.Phone, maxPhone); p != nil {
		return r, p
	}

	r.Delivery = strings.TrimSpace(text(obj, "delivery"))
	if r.Delivery != deliveryLocal && r.Delivery != deliveryCourier && r.Delivery != deliveryPickup {
		return r, invalid("delivery", "The delivery method must be local, courier or pickup.")
	}

	r.Payment = strings.TrimSpace(text(obj, "payment"))
	if r.Payment != "transfer" && r.Payment != paymentCOD && r.Payment != "shop" {
		return r, invalid("payment", "The payment method must be transfer, cod or shop.")
	}
	if r.Payment == paymentCOD && r.Delivery != deliveryLocal {
		return r, invalid("payment", "Cash on delivery is only available for delivery in town.")
	}

	r.Address = strings.TrimSpace(text(obj, "address"))
	if r.Delivery != deliveryPickup && r.Address == "" {
		return r, invalid("address", "A delivery address is needed.")
	}
	if p := storable("address", r.Address, maxAddress); p != nil {
		return r, p
	}

	r.Note = strings.TrimSpace(text(obj, "note"))
	if p := storable("note", r.Note, maxNote); p != nil {
		return r, p
	}

	if len(lines) == 0 {
		return r, invalid("lines", "Your bag is empty.")
	}
	if len(lines) > maxLines {
		return r, invalid("lines", "An order holds at most %d lines.", maxLines)
	}
	r.Lines = make([]Line, 0, len(lines))
	for i, obj := range lines {
		field := func(k string) string { return fmt.Sprintf("lines[%d].%s", i, k) }

		// An id is compared exactly, so it is not trimmed; a blank one names no bike.
		id := text(obj, "id")
		if strings.ContainsRune(id, 0) {
			return r, invalid(field("id"), "The text in “%s” contains a character that cannot be stored.", field("id"))
		}
		if strings.TrimSpace(id) == "" || length(id) > maxID {
			return r, invalid(field("id"), "Line %d needs a bike.", i+1)
		}

		variant := strings.TrimSpace(text(obj, "variant"))
		if p := storable(field("variant"), variant, maxVariant); p != nil {
			return r, p
		}

		qty, ok := quantity(obj["qty"])
		if !ok {
			return r, invalid(field("qty"), "The quantity of line %d must be a whole number from 1 to %d.", i+1, maxQty)
		}
		r.Lines = append(r.Lines, Line{ID: id, Variant: variant, Qty: qty})
	}
	return r, nil
}

// quantity reads a JSON number that is a whole number from 1 to 10. 1.5, 2.0,
// 1e0, 00 and a number too large for 64 bits are all refused, never overflowed.
func quantity(raw json.RawMessage) (int64, bool) {
	token := string(bytes.TrimSpace(raw))
	if !qtyToken.MatchString(token) {
		return 0, false
	}
	n, err := strconv.ParseInt(token, 10, 64)
	if err != nil || n < 1 || n > maxQty {
		return 0, false
	}
	return n, true
}
