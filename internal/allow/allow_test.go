package allow

import (
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func TestMatch(t *testing.T) {
	fp := "whac1:" + strings.Repeat("ab", 32)
	al, err := New(rules.Allowlist{
		Fingerprints: map[string]string{fp: "test fixture"},
		Paths:        map[string]string{"testdata/**": "fixtures"},
		Regexes:      map[string]string{`^tok_EXAMPLE`: "documentation"},
	})
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		path, fp, secret string
		want             string
	}{
		{"src/a.go", fp, "x", "fingerprint: test fixture"},
		{"testdata/deep/b.env", "whac1:other", "x", "path testdata/**: fixtures"},
		{"src/a.go", "whac1:other", "tok_EXAMPLE123", "regex ^tok_EXAMPLE: documentation"},
		{"src/a.go", "whac1:other", "tok_real", ""},
	}
	for _, tt := range tests {
		if got := al.Match(tt.path, tt.fp, tt.secret); got != tt.want {
			t.Errorf("Match(%q, %q, %q) = %q, want %q", tt.path, tt.fp, tt.secret, got, tt.want)
		}
	}
}

func TestNewRejectsInvalidEntries(t *testing.T) {
	if _, err := New(rules.Allowlist{Paths: map[string]string{"a/[": "x"}}); err == nil {
		t.Error("invalid glob: want an error")
	}
	if _, err := New(rules.Allowlist{Regexes: map[string]string{"x(": "x"}}); err == nil {
		t.Error("invalid regex: want an error")
	}
}

func TestInline(t *testing.T) {
	if !Inline([]byte(`key = "abc" # whac:ignore`)) || Inline([]byte(`key = "abc"`)) {
		t.Error("Inline")
	}
}
