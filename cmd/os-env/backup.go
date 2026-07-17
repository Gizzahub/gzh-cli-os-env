// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"fmt"
	"path/filepath"

	"github.com/gizzahub/gzh-cli-os-env/pkg/backup"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create and inspect OS config backups",
	}
	cmd.AddCommand(newBackupCreateCmd(), newBackupRestoreCmd(), newBackupDiffCmd())
	return cmd
}

func newBackupCreateCmd() *cobra.Command {
	var output string
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a portable config snapshot archive",
		Long: `Collect current locale, timezone, keyboard, displays, and shortcuts
into a gzip tar archive (manifest.json inside).`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if output == "" {
				output = "os-config-backup.tar.gz"
			}
			s, err := backup.Create(cmd.Context(), output)
			if err != nil {
				return err
			}
			abs, absErr := filepath.Abs(output)
			if absErr != nil {
				abs = output
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(),
				"Wrote %s (desktop=%s locale=%s timezone=%s)\n",
				abs, s.Desktop, s.Locale, s.Timezone)
			return err
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "", "output archive path (default os-config-backup.tar.gz)")
	return cmd
}

func newBackupRestoreCmd() *cobra.Command {
	var input string
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Show settings from a backup archive (apply not yet implemented)",
		Long: `Load a backup archive and print the stored snapshot.

Phase 6 ships read/inspect restore. Applying settings back to the host
(write paths) is deferred until cross-platform set APIs exist.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input == "" {
				return fmt.Errorf("--input is required")
			}
			s, err := backup.Load(input)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(w, "Version:  %d\n", s.Version); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "Created:  %s\n", s.CreatedAt.Format("2006-01-02T15:04:05Z")); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "Desktop:  %s\n", s.Desktop); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "Locale:   %s\n", s.Locale); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(w, "Timezone: %s\n", s.Timezone); err != nil {
				return err
			}
			if s.Keyboard != nil {
				if _, err := fmt.Fprintf(w, "Keyboard sources: %d\n", len(s.Keyboard.Sources)); err != nil {
					return err
				}
			}
			if _, err := fmt.Fprintf(w, "Displays: %d\n", len(s.Displays)); err != nil {
				return err
			}
			_, err = fmt.Fprintf(w, "Shortcuts: %d\n", len(s.Shortcuts))
			return err
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", "", "input archive path")
	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	return cmd
}

func newBackupDiffCmd() *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Diff a backup archive against the live system",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if from == "" {
				return fmt.Errorf("--from is required")
			}
			archived, err := backup.Load(from)
			if err != nil {
				return err
			}
			live, err := backup.Collect(cmd.Context())
			if err != nil {
				return err
			}
			diffs := backup.Diff(archived, live)
			w := cmd.OutOrStdout()
			if len(diffs) == 0 {
				_, err = fmt.Fprintln(w, "no differences")
				return err
			}
			for _, d := range diffs {
				if _, err := fmt.Fprintln(w, d); err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "backup archive to compare against live system")
	if err := cmd.MarkFlagRequired("from"); err != nil {
		panic(err)
	}
	return cmd
}
