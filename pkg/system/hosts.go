// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package system reports and edits OS-level settings: hosts file, locale,
// and timezone. Locale/timezone writes are supported on macOS and Linux;
// Windows set paths return unsupported errors.
package system

import (
	"os"
	"strings"
)

// runtime.GOOS string constants (goconst across Get/Set dispatch).
const (
	goosDarwin  = "darwin"
	goosLinux   = "linux"
	goosWindows = "windows"
)

// HostsEntry is one parsed line of /etc/hosts: an IP and its hostnames.
type HostsEntry struct {
	IP    string
	Names []string
}

// ParseHosts parses the contents of an /etc/hosts file into entries.
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

// GetHosts reads and parses the system hosts file (path).
func GetHosts(path string) ([]HostsEntry, error) {
	// #nosec G304 -- reading a caller-chosen path is this function's contract;
	// the caller, not this package, decides which hosts file to inspect.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseHosts(string(data))
}
