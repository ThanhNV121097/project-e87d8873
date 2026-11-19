package config

import "testing"

func adminEnv(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestTheTwoAdminSecretsAreRead(t *testing.T) {
	c, err := Load(adminEnv(map[string]string{"DATABASE_URL": "postgres://x", "ADMIN_EMAIL": "  Owner@deo.test ", "ADMIN_PASSWORD": "s3cret"}))
	if err != nil {
		t.Fatal(err)
	}
	if c.AdminEmail != "Owner@deo.test" || c.AdminPassword != "s3cret" {
		t.Errorf("email %q, password %q", c.AdminEmail, c.AdminPassword)
	}
}

// A secret pasted into a console can carry a trailing newline nobody can
// type; only that is removed, and only from the password: its leading and
// trailing spaces are part of it.
func TestOnlyTrailingLineBreaksAreRemovedFromThePassword(t *testing.T) {
	for in, want := range map[string]string{
		"s3cret\n":      "s3cret",
		"s3cret\r\n":    "s3cret",
		"s3cret\n\n":    "s3cret",
		" s3cret ":      " s3cret ",
		" s3 cret \r\n": " s3 cret ",
		"\ns3cret":      "\ns3cret",
	} {
		c, err := Load(adminEnv(map[string]string{"DATABASE_URL": "postgres://x", "ADMIN_EMAIL": "a@b.c", "ADMIN_PASSWORD": in}))
		if err != nil {
			t.Fatal(err)
		}
		if c.AdminPassword != want {
			t.Errorf("password %q became %q, want %q", in, c.AdminPassword, want)
		}
	}
}

func TestABlankSecretIsNotSetAndNeitherIsRequired(t *testing.T) {
	for name, env := range map[string]map[string]string{
		"both unset":       {},
		"white space only": {"ADMIN_EMAIL": "   ", "ADMIN_PASSWORD": " \t "},
		"line breaks only": {"ADMIN_EMAIL": "a@b.c", "ADMIN_PASSWORD": "\r\n"},
	} {
		env["DATABASE_URL"] = "postgres://x"
		c, err := Load(adminEnv(env))
		if err != nil {
			t.Errorf("%s: Load failed: %v", name, err)
			continue
		}
		if c.AdminPassword != "" || (name != "line breaks only" && c.AdminEmail != "") {
			t.Errorf("%s: email %q, password %q, want not set", name, c.AdminEmail, c.AdminPassword)
		}
	}
}

func TestLoadStillFailsOnlyForTheDatabaseAndThePort(t *testing.T) {
	if _, err := Load(adminEnv(map[string]string{"ADMIN_EMAIL": "a@b.c", "ADMIN_PASSWORD": "x"})); err == nil {
		t.Error("no DATABASE_URL: Load succeeded")
	}
	if _, err := Load(adminEnv(map[string]string{"DATABASE_URL": "postgres://x", "PORT": "nope"})); err == nil {
		t.Error("a bad port: Load succeeded")
	}
}
