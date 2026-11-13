package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScanOutputFiles checks that one run prints text on stdout and writes
// each --output file with the same content as --format on stdout.
func TestScanOutputFiles(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	dir := t.TempDir()
	jsonPath, sarifPath := filepath.Join(dir, "whac.json"), filepath.Join(dir, "whac.sarif")
	code, out, errOut := run(t, "scan", "--config-dir", t.TempDir(),
		"--output", "json="+jsonPath, "--output", "sarif="+sarifPath, "testdata/corpus")
	if code != ExitFindings {
		t.Fatalf("code %d, want %d; stderr %q", code, ExitFindings, errOut)
	}
	if !strings.Contains(out, "finding(s):") || strings.HasPrefix(out, "{") {
		t.Errorf("stdout is not the text report:\n%s", out)
	}
	for format, path := range map[string]string{"json": jsonPath, "sarif": sarifPath} {
		got, err := os.ReadFile(path) //nolint:gosec // G304: path in a test temp dir
		if err != nil {
			t.Fatal(err)
		}
		_, want, _ := run(t, "scan", "--format", format, "--config-dir", t.TempDir(), "testdata/corpus")
		if string(got) != want {
			t.Errorf("%s file differs from --format %s on stdout", format, format)
		}
		if !json.Valid(got) {
			t.Errorf("%s file is not valid JSON", format)
		}
		if strings.Contains(string(got), "EXAMPLE0a1B2") {
			t.Errorf("%s file leaks a secret without --no-redact", format)
		}
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm() != 0o600 {
			t.Errorf("%s file mode %v, want 0600", format, fi.Mode().Perm())
		}
	}
	var rep struct {
		Kind    string `json:"kind"`
		Summary struct {
			Findings int `json:"findings"`
		} `json:"summary"`
	}
	data, err := os.ReadFile(jsonPath) //nolint:gosec // G304: path in a test temp dir
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &rep); err != nil || rep.Kind != "ScanReport" || rep.Summary.Findings == 0 {
		t.Errorf("JSON report: err %v, kind %q, findings %d", err, rep.Kind, rep.Summary.Findings)
	}
	if left, _ := filepath.Glob(filepath.Join(dir, ".whac-report-*")); len(left) > 0 {
		t.Errorf("temporary files remain: %v", left)
	}
}

func TestScanOutputTextHasNoColorAndNoRedactApplies(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	t.Setenv("NO_COLOR", "")
	path := filepath.Join(t.TempDir(), "whac.txt")
	if code, _, errOut := run(t, "scan", "--no-redact", "--config-dir", t.TempDir(),
		"--format", "json", "--output", "text="+path, "testdata/corpus/app.env"); code != ExitFindings {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	got, err := os.ReadFile(path) //nolint:gosec // G304: path in a test temp dir
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(got), "\x1b[") {
		t.Error("the text file has ANSI colors")
	}
	if !strings.Contains(string(got), "ghp_EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr") {
		t.Errorf("--no-redact does not apply to the file:\n%s", got)
	}
}

func TestScanOutputErrors(t *testing.T) {
	t.Chdir(filepath.Join("..", ".."))
	dir := t.TempDir()
	tests := [][]string{
		{"--output", "json"},
		{"--output", "json="},
		{"--output", "xml=" + filepath.Join(dir, "a.xml")},
		{"--output", "json=" + filepath.Join(dir, "a.json"), "--output", "sarif=" + dir + "/./a.json"},
		{"--output", "json=" + filepath.Join(dir, "absent", "a.json")},
	}
	for _, args := range tests {
		args = append([]string{"scan", "--config-dir", t.TempDir()}, append(args, "testdata/corpus")...)
		code, _, errOut := run(t, args...)
		if code != ExitError || !strings.Contains(errOut, "--output") {
			t.Errorf("%v: code %d, stderr %q", args, code, errOut)
		}
	}
	if entries, _ := os.ReadDir(dir); len(entries) > 0 {
		t.Errorf("a failed --output left files: %v", entries)
	}
}

func TestScanOutputAgentContext(t *testing.T) {
	fakeHome(t)
	path := filepath.Join(t.TempDir(), "whac.json")
	code, out, errOut := run(t, "scan", "--home", "--config-dir", t.TempDir(), "--output", "json="+path)
	if code != ExitFindings || !strings.Contains(out, "AWS_ACCESS_KEY_ID") {
		t.Fatalf("code %d, stderr %q, out:\n%s", code, errOut, out)
	}
	got, err := os.ReadFile(path) //nolint:gosec // G304: path in a test temp dir
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(got), `"ruleId": "AWS_ACCESS_KEY_ID"`) || strings.Contains(string(got), "AKIAIOSFODNN7EXAMPLE") {
		t.Errorf("agent-context JSON file:\n%s", got)
	}
}
