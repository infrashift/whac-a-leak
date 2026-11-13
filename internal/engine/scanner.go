package engine

import (
	"slices"

	ahocorasick "github.com/petar-dambovaliev/aho-corasick"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// Scanner runs a set of pattern rules over content. One Aho-Corasick pass
// over the content selects the candidate rules; only those run their regex
// and filters (ADR-0003). A Scanner is safe for concurrent use.
type Scanner struct {
	rules []*rules.Rule
	ac    ahocorasick.AhoCorasick
	// byKeyword maps an automaton pattern index to the rules that list it.
	byKeyword [][]int
}

// NewScanner builds the prefilter for the enabled pattern rules in rs.
func NewScanner(rs []*rules.Rule) *Scanner {
	s := &Scanner{}
	index := map[string]int{}
	var keywords []string
	for _, r := range rs {
		if !r.Enabled || r.Kind != rules.KindPattern {
			continue
		}
		ri := len(s.rules)
		s.rules = append(s.rules, r)
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

// Scan returns every match in content, ordered by offset and then rule ID.
func (s *Scanner) Scan(content []byte) []Match {
	if len(s.rules) == 0 {
		return nil
	}
	// hits[ri] collects every keyword occurrence of rule ri.
	hits := make([][]span, len(s.rules))
	it := s.ac.IterOverlappingByte(content)
	for m := it.Next(); m != nil; m = it.Next() {
		h := span{m.Start(), m.End()}
		for _, ri := range s.byKeyword[m.Pattern()] {
			hits[ri] = append(hits[ri], h)
		}
	}

	var out []Match
	for ri, hs := range hits {
		if len(hs) > 0 {
			out = append(out, evalRule(s.rules[ri], content, hs, nil)...)
		}
	}
	slices.SortFunc(out, func(a, b Match) int {
		if a.Start != b.Start {
			return a.Start - b.Start
		}
		if a.Rule.ID < b.Rule.ID {
			return -1
		}
		if a.Rule.ID > b.Rule.ID {
			return 1
		}
		return 0
	})
	return out
}
