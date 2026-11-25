package migrations

import (
	"regexp"
	"strings"
	"testing"
)

// name is {UTC timestamp, 14 digits}_{snake_case}.{up|down}.sql. The runner
// applies files in filename order, so the timestamp is what orders them.
var name = regexp.MustCompile(`^\d{14}_[a-z0-9_]+\.(up|down)\.sql$`)

func TestEveryMigrationIsNamedAndPaired(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	ups, downs := map[string]bool{}, map[string]bool{}
	for _, e := range entries {
		m := name.FindStringSubmatch(e.Name())
		if m == nil {
			t.Errorf("%s: not named {YYYYMMDDHHMMSS}_{name}.up.sql or .down.sql", e.Name())
			continue
		}
		base := strings.TrimSuffix(e.Name(), "."+m[1]+".sql")
		if m[1] == "up" {
			ups[base] = true
		} else {
			downs[base] = true
		}
	}
	if len(ups) == 0 {
		t.Error("no migrations: the baseline must stay")
	}
	for base := range ups {
		if !downs[base] {
			t.Errorf("%s has no .down.sql", base)
		}
	}
	for base := range downs {
		if !ups[base] {
			t.Errorf("%s has a .down.sql and no .up.sql", base)
		}
	}
}

func read(t *testing.T, name string) string {
	t.Helper()
	body, err := FS.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(body)
}

// The runner decides whether a file may run inside a transaction by looking
// for the word CONCURRENTLY anywhere in a statement's text, strings included.
// A seeded text that held it would run the whole file outside a transaction,
// and a failure half-way would leave half the shop stored. Better this test
// fails in CI than that happens in production.
func TestNoFileThatCreatesTablesOrHoldsDataMentionsTheConcurrentWord(t *testing.T) {
	entries, err := FS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		body := strings.ToLower(read(t, e.Name()))
		holdsData := strings.Contains(body, "create table") || strings.Contains(body, "insert into")
		if holdsData && strings.Contains(body, "concurrent"+"ly") {
			t.Errorf("%s creates tables or inserts data and mentions the word that moves a file outside its transaction", e.Name())
		}
	}
}

const (
	seedUp   = "20261118000200_seed_collections.up.sql"
	seedDown = "20261118000200_seed_collections.down.sql"
)

var seedRecord = regexp.MustCompile(`\{"id":"([a-z0-9-]+)"`)

// The seed copies every record of the eight JSON files, in this order of
// tables (a reference must exist before it is made) and in the file's order.
func TestTheSeedHoldsEveryRecordOfTheEightFiles(t *testing.T) {
	wantOrder := []string{"categories", "products", "reviews", "gallery", "team", "services", "faqs", "shop"}
	wantCount := map[string]int{"categories": 4, "products": 14, "reviews": 12, "gallery": 7, "team": 4, "services": 6, "faqs": 8, "shop": 1}

	chunks := strings.Split(read(t, seedUp), "INSERT INTO ")[1:]
	if len(chunks) != len(wantOrder) {
		t.Fatalf("the seed has %d statements, want one per table (%d)", len(chunks), len(wantOrder))
	}
	for i, chunk := range chunks {
		table := strings.Fields(chunk)[0]
		if table != wantOrder[i] {
			t.Errorf("statement %d seeds %s, want %s", i+1, table, wantOrder[i])
		}
		if n := len(seedRecord.FindAllString(chunk, -1)); n != wantCount[table] {
			t.Errorf("%s: the seed holds %d records, want %d", table, n, wantCount[table])
		}
		if !strings.Contains(chunk, `ON CONFLICT ("id") DO NOTHING;`) {
			t.Errorf("%s: the statement does not end in ON CONFLICT (\"id\") DO NOTHING", table)
		}
		if strings.Contains(chunk, `"seq"`) {
			t.Errorf("%s: the seed names seq, which only the database assigns", table)
		}
	}
}

// The seed's down file removes exactly what the seed inserted.
func TestTheSeedDownFileListsEverySeededID(t *testing.T) {
	down := read(t, seedDown)
	ids := seedRecord.FindAllStringSubmatch(read(t, seedUp), -1)
	if len(ids) != 56 {
		t.Fatalf("the seed holds %d records, want 56", len(ids))
	}
	for _, m := range ids {
		if !strings.Contains(down, "'"+m[1]+"'") {
			t.Errorf("the down file does not delete %s", m[1])
		}
	}
}

// Every statement of the collections migration names its columns, so no
// column the API does not answer can reach a visitor by accident, and the
// migration creates what the contract names.
func TestTheCollectionsMigrationCreatesTheEightTablesAndTheThreeFunctions(t *testing.T) {
	body := read(t, "20261118000100_collections.up.sql")
	for _, want := range []string{
		"CREATE FUNCTION fold_text(", "CREATE FUNCTION natural_key(", "CREATE FUNCTION touch_record(",
		"CREATE TABLE categories (", "CREATE TABLE products (", "CREATE TABLE reviews (", "CREATE TABLE gallery (",
		"CREATE TABLE team (", "CREATE TABLE services (", "CREATE TABLE faqs (", "CREATE TABLE shop (",
		"trg_products_touch", "trg_categories_touch",
		"fk_products_category", "fk_reviews_bike", "fk_gallery_type",
		"ck_products_live_has_price", "ck_products_stock_nonneg", "ck_reviews_date_iso",
		"uq_products_seq", "idx_products_category", "idx_reviews_bike", "idx_gallery_type",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("the collections migration does not contain %q", want)
		}
	}
	// fold_text is created before natural_key, which calls it.
	if strings.Index(body, "CREATE FUNCTION fold_text(") > strings.Index(body, "CREATE FUNCTION natural_key(") {
		t.Error("natural_key is created before fold_text, which it calls")
	}
}
