package orders

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

const jsonType = "application/json"

func validOrder() map[string]any {
	return map[string]any{
		"name": "Minh", "phone": "0901234567", "delivery": "local", "address": "12 Hùng Vương", "payment": "transfer",
		"lines": []any{map[string]any{"id": "talon-2", "variant": "M", "qty": 2}},
	}
}

// parseWith parses the valid order after a case has changed one thing in it.
func parseWith(t *testing.T, change func(m map[string]any)) (Request, bool, *Problem) {
	t.Helper()
	m := validOrder()
	if change != nil {
		change(m)
	}
	body, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	return Parse(jsonType, body)
}

// withQty is the valid order as text with the quantity written exactly as given.
func withQty(token string) string {
	return `{"name":"Minh","phone":"0901234567","delivery":"local","address":"12 Hùng Vương","payment":"transfer","lines":[{"id":"talon-2","variant":"M","qty":` + token + `}]}`
}

func with(key string, value any) func(map[string]any) {
	return func(m map[string]any) { m[key] = value }
}

func drop(key string) func(map[string]any) {
	return func(m map[string]any) { delete(m, key) }
}

func manyLines(n int) []any {
	sizes := []string{"XS", "S", "M", "L", "XL"}
	lines := make([]any, n)
	for i := range lines {
		lines[i] = map[string]any{"id": "talon-2", "variant": sizes[i%len(sizes)], "qty": 1}
	}
	return lines
}

func wantProblem(t *testing.T, label string, p *Problem, code, field, words string) {
	t.Helper()
	if p == nil {
		t.Errorf("%s: accepted, want %s", label, code)
		return
	}
	if p.Status != 400 || p.Code != code || p.Field != field {
		t.Errorf("%s: status %d code %s field %q, want 400 %s %q", label, p.Status, p.Code, p.Field, code, field)
	}
	if !strings.Contains(strings.ToLower(p.Sentence), strings.ToLower(words)) {
		t.Errorf("%s: sentence %q does not name %q", label, p.Sentence, words)
	}
}

func TestAValidOrderIsParsedAsItWasSent(t *testing.T) {
	req, bot, p := parseWith(t, nil)
	if p != nil || bot {
		t.Fatalf("problem %v, bot %v", p, bot)
	}
	want := Request{Name: "Minh", Phone: "0901234567", Delivery: "local", Address: "12 Hùng Vương", Payment: "transfer",
		Lines: []Line{{ID: "talon-2", Variant: "M", Qty: 2}}}
	if !reflect.DeepEqual(req, want) {
		t.Errorf("request = %+v, want %+v", req, want)
	}
}

func TestEveryTextIsTrimmedExceptABikesId(t *testing.T) {
	req, _, p := parseWith(t, func(m map[string]any) {
		m["name"], m["phone"], m["address"], m["note"] = " Minh ", "\t090 ", "  12 Hùng Vương ", "  call me "
		m["lines"] = []any{map[string]any{"id": " talon-2", "variant": " M ", "qty": 1}}
	})
	if p != nil {
		t.Fatal(p)
	}
	if req.Name != "Minh" || req.Phone != "090" || req.Address != "12 Hùng Vương" || req.Note != "call me" || req.Lines[0].Variant != "M" {
		t.Errorf("texts were not trimmed: %+v", req)
	}
	if req.Lines[0].ID != " talon-2" {
		t.Errorf("id = %q: an id is compared exactly and is never trimmed", req.Lines[0].ID)
	}
}

