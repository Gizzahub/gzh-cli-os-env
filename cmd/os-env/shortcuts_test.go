// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-os-env/pkg/shortcuts"
)

func TestDescribeShortcut(t *testing.T) {
	tests := []struct {
		name     string
		shortcut shortcuts.Shortcut
		want     string
	}{
		{
			"named id renders its name",
			shortcuts.Shortcut{ID: 64, Name: "Show Spotlight search"},
			"Show Spotlight search (id 64)",
		},
		{
			// Apple documents none of these IDs. An unnamed one must read as
			// unknown, never as a guess.
			"unnamed id renders as unknown",
			shortcuts.Shortcut{ID: 15},
			"unknown shortcut (id 15)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeShortcut(tt.shortcut); got != tt.want {
				t.Errorf("describeShortcut() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDescribeBinding(t *testing.T) {
	tests := []struct {
		name     string
		shortcut shortcuts.Shortcut
		want     string
	}{
		{
			"bound and enabled",
			shortcuts.Shortcut{Binding: "Cmd+Space", Enabled: true},
			"Cmd+Space [enabled]",
		},
		{
			"bound and disabled",
			shortcuts.Shortcut{Binding: "Cmd+Space"},
			"Cmd+Space [disabled]",
		},
		{
			// An empty binding means macOS stores no override, not that the
			// shortcut has no key.
			"empty binding is the system default, not 'none'",
			shortcuts.Shortcut{Enabled: true},
			"system default binding [enabled]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeBinding(tt.shortcut); got != tt.want {
				t.Errorf("describeBinding() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDescribeBindingNeverClaimsNoShortcut guards the wording: "no shortcut"
// would be a claim the data cannot support.
func TestDescribeBindingNeverClaimsNoShortcut(t *testing.T) {
	got := describeBinding(shortcuts.Shortcut{})
	if strings.Contains(got, "no shortcut") || strings.Contains(got, "none") {
		t.Errorf("describeBinding() = %q, must not claim the shortcut is unset", got)
	}
}
