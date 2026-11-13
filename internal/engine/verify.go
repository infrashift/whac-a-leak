package engine

import (
	"bytes"
	"fmt"

	"github.com/infrashift/whac-a-leak/internal/rules"
)

// Failure is a rule sample that did not behave as declared. It names the
// sample by position and never carries the sample text.
type Failure struct {
	RuleID string
	// Sample is "mustMatch[i]" or "mustIgnore[i]".
	Sample string
	Reason string
}

func (f Failure) String() string {
	return fmt.Sprintf("%s %s: %s", f.RuleID, f.Sample, f.Reason)
}

// Verify runs every rule's own samples through the same matching code a scan
// uses (ADR-0009). Disabled rules are verified too: a broken rule is broken
// whether or not a layer has switched it off.
func Verify(rs []*rules.Rule) []Failure {
	var out []Failure
	for _, r := range rs {
		for i, s := range r.Tests.MustMatch {
			t := &trace{}
			content := []byte(s)
			if len(matchRule(r, content, bytes.ToLower(content), t)) == 0 {
				out = append(out, Failure{RuleID: r.ID, Sample: fmt.Sprintf("mustMatch[%d]", i), Reason: t.explain()})
			}
		}
		for i, s := range r.Tests.MustIgnore {
			content := []byte(s)
			if n := len(matchRule(r, content, bytes.ToLower(content), nil)); n > 0 {
				out = append(out, Failure{RuleID: r.ID, Sample: fmt.Sprintf("mustIgnore[%d]", i),
					Reason: fmt.Sprintf("produced %d finding(s), expected none", n)})
			}
		}
	}
	return out
}
