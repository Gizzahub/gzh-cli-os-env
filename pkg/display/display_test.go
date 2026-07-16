// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package display

import (
	"strings"
	"testing"
)

func TestParseSystemProfilerDisplays(t *testing.T) {
	sample := `
Graphics/Displays:

    Apple M1:

      Chipset Model: Apple M1
      Type: GPU
      Displays:
        Color LCD:
          Display Type: Built-In Retina LCD
          Resolution: 2560 x 1600 Retina
          Main Display: Yes
          Mirror: Off
          Online: Yes
        LG HDR QHD:
          Resolution: 2560 x 1440 (QHD/WQHD - Wide Quad High Definition)
          Main Display: No
          Mirror: Off
          Online: Yes
`
	got := ParseSystemProfilerDisplays(sample)
	if len(got) != 2 {
		t.Fatalf("len=%d got=%+v", len(got), got)
	}
	if got[0].Name != "Color LCD" || !got[0].Main {
		t.Fatalf("first=%+v", got[0])
	}
	if !strings.Contains(got[0].Resolution, "2560") {
		t.Fatalf("res=%q", got[0].Resolution)
	}
	if got[1].Name != "LG HDR QHD" || got[1].Main {
		t.Fatalf("second=%+v", got[1])
	}
}

func TestParseSystemProfilerDisplays_Empty(t *testing.T) {
	if got := ParseSystemProfilerDisplays(""); len(got) != 0 {
		t.Fatalf("got=%+v", got)
	}
}

func TestParseXrandr(t *testing.T) {
	sample := `Screen 0: minimum 320 x 200, current 4480 x 1440, maximum 16384 x 16384
eDP-1 connected primary 1920x1080+0+0 (normal left inverted right x axis y axis) 340mm x 190mm
   1920x1080     60.00*+  59.97    59.96    59.93
   1680x1050     59.95    59.88
HDMI-1 connected 2560x1440+1920+0 (normal left inverted right x axis y axis) 600mm x 340mm
   2560x1440     59.95*+
   1920x1080     60.00    50.00
DP-2 disconnected (normal left inverted right x axis y axis)
`
	got := ParseXrandr(sample)
	if len(got) != 2 {
		t.Fatalf("len=%d got=%+v", len(got), got)
	}
	if got[0].Name != "eDP-1" || !got[0].Main {
		t.Errorf("first=%+v", got[0])
	}
	if got[0].Resolution != "1920 x 1080" {
		t.Errorf("first res=%q", got[0].Resolution)
	}
	if got[1].Name != "HDMI-1" || got[1].Main {
		t.Errorf("second=%+v", got[1])
	}
	if got[1].Resolution != "2560 x 1440" {
		t.Errorf("second res=%q", got[1].Resolution)
	}
}

func TestParseXrandr_PrimaryOnlyNoMode(t *testing.T) {
	// Connected but no current mode (disabled output still listed as connected).
	sample := "eDP-1 connected primary (normal left inverted right x axis y axis)\n"
	got := ParseXrandr(sample)
	if len(got) != 1 {
		t.Fatalf("len=%d got=%+v", len(got), got)
	}
	if got[0].Name != "eDP-1" || !got[0].Main {
		t.Errorf("got=%+v", got[0])
	}
	if got[0].Resolution != "" {
		t.Errorf("res=%q, want empty", got[0].Resolution)
	}
}

func TestParseXrandr_Empty(t *testing.T) {
	if got := ParseXrandr(""); len(got) != 0 {
		t.Fatalf("got=%+v", got)
	}
	if got := ParseXrandr("DP-1 disconnected (normal left inverted right x axis y axis)\n"); len(got) != 0 {
		t.Fatalf("disconnected should be ignored, got=%+v", got)
	}
}

func TestParseWlrRandr(t *testing.T) {
	sample := `eDP-1 "Sharp Corporation 0x14D2 (eDP-1)"
  Enabled: yes
    current 1920x1080@60.000000Hz
HDMI-A-1 "LG Electronics LG HDR QHD"
  Enabled: yes
    current 2560x1440@59.951000Hz
`
	got := ParseWlrRandr(sample)
	if len(got) != 2 {
		t.Fatalf("len=%d got=%+v", len(got), got)
	}
	if got[0].Name != "eDP-1" || !got[0].Main {
		t.Errorf("first=%+v", got[0])
	}
	if got[0].Resolution != "1920 x 1080" {
		t.Errorf("first res=%q", got[0].Resolution)
	}
	if got[1].Name != "HDMI-A-1" || got[1].Main {
		t.Errorf("second=%+v", got[1])
	}
	if got[1].Resolution != "2560 x 1440" {
		t.Errorf("second res=%q", got[1].Resolution)
	}
}

func TestParseWlrRandr_Empty(t *testing.T) {
	if got := ParseWlrRandr(""); len(got) != 0 {
		t.Fatalf("got=%+v", got)
	}
}
