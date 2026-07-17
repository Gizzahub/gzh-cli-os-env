// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package backup creates and restores portable OS/desktop config snapshots.
//
// A snapshot is a gzip-compressed tar archive containing a single
// manifest.json with collected settings (locale, timezone, keyboard,
// displays, shortcuts). Collectors run best-effort: missing platform
// support yields empty fields, not a hard failure.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/gizzahub/gzh-cli-os-env/pkg/detector"
	"github.com/gizzahub/gzh-cli-os-env/pkg/display"
	"github.com/gizzahub/gzh-cli-os-env/pkg/input"
	"github.com/gizzahub/gzh-cli-os-env/pkg/shortcuts"
	"github.com/gizzahub/gzh-cli-os-env/pkg/system"
)

const (
	manifestName    = "manifest.json"
	snapshotVersion = 1
)

// Snapshot is the portable config payload stored in a backup archive.
type Snapshot struct {
	Version   int                  `json:"version"`
	CreatedAt time.Time            `json:"created_at"`
	Desktop   string               `json:"desktop,omitempty"`
	Locale    string               `json:"locale,omitempty"`
	Timezone  string               `json:"timezone,omitempty"`
	Keyboard  *input.Keyboard      `json:"keyboard,omitempty"`
	Displays  []display.Info       `json:"displays,omitempty"`
	Shortcuts []shortcuts.Shortcut `json:"shortcuts,omitempty"`
}

// Collect gathers a best-effort snapshot of current OS settings.
func Collect(ctx context.Context) (*Snapshot, error) {
	s := &Snapshot{
		Version:   snapshotVersion,
		CreatedAt: time.Now().UTC(),
	}
	if info, err := detector.Detect(); err == nil {
		s.Desktop = info.Desktop
		if s.Desktop == "" {
			s.Desktop = info.OS
		}
	}
	_ = ctx // reserved for collector cancellation
	if loc, err := system.GetLocale(ctx); err == nil {
		s.Locale = loc
	}
	if tz, err := system.GetTimezone(ctx); err == nil {
		s.Timezone = tz
	}
	if kb, err := input.GetKeyboard(ctx); err == nil {
		s.Keyboard = kb
	}
	if displays, err := display.List(ctx); err == nil {
		s.Displays = displays
	}
	if sc, err := shortcuts.List(ctx); err == nil {
		s.Shortcuts = sc
	}
	return s, nil
}

// EncodeJSON returns the canonical JSON encoding of s.
func EncodeJSON(s *Snapshot) ([]byte, error) {
	if s == nil {
		return nil, errors.New("nil snapshot")
	}
	return json.MarshalIndent(s, "", "  ")
}

// DecodeJSON parses a snapshot from JSON bytes.
func DecodeJSON(data []byte) (*Snapshot, error) {
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	if s.Version == 0 {
		return nil, errors.New("snapshot missing version")
	}
	return &s, nil
}

// WriteArchive writes s as a gzip tar archive to w.
func WriteArchive(w io.Writer, s *Snapshot) error {
	data, err := EncodeJSON(s)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(w)
	tw := tar.NewWriter(gz)
	hdr := &tar.Header{
		Name:    manifestName,
		Mode:    0o644,
		Size:    int64(len(data)),
		ModTime: s.CreatedAt,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return err
	}
	if _, err := tw.Write(data); err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return err
	}
	if err := tw.Close(); err != nil {
		_ = gz.Close()
		return err
	}
	return gz.Close()
}

// ReadArchive reads a snapshot from a gzip tar archive.
func ReadArchive(r io.Reader) (*Snapshot, error) {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("gzip: %w", err)
	}
	defer func() { _ = gz.Close() }()
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("tar: %w", err)
		}
		if filepath.Base(hdr.Name) != manifestName {
			continue
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		return DecodeJSON(data)
	}
	return nil, errors.New("manifest.json not found in archive")
}

// Create writes a new backup archive to path (parent dirs created as needed).
func Create(ctx context.Context, path string) (*Snapshot, error) {
	s, err := Collect(ctx)
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(path)
	if dir != "." && dir != "" {
		// 0o750: owner/group only for backup parent dirs
		if err := os.MkdirAll(dir, 0o750); err != nil {
			return nil, err
		}
	}
	f, err := os.Create(path) //nolint:gosec // path is user-supplied backup destination
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	if err := WriteArchive(f, s); err != nil {
		return nil, err
	}
	return s, nil
}

// Load reads a backup archive from path.
func Load(path string) (*Snapshot, error) {
	f, err := os.Open(path) //nolint:gosec // path is user-supplied backup source
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	return ReadArchive(f)
}

// Diff summarizes field-level differences between two snapshots (pure).
func Diff(from, to *Snapshot) []string {
	if from == nil || to == nil {
		return []string{"missing snapshot"}
	}
	var out []string
	if from.Locale != to.Locale {
		out = append(out, fmt.Sprintf("locale: %q -> %q", from.Locale, to.Locale))
	}
	if from.Timezone != to.Timezone {
		out = append(out, fmt.Sprintf("timezone: %q -> %q", from.Timezone, to.Timezone))
	}
	if from.Desktop != to.Desktop {
		out = append(out, fmt.Sprintf("desktop: %q -> %q", from.Desktop, to.Desktop))
	}
	fromN, toN := 0, 0
	if from.Keyboard != nil {
		fromN = len(from.Keyboard.Sources)
	}
	if to.Keyboard != nil {
		toN = len(to.Keyboard.Sources)
	}
	if fromN != toN {
		out = append(out, fmt.Sprintf("keyboard sources: %d -> %d", fromN, toN))
	}
	if len(from.Displays) != len(to.Displays) {
		out = append(out, fmt.Sprintf("displays: %d -> %d", len(from.Displays), len(to.Displays)))
	}
	if len(from.Shortcuts) != len(to.Shortcuts) {
		out = append(out, fmt.Sprintf("shortcuts: %d -> %d", len(from.Shortcuts), len(to.Shortcuts)))
	}
	return out
}
