// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package input

import "testing"

func TestParseGsettingsInt(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    int
		wantErr bool
	}{
		{"uint32", "uint32 30\n", 30, false},
		{"int32", "int32 500", 500, false},
		{"bare", "25\n", 25, false},
		{"surrounding space", "  6  \n", 6, false},
		{"zero", "uint32 0\n", 0, false},
		{"not a number", "true\n", 0, true},
		{"empty", "", 0, true},
		{"type only", "uint32\n", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGsettingsInt(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseGsettingsInt(%q) error = %v, wantErr %v", tt.output, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("ParseGsettingsInt(%q) = %d, want %d", tt.output, got, tt.want)
			}
		})
	}
}

const setxkbmapSample = `rules:      evdev
model:      pc105
layout:     us,kr
variant:    ,
options:    grp:alt_shift_toggle
`

func TestParseSetxkbmapQuery(t *testing.T) {
	got := ParseSetxkbmapQuery(setxkbmapSample)
	want := []Source{
		{Name: "us", Kind: "Keyboard Layout"},
		{Name: "kr", Kind: "Keyboard Layout"},
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

func TestParseSetxkbmapQuerySingle(t *testing.T) {
	got := ParseSetxkbmapQuery("layout: us\n")
	if len(got) != 1 || got[0].Name != "us" || got[0].Kind != "Keyboard Layout" {
		t.Errorf("got %+v, want single us layout", got)
	}
}

func TestParseSetxkbmapQueryMissingLayout(t *testing.T) {
	if got := ParseSetxkbmapQuery("rules: evdev\nmodel: pc105\n"); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseSetxkbmapQueryEmptyLayout(t *testing.T) {
	if got := ParseSetxkbmapQuery("layout:\n"); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseSetxkbmapQuerySkipsBlankParts(t *testing.T) {
	got := ParseSetxkbmapQuery("layout: us,,kr,\n")
	want := []Source{
		{Name: "us", Kind: "Keyboard Layout"},
		{Name: "kr", Kind: "Keyboard Layout"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("source[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
