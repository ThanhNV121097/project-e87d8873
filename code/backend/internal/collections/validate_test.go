package collections

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func spec(t *testing.T, name string) Spec {
	t.Helper()
	s, ok := Lookup(name)
	if !ok {
		t.Fatalf("no collection %s", name)
	}
	return s
}

func decode(t *testing.T, collection, body string) (Decoded, *Refusal) {
	t.Helper()
	return Decode(spec(t, collection), "application/json", []byte(body), "", false)
}

// valueOf is the bound value of one field in a decoded record.
func valueOf(t *testing.T, s Spec, rec Decoded, key string) any {
	t.Helper()
	for i, f := range s.Fields {
		if f.Key == key {
			return rec.Values[i]
		}
	}
	t.Fatalf("%s has no field %s", s.Name, key)
	return nil
}

func TestABikeIsDecodedIntoTheValuesTheStatementBinds(t *testing.T) {
	body := `{"name":"Talon 2","brand":"Giant","category":"trail","year":2027.0,"price":25000000,"was":"","status":"draft",
	  "stock":3,"sizes":["M","L"],"sizes_out":[],"deal":true,"month":null,"wheel":"29","rating":4.5,"description":"Fast."}`
	rec, ref := decode(t, "products", body)
	if ref != nil {
		t.Fatalf("refused: %+v", ref)
	}
	s := spec(t, "products")
	want := map[string]any{
		"name": "Talon 2", "brand": "Giant", "category": "trail", "year": "2027.0", "price": int64(25000000), "was": nil,
		"status": "draft", "stock": int64(3), "sizes": []string{"M", "L"}, "sizes_out": []string{}, "deal": true, "month": nil,
		"wheel": "29", "rating": "4.5", "description": "Fast.", "fact": nil, "sizes_unknown": nil,
	}
	for key, w := range want {
		if key == "sizes_unknown" {
			continue
		}
		if got := valueOf(t, s, rec, key); !reflect.DeepEqual(got, w) {
			t.Errorf("%s = %#v, want %#v", key, got, w)
		}
	}
	if len(rec.Values) != len(s.Fields) {
		t.Errorf("%d values for %d fields", len(rec.Values), len(s.Fields))
	}
	if rec.ID != "" || rec.Category == nil || *rec.Category != "trail" {
		t.Errorf("id %q, category %v, want no id and trail", rec.ID, rec.Category)
	}
	// An empty list is a real list, not the absence of one: it binds a non-nil slice.
	if got := valueOf(t, s, rec, "sizes_out").([]string); got == nil {
		t.Error("[] was decoded to a nil slice, which would be stored as NULL")
	}
}

func TestEveryEmptyMeansNoValue(t *testing.T) {
	rec, ref := decode(t, "products", `{"name":"x","brand":"","category":null,"price":"","was":null,"description":"","sizes":null,"deal":null,"wheel":""}`)
	if ref != nil {
		t.Fatalf("refused: %+v", ref)
	}
	s := spec(t, "products")
	for _, key := range []string{"brand", "category", "price", "was", "description", "sizes", "deal", "wheel", "image"} {
		if got := valueOf(t, s, rec, key); got != nil {
			t.Errorf("%s = %#v, want no value", key, got)
		}
	}
	if rec.Category != nil {
		t.Errorf("category %v, want none", rec.Category)
	}
}

func TestARidingTypeIsDecoded(t *testing.T) {
	rec, ref := decode(t, "categories", `{"name":"Gravel","line":"","terrain":"Fire roads"}`)
	if ref != nil {
		t.Fatalf("refused: %+v", ref)
	}
	if want := []any{"Gravel", nil, "Fire roads"}; !reflect.DeepEqual(rec.Values, want) {
		t.Errorf("values = %#v, want %#v", rec.Values, want)
	}
}

