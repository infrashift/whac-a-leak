package source

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// Environment renders environment entries ("NAME=value") as dotenv lines,
// sorted by name. names[i] is the variable that line i+1 belongs to; a value
// with line breaks spans several lines.
func Environment(environ []string) (content []byte, names []string) {
	sorted := slices.Clone(environ)
	slices.Sort(sorted)
	var b strings.Builder
	for _, kv := range sorted {
		name, _, ok := strings.Cut(kv, "=")
		if !ok || name == "" {
			continue
		}
		b.WriteString(kv)
		b.WriteByte('\n')
		for range strings.Count(kv, "\n") + 1 {
			names = append(names, name)
		}
	}
	return []byte(b.String()), names
}

// CredentialFile is a credential file found under a root.
type CredentialFile struct {
	// ID names the config.credentialPaths entry.
	ID string
	// Path is the display path: "~/…" under the home directory.
	Path string
	// Rel is the path relative to its root, with slashes.
	Rel     string
	Abs     string
	Format  string
	MaxSize int64
}

// CredentialFiles resolves the enabled credential paths for goos under each
// root. home is the user's home directory; it is shown as "~". A root that
// does not exist is skipped. Symlinks are followed, so a credential file that
// a dotfile manager links in is still found; only regular files count.
func CredentialFiles(roots []string, home string, paths map[string]rules.CredentialPath, goos string) ([]CredentialFile, error) {
	ids := make([]string, 0, len(paths))
	for id := range paths {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	var out []CredentialFile
	seen := map[string]bool{}
	for _, root := range roots {
		if _, err := os.Stat(root); errors.Is(err, fs.ErrNotExist) {
			continue
		}
		fsys := os.DirFS(root)
		for _, id := range ids {
			cp := paths[id]
			if !cp.Enabled || (cp.OS != "" && cp.OS != goos) {
				continue
			}
			matches, err := doublestar.Glob(fsys, cp.Glob, doublestar.WithFailOnIOErrors())
			if err != nil {
				return nil, fmt.Errorf("credential path %s: %w", id, err)
			}
			slices.Sort(matches)
			for _, rel := range matches {
				abs := filepath.Join(root, filepath.FromSlash(rel))
				if seen[abs] {
					continue
				}
				if fi, err := os.Stat(abs); err != nil || !fi.Mode().IsRegular() {
					continue
				}
				seen[abs] = true
				display := filepath.ToSlash(abs)
				if root == home {
					display = "~/" + rel
				}
				out = append(out, CredentialFile{ID: id, Path: display, Rel: rel, Abs: abs, Format: cp.Format, MaxSize: cp.MaxSize})
			}
		}
	}
	return out, nil
}
