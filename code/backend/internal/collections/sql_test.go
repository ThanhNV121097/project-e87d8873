package collections

import (
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func build(t *testing.T, collection, raw string) (string, []any) {
	t.Helper()
	s, _ := Lookup(collection)
	return BuildList(s, mustParse(t, collection, raw))
}

func wantIn(t *testing.T, sql string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if !strings.Contains(sql, p) {
			t.Errorf("the statement does not contain %q:\n%s", p, sql)
		}
	}
}

func wantNotIn(t *testing.T, sql string, parts ...string) {
	t.Helper()
	for _, p := range parts {
		if strings.Contains(sql, p) {
			t.Errorf("the statement contains %q:\n%s", p, sql)
		}
	}
}

func TestNoConditionMeansNoWhereAndTheSeedOrder(t *testing.T) {
	sql, args := build(t, "products", "")
	wantNotIn(t, sql, "\n  WHERE ")
	wantIn(t, sql,
		"WITH m AS MATERIALIZED (",
		`FROM "products" AS t`,
		`SELECT t."seq", t."id", t."image", t."name"`,
		`t."_source"`,
		"jsonb_strip_nulls(to_jsonb(m) - 'seq') AS rec",
		`row_number() OVER (ORDER BY m."seq") AS rn`,
		"p.rn > $1::bigint AND p.rn <= $2::bigint",
		"COALESCE((", "'[]'::jsonb) AS items", "(SELECT count(*) FROM m) AS total",
	)
	wantNotIn(t, sql, "OFFSET", "LIMIT", "created_at", "updated_at", `t."version"`)
	if want := []any{int64(0), int64(12)}; !reflect.DeepEqual(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
}

func TestThePageIsCutWithTheArithmeticOfPageAndLimit(t *testing.T) {
	for raw, want := range map[string][]any{
		"page=9&limit=12":  {int64(96), int64(108)},
		"page=2&limit=5":   {int64(5), int64(10)},
		"page=1&limit=500": {int64(0), int64(100)},
		"page=3&limit=5":   {int64(10), int64(15)},
	} {
		if _, args := build(t, "products", raw); !reflect.DeepEqual(args, want) {
			t.Errorf("%s: args = %#v, want %#v", raw, args, want)
		}
	}
}

func TestTheWorkedExampleOfTheContract(t *testing.T) {
	sql, args := build(t, "products", "status=live&category=Trail,e-mtb&price_to=70000000&q=Stance&sort=-price&page=2&limit=5")
	wantIn(t, sql,
		`t."category" = ANY (ARRAY(SELECT fold_text(v) FROM unnest($1::text[]) AS v))`,
		`t."price" <= $2::text::numeric`,
		`fold_text(t."status") = ANY (ARRAY(SELECT fold_text(v) FROM unnest($3::text[]) AS v))`,
		`strpos(fold_text(t."id"), fold_text($4::text)) > 0`,
		`strpos(fold_text(t."image_detail"), fold_text($4::text)) > 0`,
		"p.rn > $5::bigint AND p.rn <= $6::bigint",
		`ORDER BY m."price" DESC NULLS LAST, m."seq"`,
	)
	if i, j := strings.Index(sql, `t."category" = ANY`), strings.Index(sql, `t."price" <=`); i < 0 || j < 0 || i > j {
		t.Error("the conditions are not in the alphabetical order of their parameter names")
	}
	want := []any{[]string{"Trail", "e-mtb"}, "70000000", []string{"live"}, "Stance", int64(5), int64(10)}
	if !reflect.DeepEqual(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
	wantIn(t, sql, "\n  WHERE ")
	wantIn(t, sql, " AND ")
}

func TestEachKindOfCondition(t *testing.T) {
	fold := "ARRAY(SELECT fold_text(v) FROM unnest($1::text[]) AS v)"
	cases := []struct {
		collection, raw, want string
		arg0                  any
	}{
		{"products", "id=Talon-2", `t."id" = ANY (` + fold + `)`, []string{"Talon-2"}},
		{"products", "category=trail", `t."category" = ANY (` + fold + `)`, []string{"trail"}},
		{"products", "brand=Giant,Liv", `fold_text(t."brand") = ANY (` + fold + `)`, []string{"Giant", "Liv"}},
		{"products", "stock=0", `t."stock" = ANY ($1::text[]::numeric[])`, []string{"0"}},
		{"products", "rating=4.8", `t."rating" = ANY ($1::text[]::numeric[])`, []string{"4.8"}},
		{"products", "deal=true", `t."deal" = ANY ($1::boolean[])`, []bool{true}},
		{"products", "sizes=M", `EXISTS (SELECT 1 FROM unnest(t."sizes") AS e WHERE fold_text(e) = ANY (` + fold + `))`, []string{"M"}},
		{"products", "price_from=1", `t."price" >= $1::text::numeric`, "1"},
		{"products", "price_to=18990000", `t."price" <= $1::text::numeric`, "18990000"},
		{"reviews", "date_from=2026-09-01", `t."date" COLLATE "C" >= $1::text`, "2026-09-01"},
		{"reviews", "date_to=2026-10-31", `t."date" COLLATE "C" <= $1::text`, "2026-10-31"},
		{"reviews", "bike=stance-0", `t."bike" = ANY (` + fold + `)`, []string{"stance-0"}},
		{"faqs", "q_from=a", `t."q" COLLATE "C" >= $1::text`, "a"},
	}
	for _, c := range cases {
		sql, args := build(t, c.collection, c.raw)
		wantIn(t, sql, c.want)
		if !reflect.DeepEqual(args[0], c.arg0) {
			t.Errorf("%s?%s: first argument = %#v, want %#v", c.collection, c.raw, args[0], c.arg0)
		}
	}
}

func TestSeveralConditionsAreAnded(t *testing.T) {
	sql, _ := build(t, "products", "brand=Giant&wheel=29&category=trail")
	if n := strings.Count(sql, " AND "); n != 2+1 { // two between conditions, one in the page bounds
		t.Errorf("AND appears %d times:\n%s", n, sql)
	}
}

func TestTheOrderAlwaysEndsInSeq(t *testing.T) {
	for raw, want := range map[string]string{
		"sort=price":  `ORDER BY m."price" ASC NULLS LAST, m."seq")`,
		"sort=-price": `ORDER BY m."price" DESC NULLS LAST, m."seq")`,
		"sort=-stock": `ORDER BY m."stock" DESC NULLS LAST, m."seq")`,
		"sort=deal":   `ORDER BY m."deal" ASC NULLS LAST, m."seq")`,
		"sort=name":   `ORDER BY natural_key(m."name") COLLATE "C" ASC NULLS LAST, m."seq")`,
		"sort=-name":  `ORDER BY natural_key(m."name") COLLATE "C" DESC NULLS LAST, m."seq")`,
		"sort=id":     `ORDER BY natural_key(m."id") COLLATE "C" ASC NULLS LAST, m."seq")`,
		"":            `ORDER BY m."seq")`,
	} {
		sql, _ := build(t, "products", raw)
		wantIn(t, sql, want)
	}
}

func TestSearchOnlyTheFieldsNamed(t *testing.T) {
	sql, args := build(t, "products", "q=x&search=name,fact")
	wantIn(t, sql,
		`(strpos(fold_text(t."name"), fold_text($1::text)) > 0 OR strpos(fold_text(t."fact"), fold_text($1::text)) > 0)`)
	wantNotIn(t, sql, `strpos(fold_text(t."brand")`, `strpos(fold_text(t."id")`)
	if want := []any{"x", int64(0), int64(12)}; !reflect.DeepEqual(args, want) {
		t.Errorf("args = %#v, want %#v", args, want)
	}
}

func TestSearchTheDefaultFieldsWithOnePlaceholder(t *testing.T) {
	sql, args := build(t, "products", "q=Stance")
	s, _ := Lookup("products")
	if n := strings.Count(sql, "fold_text($1::text)"); n != len(s.DefaultSearch()) {
		t.Errorf("the words are used %d times, want once per searched field (%d)", n, len(s.DefaultSearch()))
	}
	wantIn(t, sql, `strpos(fold_text(t."id"), fold_text($1::text)) > 0`, `strpos(fold_text(t."name"), fold_text($1::text)) > 0`)
	wantNotIn(t, sql, `fold_text(t."price")`, `fold_text(t."stock")`, `fold_text(t."deal")`, `fold_text(t."sizes")`, `fold_text(t."_source")`, `fold_text(t."_invented")`, "$4")
	if len(args) != 3 || args[0] != "Stance" {
		t.Errorf("args = %#v, want the words once, then the page bounds", args)
	}

	// A blank q adds nothing at all.
	sql, args = build(t, "products", "q=%20%20")
	wantNotIn(t, sql, "strpos")
	if len(args) != 2 {
		t.Errorf("a blank q left args = %#v", args)
	}
}

var hostile = []string{
	"'; drop table products; --",
	"100%",
	"a_b",
	`back\slash`,
	`"quoted"`,
	"semi;colon",
	"Dr. $1 $$ x",
	"Đèo cửa lò",
	strings.Repeat("x", 100),
}

func TestEveryValueIsABoundParameterAndNeverPartOfTheStatement(t *testing.T) {
	s, _ := Lookup("products")
	for _, v := range hostile {
		q, err := ParseQuery(s, url.Values{"q": {v}, "brand": {v}, "price_to": {"1"}})
		if err != nil {
			t.Fatalf("%q refused: %+v", v, *err)
		}
		sql, args := BuildList(s, q)
		if strings.Contains(sql, v) {
			t.Errorf("%q is in the statement text", v)
		}
		// Alphabetical: brand, price_to; then the words.
		if !reflect.DeepEqual(args[0], []string{v}) || args[2] != v {
			t.Errorf("%q: args = %#v, want the value bound as a parameter", v, args)
		}
	}
}

func TestTwoRequestsThatDifferOnlyInValuesHaveTheSameStatement(t *testing.T) {
	a, _ := build(t, "products", "q=alpha&brand=Giant&price_to=5&page=1&limit=12")
	b, _ := build(t, "products", "q=beta&brand=Liv,Giant&price_to=9000&page=7&limit=100")
	if a != b {
		t.Errorf("the statements differ:\n%s\n---\n%s", a, b)
	}
}

func TestEveryCollectionBuildsAndNeverNamesAColumnItDoesNotAnswer(t *testing.T) {
	for _, s := range All {
		sql, _ := BuildList(s, mustParse(t, s.Name, ""))
		wantIn(t, sql, `FROM "`+s.Name+`" AS t`)
		wantNotIn(t, sql, "created_at", "updated_at")
		for _, f := range s.Fields {
			wantIn(t, sql, `t."`+f.Key+`"`)
		}
	}
}

func TestRecordStatements(t *testing.T) {
	p, _ := Lookup("products")
	sql := BuildRecord(p)
	wantIn(t, sql, `jsonb_strip_nulls(to_jsonb(r) - 'version') AS rec, r."version"`, `FROM "products" AS t WHERE t."id" = $1::text`, `t."version"`)
	wantNotIn(t, sql, `t."seq"`, "created_at")

	r, _ := Lookup("reviews")
	sql = BuildRecord(r)
	wantIn(t, sql, `jsonb_strip_nulls(to_jsonb(r)) AS rec`, `FROM "reviews" AS t WHERE t."id" = $1::text`)
	wantNotIn(t, sql, "version", `t."seq"`, "created_at")
}
