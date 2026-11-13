package engine

import (
	"fmt"
	"slices"
	"strings"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// Match is one secret found by a pattern rule.
type Match struct {
	Rule *rules.Rule
	// Start and End are the byte offsets of the secret within the content.
	Start, End int
	Secret     string
	// KeyPath is set for assignment rules.
	KeyPath string
}

// trace records why a rule produced nothing, for `whac rules verify`.
// A nil *trace records nothing.
type trace struct {
	assignment  bool // the rule is an assignment rule
	keywordHit  bool
	structured  bool // the sample parsed as a structured file
	regexHits   int
	assignments int // key/value pairs the extractors produced
	selected    int // pairs the rule's keyContext or paths selected
	rejections  []string
}

func (t *trace) reject(reason string) {
	if t != nil && !slices.Contains(t.rejections, reason) {
		t.rejections = append(t.rejections, reason)
	}
}

// explain describes the furthest stage a sample reached without a finding.
// It never quotes the sample.
func (t *trace) explain() string {
	if t.assignment {
		switch {
		case !t.structured && !t.keywordHit:
			return "no keyword appears in the sample, and it is not a structured file"
		case t.assignments == 0:
			return "the extractor found no key/value pair"
		case t.selected == 0:
			return "no key matches keyContext or paths"
		}
		return "every selected value was rejected: " + strings.Join(t.rejections, "; ")
	}
	switch {
	case !t.keywordHit:
		return "no keyword appears in the sample"
	case t.regexHits == 0:
		return "the regex does not match"
	}
	return "every regex match was rejected: " + strings.Join(t.rejections, "; ")
}

// MatchRule evaluates one rule over content the way a scan of an
// unstructured text file does, whether or not the rule is enabled.
func MatchRule(r *rules.Rule, content []byte) []Match {
	return newScanner([]*rules.Rule{r}, true).scan("sample.txt", content, nil)
}

// matchRule runs the regex and filters over one window. Callers pass the
// windows around keyword hits (see evalRule).
func matchRule(r *rules.Rule, content []byte, t *trace) []Match {
	if r.Kind != rules.KindPattern {
		return nil
	}
	var out []Match
	for _, loc := range r.Regex.FindAllSubmatchIndex(content, -1) {
		if t != nil {
			t.regexHits++
		}
		start, end := loc[2*r.SecretGroup], loc[2*r.SecretGroup+1]
		if start < 0 || start == end {
			t.reject("secret group is empty")
			continue
		}
		secret := string(content[start:end])
		if reason := rejectSecret(r, secret, ""); reason != "" {
			t.reject(reason)
			continue
		}
		out = append(out, Match{Rule: r, Start: start, End: end, Secret: secret})
	}
	return out
}

// rejectSecret returns why a candidate secret is dropped, or "". key is the
// assignment key, or "" for a pattern match.
func rejectSecret(r *rules.Rule, secret, key string) string {
	if r.Entropy != nil {
		// Safe formats have high entropy but are not secrets. The exclusion
		// belongs to the entropy heuristic: a rule without an entropy block
		// matches a precise token format and keeps every match (ADR-0006).
		if reason := safeFormat(secret, key); reason != "" {
			return reason
		}
		if reason := checkEntropy(r.Entropy, secret); reason != "" {
			return reason
		}
	}
	if len(r.Stopwords) > 0 {
		lower := strings.ToLower(secret)
		for _, w := range r.Stopwords {
			if strings.Contains(lower, w) {
				return "secret contains a stopword"
			}
		}
	}
	for i, re := range r.AllowRegexes {
		if re.MatchString(secret) {
			return fmt.Sprintf("secret matches allowRegexes[%d]", i)
		}
	}
	return ""
}
