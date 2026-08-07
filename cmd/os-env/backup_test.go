// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-os-env/pkg/backup"
)

func writeTestArchive(t *testing.T, path string, s *backup.Snapshot) {
	t.Helper()
	f, err := os.Create(path) //nolint:gosec // test fixture under t.TempDir
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := backup.WriteArchive(f, s); err != nil {
		t.Fatal(err)
	}
}

func TestRestore_InspectOnly(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.tar.gz")
	s := &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		Desktop:   "macos",
		Locale:    "en_US.UTF-8",
		Timezone:  "Asia/Seoul",
	}
	writeTestArchive(t, path, s)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"backup", "restore", "--input", path})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("restore inspect: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Locale:   en_US.UTF-8") {
		t.Fatalf("inspect output missing locale: %q", got)
	}
	if strings.Contains(got, "Applied:") || strings.Contains(got, "Planned:") {
		t.Fatalf("inspect must not apply: %q", got)
	}
}

func TestRestore_DryRun_PlansDomains(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.tar.gz")
	s := &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		Locale:    "ko_KR.UTF-8",
		Timezone:  "Asia/Seoul",
	}
	writeTestArchive(t, path, s)

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"backup", "restore", "--input", path, "--apply", "--dry-run"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("restore dry-run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Planned: locale") {
		t.Fatalf("expected planned locale: %q", got)
	}
	if !strings.Contains(got, "Planned: timezone") {
		t.Fatalf("expected planned timezone: %q", got)
	}
	if !strings.Contains(got, "Skipped: keyboard") {
		t.Fatalf("expected keyboard skip reason: %q", got)
	}
	if strings.Contains(got, "Applied:") {
		t.Fatalf("dry-run must not report Applied: %q", got)
	}
}

func TestRestore_DryRun_RequiresApply(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snap.tar.gz")
	writeTestArchive(t, path, &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		Locale:    "C",
	})

	cmd := NewRootCmd()
	cmd.SetArgs([]string{"backup", "restore", "--input", path, "--dry-run"})
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error when --dry-run without --apply")
	}
}

func TestBackupRestore_HelpMentionsApply(t *testing.T) {
	cmd := NewRootCmd()
	var found bool
	for _, c := range cmd.Commands() {
		if c.Name() != "backup" {
			continue
		}
		for _, sub := range c.Commands() {
			if sub.Name() != "restore" {
				continue
			}
			found = true
			if sub.Flags().Lookup("apply") == nil {
				t.Fatal("restore missing --apply flag")
			}
			if sub.Flags().Lookup("dry-run") == nil {
				t.Fatal("restore missing --dry-run flag")
			}
			if !strings.Contains(sub.Long, "dry-run") {
				t.Fatalf("restore Long should document dry-run: %q", sub.Long)
			}
		}
	}
	if !found {
		t.Fatal("backup restore command not found")
	}
}
