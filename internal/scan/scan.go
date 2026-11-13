// Package scan runs the scan pipeline: source → engine → allowlist →
// findings. Every source (files, staged changes, history) feeds Content.
package scan

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"slices"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/infrashift/whac-a-leak/internal/allow"
	"github.com/infrashift/whac-a-leak/internal/engine"
	"github.com/infrashift/whac-a-leak/internal/fingerprint"
	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/source"
)

// binarySniffLen is how many leading bytes are checked for a NUL byte.
const binarySniffLen = 8000

// Finding is one reported secret.
type Finding struct {
	RuleID      string
	Description string
	Severity    rules.Severity
	Path        string
	// Line and Column are 1-based; Column counts runes. EndLine differs from
	// Line for a multi-line secret such as a PEM key.
	Line, Column, EndLine int
	Fingerprint           string
	// KeyPath is the key path of an assignment finding, e.g.
	// "database.credentials.password".
	KeyPath string
	// Commit, Author and Date attribute a finding from git history.
	Commit, Author, Date string
	// Secret is the matched secret. Reports print Redacted unless the user
	// asks otherwise.
	Secret   string
	Redacted string
}

// Stats counts what a scan read and skipped.
type Stats struct {
	// Commits counts the commits of a history scan.
	Commits    int   `json:"commits,omitempty"`
	Files      int   `json:"files"`
	Bytes      int64 `json:"bytes"`
	Binary     int   `json:"skippedBinary"`
	Large      int   `json:"skippedLarge"`
	Excluded   int   `json:"skippedExcluded"`
	Unreadable int   `json:"skippedUnreadable"`
	Suppressed int   `json:"suppressed"`
}

// Result is the outcome of a scan.
type Result struct {
	Findings []Finding
	Stats    Stats
}

// Pipeline holds the compiled parts shared by every source.
type Pipeline struct {
	Rules     *rules.Ruleset
	Scanner   *engine.Scanner
	Allowlist *allow.Allowlist
}

// NewPipeline compiles the scanner and allowlist of a verified ruleset.
func NewPipeline(rs *rules.Ruleset) (*Pipeline, error) {
	al, err := allow.New(rs.Allowlist)
	if err != nil {
		return nil, err
	}
	for glob := range rs.Config.Exclude {
		if !doublestar.ValidatePattern(glob) {
			return nil, fmt.Errorf("config.exclude: invalid glob %q", glob)
		}
	}
	return &Pipeline{Rules: rs, Scanner: engine.NewScanner(rs.Enabled()), Allowlist: al}, nil
}

// Excluded reports whether config.exclude covers rel.
func (p *Pipeline) Excluded(rel string) bool {
	for glob := range p.Rules.Config.Exclude {
		if doublestar.MatchUnvalidated(glob, rel) {
			return true
		}
	}
	return false
}

// IsBinary reports whether content looks binary: a NUL byte near the start.
func IsBinary(content []byte) bool {
	return bytes.IndexByte(content[:min(len(content), binarySniffLen)], 0) >= 0
}

// Content scans one file. path is the display path; rel is the path that
// allowlist globs match and that selects the structural extractor. It
// returns the findings and the number of matches the allowlist or an inline
// marker suppressed.
func (p *Pipeline) Content(path, rel string, content []byte) (findings []Finding, suppressed int) {
	return p.content(path, rel, rel, content)
}

// content scans content as the file name; an empty name scans it as
// unstructured text.
func (p *Pipeline) content(path, rel, name string, content []byte) (findings []Finding, suppressed int) {
	matches := p.Scanner.Scan(name, content)
	if len(matches) == 0 {
		return nil, 0
	}
	lines := newLineIndex(content)
	for _, m := range matches {
		line := lines.line(m.Start)
		if allow.Inline(lines.text(content, line)) {
			suppressed++
			continue
		}
		fp := fingerprint.Of(m.Rule.ID, m.Secret)
		if p.Allowlist.Match(rel, fp, m.Secret) != "" {
			suppressed++
			continue
		}
		findings = append(findings, Finding{
			RuleID:      m.Rule.ID,
			Description: m.Rule.Description,
			Severity:    m.Rule.Severity,
			Path:        path,
			Line:        line,
			Column:      lines.column(content, line, m.Start),
			EndLine:     lines.line(max(m.End-1, m.Start)),
			Fingerprint: fp,
			KeyPath:     m.KeyPath,
			Secret:      m.Secret,
			Redacted:    Redact(m.Secret),
		})
	}
	return findings, suppressed
}

