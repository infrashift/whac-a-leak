package scan

import (
	"context"
	"os"
	"runtime"

	"github.com/infrashift/whac-a-leak/internal/exposure"
	"github.com/infrashift/whac-a-leak/internal/fingerprint"
	"github.com/infrashift/whac-a-leak/internal/source"
)

// EnvPath is the display path of environment findings.
const EnvPath = "env"

// masked replaces every value in agent-context findings.
const masked = "[masked]"

// formatNames maps a credential path's format to a file name that selects
// the extractor.
var formatNames = map[string]string{
	"yaml": "credential.yaml", "json": "credential.json", "dotenv": "credential.env",
	"properties": "credential.properties", "ini": "credential.ini", "hcl": "credential.hcl",
}

// AgentOptions selects the parts of an agent-context scan (ADR-0011).
type AgentOptions struct {
	Env, Home, Exposure bool
	// Environ is the environment to audit, as os.Environ returns it.
	Environ []string
	// HomeDir is the user's home directory; Roots are extra roots such as
	// mounted volumes. Credential paths resolve under each.
	HomeDir string
	Roots   []string
	// Exposure check settings; see package exposure.
	Getenv        func(string) string
	DockerSockets []string
	ProbeIMDS     bool
	IMDSURL       string
}

// AgentContext audits what an agent's process can reach: environment
// variables, well-known credential files and credential sockets. Values are
// always masked, because preflight output usually lands in logs the agent
// can read.
func (p *Pipeline) AgentContext(ctx context.Context, o *AgentOptions) (*Result, error) {
	var res Result
	if o.Env {
		content, names := source.Environment(o.Environ)
		res.Stats.Files++
		res.Stats.Bytes += int64(len(content))
		findings, suppressed := p.content(EnvPath, EnvPath, "environment.env", content)
		res.Stats.Suppressed += suppressed
		for i := range findings {
			f := &findings[i]
			if f.Line >= 1 && f.Line <= len(names) {
				f.KeyPath = "env." + names[f.Line-1]
			}
			f.Line, f.Column, f.EndLine = 0, 0, 0
			res.Findings = append(res.Findings, *f)
		}
	}
	if o.Home {
		roots := append([]string{o.HomeDir}, o.Roots...)
		files, err := source.CredentialFiles(roots, o.HomeDir, p.Rules.Config.CredentialPaths, runtime.GOOS)
		if err != nil {
			return nil, err
		}
		for _, cf := range files {
			if p.Excluded(cf.Rel) {
				res.Stats.Excluded++
				continue
			}
			info, err := os.Stat(cf.Abs)
			if err != nil {
				res.Stats.Unreadable++
				continue
			}
			if info.Size() > cf.MaxSize {
				res.Stats.Large++
				continue
			}
			content, err := os.ReadFile(cf.Abs)
			if err != nil {
				res.Stats.Unreadable++
				continue
			}
			if IsBinary(content) {
				res.Stats.Binary++
				continue
			}
			res.Stats.Files++
			res.Stats.Bytes += int64(len(content))
			name := cf.Rel
			if n, ok := formatNames[cf.Format]; ok {
				name = n
			}
			findings, suppressed := p.content(cf.Path, cf.Rel, name, content)
			res.Stats.Suppressed += suppressed
			res.Findings = append(res.Findings, findings...)
		}
	}
	if o.Exposure {
		for _, e := range exposure.Check(ctx, p.Rules.Config.Exposure, exposure.Options{
			Getenv: o.Getenv, DockerSockets: o.DockerSockets, ProbeIMDS: o.ProbeIMDS, IMDSURL: o.IMDSURL,
		}) {
			fp := fingerprint.Of(e.ID, e.Path)
			if p.Allowlist.Match(e.Path, fp, e.Path) != "" {
				res.Stats.Suppressed++
				continue
			}
			res.Findings = append(res.Findings, Finding{
				Category: CategoryExposure, RuleID: e.ID, Description: e.Description,
				Severity: e.Severity, Path: e.Path, Fingerprint: fp,
			})
		}
	}
	for i := range res.Findings {
		res.Findings[i].Secret = ""
		res.Findings[i].Redacted = masked
	}
	Sort(res.Findings)
	return &res, nil
}
