package rules

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// document mirrors the top level of the CUE schema (builtin/schema.cue).
type document struct {
	Rules     map[string]ruleSpec     `json:"rules"`
	Overrides map[string]overrideSpec `json:"overrides"`
	Config    configSpec              `json:"config"`
	Allowlist allowlistSpec           `json:"allowlist"`
}

type ruleSpec struct {
	ID           string       `json:"id"`
	Description  string       `json:"description"`
	Kind         string       `json:"kind"`
	Severity     string       `json:"severity"`
	Enabled      bool         `json:"enabled"`
	Keywords     []string     `json:"keywords"`
	Regex        string       `json:"regex"`
	SecretGroup  int          `json:"secretGroup"`
	Multiline    bool         `json:"multiline"`
	KeyContext   string       `json:"keyContext"`
	Paths        []string     `json:"paths"`
	Entropy      *entropySpec `json:"entropy"`
	Stopwords    []string     `json:"stopwords"`
	AllowRegexes []string     `json:"allowRegexes"`
	Tests        testsSpec    `json:"tests"`
}

type entropySpec struct {
	Min     *float64 `json:"min"`
	Max     *float64 `json:"max"`
	MinLen  int      `json:"minLen"`
	Charset string   `json:"charset"`
}

type testsSpec struct {
	MustMatch  []string `json:"mustMatch"`
	MustIgnore []string `json:"mustIgnore"`
}

type overrideSpec struct {
	Severity *string `json:"severity"`
	Enabled  *bool   `json:"enabled"`
}

type configSpec struct {
	MaxFileSize int64             `json:"maxFileSize"`
	Exclude     map[string]string `json:"exclude"`
}

type allowlistSpec struct {
	Fingerprints map[string]string `json:"fingerprints"`
	Paths        map[string]string `json:"paths"`
	Regexes      map[string]string `json:"regexes"`
}

// compile turns the decoded document into a Ruleset, applying overrides and
// collecting every error rather than stopping at the first.
func compile(doc *document) (*Ruleset, error) {
	var errs []error
	rs := &Ruleset{
		Config: Config{MaxFileSize: doc.Config.MaxFileSize, Exclude: nonNil(doc.Config.Exclude)},
		Allowlist: Allowlist{
			Fingerprints: nonNil(doc.Allowlist.Fingerprints),
			Paths:        nonNil(doc.Allowlist.Paths),
			Regexes:      nonNil(doc.Allowlist.Regexes),
		},
	}
	for re := range rs.Allowlist.Regexes {
		if _, err := regexp.Compile(re); err != nil {
			errs = append(errs, fmt.Errorf("allowlist.regexes: %w", err))
		}
	}

	ids := make([]string, 0, len(doc.Rules))
	for id := range doc.Rules {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		spec := doc.Rules[id]
		r, err := compileRule(&spec)
		if err != nil {
			errs = append(errs, fmt.Errorf("rules.%s: %w", id, err))
			continue
		}
		rs.Rules = append(rs.Rules, r)
	}

	oids := make([]string, 0, len(doc.Overrides))
	for id := range doc.Overrides {
		oids = append(oids, id)
	}
	slices.Sort(oids)
	for _, id := range oids {
		if err := applyOverride(rs, id, doc.Overrides[id]); err != nil {
			errs = append(errs, fmt.Errorf("overrides.%s: %w", id, err))
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}
	return rs, nil
}

func applyOverride(rs *Ruleset, id string, o overrideSpec) error {
	r := rs.Rule(id)
	if r == nil {
		return errors.New("no rule with this ID")
	}
	if o.Severity != nil {
		sev, err := ParseSeverity(*o.Severity)
		if err != nil {
			return err
		}
		r.Severity = sev
	}
	if o.Enabled != nil {
		r.Enabled = *o.Enabled
	}
	return nil
}

func compileRule(s *ruleSpec) (*Rule, error) {
	sev, err := ParseSeverity(s.Severity)
	if err != nil {
		return nil, err
	}
	r := &Rule{
		ID:          s.ID,
		Description: s.Description,
		Kind:        Kind(s.Kind),
		Severity:    sev,
		Enabled:     s.Enabled,
		SecretGroup: s.SecretGroup,
		Multiline:   s.Multiline,
		Paths:       s.Paths,
		Tests:       Tests{MustMatch: s.Tests.MustMatch, MustIgnore: s.Tests.MustIgnore},
	}
	switch r.Kind {
	case KindPattern:
	case KindAssignment:
		// Assignment rules need the structural extractors (Milestone 3).
		return nil, errors.New(`kind "assignment" is not supported yet; use a pattern rule`)
	default:
		return nil, fmt.Errorf("unknown kind %q", s.Kind)
	}

	for _, k := range s.Keywords {
		r.Keywords = append(r.Keywords, strings.ToLower(k))
	}
	if r.Regex, err = regexp.Compile(s.Regex); err != nil {
		return nil, fmt.Errorf("regex: %w", err)
	}
	if n := r.Regex.NumSubexp(); r.SecretGroup > n {
		return nil, fmt.Errorf("secretGroup %d, but the regex has only %d capture group(s)", r.SecretGroup, n)
	}
	if s.KeyContext != "" {
		if r.KeyContext, err = regexp.Compile(s.KeyContext); err != nil {
			return nil, fmt.Errorf("keyContext: %w", err)
		}
	}
	for i, re := range s.AllowRegexes {
		c, err := regexp.Compile(re)
		if err != nil {
			return nil, fmt.Errorf("allowRegexes[%d]: %w", i, err)
		}
		r.AllowRegexes = append(r.AllowRegexes, c)
	}
	for _, w := range s.Stopwords {
		r.Stopwords = append(r.Stopwords, strings.ToLower(w))
	}
	if e := s.Entropy; e != nil {
		r.Entropy = &Entropy{Min: e.Min, Max: e.Max, MinLen: e.MinLen, Charset: Charset(e.Charset)}
	}
	return r, nil
}

func nonNil(m map[string]string) map[string]string {
	if m == nil {
		return map[string]string{}
	}
	return m
}