func TestTheValueRulesRefuseWithTheFieldAndASentenceNamingIt(t *testing.T) {
	cod := func(delivery string) func(map[string]any) {
		return func(m map[string]any) { m["payment"], m["delivery"] = "cod", delivery }
	}
	cases := []struct {
		name   string
		change func(map[string]any)
		field  string
		words  string
	}{
		{"no name", drop("name"), "name", "name"},
		{"blank name", with("name", "   "), "name", "name"},
		{"no phone", drop("phone"), "phone", "phone"},
		{"blank phone", with("phone", " "), "phone", "phone"},
		{"no delivery", drop("delivery"), "delivery", "delivery"},
		{"a drone", with("delivery", "drone"), "delivery", "delivery"},
		{"no payment", drop("payment"), "payment", "payment"},
		{"a card", with("payment", "card"), "payment", "payment"},
		{"cash on delivery by courier", cod("courier"), "payment", "cash on delivery"},
		{"cash on delivery to collect", cod("pickup"), "payment", "cash on delivery"},
		{"no address for local", drop("address"), "address", "address"},
		{"blank address for courier", func(m map[string]any) { m["delivery"], m["address"] = "courier", "  " }, "address", "address"},
		{"no lines", drop("lines"), "lines", "bag"},
		{"empty lines", with("lines", []any{}), "lines", "bag"},
		{"21 lines", with("lines", manyLines(21)), "lines", "20 lines"},
		{"no bike id", with("lines", []any{map[string]any{"qty": 1}}), "lines[0].id", "line 1"},
		{"blank bike id", with("lines", []any{map[string]any{"id": "  ", "qty": 1}}), "lines[0].id", "line 1"},
		{"a bike id of 201 characters", with("lines", []any{map[string]any{"id": strings.Repeat("a", 201), "qty": 1}}), "lines[0].id", "line 1"},
		{"no quantity on line 2", with("lines", []any{map[string]any{"id": "a", "qty": 1}, map[string]any{"id": "b"}}), "lines[1].qty", "quantity of line 2"},
		{"a NUL in the name", with("name", "a\x00b"), "name", "cannot be stored"},
		{"a NUL in the note", with("note", "a\x00b"), "note", "cannot be stored"},
		{"a NUL in a size", with("lines", []any{map[string]any{"id": "a", "variant": "M\x00", "qty": 1}}), "lines[0].variant", "cannot be stored"},
	}
	for _, c := range cases {
		_, bot, p := parseWith(t, c.change)
		if bot {
			t.Errorf("%s: taken for a bot", c.name)
		}
		wantProblem(t, c.name, p, "VALIDATION_FAILED", c.field, c.words)
	}
}

func TestQuantityIsAWholeNumberFromOneToTen(t *testing.T) {
	for _, token := range []string{"1", "10", "5"} {
		if _, _, p := Parse(jsonType, []byte(withQty(token))); p != nil {
			t.Errorf("qty %s refused: %v", token, p)
		}
	}
	for _, token := range []string{"0", "11", "-1", "1.5", "2.0", "1e0", "-0", "1234567890123456789012345"} {
		_, _, p := Parse(jsonType, []byte(withQty(token)))
		wantProblem(t, "qty "+token, p, "VALIDATION_FAILED", "lines[0].qty", "quantity")
	}
	// 00 is not a JSON number at all, so the body is refused whole as malformed.
	_, _, p := Parse(jsonType, []byte(withQty("00")))
	wantProblem(t, "qty 00", p, "MALFORMED_REQUEST", "", "JSON object")
	for _, token := range []string{`"2"`, "null", "true", "[]", "{}"} {
		_, _, p := Parse(jsonType, []byte(withQty(token)))
		wantProblem(t, "qty "+token, p, "MALFORMED_REQUEST", "lines[0].qty", "quantity")
	}
}

