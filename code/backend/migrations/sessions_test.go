package migrations

import (
	"strings"
	"testing"
)

const (
	sessionsUp   = "20261118000400_sessions.up.sql"
	sessionsDown = "20261118000400_sessions.down.sql"
)

// The sessions migration creates exactly the sessions table, as the ERD
// has it, and depends on no other table. CI has no database, so this reads
// the text.
func TestTheSessionsMigrationCreatesExactlyTheSessionsTable(t *testing.T) {
	up := read(t, sessionsUp)
	if n := strings.Count(up, "CREATE TABLE"); n != 1 || !strings.Contains(up, `CREATE TABLE "sessions" (`) {
		t.Errorf("the up file creates %d tables, want exactly sessions", n)
	}
	for _, want := range []string{
		`"id"         uuid        NOT NULL DEFAULT gen_random_uuid()`,
		`"token_hash" bytea       NOT NULL`,
		`"email"      text        NOT NULL`,
		`"created_at" timestamptz NOT NULL DEFAULT now()`,
		`"expires_at" timestamptz NOT NULL`,
		`CONSTRAINT sessions_pkey PRIMARY KEY ("id")`,
		`CONSTRAINT uq_sessions_token_hash UNIQUE ("token_hash")`,
		`CONSTRAINT ck_sessions_token_hash_length CHECK (octet_length("token_hash") = 32)`,
		`CONSTRAINT ck_sessions_expiry_after_creation CHECK ("expires_at" > "created_at")`,
	} {
		if !strings.Contains(up, want) {
			t.Errorf("the up file does not contain %q", want)
		}
	}
	lower := strings.ToLower(up)
	for _, never := range []string{"references", "create index", "concurrent" + "ly", "insert into", "alter table", `"products"`, `"categories"`} {
		if strings.Contains(lower, never) {
			t.Errorf("the up file contains %q: it must only create the one table", never)
		}
	}

	if down := strings.TrimSpace(read(t, sessionsDown)); down != `DROP TABLE "sessions";` {
		t.Errorf("the down file is %q, want exactly DROP TABLE \"sessions\";", down)
	}
}
