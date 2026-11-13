package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infrashift/whac-a-leak/internal/exposure"
	"github.com/infrashift/whac-a-leak/internal/report"
	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

type scanFlags struct {
	format    string
	noRedact  bool
	threshold string
	staged    bool
	git       string
	// Agent-context preflight (ADR-0011).
	agentContext, env, home, probeIMDS bool
	roots                              []string
}

func (f *scanFlags) agentMode() bool { return f.agentContext || f.env || f.home }

func newScanCmd(g *globalFlags) *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan a directory, a file, staged changes or git history for secrets",
		Long: `Scan a directory or file for secrets. Inside a git work tree, whac scans the
files that git tracks or would track, so .gitignore applies.

--staged scans the staged changes of the repository at path (a pre-commit
check) and reports only findings on added lines. --git scans the lines that
each commit adds; --git=<range> limits the commits, e.g. --git=main..HEAD.

--agent-context audits what an agent's process can reach: environment
variables (--env), well-known credential files under $HOME and each --root
(--home), and credential sockets. --probe-imds adds one GET to the cloud
metadata endpoint. Values are always masked in this mode.

Exit codes: 0 no findings at or above the threshold, 1 findings, 2 error.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				if f.agentMode() {
					return &exitError{code: ExitError, err: errors.New("agent-context modes take no path; use --root for extra directories")}
				}
				root = args[0]
			}
			return runScan(cmd, g, f, root)
		},
	}
	cmd.Flags().BoolVar(&f.staged, "staged", false, "scan staged changes; report only added lines")
	cmd.Flags().StringVar(&f.git, "git", "", "scan git history; optional range as --git=<range> (default HEAD)")
	cmd.Flags().Lookup("git").NoOptDefVal = "HEAD"
	cmd.Flags().BoolVar(&f.agentContext, "agent-context", false, "preflight: audit environment variables, credential files and credential sockets")
	cmd.Flags().BoolVar(&f.env, "env", false, "audit environment variables only")
	cmd.Flags().BoolVar(&f.home, "home", false, "audit credential files under $HOME and --root only")
	cmd.Flags().StringSliceVar(&f.roots, "root", nil, "extra root for credential files, e.g. a mounted volume (repeatable)")
	cmd.Flags().BoolVar(&f.probeIMDS, "probe-imds", false, "with --agent-context: send one GET to the cloud metadata endpoint")
	cmd.Flags().StringVar(&f.format, "format", "text", "output format: text, json or sarif")
	cmd.Flags().BoolVar(&f.noRedact, "no-redact", false, "print secrets in full")
	cmd.Flags().StringVar(&f.threshold, "severity-threshold", "LOW", "report findings at or above LOW, MEDIUM, HIGH or CRITICAL")
	return cmd
}

func runScan(cmd *cobra.Command, g *globalFlags, f *scanFlags, root string) error {
	minSev, err := rules.ParseSeverity(f.threshold)
	if err != nil {
		return &exitError{code: ExitError, err: fmt.Errorf("--severity-threshold: %w", err)}
	}
	write, err := writer(f.format)
	if err != nil {
		return err
	}
	rs, err := loadVerified(cmd, g)
	if err != nil {
		return err
	}
	p, err := scan.NewPipeline(rs)
	if err != nil {
		return &exitError{code: ExitError, err: err}
	}
	modes := 0
	for _, on := range []bool{f.staged, f.git != "", f.agentMode()} {
		if on {
			modes++
		}
	}
	if modes > 1 {
		return &exitError{code: ExitError, err: errors.New("--staged, --git and the agent-context modes exclude each other")}
	}
	if f.probeIMDS && !f.agentContext {
		return &exitError{code: ExitError, err: errors.New("--probe-imds needs --agent-context")}
	}
	if len(f.roots) > 0 && !f.agentContext && !f.home {
		return &exitError{code: ExitError, err: errors.New("--root needs --agent-context or --home")}
	}
	var res *scan.Result
	switch {
	case f.agentMode():
		res, err = runAgent(cmd, p, f)
	case f.staged:
		res, err = p.Staged(cmd.Context(), root)
	case f.git != "":
		if _, statErr := os.Stat(root); statErr != nil {
			return &exitError{code: ExitError, err: fmt.Errorf("%w; to pass a revision range, use --git=<range>", statErr)}
		}
		res, err = p.History(cmd.Context(), root, f.git)
	default:
		res, err = p.Files(cmd.Context(), root)
	}
	if err != nil {
		return &exitError{code: ExitError, err: err}
	}
	res.Findings = scan.Filter(res.Findings, minSev)

	out := cmd.OutOrStdout()
	v, _, _ := buildVersion()
	opts := report.Options{
		NoRedact: f.noRedact, Color: f.format == "text" && colorEnabled(out),
		Rules: rs.Enabled(), ToolVersion: v,
	}
	if err := write(out, res, opts); err != nil {
		return &exitError{code: ExitError, err: err}
	}
	if len(res.Findings) > 0 {
		return &exitError{code: ExitFindings}
	}
	return nil
}

func runAgent(cmd *cobra.Command, p *scan.Pipeline, f *scanFlags) (*scan.Result, error) {
	for _, r := range f.roots {
		if fi, err := os.Stat(r); err != nil || !fi.IsDir() {
			return nil, fmt.Errorf("--root %s is not a directory", r)
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("home directory: %w", err)
	}
	all := f.agentContext
	return p.AgentContext(cmd.Context(), &scan.AgentOptions{
		Env:  all || f.env,
		Home: all || f.home,
		// Exposure checks run only in the full preflight.
		Exposure:      all,
		Environ:       os.Environ(),
		HomeDir:       home,
		Roots:         f.roots,
		Getenv:        os.Getenv,
		DockerSockets: exposure.DefaultDockerSockets,
		ProbeIMDS:     f.probeIMDS,
	})
}

func writer(format string) (func(io.Writer, *scan.Result, report.Options) error, error) {
	switch format {
	case "text":
		return report.Text, nil
	case "json":
		return report.JSON, nil
	case "sarif":
		return report.SARIF, nil
	}
	return nil, &exitError{code: ExitError, err: fmt.Errorf("unknown --format %q (want text, json or sarif)", format)}
}

// colorEnabled reports whether w is a terminal and NO_COLOR is unset.
func colorEnabled(w io.Writer) bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
