package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const fakeToken = "ghp_EXAMPLE0a1B2c3D4e5F6g7H8i9J0kLmNoPqr" //nolint:gosec // G101: fake token with an EXAMPLE marker

func gitRepo(t *testing.T) (string, func(args ...string)) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	t.Setenv("GIT_AUTHOR_NAME", "Ada")
	t.Setenv("GIT_AUTHOR_EMAIL", "ada@example.invalid")
	t.Setenv("GIT_COMMITTER_NAME", "Ada")
	t.Setenv("GIT_COMMITTER_EMAIL", "ada@example.invalid")
	dir := t.TempDir()
	g := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...) //nolint:gosec // G204: test helper
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	g("init", "-q", "-b", "main")
	return dir, g
}

func TestScanStaged(t *testing.T) {
	dir, g := gitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	g("add", ".")
	g("commit", "-q", "-m", "init")
	if code, out, _ := run(t, "scan", "--staged", "--config-dir", t.TempDir(), dir); code != ExitOK || !strings.Contains(out, "No findings.") {
		t.Errorf("nothing staged: code %d, out %q", code, out)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.env"), []byte("GH_TOKEN="+fakeToken+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	g("add", "app.env")
	code, out, errOut := run(t, "scan", "--staged", "--config-dir", t.TempDir(), dir)
	if code != ExitFindings || !strings.Contains(out, "GITHUB_TOKEN  app.env:1:10") {
		t.Errorf("staged secret: code %d, out %q, stderr %q", code, out, errOut)
	}
}

func TestScanGitJSON(t *testing.T) {
	dir, g := gitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "app.env"), []byte("GH_TOKEN="+fakeToken+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	g("add", ".")
	g("commit", "-q", "-m", "oops")
	g("rm", "-q", "app.env")
	g("commit", "-q", "-m", "remove")

	// The secret is gone from the work tree but stays in history.
	if code, _, _ := run(t, "scan", "--config-dir", t.TempDir(), dir); code != ExitOK {
		t.Errorf("work tree scan: code %d, want %d", code, ExitOK)
	}
	code, out, errOut := run(t, "scan", "--git", "--format", "json", "--config-dir", t.TempDir(), dir)
	if code != ExitFindings {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	var rep struct {
		Summary  struct{ Commits int }
		Findings []struct{ Commit, Author, Path string }
	}
	if err := json.Unmarshal([]byte(out), &rep); err != nil {
		t.Fatal(err)
	}
	if rep.Summary.Commits != 1 || len(rep.Findings) != 1 || len(rep.Findings[0].Commit) != 40 || rep.Findings[0].Author != "Ada" || rep.Findings[0].Path != "app.env" {
		t.Errorf("report = %+v", rep)
	}
	if code, _, errOut := run(t, "scan", "--git=no-such-ref", dir); code != ExitError || !strings.Contains(errOut, "no-such-ref") {
		t.Errorf("bad range: code %d, stderr %q", code, errOut)
	}
}

func TestScanGitFlagErrors(t *testing.T) {
	if code, _, errOut := run(t, "scan", "--staged", "--git"); code != ExitError || !strings.Contains(errOut, "exclude each other") {
		t.Errorf("both flags: code %d, stderr %q", code, errOut)
	}
	if code, _, errOut := run(t, "scan", "--git", "main..HEAD"); code != ExitError || !strings.Contains(errOut, "--git=<range>") {
		t.Errorf("range without =: code %d, stderr %q", code, errOut)
	}
}
