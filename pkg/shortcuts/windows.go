// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package shortcuts

import (
	"context"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

// ParseWindowsHotkeyList parses dump lines of the form:
//
//	Name=Copy;Binding=Ctrl+C;Enabled=true
//	Name=Paste;Binding=Ctrl+V;Enabled=true
//
// ID is always 0 on Windows (no Apple symbolic IDs). Field order is free.
// Malformed lines are skipped. Results are sorted by Name. Pure function for
// offline unit tests.
func ParseWindowsHotkeyList(output string) []Shortcut {
	var out []Shortcut
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		sc, ok := parseWindowsHotkeyLine(line)
		if !ok {
			continue
		}
		out = append(out, sc)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// parseWindowsHotkeyLine parses one Name=...;Binding=...;Enabled=... record.
func parseWindowsHotkeyLine(line string) (Shortcut, bool) {
	var sc Shortcut
	var haveName bool
	// Default Enabled=true when Binding is present and Enabled is omitted,
	// matching the dump format's usual "listed means active" convention.
	enabledSet := false
	enabled := true

	for _, part := range strings.Split(line, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, val, ok := strings.Cut(part, "=")
		if !ok {
			return Shortcut{}, false
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		switch key {
		case "Name":
			if val == "" {
				return Shortcut{}, false
			}
			sc.Name = val
			haveName = true
		case "Binding":
			sc.Binding = val
		case "Enabled":
			b, err := strconv.ParseBool(val)
			if err != nil {
				// Also accept 1/0 which PowerShell dumps sometimes emit.
				switch val {
				case "1":
					b = true
				case "0":
					b = false
				default:
					return Shortcut{}, false
				}
			}
			enabled = b
			enabledSet = true
		}
	}
	if !haveName {
		return Shortcut{}, false
	}
	if !enabledSet {
		// Unbound entries without an explicit Enabled flag are disabled.
		enabled = sc.Binding != ""
	}
	if sc.Binding == "" {
		enabled = false
	}
	sc.Enabled = enabled
	sc.ID = 0
	return sc, true
}

// listWindows probes for enumerable hotkeys. Windows has no public API for
// system-wide shortcuts comparable to macOS symbolichotkeys or GNOME
// gsettings. We surface a few Accessibility toggle states when present;
// empty result means supported with none discovered (not ErrUnsupported).
// Offline dumps parse via ParseWindowsHotkeyList.
func listWindows(ctx context.Context) ([]Shortcut, error) {
	ps := `$ErrorActionPreference='SilentlyContinue';` +
		`$sk = Get-ItemProperty 'HKCU:\Control Panel\Accessibility\StickyKeys' -ErrorAction SilentlyContinue;` +
		`$fk = Get-ItemProperty 'HKCU:\Control Panel\Accessibility\Keyboard Preference' -ErrorAction SilentlyContinue;` +
		`if ($null -ne $sk) { $en = if (($sk.Flags -band 1) -ne 0) { 'true' } else { 'false' }; Write-Output ("Name=StickyKeys;Binding=Shift x5;Enabled=$en") };` +
		`if ($null -ne $fk) { Write-Output 'Name=KeyboardPreference;Binding=;Enabled=true' }`
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", ps).Output()
	if err != nil {
		// Windows path is supported; hotkey discovery is best-effort.
		return []Shortcut{}, nil //nolint:nilerr // empty list is intentional when PowerShell fails
	}
	return ParseWindowsHotkeyList(string(out)), nil
}
