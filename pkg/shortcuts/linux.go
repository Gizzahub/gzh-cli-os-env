// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package shortcuts

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// kglobalAccelFile overrides the default ~/.config/kglobalshortcutsrc path.
// Empty means resolve from the user home directory at call time.
var kglobalAccelFile string

// ParseGsettingsKeybindings parses `gsettings list-recursively` lines for a
// keybinding schema. Pure function for offline unit tests.
//
// Example lines:
//
//	org.gnome.desktop.wm.keybindings switch-windows ['<Alt>Tab']
//	org.gnome.desktop.wm.keybindings close ['<Alt>F4']
//	org.gnome.desktop.wm.keybindings maximize @as []
func ParseGsettingsKeybindings(output string) []Shortcut {
	var out []Shortcut
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		// schema key value… — schema and key are single tokens; value is the rest.
		schemaEnd := strings.IndexByte(line, ' ')
		if schemaEnd < 0 {
			continue
		}
		rest := strings.TrimSpace(line[schemaEnd+1:])
		keyEnd := strings.IndexByte(rest, ' ')
		if keyEnd < 0 {
			continue
		}
		name := rest[:keyEnd]
		value := strings.TrimSpace(rest[keyEnd+1:])
		bindings := parseGsettingsStringList(value)
		binding := strings.Join(bindings, ", ")
		out = append(out, Shortcut{
			ID:      0,
			Name:    name,
			Enabled: len(bindings) > 0,
			Binding: binding,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// parseGsettingsStringList extracts non-empty strings from a gsettings list
// value such as "['<Alt>Tab']", "['a', 'b']", or "@as []".
func parseGsettingsStringList(value string) []string {
	value = strings.TrimSpace(value)
	// Drop optional GVariant type tag: "@as []"
	if strings.HasPrefix(value, "@") {
		if i := strings.IndexByte(value, ' '); i >= 0 {
			value = strings.TrimSpace(value[i+1:])
		}
	}
	if !strings.HasPrefix(value, "[") || !strings.HasSuffix(value, "]") {
		return nil
	}
	inner := strings.TrimSpace(value[1 : len(value)-1])
	if inner == "" {
		return nil
	}

	var out []string
	for _, part := range splitGsettingsListItems(inner) {
		s := strings.TrimSpace(part)
		s = strings.Trim(s, `"'`)
		if s == "" {
			continue
		}
		out = append(out, s)
	}
	return out
}

// splitGsettingsListItems splits a comma-separated GVariant string list body
// without treating commas inside quotes as separators.
func splitGsettingsListItems(inner string) []string {
	var items []string
	var b strings.Builder
	inQuote := false
	quote := rune(0)
	for _, r := range inner {
		switch {
		case (r == '\'' || r == '"') && !inQuote:
			inQuote = true
			quote = r
			b.WriteRune(r)
		case inQuote && r == quote:
			inQuote = false
			b.WriteRune(r)
		case r == ',' && !inQuote:
			items = append(items, b.String())
			b.Reset()
		default:
			b.WriteRune(r)
		}
	}
	if b.Len() > 0 {
		items = append(items, b.String())
	}
	return items
}

// ParseKGlobalAccel parses a kglobalshortcutsrc file (INI-like). Pure function
// for offline unit tests.
//
// Lines look like:
//
//	Walk Through Windows=Alt+Tab,Alt+Tab,Walk Through Windows
//	Window Close=none,Alt+F4,Close Window
//
// Fields are current,default,description. current may use ';' for alternatives
// and "none" (or empty) when unbound.
func ParseKGlobalAccel(rcText string) []Shortcut {
	var out []Shortcut
	for _, raw := range strings.Split(rcText, "\n") {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "[") {
			continue
		}
		eq := strings.IndexByte(line, '=')
		if eq <= 0 {
			continue
		}
		name := strings.TrimSpace(line[:eq])
		value := line[eq+1:]
		// current,default,description — description may contain commas.
		parts := strings.SplitN(value, ",", 3)
		current := ""
		if len(parts) > 0 {
			current = strings.TrimSpace(parts[0])
		}
		binding, enabled := normalizeKDEBinding(current)
		out = append(out, Shortcut{
			ID:      0,
			Name:    name,
			Enabled: enabled,
			Binding: binding,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// normalizeKDEBinding turns a kglobalaccel "current" field into a display
// binding. "none"/empty means unbound; ';' separates alternate chords.
func normalizeKDEBinding(current string) (binding string, enabled bool) {
	if current == "" || strings.EqualFold(current, "none") {
		return "", false
	}
	// Alternatives are ';'-separated; present them comma-separated for display.
	alts := strings.Split(current, ";")
	var kept []string
	for _, a := range alts {
		a = strings.TrimSpace(a)
		if a == "" || strings.EqualFold(a, "none") {
			continue
		}
		kept = append(kept, a)
	}
	if len(kept) == 0 {
		return "", false
	}
	return strings.Join(kept, ", "), true
}

// listLinux dispatches by XDG_CURRENT_DESKTOP (GNOME primary, KDE secondary).
func listLinux(ctx context.Context) ([]Shortcut, error) {
	de := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	switch {
	case strings.Contains(de, "gnome"):
		return listGNOME(ctx)
	case strings.Contains(de, "kde"), strings.Contains(de, "plasma"):
		return listKDE()
	default:
		return nil, ErrUnsupported
	}
}

func listGNOME(ctx context.Context) ([]Shortcut, error) {
	// #nosec G204 -- fixed schema; no user input in the argument list.
	out, err := exec.CommandContext(ctx, "gsettings", "list-recursively", "org.gnome.desktop.wm.keybindings").Output()
	if err != nil {
		return nil, fmt.Errorf("shortcuts: gsettings list-recursively: %w", err)
	}
	return ParseGsettingsKeybindings(string(out)), nil
}

func listKDE() ([]Shortcut, error) {
	path := resolveKGlobalAccelFile()
	if path == "" {
		return nil, fmt.Errorf("shortcuts: cannot resolve kglobalshortcutsrc path")
	}
	// #nosec G304 -- path is the standard user config file or a test override.
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("shortcuts: read kglobalshortcutsrc: %w", err)
	}
	return ParseKGlobalAccel(string(data)), nil
}

func resolveKGlobalAccelFile() string {
	if kglobalAccelFile != "" {
		return kglobalAccelFile
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "kglobalshortcutsrc")
}
