package collections

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"
)

const (
	defaultPage  = 1
	defaultLimit = 12
	maxLimit     = 100
	maxPage      = 1_000_000_000
	maxQueryLen  = 100
)

var (
	wholeNumber = regexp.MustCompile(`^[0-9]+$`)
	// A plain decimal, with a length cap so that a number no database can
	// hold is a 400 here and not a failed statement there.
	plainDecimal = regexp.MustCompile(`^-?[0-9]{1,30}(\.[0-9]{1,30})?$`)
)

// ownNames are the parameters that are the API's, not a field's.
var ownNames = map[string]bool{"page": true, "limit": true, "q": true, "search": true, "sort": true}

// Op is how a condition compares.
type Op int

const (
	// OpEq keeps records whose field equals any of the values.
	OpEq Op = iota
	// OpFrom keeps records whose field is at or above the value.
	OpFrom
	// OpTo keeps records whose field is at or below the value.
	OpTo
)

// Condition is one narrowing of the list on one field.
type Condition struct {
	Field Field
	Op    Op
	// Values are the alternatives of an equality: text, or the digits of a
	// number (cast by the database, never parsed to a float here).
	Values []string
	// Bools are the alternatives of an equality on a boolean field.
	Bools []bool
	// Value is the end of a range.
	Value string
}

// Query is a list request that has been read and checked. Everything after
// ParseQuery trusts it.
type Query struct {
	Page, Limit int64
	// Sort is nil when no order is asked: the seed's order.
	Sort *Field
	Desc bool
	// Search is nil when q searches the default set.
	Search []Field
	// Words are the trimmed search words; empty means no search.
	Words string
	// Conditions are in the alphabetical order of their parameter names.
	Conditions []Condition
}

// QueryError is the first refusal of a query: Field is the parameter or field
// at fault and Sentence is the one sentence the answer carries.
type QueryError struct {
	Field    string
	Sentence string
}

func fail(field, format string, a ...any) *QueryError {
	return &QueryError{Field: field, Sentence: fmt.Sprintf(format, a...)}
}

// ParseQuery reads the whole query string of a list request. The checks run
// in the order the service contract lists them and the first failure answers.
func ParseQuery(c Collection, v url.Values) (Query, *QueryError) {
	names := make([]string, 0, len(v))
	for name := range v {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		if len(v[name]) > 1 {
			return Query{}, fail(name, "The parameter “%s” was given more than once. Use a comma to give several values.", name)
		}
	}
	// Text that is not valid UTF-8, or holds a NUL, cannot be stored in or
	// compared by the database: refuse it here rather than fail there.
	for _, name := range names {
		if !cleanText(name) || !cleanText(v.Get(name)) {
			return Query{}, fail(name, "The parameter “%s” is not valid text.", name)
		}
	}

	q := Query{Page: defaultPage, Limit: defaultLimit}
	if _, ok := v["page"]; ok {
		n, ok := parseWhole(v.Get("page"))
		if !ok || n < 1 || n > maxPage {
			return Query{}, fail("page", "The page must be a whole number from 1 to %d.", maxPage)
		}
		q.Page = n
	}
	if _, ok := v["limit"]; ok {
		n, ok := parseWhole(v.Get("limit"))
		if !ok || n < 1 {
			return Query{}, fail("limit", "The limit must be a whole number of 1 or more.")
		}
		q.Limit = min(n, maxLimit)
	}

	if err := parseSort(c, strings.TrimSpace(v.Get("sort")), &q); err != nil {
		return Query{}, err
	}
	if err := parseSearch(c, v.Get("search"), &q); err != nil {
		return Query{}, err
	}

	words := strings.TrimSpace(v.Get("q"))
	if utf8.RuneCountInString(words) > maxQueryLen {
		return Query{}, fail("q", "The search words must be at most %d characters.", maxQueryLen)
	}
	q.Words = words

	for _, name := range names {
		if ownNames[name] {
			continue
		}
		cond, ok, err := parseCondition(c, name, v.Get(name))
		if err != nil {
			return Query{}, err
		}
		if ok {
			q.Conditions = append(q.Conditions, cond)
		}
	}
	return q, nil
}

