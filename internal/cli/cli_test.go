package cli

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = Execute(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func layerDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(src), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// failingLayer holds a rule whose only positive sample is too short for its
// own minLen, so verification fails.
const failingLayer = `package whac
rules: ACME_API_TOKEN: {
	description: "Acme API token"
	keywords: ["acme_"]
	regex: #"\b(acme_[A-Za-z0-9]{32})\b"#
	entropy: minLen: 40
	tests: mustMatch: ["token = acme_EXAMPLE0123456789abcdefghijklmno"]
}
`

func TestVersion(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != ExitOK || !strings.HasPrefix(out, "whac ") || !strings.Contains(out, "commit") {
		t.Errorf("version: code %d, out %q", code, out)
	}
}

func TestVersionLdflags(t *testing.T) {
	old := []string{version, commit, date}
	t.Cleanup(func() { version, commit, date = old[0], old[1], old[2] })
	version, commit, date = "v1.2.3", "abc123", "2026-11-12"
	if _, out, _ := run(t, "version"); out != "whac v1.2.3 (commit abc123, built 2026-11-12)\n" {
		t.Errorf("version = %q", out)
	}
}

func TestRulesList(t *testing.T) {
	if code, _, _ := run(t, "rules", "list", "--config-dir", filepath.Join(t.TempDir(), "absent")); code != ExitError {
		t.Fatalf("an explicit --config-dir that does not exist must fail; code %d", code)
	}
	code, out, errOut := run(t, "rules", "list")
	if code != ExitOK {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	for _, id := range []string{"AWS_ACCESS_KEY_ID", "GITHUB_TOKEN", "PRIVATE_KEY", "JWT", "GENERIC_API_KEY"} {
		if !strings.Contains(out, id) {
			t.Errorf("list output lacks %s:\n%s", id, out)
		}
	}
}

func TestRulesListJSONWithLayer(t *testing.T) {
	dir := layerDir(t, map[string]string{"org.cue": "package whac\noverrides: JWT: {enabled: false, severity: \"LOW\"}\n"})
	code, out, errOut := run(t, "rules", "list", "--format", "json", "--config-dir", dir)
	if code != ExitOK {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	var list []ruleJSON
	if err := json.Unmarshal([]byte(out), &list); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, out)
	}
	for _, r := range list {
		if r.ID == "JWT" && (r.Enabled || r.Severity != rules.SeverityLow) {
			t.Errorf("override not applied: %+v", r)
		}
	}
}

func TestRulesListBadFormat(t *testing.T) {
	if code, _, errOut := run(t, "rules", "list", "--format", "xml"); code != ExitError || !strings.Contains(errOut, "--format") {
		t.Errorf("code %d, stderr %q", code, errOut)
	}
}

func TestRulesListRefusesFailingRules(t *testing.T) {
	dir := layerDir(t, map[string]string{"org.cue": failingLayer})
	code, _, errOut := run(t, "rules", "list", "--config-dir", dir)
	if code != ExitError || !strings.Contains(errOut, "ACME_API_TOKEN mustMatch[0]") {
		t.Errorf("code %d, stderr %q", code, errOut)
	}
	if strings.Contains(errOut, "acme_EXAMPLE") {
		t.Errorf("stderr leaks the sample: %q", errOut)
	}
}

func TestRulesVerify(t *testing.T) {
	code, out, errOut := run(t, "rules", "verify")
	if code != ExitOK || !strings.Contains(out, "0 failure(s)") || strings.Contains(out, "FAIL") {
		t.Errorf("code %d, out %q, stderr %q", code, out, errOut)
	}

	dir := layerDir(t, map[string]string{"org.cue": failingLayer})
	code, out, _ = run(t, "rules", "verify", "--config-dir", dir)
	if code != ExitError || !strings.Contains(out, "FAIL  ACME_API_TOKEN") || !strings.Contains(out, "minLen 40") {
		t.Errorf("code %d, out %q", code, out)
	}
}

func TestInvalidLayerExitsTwo(t *testing.T) {
	dir := layerDir(t, map[string]string{"org.cue": "package whac\nrule: X: {}\n"})
	code, _, errOut := run(t, "rules", "verify", "--config-dir", dir)
	if code != ExitError || !strings.Contains(errOut, "unknown top-level field") {
		t.Errorf("code %d, stderr %q", code, errOut)
	}
}

func TestExportSchema(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".whac")
	code, out, errOut := run(t, "rules", "export-schema", "--config-dir", dir)
	if code != ExitOK || !strings.Contains(out, "wrote") {
		t.Fatalf("code %d, out %q, stderr %q", code, out, errOut)
	}
	target := filepath.Join(dir, "schema.cue")
	data, err := fs.ReadFile(os.DirFS(dir), "schema.cue")
	if err != nil || !bytes.HasPrefix(data, []byte(rules.GeneratedHeader)) {
		t.Fatalf("schema copy missing header: %v", err)
	}

	if _, out, _ := run(t, "rules", "export-schema", "--config-dir", dir); !strings.Contains(out, "up to date") {
		t.Errorf("second export: %q", out)
	}
	// The exported copy is skipped by the loader, so the rules still load.
	if code, _, errOut := run(t, "rules", "verify", "--config-dir", dir); code != ExitOK {
		t.Errorf("verify with exported schema: code %d, stderr %q", code, errOut)
	}

	if err := os.WriteFile(target, []byte("package whac\n// mine\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if code, _, errOut := run(t, "rules", "export-schema", "--config-dir", dir); code != ExitError || !strings.Contains(errOut, "--force") {
		t.Errorf("foreign file: code %d, stderr %q", code, errOut)
	}
	if code, _, errOut := run(t, "rules", "export-schema", "--force", "--config-dir", dir); code != ExitOK {
		t.Errorf("--force: code %d, stderr %q", code, errOut)
	}
}

func TestUnknownCommand(t *testing.T) {
	if code, _, errOut := run(t, "frobnicate"); code != ExitError || !strings.Contains(errOut, "unknown command") {
		t.Errorf("code %d, stderr %q", code, errOut)
	}
}

// TestDocsDemoRuleVerifies keeps the landing-page demo rule valid: the docs
// site shows docs/src/demo/rule.cue as a working layer.
func TestDocsDemoRuleVerifies(t *testing.T) {
	code, out, errOut := run(t, "rules", "verify", "--config-dir", filepath.Join("..", "..", "docs", "src", "demo"))
	if code != ExitOK || !strings.Contains(out, "PASS  ACME_API_TOKEN") {
		t.Errorf("code %d, out %q, stderr %q", code, out, errOut)
	}
}
