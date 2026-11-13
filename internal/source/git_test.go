package source

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// isolateGit keeps the user's global and system git config (signing, hooks,
// prefixes) out of the test repositories.
func isolateGit(t *testing.T) {
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
}

// newRepo creates an isolated git repository and returns a helper that runs
// git in it.
func newRepo(t *testing.T) (string, func(args ...string)) {
	t.Helper()
	isolateGit(t)
	dir := t.TempDir()
	g := func(args ...string) {
		t.Helper()
		cmd := exec.CommandContext(t.Context(), "git", append([]string{"-C", dir}, args...)...) //nolint:gosec // G204: test helper with fixed git subcommands
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	g("init", "-q", "-b", "main")
	return dir, g
}

func writeFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	p := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestStaged(t *testing.T) {
	dir, g := newRepo(t)
	writeFile(t, dir, "app.env", "A=1\nB=2\n")
	g("add", ".")
	g("commit", "-q", "-m", "first")
	writeFile(t, dir, "app.env", "A=1\nNEW=staged\nB=2\n")
	writeFile(t, dir, "new.txt", "hello\n")
	g("add", ".")
	writeFile(t, dir, "app.env", "A=1\nNEW=staged\nB=2\nLATER=unstaged\n")

	root, err := RepoRoot(context.Background(), dir)
	if err != nil {
		t.Fatal(err)
	}
	files, err := Staged(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0].Path != "app.env" || files[1].Path != "new.txt" {
		t.Fatalf("files = %+v", files)
	}
	app := files[0]
	if strings.Contains(string(app.Content), "LATER") {
		t.Error("staged content includes an unstaged change")
	}
	if len(app.Added) != 1 || app.Added[0].Line != 2 || string(app.Added[0].Text) != "NEW=staged" {
		t.Errorf("added = %+v", app.Added)
	}
}

func TestHistory(t *testing.T) {
	dir, g := newRepo(t)
	writeFile(t, dir, "a.env", "TOKEN=one\n")
	g("add", ".")
	g("commit", "-q", "-m", "first")
	g("mv", "a.env", "b.env")
	g("commit", "-q", "-m", "rename only")
	writeFile(t, dir, "b.env", "TOKEN=one\nOTHER=two\n")
	writeFile(t, dir, "bin.dat", "x\x00y")
	g("add", ".")
	g("commit", "-q", "-m", "third")

	var got []FileDiff
	if err := History(context.Background(), dir, "HEAD", func(d FileDiff) error { got = append(got, d); return nil }); err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %+v; want a.env in commit 0 and b.env in commit 2 (the rename adds nothing, binaries are skipped)", got)
	}
	if got[0].Path != "a.env" || got[0].Commit.Seq != 0 || got[0].Commit.Author != "Ada" || len(got[0].Commit.Hash) != 40 {
		t.Errorf("first = %+v", got[0])
	}
	if got[1].Path != "b.env" || got[1].Commit.Seq != 2 || len(got[1].Added) != 1 || got[1].Added[0].Line != 2 {
		t.Errorf("third = %+v", got[1])
	}
	if err := History(context.Background(), dir, "--output=/tmp/x", func(FileDiff) error { return nil }); err == nil {
		t.Error("an option-like revision must be rejected")
	}
	if err := History(context.Background(), dir, "no-such-branch", func(FileDiff) error { return nil }); err == nil {
		t.Error("an unknown revision must fail")
	}
}

func TestRepoRootOutsideGit(t *testing.T) {
	isolateGit(t)
	if _, err := RepoRoot(context.Background(), t.TempDir()); err == nil {
		t.Error("want an error outside a git work tree")
	}
}
