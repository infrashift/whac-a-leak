package rules

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// acmeRule is a valid user rule for a fictional provider.
const acmeRule = `package whac

rules: ACME_API_TOKEN: {
	description: "Acme API token"
	keywords: ["acme_"]
	regex: #"\b(acme_[A-Za-z0-9]{32})\b"#
	tests: mustMatch: ["token = acme_EXAMPLE0123456789abcdefghijklmno"]
}
`

func writeLayers(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func mustLoad(t *testing.T, files map[string]string) *Ruleset {
	t.Helper()
	rs, err := Load(Options{ConfigDir: writeLayers(t, files), RequireConfigDir: true})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return rs
}

func loadErr(t *testing.T, files map[string]string) string {
	t.Helper()
	_, err := Load(Options{ConfigDir: writeLayers(t, files), RequireConfigDir: true})
	if err == nil {
		t.Fatal("Load succeeded, want an error")
	}
	return err.Error()
}

func TestLoadBuiltins(t *testing.T) {
	rs, err := Load(Options{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(rs.Rules) < 5 {
		t.Fatalf("got %d builtin rules, want at least 5", len(rs.Rules))
	}
	for i, r := range rs.Rules {
		if i > 0 && rs.Rules[i-1].ID >= r.ID {
			t.Errorf("rules not sorted: %s before %s", rs.Rules[i-1].ID, r.ID)
		}
		if r.Kind != KindPattern || r.Regex == nil || len(r.Keywords) == 0 || len(r.Tests.MustMatch) == 0 {
			t.Errorf("%s: incomplete compiled rule %+v", r.ID, r)
		}
		for _, k := range r.Keywords {
			if k != strings.ToLower(k) {
				t.Errorf("%s: keyword %q not lowercased", r.ID, k)
			}
		}
	}
	if rs.Config.MaxFileSize != 2097152 {
		t.Errorf("MaxFileSize = %d, want the schema default", rs.Config.MaxFileSize)
	}
	if len(rs.Layers) != 0 {
		t.Errorf("Layers = %v, want none", rs.Layers)
	}
}

func TestLoadMissingConfigDir(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "nope")
	if _, err := Load(Options{ConfigDir: missing}); err != nil {
		t.Errorf("optional missing dir: %v", err)
	}
	if _, err := Load(Options{ConfigDir: missing, RequireConfigDir: true}); err == nil {
		t.Error("required missing dir: want an error")
	}
}

func TestLayering(t *testing.T) {
	rs := mustLoad(t, map[string]string{
		"a.cue": acmeRule,
		"b.cue": `package whac
overrides: JWT: enabled: false
overrides: GENERIC_API_KEY: severity: "LOW"
config: exclude: "vendor/**": "third-party code"
allowlist: fingerprints: "whac1:` + strings.Repeat("ab", 32) + `": "test fixture"
`,
	})
	if r := rs.Rule("ACME_API_TOKEN"); r == nil || r.Severity != SeverityHigh || !r.Enabled || r.SecretGroup != 1 {
		t.Errorf("added rule = %+v, want HIGH (schema default), enabled, secretGroup 1", r)
	}
	if r := rs.Rule("JWT"); r == nil || r.Enabled {
		t.Errorf("JWT should be disabled by override: %+v", r)
	}
	if r := rs.Rule("GENERIC_API_KEY"); r == nil || r.Severity != SeverityLow {
		t.Errorf("GENERIC_API_KEY should be re-graded LOW: %+v", r)
	}
	if got := rs.Config.Exclude["vendor/**"]; got != "third-party code" {
		t.Errorf("exclude = %q", got)
	}
	if len(rs.Allowlist.Fingerprints) != 1 {
		t.Errorf("fingerprints = %v", rs.Allowlist.Fingerprints)
	}
	if len(rs.Layers) != 2 {
		t.Errorf("Layers = %v, want 2 files", rs.Layers)
	}
	for _, r := range rs.Enabled() {
		if r.ID == "JWT" {
			t.Error("Enabled() includes a disabled rule")
		}
	}
}

func TestLayerTightensBuiltin(t *testing.T) {
	rs := mustLoad(t, map[string]string{"a.cue": `package whac
rules: GENERIC_API_KEY: stopwords: ["dummy"]
`})
	if got := rs.Rule("GENERIC_API_KEY").Stopwords; len(got) != 1 || got[0] != "dummy" {
		t.Errorf("stopwords = %v", got)
	}
}

func TestLayerMayReferenceSchemaDefinitions(t *testing.T) {
	rs := mustLoad(t, map[string]string{"a.cue": `package whac
_base: #SecretRule & {severity: "LOW", keywords: ["zz_"], tests: mustMatch: ["zz_abc"]}
rules: ZZ: _base & {description: "zz", regex: "(zz_[a-z]+)"}
`})
	if r := rs.Rule("ZZ"); r == nil || r.Severity != SeverityLow {
		t.Errorf("ZZ = %+v", r)
	}
}

func TestGeneratedSchemaCopyIsSkipped(t *testing.T) {
	// A stale exported schema that would conflict if it were loaded.
	stale := GeneratedHeader + "\npackage whac\nrules: [string]: severity: \"LOW\"\n"
	rs := mustLoad(t, map[string]string{"schema.cue": stale})
	if len(rs.Layers) != 0 {
		t.Errorf("generated schema copy was loaded: %v", rs.Layers)
	}
}

func TestLoadErrors(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  []string // substrings of the error
	}{
		{
			name:  "re-grading through rules conflicts with the concrete builtin severity",
			files: map[string]string{"a.cue": "package whac\nrules: AWS_ACCESS_KEY_ID: severity: \"LOW\"\n"},
			want:  []string{"rules.AWS_ACCESS_KEY_ID.severity", "a.cue:2"},
		},
		{
			name: "two layers disagree",
			files: map[string]string{
				"a.cue": "package whac\noverrides: JWT: severity: \"LOW\"\n",
				"b.cue": "package whac\noverrides: JWT: severity: \"HIGH\"\n",
			},
			want: []string{"overrides.JWT.severity", "a.cue", "b.cue"},
		},
		{
			name:  "override for an unknown rule",
			files: map[string]string{"a.cue": "package whac\noverrides: NOPE: enabled: false\n"},
			want:  []string{"overrides.NOPE", "no rule with this ID"},
		},
		{
			name:  "misspelt rule field is rejected by the closed definition",
			files: map[string]string{"a.cue": "package whac\nrules: JWT: severty: \"LOW\"\n"},
			want:  []string{"severty", "not allowed"},
		},
		{
			name:  "misspelt top-level field",
			files: map[string]string{"a.cue": "package whac\nrule: JWT: {}\n"},
			want:  []string{"rule: unknown top-level field", "a.cue:2"},
		},
		{
			name:  "wrong package",
			files: map[string]string{"a.cue": "package other\nrules: {}\n"},
			want:  []string{"package name"},
		},
		{
			name:  "invalid regex",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", keywords: [\"x\"], regex: \"x(\", tests: mustMatch: [\"x\"]}\n"},
			want:  []string{"rules.BAD", "missing closing )"},
		},
		{
			name:  "sample does not match the regex (CUE-side check)",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", keywords: [\"x\"], regex: \"x+\", tests: mustMatch: [\"yy\"]}\n"},
			want:  []string{"rules.BAD.tests.mustMatch.0"},
		},
		{
			name:  "pattern rule without keywords",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", keywords: [], regex: \"x+\", tests: mustMatch: [\"x\"]}\n"},
			want:  []string{"rules.BAD.keywords"},
		},
		{
			name:  "secretGroup beyond the capture groups",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", keywords: [\"x\"], regex: \"x+\", tests: mustMatch: [\"x\"]}\n"},
			want:  []string{"rules.BAD", "secretGroup 1"},
		},
		{
			name:  "entropy max below min",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", keywords: [\"x\"], regex: \"(x+)\", entropy: {min: 4, max: 3}, tests: mustMatch: [\"x\"]}\n"},
			want:  []string{"rules.BAD.entropy.max"},
		},
		{
			name:  "assignment rules are not supported yet",
			files: map[string]string{"a.cue": "package whac\nrules: BAD: {description: \"x\", kind: \"assignment\", keyContext: \"(?i)token\", tests: mustMatch: [\"x\"]}\n"},
			want:  []string{"rules.BAD", "assignment"},
		},
		{
			name:  "malformed allowlist fingerprint",
			files: map[string]string{"a.cue": "package whac\nallowlist: fingerprints: \"whac1:nothex\": \"x\"\n"},
			want:  []string{"allowlist.fingerprints"},
		},
		{
			name:  "allowlist entry without a reason",
			files: map[string]string{"a.cue": "package whac\nallowlist: paths: \"testdata/**\": \"\"\n"},
			want:  []string{"allowlist.paths"},
		},
		{
			name:  "invalid allowlist regex",
			files: map[string]string{"a.cue": "package whac\nallowlist: regexes: \"x(\": \"why\"\n"},
			want:  []string{"allowlist.regexes", "missing closing )"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := loadErr(t, tt.files)
			for _, w := range tt.want {
				if !strings.Contains(msg, w) {
					t.Errorf("error does not contain %q:\n%s", w, msg)
				}
			}
		})
	}
}

