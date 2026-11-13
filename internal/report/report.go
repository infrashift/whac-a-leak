// Package report writes scan results as text or JSON.
package report

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

// APIVersion and Kind identify the JSON report format.
const (
	APIVersion = "whac.infrashift.io/v1alpha1"
	Kind       = "ScanReport"
)

// Options controls how a report looks.
type Options struct {
	// NoRedact prints secrets in full. The default prints Finding.Redacted.
	NoRedact bool
	// Color adds ANSI colors to text output.
	Color bool
}

func secretText(f *scan.Finding, o Options) string {
	if o.NoRedact {
		return f.Secret
	}
	return f.Redacted
}

type jsonFinding struct {
	RuleID      string         `json:"ruleId"`
	Description string         `json:"description"`
	Severity    rules.Severity `json:"severity"`
	Path        string         `json:"path"`
	Line        int            `json:"line"`
	Column      int            `json:"column"`
	EndLine     int            `json:"endLine"`
	Fingerprint string         `json:"fingerprint"`
	Secret      string         `json:"secret"`
	Redacted    bool           `json:"redacted"`
}

type jsonSummary struct {
	Findings   int            `json:"findings"`
	BySeverity map[string]int `json:"bySeverity"`
	scan.Stats
}

type jsonReport struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Summary    jsonSummary   `json:"summary"`
	Findings   []jsonFinding `json:"findings"`
}

// JSON writes the result as one indented JSON document.
func JSON(w io.Writer, res *scan.Result, o Options) error {
	rep := jsonReport{
		APIVersion: APIVersion,
		Kind:       Kind,
		Summary:    jsonSummary{Findings: len(res.Findings), BySeverity: bySeverity(res.Findings), Stats: res.Stats},
		Findings:   make([]jsonFinding, 0, len(res.Findings)),
	}
	for i := range res.Findings {
		f := &res.Findings[i]
		rep.Findings = append(rep.Findings, jsonFinding{
			RuleID: f.RuleID, Description: f.Description, Severity: f.Severity,
			Path: f.Path, Line: f.Line, Column: f.Column, EndLine: f.EndLine,
			Fingerprint: f.Fingerprint, Secret: secretText(f, o), Redacted: !o.NoRedact,
		})
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}

var severityColor = map[rules.Severity]string{
	rules.SeverityCritical: "\x1b[1;31m",
	rules.SeverityHigh:     "\x1b[31m",
	rules.SeverityMedium:   "\x1b[33m",
	rules.SeverityLow:      "\x1b[36m",
}

const colorReset = "\x1b[0m"

// Text writes one block per finding and a summary line.
func Text(w io.Writer, res *scan.Result, o Options) error {
	p := &printer{w: w}
	for i := range res.Findings {
		f := &res.Findings[i]
		sev := fmt.Sprintf("%-8s", f.Severity)
		if o.Color {
			sev = severityColor[f.Severity] + sev + colorReset
		}
		p.printf("%s  %s  %s:%d:%d\n", sev, f.RuleID, f.Path, f.Line, f.Column)
		p.printf("          %s\n", f.Description)
		p.printf("          secret: %s\n", secretText(f, o))
		p.printf("          fingerprint: %s\n\n", f.Fingerprint)
	}
	p.printf("%s\n", Summary(res))
	return p.err
}

// Summary is a one-line description of a result.
func Summary(res *scan.Result) string {
	s := res.Stats
	var b strings.Builder
	switch n := len(res.Findings); n {
	case 0:
		b.WriteString("No findings.")
	default:
		counts := bySeverity(res.Findings)
		var parts []string
		for _, sev := range []rules.Severity{rules.SeverityCritical, rules.SeverityHigh, rules.SeverityMedium, rules.SeverityLow} {
			if c := counts[sev.String()]; c > 0 {
				parts = append(parts, fmt.Sprintf("%d %s", c, sev))
			}
		}
		fmt.Fprintf(&b, "%d finding(s): %s.", n, strings.Join(parts, ", "))
	}
	fmt.Fprintf(&b, " Scanned %d file(s), %s.", s.Files, humanBytes(s.Bytes))
	if skipped := s.Binary + s.Large + s.Excluded + s.Unreadable; skipped > 0 {
		fmt.Fprintf(&b, " Skipped %d: %d binary, %d large, %d excluded, %d unreadable.",
			skipped, s.Binary, s.Large, s.Excluded, s.Unreadable)
	}
	if s.Suppressed > 0 {
		fmt.Fprintf(&b, " Suppressed %d.", s.Suppressed)
	}
	return b.String()
}

func bySeverity(fs []scan.Finding) map[string]int {
	m := map[string]int{}
	for i := range fs {
		m[fs[i].Severity.String()]++
	}
	return m
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// printer keeps the first write error.
type printer struct {
	w   io.Writer
	err error
}

func (p *printer) printf(format string, a ...any) {
	if p.err == nil {
		_, p.err = fmt.Fprintf(p.w, format, a...)
	}
}
