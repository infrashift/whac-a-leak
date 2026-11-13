package report

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

// SARIF constants.
const (
	sarifSchema  = "https://json.schemastore.org/sarif-2.1.0.json"
	sarifVersion = "2.1.0"
	// fingerprintKey names the whac1 fingerprint in partialFingerprints.
	fingerprintKey = "whacFingerprint/v1"
	infoURI        = "https://infrashift.github.io/whac-a-leak/"
)

type sarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version,omitempty"`
	InformationURI string      `json:"informationUri"`
	Rules          []sarifRule `json:"rules"`
}

type sarifText struct {
	Text string `json:"text"`
}

type sarifRule struct {
	ID                   string          `json:"id"`
	Name                 string          `json:"name"`
	ShortDescription     sarifText       `json:"shortDescription"`
	FullDescription      sarifText       `json:"fullDescription"`
	HelpURI              string          `json:"helpUri"`
	DefaultConfiguration sarifRuleConfig `json:"defaultConfiguration"`
	Properties           sarifRuleProps  `json:"properties"`
}

type sarifRuleConfig struct {
	Level string `json:"level"`
}

type sarifRuleProps struct {
	// GitHub code scanning reads security-severity (0.0–10.0).
	SecuritySeverity string   `json:"security-severity"`
	Tags             []string `json:"tags"`
}

type sarifResult struct {
	RuleID              string            `json:"ruleId"`
	RuleIndex           int               `json:"ruleIndex"`
	Level               string            `json:"level"`
	Message             sarifText         `json:"message"`
	Locations           []sarifLocation   `json:"locations"`
	PartialFingerprints map[string]string `json:"partialFingerprints"`
	Properties          map[string]string `json:"properties,omitempty"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysical `json:"physicalLocation"`
}

type sarifPhysical struct {
	ArtifactLocation sarifArtifact `json:"artifactLocation"`
	Region           sarifRegion   `json:"region"`
}

type sarifArtifact struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn"`
	EndLine     int `json:"endLine"`
}

func sarifLevel(s rules.Severity) string {
	switch s {
	case rules.SeverityCritical, rules.SeverityHigh:
		return "error"
	case rules.SeverityMedium:
		return "warning"
	case rules.SeverityLow:
		return "note"
	}
	return "warning"
}

func securitySeverity(s rules.Severity) string {
	switch s {
	case rules.SeverityCritical:
		return "9.5"
	case rules.SeverityHigh:
		return "8.0"
	case rules.SeverityMedium:
		return "5.5"
	case rules.SeverityLow:
		return "3.0"
	}
	return "5.5"
}

// SARIF writes a SARIF 2.1.0 log. Secrets are always redacted, because SARIF
// files go to code scanning services; Options.NoRedact does not apply.
// Options.Rules lists the rules in tool.driver.rules; a finding whose rule is
// missing from it adds a minimal entry.
func SARIF(w io.Writer, res *scan.Result, o Options) error {
	driver := sarifDriver{Name: "whac", Version: o.ToolVersion, InformationURI: infoURI, Rules: []sarifRule{}}
	index := map[string]int{}
	addRule := func(id, desc string, sev rules.Severity) int {
		if i, ok := index[id]; ok {
			return i
		}
		index[id] = len(driver.Rules)
		driver.Rules = append(driver.Rules, sarifRule{
			ID: id, Name: id,
			ShortDescription:     sarifText{Text: desc},
			FullDescription:      sarifText{Text: desc},
			HelpURI:              infoURI + "docs/reference/default-rules/",
			DefaultConfiguration: sarifRuleConfig{Level: sarifLevel(sev)},
			Properties:           sarifRuleProps{SecuritySeverity: securitySeverity(sev), Tags: []string{"security", "secret"}},
		})
		return index[id]
	}
	for _, r := range o.Rules {
		addRule(r.ID, r.Description, r.Severity)
	}
	results := make([]sarifResult, 0, len(res.Findings))
	for i := range res.Findings {
		f := &res.Findings[i]
		msg := fmt.Sprintf("%s: %s", f.Description, f.Redacted)
		if f.KeyPath != "" {
			msg += fmt.Sprintf(" (key %s)", f.KeyPath)
		}
		r := sarifResult{
			RuleID:    f.RuleID,
			RuleIndex: addRule(f.RuleID, f.Description, f.Severity),
			Level:     sarifLevel(f.Severity),
			Message:   sarifText{Text: msg},
			Locations: []sarifLocation{{PhysicalLocation: sarifPhysical{
				ArtifactLocation: sarifArtifact{URI: f.Path},
				Region:           sarifRegion{StartLine: f.Line, StartColumn: f.Column, EndLine: f.EndLine},
			}}},
			PartialFingerprints: map[string]string{fingerprintKey: f.Fingerprint},
		}
		if f.Commit != "" {
			r.Properties = map[string]string{"commit": f.Commit, "author": f.Author, "date": f.Date}
		}
		results = append(results, r)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(sarifLog{
		Schema:  sarifSchema,
		Version: sarifVersion,
		Runs:    []sarifRun{{Tool: sarifTool{Driver: driver}, Results: results}},
	})
}
