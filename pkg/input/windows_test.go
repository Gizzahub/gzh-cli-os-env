// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package input

import "testing"

func TestParsePowerShellLanguageListFormatList(t *testing.T) {
	in := "LanguageTag : en-US\nLanguageTag : ko-KR\n"
	got := ParsePowerShellLanguageList(in)
	want := []Source{
		{Name: "en-US", Kind: "LanguageTag"},
		{Name: "ko-KR", Kind: "LanguageTag"},
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

func TestParsePowerShellLanguageListBareLines(t *testing.T) {
	got := ParsePowerShellLanguageList("en-US\nko-KR\n")
	if len(got) != 2 || got[0].Name != "en-US" || got[1].Name != "ko-KR" {
		t.Fatalf("got %+v", got)
	}
	if got[0].Kind != "LanguageTag" {
		t.Errorf("kind = %q, want LanguageTag", got[0].Kind)
	}
}

func TestParsePowerShellLanguageListDedupes(t *testing.T) {
	got := ParsePowerShellLanguageList("en-US\nen-US\nLanguageTag : en-US\n")
	if len(got) != 1 || got[0].Name != "en-US" {
		t.Errorf("got %+v, want single en-US", got)
	}
}

func TestParsePowerShellLanguageListSkipsNoise(t *testing.T) {
	in := "EstimatedChargeRemaining : 85\nLanguageTag : en-US\nnot a tag\n# comment\n"
	got := ParsePowerShellLanguageList(in)
	if len(got) != 1 || got[0].Name != "en-US" {
		t.Errorf("got %+v, want only en-US", got)
	}
}

func TestParsePowerShellLanguageListEmpty(t *testing.T) {
	if got := ParsePowerShellLanguageList(""); len(got) != 0 {
		t.Errorf("got %+v, want empty", got)
	}
}

func TestParseKeyboardDelay(t *testing.T) {
	tests := []struct {
		name    string
		output  string
		want    Setting
		wantErr bool
	}{
		{"equals", "KeyboardDelay=1\n", Setting{Value: 1, Set: true}, false},
		{"colon", "KeyboardDelay : 2\n", Setting{Value: 2, Set: true}, false},
		{"bare", "3\n", Setting{Value: 3, Set: true}, false},
		{"zero", "KeyboardDelay=0", Setting{Value: 0, Set: true}, false},
		{"empty", "", Setting{}, true},
		{"not a number", "KeyboardDelay=slow\n", Setting{}, true},
		{"wrong key", "KeyboardSpeed=31\n", Setting{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseKeyboardDelay(tt.output)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseKeyboardDelay(%q) error = %v, wantErr %v", tt.output, err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("ParseKeyboardDelay(%q) = %+v, want %+v", tt.output, got, tt.want)
			}
		})
	}
}
