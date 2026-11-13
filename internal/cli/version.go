package cli

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Set by -ldflags "-X github.com/infrashift/whac-a-leak/internal/cli.version=…"
// (see the Makefile); otherwise filled from the module's build info.
var (
	version = ""
	commit  = ""
	date    = ""
)

func buildVersion() (v, c, d string) {
	v, c, d = version, commit, date
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return orDefault(v, "dev"), orDefault(c, "unknown"), orDefault(d, "unknown")
	}
	if v == "" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if c == "" {
				c = s.Value
			}
		case "vcs.time":
			if d == "" {
				d = s.Value
			}
		}
	}
	return orDefault(v, "dev"), orDefault(c, "unknown"), orDefault(d, "unknown")
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version, commit and build date",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			v, c, d := buildVersion()
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "whac %s (commit %s, built %s)\n", v, c, d)
			return err
		},
	}
}
