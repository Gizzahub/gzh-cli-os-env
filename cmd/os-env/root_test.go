// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"bytes"
	"testing"
)

// TestRootCmd_HasSubcommands guards the command surface against accidental
// removal during refactors.
func TestRootCmd_HasSubcommands(t *testing.T) {
	cmd := NewRootCmd()
	names := map[string]bool{}
	for _, c := range cmd.Commands() {
		names[c.Name()] = true
	}
	for _, want := range []string{"detect", "power", "system", "display", "shortcuts", "input"} {
		if !names[want] {
			t.Errorf("root command missing subcommand %q", want)
		}
	}
}

// TestDetectCmd_Executes runs the detect subcommand end-to-end. Detect is
// platform-agnostic (returns a known value on every GOOS), so this is safe
// in CI. Platform-specific commands (power battery, system locale) are not
// exercised here because they require macOS.
func TestDetectCmd_Executes(t *testing.T) {
	cmd := NewRootCmd()
	cmd.SetArgs([]string{"detect"})
	out := &bytes.Buffer{}
	cmd.SetOut(out)
	cmd.SetErr(&bytes.Buffer{})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("detect Execute: %v", err)
	}
	if out.Len() == 0 {
		t.Error("detect produced no output")
	}
}