// Files scans the files under root with one worker per CPU.
func (p *Pipeline) Files(ctx context.Context, root string) (*Result, error) {
	files, err := source.Files(ctx, root)
	if err != nil {
		return nil, err
	}
	var (
		mu  sync.Mutex
		res Result
		wg  sync.WaitGroup
	)
	jobs := make(chan source.File)
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for f := range jobs {
				findings, st := p.file(f)
				mu.Lock()
				res.Findings = append(res.Findings, findings...)
				res.Stats.add(st)
				mu.Unlock()
			}
		})
	}
feed:
	for _, f := range files {
		select {
		case jobs <- f:
		case <-ctx.Done():
			break feed
		}
	}
	close(jobs)
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	Sort(res.Findings)
	return &res, nil
}

// file scans one file. A file that vanished or cannot be read counts as
// unreadable; it does not stop the scan.
func (p *Pipeline) file(f source.File) ([]Finding, Stats) {
	var st Stats
	if p.Excluded(f.Rel) {
		st.Excluded++
		return nil, st
	}
	info, err := os.Lstat(f.Abs)
	if err != nil {
		st.Unreadable++
		return nil, st
	}
	if info.Size() > p.Rules.Config.MaxFileSize {
		st.Large++
		return nil, st
	}
	content, err := os.ReadFile(f.Abs)
	if err != nil {
		st.Unreadable++
		return nil, st
	}
	if IsBinary(content) {
		st.Binary++
		return nil, st
	}
	st.Files++
	st.Bytes += int64(len(content))
	findings, suppressed := p.Content(f.Path, f.Rel, content)
	st.Suppressed += suppressed
	return findings, st
}

func (s *Stats) add(o Stats) {
	s.Commits += o.Commits
	s.Files += o.Files
	s.Bytes += o.Bytes
	s.Binary += o.Binary
	s.Large += o.Large
	s.Excluded += o.Excluded
	s.Unreadable += o.Unreadable
	s.Suppressed += o.Suppressed
}

// Sort orders findings by path, line, column and rule ID.
func Sort(fs []Finding) {
	slices.SortStableFunc(fs, func(a, b Finding) int {
		switch {
		case a.Path != b.Path:
			return strings.Compare(a.Path, b.Path)
		case a.Line != b.Line:
			return a.Line - b.Line
		case a.Column != b.Column:
			return a.Column - b.Column
		}
		return strings.Compare(a.RuleID, b.RuleID)
	})
}

// Filter keeps the findings at or above min.
func Filter(fs []Finding, minSev rules.Severity) []Finding {
	out := fs[:0:0]
	for i := range fs {
		if fs[i].Severity >= minSev {
			out = append(out, fs[i])
		}
	}
	return out
}

// Redact keeps the first and last four characters of a long single-line
// secret and hides everything else.
func Redact(secret string) string {
	if strings.ContainsAny(secret, "\r\n") || utf8.RuneCountInString(secret) < 16 {
		return "****"
	}
	r := []rune(secret)
	return string(r[:4]) + "****" + string(r[len(r)-4:])
}

// lineIndex maps byte offsets to 1-based line numbers.
type lineIndex []int // start offset of each line

func newLineIndex(content []byte) lineIndex {
	idx := lineIndex{0}
	for i, b := range content {
		if b == '\n' {
			idx = append(idx, i+1)
		}
	}
	return idx
}

func (l lineIndex) line(off int) int {
	return sort.Search(len(l), func(i int) bool { return l[i] > off })
}

func (l lineIndex) text(content []byte, line int) []byte {
	start := l[line-1]
	end := len(content)
	if line < len(l) {
		end = l[line] - 1
	}
	return content[start:end]
}

func (l lineIndex) column(content []byte, line, off int) int {
	return utf8.RuneCount(content[l[line-1]:off]) + 1
}