func TestRefusedBodies(t *testing.T) {
	long := func(n int) string { return strings.Repeat("a", n) }
	cases := []struct {
		name, collection, body string
		status                 int
		code, field, words     string
	}{
		{"an undefined key", "products", `{"name":"x","colourful":"red"}`, 400, "MALFORMED_REQUEST", "colourful", "not part of a bike"},
		{"an undefined key of a riding type", "categories", `{"name":"x","colourful":"red"}`, 400, "MALFORMED_REQUEST", "colourful", "not part of a riding type"},
		{"_source is undefined", "products", `{"name":"x","_source":"y"}`, 400, "MALFORMED_REQUEST", "_source", "not part of"},
		{"seq is undefined", "products", `{"name":"x","seq":1}`, 400, "MALFORMED_REQUEST", "seq", "not part of"},
		{"version is undefined", "categories", `{"name":"x","version":1}`, 400, "MALFORMED_REQUEST", "version", "not part of"},
		{"the first undefined key by name", "products", `{"zebra":1,"apple":2,"name":"x"}`, 400, "MALFORMED_REQUEST", "apple", "not part of"},
		{"no name", "products", `{"brand":"Giant"}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"a blank name", "products", `{"name":"   "}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"a name of a no-break space", "products", `{"name":" "}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"an empty name", "products", `{"name":""}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"a null name", "products", `{"name":null}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"a name that is a number", "products", `{"name":5}`, 400, "MALFORMED_REQUEST", "name", "must be text"},
		{"a riding type with no name", "categories", `{"line":"x"}`, 400, "VALIDATION_FAILED", "name", "needed"},
		{"price cheap", "products", `{"name":"x","price":"cheap"}`, 400, "MALFORMED_REQUEST", "price", "must be a number"},
		{"price -1", "products", `{"name":"x","price":-1}`, 400, "VALIDATION_FAILED", "price", "whole number"},
		{"price 1.5", "products", `{"name":"x","price":1.5}`, 400, "VALIDATION_FAILED", "price", "whole number"},
		{"price 5.0", "products", `{"name":"x","price":5.0}`, 400, "VALIDATION_FAILED", "price", "whole number"},
		{"price 1e3", "products", `{"name":"x","price":1e3}`, 400, "VALIDATION_FAILED", "price", "whole number"},
		{"price above 2^53-1", "products", `{"name":"x","price":9007199254740992}`, 400, "VALIDATION_FAILED", "price", "whole number"},
		{"was -1", "products", `{"name":"x","was":-1}`, 400, "VALIDATION_FAILED", "was", "whole number"},
		{"stock -1", "products", `{"name":"x","stock":-1}`, 400, "VALIDATION_FAILED", "stock", "whole number"},
		{"stock 2.5", "products", `{"name":"x","stock":2.5}`, 400, "VALIDATION_FAILED", "stock", "whole number"},
		{"stock text", "products", `{"name":"x","stock":"many"}`, 400, "MALFORMED_REQUEST", "stock", "must be a number"},
		{"year text", "products", `{"name":"x","year":"soon"}`, 400, "MALFORMED_REQUEST", "year", "must be a number"},
		{"a year token of 41 characters", "products", `{"name":"x","year":1` + strings.Repeat("0", 40) + `}`, 400, "VALIDATION_FAILED", "year", "must be a number"},
		{"a rating out of range", "products", `{"name":"x","rating":1e999}`, 400, "VALIDATION_FAILED", "rating", "must be a number"},
		{"sizes not a list", "products", `{"name":"x","sizes":"M"}`, 400, "MALFORMED_REQUEST", "sizes", "list of text"},
		{"sizes with a number", "products", `{"name":"x","sizes":["M",5]}`, 400, "MALFORMED_REQUEST", "sizes", "list of text"},
		{"51 sizes", "products", `{"name":"x","sizes":[` + strings.TrimSuffix(strings.Repeat(`"a",`, 51), ",") + `]}`, 400, "VALIDATION_FAILED", "sizes", "at most 50"},
		{"a size of 5001 characters", "products", `{"name":"x","sizes":["` + long(5001) + `"]}`, 400, "VALIDATION_FAILED", "sizes", "at most 5000"},
		{"deal yes", "products", `{"name":"x","deal":"yes"}`, 400, "MALFORMED_REQUEST", "deal", "true or false"},
		{"month 1", "products", `{"name":"x","month":1}`, 400, "MALFORMED_REQUEST", "month", "true or false"},
		{"status sold", "products", `{"name":"x","status":"sold"}`, 400, "VALIDATION_FAILED", "status", "draft, live"},
		{"wheel 26", "products", `{"name":"x","wheel":"26"}`, 400, "VALIDATION_FAILED", "wheel", "29, mixed"},
		{"wheel the number 29", "products", `{"name":"x","wheel":29}`, 400, "MALFORMED_REQUEST", "wheel", "must be text"},
		{"brand Trek", "products", `{"name":"x","brand":"Trek"}`, 400, "VALIDATION_FAILED", "brand", "Giant, Liv"},
		{"a live bike with no price", "products", `{"name":"x","status":"live"}`, 400, "VALIDATION_FAILED", "price", "A live bike needs a price."},
		{"a live bike with price 0", "products", `{"name":"x","status":"live","price":0}`, 400, "VALIDATION_FAILED", "price", "A live bike needs a price."},
		{"a live bike with an empty price", "products", `{"name":"x","status":"live","price":""}`, 400, "VALIDATION_FAILED", "price", "A live bike needs a price."},
		{"an id with capitals", "products", `{"id":"Bad","name":"x"}`, 400, "VALIDATION_FAILED", "id", "lowercase"},
		{"an id with a space", "products", `{"id":"my bike","name":"x"}`, 400, "VALIDATION_FAILED", "id", "lowercase"},
		{"an id starting with a hyphen", "products", `{"id":"-bike","name":"x"}`, 400, "VALIDATION_FAILED", "id", "lowercase"},
		{"an empty id", "products", `{"id":"","name":"x"}`, 400, "VALIDATION_FAILED", "id", "lowercase"},
		{"an id of 65 characters", "products", `{"id":"` + long(65) + `","name":"x"}`, 400, "VALIDATION_FAILED", "id", "lowercase"},
		{"an id that is a number", "products", `{"id":7,"name":"x"}`, 400, "MALFORMED_REQUEST", "id", "must be text"},
		{"text of 5001 characters", "products", `{"name":"x","description":"` + long(5001) + `"}`, 400, "VALIDATION_FAILED", "description", "at most 5000"},
		{"a name of 5001 characters", "categories", `{"name":"` + long(5001) + `"}`, 400, "VALIDATION_FAILED", "name", "at most 5000"},
		{"a NUL in a text", "products", `{"name":"x","fact":"a\u0000b"}`, 400, "VALIDATION_FAILED", "fact", "cannot be stored"},
		{"a NUL in a list entry", "products", `{"name":"x","sizes":["a\u0000"]}`, 400, "VALIDATION_FAILED", "sizes", "cannot be stored"},
		{"an array body", "products", `[]`, 400, "MALFORMED_REQUEST", "", "JSON object"},
		{"a null body", "products", `null`, 400, "MALFORMED_REQUEST", "", "JSON object"},
		{"an empty body", "products", ``, 400, "MALFORMED_REQUEST", "", "JSON object"},
		{"trailing data", "products", `{"name":"x"} {"name":"y"}`, 400, "MALFORMED_REQUEST", "", "JSON object"},
		// Two rules broken at once: the field table's order decides which answers.
		{"id before name", "products", `{"id":"Bad Id","name":""}`, 400, "VALIDATION_FAILED", "id", ""},
		{"brand before price", "products", `{"name":"x","brand":"Trek","price":-1}`, 400, "VALIDATION_FAILED", "brand", ""},
		{"stock before the live rule", "products", `{"name":"x","status":"live","stock":-1}`, 400, "VALIDATION_FAILED", "stock", ""},
		{"a bad price before a missing riding type", "products", `{"name":"x","category":"unicycle","price":-1}`, 400, "VALIDATION_FAILED", "price", ""},
	}
	for _, c := range cases {
		_, ref := decode(t, c.collection, c.body)
		if ref == nil {
			t.Errorf("%s: accepted, want %d %s", c.name, c.status, c.code)
			continue
		}
		if ref.Status != c.status || ref.Code != c.code || ref.Field != c.field || !strings.Contains(ref.Sentence, c.words) {
			t.Errorf("%s: %d %s field %q %q, want %d %s field %q containing %q", c.name, ref.Status, ref.Code, ref.Field, ref.Sentence, c.status, c.code, c.field, c.words)
		}
	}
}

func TestAcceptedBodiesAtTheEdges(t *testing.T) {
	long := func(n int, r string) string { return strings.Repeat(r, n) }
	for name, body := range map[string]string{
		"an id of 64 characters":          `{"id":"` + long(64, "a") + `","name":"x"}`,
		"a client id":                     `{"id":"my-bike-1","name":"x"}`,
		"text of 5000 characters":         `{"name":"x","description":"` + long(5000, "a") + `"}`,
		"5000 multi-byte characters":      `{"name":"x","description":"` + long(5000, "Đ") + `"}`,
		"a price of 9007199254740991":     `{"name":"x","price":9007199254740991}`,
		"a price of -0":                   `{"name":"x","price":-0}`,
		"stock 0":                         `{"name":"x","stock":0}`,
		"a year token of 40 characters":   `{"name":"x","year":1` + long(39, "0") + `}`,
		"a draft with no price":           `{"name":"x","status":"draft"}`,
		"no status and no price":          `{"name":"x"}`,
		"a live bike with a price":        `{"name":"x","status":"live","price":1}`,
		"a duplicate key keeps the last":  `{"name":"","name":"x"}`,
		"a riding type that is not found": `{"name":"x","category":"unicycle"}`,
		"the hidden fields as read":       `{"name":"x","price_listed":"1.000.000 ₫","_invented":["rating"]}`,
	} {
		if _, ref := decode(t, "products", body); ref != nil {
			t.Errorf("%s: refused: %+v", name, ref)
		}
	}
	if rec, _ := decode(t, "products", `{"name":"x","price":-0}`); valueOf(t, spec(t, "products"), rec, "price") != int64(0) {
		t.Error("-0 is not 0")
	}
}

func TestAChangeMustNameItsOwnIDOrNone(t *testing.T) {
	s := spec(t, "products")
	if _, ref := Decode(s, "application/json", []byte(`{"id":"other","name":"x"}`), "talon-2", true); ref == nil || ref.Field != "id" || ref.Code != "VALIDATION_FAILED" {
		t.Errorf("a different id: %+v, want a 400 on id", ref)
	}
	for _, body := range []string{`{"id":"talon-2","name":"x"}`, `{"name":"x"}`, `{"id":null,"name":"x"}`} {
		if _, ref := Decode(s, "application/json", []byte(body), "talon-2", true); ref != nil {
			t.Errorf("%s: refused: %+v", body, ref)
		}
	}
}

func TestOnlyJSONIsRead(t *testing.T) {
	s := spec(t, "products")
	for _, ct := range []string{"", "text/plain", "application/x-www-form-urlencoded", "application/jsonx", "garbage;;"} {
		if _, ref := Decode(s, ct, []byte(`{"name":"x"}`), "", false); ref == nil || ref.Code != "MALFORMED_REQUEST" || !strings.Contains(ref.Sentence, "JSON") {
			t.Errorf("content type %q: %+v, want a 400 asking for JSON", ct, ref)
		}
	}
	for _, ct := range []string{"application/json", "Application/JSON", "application/json; charset=utf-8"} {
		if _, ref := Decode(s, ct, []byte(`{"name":"x"}`), "", false); ref != nil {
			t.Errorf("content type %q: refused: %+v", ct, ref)
		}
	}
}

// The closed sets are the schema file's own options: a drift between them
// would let the validator accept what the database refuses, or the reverse.
func TestEveryClosedSetMatchesTheDatabaseChecks(t *testing.T) {
	want := map[string][]string{"brand": {"Giant", "Liv"}, "status": {"draft", "live"}, "wheel": {"29", "mixed"}}
	if !reflect.DeepEqual(closedSets, want) {
		t.Errorf("closedSets = %v, want %v", closedSets, want)
	}
	s := spec(t, "products")
	for key := range closedSets {
		if f, ok := s.Field(key); !ok || f.Kind != KindText {
			t.Errorf("%s is not a text field of products", key)
		}
	}
	for key := range wholeNumbers {
		if f, ok := s.Field(key); !ok || f.Kind != KindNumber {
			t.Errorf("%s is not a number field of products", key)
		}
	}
}

func TestEveryFieldOfAWritableCollectionIsDecoded(t *testing.T) {
	// A field added to the table with no decoding rule would be silently
	// bound as text: every kind the writable collections hold has a rule.
	for _, s := range All {
		if !s.Writable() {
			continue
		}
		for _, f := range s.Fields {
			switch f.Kind {
			case KindText, KindRef, KindNumber, KindBool, KindList:
			default:
				t.Errorf("%s.%s has a kind with no write rule: %d", s.Name, f.Key, f.Kind)
			}
		}
	}
	if fmt.Sprint(spec(t, "products").Writable(), spec(t, "categories").Writable(), spec(t, "reviews").Writable()) != "true true false" {
		t.Error("only products and categories are writable")
	}
}
