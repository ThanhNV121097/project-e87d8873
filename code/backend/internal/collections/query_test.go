package collections

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func parse(t *testing.T, collection, raw string) (Query, *QueryError) {
	t.Helper()
	s, ok := Lookup(collection)
	if !ok {
		t.Fatalf("no collection %q", collection)
	}
	v, err := url.ParseQuery(raw)
	if err != nil {
		t.Fatalf("bad test query %q: %v", raw, err)
	}
	return ParseQuery(s, v)
}

func mustParse(t *testing.T, collection, raw string) Query {
	t.Helper()
	q, err := parse(t, collection, raw)
	if err != nil {
		t.Fatalf("%s?%s: refused: %+v", collection, raw, *err)
	}
	return q
}

// wantRefused checks the refusal names the field and says it in a sentence.
func wantRefused(t *testing.T, collection, raw, field string) {
	t.Helper()
	_, err := parse(t, collection, raw)
	if err == nil {
		t.Errorf("%s?%s: accepted, want a refusal naming %q", collection, raw, field)
		return
	}
	if err.Field != field {
		t.Errorf("%s?%s: field = %q, want %q", collection, raw, err.Field, field)
	}
	if err.Sentence == "" || !strings.HasSuffix(err.Sentence, ".") {
		t.Errorf("%s?%s: sentence = %q, want one sentence", collection, raw, err.Sentence)
	}
	if field != "" && field != "sort" && field != "q" && !strings.Contains(err.Sentence, field) {
		t.Errorf("%s?%s: sentence %q does not name %q", collection, raw, err.Sentence, field)
	}
}

func TestDefaults(t *testing.T) {
	q := mustParse(t, "products", "")
	if q.Page != 1 || q.Limit != 12 || q.Sort != nil || q.Search != nil || q.Words != "" || len(q.Conditions) != 0 {
		t.Errorf("defaults = %+v", q)
	}
}

func TestLimitAbove100IsAnsweredAs100(t *testing.T) {
	for raw, want := range map[string]int64{"limit=100": 100, "limit=101": 100, "limit=500": 100, "limit=99": 99, "limit=1": 1, "limit=007": 7} {
		if got := mustParse(t, "products", raw).Limit; got != want {
			t.Errorf("%s: limit = %d, want %d", raw, got, want)
		}
	}
	if got := mustParse(t, "products", "page=9&limit=5"); got.Page != 9 || got.Limit != 5 {
		t.Errorf("page and limit = %d, %d", got.Page, got.Limit)
	}
}

func TestPageAndLimitMustBeWholeNumbers(t *testing.T) {
	for _, raw := range []string{"page=0", "page=-1", "page=abc", "page=", "page=1.0", "page=+1", "page=%201", "page=1000000001", "page=99999999999999999999"} {
		wantRefused(t, "products", raw, "page")
	}
	for _, raw := range []string{"limit=0", "limit=-1", "limit=abc", "limit=", "limit=1.5", "limit=99999999999999999999"} {
		wantRefused(t, "products", raw, "limit")
	}
	if got := mustParse(t, "products", "page=1000000000").Page; got != 1_000_000_000 {
		t.Errorf("page = %d, want the largest allowed", got)
	}
}

func TestSearchWords(t *testing.T) {
	if got := mustParse(t, "products", "q=%20%20cua%20lo%20").Words; got != "cua lo" {
		t.Errorf("words = %q, want them trimmed", got)
	}
	for _, raw := range []string{"q=", "q=%20%20%20", "q=%C2%A0"} {
		if got := mustParse(t, "products", raw).Words; got != "" {
			t.Errorf("%s: words = %q, want none", raw, got)
		}
	}
	if got := mustParse(t, "products", "q="+strings.Repeat("x", 100)).Words; len(got) != 100 {
		t.Errorf("100 characters were not accepted")
	}
	if got := mustParse(t, "products", "q="+url.QueryEscape(strings.Repeat("đ", 100))).Words; got == "" {
		t.Errorf("100 characters of more than one byte each were not accepted")
	}
	wantRefused(t, "products", "q="+strings.Repeat("x", 101), "q")
	wantRefused(t, "products", "q="+url.QueryEscape(strings.Repeat("đ", 101)), "q")
	wantRefused(t, "products", "q=%FF", "q")
	wantRefused(t, "products", "q=a%00b", "q")
}

