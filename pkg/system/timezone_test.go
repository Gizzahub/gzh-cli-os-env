// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package system

import "testing"

func TestParseTimezoneLink_Zoneinfo(t *testing.T) {
	cases := map[string]string{
		"/var/db/timezone/zoneinfo/Asia/Seoul":            "Asia/Seoul",
		"/var/db/timezone/zoneinfo/America/New_York":      "America/New_York",
		"  /var/db/timezone/zoneinfo/Europe/London  \n":   "Europe/London",
	}
	for in, want := range cases {
		if got := ParseTimezoneLink(in); got != want {
			t.Errorf("ParseTimezoneLink(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestParseTimezoneLink_Basename(t *testing.T) {
	if got, want := ParseTimezoneLink("/usr/share/zoneinfo/Asia/Tokyo"), "Asia/Tokyo"; got != want {
		t.Errorf("ParseTimezoneLink basename = %q, want %q", got, want)
	}
}

func TestParseTimezoneLink_Plain(t *testing.T) {
	if got, want := ParseTimezoneLink("UTC"), "UTC"; got != want {
		t.Errorf("ParseTimezoneLink plain = %q, want %q", got, want)
	}
}
