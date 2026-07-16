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
