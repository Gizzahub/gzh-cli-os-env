// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package input

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// ParsePowerShellLanguageList extracts language tags from PowerShell output.
// Pure function for offline unit tests.
//
// Accepts either Format-List style:
//
//	LanguageTag : en-US
//	LanguageTag : ko-KR
//
// or one bare tag per line:
//
//	en-US
//	ko-KR
//
// Each tag becomes Source{Name: tag, Kind: "LanguageTag"}.
func ParsePowerShellLanguageList(output string) []Source {
	var out []Source
	seen := make(map[string]struct{})
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		tag := extractLanguageTag(line)
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, Source{Name: tag, Kind: "LanguageTag"})
	}
	return out
}

// extractLanguageTag pulls a BCP-47-like tag from a Format-List line or a bare
// tag line. Non-LanguageTag property lines are rejected.
func extractLanguageTag(line string) string {
	if key, val, ok := strings.Cut(line, ":"); ok {
		if !strings.EqualFold(strings.TrimSpace(key), "LanguageTag") {
			return ""
		}
		return normalizeLanguageTag(strings.TrimSpace(val))
	}
	return normalizeLanguageTag(line)
}

// normalizeLanguageTag accepts simple BCP-47 tags (en, en-US, zh-Hans-CN).
func normalizeLanguageTag(s string) string {
	s = strings.TrimSpace(s)
	if s == "" || strings.ContainsAny(s, " \t") {
		return ""
	}
	if len(s) < 2 || len(s) > 16 {
		return ""
	}
	for i, r := range s {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z':
			continue
		case r >= '0' && r <= '9' && i > 0:
			continue
		case r == '-' && i > 0:
			continue
		default:
			return ""
		}
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") || strings.Contains(s, "--") {
		return ""
	}
	return s
}

// ParseKeyboardDelay parses a KeyboardDelay value from a registry / PowerShell
// dump. Pure function for offline unit tests.
//
// Accepts:
//
//	KeyboardDelay=1
//	KeyboardDelay : 1
//	1
//
// Windows stores KeyboardDelay as 0–3 (shortest to longest). On success Set is
// true; parse failure returns an unset Setting and an error.
func ParseKeyboardDelay(output string) (Setting, error) {
	s := strings.TrimSpace(output)
	if s == "" {
		return Setting{}, fmt.Errorf("input: empty keyboard delay")
	}
	for _, raw := range strings.Split(s, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		val := line
		if key, v, ok := strings.Cut(line, "="); ok {
			if !strings.EqualFold(strings.TrimSpace(key), "KeyboardDelay") {
				continue
			}
			val = strings.TrimSpace(v)
		} else if key, v, ok := strings.Cut(line, ":"); ok {
			if !strings.EqualFold(strings.TrimSpace(key), "KeyboardDelay") {
				continue
			}
			val = strings.TrimSpace(v)
		}
		n, err := strconv.Atoi(val)
		if err != nil {
			return Setting{}, fmt.Errorf("input: parse keyboard delay %q: %w", val, err)
		}
		return Setting{Value: n, Set: true}, nil
	}
	return Setting{}, fmt.Errorf("input: no keyboard delay in dump")
}

// keyboardWindows reads input languages via Get-WinUserLanguageList and
// optionally KeyboardDelay from HKCU:\Control Panel\Keyboard. RepeatRate is
// left unset: Windows KeyboardSpeed is not a direct analog of macOS KeyRepeat.
func keyboardWindows(ctx context.Context) (*Keyboard, error) {
	kb := &Keyboard{
		RepeatDelay: readWindowsKeyboardDelay(ctx),
	}
	sources, err := inputSourcesWindows(ctx)
	if err != nil {
		// Language list is best-effort when PowerShell is unavailable mid-probe;
		// still return delay if we got it. A hard failure with no data is rare.
		sources = nil
	}
	kb.Sources = sources
	return kb, nil
}

func inputSourcesWindows(ctx context.Context) ([]Source, error) {
	// Emit one LanguageTag per line for a stable pure-parser contract.
	const ps = `$ErrorActionPreference='Stop'; ` +
		`(Get-WinUserLanguageList).LanguageTag`
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", ps).Output()
	if err != nil {
		return nil, fmt.Errorf("input: Get-WinUserLanguageList: %w", err)
	}
	return ParsePowerShellLanguageList(string(out)), nil
}

func readWindowsKeyboardDelay(ctx context.Context) Setting {
	const ps = `$ErrorActionPreference='SilentlyContinue'; ` +
		`(Get-ItemProperty 'HKCU:\Control Panel\Keyboard').KeyboardDelay`
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", ps).Output()
	if err != nil {
		return Setting{}
	}
	s, err := ParseKeyboardDelay(string(out))
	if err != nil {
		return Setting{}
	}
	return s
}
