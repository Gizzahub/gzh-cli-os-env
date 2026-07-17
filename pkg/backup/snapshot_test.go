// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package backup

import (
	"bytes"
	"testing"
	"time"
)

func TestEncodeDecodeJSON_RoundTrip(t *testing.T) {
	in := &Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
		Desktop:   "macos",
		Locale:    "en_US.UTF-8",
		Timezone:  "Asia/Seoul",
	}
	data, err := EncodeJSON(in)
	if err != nil {
		t.Fatal(err)
	}
	out, err := DecodeJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if out.Version != 1 || out.Locale != "en_US.UTF-8" || out.Timezone != "Asia/Seoul" {
		t.Fatalf("unexpected decode: %+v", out)
	}
}

func TestDecodeJSON_MissingVersion(t *testing.T) {
	if _, err := DecodeJSON([]byte(`{"locale":"x"}`)); err == nil {
		t.Fatal("expected error for missing version")
	}
}

func TestWriteReadArchive_RoundTrip(t *testing.T) {
	in := &Snapshot{
		Version:   1,
		CreatedAt: time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
		Desktop:   "gnome",
		Locale:    "ko_KR.UTF-8",
		Timezone:  "Asia/Seoul",
	}
	var buf bytes.Buffer
	if err := WriteArchive(&buf, in); err != nil {
		t.Fatal(err)
	}
	out, err := ReadArchive(&buf)
	if err != nil {
		t.Fatal(err)
	}
	if out.Desktop != "gnome" || out.Locale != "ko_KR.UTF-8" {
		t.Fatalf("archive round-trip: %+v", out)
	}
}

func TestDiff(t *testing.T) {
	a := &Snapshot{Version: 1, Locale: "en_US", Timezone: "UTC"}
	b := &Snapshot{Version: 1, Locale: "ko_KR", Timezone: "UTC"}
	d := Diff(a, b)
	if len(d) != 1 || d[0] != `locale: "en_US" -> "ko_KR"` {
		t.Fatalf("diff=%v", d)
	}
	if Diff(nil, b)[0] != "missing snapshot" {
		t.Fatal("expected missing snapshot")
	}
}

func TestCreateLoad_FileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/os-config-backup.tar.gz"
	// Create uses live Collect — may be sparse on CI; still must write valid archive.
	if _, err := Create(t.Context(), path); err != nil {
		t.Fatal(err)
	}
	s, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != 1 {
		t.Fatalf("version=%d", s.Version)
	}
}