func TestLoadErrorType(t *testing.T) {
	_, err := Load(Options{ConfigDir: writeLayers(t, map[string]string{"a.cue": "package whac\nrules: JWT: severty: 1\n"}), RequireConfigDir: true})
	var le *LoadError
	if !errors.As(err, &le) {
		t.Fatalf("err = %T %v, want *LoadError", err, err)
	}
}

func TestSchemaSource(t *testing.T) {
	if !strings.Contains(string(SchemaSource()), "#SecretRule:") {
		t.Error("SchemaSource does not contain #SecretRule")
	}
}

func TestSeverity(t *testing.T) {
	for _, s := range []Severity{SeverityLow, SeverityMedium, SeverityHigh, SeverityCritical} {
		got, err := ParseSeverity(s.String())
		if err != nil || got != s {
			t.Errorf("round trip %v: got %v, %v", s, got, err)
		}
	}
	if _, err := ParseSeverity("SEVERE"); err == nil {
		t.Error("ParseSeverity(SEVERE): want an error")
	}
	if SeverityLow >= SeverityMedium || SeverityHigh >= SeverityCritical {
		t.Error("severities are not ordered")
	}
	if b, err := SeverityCritical.MarshalText(); err != nil || string(b) != "CRITICAL" {
		t.Errorf("MarshalText = %q, %v", b, err)
	}
	if _, err := Severity(0).MarshalText(); err == nil {
		t.Error("MarshalText(0): want an error")
	}
	var s Severity
	if err := s.UnmarshalText([]byte("MEDIUM")); err != nil || s != SeverityMedium {
		t.Errorf("UnmarshalText = %v, %v", s, err)
	}
	if err := s.UnmarshalText([]byte("nope")); err == nil {
		t.Error("UnmarshalText(nope): want an error")
	}
}

func BenchmarkLoad(b *testing.B) {
	for b.Loop() {
		if _, err := Load(Options{}); err != nil {
			b.Fatal(err)
		}
	}
}
