package engine

import (
	"fmt"
	"strings"

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
		sc := newScanner([]*rules.Rule{r}, true)
		for i, s := range r.Tests.MustMatch {
			t := &trace{}
			if len(sc.scan(s.Name(), []byte(s.Content), t)) == 0 {
				reason := t.explain()
				if r.Kind == rules.KindPattern && t.keywordHit && t.regexHits == 0 && !r.Multiline && strings.ContainsRune(s.Content, '\n') {
					reason += " on the keyword's line; set multiline: true for a secret that spans lines"
				}
				out = append(out, Failure{RuleID: r.ID, Sample: fmt.Sprintf("mustMatch[%d]", i), Reason: reason})
			}
		}
		for i, s := range r.Tests.MustIgnore {
			if n := len(sc.scan(s.Name(), []byte(s.Content), nil)); n > 0 {
				out = append(out, Failure{RuleID: r.ID, Sample: fmt.Sprintf("mustIgnore[%d]", i),
					Reason: fmt.Sprintf("produced %d finding(s), expected none", n)})
			}
		}
	}
	return out
}
