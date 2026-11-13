package scan

import (
	"bytes"
	"context"
	"runtime"
	"sync"

	"github.com/infrashift/whac-a-leak/internal/source"
)

// gate applies the exclude, size and binary gates to one file version. It
// returns false when the content is skipped.
func (p *Pipeline) gate(rel string, content []byte, st *Stats) bool {
	switch {
	case p.Excluded(rel):
		st.Excluded++
	case int64(len(content)) > p.Rules.Config.MaxFileSize:
		st.Large++
	case IsBinary(content):
		st.Binary++
	default:
		st.Files++
		st.Bytes += int64(len(content))
		return true
	}
	return false
}

// Staged scans the staged version of each changed file in the repository
// that holds dir, and reports only findings on the lines the commit adds.
// Paths are relative to the repository root.
func (p *Pipeline) Staged(ctx context.Context, dir string) (*Result, error) {
	root, err := source.RepoRoot(ctx, dir)
	if err != nil {
		return nil, err
	}
	files, err := source.Staged(ctx, root)
	if err != nil {
		return nil, err
	}
	var res Result
	for _, f := range files {
		if !p.gate(f.Path, f.Content, &res.Stats) {
			continue
		}
		added := map[int]bool{}
		for _, a := range f.Added {
			added[a.Line] = true
		}
		findings, suppressed := p.Content(f.Path, f.Path, f.Content)
		res.Stats.Suppressed += suppressed
		for i := range findings {
			if touches(added, findings[i].Line, findings[i].EndLine) {
				res.Findings = append(res.Findings, findings[i])
			}
		}
	}
	Sort(res.Findings)
	return &res, nil
}

// touches reports whether any line in [start, end] is an added line.
func touches(added map[int]bool, start, end int) bool {
	for l := start; l <= end; l++ {
		if added[l] {
			return true
		}
	}
	return false
}

// History scans the lines that each commit in rev adds. Hunks are partial
// files, so whac scans them as unstructured text (ADR-0005). A secret that
// several commits add to one path is reported once, at its earliest commit.
func (p *Pipeline) History(ctx context.Context, dir, rev string) (*Result, error) {
	root, err := source.RepoRoot(ctx, dir)
	if err != nil {
		return nil, err
	}
	type key struct{ fingerprint, path string }
	var (
		mu       sync.Mutex
		res      Result
		earliest = map[key]int{} // index into res.Findings
		seqs     = map[key]int{}
		commits  = map[int]bool{}
		wg       sync.WaitGroup
	)
	jobs := make(chan source.FileDiff)
	for range runtime.GOMAXPROCS(0) {
		wg.Go(func() {
			for d := range jobs {
				findings, st := p.diff(&d)
				mu.Lock()
				res.Stats.add(st)
				commits[d.Commit.Seq] = true
				for i := range findings {
					f := &findings[i]
					k := key{f.Fingerprint, f.Path}
					if i, ok := earliest[k]; ok {
						if d.Commit.Seq < seqs[k] {
							res.Findings[i], seqs[k] = *f, d.Commit.Seq
						}
						continue
					}
					earliest[k], seqs[k] = len(res.Findings), d.Commit.Seq
					res.Findings = append(res.Findings, *f)
				}
				mu.Unlock()
			}
		})
	}
	herr := source.History(ctx, root, rev, func(d source.FileDiff) error {
		select {
		case jobs <- d:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	close(jobs)
	wg.Wait()
	if herr != nil {
		return nil, herr
	}
	res.Stats.Commits = len(commits)
	Sort(res.Findings)
	return &res, nil
}

// diff scans the added lines of one file in one commit and maps findings
// back to line numbers in that commit's version of the file.
func (p *Pipeline) diff(d *source.FileDiff) ([]Finding, Stats) {
	var (
		st    Stats
		buf   bytes.Buffer
		lines = make([]int, 0, len(d.Added))
	)
	for _, a := range d.Added {
		buf.Write(a.Text)
		buf.WriteByte('\n')
		lines = append(lines, a.Line)
	}
	content := buf.Bytes()
	if !p.gate(d.Path, content, &st) {
		return nil, st
	}
	findings, suppressed := p.content(d.Path, d.Path, "", content)
	st.Suppressed += suppressed
	for i := range findings {
		f := &findings[i]
		f.Line, f.EndLine = lines[f.Line-1], lines[f.EndLine-1]
		f.Commit, f.Author, f.Date = d.Commit.Hash, d.Commit.Author, d.Commit.Date
	}
	return findings, st
}
