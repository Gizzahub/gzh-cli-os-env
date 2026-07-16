// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package shortcuts reports keyboard shortcuts. Phase 3 covers macOS via
// com.apple.symbolichotkeys; other platforms return ErrUnsupported.
package shortcuts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

// ErrUnsupported is returned when the platform has no shortcuts backend.
var ErrUnsupported = errors.New("shortcuts: query unsupported on this platform")

// Shortcut describes one system keyboard shortcut.
type Shortcut struct {
	ID      int    // Apple symbolic hotkey ID
	Name    string // human-readable name; empty when the ID is not in knownHotkeys
	Enabled bool
	Binding string // e.g. "Cmd+Space"; empty when the shortcut uses its built-in default
}

// Named reports whether this shortcut's ID maps to a known name. Callers must
// not present an unnamed shortcut as anything but its raw ID.
func (s Shortcut) Named() bool { return s.Name != "" }

// knownHotkeys maps Apple symbolic hotkey IDs to names.
//
// Apple does not document these IDs. This table is deliberately limited to
// entries confirmed against a live system: an ID absent here is reported as
// unknown rather than guessed, because a wrong name is worse than no name.
// Add entries only after verifying against System Settings > Keyboard.
var knownHotkeys = map[int]string{
	60: "Select the previous input source",
	61: "Select the next input source",
	64: "Show Spotlight search",
	65: "Show Spotlight file search window",
	79: "Move left a space",
	80: "Move left a space (alternate)",
	81: "Move right a space",
	82: "Move right a space (alternate)",
}

// Modifier bits used in a symbolic hotkey's third parameter.
const (
	maskShift   = 0x020000
	maskControl = 0x040000
	maskOption  = 0x080000
	maskCommand = 0x100000
)

// noParameter marks an unset ASCII code or key code (Apple uses 65535).
const noParameter = 65535

// keyCodes maps virtual key codes to display names. Only keys that carry no
// printable ASCII in parameter 0 need an entry.
var keyCodes = map[int]string{
	36:  "Return",
	48:  "Tab",
	49:  "Space",
	51:  "Delete",
	53:  "Escape",
	114: "Help",
	115: "Home",
	116: "PageUp",
	117: "ForwardDelete",
	119: "End",
	121: "PageDown",
	123: "Left",
	124: "Right",
	125: "Down",
	126: "Up",
}

// FormatBinding renders a symbolic hotkey's parameters as "Cmd+Shift+4".
// parameters is [asciiCode, keyCode, modifierMask]; it returns "" when the
// parameters describe no usable binding, which callers must render as the
// system default rather than as "no shortcut".
func FormatBinding(parameters []int) string {
	if len(parameters) < 3 {
		return ""
	}
	ascii, keyCode, mask := parameters[0], parameters[1], parameters[2]

	var parts []string
	for _, m := range []struct {
		bit  int
		name string
	}{
		{maskControl, "Ctrl"},
		{maskOption, "Option"},
		{maskShift, "Shift"},
		{maskCommand, "Cmd"},
	} {
		if mask&m.bit != 0 {
			parts = append(parts, m.name)
		}
	}

	key := formatKey(ascii, keyCode)
	if key == "" {
		return ""
	}
	return strings.Join(append(parts, key), "+")
}

// formatKey prefers the key code so that named keys (Space, Tab) win over their
// ASCII spelling, and falls back to the printable ASCII character.
func formatKey(ascii, keyCode int) string {
	if name, ok := keyCodes[keyCode]; ok {
		return name
	}
	if ascii != noParameter && ascii > 32 && ascii < 127 {
		return strings.ToUpper(string(rune(ascii)))
	}
	if keyCode != noParameter {
		return "Key" + strconv.Itoa(keyCode)
	}
	return ""
}

// symbolicHotkeys mirrors the JSON shape of com.apple.symbolichotkeys.
// The top-level tag carries Apple's key name verbatim, so the project's
// snake_case tag convention cannot apply here.
//
//nolint:tagliatelle // external schema: keys are Apple's, not ours
type symbolicHotkeys struct {
	AppleSymbolicHotKeys map[string]struct {
		Enabled bool `json:"enabled"`
		Value   struct {
			Parameters []int `json:"parameters"`
		} `json:"value"`
	} `json:"AppleSymbolicHotKeys"`
}

// ParseSymbolicHotkeysJSON parses com.apple.symbolichotkeys converted to JSON.
// Pure function for offline unit tests. Results are sorted by ID.
func ParseSymbolicHotkeysJSON(data []byte) ([]Shortcut, error) {
	var raw symbolicHotkeys
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("shortcuts: parse symbolichotkeys: %w", err)
	}

	out := make([]Shortcut, 0, len(raw.AppleSymbolicHotKeys))
	for key, entry := range raw.AppleSymbolicHotKeys {
		id, err := strconv.Atoi(key)
		if err != nil {
			// A non-numeric key is not a hotkey; skip rather than fail the dump.
			continue
		}
		out = append(out, Shortcut{
			ID:      id,
			Name:    knownHotkeys[id],
			Enabled: entry.Enabled,
			Binding: FormatBinding(entry.Value.Parameters),
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

// List returns system keyboard shortcuts for the current platform.
// Canceling ctx aborts the underlying platform query.
func List(ctx context.Context) ([]Shortcut, error) {
	switch runtime.GOOS {
	case "darwin":
		return listMacOS(ctx)
	default:
		return nil, ErrUnsupported
	}
}

// listMacOS exports the hotkey domain and converts it to JSON with plutil.
// `defaults read` emits legacy plist, so the export/convert pair is what makes
// the payload parseable with the standard library alone.
func listMacOS(ctx context.Context) ([]Shortcut, error) {
	export := exec.CommandContext(ctx, "defaults", "export", "com.apple.symbolichotkeys", "-")
	plist, err := export.Output()
	if err != nil {
		return nil, fmt.Errorf("shortcuts: export symbolichotkeys: %w", err)
	}

	convert := exec.CommandContext(ctx, "plutil", "-convert", "json", "-o", "-", "-")
	convert.Stdin = strings.NewReader(string(plist))
	data, err := convert.Output()
	if err != nil {
		return nil, fmt.Errorf("shortcuts: convert symbolichotkeys to json: %w", err)
	}
	return ParseSymbolicHotkeysJSON(data)
}
