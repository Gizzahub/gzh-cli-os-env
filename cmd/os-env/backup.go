// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/gizzahub/gzh-cli-os-env/pkg/backup"
	"github.com/spf13/cobra"
)

func newBackupCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Create and restore OS config backups",
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
	var (
		input  string
		apply  bool
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Inspect or apply settings from a backup archive",
		Long: `Load a backup archive and either print the stored snapshot (default)
or apply supported domains back to the host.

Without --apply the command is inspect-only.

With --apply, locale and timezone are written via platform set APIs
(macOS/Linux). keyboard, displays, and shortcuts are skipped with an
explicit reason until set APIs exist. Windows set paths return
unsupported errors from pkg/system.

Use --apply --dry-run to print the planned domains without writing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if input == "" {
				return fmt.Errorf("--input is required")
			}
			if dryRun && !apply {
				return fmt.Errorf("--dry-run requires --apply")
			}
			s, err := backup.Load(input)
			if err != nil {
				return err
			}
			w := cmd.OutOrStdout()
			if !apply {
				return printSnapshotInspect(w, s)
			}
			result, err := backup.Apply(cmd.Context(), s, backup.ApplyOptions{DryRun: dryRun})
			if result != nil {
				if printErr := printApplyResult(w, result, dryRun); printErr != nil {
					return printErr
				}
			}
			return err
		},
	}
	cmd.Flags().StringVarP(&input, "input", "i", "", "input archive path")
	cmd.Flags().BoolVar(&apply, "apply", false, "apply supported domains to the host")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "with --apply, plan domains without writing")
	if err := cmd.MarkFlagRequired("input"); err != nil {
		panic(err)
	}
	return cmd
}

func printSnapshotInspect(w io.Writer, s *backup.Snapshot) error {
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
	_, err := fmt.Fprintf(w, "Shortcuts: %d\n", len(s.Shortcuts))
	return err
}

func printApplyResult(w io.Writer, result *backup.ApplyResult, dryRun bool) error {
	label, domains := "Applied", result.Applied
	if dryRun {
		label, domains = "Planned", result.Planned
	}
	if err := printDomains(w, label, domains); err != nil {
		return err
	}
	for _, sk := range result.Skipped {
		if _, err := fmt.Fprintf(w, "Skipped: %s (%s)\n", sk.Domain, sk.Reason); err != nil {
			return err
		}
	}
	return nil
}

// printDomains writes one "<label>: <domain>" line per domain, or
// "<label>: (none)" when there are none.
func printDomains(w io.Writer, label string, domains []string) error {
	if len(domains) == 0 {
		_, err := fmt.Fprintf(w, "%s: (none)\n", label)
		return err
	}
	for _, d := range domains {
		if _, err := fmt.Fprintf(w, "%s: %s\n", label, d); err != nil {
			return err
		}
	}
	return nil
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
