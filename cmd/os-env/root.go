// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import "github.com/spf13/cobra"

// NewRootCmd creates the root command for OS / desktop environment management.
// Designed for direct use or wrapping by a parent CLI (e.g. gzh-cli).
func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "os-env",
		Short: "Manage OS and desktop environment configurations",
		Long: `Manage desktop environment and OS-level settings across platforms.

Provides unified management of:
- Desktop environments (KDE Plasma, GNOME, macOS, Windows)
- Power & battery
- System settings (hosts, locale, timezone, paths)
- Input devices (keyboard, mouse, touchpad)
- Display settings

Examples:
  # Detect the current desktop environment
  os-env detect`,
		SilenceUsage: true,
	}

	cmd.AddCommand(newDetectCmd())
	cmd.AddCommand(newPowerCmd())
	cmd.AddCommand(newSystemCmd())
	cmd.AddCommand(newDisplayCmd())

	return cmd
}
