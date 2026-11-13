package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infrashift/whac-a-leak/internal/report"
	"github.com/infrashift/whac-a-leak/internal/rules"
	"github.com/infrashift/whac-a-leak/internal/scan"
)

type scanFlags struct {
	format    string
	noRedact  bool
	threshold string
}

func newScanCmd(g *globalFlags) *cobra.Command {
	f := &scanFlags{}
	cmd := &cobra.Command{
		Use:   "scan [path]",
		Short: "Scan a directory or file for secrets",
		Long: `Scan a directory or file for secrets. Inside a git work tree, whac scans the
files that git tracks or would track, so .gitignore applies. Exit codes: 0 no
findings at or above the threshold, 1 findings, 2 error.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := "."
			if len(args) == 1 {
				root = args[0]
			}
			return runScan(cmd, g, f, root)
		},
	}
	cmd.Flags().StringVar(&f.format, "format", "text", "output format: text or json")
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
	res, err := p.Files(cmd.Context(), root)
	if err != nil {
		return &exitError{code: ExitError, err: err}
	}
	res.Findings = scan.Filter(res.Findings, minSev)

	out := cmd.OutOrStdout()
	opts := report.Options{NoRedact: f.noRedact, Color: f.format == "text" && colorEnabled(out)}
	if err := write(out, res, opts); err != nil {
		return &exitError{code: ExitError, err: err}
	}
	if len(res.Findings) > 0 {
		return &exitError{code: ExitFindings}
	}
	return nil
}

func writer(format string) (func(io.Writer, *scan.Result, report.Options) error, error) {
	switch format {
	case "text":
		return report.Text, nil
	case "json":
		return report.JSON, nil
	}
	return nil, &exitError{code: ExitError, err: fmt.Errorf("unknown --format %q (want text or json)", format)}
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
