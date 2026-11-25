package collections

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

// Writable reports whether the owner may add, change and delete records of
// the collection: the two that carry the compare-and-swap version, bikes and
// riding types. The other six are read-only; the API answers 405 to a write.
func (c Spec) Writable() bool { return c.Versioned }

// The statements of the owner's writes. Like the reads they are fixed text
// built from the field table alone: a column name is quoted and comes from
// there, a value from a request is only ever a numbered placeholder, and
// every placeholder has an explicit cast. seq, version, created_at,
// updated_at and _source are never in an INSERT or SET list: seq is the
// database's, version and updated_at belong to the trigger, and _source is
// untouched by a replace and NULL on an add.

type writeSQL struct {
	add, replace, lookup, version, remove string
}

var writeStatements = func() map[string]writeSQL {
	m := map[string]writeSQL{}
	for _, s := range All {
		if s.Writable() {
			m[s.Name] = buildWrites(s)
		}
	}
	return m
}()

// castOf is the cast every placeholder of a field carries. A number that is
// not a whole amount is bound as its text and cast through text, so a stored
// 2027.0 reads back as 2027.0 and nothing passes through a float.
func castOf(f Field) string {
	switch f.Kind {
	case KindNumber:
		if wholeNumbers[f.Key] {
			return "::bigint"
		}
		return "::text::numeric"
	case KindBool:
		return "::boolean"
	case KindList:
		return "::text[]"
	}
	return "::text"
}

// returningList is the columns an add or a replace answers: the data story's
// answer columns, then the version, so the answer is the record a later read
// returns.
func returningList(s Spec) string {
	cols := append(s.AnswerColumns(), "version")
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = quote(col)
	}
	return strings.Join(parts, ", ")
}

const answerRow = `SELECT jsonb_strip_nulls(to_jsonb(w) - 'version') AS rec, w."version" FROM w`

func buildWrites(s Spec) writeSQL {
	table := quote(s.Name)
	cols := make([]string, len(s.Fields))
	adds := make([]string, len(s.Fields))
	sets := make([]string, len(s.Fields))
	for i, f := range s.Fields {
		cast := castOf(f)
		cols[i] = quote(f.Key)
		adds[i] = "$" + strconv.Itoa(i+2) + cast
		sets[i] = quote(f.Key) + " = $" + strconv.Itoa(i+3) + cast
	}
	w := writeSQL{
		add: "WITH w AS (\n  INSERT INTO " + table + " (\"id\", " + strings.Join(cols, ", ") + ")\n  VALUES ($1::text, " + strings.Join(adds, ", ") +
			")\n  ON CONFLICT (\"id\") DO NOTHING\n  RETURNING " + returningList(s) + ")\n" + answerRow,
		replace: "WITH w AS (\n  UPDATE " + table + " SET " + strings.Join(sets, ", ") +
			"\n  WHERE \"id\" = $1::text AND \"version\" = $2::bigint\n  RETURNING " + returningList(s) + ")\n" + answerRow,
		version: "SELECT \"version\" FROM " + table + " WHERE \"id\" = $1::text",
		remove:  "DELETE FROM " + table + " WHERE \"id\" = $1::text RETURNING \"id\"",
	}
	w.lookup = "SELECT (SELECT t.\"version\" FROM " + table + " AS t WHERE t.\"id\" = $1::text) AS version, true AS category_ok"
	if _, ok := s.Field("category"); ok {
		w.lookup = "SELECT (SELECT t.\"version\" FROM " + table + " AS t WHERE t.\"id\" = $1::text) AS version, " +
			"($2::text IS NULL OR EXISTS (SELECT 1 FROM \"categories\" AS c WHERE c.\"id\" = $2::text)) AS category_ok"
	}
	return w
}

const (
	sqlCategoryExists = `SELECT EXISTS (SELECT 1 FROM "categories" WHERE "id" = $1::text)`
	sqlInUse          = `SELECT c."name", (SELECT count(*) FROM "products" AS p WHERE p."category" = c."id") FROM "categories" AS c WHERE c."id" = $1::text`
)

