package engine

import (
	"math"
	"regexp"
	"strings"
	"testing"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

func ptr(f float64) *float64 { return &f }

func TestShannon(t *testing.T) {
	tests := []struct {
		in   string
		want float64
	}{
		{"", 0},
		{"aaaa", 0},
		{"ab", 1},
		{"abcd", 2},
		{"0123456789abcdef", 4},
	}
	for _, tt := range tests {
		if got := Shannon(tt.in); math.Abs(got-tt.want) > 1e-9 {
			t.Errorf("Shannon(%q) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestClassify(t *testing.T) {
	tests := []struct {
		in   string
		want rules.Charset
	}{
		{"deadBEEF0123", rules.CharsetHex},
		{"AKIAIOSFODNN7EXAMPLE", rules.CharsetAlnum},
		{"abc_def-123", rules.CharsetBase64URL},
		{"abc+def/12==", rules.CharsetBase64},
		{"pa$$word!", rules.CharsetAny},
	}
	for _, tt := range tests {
		if got := Classify(tt.in); got != tt.want {
			t.Errorf("Classify(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestMinEntropy(t *testing.T) {
	explicit := &rules.Entropy{Min: ptr(3.3)}
	if got := MinEntropy(explicit, "anything"); got != 3.3 {
		t.Errorf("explicit min = %v", got)
	}
	// 40 hex chars: the alphabet (16) bounds the reachable entropy, not the length.
	hex := strings.Repeat("0123456789abcdef", 3)[:40]
	if got, want := MinEntropy(&rules.Entropy{}, hex), defaultMinFraction*4; math.Abs(got-want) > 1e-9 {
		t.Errorf("hex default = %v, want %v", got, want)
	}
	// 20 alnum chars: the length (20) bounds it, not the alphabet (62).
	alnum := "AKIAIOSFODNN7EXAMPLE" //nolint:gosec // G101: AWS's published documentation example key, not a credential
	if got, want := MinEntropy(&rules.Entropy{}, alnum), defaultMinFraction*math.Log2(20); math.Abs(got-want) > 1e-9 {
		t.Errorf("alnum default = %v, want %v", got, want)
	}
	// A declared charset overrides classification.
	if got, want := MinEntropy(&rules.Entropy{Charset: rules.CharsetHex}, strings.Repeat("z", 40)), defaultMinFraction*4; math.Abs(got-want) > 1e-9 {
		t.Errorf("declared charset = %v, want %v", got, want)
	}
	if got := MinEntropy(&rules.Entropy{}, "a"); got != 0 {
		t.Errorf("one-char secret = %v, want 0", got)
	}
}

func TestCheckEntropy(t *testing.T) {
	e := &rules.Entropy{MinLen: 8, Max: ptr(3.5)}
	if r := checkEntropy(e, "short"); !strings.Contains(r, "minLen") {
		t.Errorf("short: %q", r)
	}
	if r := checkEntropy(e, "aaaaaaaaaaaa"); !strings.Contains(r, "below minimum") {
		t.Errorf("flat: %q", r)
	}
	if r := checkEntropy(e, "0123456789abcdefghij"); !strings.Contains(r, "above maximum") {
		t.Errorf("too random: %q", r)
	}
	if r := checkEntropy(e, "abcdefgh"); r != "" {
		t.Errorf("good: %q", r)
	}
}

func rule(t *testing.T, re string, mutate func(*rules.Rule)) *rules.Rule {
	t.Helper()
	r := &rules.Rule{
		ID:          "TEST",
		Kind:        rules.KindPattern,
		Keywords:    []string{"tok_"},
		Regex:       regexp.MustCompile(re),
		SecretGroup: 1,
	}
	if mutate != nil {
		mutate(r)
	}
	return r
}

func TestMatchRule(t *testing.T) {
	content := []byte("a = TOK_Abc123XyZw\nb = tok_q9Wv7Lm2Rt\n")
	r := rule(t, `(?i)\b(tok_[A-Za-z0-9]{10})\b`, nil)
	ms := MatchRule(r, content)
	if len(ms) != 2 {
		t.Fatalf("got %d matches, want 2: %+v", len(ms), ms)
	}
	for _, m := range ms {
		if string(content[m.Start:m.End]) != m.Secret || m.Rule != r {
			t.Errorf("match offsets do not point at the secret: %+v", m)
		}
	}
	if ms[0].Secret != "TOK_Abc123XyZw" {
		t.Errorf("keyword gate must be case-insensitive; first secret = %q", ms[0].Secret)
	}
}

func TestMatchRuleStages(t *testing.T) {
	content := []byte("x tok_aaaaaaaaaa y tok_Q9wv7Lm2Rt")
	tests := []struct {
		name   string
		mutate func(*rules.Rule)
		want   int
	}{
		{"no keyword", func(r *rules.Rule) { r.Keywords = []string{"absent"} }, 0},
		{"entropy drops the flat value", func(r *rules.Rule) { r.Entropy = &rules.Entropy{MinLen: 8} }, 1},
		{"stopword", func(r *rules.Rule) { r.Stopwords = []string{"q9wv"} }, 1},
		{"allowRegex", func(r *rules.Rule) { r.AllowRegexes = []*regexp.Regexp{regexp.MustCompile(`^tok_a+$`)} }, 1},
		{"whole match", func(r *rules.Rule) { r.SecretGroup = 0 }, 2},
		{"assignment kind is not a pattern", func(r *rules.Rule) { r.Kind = rules.KindAssignment }, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := rule(t, `(tok_[A-Za-z0-9]{10})`, tt.mutate)
			if got := len(MatchRule(r, content)); got != tt.want {
				t.Errorf("got %d matches, want %d", got, tt.want)
			}
		})
	}
}

func TestMatchRuleEmptyGroup(t *testing.T) {
	r := rule(t, `tok_(x*)`, nil)
	tr := &trace{keywordHit: true}
	if ms := matchRule(r, []byte("tok_"), tr); len(ms) != 0 {
		t.Fatalf("empty group produced %v", ms)
	}
	if !strings.Contains(tr.explain(), "secret group is empty") {
		t.Errorf("explain = %q", tr.explain())
	}
}

func TestVerifyBuiltins(t *testing.T) {
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range Verify(rs.Rules) {
		t.Errorf("builtin sample failed: %s", f)
	}
}

func TestVerifyReportsFailuresWithoutSampleText(t *testing.T) {
	const secret = "tok_SECRETSAMPLE"
	tests := []struct {
		name       string
		r          *rules.Rule
		wantSample string
		wantReason string
	}{
		{
			name: "mustMatch without keyword",
			r: rule(t, `(tok_[A-Z]+)`, func(r *rules.Rule) {
				r.Keywords = []string{"nope"}
				r.Tests.MustMatch = []rules.Sample{{Content: secret}}
			}),
			wantSample: "mustMatch[0]", wantReason: "no keyword",
		},
		{
			name:       "mustMatch regex miss",
			r:          rule(t, `(tok_[0-9]+)`, func(r *rules.Rule) { r.Tests.MustMatch = []rules.Sample{{Content: secret}} }),
			wantSample: "mustMatch[0]", wantReason: "regex does not match",
		},
		{
			name: "mustMatch rejected by entropy",
			r: rule(t, `(tok_[A-Z]+)`, func(r *rules.Rule) {
				r.Entropy = &rules.Entropy{MinLen: 64}
				r.Tests.MustMatch = []rules.Sample{{Content: secret}}
			}),
			wantSample: "mustMatch[0]", wantReason: "minLen 64",
		},
		{
			name: "mustIgnore produces a finding",
			r: rule(t, `(tok_[A-Z]+)`, func(r *rules.Rule) {
				r.Tests.MustMatch = []rules.Sample{{Content: "tok_OK"}}
				r.Tests.MustIgnore = []rules.Sample{{Content: "fine"}, {Content: secret}}
			}),
			wantSample: "mustIgnore[1]", wantReason: "expected none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := Verify([]*rules.Rule{tt.r})
			if len(fs) != 1 {
				t.Fatalf("got %d failures, want 1: %v", len(fs), fs)
			}
			f := fs[0]
			if f.Sample != tt.wantSample || !strings.Contains(f.Reason, tt.wantReason) {
				t.Errorf("failure = %s; want sample %s, reason containing %q", f, tt.wantSample, tt.wantReason)
			}
			if strings.Contains(f.String(), "SECRETSAMPLE") {
				t.Errorf("failure leaks the sample text: %s", f)
			}
		})
	}
}

func BenchmarkVerifyBuiltins(b *testing.B) {
	rs, err := rules.Load(rules.Options{})
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		Verify(rs.Rules)
	}
}
