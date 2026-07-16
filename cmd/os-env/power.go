// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import (
	"errors"
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/power"
	"github.com/spf13/cobra"
)

func newPowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "Power and battery management",
	}
	cmd.AddCommand(newBatteryCmd())
	return cmd
}

func newBatteryCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "battery",
		Short: "Show current battery status",
		Long: `Show the current battery status (source, percent, charging state).

On macOS this reads ` + "`pmset -g batt`" + `. Other platforms are not yet supported.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			status, err := power.GetBattery(cmd.Context())
			if err != nil {
				if errors.Is(err, power.ErrUnsupported) {
					return fmt.Errorf("battery status is not supported on this platform yet")
				}
				return err
			}
			out := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(out, "Source:   %s\n", status.Source); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "Percent:  %d%%\n", status.Percent); err != nil {
				return err
			}
			charging := "no"
			if status.Charging {
				charging = "yes"
			}
			_, err = fmt.Fprintf(out, "Charging: %s\n", charging)
			return err
		},
	}
}
