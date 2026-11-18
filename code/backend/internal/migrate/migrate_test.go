package migrate

import (
	"reflect"
	"testing"
	"testing/fstest"
)

func TestLoadReadsUpFilesInFilenameOrder(t *testing.T) {
	fsys := fstest.MapFS{
		"20260102000000_second.up.sql":   {Data: []byte("SELECT 2;")},
		"20260101000000_first.up.sql":    {Data: []byte("SELECT 1;")},
		"20260101000000_first.down.sql":  {Data: []byte("SELECT 0;")},
		"20260102000000_second.down.sql": {Data: []byte("SELECT 0;")},
		"README.md":                      {Data: []byte("not sql")},
	}
	got, err := Load(fsys)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := []Migration{
		{Version: "20260101000000_first", SQL: "SELECT 1;"},
		{Version: "20260102000000_second", SQL: "SELECT 2;"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Load = %+v, want %+v", got, want)
	}
}

func TestSplitStatements(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want []string
	}{
		{"empty", "", nil},
		{"one statement", "SELECT 1;", []string{"SELECT 1"}},
		{"no final semicolon", "SELECT 1", []string{"SELECT 1"}},
		{"two statements", "SELECT 1;\nSELECT 2;\n", []string{"SELECT 1", "SELECT 2"}},
		{"semicolon in a string", "INSERT INTO t VALUES ('a;b');", []string{"INSERT INTO t VALUES ('a;b')"}},
		{"doubled quote", "SELECT 'it''s; fine';", []string{"SELECT 'it''s; fine'"}},
		{"semicolon in a quoted name", `SELECT 1 AS "a;b";`, []string{`SELECT 1 AS "a;b"`}},
		{"dollar quoted body", "DO $$ BEGIN PERFORM 1; END $$;", []string{"DO $$ BEGIN PERFORM 1; END $$"}},
		{"tagged dollar quote", "SELECT $f$a;b$f$;", []string{"SELECT $f$a;b$f$"}},
		{"placeholder is not a tag", "SELECT $1;", []string{"SELECT $1"}},
		{"line comment dropped", "-- a; comment\nSELECT 1;", []string{"SELECT 1"}},
		{"block comment dropped", "SELECT 1; /* x; y */", []string{"SELECT 1"}},
		{"trailing comment only", "SELECT 1; -- done", []string{"SELECT 1"}},
		{"only comments", "-- nothing\n/* here */", nil},
		{"concurrent index", "CREATE INDEX CONCURRENTLY IF NOT EXISTS i ON t (a);", []string{"CREATE INDEX CONCURRENTLY IF NOT EXISTS i ON t (a)"}},
		{"unterminated string runs to the end", "SELECT 'abc; SELECT 2;", []string{"SELECT 'abc; SELECT 2;"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := splitStatements(c.sql); !reflect.DeepEqual(got, c.want) {
				t.Errorf("splitStatements(%q) = %q, want %q", c.sql, got, c.want)
			}
		})
	}
}

func TestNeedsNoTransaction(t *testing.T) {
	cases := []struct {
		name string
		sql  string
		want bool
	}{
		{"plain DDL", "CREATE TABLE t (id int);", false},
		{"concurrent index", "create index concurrently i on t (a);", true},
		{"mentioned only in a comment", "-- not concurrently\nCREATE TABLE t (id int);", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := needsNoTransaction(splitStatements(c.sql)); got != c.want {
				t.Errorf("needsNoTransaction = %v, want %v", got, c.want)
			}
		})
	}
}
