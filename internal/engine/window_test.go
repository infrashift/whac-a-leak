package engine

import (
	"slices"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func TestLineWindows(t *testing.T) {
	content := []byte("aa\nbb key cc key\ndd\nee key\n")
	//                 0  3          15  17 20
	hits := keywordHits(&rules.Rule{Keywords: []string{"key"}}, content)
	if len(hits) != 3 {
		t.Fatalf("hits = %v", hits)
	}
	got := lineWindows(content, hits, false)
	want := []span{{3, 16}, {20, 26}}
	if !slices.Equal(got, want) {
		t.Errorf("windows = %v, want %v", got, want)
	}
	if got := lineWindows(content, hits, true); !slices.Equal(got, []span{{3, len(content)}}) {
		t.Errorf("multiline windows = %v", got)
	}
	if got := lineWindows([]byte("key"), []span{{0, 3}}, false); !slices.Equal(got, []span{{0, 3}}) {
		t.Errorf("no newline: %v", got)
	}
}

func TestKeywordHitsFoldsASCIIOnly(t *testing.T) {
	// "İ" (U+0130) lowers to a longer UTF-8 sequence; ASCII folding keeps
	// every offset valid.
	content := []byte("İİ TOK_x tok_y")
	hits := keywordHits(&rules.Rule{Keywords: []string{"tok_"}}, content)
	if len(hits) != 2 || string(content[hits[0].start:hits[0].end]) != "TOK_" {
		t.Errorf("hits = %v", hits)
	}
}

func TestMultilineRequiredForSpanningSecrets(t *testing.T) {
	r := rule(t, `(?s)(BEGIN.+?END)`, func(r *rules.Rule) {
		r.Keywords = []string{"begin"}
		r.Tests.MustMatch = []string{"BEGIN\nbody\nEND"}
	})
	fs := Verify([]*rules.Rule{r})
	if len(fs) != 1 || !strings.Contains(fs[0].Reason, "multiline: true") {
		t.Fatalf("failures = %v, want a multiline hint", fs)
	}
	r.Multiline = true
	if fs := Verify([]*rules.Rule{r}); len(fs) != 0 {
		t.Errorf("multiline rule: %v", fs)
	}
}

func TestMatchOffsetsAcrossWindows(t *testing.T) {
	content := []byte("line one\nx = tok_q9Wv7Lm2Rt\nline three tok_Abc123XyZw\n")
	ms := MatchRule(rule(t, `(tok_[A-Za-z0-9]{10})`, nil), content)
	if len(ms) != 2 {
		t.Fatalf("matches = %v", ms)
	}
	for _, m := range ms {
		if string(content[m.Start:m.End]) != m.Secret {
			t.Errorf("offset mismatch: %+v", m)
		}
	}
}
