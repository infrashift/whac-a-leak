package scan

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

const token2 = "ghp_EXAMPLE9z8Y7x6W5v4U3t2S1r0QpOnMlKjIh" //nolint:gosec // G101: fake token with an EXAMPLE marker

func gitRepo(t *testing.T) (string, func(args ...string)) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "Ada")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "ada@example.invalid")
	}
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

func put(t *testing.T, dir, rel, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestStagedReportsOnlyAddedLines(t *testing.T) {
	dir, g := gitRepo(t)
	put(t, dir, "app.env", "OLD="+token+"\n")
	g("add", ".")
	g("commit", "-q", "-m", "old secret")
	put(t, dir, "app.env", "OLD="+token+"\nNEW="+token2+"\n")
	g("add", ".")
	put(t, dir, "app.env", "OLD="+token+"\nNEW="+token2+"\nUNSTAGED="+token2+"x\n")

	res, err := pipeline(t, nil).Staged(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 || res.Findings[0].Line != 2 || res.Findings[0].Secret != token2 || res.Findings[0].Path != "app.env" {
		t.Errorf("findings = %+v; want only the staged NEW line", res.Findings)
	}
}

func TestHistoryAttributesEarliestCommit(t *testing.T) {
	dir, g := gitRepo(t)
	put(t, dir, "a.env", "A=1\nT="+token+"\n")
	g("add", ".")
	g("commit", "-q", "-m", "add secret")
	put(t, dir, "a.env", "A=2\nT="+token+"\n") // the secret line is unchanged
	g("commit", "-q", "-am", "unrelated edit")
	put(t, dir, "a.env", "A=2\n")
	g("commit", "-q", "-am", "remove secret")
	put(t, dir, "a.env", "A=2\nT="+token+"\n")
	g("commit", "-q", "-am", "add it again")

	res, err := pipeline(t, nil).History(context.Background(), dir, "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Findings) != 1 {
		t.Fatalf("findings = %+v; want one finding for one secret in one path", res.Findings)
	}
	f := res.Findings[0]
	first := gitOut(t, dir, "rev-list", "--max-parents=0", "HEAD")
	if f.Commit != first || f.Author != "Ada" || f.Line != 2 || f.Date == "" {
		t.Errorf("finding = %+v; want the first commit %s at line 2", f, first)
	}
	if res.Stats.Commits != 3 {
		t.Errorf("commits with added lines = %d, want 3", res.Stats.Commits)
	}
}

func TestHistoryErrors(t *testing.T) {
	dir, _ := gitRepo(t)
	if _, err := pipeline(t, nil).History(context.Background(), dir, "HEAD"); err == nil {
		t.Error("a repository without commits: want an error")
	}
	if _, err := pipeline(t, nil).Staged(context.Background(), t.TempDir()); err == nil {
		t.Error("outside a repository: want an error")
	}
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...).Output() //nolint:gosec // G204: test helper
	if err != nil {
		t.Fatal(err)
	}
	return string(out[:len(out)-1])
}
