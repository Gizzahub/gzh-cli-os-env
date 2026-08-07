// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package cfgsync mirrors OS config backup archives to a remote store.
//
// Transport is intentionally minimal: a local directory (or a git working
// tree path) holds a single allowlisted archive. Git URLs can be treated as
// directory remotes after the user clones them; pure filesystem remotes keep
// unit tests offline.
//
// Secret policy: only the backup archive payload is transferred, and that
// archive is restricted to the allowlisted entry "manifest.json" (see
// pkg/backup). Tokens, credentials, and env secrets must never be collected
// into the snapshot or placed beside the archive in the remote.
package cfgsync

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gizzahub/gzh-cli-os-env/pkg/backup"
)

const (
	// DefaultArchiveName is the filename stored at the remote root.
	DefaultArchiveName = "os-config-backup.tar.gz"

	// ManifestEntry is the only file allowed inside a sync archive.
	// Secret policy: never add credentials, tokens, or env dumps here.
	ManifestEntry = "manifest.json"
)

// AllowedArchiveEntries returns the secret-policy allowlist for archive
// members. Push/pull refuse archives that contain any other entry.
func AllowedArchiveEntries() []string {
	return []string{ManifestEntry}
}

// Options configure a push or pull operation.
type Options struct {
	// Remote is a local directory used as the sync store (git-friendly path).
	Remote string
	// Archive is the local archive path. For push: source (optional create);
	// for pull: destination. Empty uses DefaultArchiveName under a temp or CWD
	// depending on the operation — callers should set it explicitly.
	Archive string
	// DryRun plans actions without writing files or applying settings.
	DryRun bool
	// Apply, after a successful pull, runs backup.Apply on the loaded snapshot.
	// Ignored by Push and Status.
	Apply bool
	// Writer is forwarded to backup.Apply when Apply is true (tests inject fakes).
	Writer backup.DomainWriter
}

// Result summarizes planned or completed sync actions.
type Result struct {
	// Planned lists human-readable actions (always filled for dry-run; also
	// filled with the intent before execute for real runs).
	Planned []string
	// Done lists actions that were executed (empty when DryRun).
	Done []string
	// RemoteArchive is the resolved path of the remote archive.
	RemoteArchive string
	// LocalArchive is the local archive path used/produced.
	LocalArchive string
	// Snapshot is the loaded snapshot when available (status, pull, validated push).
	Snapshot *backup.Snapshot
	// ApplyResult is set when Options.Apply ran (including dry-run apply plan).
	ApplyResult *backup.ApplyResult
}

// Status reports whether the remote has a valid allowlisted archive.
func Status(remote string) (*Result, error) {
	if err := validateRemote(remote); err != nil {
		return nil, err
	}
	remoteArchive := filepath.Join(remote, DefaultArchiveName)
	res := &Result{RemoteArchive: remoteArchive}

	info, err := os.Stat(remoteArchive)
	if err != nil {
		if os.IsNotExist(err) {
			res.Planned = append(res.Planned, "remote archive missing")
			return res, nil
		}
		return nil, fmt.Errorf("stat remote archive: %w", err)
	}
	if info.IsDir() {
		return nil, fmt.Errorf("remote archive is a directory: %s", remoteArchive)
	}

	if err := ValidateArchiveAllowlist(remoteArchive); err != nil {
		return nil, err
	}
	s, err := backup.Load(remoteArchive)
	if err != nil {
		return nil, fmt.Errorf("load remote archive: %w", err)
	}
	res.Snapshot = s
	res.Planned = append(res.Planned, fmt.Sprintf(
		"remote ready: version=%d locale=%q timezone=%q created=%s",
		s.Version, s.Locale, s.Timezone, s.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
	))
	return res, nil
}

// Push copies a local backup archive into the remote directory.
//
// If opts.Archive is empty, a fresh snapshot is collected into a temp file.
// The archive is always validated against the secret allowlist before copy.
func Push(ctx context.Context, opts Options) (*Result, error) {
	if err := validateRemote(opts.Remote); err != nil {
		return nil, err
	}
	remoteArchive := filepath.Join(opts.Remote, DefaultArchiveName)
	res := &Result{RemoteArchive: remoteArchive}

	localArchive := opts.Archive
	var tmpDir string
	if localArchive == "" {
		if opts.DryRun {
			res.Planned = append(res.Planned, "create snapshot archive (temp)")
			res.LocalArchive = "(new snapshot)"
			res.Planned = append(res.Planned, fmt.Sprintf("copy (new snapshot) -> %s", remoteArchive))
			return res, nil
		}
		var err error
		tmpDir, err = os.MkdirTemp("", "os-env-sync-")
		if err != nil {
			return nil, err
		}
		defer cleanupDir(tmpDir)
		localArchive = filepath.Join(tmpDir, DefaultArchiveName)
		if _, err := backup.Create(ctx, localArchive); err != nil {
			return nil, fmt.Errorf("create snapshot: %w", err)
		}
		res.Done = append(res.Done, "created snapshot archive")
	}
	res.LocalArchive = localArchive

	if err := ValidateArchiveAllowlist(localArchive); err != nil {
		return res, err
	}
	s, err := backup.Load(localArchive)
	if err != nil {
		return res, fmt.Errorf("load local archive: %w", err)
	}
	res.Snapshot = s

	res.Planned = append(res.Planned, fmt.Sprintf("copy %s -> %s", localArchive, remoteArchive))
	if opts.DryRun {
		return res, nil
	}

	if err := os.MkdirAll(opts.Remote, 0o750); err != nil {
		return res, fmt.Errorf("mkdir remote: %w", err)
	}
	if err := copyFile(localArchive, remoteArchive); err != nil {
		return res, err
	}
	res.Done = append(res.Done, fmt.Sprintf("pushed %s", remoteArchive))
	return res, nil
}