// Writer is the store behind the owner's writes. Each method is one
// statement answered by one row, so the pool's QueryRow is all it needs; a
// test stands in with a fake that can fail any step.
type Writer interface {
	// Add inserts a record; added is false when the id is already taken.
	Add(ctx context.Context, s Spec, id string, values []any) (rec []byte, version int64, added bool, err error)
	// Lookup reads the record's version (found is false when there is none)
	// and whether the riding type named exists (true when none is named).
	Lookup(ctx context.Context, s Spec, id string, category *string) (version int64, found, categoryOK bool, err error)
	// CategoryExists reports whether a riding type with the id exists.
	CategoryExists(ctx context.Context, id string) (bool, error)
	// Replace overwrites the record when its version is still the one given;
	// replaced is false when no row matched.
	Replace(ctx context.Context, s Spec, id string, version int64, values []any) (rec []byte, newVersion int64, replaced bool, err error)
	// Version reads the record's current version, after a replace matched nothing.
	Version(ctx context.Context, s Spec, id string) (version int64, found bool, err error)
	// Remove deletes the record; found is false when there was none. A riding
	// type that a bike uses fails with a foreign-key error (SQLSTATE 23503).
	Remove(ctx context.Context, s Spec, id string) (found bool, err error)
	// InUse reads a riding type's name and how many bikes use it.
	InUse(ctx context.Context, id string) (name string, bikes int64, found bool, err error)
}

type pgWriter struct{ db DB }

func (p pgWriter) Add(ctx context.Context, s Spec, id string, values []any) ([]byte, int64, bool, error) {
	var (
		rec     []byte
		version int64
	)
	err := p.db.QueryRow(ctx, writeStatements[s.Name].add, append([]any{id}, values...)...).Scan(&rec, &version)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, false, nil
	}
	return rec, version, err == nil, err
}

func (p pgWriter) Lookup(ctx context.Context, s Spec, id string, category *string) (int64, bool, bool, error) {
	args := []any{id}
	if _, ok := s.Field("category"); ok {
		args = append(args, category)
	}
	var (
		version    *int64
		categoryOK bool
	)
	if err := p.db.QueryRow(ctx, writeStatements[s.Name].lookup, args...).Scan(&version, &categoryOK); err != nil {
		return 0, false, false, err
	}
	if version == nil {
		return 0, false, categoryOK, nil
	}
	return *version, true, categoryOK, nil
}

func (p pgWriter) CategoryExists(ctx context.Context, id string) (bool, error) {
	var ok bool
	err := p.db.QueryRow(ctx, sqlCategoryExists, id).Scan(&ok)
	return ok, err
}

func (p pgWriter) Replace(ctx context.Context, s Spec, id string, version int64, values []any) ([]byte, int64, bool, error) {
	var (
		rec        []byte
		newVersion int64
	)
	err := p.db.QueryRow(ctx, writeStatements[s.Name].replace, append([]any{id, version}, values...)...).Scan(&rec, &newVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, 0, false, nil
	}
	return rec, newVersion, err == nil, err
}

func (p pgWriter) Version(ctx context.Context, s Spec, id string) (int64, bool, error) {
	var version int64
	err := p.db.QueryRow(ctx, writeStatements[s.Name].version, id).Scan(&version)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	return version, err == nil, err
}

func (p pgWriter) Remove(ctx context.Context, s Spec, id string) (bool, error) {
	var gone string
	err := p.db.QueryRow(ctx, writeStatements[s.Name].remove, id).Scan(&gone)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	return err == nil, err
}

func (p pgWriter) InUse(ctx context.Context, id string) (string, int64, bool, error) {
	var (
		name  string
		bikes int64
	)
	err := p.db.QueryRow(ctx, sqlInUse, id).Scan(&name, &bikes)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", 0, false, nil
	}
	return name, bikes, err == nil, err
}
