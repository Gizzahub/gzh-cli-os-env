// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"fmt"
	"io"

	"github.com/gizzahub/gzh-cli-os-env/pkg/cfgsync"
	"github.com/spf13/cobra"
)

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Push/pull OS config snapshots to a remote store",
		Long: `Mirror portable config backup archives to a local remote directory.

The remote is a filesystem path (git working tree friendly). Clone a git
repository yourself, then pass the local path as --remote.

Secret policy: only the backup archive (manifest.json allowlist) is
transferred. Tokens and credentials are never part of the default payload.`,
	}
	cmd.AddCommand(newSyncStatusCmd(), newSyncPushCmd(), newSyncPullCmd())
	return cmd
}

func newSyncStatusCmd() *cobra.Command {
	var remote string
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show remote archive presence and snapshot summary",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := cfgsync.Status(remote)
			if err != nil {
				return err
			}
			return printSyncStatus(cmd.OutOrStdout(), res)
		},
	}
	cmd.Flags().StringVar(&remote, "remote", "", "local remote directory (required)")
	if err := cmd.MarkFlagRequired("remote"); err != nil {
		panic(err)
	}
	return cmd
}

func newSyncPushCmd() *cobra.Command {
	var (
		remote string
		input  string
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "push",
		Short: "Push a config archive to the remote store",
		Long: `Copy a backup archive into --remote as os-config-backup.tar.gz.

Without --input, a fresh snapshot is collected first. Archives are validated
against the secret allowlist (manifest.json only) before copy.

Use --dry-run to print planned actions without writing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			res, err := cfgsync.Push(cmd.Context(), cfgsync.Options{
				Remote:  remote,
				Archive: input,
				DryRun:  dryRun,
			})
			if err != nil {
				return err
			}
			return printSyncResult(cmd.OutOrStdout(), res, dryRun)
		},
	}
	cmd.Flags().StringVar(&remote, "remote", "", "local remote directory (required)")
	cmd.Flags().StringVarP(&input, "input", "i", "", "local archive to push (default: create fresh snapshot)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "plan actions without writing")
	if err := cmd.MarkFlagRequired("remote"); err != nil {
		panic(err)
	}
	return cmd
}

func newSyncPullCmd() *cobra.Command {
	var (
		remote string
		output string
		apply  bool
		dryRun bool
	)
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Pull a config archive from the remote store",
		Long: `Copy os-config-backup.tar.gz from --remote to a local path.

Without --apply the command only places the archive file.
With --apply, supported domains are written via backup.Apply (locale/timezone).
Use --apply --dry-run to plan copy + apply without writing.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if output == "" {
				output = cfgsync.DefaultArchiveName
			}
			res, err := cfgsync.Pull(cmd.Context(), cfgsync.Options{
				Remote:  remote,
				Archive: output,
				Apply:   apply,
				DryRun:  dryRun,
			})
			if printErr := printSyncResult(cmd.OutOrStdout(), res, dryRun); printErr != nil {
				return printErr
			}
			if res != nil && res.ApplyResult != nil {
				if printErr := printApplyResult(cmd.OutOrStdout(), res.ApplyResult, dryRun); printErr != nil {
					return printErr
				}
			}
			return err
		},
	}
	cmd.Flags().StringVar(&remote, "remote", "", "local remote directory (required)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "local destination archive (default os-config-backup.tar.gz)")
	cmd.Flags().BoolVar(&apply, "apply", false, "after pull, apply supported domains to the host")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "plan actions without writing")
	if err := cmd.MarkFlagRequired("remote"); err != nil {
		panic(err)
	}
	return cmd
}

func printSyncStatus(w io.Writer, res *cfgsync.Result) error {
	if res == nil {
		return nil
	}
	if _, err := fmt.Fprintf(w, "Remote: %s\n", res.RemoteArchive); err != nil {
		return err
	}
	if res.Snapshot == nil {
		for _, p := range res.Planned {
			if _, err := fmt.Fprintf(w, "Status: %s\n", p); err != nil {
				return err
			}
		}
		return nil
	}
	return printSnapshotInspect(w, res.Snapshot)
}

func printSyncResult(w io.Writer, res *cfgsync.Result, dryRun bool) error {
	if res == nil {
		return nil
	}
	if dryRun {
		if len(res.Planned) == 0 {
			if _, err := fmt.Fprintln(w, "Planned: (none)"); err != nil {
				return err
			}
		} else {
			for _, p := range res.Planned {
				if _, err := fmt.Fprintf(w, "Planned: %s\n", p); err != nil {
					return err
				}
			}
		}
		return nil
	}
	for _, d := range res.Done {
		if _, err := fmt.Fprintf(w, "Done: %s\n", d); err != nil {
			return err
		}
	}
	if res.Snapshot != nil && len(res.Done) == 0 {
		// Status-like: no done, show inspect when present.
		return printSnapshotInspect(w, res.Snapshot)
	}
	return nil
}