// Pull copies the remote archive to a local path and optionally applies it.
//
// Without Apply, only the archive file is placed. With Apply, backup.Apply is
// reused (unit-37). DryRun plans both the copy and apply domains without writes.
func Pull(ctx context.Context, opts Options) (*Result, error) {
	if err := validateRemote(opts.Remote); err != nil {
		return nil, err
	}
	if opts.Archive == "" {
		return nil, errors.New("local archive path required for pull")
	}
	remoteArchive := filepath.Join(opts.Remote, DefaultArchiveName)
	res := &Result{
		RemoteArchive: remoteArchive,
		LocalArchive:  opts.Archive,
	}

	if _, err := os.Stat(remoteArchive); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("remote archive not found: %s", remoteArchive)
		}
		return nil, fmt.Errorf("stat remote archive: %w", err)
	}
	if err := ValidateArchiveAllowlist(remoteArchive); err != nil {
		return nil, err
	}
	s, err := backup.Load(remoteArchive)
	if err != nil {
		return nil, fmt.Errorf("load remote archive: %w", err)
	}
	res.Snapshot = s

	res.Planned = append(res.Planned, fmt.Sprintf("copy %s -> %s", remoteArchive, opts.Archive))
	if opts.Apply {
		res.Planned = append(res.Planned, "apply supported domains from snapshot")
	}

	if opts.DryRun {
		if opts.Apply {
			ar, err := backup.Apply(ctx, s, backup.ApplyOptions{DryRun: true, Writer: opts.Writer})
			res.ApplyResult = ar
			if err != nil {
				return res, err
			}
		}
		return res, nil
	}

	dir := filepath.Dir(opts.Archive)
	if dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return res, fmt.Errorf("mkdir local: %w", err)
		}
	}
	if err := copyFile(remoteArchive, opts.Archive); err != nil {
		return res, err
	}
	res.Done = append(res.Done, fmt.Sprintf("pulled %s", opts.Archive))

	if opts.Apply {
		ar, err := backup.Apply(ctx, s, backup.ApplyOptions{DryRun: false, Writer: opts.Writer})
		res.ApplyResult = ar
		if err != nil {
			return res, err
		}
		res.Done = append(res.Done, "applied snapshot")
	}
	return res, nil
}

// ValidateArchiveAllowlist ensures the gzip-tar only contains allowlisted
// entries (manifest.json). Extra files are rejected so secrets cannot ride
// along in the default payload.
func ValidateArchiveAllowlist(path string) error {
	f, err := os.Open(path) //nolint:gosec // path is caller-supplied archive
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()

	tr := tar.NewReader(gz)
	allowed := map[string]struct{}{}
	for _, name := range AllowedArchiveEntries() {
		allowed[name] = struct{}{}
	}
	foundManifest := false
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("tar: %w", err)
		}
		name := filepath.Base(hdr.Name)
		// Reject absolute paths / parent traversal in member names.
		if strings.Contains(hdr.Name, "..") || filepath.IsAbs(hdr.Name) {
			return fmt.Errorf("secret policy: disallowed archive path %q", hdr.Name)
		}
		if _, ok := allowed[name]; !ok {
			return fmt.Errorf("secret policy: archive entry %q not in allowlist %v", hdr.Name, AllowedArchiveEntries())
		}
		if name == ManifestEntry {
			foundManifest = true
		}
	}
	if !foundManifest {
		return errors.New("secret policy: manifest.json missing from archive")
	}
	return nil
}

func validateRemote(remote string) error {
	if strings.TrimSpace(remote) == "" {
		return errors.New("remote path is required")
	}
	// Refuse network-looking remotes in this minimal transport; callers use a
	// local clone path so tests stay offline and no tokens hit the wire here.
	if isNetworkRemote(remote) {
		return fmt.Errorf("network remotes are not supported directly; clone to a local path first: %s", remote)
	}
	return nil
}

func isNetworkRemote(remote string) bool {
	r := strings.ToLower(remote)
	if strings.HasPrefix(r, "http://") || strings.HasPrefix(r, "https://") ||
		strings.HasPrefix(r, "git://") || strings.HasPrefix(r, "ssh://") {
		return true
	}
	// git@host:path
	if strings.HasPrefix(remote, "git@") {
		return true
	}
	return false
}

func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // caller-supplied paths
	if err != nil {
		return fmt.Errorf("open source: %w", err)
	}
	defer func() { _ = in.Close() }()

	// 0o600: sync archives may hold personal desktop prefs — owner-only.
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600) //nolint:gosec // caller-supplied dest path
	if err != nil {
		return fmt.Errorf("create dest: %w", err)
	}
	defer func() { _ = out.Close() }()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy: %w", err)
	}
	return out.Close()
}

// cleanupDir best-effort removes a temp directory created for push snapshots.
func cleanupDir(dir string) {
	if dir == "" {
		return
	}
	err := os.RemoveAll(dir)
	if err == nil || os.IsNotExist(err) {
		return
	}
	// Non-fatal: temp cleanup failure must not mask a successful push.
}
