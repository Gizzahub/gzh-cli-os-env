// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package display reports connected displays. Phase 3 covers macOS via
// system_profiler SPDisplaysDataType; other platforms return ErrUnsupported.
package display

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// ErrUnsupported is returned when the platform has no display backend.
var ErrUnsupported = errors.New("display: query unsupported on this platform")

// Info describes a single connected display.
type Info struct {
	Name       string
	Resolution string // e.g. "2560 x 1440"
	Main       bool
}

var (
	reDisplayName = regexp.MustCompile(`(?m)^\s{4}([^\n:]+):\s*$`)
	reResolution  = regexp.MustCompile(`(?m)^\s+Resolution:\s*(.+)$`)
	reMain        = regexp.MustCompile(`(?m)^\s+Main Display:\s*Yes`)
)

// ParseSystemProfilerDisplays parses `system_profiler SPDisplaysDataType` text.
// Pure function for offline unit tests.
func ParseSystemProfilerDisplays(output string) []Info {
	var displays []Info
	// Split by top-level GPU/display sections roughly by lines starting with 4 spaces name:
	// Walk line by line.
	lines := strings.Split(output, "\n")
	var cur *Info
	flush := func() {
		if cur != nil && (cur.Name != "" || cur.Resolution != "") {
			displays = append(displays, *cur)
		}
		cur = nil
	}
	for _, line := range lines {
		if m := reDisplayName.FindStringSubmatch(line); m != nil {
			name := strings.TrimSpace(m[1])
			// skip known non-display keys under GPU
			switch name {
			case "Chipset Model", "Bus", "VRAM (Total)", "Vendor", "Device ID",
				"Revision ID", "Metal Support", "Displays", "Total Number of Cores",
				"Type", "Display Type", "Framebuffer Depth", "Mirror", "Online",
				"Automatically Adjust Brightness", "Connection Type", "Television":
				continue
			}
			// Heuristic: display names often appear under Displays: section with Resolution nearby.
			// Keep candidates that later get a Resolution.
			flush()
			cur = &Info{Name: name}
			continue
		}
		if cur == nil {
			continue
		}
		if m := reResolution.FindStringSubmatch(line); m != nil {
			cur.Resolution = strings.TrimSpace(m[1])
			continue
		}
		if reMain.MatchString(line) {
			cur.Main = true
		}
	}
	flush()
	// Drop entries without resolution (likely not displays)
	out := make([]Info, 0, len(displays))
	for _, d := range displays {
		if d.Resolution != "" {
			out = append(out, d)
		}
	}
	return out
}

// List returns connected displays for the current platform.
func List() ([]Info, error) {
	switch runtime.GOOS {
	case "darwin":
		return listMacOS()
	default:
		return nil, ErrUnsupported
	}
}

func listMacOS() ([]Info, error) {
	out, err := exec.Command("system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return nil, err
	}
	return ParseSystemProfilerDisplays(string(out)), nil
}