func TestQIsAlwaysTheSearchWordsEvenOnFaqs(t *testing.T) {
	q := mustParse(t, "faqs", "q=free")
	if q.Words != "free" || len(q.Conditions) != 0 {
		t.Errorf("faqs?q=free = %+v, want search words and no condition", q)
	}
	q = mustParse(t, "faqs", "q=free&search=q")
	if len(q.Search) != 1 || q.Search[0].Key != "q" {
		t.Errorf("search = %+v, want the field q", q.Search)
	}
	// The field's own range names are not the API's: they are conditions.
	q = mustParse(t, "faqs", "q_from=a")
	if len(q.Conditions) != 1 || q.Conditions[0].Field.Key != "q" || q.Conditions[0].Op != OpFrom {
		t.Errorf("faqs?q_from=a = %+v", q.Conditions)
	}
}

func TestSearchFields(t *testing.T) {
	q := mustParse(t, "products", "q=x&search=name,%20brand,,")
	if len(q.Search) != 2 || q.Search[0].Key != "name" || q.Search[1].Key != "brand" {
		t.Errorf("search = %+v, want name then brand", q.Search)
	}
	if q := mustParse(t, "products", "search=&q=x"); q.Search != nil {
		t.Errorf("an empty search = %+v, want the default", q.Search)
	}
	if q := mustParse(t, "products", "search=id"); len(q.Search) != 1 || q.Search[0].Kind != KindID {
		t.Errorf("search=id = %+v", q.Search)
	}
	wantRefused(t, "faqs", "search=nothing", "nothing")
	for _, name := range []string{"price", "stock", "deal", "sizes", "_source", "seq"} {
		wantRefused(t, "products", "search="+name, name)
	}
	wantRefused(t, "products", "search=name,nothing", "nothing")
}

func TestSort(t *testing.T) {
	q := mustParse(t, "products", "sort=-price")
	if q.Sort == nil || q.Sort.Key != "price" || !q.Desc {
		t.Errorf("sort=-price = %+v", q)
	}
	q = mustParse(t, "products", "sort=%20name%20")
	if q.Sort == nil || q.Sort.Key != "name" || q.Desc {
		t.Errorf("sort=name = %+v", q)
	}
	if q := mustParse(t, "products", "sort="); q.Sort != nil {
		t.Errorf("an empty sort = %+v, want none", q.Sort)
	}
	if q := mustParse(t, "products", "sort=id"); q.Sort == nil || q.Sort.Kind != KindID {
		t.Errorf("sort=id = %+v", q.Sort)
	}
	wantRefused(t, "products", "sort=colourful", "colourful")
	wantRefused(t, "products", "sort=-colourful", "colourful")
	wantRefused(t, "products", "sort=sizes", "sizes")
	wantRefused(t, "products", "sort=_source", "_source")
	wantRefused(t, "products", "sort=-", "sort")
}

func TestEqualityConditions(t *testing.T) {
	q := mustParse(t, "products", "category=trail,e-mtb&brand=Giant&wheel=29")
	if len(q.Conditions) != 3 {
		t.Fatalf("conditions = %+v", q.Conditions)
	}
	// Alphabetical by parameter name: brand, category, wheel.
	for i, key := range []string{"brand", "category", "wheel"} {
		if q.Conditions[i].Field.Key != key || q.Conditions[i].Op != OpEq {
			t.Errorf("condition %d = %+v, want an equality on %s", i, q.Conditions[i], key)
		}
	}
	if want := []string{"trail", "e-mtb"}; !reflect.DeepEqual(q.Conditions[1].Values, want) {
		t.Errorf("category values = %v, want %v", q.Conditions[1].Values, want)
	}
	if got := mustParse(t, "products", "category=trail,,e-mtb,").Conditions[0].Values; !reflect.DeepEqual(got, []string{"trail", "e-mtb"}) {
		t.Errorf("empty parts were kept: %v", got)
	}
}

func TestAnEmptyValueFiltersNothing(t *testing.T) {
	for _, raw := range []string{"category=", "category=,", "price_from=", "price_to=", "stock=", "deal=", "sizes=", "deal=,"} {
		if q := mustParse(t, "products", raw); len(q.Conditions) != 0 {
			t.Errorf("%s: conditions = %+v, want none", raw, q.Conditions)
		}
	}
}

