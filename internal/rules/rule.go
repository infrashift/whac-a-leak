// Package rules loads Whac-a-Leak's CUE rule layers and compiles them into Go
// matchers. CUE is evaluated only here, once, at load time (ADR-0003); the
// scan engine works on the compiled Rule values.
package rules

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
)

// Severity orders findings from LOW to CRITICAL.
type Severity int

// Severities in ascending order.
const (
	SeverityLow Severity = iota + 1
	SeverityMedium
	SeverityHigh
	SeverityCritical
)

var severityNames = map[Severity]string{
	SeverityLow:      "LOW",
	SeverityMedium:   "MEDIUM",
	SeverityHigh:     "HIGH",
	SeverityCritical: "CRITICAL",
}

// ParseSeverity parses the CUE spelling of a severity.
func ParseSeverity(s string) (Severity, error) {
	for sev, name := range severityNames {
		if name == s {
			return sev, nil
		}
	}
	return 0, fmt.Errorf("unknown severity %q (want LOW, MEDIUM, HIGH or CRITICAL)", s)
}

// String returns the CUE spelling of the severity.
func (s Severity) String() string {
	if name, ok := severityNames[s]; ok {
		return name
	}
	return fmt.Sprintf("Severity(%d)", int(s))
}

// MarshalText encodes the severity by name, e.g. in JSON reports.
func (s Severity) MarshalText() ([]byte, error) {
	if _, ok := severityNames[s]; !ok {
		return nil, fmt.Errorf("invalid severity %d", int(s))
	}
	return []byte(s.String()), nil
}

// UnmarshalText decodes a severity name, e.g. from a JSON report.
func (s *Severity) UnmarshalText(b []byte) error {
	sev, err := ParseSeverity(string(b))
	if err != nil {
		return err
	}
	*s = sev
	return nil
}

// Kind says what a rule is evaluated against.
type Kind string

// Rule kinds.
const (
	// KindPattern rules run a keyword-gated regex over raw content.
	KindPattern Kind = "pattern"
	// KindAssignment rules run over (key path, value) pairs from structured
	// files; they arrive with structure-aware scanning (Milestone 3).
	KindAssignment Kind = "assignment"
)

// Charset is the alphabet a secret is drawn from; it bounds the entropy a
// secret can reach (ADR-0006).
type Charset string

// Charsets, from the smallest alphabet to the largest.
const (
	CharsetHex       Charset = "hex"
	CharsetAlnum     Charset = "alnum"
	CharsetBase64    Charset = "base64"
	CharsetBase64URL Charset = "base64url"
	// CharsetAny is any other printable text.
	CharsetAny Charset = "any"
)

// Entropy holds a rule's entropy constraints.
type Entropy struct {
	// Min and Max are bits per character; nil means unset. When Min is unset
	// the engine derives a length- and charset-aware default.
	Min, Max *float64
	MinLen   int
	// Charset is the alphabet the thresholds assume; empty means the engine
	// classifies each secret.
	Charset Charset
}

// Tests are a rule's own samples (ADR-0009).
type Tests struct {
	MustMatch  []string
	MustIgnore []string
}

// Rule is a compiled, ready-to-run rule.
type Rule struct {
	ID          string
	Description string
	Kind        Kind
	Severity    Severity
	Enabled     bool

	// Keywords are lowercased; content is matched case-insensitively.
	Keywords    []string
	Regex       *regexp.Regexp
	SecretGroup int
	KeyContext  *regexp.Regexp
	Paths       []string

	Entropy      *Entropy
	Stopwords    []string // lowercased
	AllowRegexes []*regexp.Regexp

	Tests Tests
}

// Config is the scan configuration after layering.
type Config struct {
	MaxFileSize int64
	// Exclude maps doublestar globs to the reason they are excluded.
	Exclude map[string]string
}

// Allowlist holds suppressions after layering; every entry carries a reason.
type Allowlist struct {
	Fingerprints map[string]string
	Paths        map[string]string
	Regexes      map[string]string
}

// Ruleset is the result of loading and compiling every layer.
type Ruleset struct {
	// Rules holds every rule, enabled or not, sorted by ID.
	Rules     []*Rule
	Config    Config
	Allowlist Allowlist
	// Layers lists the user files that were unified with the builtins.
	Layers []string
}

// Enabled returns the enabled rules, sorted by ID.
func (rs *Ruleset) Enabled() []*Rule {
	out := make([]*Rule, 0, len(rs.Rules))
	for _, r := range rs.Rules {
		if r.Enabled {
			out = append(out, r)
		}
	}
	return out
}

// Rule returns the rule with the given ID, or nil.
func (rs *Ruleset) Rule(id string) *Rule {
	i, ok := slices.BinarySearchFunc(rs.Rules, id, func(r *Rule, id string) int {
		return cmp.Compare(r.ID, id)
	})
	if !ok {
		return nil
	}
	return rs.Rules[i]
}
