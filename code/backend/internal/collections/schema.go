// Package collections serves the shop's eight record collections: it reads a
// request's query, turns it into one SQL statement that carries every value
// as a bound parameter, and answers the page or the record.
package collections

// Kind is how a field behaves in a condition, a search and a sort. It comes
// from the field's type in code/frontend/src/data/schema.json.
type Kind int

const (
	// KindID is the record's id.
	KindID Kind = iota
	// KindText is a text, long text, image or url field.
	KindText
	// KindRef is a reference to a record of another collection.
	KindRef
	// KindNumber is a number or a price.
	KindNumber
	// KindBool is a true/false field.
	KindBool
	// KindList is an ordered list of text.
	KindList
)

// TextLike reports whether the kind is compared, searched and ordered as text.
func (k Kind) TextLike() bool { return k == KindID || k == KindText || k == KindRef }

// Field is one column of a collection. Key is the schema's key and the
// column's name.
type Field struct {
	Key  string
	Kind Kind
}

// Spec is one collection. Fields lists the schema's fields in schema order,
// without the id; Noun is the plural the 400 sentences use. The table has the
// collection's name. Versioned collections carry the compare-and-swap token
// the record answer sends as its ETag.
type Spec struct {
	Name      string
	Noun      string
	Versioned bool
	Fields    []Field
}

func txt(k string) Field  { return Field{k, KindText} }
func ref(k string) Field  { return Field{k, KindRef} }
func num(k string) Field  { return Field{k, KindNumber} }
func flag(k string) Field { return Field{k, KindBool} }
func list(k string) Field { return Field{k, KindList} }

// All is the closed set of collections, and the only place a column name
// comes from: no request ever supplies one.
var All = []Spec{
	{Name: "products", Noun: "bikes", Versioned: true, Fields: []Field{
		txt("image"), txt("name"), txt("brand"), ref("category"), num("year"), num("price"), num("was"),
		txt("price_listed"), txt("fact"), txt("status"), num("stock"), list("sizes"), list("sizes_out"),
		list("size_fit"), flag("deal"), flag("month"), txt("wheel"), txt("travel"), txt("frame"), txt("fork"),
		txt("shock"), txt("motor"), txt("drivetrain"), txt("brakes"), txt("wheels"), txt("tyres"),
		txt("weight"), txt("colour"), txt("description"), txt("ride"), num("rating"), num("review_count"),
		txt("image_back"), txt("image_detail"), list("_invented"),
	}},
	{Name: "categories", Noun: "riding types", Versioned: true, Fields: []Field{
		txt("name"), txt("line"), txt("terrain"),
	}},
	{Name: "reviews", Noun: "reviews", Fields: []Field{
		txt("name"), num("rating"), txt("date"), ref("bike"), txt("text"),
	}},
	{Name: "team", Noun: "team members", Fields: []Field{
		txt("name"), txt("role"), txt("line"), txt("rides"),
	}},
	{Name: "gallery", Noun: "photographs", Fields: []Field{
		txt("image"), txt("title"), txt("caption"), ref("type"),
	}},
	{Name: "services", Noun: "workshop jobs", Fields: []Field{
		txt("name"), txt("line"), txt("time"), num("price"),
	}},
	{Name: "faqs", Noun: "questions", Fields: []Field{
		txt("q"), txt("a"),
	}},
	{Name: "shop", Noun: "shop records", Fields: []Field{
		num("founded"), num("bikes_built"), num("rating"), num("review_count"), num("reply_hours"),
		num("warranty_years"), num("free_service_months"), txt("delivery_local"), txt("delivery_nationwide"),
		num("delivery_fee"), txt("pickup"), list("payment"), txt("confirm"), txt("stock_rule"), list("_invented"),
	}},
}

// Lookup finds a collection by its exact name.
func Lookup(name string) (Spec, bool) {
	for _, c := range All {
		if c.Name == name {
			return c, true
		}
	}
	return Spec{}, false
}

// Field finds a field by key; "id" is the record's own id. Names that are not
// schema keys (_source, seq, version, created_at, updated_at) are not found.
func (c Spec) Field(key string) (Field, bool) {
	if key == "id" {
		return Field{"id", KindID}, true
	}
	for _, f := range c.Fields {
		if f.Key == key {
			return f, true
		}
	}
	return Field{}, false
}

// AnswerColumns are the columns a record is built from, in this order: the
// id, every schema field, then the record's origin. A column a table gains
// later is not in the answer until it is added here.
func (c Spec) AnswerColumns() []string {
	cols := make([]string, 0, len(c.Fields)+2)
	cols = append(cols, "id")
	for _, f := range c.Fields {
		cols = append(cols, f.Key)
	}
	return append(cols, "_source")
}

// DefaultSearch is what q searches when no search= names fields: the id and
// every text-like schema field, in schema order.
func (c Spec) DefaultSearch() []Field {
	out := []Field{{"id", KindID}}
	for _, f := range c.Fields {
		if f.Kind.TextLike() {
			out = append(out, f)
		}
	}
	return out
}
