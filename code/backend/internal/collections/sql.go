package collections

import (
	"strconv"
	"strings"
)

// Every statement below is built from fixed fragments and the fixed field
// table of schema.go only: a column name is written into the text from that
// table, quoted, and never from a request; a value from a request is written
// only as a numbered placeholder. Every placeholder carries an explicit cast,
// so a prepared statement never has to guess a type.

// quote writes a column or table name from the fixed table as an identifier.
func quote(name string) string { return `"` + name + `"` }

// answerColumns lists the columns a record is built from, qualified by alias.
func answerColumns(s Spec, alias string) string {
	cols := s.AnswerColumns()
	parts := make([]string, len(cols))
	for i, col := range cols {
		parts[i] = alias + "." + quote(col)
	}
	return strings.Join(parts, ", ")
}

// foldedValues turns a text[] placeholder into the array of its folded values.
func foldedValues(p string) string {
	return "ARRAY(SELECT fold_text(v) FROM unnest(" + p + "::text[]) AS v)"
}

// conditionSQL writes one condition. arg stores a value and returns its
// placeholder; each branch calls it exactly once.
func conditionSQL(cd Condition, arg func(any) string) string {
	col := "t." + quote(cd.Field.Key)
	if cd.Op != OpEq {
		op := ">="
		if cd.Op == OpTo {
			op = "<="
		}
		p := arg(cd.Value)
		if cd.Field.Kind == KindNumber {
			return col + " " + op + " " + p + "::text::numeric"
		}
		return col + ` COLLATE "C" ` + op + " " + p + "::text"
	}
	switch cd.Field.Kind {
	case KindID, KindRef:
		// A stored id or reference is already folded by construction, so only
		// the values are folded: the same answer, and the index serves it.
		return col + " = ANY (" + foldedValues(arg(cd.Values)) + ")"
	case KindNumber:
		return col + " = ANY (" + arg(cd.Values) + "::text[]::numeric[])"
	case KindBool:
		return col + " = ANY (" + arg(cd.Bools) + "::boolean[])"
	case KindList:
		return "EXISTS (SELECT 1 FROM unnest(" + col + ") AS e WHERE fold_text(e) = ANY (" + foldedValues(arg(cd.Values)) + "))"
	default:
		return "fold_text(" + col + ") = ANY (" + foldedValues(arg(cd.Values)) + ")"
	}
}

// searchSQL writes the search words as one condition over the searched
// fields. strpos has no wildcards, so %, _ and \ mean themselves and nothing
// is escaped; p is the one placeholder every field reuses.
func searchSQL(s Spec, q Query, p string) string {
	fields := q.Search
	if len(fields) == 0 {
		fields = s.DefaultSearch()
	}
	terms := make([]string, len(fields))
	for i, f := range fields {
		terms[i] = "strpos(fold_text(t." + quote(f.Key) + "), fold_text(" + p + "::text)) > 0"
	}
	return "(" + strings.Join(terms, " OR ") + ")"
}

// orderSQL writes the window's order. It always ends in seq, so equal values
// keep the order they were seeded or added in, in both directions, and a
// record with no value comes after those that have one.
func orderSQL(q Query) string {
	if q.Sort == nil {
		return `m."seq"`
	}
	col := "m." + quote(q.Sort.Key)
	dir := "ASC"
	if q.Desc {
		dir = "DESC"
	}
	if q.Sort.Kind.TextLike() {
		return "natural_key(" + col + `) COLLATE "C" ` + dir + ` NULLS LAST, m."seq"`
	}
	return col + " " + dir + ` NULLS LAST, m."seq"`
}

// BuildList writes the statement of a list request and its arguments. It
// answers exactly one row, (total bigint, items jsonb), even for an empty
// page: the page is cut with row_number() and not OFFSET/LIMIT, so there is no
// second statement to disagree with the first about the total.
//
// Placeholders are numbered in a fixed order: the field conditions in the
// alphabetical order of their parameter names, the search words, then the two
// ends of the page.
func BuildList(s Spec, q Query) (string, []any) {
	var args []any
	arg := func(v any) string {
		args = append(args, v)
		return "$" + strconv.Itoa(len(args))
	}

	var where []string
	for _, cd := range q.Conditions {
		where = append(where, conditionSQL(cd, arg))
	}
	if q.Words != "" {
		where = append(where, searchSQL(s, q, arg(q.Words)))
	}

	var b strings.Builder
	b.WriteString("WITH m AS MATERIALIZED (\n  SELECT t.\"seq\", " + answerColumns(s, "t") + "\n  FROM " + quote(s.Name) + " AS t")
	if len(where) > 0 {
		b.WriteString("\n  WHERE " + strings.Join(where, " AND "))
	}
	from := (q.Page - 1) * q.Limit
	b.WriteString("\n)\nSELECT\n  (SELECT count(*) FROM m) AS total,\n  COALESCE((\n    SELECT jsonb_agg(p.rec ORDER BY p.rn)\n    FROM (\n      SELECT jsonb_strip_nulls(to_jsonb(m) - 'seq') AS rec,\n             row_number() OVER (ORDER BY " + orderSQL(q) + ") AS rn\n      FROM m\n    ) AS p\n    WHERE p.rn > ")
	b.WriteString(arg(from) + "::bigint AND p.rn <= ")
	b.WriteString(arg(from+q.Limit) + "::bigint\n  ), '[]'::jsonb) AS items")
	return b.String(), args
}

// BuildRecord writes the statement of a record request; its one argument is
// the id. A versioned collection also answers its version, which becomes the
// ETag and never part of the record.
func BuildRecord(s Spec) string {
	if s.Versioned {
		return "SELECT jsonb_strip_nulls(to_jsonb(r) - 'version') AS rec, r.\"version\"\nFROM (SELECT " +
			answerColumns(s, "t") + ", t.\"version\" FROM " + quote(s.Name) + " AS t WHERE t.\"id\" = $1::text) AS r"
	}
	return "SELECT jsonb_strip_nulls(to_jsonb(r)) AS rec\nFROM (SELECT " +
		answerColumns(s, "t") + " FROM " + quote(s.Name) + " AS t WHERE t.\"id\" = $1::text) AS r"
}
