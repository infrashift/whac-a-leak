package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/infrashift/whac-a-leak/internal/fingerprint"
	"github.com/infrashift/whac-a-leak/internal/rules"
)

const allowlistFile = "allowlist.cue"

const allowlistHeader = `package whac

// Allowlist entries for false positives: test fixtures and documented examples.
// Rotate a real secret. Do not allowlist it.
`

func newAllowCmd(g *globalFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "allow",
		Short: "Manage allowlist entries in the config dir",
	}
	cmd.AddCommand(newAllowAddCmd(g))
	return cmd
}

func newAllowAddCmd(g *globalFlags) *cobra.Command {
	var reason string
	cmd := &cobra.Command{
		Use:   "add <fingerprint>",
		Short: "Allowlist a false positive by its whac1 fingerprint",
		Long: `Append an allowlist entry to <config-dir>/allowlist.cue. Use it only for
false positives such as test fixtures. Rotate a real secret: a committed
fingerprint of a short secret is open to an offline guess.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fp := args[0]
			if !fingerprint.Valid(fp) {
				return &exitError{code: ExitError, err: fmt.Errorf("%q is not a whac1 fingerprint", fp)}
			}
			if strings.TrimSpace(reason) == "" {
				return &exitError{code: ExitError, err: errors.New("--reason is required")}
			}
			// The config dir may not exist yet: allow add creates it.
			rs, err := loadVerifiedWith(rules.Options{ConfigDir: g.configDir})
			if err != nil {
				return err
			}
			if r, ok := rs.Allowlist.Fingerprints[fp]; ok {
				_, werr := fmt.Fprintf(cmd.OutOrStdout(), "%s is already allowlisted: %s\n", fp, r)
				return werr
			}
			target := filepath.Join(g.configDir, allowlistFile)
			if aerr := appendEntry(target, fp, reason); aerr != nil {
				return &exitError{code: ExitError, err: aerr}
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "added %s to %s\n", fp, target)
			return err
		},
	}
	cmd.Flags().StringVar(&reason, "reason", "", "why this finding is a false positive (required)")
	return cmd
}

// appendEntry adds one line to the allowlist file and checks that every
// layer still loads. On failure it restores the previous file.
func appendEntry(target, fp, reason string) error {
	old, err := os.ReadFile(target) //nolint:gosec // G304: the path is the user's own config dir
	existed := err == nil
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	content := old
	if !existed {
		content = []byte(allowlistHeader)
	}
	if len(content) > 0 && content[len(content)-1] != '\n' {
		content = append(content, '\n')
	}
	quoted, err := json.Marshal(reason) // a JSON string is a valid CUE string
	if err != nil {
		return err
	}
	content = append(content, fmt.Sprintf("\nallowlist: fingerprints: %q: %s\n", fp, quoted)...)

	if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
		return err
	}
	if err := os.WriteFile(target, content, 0o644); err != nil { //nolint:gosec // G306: a layer meant to be committed
		return err
	}
	if _, err := rules.Load(rules.Options{ConfigDir: filepath.Dir(target), RequireConfigDir: true}); err != nil {
		var restoreErr error
		if existed {
			restoreErr = os.WriteFile(target, old, 0o644) //nolint:gosec // G306: restore the previous layer
		} else {
			restoreErr = os.Remove(target)
		}
		return errors.Join(fmt.Errorf("the new entry breaks the config; %s is unchanged: %w", target, err), restoreErr)
	}
	return nil
}
