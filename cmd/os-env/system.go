// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gizzahub/gzh-cli-os-env/pkg/system"
	"github.com/spf13/cobra"
)

func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "System settings (hosts, locale, timezone)",
	}
	cmd.AddCommand(newSystemHostsCmd(), newSystemLocaleCmd(), newSystemTimezoneCmd())
	return cmd
}

func newSystemHostsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hosts",
		Short: "List /etc/hosts entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			entries, err := system.GetHosts("/etc/hosts")
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			for _, e := range entries {
				if _, err := fmt.Fprintf(w, "%s\t%s\n", e.IP, strings.Join(e.Names, " ")); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

func newSystemLocaleCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "locale",
		Short: "Show the current locale",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			loc, err := system.GetLocale(cmd.Context())
			if err != nil {
				if errors.Is(err, system.ErrLocaleUnsupported) {
					return fmt.Errorf("locale is not supported on this platform yet")
				}
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), loc)
			return err
		},
	}
}

func newSystemTimezoneCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "timezone",
		Short: "Show the current timezone",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			tz, err := system.GetTimezone(cmd.Context())
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(cmd.OutOrStdout(), tz)
			return err
		},
	}
}
