// Package source lists the content that whac scans.
package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// File is one file to scan.
type File struct {
	// Path is the display path: the scan root joined with Rel, with slashes.
	Path string
	// Rel is the path relative to the scan root, with slashes. Exclude and
	// allowlist globs match against it.
	Rel string
	// Abs is the path to open.
	Abs string
}

// Files lists the regular files under root. Inside a git work tree it uses
// `git ls-files -co --exclude-standard`, so .gitignore applies; elsewhere it
// walks the directory and skips .git. Symlinks are never followed. A root
// that is a file yields that file.
func Files(ctx context.Context, root string) ([]File, error) {
	info, err := os.Lstat(root)
	if err != nil {
		return nil, err
	}
	if info.Mode().IsRegular() {
		return []File{{Path: filepath.ToSlash(root), Rel: filepath.Base(root), Abs: root}}, nil
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%s is not a regular file or directory", root)
	}

	rels, err := gitFiles(ctx, root)
	if errors.Is(err, errNotGit) {
		rels, err = walkFiles(root)
	}
	if err != nil {
		return nil, err
	}
	slices.Sort(rels)
	files := make([]File, 0, len(rels))
	for _, rel := range rels {
		abs := filepath.Join(root, filepath.FromSlash(rel))
		fi, err := os.Lstat(abs)
		if err != nil || !fi.Mode().IsRegular() {
			continue // deleted from the work tree, a symlink, or a submodule
		}
		files = append(files, File{Path: filepath.ToSlash(filepath.Join(root, filepath.FromSlash(rel))), Rel: rel, Abs: abs})
	}
	return files, nil
}

var errNotGit = errors.New("not a git work tree")

func gitFiles(ctx context.Context, root string) ([]string, error) {
	if _, err := exec.LookPath("git"); err != nil {
		return nil, errNotGit
	}
	// G204: no shell is involved; root is passed to git as one argument.
	probe := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "--is-inside-work-tree") //nolint:gosec // G204: see above
	if out, err := probe.Output(); err != nil || strings.TrimSpace(string(out)) != "true" {
		return nil, errNotGit
	}
	cmd := exec.CommandContext(ctx, "git", "-C", root, "ls-files", "-z", "--cached", "--others", "--exclude-standard") //nolint:gosec // G204: no shell; root is one argument
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var rels []string
	for _, p := range bytes.Split(out, []byte{0}) {
		if len(p) > 0 {
			rels = append(rels, string(p))
		}
	}
	return slices.Compact(rels), nil
}

func walkFiles(root string) ([]string, error) {
	var rels []string
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		rels = append(rels, filepath.ToSlash(rel))
		return nil
	})
	return rels, err
}
