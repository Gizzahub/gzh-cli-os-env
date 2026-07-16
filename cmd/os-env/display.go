// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import (
	"errors"
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/display"
	"github.com/spf13/cobra"
)

func newDisplayCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "display",
		Short: "Display / monitor management",
	}
	cmd.AddCommand(newDisplayListCmd())
	return cmd
}

func newDisplayListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List connected displays",
		Long: `List connected displays and their resolutions.

On macOS this reads system_profiler SPDisplaysDataType. Other platforms are
not yet supported.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items, err := display.List(cmd.Context())
			if err != nil {
				if errors.Is(err, display.ErrUnsupported) {
					return fmt.Errorf("display list is not supported on this platform yet")
				}
				return err
			}
			if len(items) == 0 {
				_, err := fmt.Fprintln(cmd.OutOrStdout(), "No displays found")
				return err
			}
			out := cmd.OutOrStdout()
			for _, d := range items {
				main := ""
				if d.Main {
					main = " [main]"
				}
				if _, err := fmt.Fprintf(out, "%s%s\n  Resolution: %s\n", d.Name, main, d.Resolution); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
