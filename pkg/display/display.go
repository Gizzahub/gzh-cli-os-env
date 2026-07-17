// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package display reports connected displays. Phase 3 covers macOS via
// system_profiler SPDisplaysDataType; Phase 4 adds Linux via xrandr
// (with optional wlr-randr fallback); Phase 5 adds Windows via WMIC
// Win32_VideoController. Other platforms return ErrUnsupported.
package display

import (
	"context"
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
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

	// xrandr connected line, e.g.:
	//   eDP-1 connected primary 1920x1080+0+0 (normal left inverted right x axis y axis) 340mm x 190mm
	//   HDMI-1 connected 2560x1440+1920+0 (normal left inverted right x axis y axis) 600mm x 340mm
	//   DP-1 connected primary (normal left inverted right x axis y axis)
	reXrandrConnected = regexp.MustCompile(
		`^(\S+)\s+connected(?:\s+primary)?(?:\s+(\d+)x(\d+)\+\d+\+\d+)?`,
	)

	// wlr-randr output block header + current mode, e.g.:
	//   eDP-1 "Sharp ..."
	//     ...
	//     current 1920x1080@60.000000Hz
	reWlrName    = regexp.MustCompile(`^(\S+)\s+"`)
	reWlrCurrent = regexp.MustCompile(`^\s+current\s+(\d+)x(\d+)`)
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

// ParseXrandr parses `xrandr --query` output for connected displays.
// Pure function for offline unit tests.
//
// Example lines:
//
//	eDP-1 connected primary 1920x1080+0+0 (normal left inverted right x axis y axis) 340mm x 190mm
//	HDMI-1 connected 2560x1440+1920+0 (normal left inverted right x axis y axis) 600mm x 340mm
//	DP-2 disconnected (normal left inverted right x axis y axis)
func ParseXrandr(output string) []Info {
	var displays []Info
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(line, " connected") {
			continue
		}
		// Skip "disconnected" which also contains the substring carefully:
		// " connected" with leading space avoids matching "disconnected".
		if strings.Contains(line, " disconnected") {
			continue
		}
		m := reXrandrConnected.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		info := Info{Name: m[1]}
		if strings.Contains(line, " primary ") || strings.HasSuffix(line, " primary") ||
			strings.Contains(line, " connected primary") {
			info.Main = true
		}
		if m[2] != "" && m[3] != "" {
			info.Resolution = m[2] + " x " + m[3]
		}
		displays = append(displays, info)
	}
	return displays
}

// ParseWlrRandr parses `wlr-randr` text output for connected displays.
// Pure function for offline unit tests. Optional Wayland fallback.
func ParseWlrRandr(output string) []Info {
	var displays []Info
	var cur *Info
	flush := func() {
		if cur != nil && cur.Name != "" {
			displays = append(displays, *cur)
		}
		cur = nil
	}
	for _, line := range strings.Split(output, "\n") {
		if m := reWlrName.FindStringSubmatch(line); m != nil {
			flush()
			cur = &Info{Name: m[1]}
			continue
		}
		if cur == nil {
			continue
		}
		if m := reWlrCurrent.FindStringSubmatch(line); m != nil {
			cur.Resolution = m[1] + " x " + m[2]
			continue
		}
		if strings.Contains(line, "Enabled: yes") || strings.Contains(line, "Enabled: true") {
			// keep; no-op marker that output is active
			continue
		}
		// First output is treated as main if none marked; wlr-randr has no
		// universal primary flag across compositors.
	}
	flush()
	if len(displays) > 0 {
		displays[0].Main = true
	}
	return displays
}

// ParseWmicDesktopMonitor parses WMIC /format:list output for display
// adapters or monitors. Accepts either DesktopMonitor fields:
//
//	Name=Generic PnP Monitor
//	ScreenHeight=1080
//	ScreenWidth=1920
//
// or VideoController fields:
//
//	Name=Intel(R) UHD Graphics
//	CurrentHorizontalResolution=1920
//	CurrentVerticalResolution=1080
//
// Multiple blank-line-separated records become multiple Info entries.
// The first entry with a resolution is marked Main. Pure function for tests.
//
//nolint:gocognit,gocyclo // WMIC list-format state machine is linear but branch-heavy
func ParseWmicDesktopMonitor(output string) []Info {
	var displays []Info
	var name, w, h string
	flush := func() {
		if name == "" && w == "" && h == "" {
			return
		}
		info := Info{Name: name}
		if w != "" && h != "" {
			info.Resolution = w + " x " + h
		}
		if info.Name != "" || info.Resolution != "" {
			displays = append(displays, info)
		}
		name, w, h = "", "", ""
	}
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			flush()
			continue
		}
		eq := strings.Index(line, "=")
		if eq <= 0 {
			continue
		}
		key := strings.ToLower(strings.TrimSpace(line[:eq]))
		val := strings.TrimSpace(line[eq+1:])
		switch key {
		case "name":
			// New Name within a record without blank separator: flush previous.
			if name != "" || w != "" || h != "" {
				flush()
			}
			name = val
		case "screenwidth", "currenthorizontalresolution":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				w = strconv.Itoa(n)
			}
		case "screenheight", "currentverticalresolution":
			if n, err := strconv.Atoi(val); err == nil && n > 0 {
				h = strconv.Itoa(n)
			}
		}
	}
	flush()
	// Drop entries with neither name nor resolution; prefer those with resolution.
	out := make([]Info, 0, len(displays))
	for _, d := range displays {
		if d.Resolution == "" && d.Name == "" {
			continue
		}
		// Skip zero-resolution adapters (inactive heads often report 0).
		if d.Resolution == "" {
			continue
		}
		out = append(out, d)
	}
	if len(out) > 0 {
		out[0].Main = true
	}
	return out
}

// List returns connected displays for the current platform.
// Canceling ctx aborts the underlying platform query.
func List(ctx context.Context) ([]Info, error) {
	switch runtime.GOOS {
	case "darwin":
		return listMacOS(ctx)
	case "linux":
		return listLinux(ctx)
	case "windows":
		return listWindows(ctx)
	default:
		return nil, ErrUnsupported
	}
}

func listMacOS(ctx context.Context) ([]Info, error) {
	out, err := exec.CommandContext(ctx, "system_profiler", "SPDisplaysDataType").Output()
	if err != nil {
		return nil, err
	}
	return ParseSystemProfilerDisplays(string(out)), nil
}

func listLinux(ctx context.Context) ([]Info, error) {
	if out, err := exec.CommandContext(ctx, "xrandr", "--query").Output(); err == nil {
		if displays := ParseXrandr(string(out)); len(displays) > 0 {
			return displays, nil
		}
	}
	// Wayland fallback.
	if out, err := exec.CommandContext(ctx, "wlr-randr").Output(); err == nil {
		if displays := ParseWlrRandr(string(out)); len(displays) > 0 {
			return displays, nil
		}
	}
	return nil, ErrUnsupported
}

func listWindows(ctx context.Context) ([]Info, error) {
	// #nosec G204 -- fixed wmic path/args, no user input
	out, err := exec.CommandContext(ctx, "wmic", "path", "Win32_VideoController",
		"get", "Name,CurrentHorizontalResolution,CurrentVerticalResolution",
		"/format:list").Output()
	if err != nil {
		return nil, err
	}
	displays := ParseWmicDesktopMonitor(string(out))
	if len(displays) == 0 {
		return nil, ErrUnsupported
	}
	return displays, nil
}
