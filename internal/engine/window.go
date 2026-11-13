package engine

import (
	"bytes"
	"slices"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// multilineSpan bounds the window of a multiline rule.
const multilineSpan = 64 << 10

// span is a byte range [start, end) of content.
type span struct{ start, end int }

// asciiLower folds A-Z only, so byte offsets stay valid. The Aho-Corasick
// prefilter folds case the same way.
func asciiLower(b []byte) []byte {
	out := make([]byte, len(b))
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		out[i] = c
	}
	return out
}

// keywordHits finds every occurrence of the rule's keywords. It is the
// single-rule form of the Scanner's prefilter.
func keywordHits(r *rules.Rule, content []byte) []span {
	lower := asciiLower(content)
	var hits []span
	for _, k := range r.Keywords {
		kb := []byte(k)
		for i := 0; i <= len(lower)-len(kb); {
			j := bytes.Index(lower[i:], kb)
			if j < 0 {
				break
			}
			hits = append(hits, span{i + j, i + j + len(kb)})
			i += j + 1
		}
	}
	return hits
}

// lineWindows turns keyword hits into the merged windows that a rule's regex
// runs on. A hit covers its whole line; for a multiline rule it covers
// multilineSpan bytes from the start of its line, rounded up to a line end.
func lineWindows(content []byte, hits []span, multiline bool) []span {
	slices.SortFunc(hits, func(a, b span) int { return a.start - b.start })
	var out []span
	for _, h := range hits {
		start := bytes.LastIndexByte(content[:h.start], '\n') + 1
		limit := h.end
		if multiline {
			limit = max(limit, min(len(content), start+multilineSpan))
		}
		end := len(content)
		if i := bytes.IndexByte(content[limit:], '\n'); i >= 0 {
			end = limit + i
		}
		if n := len(out); n > 0 && start <= out[n-1].end {
			out[n-1].end = max(out[n-1].end, end)
			continue
		}
		out = append(out, span{start, end})
	}
	return out
}

// evalRule runs the rule's regex and filters on each window around the
// keyword hits and returns matches with offsets into content.
func evalRule(r *rules.Rule, content []byte, hits []span, t *trace) []Match {
	var out []Match
	for _, w := range lineWindows(content, hits, r.Multiline) {
		for _, m := range matchRule(r, content[w.start:w.end], t) {
			m.Start += w.start
			m.End += w.start
			out = append(out, m)
		}
	}
	return out
}
