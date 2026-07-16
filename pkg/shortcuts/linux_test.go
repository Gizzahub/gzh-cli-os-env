// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package shortcuts

import "testing"

const gsettingsSample = `org.gnome.desktop.wm.keybindings switch-windows ['<Alt>Tab']
org.gnome.desktop.wm.keybindings close ['<Alt>F4']
org.gnome.desktop.wm.keybindings maximize @as []
org.gnome.desktop.wm.keybindings minimize ['']
org.gnome.desktop.wm.keybindings move-to-workspace-1 ['<Super><Shift>Home', '<Super><Shift>KP_Home']
`

func TestParseGsettingsKeybindings(t *testing.T) {
	got := ParseGsettingsKeybindings(gsettingsSample)
	want := []Shortcut{
		{ID: 0, Name: "close", Enabled: true, Binding: "<Alt>F4"},
		{ID: 0, Name: "maximize", Enabled: false, Binding: ""},
		{ID: 0, Name: "minimize", Enabled: false, Binding: ""},
		{ID: 0, Name: "move-to-workspace-1", Enabled: true, Binding: "<Super><Shift>Home, <Super><Shift>KP_Home"},
		{ID: 0, Name: "switch-windows", Enabled: true, Binding: "<Alt>Tab"},
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

func TestParseGsettingsKeybindingsEmpty(t *testing.T) {
	if got := ParseGsettingsKeybindings(""); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseGsettingsKeybindingsSkipsMalformed(t *testing.T) {
	got := ParseGsettingsKeybindings("not-a-line\norg.gnome.desktop.wm.keybindings only-key\n")
	if len(got) != 0 {
		t.Errorf("got %+v, want empty for malformed lines", got)
	}
}

const kglobalSample = `[kwin]
Walk Through Windows=Alt+Tab,Alt+Tab,Walk Through Windows
Window Close=Alt+F4,Alt+F4,Close Window
Activate Window Demanding Attention=none,Ctrl+Alt+A,Activate Window Demanding Attention
Switch One Desktop Left=Meta+Ctrl+Left;Meta+Alt+Left,Meta+Ctrl+Left,Switch One Desktop Left
# comment line
[plasmashell]
activate task manager entry 1=Meta+1,Meta+1,Activate Task Manager Entry 1
`

func TestParseKGlobalAccel(t *testing.T) {
	got := ParseKGlobalAccel(kglobalSample)
	want := []Shortcut{
		{ID: 0, Name: "Activate Window Demanding Attention", Enabled: false, Binding: ""},
		{ID: 0, Name: "Switch One Desktop Left", Enabled: true, Binding: "Meta+Ctrl+Left, Meta+Alt+Left"},
		{ID: 0, Name: "Walk Through Windows", Enabled: true, Binding: "Alt+Tab"},
		{ID: 0, Name: "Window Close", Enabled: true, Binding: "Alt+F4"},
		{ID: 0, Name: "activate task manager entry 1", Enabled: true, Binding: "Meta+1"},
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

func TestParseKGlobalAccelEmpty(t *testing.T) {
	if got := ParseKGlobalAccel(""); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseKGlobalAccelSkipsSectionsAndComments(t *testing.T) {
	got := ParseKGlobalAccel("[General]\n# hi\n\n")
	if len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}
