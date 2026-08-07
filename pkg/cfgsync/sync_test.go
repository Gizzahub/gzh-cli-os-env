// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package cfgsync

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gizzahub/gzh-cli-os-env/pkg/backup"
)

func writeArchive(t *testing.T, path string, s *backup.Snapshot) {
	t.Helper()
	f, err := os.Create(path) //nolint:gosec // test path under TempDir
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	if err := backup.WriteArchive(f, s); err != nil {
		t.Fatal(err)
	}
}

func sampleSnapshot() *backup.Snapshot {
	return &backup.Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC),
		Desktop:   "macos",
		Locale:    "en_US.UTF-8",
		Timezone:  "Asia/Seoul",
	}
}

func TestSync_PushPlacesArchive(t *testing.T) {
	remote := t.TempDir()
	local := filepath.Join(t.TempDir(), "snap.tar.gz")
	writeArchive(t, local, sampleSnapshot())

	res, err := Push(context.Background(), Options{
		Remote:  remote,
		Archive: local,
	})
	if err != nil {
		t.Fatalf("push: %v", err)
	}
	remotePath := filepath.Join(remote, DefaultArchiveName)
	if res.RemoteArchive != remotePath {
		t.Fatalf("remote archive path: got %q want %q", res.RemoteArchive, remotePath)
	}
	if _, err := os.Stat(remotePath); err != nil {
		t.Fatalf("expected remote archive: %v", err)
	}
	loaded, err := backup.Load(remotePath)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Locale != "en_US.UTF-8" || loaded.Timezone != "Asia/Seoul" {
		t.Fatalf("unexpected remote snapshot: %+v", loaded)
	}
}

func TestSync_PullLoadsArchive(t *testing.T) {
	remote := t.TempDir()
	writeArchive(t, filepath.Join(remote, DefaultArchiveName), sampleSnapshot())

	local := filepath.Join(t.TempDir(), "pulled.tar.gz")
	res, err := Pull(context.Background(), Options{
		Remote:  remote,
		Archive: local,
	})
	if err != nil {
		t.Fatalf("pull: %v", err)
	}
	if res.Snapshot == nil || res.Snapshot.Locale != "en_US.UTF-8" {
		t.Fatalf("snapshot not loaded: %+v", res.Snapshot)
	}
	if _, err := os.Stat(local); err != nil {
		t.Fatalf("local archive missing: %v", err)
	}
	loaded, err := backup.Load(local)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.Timezone != "Asia/Seoul" {
		t.Fatalf("pulled timezone: %q", loaded.Timezone)
	}
}

func TestSync_PushPull_RoundTrip(t *testing.T) {
	remote := t.TempDir()
	src := filepath.Join(t.TempDir(), "src.tar.gz")
	writeArchive(t, src, sampleSnapshot())

	if _, err := Push(context.Background(), Options{Remote: remote, Archive: src}); err != nil {
		t.Fatalf("push: %v", err)
	}
	dst := filepath.Join(t.TempDir(), "dst.tar.gz")
	if _, err := Pull(context.Background(), Options{Remote: remote, Archive: dst}); err != nil {
		t.Fatalf("pull: %v", err)
	}
	a, _ := backup.Load(src)
	b, _ := backup.Load(dst)
	if a.Locale != b.Locale || a.Timezone != b.Timezone || a.Version != b.Version {
		t.Fatalf("round-trip mismatch: %+v vs %+v", a, b)
	}
}

func TestSync_Status_ReadyAndMissing(t *testing.T) {
	remote := t.TempDir()
	st, err := Status(remote)
	if err != nil {
		t.Fatalf("status empty: %v", err)
	}
	if st.Snapshot != nil {
		t.Fatal("expected no snapshot when missing")
	}
	if len(st.Planned) == 0 || !strings.Contains(st.Planned[0], "missing") {
		t.Fatalf("expected missing plan: %v", st.Planned)
	}

	writeArchive(t, filepath.Join(remote, DefaultArchiveName), sampleSnapshot())
	st, err = Status(remote)
	if err != nil {
		t.Fatalf("status ready: %v", err)
	}
	if st.Snapshot == nil || st.Snapshot.Locale != "en_US.UTF-8" {
		t.Fatalf("status snapshot: %+v", st.Snapshot)
	}
	if !strings.Contains(strings.Join(st.Planned, " "), "remote ready") {
		t.Fatalf("expected ready message: %v", st.Planned)
	}
}