func cleanText(s string) bool {
	return utf8.ValidString(s) && !strings.ContainsRune(s, 0)
}

// parseWhole reads a value that is nothing but decimal digits. A digit string
// too long for an int64 is not a whole number this API accepts.
func parseWhole(s string) (int64, bool) {
	if !wholeNumber.MatchString(s) {
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}

func parseSort(c Collection, raw string, q *Query) *QueryError {
	if raw == "" {
		return nil
	}
	name, desc := strings.CutPrefix(raw, "-")
	if name == "" {
		return fail("sort", "There is no field to sort %s by.", c.Noun)
	}
	f, ok := c.Field(name)
	if !ok {
		return fail(name, "There is no field “%s” to sort %s by.", name, c.Noun)
	}
	if f.Kind == KindList {
		return fail(name, "“%s” is a list and cannot be sorted.", name)
	}
	q.Sort, q.Desc = &f, desc
	return nil
}

func parseSearch(c Collection, raw string, q *Query) *QueryError {
	for _, part := range strings.Split(raw, ",") {
		name := strings.TrimSpace(part)
		if name == "" {
			continue
		}
		f, ok := c.Field(name)
		if !ok {
			return fail(name, "There is no field “%s” to search %s by.", name, c.Noun)
		}
		if !f.Kind.TextLike() {
			return fail(name, "“%s” cannot be searched.", name)
		}
		q.Search = append(q.Search, f)
	}
	return nil
}

// parseCondition classifies a parameter by its name, whatever its value, even
// an empty one: an unknown name is refused, and an empty value only means "no
// condition" for a name that is a field. ok is false for no condition.
func parseCondition(c Collection, name, raw string) (cond Condition, ok bool, err *QueryError) {
	if f, found := c.Field(name); found {
		return parseEquality(f, name, raw)
	}
	for _, suffix := range []struct {
		text string
		op   Op
	}{{"_from", OpFrom}, {"_to", OpTo}} {
		base, cut := strings.CutSuffix(name, suffix.text)
		if !cut {
			continue
		}
		f, found := c.Field(base)
		if !found {
			continue
		}
		return parseRange(f, suffix.op, name, raw)
	}
	return Condition{}, false, fail(name, "There is no field “%s” to filter %s by.", name, c.Noun)
}

func parseEquality(f Field, name, raw string) (Condition, bool, *QueryError) {
	var parts []string
	for _, p := range strings.Split(raw, ",") {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if len(parts) == 0 {
		return Condition{}, false, nil
	}
	cond := Condition{Field: f, Op: OpEq}
	switch f.Kind {
	case KindNumber:
		for _, p := range parts {
			if !plainDecimal.MatchString(p) {
				return Condition{}, false, fail(name, "The value of “%s” must be a number.", name)
			}
		}
		cond.Values = parts
	case KindBool:
		for _, p := range parts {
			switch p {
			case "true":
				cond.Bools = append(cond.Bools, true)
			case "false":
				cond.Bools = append(cond.Bools, false)
			default:
				return Condition{}, false, fail(name, "The value of “%s” must be true or false.", name)
			}
		}
	default:
		cond.Values = parts
	}
	return cond, true, nil
}

func parseRange(f Field, op Op, name, raw string) (Condition, bool, *QueryError) {
	if f.Kind == KindBool || f.Kind == KindList {
		return Condition{}, false, fail(name, "“%s” cannot be used: a range needs a number, a price or text.", name)
	}
	if raw == "" {
		return Condition{}, false, nil
	}
	if f.Kind == KindNumber && !plainDecimal.MatchString(raw) {
		return Condition{}, false, fail(name, "The value of “%s” must be a number.", name)
	}
	return Condition{Field: f, Op: op, Value: raw}, true, nil
}
