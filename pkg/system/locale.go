// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ErrLocaleUnsupported is returned on platforms without a locale backend.
var ErrLocaleUnsupported = errors.New("system: locale query unsupported on this platform")

// ErrLocaleSetUnsupported is returned when setting locale is not supported.
var ErrLocaleSetUnsupported = errors.New("system: locale set unsupported on this platform")

// GetLocale returns the user's locale identifier. On macOS it reads
// `defaults read .GlobalPreferences AppleLocale` (e.g. "ko_KR").
// On Linux it prefers $LANG, then falls back to `localectl status`.
// On Windows it runs `(Get-Culture).Name` via PowerShell (e.g. "ko-KR").
// Canceling ctx aborts the underlying platform query.
func GetLocale(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case goosDarwin:
		return getLocaleMacOS(ctx)
	case goosLinux:
		return getLocaleLinux(ctx)
	case goosWindows:
		return getLocaleWindows(ctx)
	default:
		return "", ErrLocaleUnsupported
	}
}

// SetLocale writes the user locale identifier.
// macOS: `defaults write -g AppleLocale -string <locale>`.
// Linux: `localectl set-locale LANG=<locale>` (may require elevated privileges).
// Windows and other platforms return ErrLocaleSetUnsupported.
// Empty locale returns an error. Canceling ctx aborts the command.
func SetLocale(ctx context.Context, locale string) error {
	locale = strings.TrimSpace(locale)
	if locale == "" {
		return errors.New("system: empty locale")
	}
	switch runtime.GOOS {
	case goosDarwin:
		return setLocaleMacOS(ctx, locale)
	case goosLinux:
		return setLocaleLinux(ctx, locale)
	case goosWindows:
		return ErrLocaleSetUnsupported
	default:
		return ErrLocaleSetUnsupported
	}
}

func setLocaleMacOS(ctx context.Context, locale string) error {
	// Match GetLocale which reads .GlobalPreferences / NSGlobalDomain AppleLocale.
	// #nosec G204 -- locale is validated non-empty; argv is fixed command + flag form
	cmd := exec.CommandContext(ctx, "defaults", "write", "-g", "AppleLocale", "-string", locale)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("system: set locale: %w: %s", err, msg)
		}
		return fmt.Errorf("system: set locale: %w", err)
	}
	return nil
}

func setLocaleLinux(ctx context.Context, locale string) error {
	// localectl expects LANG=... form; accept bare "ko_KR.UTF-8" too.
	value := locale
	if !strings.Contains(value, "=") {
		value = "LANG=" + locale
	}
	// #nosec G204 -- fixed binary; value is a single argv element
	cmd := exec.CommandContext(ctx, "localectl", "set-locale", value)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg != "" {
			return fmt.Errorf("system: set locale: %w: %s", err, msg)
		}
		return fmt.Errorf("system: set locale: %w", err)
	}
	return nil
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
