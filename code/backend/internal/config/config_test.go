package config

import "testing"

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadPortPrecedence(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"PORT wins", map[string]string{"DATABASE_URL": "postgres://x", "PORT": "9000", "APP_PORT": "9001"}, "9000"},
		{"APP_PORT next", map[string]string{"DATABASE_URL": "postgres://x", "APP_PORT": "9001"}, "9001"},
		{"blank PORT is unset", map[string]string{"DATABASE_URL": "postgres://x", "PORT": "  ", "APP_PORT": "9001"}, "9001"},
		{"default", map[string]string{"DATABASE_URL": "postgres://x"}, "8080"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Load(env(c.env))
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			if got.Port != c.want {
				t.Errorf("port = %q, want %q", got.Port, c.want)
			}
		})
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	if _, err := Load(env(map[string]string{"PORT": "8080"})); err == nil {
		t.Fatal("Load accepted a missing DATABASE_URL")
	}
	if _, err := Load(env(map[string]string{"DATABASE_URL": "   "})); err == nil {
		t.Fatal("Load accepted a blank DATABASE_URL")
	}
}

func TestLoadRejectsBadPort(t *testing.T) {
	for _, p := range []string{"abc", "0", "70000", "-1"} {
		if _, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x", "PORT": p})); err == nil {
			t.Errorf("Load accepted port %q", p)
		}
	}
}
