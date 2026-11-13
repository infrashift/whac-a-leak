// Package allow decides whether a finding is suppressed: by fingerprint, by
// path glob, by a regex over the secret, or by an inline marker.
package allow

import (
	"bytes"
	"fmt"
	"regexp"
	"slices"

	"github.com/bmatcuk/doublestar/v4"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// InlineMarker on the line of a finding suppresses that finding.
const InlineMarker = "whac:ignore"

type entry[T any] struct {
	key    T
	text   string
	reason string
}

// Allowlist is the compiled allowlist of a ruleset.
type Allowlist struct {
	fingerprints map[string]string
	paths        []entry[string]
	regexes      []entry[*regexp.Regexp]
}

// New compiles the allowlist entries of a ruleset.
func New(a rules.Allowlist) (*Allowlist, error) {
	al := &Allowlist{fingerprints: a.Fingerprints}
	for _, glob := range sortedKeys(a.Paths) {
		if !doublestar.ValidatePattern(glob) {
			return nil, fmt.Errorf("allowlist.paths: invalid glob %q", glob)
		}
		al.paths = append(al.paths, entry[string]{key: glob, text: glob, reason: a.Paths[glob]})
	}
	for _, re := range sortedKeys(a.Regexes) {
		c, err := regexp.Compile(re)
		if err != nil {
			return nil, fmt.Errorf("allowlist.regexes: %w", err)
		}
		al.regexes = append(al.regexes, entry[*regexp.Regexp]{key: c, text: re, reason: a.Regexes[re]})
	}
	return al, nil
}

// Match returns the allowlist entry that suppresses a finding, as
// "kind key: reason", or "" if no entry matches. path uses forward slashes.
func (a *Allowlist) Match(path, fingerprint, secret string) string {
	if reason, ok := a.fingerprints[fingerprint]; ok {
		return "fingerprint: " + reason
	}
	for _, e := range a.paths {
		if doublestar.MatchUnvalidated(e.key, path) {
			return fmt.Sprintf("path %s: %s", e.text, e.reason)
		}
	}
	for _, e := range a.regexes {
		if e.key.MatchString(secret) {
			return fmt.Sprintf("regex %s: %s", e.text, e.reason)
		}
	}
	return ""
}

// Inline reports whether a line carries the inline marker.
func Inline(line []byte) bool {
	return bytes.Contains(line, []byte(InlineMarker))
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	return keys
}
