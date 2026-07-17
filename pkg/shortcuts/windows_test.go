// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package shortcuts

import "testing"

const windowsHotkeySample = `Name=Copy;Binding=Ctrl+C;Enabled=true
Name=Paste;Binding=Ctrl+V;Enabled=true
Name=Disabled;Binding=Ctrl+X;Enabled=false
Name=Unbound;Binding=;Enabled=true
# comment
`

func TestParseWindowsHotkeyList(t *testing.T) {
	got := ParseWindowsHotkeyList(windowsHotkeySample)
	want := []Shortcut{
		{ID: 0, Name: "Copy", Binding: "Ctrl+C", Enabled: true},
		{ID: 0, Name: "Disabled", Binding: "Ctrl+X", Enabled: false},
		{ID: 0, Name: "Paste", Binding: "Ctrl+V", Enabled: true},
		{ID: 0, Name: "Unbound", Binding: "", Enabled: false},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d shortcuts, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("shortcut[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseWindowsHotkeyListFieldOrder(t *testing.T) {
	got := ParseWindowsHotkeyList("Enabled=false;Binding=Alt+Tab;Name=Switch\n")
	if len(got) != 1 {
		t.Fatalf("got %+v, want one shortcut", got)
	}
	want := Shortcut{ID: 0, Name: "Switch", Binding: "Alt+Tab", Enabled: false}
	if got[0] != want {
		t.Errorf("got %+v, want %+v", got[0], want)
	}
}

func TestParseWindowsHotkeyListEnabledVariants(t *testing.T) {
	got := ParseWindowsHotkeyList("Name=A;Binding=Ctrl+A;Enabled=1\nName=B;Binding=Ctrl+B;Enabled=0\n")
	if len(got) != 2 {
		t.Fatalf("got %+v", got)
	}
	if !got[0].Enabled || got[1].Enabled {
		t.Errorf("enabled flags: %+v", got)
	}
}

func TestParseWindowsHotkeyListOmitsEnabledWhenBound(t *testing.T) {
	got := ParseWindowsHotkeyList("Name=Copy;Binding=Ctrl+C\n")
	if len(got) != 1 || !got[0].Enabled {
		t.Errorf("got %+v, want enabled when binding present and Enabled omitted", got)
	}
}

func TestParseWindowsHotkeyListEmpty(t *testing.T) {
	if got := ParseWindowsHotkeyList(""); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseWindowsHotkeyListSkipsMalformed(t *testing.T) {
	got := ParseWindowsHotkeyList("not-a-line\nName=\nBinding=Ctrl+C;Enabled=true\nName=OnlyName\n")
	// Name=OnlyName has no Binding; Enabled defaults false for empty binding.
	if len(got) != 1 || got[0].Name != "OnlyName" || got[0].Enabled {
		t.Errorf("got %+v, want only OnlyName disabled", got)
	}
}

func TestParseWindowsHotkeyListSortsByName(t *testing.T) {
	got := ParseWindowsHotkeyList("Name=Zed;Binding=Z;Enabled=true\nName=Alpha;Binding=A;Enabled=true\n")
	if len(got) != 2 || got[0].Name != "Alpha" || got[1].Name != "Zed" {
		t.Errorf("not sorted by name: %+v", got)
	}
}
