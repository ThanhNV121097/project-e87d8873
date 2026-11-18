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
