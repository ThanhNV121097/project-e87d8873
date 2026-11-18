package collections

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A column in the field table, in the schema file and in the migration must
// not drift apart: this test is what turns a missing or misspelt key into a
// red build instead of a record that quietly loses a field.
func TestFieldTableMatchesTheSchemaFile(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "frontend", "src", "data", "schema.json"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Skip("schema.json is not in this checkout (a Docker build context holds code/backend only)")
	}
	if err != nil {
		t.Fatal(err)
	}
	var file struct {
		Collections map[string]struct {
			Fields []struct {
				Key  string `json:"key"`
				Type string `json:"type"`
			} `json:"fields"`
		} `json:"collections"`
	}
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatal(err)
	}

	if len(file.Collections) != len(All) {
		t.Errorf("schema.json has %d collections, the field table %d", len(file.Collections), len(All))
	}
	for _, s := range All {
		def, ok := file.Collections[s.Name]
		if !ok {
			t.Errorf("%s is in the field table and not in schema.json", s.Name)
			continue
		}
		if len(def.Fields) != len(s.Fields) {
			t.Errorf("%s: schema.json has %d fields, the field table %d", s.Name, len(def.Fields), len(s.Fields))
			continue
		}
		for i, f := range def.Fields {
			if s.Fields[i].Key != f.Key {
				t.Errorf("%s field %d: schema.json says %q, the field table %q", s.Name, i, f.Key, s.Fields[i].Key)
				continue
			}
			if want := kindOf(t, f.Type); s.Fields[i].Kind != want {
				t.Errorf("%s.%s: type %q is kind %d, the field table says %d", s.Name, f.Key, f.Type, want, s.Fields[i].Kind)
			}
		}
	}
}

func kindOf(t *testing.T, typ string) Kind {
	t.Helper()
	switch {
	case typ == "text" || typ == "longtext" || typ == "image" || typ == "url":
		return KindText
	case typ == "number" || typ == "price":
		return KindNumber
	case typ == "bool":
		return KindBool
	case typ == "list":
		return KindList
	case strings.HasPrefix(typ, "ref:"):
		return KindRef
	}
	t.Errorf("schema.json has a field type the field table does not know: %q", typ)
	return KindText
}

func TestLookupIsExact(t *testing.T) {
	for _, name := range []string{"products", "categories", "reviews", "gallery", "team", "services", "faqs", "shop"} {
		if _, ok := Lookup(name); !ok {
			t.Errorf("%s is not found", name)
		}
	}
	for _, name := range []string{"", "nothing", "PRODUCTS", "product", "orders", "me", "sessions", "products "} {
		if _, ok := Lookup(name); ok {
			t.Errorf("%q: found, want no such collection", name)
		}
	}
}

func TestOnlyProductsAndCategoriesAreVersioned(t *testing.T) {
	for _, s := range All {
		want := s.Name == "products" || s.Name == "categories"
		if s.Versioned != want {
			t.Errorf("%s: Versioned = %v, want %v", s.Name, s.Versioned, want)
		}
	}
}

func TestFieldNamesThatAreNotSchemaKeysAreNotFound(t *testing.T) {
	p, _ := Lookup("products")
	for _, name := range []string{"_source", "seq", "version", "created_at", "updated_at", "colourful", ""} {
		if _, ok := p.Field(name); ok {
			t.Errorf("%q: found as a field", name)
		}
	}
	for _, name := range []string{"id", "price", "_invented"} {
		if _, ok := p.Field(name); !ok {
			t.Errorf("%q: not found as a field", name)
		}
	}
}

func TestDefaultSearchIsTheIDAndTheTextLikeFieldsOnly(t *testing.T) {
	p, _ := Lookup("products")
	got := map[string]bool{}
	for _, f := range p.DefaultSearch() {
		got[f.Key] = true
	}
	for _, key := range []string{"id", "name", "brand", "category", "price_listed", "fact", "status", "wheel", "image", "image_back", "image_detail", "description", "ride"} {
		if !got[key] {
			t.Errorf("%s is not searched by default", key)
		}
	}
	for _, key := range []string{"price", "was", "stock", "year", "rating", "deal", "month", "sizes", "sizes_out", "size_fit", "_invented", "_source"} {
		if got[key] {
			t.Errorf("%s is searched by default, and must not be", key)
		}
	}
}

func TestAnswerColumnsStartWithTheIDAndEndWithTheSource(t *testing.T) {
	for _, s := range All {
		cols := s.AnswerColumns()
		if cols[0] != "id" || cols[len(cols)-1] != "_source" || len(cols) != len(s.Fields)+2 {
			t.Errorf("%s: columns = %v", s.Name, cols)
		}
		for _, col := range cols {
			if col == "seq" || col == "version" || col == "created_at" || col == "updated_at" {
				t.Errorf("%s: the answer would carry %s", s.Name, col)
			}
		}
	}
}