func TestTextLimitsHoldAtTheLimitAndRefuseOnePast(t *testing.T) {
	line := func(variant string) func(map[string]any) {
		return func(m map[string]any) {
			m["lines"] = []any{map[string]any{"id": "talon-2", "variant": variant, "qty": 1}}
		}
	}
	cases := []struct {
		name  string
		limit int
		make  func(text string) func(map[string]any)
		field string
	}{
		{"name", 100, func(s string) func(map[string]any) { return with("name", s) }, "name"},
		{"phone", 30, func(s string) func(map[string]any) { return with("phone", s) }, "phone"},
		{"address", 300, func(s string) func(map[string]any) { return with("address", s) }, "address"},
		{"note", 1000, func(s string) func(map[string]any) { return with("note", s) }, "note"},
		{"size", 20, line, "lines[0].variant"},
	}
	for _, c := range cases {
		if _, _, p := parseWith(t, c.make(strings.Repeat("a", c.limit))); p != nil {
			t.Errorf("%s at %d characters refused: %v", c.name, c.limit, p)
		}
		_, _, p := parseWith(t, c.make(strings.Repeat("a", c.limit+1)))
		wantProblem(t, c.name+" one past", p, "VALIDATION_FAILED", c.field, "too long")
	}
	// Characters are counted, not bytes: 100 letters of two bytes each fit.
	if _, _, p := parseWith(t, with("name", strings.Repeat("Đ", 100))); p != nil {
		t.Errorf("a name of 100 Đ refused: %v", p)
	}
	// The limit counts the text after its spaces are trimmed.
	if _, _, p := parseWith(t, with("name", "  "+strings.Repeat("a", 100)+"  ")); p != nil {
		t.Errorf("a name of 100 characters inside spaces refused: %v", p)
	}
}

func TestAnOrderHoldsOneToTwentyLines(t *testing.T) {
	for _, n := range []int{1, 20} {
		if req, _, p := parseWith(t, with("lines", manyLines(n))); p != nil || len(req.Lines) != n {
			t.Errorf("%d lines: problem %v, got %d", n, p, len(req.Lines))
		}
	}
}

func TestAPickupNeedsNoAddressAndKeepsOneItIsGiven(t *testing.T) {
	if _, _, p := parseWith(t, func(m map[string]any) { m["delivery"] = "pickup"; delete(m, "address") }); p != nil {
		t.Errorf("a pickup with no address refused: %v", p)
	}
	req, _, p := parseWith(t, func(m map[string]any) { m["delivery"], m["address"] = "pickup", " the shop " })
	if p != nil || req.Address != "the shop" {
		t.Errorf("a pickup with an address: %+v, %v", req, p)
	}
	_, _, p = parseWith(t, func(m map[string]any) { m["delivery"], m["address"] = "pickup", strings.Repeat("a", 301) })
	wantProblem(t, "a pickup address one past", p, "VALIDATION_FAILED", "address", "too long")
	if _, _, p := parseWith(t, func(m map[string]any) { m["payment"] = "shop"; m["delivery"] = "pickup" }); p != nil {
		t.Errorf("pickup paid at the shop refused: %v", p)
	}
	if _, _, p := parseWith(t, func(m map[string]any) { m["payment"] = "cod" }); p != nil {
		t.Errorf("cash on delivery in town refused: %v", p)
	}
}

func TestAKeyTheContractDoesNotDefineIsRefusedAtEveryLevel(t *testing.T) {
	_, _, p := parseWith(t, with("colour", "red"))
	wantProblem(t, "top level", p, "MALFORMED_REQUEST", "colour", "colour")

	_, _, p = parseWith(t, func(m map[string]any) { m["zeta"], m["alpha"] = 1, 2 })
	wantProblem(t, "two undefined keys", p, "MALFORMED_REQUEST", "alpha", "alpha")

	_, _, p = parseWith(t, with("lines", []any{map[string]any{"id": "a", "qty": 1, "colour": "red"}}))
	wantProblem(t, "a line", p, "MALFORMED_REQUEST", "lines[0].colour", "colour")
}

