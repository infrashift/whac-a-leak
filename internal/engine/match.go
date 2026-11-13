package engine

import (
	"bytes"
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
}

// trace records why a rule produced nothing, for `whac rules verify`.
// A nil *trace records nothing.
type trace struct {
	keywordHit bool
	regexHits  int
	rejections []string
}

func (t *trace) reject(reason string) {
	if t != nil && !slices.Contains(t.rejections, reason) {
		t.rejections = append(t.rejections, reason)
	}
}

// explain describes the furthest stage a sample reached without a finding.
// It never quotes the sample.
func (t *trace) explain() string {
	switch {
	case !t.keywordHit:
		return "no keyword appears in the sample"
	case t.regexHits == 0:
		return "the regex does not match"
	}
	return "every regex match was rejected: " + strings.Join(t.rejections, "; ")
}

// keywordHit reports whether any of the rule's keywords occurs in lower, the
// lowercased content. It is the per-rule form of the prefilter.
func keywordHit(r *rules.Rule, lower []byte) bool {
	for _, k := range r.Keywords {
		if bytes.Contains(lower, []byte(k)) {
			return true
		}
	}
	return false
}

// MatchRule evaluates one pattern rule over content: keyword gate, regex,
// secret group, entropy, stopwords and allowRegexes.
func MatchRule(r *rules.Rule, content []byte) []Match {
	return matchRule(r, content, bytes.ToLower(content), nil)
}

// matchRule runs every stage after the prefilter. lower is the lowercased
// content, computed once per chunk by the caller.
func matchRule(r *rules.Rule, content, lower []byte, t *trace) []Match {
	if r.Kind != rules.KindPattern || !keywordHit(r, lower) {
		return nil
	}
	if t != nil {
		t.keywordHit = true
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
		if reason := rejectSecret(r, secret); reason != "" {
			t.reject(reason)
			continue
		}
		out = append(out, Match{Rule: r, Start: start, End: end, Secret: secret})
	}
	return out
}

// rejectSecret returns why a candidate secret is dropped, or "".
func rejectSecret(r *rules.Rule, secret string) string {
	if r.Entropy != nil {
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
