package cli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/infrashift/whac-a-leak/internal/engine"
	"github.com/infrashift/whac-a-leak/internal/rules"
)

func newRulesCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "List, verify and export the rules",
	}
	cmd.AddCommand(newRulesListCmd(g), newRulesVerifyCmd(g), newRulesExportSchemaCmd(g))
	return cmd
}

func loadOptions(cmd *cobra.Command, g *globalFlags) rules.Options {
	return rules.Options{
		ConfigDir:        g.configDir,
		RequireConfigDir: cmd.Flags().Changed("config-dir"),
	}
}

// loadVerified loads every layer and runs the rules' own tests; any failing
// sample is a configuration error (ADR-0009). Commands that scan use this.
func loadVerified(cmd *cobra.Command, g *globalFlags) (*rules.Ruleset, error) {
	rs, err := rules.Load(loadOptions(cmd, g))
	if err != nil {
		return nil, &exitError{code: ExitError, err: err}
	}
	if failures := engine.Verify(rs.Rules); len(failures) > 0 {
		lines := make([]string, len(failures))
		for i, f := range failures {
			lines[i] = "  " + f.String()
		}
		return nil, &exitError{code: ExitError, err: fmt.Errorf(
			"%d rule sample(s) failed; run `whac rules verify` for details:\n%s",
			len(failures), strings.Join(lines, "\n"))}
	}
	return rs, nil
}

type ruleJSON struct {
	ID          string         `json:"id"`
	Kind        rules.Kind     `json:"kind"`
	Severity    rules.Severity `json:"severity"`
	Enabled     bool           `json:"enabled"`
	Description string         `json:"description"`
	Keywords    []string       `json:"keywords"`
}

func newRulesListCmd(g *globalFlags) *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the effective rules after layering",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := loadVerified(cmd, g)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			switch format {
			case "text":
				return writeRulesTable(out, rs)
			case "json":
				list := make([]ruleJSON, len(rs.Rules))
				for i, r := range rs.Rules {
					list[i] = ruleJSON{ID: r.ID, Kind: r.Kind, Severity: r.Severity, Enabled: r.Enabled,
						Description: r.Description, Keywords: r.Keywords}
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(list)
			}
			return &exitError{code: ExitError, err: fmt.Errorf("unknown --format %q (want text or json)", format)}
		},
	}
	cmd.Flags().StringVar(&format, "format", "text", "output format: text or json")
	return cmd
}

func writeRulesTable(out io.Writer, rs *rules.Ruleset) error {
	tw := tabwriter.NewWriter(out, 0, 4, 2, ' ', 0)
	p := &printer{w: tw}
	p.printf("ID\tKIND\tSEVERITY\tENABLED\tDESCRIPTION\n")
	for _, r := range rs.Rules {
		p.printf("%s\t%s\t%s\t%t\t%s\n", r.ID, r.Kind, r.Severity, r.Enabled, r.Description)
	}
	if p.err != nil {
		return p.err
	}
	if err := tw.Flush(); err != nil {
		return err
	}
	if len(rs.Layers) > 0 {
		_, err := fmt.Fprintf(out, "\nlayers: %s\n", strings.Join(rs.Layers, ", "))
		return err
	}
	return nil
}

func newRulesVerifyCmd(g *globalFlags) *cobra.Command {
	return &cobra.Command{
		Use:   "verify",
		Short: "Run every rule's mustMatch and mustIgnore samples through the engine",
		Long: `Run every rule's own samples through the same matching code a scan uses.
Exits 2 if any sample misbehaves. Failures name the sample by position and
never print its text.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rs, err := rules.Load(loadOptions(cmd, g))
			if err != nil {
				return &exitError{code: ExitError, err: err}
			}
			failures := engine.Verify(rs.Rules)
			byRule := map[string][]engine.Failure{}
			for _, f := range failures {
				byRule[f.RuleID] = append(byRule[f.RuleID], f)
			}
			p := &printer{w: cmd.OutOrStdout()}
			samples := 0
			for _, r := range rs.Rules {
				n := len(r.Tests.MustMatch) + len(r.Tests.MustIgnore)
				samples += n
				failed := byRule[r.ID]
				if len(failed) == 0 {
					p.printf("PASS  %-28s %d sample(s)\n", r.ID, n)
					continue
				}
				p.printf("FAIL  %-28s %d of %d sample(s)\n", r.ID, len(failed), n)
				for _, f := range failed {
					p.printf("        %s: %s\n", f.Sample, f.Reason)
				}
			}
			p.printf("\n%d rule(s), %d sample(s), %d failure(s)\n", len(rs.Rules), samples, len(failures))
			if p.err != nil {
				return &exitError{code: ExitError, err: p.err}
			}
			if len(failures) > 0 {
				return &exitError{code: ExitError}
			}
			return nil
		},
	}
}

func newRulesExportSchemaCmd(g *globalFlags) *cobra.Command {
	var force bool
	cmd := &cobra.Command{
		Use:   "export-schema",
		Short: "Write the rule schema into the config dir for editors and `cue vet`",
		Long: `Write the embedded CUE schema to <config-dir>/schema.cue so editors and
` + "`cue vet`" + ` understand your layers. The copy carries a generated-file header and
is ignored by the loader, so a stale copy never conflicts with the schema
built into whac. Re-run after upgrading whac.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			content := append([]byte(rules.GeneratedHeader+"\n\n"), rules.SchemaSource()...)
			target := filepath.Join(g.configDir, "schema.cue")

			existing, readErr := os.ReadFile(target) //nolint:gosec // G304: the path is the user's own config dir
			switch {
			case readErr == nil && bytes.Equal(existing, content):
				_, err := fmt.Fprintf(cmd.OutOrStdout(), "%s is up to date\n", target)
				return err
			case readErr == nil && !bytes.HasPrefix(existing, []byte(rules.GeneratedHeader)) && !force:
				return &exitError{code: ExitError, err: fmt.Errorf(
					"%s exists and was not written by whac; use --force to overwrite it", target)}
			case readErr != nil && !errors.Is(readErr, fs.ErrNotExist):
				return &exitError{code: ExitError, err: readErr}
			}

			if err := os.MkdirAll(g.configDir, 0o750); err != nil {
				return &exitError{code: ExitError, err: err}
			}
			if err := os.WriteFile(target, content, 0o644); err != nil { //nolint:gosec // G306: a schema meant to be committed and read by editors
				return &exitError{code: ExitError, err: err}
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", target)
			return err
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "overwrite a schema.cue that was not written by whac")
	return cmd
}
