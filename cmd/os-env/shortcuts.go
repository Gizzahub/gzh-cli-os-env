// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import (
	"errors"
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/shortcuts"
	"github.com/spf13/cobra"
)

func newShortcutsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shortcuts",
		Short: "Keyboard shortcut management",
	}
	cmd.AddCommand(newShortcutsListCmd())
	return cmd
}

func newShortcutsListCmd() *cobra.Command {
	var all bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List system keyboard shortcuts",
		Long: `List system keyboard shortcuts and their key bindings.

On macOS this reads com.apple.symbolichotkeys. Apple does not document the
shortcut IDs, so an ID with no confirmed name is reported as "unknown" rather
than guessed. Other platforms are not yet supported.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items, err := shortcuts.List(cmd.Context())
			if err != nil {
				if errors.Is(err, shortcuts.ErrUnsupported) {
					return fmt.Errorf("shortcuts list is not supported on this platform yet")
				}
				return err
			}

			out := cmd.OutOrStdout()
			shown := 0
			for _, s := range items {
				if !all && !s.Enabled {
					continue
				}
				shown++
				if _, err := fmt.Fprintf(out, "%s\n  %s\n", describeShortcut(s), describeBinding(s)); err != nil {
					return err
				}
			}
			if shown == 0 {
				_, err := fmt.Fprintln(out, "No shortcuts found")
				return err
			}
			if !all {
				_, err := fmt.Fprintf(out, "\n%d enabled of %d total (--all to include disabled)\n", shown, len(items))
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "Include disabled shortcuts")
	return cmd
}

// describeShortcut names a shortcut, falling back to its raw ID when the ID is
// not in the confirmed table.
func describeShortcut(s shortcuts.Shortcut) string {
	if s.Named() {
		return fmt.Sprintf("%s (id %d)", s.Name, s.ID)
	}
	return fmt.Sprintf("unknown shortcut (id %d)", s.ID)
}

// describeBinding reports an empty binding as the system default, which is what
// macOS stores it as — not as "no shortcut".
func describeBinding(s shortcuts.Shortcut) string {
	state := "enabled"
	if !s.Enabled {
		state = "disabled"
	}
	binding := s.Binding
	if binding == "" {
		binding = "system default binding"
	}
	return fmt.Sprintf("%s [%s]", binding, state)
}