func TestNumbersAndBooleans(t *testing.T) {
	q := mustParse(t, "products", "stock=0&deal=true,false&price_from=1&price_to=70000000&year=2027,2026")
	by := map[string]Condition{}
	for _, c := range q.Conditions {
		by[c.Field.Key+map[Op]string{OpEq: "", OpFrom: "_from", OpTo: "_to"}[c.Op]] = c
	}
	if !reflect.DeepEqual(by["stock"].Values, []string{"0"}) {
		t.Errorf("stock=0 kept as %v: a zero is a value, not an absence", by["stock"].Values)
	}
	if !reflect.DeepEqual(by["deal"].Bools, []bool{true, false}) {
		t.Errorf("deal = %v", by["deal"].Bools)
	}
	if by["price_from"].Value != "1" || by["price_to"].Value != "70000000" {
		t.Errorf("price range = %q..%q", by["price_from"].Value, by["price_to"].Value)
	}
	if !reflect.DeepEqual(by["year"].Values, []string{"2027", "2026"}) {
		t.Errorf("year = %v", by["year"].Values)
	}
	for _, ok := range []string{"stock=-3", "rating=4.8", "price_to=18990000.5", "stock=007"} {
		mustParse(t, "products", ok)
	}
	for _, raw := range []string{"stock=abc", "stock=1e3", "stock=1,x", "stock=.5", "stock=5.", "stock=0x10", "stock=" + strings.Repeat("9", 31)} {
		wantRefused(t, "products", raw, "stock")
	}
	wantRefused(t, "products", "price_from=cheap", "price_from")
	wantRefused(t, "products", "price_to=1,2", "price_to")
	wantRefused(t, "products", "deal=maybe", "deal")
	wantRefused(t, "products", "deal=true,1", "deal")
	wantRefused(t, "products", "deal=TRUE", "deal")
}

func TestRanges(t *testing.T) {
	q := mustParse(t, "reviews", "date_from=2026-09-01&date_to=2026-10-31")
	if len(q.Conditions) != 2 || q.Conditions[0].Op != OpFrom || q.Conditions[1].Op != OpTo || q.Conditions[0].Value != "2026-09-01" {
		t.Errorf("date range = %+v", q.Conditions)
	}
	// A range's value is never split on a comma.
	if got := mustParse(t, "products", "name_from=a,b").Conditions[0].Value; got != "a,b" {
		t.Errorf("range value = %q, want it whole", got)
	}
	wantRefused(t, "products", "deal_from=1", "deal_from")
	wantRefused(t, "products", "month_to=true", "month_to")
	wantRefused(t, "products", "sizes_from=S", "sizes_from")
	wantRefused(t, "products", "sizes_out_to=", "sizes_out_to")
}

func TestUnknownNamesAreRefusedEvenWhenEmpty(t *testing.T) {
	for _, name := range []string{"colourful", "colourful_from", "colourful_to", "_source", "seq", "version", "created_at", "updated_at", "Price", "_from", "_to", "page2", "price_between"} {
		wantRefused(t, "products", name+"=red", name)
		wantRefused(t, "products", name+"=", name)
	}
}

func TestListAndIDConditions(t *testing.T) {
	q := mustParse(t, "products", "sizes=M&id=talon-2,Stance-0&_invented=weight")
	kinds := map[string]Kind{}
	for _, c := range q.Conditions {
		kinds[c.Field.Key] = c.Field.Kind
	}
	if kinds["sizes"] != KindList || kinds["id"] != KindID || kinds["_invented"] != KindList {
		t.Errorf("kinds = %v", kinds)
	}
}

func TestARepeatedParameterIsRefused(t *testing.T) {
	wantRefused(t, "products", "category=a&category=b", "category")
	wantRefused(t, "products", "q=a&q=b", "q")
	wantRefused(t, "products", "page=1&page=2", "page")
	wantRefused(t, "products", "sort=price&sort=name", "sort")
}

func TestTheChecksRunInTheContractsOrder(t *testing.T) {
	cases := []struct{ raw, field string }{
		{"colourful=1&category=a&category=b", "category"},     // repeated first
		{"colourful=1&page=0&limit=0", "page"},                // then page
		{"colourful=1&limit=0&sort=nope", "limit"},            // then limit
		{"colourful=1&sort=nope&search=nope2", "nope"},        // then sort
		{"colourful=1&search=nope2&q=" + strings.Repeat("x", 101), "nope2"}, // then search
		{"colourful=1&q=" + strings.Repeat("x", 101), "q"},    // then q
		{"zzz=1&aaa=1", "aaa"},                                // then the rest, alphabetically
		{"stock=abc&deal=maybe", "deal"},                      // alphabetical: deal before stock
	}
	for _, c := range cases {
		wantRefused(t, "products", c.raw, c.field)
	}
}

func TestConditionsAreInTheAlphabeticalOrderOfTheirNames(t *testing.T) {
	q := mustParse(t, "products", "status=live&price_to=9&brand=Giant&category=trail")
	var got []string
	for _, c := range q.Conditions {
		got = append(got, c.Field.Key)
	}
	if want := []string{"brand", "category", "price", "status"}; !reflect.DeepEqual(got, want) {
		t.Errorf("order = %v, want %v", got, want)
	}
}
