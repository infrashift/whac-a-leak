package engine

import (
	"strings"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/structural"
)

// keyPathForGlob turns a dotted key path into the lowercased,
// slash-separated form that rules.PathGlob produces for globs. A "/" inside
// a segment is escaped so it cannot act as a separator.
func keyPathForGlob(dotted string) string {
	return strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(dotted, "/", "%2f"), ".", "/"))
}

// selects reports whether an assignment rule applies to a key. A rule that
// sets both keyContext and paths needs both to match.
func selects(r *rules.Rule, a *structural.Assignment) bool {
	if r.KeyContext != nil && !r.KeyContext.MatchString(a.Key) {
		return false
	}
	if len(r.PathGlobs) > 0 {
		p := keyPathForGlob(a.KeyPath)
		for _, g := range r.PathGlobs {
			if doublestar.MatchUnvalidated(g, p) {
				return true
			}
		}
		return false
	}
	return true
}

// matchAssignment evaluates an assignment rule on one key/value pair:
// selector, optional value regex, placeholder filter, then the same secret
// checks as a pattern rule.
func matchAssignment(r *rules.Rule, a *structural.Assignment, t *trace) (Match, bool) {
	if !selects(r, a) {
		return Match{}, false
	}
	if t != nil {
		t.selected++
	}
	secret, start := a.Value, a.Start
	if r.Regex != nil {
		loc := r.Regex.FindStringSubmatchIndex(a.Value)
		if loc == nil {
			t.reject("the value does not match the regex")
			return Match{}, false
		}
		s, e := loc[2*r.SecretGroup], loc[2*r.SecretGroup+1]
		if s < 0 || s == e {
			t.reject("secret group is empty")
			return Match{}, false
		}
		secret, start = a.Value[s:e], a.Start+s
	}
	if reason := placeholder(secret); reason != "" {
		t.reject(reason)
		return Match{}, false
	}
	if reason := rejectSecret(r, secret, a.Key); reason != "" {
		t.reject(reason)
		return Match{}, false
	}
	return Match{Rule: r, Start: start, End: start + len(secret), Secret: secret, KeyPath: a.KeyPath}, true
}
