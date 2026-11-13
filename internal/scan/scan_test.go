package scan

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

const token = "ghp_EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr" //nolint:gosec // G101: fake token with an EXAMPLE marker

func pipeline(t *testing.T, mutate func(*rules.Ruleset)) *Pipeline {
	t.Helper()
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if mutate != nil {
		mutate(rs)
	}
	p, err := NewPipeline(rs)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestContentPositions(t *testing.T) {
	p := pipeline(t, nil)
	content := []byte("first\n  ключ: " + token + "\n")
	fs, suppressed := p.Content("a.yaml", "a.yaml", content)
	if len(fs) != 1 || suppressed != 0 {
		t.Fatalf("findings = %+v, suppressed %d", fs, suppressed)
	}
	f := fs[0]
	if f.Line != 2 || f.Column != 9 || f.EndLine != 2 || f.RuleID != "GITHUB_TOKEN" {
		t.Errorf("finding = %+v", f)
	}
	if f.Redacted != "ghp_****oPqr" || f.Secret != token || !strings.HasPrefix(f.Fingerprint, "whac1:") {
		t.Errorf("finding = %+v", f)
	}
}

func TestContentMultilineEndLine(t *testing.T) {
	p := pipeline(t, nil)
	pem := "x\n-----BEGIN RSA PRIVATE KEY-----\n" +
		strings.Repeat("EXAMPLEKEYbodyNotARealKey0123456789abcdefghijklmn\n", 3) +
		"-----END RSA PRIVATE KEY-----\n"
	fs, _ := p.Content("k", "k", []byte(pem))
	if len(fs) != 1 || fs[0].Line != 3 || fs[0].EndLine != 5 || fs[0].Redacted != "****" {
		t.Errorf("findings = %+v", fs)
	}
}

func TestContentSuppression(t *testing.T) {
	p := pipeline(t, func(rs *rules.Ruleset) {
		rs.Allowlist.Paths["fixtures/**"] = "test fixtures"
	})
	if fs, n := p.Content("x", "x", []byte("t = "+token+" # whac:ignore\n")); len(fs) != 0 || n != 1 {
		t.Errorf("inline marker: %v, %d", fs, n)
	}
	if fs, n := p.Content("fixtures/a", "fixtures/a", []byte(token)); len(fs) != 0 || n != 1 {
		t.Errorf("path allowlist: %v, %d", fs, n)
	}
}

func TestFilesStats(t *testing.T) {
	root := t.TempDir()
	write := func(name string, data []byte) {
		if err := os.WriteFile(filepath.Join(root, name), data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write("a.env", []byte("GH="+token+"\n"))
	write("b.bin", []byte("x\x00"+token))
	write("big.txt", []byte(strings.Repeat("a", 64)))
	write("skip.log", []byte(token))
	p := pipeline(t, func(rs *rules.Ruleset) {
		rs.Config.MaxFileSize = 60
		rs.Config.Exclude["*.log"] = "logs"
	})
	res, err := p.Files(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 || res.Stats.Files != 1 || res.Stats.Binary != 1 || res.Stats.Large != 1 || res.Stats.Excluded != 1 {
		t.Errorf("result = %+v", res)
	}
}

func TestNewPipelineRejectsBadGlob(t *testing.T) {
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	rs.Config.Exclude["a/["] = "x"
	if _, err := NewPipeline(rs); err == nil {
		t.Error("want an error for an invalid exclude glob")
	}
}

func TestRedact(t *testing.T) {
	tests := map[string]string{
		"short":                    "****",
		"0123456789abcdef":         "0123****cdef",
		"line1\nline2-and-more-xx": "****",
		"ключключключключ":         "ключ****ключ",
	}
	for in, want := range tests {
		if got := Redact(in); got != want {
			t.Errorf("Redact(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFilterAndSort(t *testing.T) {
	fs := []Finding{
		{Path: "b", Line: 1, Severity: rules.SeverityLow, RuleID: "X"},
		{Path: "a", Line: 2, Severity: rules.SeverityCritical, RuleID: "Y"},
		{Path: "a", Line: 2, Column: 1, Severity: rules.SeverityHigh, RuleID: "A"},
	}
	Sort(fs)
	if fs[0].Path != "a" || fs[0].RuleID != "Y" || fs[2].Path != "b" {
		t.Errorf("sorted = %+v", fs)
	}
	if got := Filter(fs, rules.SeverityHigh); len(got) != 2 {
		t.Errorf("Filter = %+v", got)
	}
}

func TestIsBinary(t *testing.T) {
	if IsBinary([]byte("text")) || !IsBinary([]byte("a\x00b")) || IsBinary(nil) {
		t.Error("IsBinary")
	}
}
