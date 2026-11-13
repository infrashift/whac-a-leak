package engine

import (
	"cmp"
	"slices"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/structural"
)

// Scanner runs a set of rules over content. One Aho-Corasick pass over the
// content selects the candidate rules; only those run their regex and
// filters on the lines around their keywords (ADR-0003, ADR-0013).
// Assignment rules run on the key/value pairs of structured files, and on
// unstructured text around their keywords (ADR-0005). A Scanner is safe for
// concurrent use.
type Scanner struct {
	rules []*rules.Rule
	ac    ahocorasick.AhoCorasick
	// byKeyword maps an automaton pattern index to the rules that list it.
	byKeyword [][]int
	assign    bool // at least one assignment rule
}

// NewScanner builds the prefilter for the enabled rules in rs.
func NewScanner(rs []*rules.Rule) *Scanner { return newScanner(rs, false) }

func newScanner(rs []*rules.Rule, includeDisabled bool) *Scanner {
	s := &Scanner{}
	index := map[string]int{}
	var keywords []string
	for _, r := range rs {
		if !r.Enabled && !includeDisabled {
			continue
		}
		ri := len(s.rules)
		s.rules = append(s.rules, r)
		s.assign = s.assign || r.Kind == rules.KindAssignment
		for _, k := range r.Keywords {
			ki, ok := index[k]
			if !ok {
				ki = len(keywords)
				index[k] = ki
				keywords = append(keywords, k)
				s.byKeyword = append(s.byKeyword, nil)
			}
			if !slices.Contains(s.byKeyword[ki], ri) {
				s.byKeyword[ki] = append(s.byKeyword[ki], ri)
			}
		}
	}
	// StandardMatch with overlapping iteration reports every keyword
	// occurrence. Leftmost match kinds skip a keyword that overlaps a longer
	// one, which would hide a candidate rule.
	b := ahocorasick.NewAhoCorasickBuilder(ahocorasick.Opts{
		AsciiCaseInsensitive: true,
		MatchKind:            ahocorasick.StandardMatch,
		DFA:                  true,
	})
	s.ac = b.Build(keywords)
	return s
}

// Rules returns the rules the scanner runs.
func (s *Scanner) Rules() []*rules.Rule { return s.rules }

// Scan returns the matches in content. name is the file name; it selects the
// structural extractor. Matches are ordered by offset. When two rules report
// the same span, only the higher severity (then the lower rule ID) remains.
func (s *Scanner) Scan(name string, content []byte) []Match {
	return s.scan(name, content, nil)
}

func (s *Scanner) scan(name string, content []byte, t *trace) []Match {
	if len(s.rules) == 0 {
		return nil
	}
	// hits[ri] collects every keyword occurrence of rule ri.
	hits := make([][]span, len(s.rules))
	if len(s.byKeyword) > 0 {
		it := s.ac.IterOverlappingByte(content)
		for m := it.Next(); m != nil; m = it.Next() {
			h := span{m.Start(), m.End()}
			for _, ri := range s.byKeyword[m.Pattern()] {
				hits[ri] = append(hits[ri], h)
			}
		}
	}

	var (
		out        []Match
		assigns    []structural.Assignment
		structured bool
	)
	if s.assign {
		if f := structural.Detect(name); f != structural.FormatNone {
			if a, err := structural.Extract(f, content); err == nil {
				assigns, structured = a, true
			}
		}
	}
	for ri, r := range s.rules {
		if t != nil {
			t.assignment = r.Kind == rules.KindAssignment
			t.keywordHit = t.keywordHit || len(hits[ri]) > 0
			t.structured = structured
		}
		switch r.Kind {
		case rules.KindPattern:
			if len(hits[ri]) > 0 {
				out = append(out, evalRule(r, content, hits[ri], t)...)
			}
		case rules.KindAssignment:
			candidates := assigns
			if !structured {
				candidates = nil
				for _, w := range lineWindows(content, hits[ri], false) {
					candidates = append(candidates, structural.Generic(content[w.start:w.end], w.start)...)
				}
			}
			if t != nil {
				t.assignments += len(candidates)
			}
			for i := range candidates {
				if m, ok := matchAssignment(r, &candidates[i], t); ok {
					out = append(out, m)
				}
			}
		}
	}
	return dedupe(out)
}

// dedupe orders matches by offset and keeps one match per span.
func dedupe(ms []Match) []Match {
	slices.SortFunc(ms, func(a, b Match) int {
		return cmp.Or(
			cmp.Compare(a.Start, b.Start),
			cmp.Compare(a.End, b.End),
			cmp.Compare(b.Rule.Severity, a.Rule.Severity),
			cmp.Compare(a.Rule.ID, b.Rule.ID),
		)
	})
	return slices.CompactFunc(ms, func(a, b Match) bool { return a.Start == b.Start && a.End == b.End })
}
