package source

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// gitArgs pins the options that user configuration could change, so the
// output always has the shape the parser expects.
var gitArgs = []string{
	"-c", "core.quotePath=false",
	"-c", "log.showSignature=false",
	"-c", "diff.noprefix=false",
	"-c", "color.ui=never",
}

var diffArgs = []string{"-U0", "--no-color", "--no-ext-diff", "--no-textconv", "--no-relative", "--src-prefix=a/", "--dst-prefix=b/", "-M"}

func gitCmd(ctx context.Context, dir string, args ...string) *exec.Cmd {
	full := append(append([]string{"-C", dir}, gitArgs...), args...)
	return exec.CommandContext(ctx, "git", full...) //nolint:gosec // G204: no shell; fixed git subcommands, dir is one argument
}

func gitOutput(ctx context.Context, dir string, args ...string) ([]byte, error) {
	cmd := gitCmd(ctx, dir, args...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", args[0], err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

// RepoRoot returns the top directory of the git work tree that holds dir.
func RepoRoot(ctx context.Context, dir string) (string, error) {
	out, err := gitOutput(ctx, dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", fmt.Errorf("%s is not inside a git work tree: %w", dir, err)
	}
	return strings.TrimSpace(string(out)), nil
}

// StagedFile is a file with staged changes: its full staged content and the
// lines that the staged diff adds.
type StagedFile struct {
	Path    string
	Content []byte
	Added   []AddedLine
}

// Staged lists the files that `git diff --cached` adds lines to, with their
// staged (index) content. root is the repository root.
func Staged(ctx context.Context, root string) ([]StagedFile, error) {
	cmd := gitCmd(ctx, root, append([]string{"diff", "--cached"}, diffArgs...)...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	var diffs []FileDiff
	perr := parseDiff(out, func(d FileDiff) error { diffs = append(diffs, d); return nil })
	if err := cmd.Wait(); err != nil {
		return nil, fmt.Errorf("git diff --cached: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	if perr != nil {
		return nil, perr
	}
	files := make([]StagedFile, 0, len(diffs))
	for _, d := range diffs {
		content, err := gitOutput(ctx, root, "cat-file", "blob", ":"+d.Path)
		if err != nil {
			return nil, err
		}
		files = append(files, StagedFile{Path: d.Path, Content: content, Added: d.Added})
	}
	return files, nil
}

// History streams the added lines of each file in each commit of rev (for
// example "HEAD" or "main..feature"), oldest commit first. fn runs once per
// file per commit. Merge commits add no lines of their own.
func History(ctx context.Context, root, rev string, fn func(FileDiff) error) error {
	if strings.HasPrefix(rev, "-") {
		return fmt.Errorf("invalid revision range %q", rev)
	}
	args := append([]string{"log", "-p", "--reverse", commitFormat}, diffArgs...)
	args = append(args, rev, "--")
	cmd := gitCmd(ctx, root, args...)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	perr := parseDiff(out, fn)
	if perr != nil {
		// Stop git; its remaining output is no longer read.
		_, _ = io.Copy(io.Discard, out)
	}
	werr := cmd.Wait()
	if perr != nil {
		return perr
	}
	if werr != nil {
		var exitErr *exec.ExitError
		if errors.As(werr, &exitErr) {
			return fmt.Errorf("git log %s: %s", rev, strings.TrimSpace(stderr.String()))
		}
		return werr
	}
	return nil
}
