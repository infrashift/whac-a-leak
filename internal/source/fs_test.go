package source

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func write(t *testing.T, root, rel, content string) {
	t.Helper()
	p := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(p), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func rels(files []File) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Rel
	}
	return out
}

func TestFilesWalk(t *testing.T) {
	root := t.TempDir()
	write(t, root, "a.txt", "a")
	write(t, root, "sub/b.txt", "b")
	write(t, root, ".git/config", "not scanned")
	if err := os.Symlink(filepath.Join(root, "a.txt"), filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}
	files, err := Files(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rels(files), []string{"a.txt", "sub/b.txt"}; !slices.Equal(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
	if files[1].Path != filepath.ToSlash(filepath.Join(root, "sub", "b.txt")) {
		t.Errorf("Path = %q", files[1].Path)
	}
}

func TestFilesGitRespectsGitignore(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	root := t.TempDir()
	if out, err := exec.CommandContext(t.Context(), "git", "-C", root, "init", "-q").CombinedOutput(); err != nil { //nolint:gosec // G204: fixed git arguments, temp dir path
		t.Fatalf("git init: %v %s", err, out)
	}
	write(t, root, ".gitignore", "ignored.txt\nbuild/\n")
	write(t, root, "kept.txt", "k")
	write(t, root, "ignored.txt", "i")
	write(t, root, "build/out.txt", "o")
	files, err := Files(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := rels(files), []string{".gitignore", "kept.txt"}; !slices.Equal(got, want) {
		t.Errorf("files = %v, want %v", got, want)
	}
}

func TestFilesSingleFile(t *testing.T) {
	root := t.TempDir()
	write(t, root, "one.env", "x")
	files, err := Files(context.Background(), filepath.Join(root, "one.env"))
	if err != nil || len(files) != 1 || files[0].Rel != "one.env" {
		t.Errorf("files = %v, err %v", files, err)
	}
	if _, err := Files(context.Background(), filepath.Join(root, "missing")); err == nil {
		t.Error("missing root: want an error")
	}
}
