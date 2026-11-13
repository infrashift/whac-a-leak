package engine

import (
	"bytes"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func TestScannerSelectsOnlyCandidateRules(t *testing.T) {
	a := rule(t, `(tok_[A-Za-z0-9]{10})`, func(r *rules.Rule) { r.ID = "A" })
	b := rule(t, `(key_[A-Za-z0-9]{10})`, func(r *rules.Rule) { r.ID = "B"; r.Keywords = []string{"key_"} })
	off := rule(t, `(tok_[A-Za-z0-9]{10})`, func(r *rules.Rule) { r.ID = "OFF"; r.Enabled = false })
	a.Enabled, b.Enabled = true, true
	s := NewScanner([]*rules.Rule{a, b, off})
	if len(s.Rules()) != 2 {
		t.Fatalf("Rules() = %d, want 2 (disabled rule dropped)", len(s.Rules()))
	}
	ms := s.Scan("a.txt", []byte("x key_q9Wv7Lm2Rt y tok_Abc123XyZw"))
	if len(ms) != 2 || ms[0].Rule.ID != "B" || ms[1].Rule.ID != "A" {
		t.Fatalf("matches = %+v, want B then A in offset order", ms)
	}
	if got := s.Scan("a.txt", []byte("nothing to see")); len(got) != 0 {
		t.Errorf("no keyword: %v", got)
	}
	if got := NewScanner(nil).Scan("a.txt", []byte("tok_q9Wv7Lm2Rt")); got != nil {
		t.Errorf("empty scanner: %v", got)
	}
}

// TestScannerOverlappingKeywords guards against a leftmost-longest
// prefilter, which reports "secret_key" and hides the overlapping "key_".
func TestScannerOverlappingKeywords(t *testing.T) {
	long := rule(t, `(secret_key=[a-z]+)`, func(r *rules.Rule) { r.ID = "LONG"; r.Keywords = []string{"secret_key"}; r.SecretGroup = 1 })
	short := rule(t, `(key_[A-Za-z0-9]{3})`, func(r *rules.Rule) { r.ID = "SHORT"; r.Keywords = []string{"key_"} })
	long.Enabled, short.Enabled = true, true
	ms := NewScanner([]*rules.Rule{long, short}).Scan("a.txt", []byte("secret_key_abc"))
	ids := map[string]bool{}
	for _, m := range ms {
		ids[m.Rule.ID] = true
	}
	if !ids["SHORT"] {
		t.Errorf("overlapping keyword hid rule SHORT: %+v", ms)
	}
}

// TestScannerAgreesWithSingleRules checks that the multi-rule prefilter
// finds the same spans as each rule on its own. When two rules report one
// span, the scanner keeps one match, so the test compares spans.
func TestScannerAgreesWithSingleRules(t *testing.T) {
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	s := NewScanner(rs.Enabled())
	for _, r := range rs.Enabled() {
		for _, sample := range append(append([]rules.Sample{}, r.Tests.MustMatch...), r.Tests.MustIgnore...) {
			want := map[span]bool{}
			for _, other := range rs.Enabled() {
				for _, m := range newScanner([]*rules.Rule{other}, true).Scan(sample.Name(), []byte(sample.Content)) {
					want[span{m.Start, m.End}] = true
				}
			}
			got := map[span]bool{}
			for _, m := range s.Scan(sample.Name(), []byte(sample.Content)) {
				got[span{m.Start, m.End}] = true
			}
			if len(got) != len(want) {
				t.Errorf("%s sample %q: scanner spans %v, single-rule spans %v", r.ID, sample.Name(), got, want)
			}
		}
	}
}

func TestDedupeKeepsHigherSeverity(t *testing.T) {
	low := &rules.Rule{ID: "A_LOW", Severity: rules.SeverityLow}
	high := &rules.Rule{ID: "Z_HIGH", Severity: rules.SeverityHigh}
	ms := dedupe([]Match{{Rule: low, Start: 1, End: 5}, {Rule: high, Start: 1, End: 5}, {Rule: low, Start: 7, End: 9}})
	if len(ms) != 2 || ms[0].Rule != high {
		t.Errorf("dedupe = %+v", ms)
	}
}

func TestSafeFormatUUID(t *testing.T) {
	r := rule(t, `(?i)api_key = "([^"]+)"`, func(r *rules.Rule) {
		r.Keywords = []string{"api_key"}
		r.Entropy = &rules.Entropy{MinLen: 8}
	})
	if ms := MatchRule(r, []byte(`api_key = "550e8400-e29b-41d4-a716-446655440000"`)); len(ms) != 0 {
		t.Errorf("UUID reported: %v", ms)
	}
	r.Entropy = nil // a precise-format rule keeps UUID-shaped secrets
	if ms := MatchRule(r, []byte(`api_key = "550e8400-e29b-41d4-a716-446655440000"`)); len(ms) != 1 {
		t.Errorf("rule without entropy: %v", ms)
	}
}

// corpus returns n bytes of code-like text with a keyword on about one line
// in two hundred: the keyword-sparse case the throughput target is set for.
func corpus(n int) []byte {
	rng := rand.New(rand.NewPCG(1, 2)) //nolint:gosec // G404: deterministic benchmark data, not security
	words := strings.Fields("func return if else for range var const type struct map string int error nil " +
		"value result config handler request response client server buffer index count total item list")
	var b bytes.Buffer
	for line := 0; b.Len() < n; line++ {
		indent := strings.Repeat("\t", rng.IntN(3))
		b.WriteString(indent)
		for w := 0; w < 4+rng.IntN(8); w++ {
			b.WriteString(words[rng.IntN(len(words))])
			b.WriteByte(' ')
		}
		if line%200 == 0 {
			fmt.Fprintf(&b, "token := lookup(%d)", line)
		}
		b.WriteByte('\n')
	}
	return b.Bytes()[:n]
}

func BenchmarkScannerBuiltins(b *testing.B) {
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		b.Fatal(err)
	}
	s := NewScanner(rs.Enabled())
	data := corpus(4 << 20)
	b.SetBytes(int64(len(data)))
	b.ResetTimer()
	for b.Loop() {
		s.Scan("bench.go", data)
	}
}
