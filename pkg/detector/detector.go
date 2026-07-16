// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package detector identifies the current operating system and desktop
// environment. Detection is the entry point for every other os-env feature,
// since each platform/DE exposes settings through a different tool
// (dconf, gsettings, defaults, pmset, kwriteconfig5, ...).
package detector

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// Known desktop environment identifiers.
const (
	MacOS   = "macos"
	Windows = "windows"
	KDE     = "kde"
	GNOME   = "gnome"
	Unknown = "unknown"
)

// Info describes the detected environment.
type Info struct {
	OS      string // runtime.GOOS value: darwin, linux, windows, ...
	Desktop string // one of the Known constants
}

// Detect identifies the current OS and, on Linux, the desktop environment.
// It never returns an error for recognized platforms; callers can rely on
// a zero-error result to branch on Desktop.
func Detect() (Info, error) {
	info := Info{OS: runtime.GOOS}
	switch runtime.GOOS {
	case "darwin":
		info.Desktop = MacOS
	case "windows":
		info.Desktop = Windows
	case "linux":
		info.Desktop = detectLinuxDesktop()
	default:
		info.Desktop = Unknown
	}
	return info, nil
}

// detectLinuxDesktop distinguishes KDE Plasma from GNOME using session env
// vars. Falls back to Unknown when neither is detectable (e.g. headless).
func detectLinuxDesktop() string {
	for _, env := range []string{"XDG_CURRENT_DESKTOP", "DESKTOP_SESSION"} {
		v := strings.ToLower(os.Getenv(env))
		switch {
		case strings.Contains(v, "kde"),
			strings.Contains(v, "plasma"),
			strings.Contains(v, "kxsession"):
			return KDE
		case strings.Contains(v, "gnome"),
			strings.Contains(v, "ubuntu"),
			strings.Contains(v, "gdm"),
			strings.Contains(v, "pop"):
			return GNOME
		}
	}
	return Unknown
}

// String returns a compact "os/desktop" representation.
func (i Info) String() string {
	return fmt.Sprintf("%s/%s", i.OS, i.Desktop)
}
