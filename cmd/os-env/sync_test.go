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
	"github.com/gizzahub/gzh-cli-os-env/pkg/cfgsync"
)

func TestSync_CLI_PushPullStatus(t *testing.T) {
	remote := t.TempDir()
	src := filepath.Join(t.TempDir(), "src.tar.gz")
	writeTestArchive(t, src, &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		Locale:    "ko_KR.UTF-8",
		Timezone:  "Asia/Seoul",
		Desktop:   "macos",
	})

	// status missing
	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "status", "--remote", remote})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status missing: %v", err)
	}
	if !strings.Contains(out.String(), "missing") {
		t.Fatalf("expected missing status: %q", out.String())
	}

	// push
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "push", "--remote", remote, "--input", src})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("push: %v", err)
	}
	if !strings.Contains(out.String(), "Done:") {
		t.Fatalf("expected Done output: %q", out.String())
	}
	if _, err := os.Stat(filepath.Join(remote, cfgsync.DefaultArchiveName)); err != nil {
		t.Fatalf("remote archive missing after push: %v", err)
	}

	// status ready
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "status", "--remote", remote})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("status ready: %v", err)
	}
	if !strings.Contains(out.String(), "Locale:   ko_KR.UTF-8") {
		t.Fatalf("status inspect: %q", out.String())
	}

	// pull dry-run
	dst := filepath.Join(t.TempDir(), "pulled.tar.gz")
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "pull", "--remote", remote, "--output", dst, "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pull dry-run: %v", err)
	}
	if !strings.Contains(out.String(), "Planned:") {
		t.Fatalf("expected Planned: %q", out.String())
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatal("dry-run must not create local archive")
	}

	// pull
	cmd = NewRootCmd()
	out = &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "pull", "--remote", remote, "--output", dst})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("pull: %v", err)
	}
	loaded, err := backup.Load(dst)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Timezone != "Asia/Seoul" {
		t.Fatalf("pulled timezone: %q", loaded.Timezone)
	}
}

func TestSync_CLI_PushDryRun(t *testing.T) {
	remote := t.TempDir()
	src := filepath.Join(t.TempDir(), "src.tar.gz")
	writeTestArchive(t, src, &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 0, 0, 0, 0, time.UTC),
		Locale:    "C",
	})

	cmd := NewRootCmd()
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{"sync", "push", "--remote", remote, "--input", src, "--dry-run"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("push dry-run: %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Planned:") {
		t.Fatalf("expected Planned: %q", got)
	}
	if strings.Contains(got, "Done:") {
		t.Fatalf("dry-run must not Done: %q", got)
	}
	if _, err := os.Stat(filepath.Join(remote, cfgsync.DefaultArchiveName)); !os.IsNotExist(err) {
		t.Fatal("dry-run must not write remote")
	}
}

func TestSync_CLI_HelpMentionsRemote(t *testing.T) {
	cmd := NewRootCmd()
	var found bool
	for _, c := range cmd.Commands() {
		if c.Name() != "sync" {
			continue
		}
		found = true
		if !strings.Contains(c.Long, "Secret policy") {
			t.Fatalf("sync Long should document secret policy: %q", c.Long)
		}
		names := map[string]bool{}
		for _, sub := range c.Commands() {
			names[sub.Name()] = true
			if sub.Flags().Lookup("remote") == nil {
				t.Fatalf("%s missing --remote", sub.Name())
			}
		}
		for _, want := range []string{"status", "push", "pull"} {
			if !names[want] {
				t.Fatalf("missing subcommand %s", want)
			}
		}
	}
	if !found {
		t.Fatal("sync command not found")
	}
}
