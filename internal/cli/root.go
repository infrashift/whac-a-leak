// Package cli implements the whac command line (ADR-0002).
package cli

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

// Exit codes (see the CLI contract spec).
const (
	ExitOK       = 0
	ExitFindings = 1
	ExitError    = 2
)

// exitError carries a specific exit code. A nil err means the command already
// reported what happened and only the exit code remains.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string {
	if e.err == nil {
		return fmt.Sprintf("exit status %d", e.code)
	}
	return e.err.Error()
}

func (e *exitError) Unwrap() error { return e.err }

type globalFlags struct {
	configDir string
}

const defaultConfigDir = ".whac"

// Execute runs whac with args and returns the process exit code.
func Execute(args []string, stdout, stderr io.Writer) int {
	root := newRootCmd(stdout, stderr)
	root.SetArgs(args)
	err := root.Execute()
	if err == nil {
		return ExitOK
	}
	var ee *exitError
	if errors.As(err, &ee) {
		if ee.err != nil {
			_, _ = fmt.Fprintf(stderr, "whac: %v\n", ee.err) // nothing useful to do if stderr fails
		}
		return ee.code
	}
	_, _ = fmt.Fprintf(stderr, "whac: %v\n", err) // nothing useful to do if stderr fails
	return ExitError
}

func newRootCmd(stdout, stderr io.Writer) *cobra.Command {
	g := &globalFlags{}
	root := &cobra.Command{
		Use:   "whac",
		Short: "Whac-a-Leak: a lightweight secret scanner with self-testing CUE rules",
		Long: `Whac-a-Leak (whac) finds secrets before they leak. Rules are CUE: typed,
layered by organization policy from .whac/*.cue, and self-testing. Every
rule's own samples run through the engine whenever rules load.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.PersistentFlags().StringVar(&g.configDir, "config-dir", defaultConfigDir,
		"directory of user rule layers (*.cue, package whac)")

	root.AddCommand(newVersionCmd(), newRulesCmd(g))
	return root
}
