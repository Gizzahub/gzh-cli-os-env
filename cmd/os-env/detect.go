// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/detector"
	"github.com/spf13/cobra"
)

func newDetectCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detect",
		Short: "Detect the current OS and desktop environment",
		Long: `Detect and print the current operating system and desktop environment
(KDE Plasma, GNOME, macOS, or Windows).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			info, err := detector.Detect()
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "OS:       %s\n", info.OS); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(cmd.OutOrStdout(), "Desktop:  %s\n", info.Desktop); err != nil {
				return err
			}
			return nil
		},
	}
}
