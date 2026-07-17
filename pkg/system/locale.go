// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ErrLocaleUnsupported is returned on platforms without a locale backend.
var ErrLocaleUnsupported = errors.New("system: locale query unsupported on this platform")

// GetLocale returns the user's locale identifier. On macOS it reads
// `defaults read .GlobalPreferences AppleLocale` (e.g. "ko_KR").
// On Linux it prefers $LANG, then falls back to `localectl status`.
// On Windows it runs `(Get-Culture).Name` via PowerShell (e.g. "ko-KR").
// Canceling ctx aborts the underlying platform query.
func GetLocale(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getLocaleMacOS(ctx)
	case "linux":
		return getLocaleLinux(ctx)
	case "windows":
		return getLocaleWindows(ctx)
	default:
		return "", ErrLocaleUnsupported
	}
}

func getLocaleMacOS(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "defaults", "read", ".GlobalPreferences", "AppleLocale").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func getLocaleLinux(ctx context.Context) (string, error) {
	if lang := strings.TrimSpace(os.Getenv("LANG")); lang != "" {
		return lang, nil
	}
	out, err := exec.CommandContext(ctx, "localectl", "status").Output()
	if err != nil {
		return "", err
	}
	loc := ParseLocalectlStatus(string(out))
	if loc == "" {
		return "", errors.New("system: could not parse localectl status")
	}
	return loc, nil
}

// ParseLocalectlStatus extracts LANG from `localectl status` output.
// Looks for a line like:
//
//	System Locale: LANG=ko_KR.UTF-8
//
// or multi-value forms such as:
//
//	System Locale: LANG=en_US.UTF-8
//	               LC_TIME=C.UTF-8
//
// Pure function for offline unit tests.
func ParseLocalectlStatus(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		// Prefer explicit LANG= token anywhere on the line.
		if idx := strings.Index(line, "LANG="); idx >= 0 {
			rest := line[idx+len("LANG="):]
			// Stop at whitespace or semicolon if present.
			if end := strings.IndexAny(rest, " \t;"); end >= 0 {
				rest = rest[:end]
			}
			return strings.TrimSpace(rest)
		}
	}
	return ""
}

// ParsePowerShellCulture returns the first non-empty trimmed line from
// PowerShell culture output (e.g. "ko-KR" from `(Get-Culture).Name`).
// Pure function for offline unit tests.
func ParsePowerShellCulture(output string) string {
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func getLocaleWindows(ctx context.Context) (string, error) {
	// #nosec G204 -- fixed PowerShell command, no user input
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command",
		"(Get-Culture).Name").Output()
	if err != nil {
		return "", err
	}
	loc := ParsePowerShellCulture(string(out))
	if loc == "" {
		return "", errors.New("system: could not parse PowerShell culture")
	}
	return loc, nil
}
