// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package shortcuts

import "testing"

func TestFormatBinding(t *testing.T) {
	tests := []struct {
		name       string
		parameters []int
		want       string
	}{
		// Values captured from a live macOS system.
		{"spotlight cmd+space", []int{65535, 49, 1048576}, "Cmd+Space"},
		{"finder search cmd+option+space", []int{65535, 49, 1572864}, "Option+Cmd+Space"},
		{"previous input source ctrl+space", []int{32, 49, 262144}, "Ctrl+Space"},
		{"next input source ctrl+option+space", []int{32, 49, 786432}, "Ctrl+Option+Space"},
		{"unset parameters", []int{65535, 65535, 0}, ""},
		{"printable ascii", []int{52, 21, 1179648}, "Shift+Cmd+4"},
		{"no modifiers", []int{65535, 36, 0}, "Return"},
		{"all modifiers", []int{65535, 48, 0x1E0000}, "Ctrl+Option+Shift+Cmd+Tab"},
		{"unknown key code", []int{65535, 999, 1048576}, "Cmd+Key999"},
		{"short slice", []int{65535, 49}, ""},
		{"nil slice", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatBinding(tt.parameters); got != tt.want {
				t.Errorf("FormatBinding(%v) = %q, want %q", tt.parameters, got, tt.want)
			}
		})
	}
}

// liveDump is a trimmed capture of `defaults export com.apple.symbolichotkeys -`
// piped through `plutil -convert json`.
const liveDump = `{"AppleSymbolicHotKeys":{
  "15":{"enabled":false},
  "60":{"enabled":true,"value":{"type":"standard","parameters":[32,49,262144]}},
  "64":{"enabled":true,"value":{"type":"standard","parameters":[65535,49,1048576]}},
  "79":{"enabled":true},
  "164":{"enabled":false,"value":{"type":"standard","parameters":[65535,65535,0]}}
}}`

func TestParseSymbolicHotkeysJSON(t *testing.T) {
	got, err := ParseSymbolicHotkeysJSON([]byte(liveDump))
	if err != nil {
		t.Fatalf("ParseSymbolicHotkeysJSON() error = %v", err)
	}

	want := []Shortcut{
		{ID: 15, Name: "", Enabled: false, Binding: ""},
		{ID: 60, Name: "Select the previous input source", Enabled: true, Binding: "Ctrl+Space"},
		{ID: 64, Name: "Show Spotlight search", Enabled: true, Binding: "Cmd+Space"},
		{ID: 79, Name: "Move left a space", Enabled: true, Binding: ""},
		{ID: 164, Name: "", Enabled: false, Binding: ""},
	}

	if len(got) != len(want) {
		t.Fatalf("got %d shortcuts, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("shortcut[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseSymbolicHotkeysJSONSortsByID(t *testing.T) {
	// Map iteration is random; IDs must come back ascending and numeric so that
	// 164 does not sort between 15 and 60 as it would lexically.
	got, err := ParseSymbolicHotkeysJSON([]byte(liveDump))
	if err != nil {
		t.Fatalf("ParseSymbolicHotkeysJSON() error = %v", err)
	}
	for i := 1; i < len(got); i++ {
		if got[i-1].ID >= got[i].ID {
			t.Fatalf("not sorted ascending: %d before %d", got[i-1].ID, got[i].ID)
		}
	}
}

func TestParseSymbolicHotkeysJSONSkipsNonNumericKeys(t *testing.T) {
	got, err := ParseSymbolicHotkeysJSON([]byte(`{"AppleSymbolicHotKeys":{"AppleSymbolicHotKeysAlt":{"enabled":true},"64":{"enabled":true}}}`))
	if err != nil {
		t.Fatalf("ParseSymbolicHotkeysJSON() error = %v", err)
	}
	if len(got) != 1 || got[0].ID != 64 {
		t.Errorf("got %+v, want only ID 64", got)
	}
}

func TestParseSymbolicHotkeysJSONInvalid(t *testing.T) {
	if _, err := ParseSymbolicHotkeysJSON([]byte(`not json`)); err == nil {
		t.Error("ParseSymbolicHotkeysJSON() error = nil, want parse error")
	}
}

func TestParseSymbolicHotkeysJSONEmptyDomain(t *testing.T) {
	got, err := ParseSymbolicHotkeysJSON([]byte(`{}`))
	if err != nil {
		t.Fatalf("ParseSymbolicHotkeysJSON() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d shortcuts, want 0", len(got))
	}
}

func TestShortcutNamed(t *testing.T) {
	if !(Shortcut{ID: 64, Name: "Show Spotlight search"}).Named() {
		t.Error("Named() = false for a mapped ID, want true")
	}
	if (Shortcut{ID: 999}).Named() {
		t.Error("Named() = true for an unmapped ID, want false")
	}
}
