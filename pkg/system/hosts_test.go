// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import (
	"runtime"
	"strings"
	"testing"
)

func TestDefaultHostsPath_CurrentGOOS(t *testing.T) {
	path := DefaultHostsPath()
	if path == "" {
		t.Fatal("DefaultHostsPath() returned empty string")
	}
	switch runtime.GOOS {
	case goosWindows:
		if !strings.HasSuffix(strings.ToLower(path), `\system32\drivers\etc\hosts`) {
			t.Errorf("DefaultHostsPath() = %q, want ...\\System32\\drivers\\etc\\hosts", path)
		}
	default:
		if path != "/etc/hosts" {
			t.Errorf("DefaultHostsPath() = %q, want /etc/hosts", path)
		}
	}
}

func TestDefaultHostsPath_WindowsWithSystemRoot(t *testing.T) {
	got := hostsPathFor(goosWindows, `D:\WinRoot`)
	want := `D:\WinRoot\System32\drivers\etc\hosts`
	if got != want {
		t.Errorf("hostsPathFor(windows, D:\\WinRoot) = %q, want %q", got, want)
	}
}

func TestDefaultHostsPath_WindowsEmptySystemRoot(t *testing.T) {
	got := hostsPathFor(goosWindows, "")
	want := `C:\Windows\System32\drivers\etc\hosts`
	if got != want {
		t.Errorf("hostsPathFor(windows, \"\") = %q, want %q", got, want)
	}
}

func TestDefaultHostsPath_Unix(t *testing.T) {
	for _, goos := range []string{goosDarwin, goosLinux, "freebsd", "openbsd"} {
		got := hostsPathFor(goos, `C:\Windows`) // SystemRoot must be ignored off Windows
		if got != "/etc/hosts" {
			t.Errorf("hostsPathFor(%q, ...) = %q, want /etc/hosts", goos, got)
		}
	}
}

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
