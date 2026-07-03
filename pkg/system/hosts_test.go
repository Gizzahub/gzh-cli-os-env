// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package system

import "testing"

func TestParseHosts(t *testing.T) {
	in := `# Host Database
#
# 127.0.0.1       commented.example.com
127.0.0.1 localhost
::1 localhost

255.255.255.255 broadcasthost
192.168.1.10 myhost alias.example.com   # inline comment
`
	entries, err := ParseHosts(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	wantCount := 4 // commented line, blanks, and header skipped
	if len(entries) != wantCount {
		t.Fatalf("len(entries) = %d, want %d", len(entries), wantCount)
	}
	if entries[0].IP != "127.0.0.1" {
		t.Errorf("entries[0].IP = %q, want 127.0.0.1", entries[0].IP)
	}
	last := entries[3]
	if last.IP != "192.168.1.10" {
		t.Errorf("last.IP = %q, want 192.168.1.10", last.IP)
	}
	if len(last.Names) != 2 || last.Names[1] != "alias.example.com" {
		t.Errorf("last.Names = %v, want [myhost alias.example.com]", last.Names)
	}
}

func TestParseHosts_Empty(t *testing.T) {
	entries, err := ParseHosts("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("len(entries) = %d, want 0", len(entries))
	}
}

func TestParseHosts_SkipsLoneIP(t *testing.T) {
	entries, err := ParseHosts("10.0.0.1\n127.0.0.1 localhost\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("len(entries) = %d, want 1 (lone IP skipped)", len(entries))
	}
	if entries[0].IP != "127.0.0.1" {
		t.Errorf("entries[0].IP = %q, want 127.0.0.1", entries[0].IP)
	}
}