func TestSync_DryRun_NoWrite(t *testing.T) {
	remote := t.TempDir()
	local := filepath.Join(t.TempDir(), "snap.tar.gz")
	writeArchive(t, local, sampleSnapshot())

	res, err := Push(context.Background(), Options{
		Remote:  remote,
		Archive: local,
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("push dry-run: %v", err)
	}
	if len(res.Done) != 0 {
		t.Fatalf("dry-run must not Done: %v", res.Done)
	}
	if _, err := os.Stat(filepath.Join(remote, DefaultArchiveName)); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not create remote archive, stat err=%v", err)
	}
	if len(res.Planned) == 0 {
		t.Fatal("expected planned actions")
	}

	writeArchive(t, filepath.Join(remote, DefaultArchiveName), sampleSnapshot())
	dst := filepath.Join(t.TempDir(), "out.tar.gz")
	res, err = Pull(context.Background(), Options{
		Remote:  remote,
		Archive: dst,
		DryRun:  true,
	})
	if err != nil {
		t.Fatalf("pull dry-run: %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Fatalf("pull dry-run must not write local archive")
	}
	if !strings.Contains(strings.Join(res.Planned, " "), "copy") {
		t.Fatalf("expected copy plan: %v", res.Planned)
	}
}

func TestSync_Pull_ApplyDryRun(t *testing.T) {
	remote := t.TempDir()
	writeArchive(t, filepath.Join(remote, DefaultArchiveName), sampleSnapshot())
	dst := filepath.Join(t.TempDir(), "out.tar.gz")

	w := &fakeWriter{}
	res, err := Pull(context.Background(), Options{
		Remote:  remote,
		Archive: dst,
		Apply:   true,
		DryRun:  true,
		Writer:  w,
	})
	if err != nil {
		t.Fatalf("pull apply dry-run: %v", err)
	}
	if w.locale != "" || w.tz != "" {
		t.Fatalf("dry-run must not call writer: %+v", w)
	}
	if res.ApplyResult == nil {
		t.Fatal("expected ApplyResult")
	}
	if len(res.ApplyResult.Planned) == 0 {
		t.Fatalf("expected planned domains: %+v", res.ApplyResult)
	}
}

func TestSync_Pull_ApplyViaFakeWriter(t *testing.T) {
	remote := t.TempDir()
	writeArchive(t, filepath.Join(remote, DefaultArchiveName), sampleSnapshot())
	dst := filepath.Join(t.TempDir(), "out.tar.gz")

	w := &fakeWriter{}
	res, err := Pull(context.Background(), Options{
		Remote:  remote,
		Archive: dst,
		Apply:   true,
		Writer:  w,
	})
	if err != nil {
		t.Fatalf("pull apply: %v", err)
	}
	if w.locale != "en_US.UTF-8" || w.tz != "Asia/Seoul" {
		t.Fatalf("writer not applied: %+v", w)
	}
	if res.ApplyResult == nil || len(res.ApplyResult.Applied) == 0 {
		t.Fatalf("expected applied domains: %+v", res.ApplyResult)
	}
	if _, err := os.Stat(dst); err != nil {
		t.Fatalf("archive should still be placed: %v", err)
	}
}

func TestSync_SecretsNotInDefaultPayload(t *testing.T) {
	// Default archive allowlist is manifest.json only; extra entries rejected.
	dir := t.TempDir()
	bad := filepath.Join(dir, "bad.tar.gz")
	if err := writeArchiveWithExtra(bad, sampleSnapshot(), "secrets.env", []byte("TOKEN=super-secret\n")); err != nil {
		t.Fatal(err)
	}
	if err := ValidateArchiveAllowlist(bad); err == nil {
		t.Fatal("expected allowlist rejection for secrets.env")
	} else if !strings.Contains(err.Error(), "secret policy") {
		t.Fatalf("expected secret policy error: %v", err)
	}

	// Clean backup archive is accepted and contains no secret-like fields.
	good := filepath.Join(dir, "good.tar.gz")
	writeArchive(t, good, sampleSnapshot())
	if err := ValidateArchiveAllowlist(good); err != nil {
		t.Fatalf("clean archive rejected: %v", err)
	}
	data, err := os.ReadFile(good) //nolint:gosec
	if err != nil {
		t.Fatal(err)
	}
	// Raw archive bytes should not contain common token markers from fixture.
	if strings.Contains(string(data), "TOKEN=") || strings.Contains(string(data), "super-secret") {
		t.Fatal("archive payload contains secret marker")
	}
	// Snapshot JSON fields are allowlisted config only (no credential keys).
	s, err := backup.Load(good)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := backup.EncodeJSON(s)
	if err != nil {
		t.Fatal(err)
	}
	lower := strings.ToLower(string(raw))
	for _, banned := range []string{"token", "password", "secret", "credential", "api_key"} {
		if strings.Contains(lower, banned) {
			t.Fatalf("snapshot JSON contains banned key fragment %q: %s", banned, raw)
		}
	}
}

func TestSync_PushRejectsDisallowedArchive(t *testing.T) {
	remote := t.TempDir()
	bad := filepath.Join(t.TempDir(), "bad.tar.gz")
	if err := writeArchiveWithExtra(bad, sampleSnapshot(), "tokens.json", []byte(`{"token":"x"}`)); err != nil {
		t.Fatal(err)
	}
	_, err := Push(context.Background(), Options{Remote: remote, Archive: bad})
	if err == nil || !strings.Contains(err.Error(), "secret policy") {
		t.Fatalf("expected secret policy error, got %v", err)
	}
	if _, err := os.Stat(filepath.Join(remote, DefaultArchiveName)); !os.IsNotExist(err) {
		t.Fatal("rejected push must not place remote archive")
	}
}

func TestSync_NetworkRemoteRejected(t *testing.T) {
	_, err := Status("https://example.com/repo.git")
	if err == nil || !strings.Contains(err.Error(), "network remotes") {
		t.Fatalf("expected network remote rejection: %v", err)
	}
	_, err = Push(context.Background(), Options{
		Remote:  "git@github.com:org/repo.git",
		Archive: "x",
	})
	if err == nil || !strings.Contains(err.Error(), "network remotes") {
		t.Fatalf("expected network remote rejection: %v", err)
	}
}

func TestSync_AllowedArchiveEntries(t *testing.T) {
	got := AllowedArchiveEntries()
	if len(got) != 1 || got[0] != ManifestEntry {
		t.Fatalf("allowlist: %v", got)
	}
}

type fakeWriter struct {
	locale string
	tz     string
}

func (f *fakeWriter) SetLocale(_ context.Context, locale string) error {
	f.locale = locale
	return nil
}

func (f *fakeWriter) SetTimezone(_ context.Context, tz string) error {
	f.tz = tz
	return nil
}

func writeArchiveWithExtra(path string, s *backup.Snapshot, extraName string, extraData []byte) error {
	f, err := os.Create(path) //nolint:gosec
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	manifest, err := backup.EncodeJSON(s)
	if err != nil {
		return err
	}
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)
	for _, ent := range []struct {
		name string
		data []byte
	}{
		{ManifestEntry, manifest},
		{extraName, extraData},
	} {
		hdr := &tar.Header{
			Name:    ent.name,
			Mode:    0o644,
			Size:    int64(len(ent.data)),
			ModTime: s.CreatedAt,
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(ent.data); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}
