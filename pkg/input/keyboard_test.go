// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package input

import "testing"

// liveSources is a capture of HIToolbox AppleEnabledInputSources extracted with
// `plutil -extract AppleEnabledInputSources json`.
const liveSources = `[
  {"InputSourceKind":"Keyboard Layout","KeyboardLayout Name":"ABC","KeyboardLayout ID":252},
  {"Bundle ID":"com.apple.inputmethod.Korean","Input Mode":"com.apple.inputmethod.Korean.2SetKorean","InputSourceKind":"Input Mode"},
  {"Bundle ID":"com.apple.inputmethod.Korean","InputSourceKind":"Keyboard Input Method"}
]`

func TestParseInputSourcesJSON(t *testing.T) {
	got, err := ParseInputSourcesJSON([]byte(liveSources))
	if err != nil {
		t.Fatalf("ParseInputSourcesJSON() error = %v", err)
	}

	want := []Source{
		{Name: "ABC", Kind: "Keyboard Layout"},
		{Name: "com.apple.inputmethod.Korean.2SetKorean", Kind: "Input Mode"},
		{Name: "com.apple.inputmethod.Korean", Kind: "Keyboard Input Method"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d sources, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestParseInputSourcesJSONSkipsUnnamed(t *testing.T) {
	got, err := ParseInputSourcesJSON([]byte(`[{"InputSourceKind":"Keyboard Layout"}]`))
	if err != nil {
		t.Fatalf("ParseInputSourcesJSON() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %+v, want no sources for an entry with no identity", got)
	}
}

func TestParseInputSourcesJSONInvalid(t *testing.T) {
	if _, err := ParseInputSourcesJSON([]byte(`{"not":"an array"}`)); err == nil {
		t.Error("ParseInputSourcesJSON() error = nil, want parse error")
	}
}

func TestParseInputSourcesJSONEmpty(t *testing.T) {
	got, err := ParseInputSourcesJSON([]byte(`[]`))
	if err != nil {
		t.Fatalf("ParseInputSourcesJSON() error = %v", err)
	}
	if len(got) != 0 {
		t.Errorf("got %d sources, want 0", len(got))
	}
}

func TestParseDefaultsInt(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int
		wantErr bool
	}{
		{"plain", "2", 2, false},
		{"trailing newline", "25\n", 25, false},
		{"zero", "0\n", 0, false},
		{"surrounding space", "  6  \n", 6, false},
		{"not a number", "yes\n", 0, true},
		{"empty", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDefaultsInt(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseDefaultsInt(%q) error = %v, wantErr %v", tt.output, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("ParseDefaultsInt(%q) = %d, want %d", tt.output, got, tt.want)
			}
		})
	}
}

func TestSettingMilliseconds(t *testing.T) {
	tests := []struct {
		name    string
		setting Setting
		want    int
	}{
		{"two ticks", Setting{Value: 2, Set: true}, 30},
		{"initial delay", Setting{Value: 25, Set: true}, 375},
		{"unset reads as zero", Setting{}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.setting.Milliseconds(); got != tt.want {
				t.Errorf("Milliseconds() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestSettingUnsetIsDistinctFromZero(t *testing.T) {
	// A key set to 0 and a key never set both carry Value 0; only Set separates
	// them, and the CLI must not print a default as though it were measured.
	unset := Setting{}
	zero := Setting{Set: true}
	if unset.Value != zero.Value {
		t.Fatalf("precondition: both should carry Value 0, got %d and %d", unset.Value, zero.Value)
	}
	if unset.Set == zero.Set {
		t.Error("unset and explicit zero are indistinguishable")
	}
}