func TestTheBrowsersOwnMoneyIsAcceptedAndNeverRead(t *testing.T) {
	body := `{"name":"Minh","phone":"0901234567","delivery":"local","address":"x","payment":"transfer",
		"total":"not a number","subtotal":1,"delivery_fee":null,"number":"X","placed":[],
		"lines":[{"id":"talon-2","variant":"M","qty":1,"price":null,"name":"Cheap","image":7}]}`
	req, bot, p := Parse(jsonType, []byte(body))
	if p != nil || bot {
		t.Fatalf("problem %v, bot %v", p, bot)
	}
	if !reflect.DeepEqual(req.Lines, []Line{{ID: "talon-2", Variant: "M", Qty: 1}}) {
		t.Errorf("lines = %+v: the advisory keys must not reach the request", req.Lines)
	}
}

func TestAWrongJSONTypeIsMalformedAndAnAbsentKeyIsNot(t *testing.T) {
	cases := []struct {
		name   string
		change func(map[string]any)
		field  string
	}{
		{"a number for a name", with("name", 5), "name"},
		{"null for a phone", with("phone", nil), "phone"},
		{"a number for a note", with("note", 5), "note"},
		{"a number for the hidden field", with("website", 5), "website"},
		{"a string for the lines", with("lines", "x"), "lines"},
		{"null for the lines", with("lines", nil), "lines"},
		{"a line that is not an object", with("lines", []any{"talon-2"}), "lines[0]"},
		{"a number for a bike id", with("lines", []any{map[string]any{"id": 5, "qty": 1}}), "lines[0].id"},
		{"a number for a size", with("lines", []any{map[string]any{"id": "a", "variant": 5, "qty": 1}}), "lines[0].variant"},
	}
	for _, c := range cases {
		_, _, p := parseWith(t, c.change)
		wantProblem(t, c.name, p, "MALFORMED_REQUEST", c.field, "")
	}
}

func TestAShapeErrorComesBeforeAValueError(t *testing.T) {
	body := `{"name":"","phone":"1","delivery":"local","address":"x","payment":"transfer","lines":[{"id":"a","qty":"2"}]}`
	_, _, p := Parse(jsonType, []byte(body))
	wantProblem(t, "qty as text with a blank name", p, "MALFORMED_REQUEST", "lines[0].qty", "quantity")
}

func TestAFilledHiddenFieldIsABotWhateverElseIsWrong(t *testing.T) {
	for _, body := range []string{
		`{"website":"x"}`,
		`{"website":" "}`,
		`{"website":"x","lines":5,"zzz":1,"name":7}`,
	} {
		if _, bot, p := Parse(jsonType, []byte(body)); !bot || p != nil {
			t.Errorf("%s: bot %v, problem %v, want a bot", body, bot, p)
		}
	}
	for _, change := range []func(map[string]any){with("website", ""), nil} {
		if _, bot, p := parseWith(t, change); bot || p != nil {
			t.Errorf("an empty or absent hidden field: bot %v, problem %v", bot, p)
		}
	}
}

func TestABodyThatIsNotOneJSONObjectIsMalformed(t *testing.T) {
	for _, body := range []string{"[]", `"x"`, "null", "5", "", "  ", "{", "{}{}", `{"a":1} x`, `{"name":}`} {
		_, bot, p := Parse(jsonType, []byte(body))
		if bot {
			t.Errorf("%q: taken for a bot", body)
		}
		wantProblem(t, "body "+body, p, "MALFORMED_REQUEST", "", "JSON object")
	}
	if _, _, p := Parse(jsonType, []byte(withQty("1")+" \n")); p != nil {
		t.Errorf("white space after the object refused: %v", p)
	}
}

func TestOnlyJSONIsAccepted(t *testing.T) {
	body := []byte(withQty("1"))
	for _, ct := range []string{"application/json", "application/json; charset=utf-8", "Application/JSON"} {
		if _, _, p := Parse(ct, body); p != nil {
			t.Errorf("%q refused: %v", ct, p)
		}
	}
	for _, ct := range []string{"", "text/plain", "application/x-www-form-urlencoded", "application/jsonp", "json"} {
		_, _, p := Parse(ct, body)
		wantProblem(t, "content type "+ct, p, "MALFORMED_REQUEST", "", "JSON")
	}
}
