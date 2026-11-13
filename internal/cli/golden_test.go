package cli

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

var update = flag.Bool("update", false, "rewrite testdata/golden files")

// TestCorpusGolden scans testdata/corpus from the repository root and
// compares the JSON report with testdata/golden/corpus.json. Run
// `go test ./internal/cli -run Golden -update` after an intended change.
func TestCorpusGolden(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	code, out, errOut := run(t, "scan", "--format", "json", "--config-dir", t.TempDir(), "testdata/corpus")
	if code != ExitFindings {
		t.Fatalf("code %d, want %d; stderr %q", code, ExitFindings, errOut)
	}
	golden := filepath.Join("testdata", "golden", "corpus.json")
	if *update {
		if err := os.WriteFile(golden, []byte(out), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden) //nolint:gosec // G304: fixed path to a golden file
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if !bytes.Equal([]byte(out), want) {
		t.Errorf("report differs from %s; run with -update after an intended change.\ngot:\n%s", golden, out)
	}
	if strings.Contains(out, "EXAMPLE0a1B2") {
		t.Error("the JSON report leaks a secret without --no-redact")
	}
}

func TestScanNoRedactAndThreshold(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	_, out, _ := run(t, "scan", "--no-redact", "--config-dir", t.TempDir(), "testdata/corpus/app.env")
	if !strings.Contains(out, "ghp_EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr") {
		t.Errorf("--no-redact did not print the secret:\n%s", out)
	}
	code, out, _ := run(t, "scan", "--severity-threshold", "CRITICAL", "--config-dir", t.TempDir(), "testdata/corpus/notes.md")
	if code != ExitOK || !strings.Contains(out, "No findings.") {
		t.Errorf("JWT is MEDIUM, so a CRITICAL threshold reports nothing; code %d:\n%s", code, out)
	}
	if code, _, errOut := run(t, "scan", "--severity-threshold", "SEVERE", "testdata/corpus"); code != ExitError || !strings.Contains(errOut, "--severity-threshold") {
		t.Errorf("bad threshold: code %d, stderr %q", code, errOut)
	}
	if code, _, errOut := run(t, "scan", "--format", "xml", "testdata/corpus"); code != ExitError || !strings.Contains(errOut, "--format") {
		t.Errorf("bad format: code %d, stderr %q", code, errOut)
	}
	if code, _, _ := run(t, "scan", "testdata/does-not-exist"); code != ExitError {
		t.Errorf("missing path: code %d, want %d", code, ExitError)
	}
}

func TestScanLayersApply(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	dir := layerDir(t, map[string]string{"org.cue": `package whac
config: exclude: "keys/**": "fixtures"
overrides: JWT: enabled: false
allowlist: paths: "config.yaml": "fixture"
`})
	code, out, errOut := run(t, "scan", "--config-dir", dir, "testdata/corpus")
	if code != ExitFindings {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	for _, gone := range []string{"PRIVATE_KEY", "JWT", "config.yaml"} {
		if strings.Contains(out, gone) {
			t.Errorf("layer did not remove %s:\n%s", gone, out)
		}
	}
	if !strings.Contains(out, "2 excluded") || !strings.Contains(out, "app.env") {
		t.Errorf("unexpected report:\n%s", out)
	}
}

func TestAllowAdd(t *testing.T) {
	dir := filepath.Join(t.TempDir(), ".whac")
	fp := "whac1:" + strings.Repeat("0f", 32)
	code, out, errOut := run(t, "allow", "add", fp, "--reason", `fixture "quoted"`, "--config-dir", dir)
	if code != ExitOK || !strings.Contains(out, "added") {
		t.Fatalf("code %d, out %q, stderr %q", code, out, errOut)
	}
	if _, out, _ = run(t, "allow", "add", fp, "--reason", "again", "--config-dir", dir); !strings.Contains(out, "already allowlisted") {
		t.Errorf("duplicate: %q", out)
	}
	if code, _, _ := run(t, "allow", "add", "whac1:short", "--reason", "x", "--config-dir", dir); code != ExitError {
		t.Errorf("malformed fingerprint: code %d", code)
	}
	if code, _, errOut := run(t, "allow", "add", fp, "--config-dir", dir); code != ExitError || !strings.Contains(errOut, "--reason") {
		t.Errorf("missing reason: code %d, stderr %q", code, errOut)
	}
}

func TestAllowAddRestoresOnBrokenConfig(t *testing.T) {
	// An existing allowlist.cue that only loads while it holds no fingerprint
	// entries: the new entry breaks it, so allow add must restore the file.
	broken := "package whac\nallowlist: fingerprints: close({})\n"
	dir := layerDir(t, map[string]string{"allowlist.cue": broken})
	fp := "whac1:" + strings.Repeat("0f", 32)
	if code, _, errOut := run(t, "allow", "add", fp, "--reason", "x", "--config-dir", dir); code != ExitError || !strings.Contains(errOut, "unchanged") {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	got, err := os.ReadFile(filepath.Join(dir, "allowlist.cue")) //nolint:gosec // G304: test temp dir
	if err != nil || string(got) != broken {
		t.Errorf("file not restored: %q, %v", got, err)
	}
}

// TestDocsDemoOutputIsCurrent fails when the landing-page demo output no
// longer matches what whac prints. Run `make docs-demo` to refresh it.
func TestDocsDemoOutputIsCurrent(t *testing.T) {
	t.Chdir(filepath.Join("..", "..", "docs", "src", "demo"))
	code, out, errOut := run(t, "scan", "--config-dir", ".", "app.env")
	if code != ExitFindings {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	want, err := os.ReadFile("scan-output.txt")
	if err != nil {
		t.Fatal(err)
	}
	if out != string(want) {
		t.Errorf("docs/src/demo/scan-output.txt is stale; run `make docs-demo`.\ngot:\n%s", out)
	}
}

var cueBlockRe = regexp.MustCompile("(?s)```cue\n(.*?)```")

// TestDocsCueExamplesVerify loads every complete CUE layer in the docs
// (a ```cue block with a package clause) and runs its rule samples, so a
// docs example cannot drift from the schema or the engine.
func TestDocsCueExamplesVerify(t *testing.T) {
	root := filepath.Join("..", "..", "docs", "src", "content", "docs")
	n := 0
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasPrefix(filepath.Ext(p), ".md") {
			return err
		}
		data, err := os.ReadFile(p) //nolint:gosec // G304: docs files in the repository
		if err != nil {
			return err
		}
		for i, m := range cueBlockRe.FindAllSubmatch(data, -1) {
			block := dedent(string(m[1]))
			if !strings.HasPrefix(block, "package whac") {
				continue
			}
			n++
			dir := layerDir(t, map[string]string{"example.cue": block})
			if code, out, errOut := run(t, "rules", "verify", "--config-dir", dir); code != ExitOK {
				t.Errorf("%s block %d: code %d\n%s%s", p, i, code, out, errOut)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Error("found no CUE examples in the docs")
	}
}

// dedent removes the indentation that Markdown list items add to a block.
func dedent(s string) string {
	lines := strings.Split(s, "\n")
	prefix := ""
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		lead := l[:len(l)-len(strings.TrimLeft(l, " "))]
		if prefix == "" || len(lead) < len(prefix) {
			prefix = lead
		}
	}
	for i, l := range lines {
		lines[i] = strings.TrimPrefix(l, prefix)
	}
	return strings.Join(lines, "\n")
}
