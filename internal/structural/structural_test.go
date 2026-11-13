package structural

import (
	"testing"
)

type want struct{ path, key, value string }

func check(t *testing.T, content string, got []Assignment, wants []want) {
	t.Helper()
	if len(got) != len(wants) {
		t.Fatalf("got %d assignments, want %d: %+v", len(got), len(wants), got)
	}
	for i, w := range wants {
		a := got[i]
		if a.KeyPath != w.path || a.Key != w.key || a.Value != w.value {
			t.Errorf("[%d] = %+v, want %+v", i, a, w)
		}
		if raw := content[a.Start:a.End]; raw != a.Value {
			t.Errorf("[%d] span %d:%d = %q, want the value %q", i, a.Start, a.End, raw, a.Value)
		}
	}
}

func TestDetect(t *testing.T) {
	tests := map[string]Format{
		"a/config.yaml": FormatYAML, "x.YML": FormatYAML, "pkg.json": FormatJSON,
		".env": FormatDotenv, ".env.local": FormatDotenv, "prod.env": FormatDotenv,
		"app.properties": FormatProperties, "setup.cfg": FormatINI, "php.ini": FormatINI,
		"main.tf": FormatHCL, "vars.tfvars": FormatHCL, "main.go": FormatNone, "README": FormatNone,
	}
	for name, f := range tests {
		if got := Detect(name); got != f {
			t.Errorf("Detect(%q) = %q, want %q", name, got, f)
		}
	}
}

func TestYAML(t *testing.T) {
	content := "db:\n  credentials:\n    password: \"p@ss word\"\n    port: 5432\n    tls: true\n  hosts:\n    - token: 'tok-1'\nключ: значение\n---\nsecond: doc\n"
	got, err := Extract(FormatYAML, []byte(content))
	if err != nil {
		t.Fatal(err)
	}
	check(t, content, got, []want{
		{"db.credentials.password", "password", "p@ss word"},
		{"db.credentials.port", "port", "5432"},
		{"db.hosts.0.token", "token", "tok-1"},
		{"ключ", "ключ", "значение"},
		{"second", "second", "doc"},
	})
	if _, err := Extract(FormatYAML, []byte("a: [unclosed\n")); err == nil {
		t.Error("invalid YAML: want an error")
	}
}

func TestJSON(t *testing.T) {
	content := `{"auth": {"token": "t\"ok", "n": 1, "list": ["a", {"secret": "s"}]}, "top": "v"}`
	got, err := Extract(FormatJSON, []byte(content))
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %+v", got)
	}
	if got[0].KeyPath != "auth.token" || got[0].Value != `t"ok` || content[got[0].Start:got[0].End] != `t\"ok` {
		t.Errorf("escaped value: %+v", got[0])
	}
	check(t, content, got[1:], []want{{"auth.list.0", "0", "a"}, {"auth.list.1.secret", "secret", "s"}, {"top", "top", "v"}})
	if _, err := Extract(FormatJSON, []byte(`{"a": `)); err == nil {
		t.Error("invalid JSON: want an error")
	}
}

func TestDotenv(t *testing.T) {
	content := "# comment\nexport API_TOKEN=\"abc def\"\nDB_PASSWORD=hunter2 # note\nEMPTY=\nSINGLE='x'\n"
	got, _ := Extract(FormatDotenv, []byte(content))
	check(t, content, got, []want{
		{"API_TOKEN", "API_TOKEN", "abc def"},
		{"DB_PASSWORD", "DB_PASSWORD", "hunter2"},
		{"EMPTY", "EMPTY", ""},
		{"SINGLE", "SINGLE", "x"},
	})
}

func TestPropertiesAndINI(t *testing.T) {
	props := "# c\n! c\ndb.password = hunter2\nkey: value\n"
	got, _ := Extract(FormatProperties, []byte(props))
	check(t, props, got, []want{{"db.password", "password", "hunter2"}, {"key", "key", "value"}})

	ini := "; c\n[database]\npassword = \"hunter2\"\n[other]\nuser=me\n"
	got, _ = Extract(FormatINI, []byte(ini))
	check(t, ini, got, []want{{"database.password", "password", "hunter2"}, {"other.user", "user", "me"}})
}

func TestHCL(t *testing.T) {
	content := "resource \"aws_db\" \"main\" {\n  password = \"hunter2\"\n  user     = var.user\n  nested {\n    token = \"t\"\n  }\n}\nregion = \"us-east-1\"\n"
	got, _ := Extract(FormatHCL, []byte(content))
	check(t, content, got, []want{
		{"resource.aws_db.main.password", "password", "hunter2"},
		{"resource.aws_db.main.nested.token", "token", "t"},
		{"region", "region", "us-east-1"},
	})
}

func TestGeneric(t *testing.T) {
	content := `const sessionSecret = "abc"; cfg.apiKey: 'k', x := ` + "`raw`" + `, flag = bare_value
if token == "nope" { other != "no" }
`
	got := Generic([]byte(content), 0)
	check(t, content, got, []want{
		{"sessionSecret", "sessionSecret", "abc"},
		{"cfg.apiKey", "apiKey", "k"},
		{"x", "x", "raw"},
		{"flag", "flag", "bare_value"},
	})
	if got := Generic([]byte("a = \"b\""), 100); got[0].Start != 105 {
		t.Errorf("base offset: %+v", got)
	}
}
