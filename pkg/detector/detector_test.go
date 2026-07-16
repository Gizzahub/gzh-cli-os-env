// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package detector

import (
	"runtime"
	"testing"
)

func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("Detect() unexpected error: %v", err)
	}
	if info.OS != runtime.GOOS {
		t.Errorf("OS = %q, want %q", info.OS, runtime.GOOS)
	}
	// Desktop is derived from OS/env; must be a known value and non-empty.
	switch info.Desktop {
	case MacOS, Windows, KDE, GNOME, Unknown:
	default:
		t.Errorf("Desktop = %q, not a known constant", info.Desktop)
	}
}

func TestDetectLinuxDesktopKnownValues(t *testing.T) {
	cases := map[string]string{
		"KDE Plasma":     "plasma-kde",
		"Kubuntu":        "kxsession",
		"GNOME":          "gnome",
		"Ubuntu (GNOME)": "ubuntu:GNOME",
		"Pop!_OS":        "pop:GNOME",
	}
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DESKTOP_SESSION", "")
	for name, session := range cases {
		t.Run(name, func(t *testing.T) {
			t.Setenv("XDG_CURRENT_DESKTOP", session)
			got := detectLinuxDesktop()
			if got == Unknown {
				t.Errorf("detectLinuxDesktop() = Unknown for %q", session)
			}
		})
	}
}

func TestDetectLinuxDesktopEmpty(t *testing.T) {
	t.Setenv("XDG_CURRENT_DESKTOP", "")
	t.Setenv("DESKTOP_SESSION", "")
	if got := detectLinuxDesktop(); got != Unknown {
		t.Errorf("detectLinuxDesktop() = %q, want %q", got, Unknown)
	}
}

func TestInfoString(t *testing.T) {
	info := Info{OS: "darwin", Desktop: MacOS}
	if got, want := info.String(), "darwin/macos"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
