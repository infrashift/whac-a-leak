package report

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

func result() *scan.Result {
	return &scan.Result{
		Findings: []scan.Finding{{
			RuleID: "GITHUB_TOKEN", Description: "GitHub token", Severity: rules.SeverityCritical,
			Path: "a.env", Line: 2, Column: 10, EndLine: 2, Fingerprint: "whac1:ab",
			Secret: "ghp_SECRETVALUE1234", Redacted: "ghp_****1234",
		}},
		Stats: scan.Stats{Files: 3, Bytes: 2048, Binary: 1, Suppressed: 2},
	}
}

func TestText(t *testing.T) {
	var b bytes.Buffer
	if err := Text(&b, result(), Options{}); err != nil {
		t.Fatal(err)
	}
	out := b.String()
	for _, want := range []string{"CRITICAL  GITHUB_TOKEN  a.env:2:10", "secret: ghp_****1234", "fingerprint: whac1:ab",
		"1 finding(s): 1 CRITICAL. Scanned 3 file(s), 2.0 KiB. Skipped 1: 1 binary", "Suppressed 2."} {
		if !strings.Contains(out, want) {
			t.Errorf("text lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "SECRETVALUE") || strings.Contains(out, "\x1b[") {
		t.Errorf("text leaks the secret or uses color:\n%s", out)
	}
	b.Reset()
	if err := Text(&b, result(), Options{NoRedact: true, Color: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "ghp_SECRETVALUE1234") || !strings.Contains(b.String(), "\x1b[1;31m") {
		t.Errorf("NoRedact/Color:\n%s", b.String())
	}
}

func TestJSON(t *testing.T) {
	var b bytes.Buffer
	if err := JSON(&b, result(), Options{}); err != nil {
		t.Fatal(err)
	}
	var rep map[string]any
	if err := json.Unmarshal(b.Bytes(), &rep); err != nil {
		t.Fatal(err)
	}
	if rep["apiVersion"] != APIVersion || rep["kind"] != Kind {
		t.Errorf("envelope = %v", rep)
	}
	list, ok := rep["findings"].([]any)
	if !ok || len(list) != 1 {
		t.Fatalf("findings = %v", rep["findings"])
	}
	f, ok := list[0].(map[string]any)
	if !ok {
		t.Fatalf("finding = %v", list[0])
	}
	if f["secret"] != "ghp_****1234" || f["redacted"] != true || f["severity"] != "CRITICAL" {
		t.Errorf("finding = %v", f)
	}
	b.Reset()
	if err := JSON(&b, &scan.Result{}, Options{}); err != nil || !strings.Contains(b.String(), `"findings": []`) {
		t.Errorf("empty report must hold an empty array: %s", b.String())
	}
}

func TestSummaryAndBytes(t *testing.T) {
	if got := Summary(&scan.Result{Stats: scan.Stats{Files: 1, Bytes: 10}}); got != "No findings. Scanned 1 file(s), 10 B." {
		t.Errorf("Summary = %q", got)
	}
	for n, want := range map[int64]string{0: "0 B", 1023: "1023 B", 1024: "1.0 KiB", 5 << 20: "5.0 MiB"} {
		if got := humanBytes(n); got != want {
			t.Errorf("humanBytes(%d) = %q, want %q", n, got, want)
		}
	}
}
