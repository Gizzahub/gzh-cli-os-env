// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package input

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ParseGsettingsInt parses a scalar `gsettings get` integer printout.
// Accepts bare numbers and GVariant-typed forms such as "uint32 30".
func ParseGsettingsInt(output string) (int, error) {
	s := strings.TrimSpace(output)
	if s == "" {
		return 0, fmt.Errorf("input: empty gsettings int")
	}
	fields := strings.Fields(s)
	// Last token is the numeric value after an optional type tag.
	n, err := strconv.Atoi(fields[len(fields)-1])
	if err != nil {
		return 0, fmt.Errorf("input: parse gsettings int %q: %w", s, err)
	}
	return n, nil
}

// ParseSetxkbmapQuery extracts keyboard layouts from `setxkbmap -query`.
// Pure function for offline unit tests.
//
// Example input:
//
//	rules:      evdev
//	model:      pc105
//	layout:     us,kr
//	variant:    ,
//	options:    grp:alt_shift_toggle
func ParseSetxkbmapQuery(output string) []Source {
	for _, raw := range strings.Split(output, "\n") {
		line := strings.TrimSpace(raw)
		if !strings.HasPrefix(line, "layout:") {
			continue
		}
		layouts := strings.TrimSpace(strings.TrimPrefix(line, "layout:"))
		if layouts == "" {
			return nil
		}
		parts := strings.Split(layouts, ",")
		out := make([]Source, 0, len(parts))
		for _, p := range parts {
			name := strings.TrimSpace(p)
			if name == "" {
				continue
			}
			out = append(out, Source{Name: name, Kind: "Keyboard Layout"})
		}
		return out
	}
	return nil
}

// keyboardLinux dispatches by XDG_CURRENT_DESKTOP.
// GNOME: gsettings repeat-interval/delay + setxkbmap layouts.
// KDE: setxkbmap layouts; gsettings tried opportunistically (often absent).
func keyboardLinux(ctx context.Context) (*Keyboard, error) {
	de := strings.ToLower(os.Getenv("XDG_CURRENT_DESKTOP"))
	kb := &Keyboard{}

	switch {
	case strings.Contains(de, "gnome"):
		kb.RepeatRate = readGsettingsInt(ctx, "repeat-interval")
		kb.RepeatDelay = readGsettingsInt(ctx, "delay")
	case strings.Contains(de, "kde"), strings.Contains(de, "plasma"):
		// KDE stores repeat settings outside gsettings; try anyway, keep unset on miss.
		kb.RepeatRate = readGsettingsInt(ctx, "repeat-interval")
		kb.RepeatDelay = readGsettingsInt(ctx, "delay")
	default:
		return nil, ErrUnsupported
	}

	sources, err := inputSourcesLinux(ctx)
	if err != nil {
		// setxkbmap is optional; a missing binary still yields repeat settings.
		sources = nil
	}
	kb.Sources = sources
	return kb, nil
}

func readGsettingsInt(ctx context.Context, key string) Setting {
	const schema = "org.gnome.desktop.peripherals.keyboard"
	// #nosec G204 -- schema/key are package-controlled constants
	out, err := exec.CommandContext(ctx, "gsettings", "get", schema, key).Output()
	if err != nil {
		return Setting{}
	}
	v, err := ParseGsettingsInt(string(out))
	if err != nil {
		return Setting{}
	}
	return Setting{Value: v, Set: true}
}

func inputSourcesLinux(ctx context.Context) ([]Source, error) {
	out, err := exec.CommandContext(ctx, "setxkbmap", "-query").Output()
	if err != nil {
		return nil, fmt.Errorf("input: setxkbmap -query: %w", err)
	}
	return ParseSetxkbmapQuery(string(out)), nil
}
