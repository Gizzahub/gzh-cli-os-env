// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package system reports and edits OS-level settings: hosts file, locale,
// and timezone. Hosts is read-only on all platforms (list via GetHosts +
// DefaultHostsPath). Locale/timezone writes are supported on macOS and Linux;
// Windows set paths return unsupported errors.
package system

import (
	"os"
	"runtime"
	"strings"
)

// runtime.GOOS string constants (goconst across Get/Set dispatch).
const (
	goosDarwin  = "darwin"
	goosLinux   = "linux"
	goosWindows = "windows"
)

// defaultWindowsSystemRoot is used when SystemRoot is unset (unusual but possible).
const defaultWindowsSystemRoot = `C:\Windows`

// HostsEntry is one parsed line of a hosts file: an IP and its hostnames.
type HostsEntry struct {
	IP    string
	Names []string
}

// DefaultHostsPath returns the platform hosts file path for the current OS.
// On Windows it is %SystemRoot%\System32\drivers\etc\hosts (SystemRoot defaults
// to C:\Windows when unset). On Unix-like systems it is /etc/hosts.
// Listing is supported on all platforms that can read that path; this package
// does not provide write (add/remove) APIs — hosts editing often needs elevated
// privileges and is out of scope.
func DefaultHostsPath() string {
	return hostsPathFor(runtime.GOOS, os.Getenv("SystemRoot"))
}

// hostsPathFor is the pure path selector used by DefaultHostsPath.
// goos is a runtime.GOOS value; systemRoot is Windows SystemRoot (may be empty).
func hostsPathFor(goos, systemRoot string) string {
	if goos == goosWindows {
		root := systemRoot
		if root == "" {
			root = defaultWindowsSystemRoot
		}
		// Fixed Windows path segment; do not use filepath.Join so tests on
		// non-Windows keep backslash separators.
		return root + `\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

// ParseHosts parses the contents of a hosts file into entries.
// Blank lines and lines whose first non-space character is '#' are
// skipped. Lines without at least an IP and one hostname are skipped.
// It is a pure function so the parsing logic can be tested anywhere.
func ParseHosts(content string) ([]HostsEntry, error) {
	var entries []HostsEntry
	for _, raw := range strings.Split(content, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Strip an inline comment.
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		entries = append(entries, HostsEntry{IP: fields[0], Names: fields[1:]})
	}
	return entries, nil
}

// GetHosts reads and parses the system hosts file at path.
// Prefer DefaultHostsPath() for the current platform's default location.
func GetHosts(path string) ([]HostsEntry, error) {
	// #nosec G304 -- reading a caller-chosen path is this function's contract;
	// the caller, not this package, decides which hosts file to inspect.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHosts(string(data))
}
