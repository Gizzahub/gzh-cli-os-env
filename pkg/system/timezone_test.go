// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import "testing"

func TestParseTimezoneLink_Zoneinfo(t *testing.T) {
	cases := map[string]string{
		"/var/db/timezone/zoneinfo/Asia/Seoul":          "Asia/Seoul",
		"/var/db/timezone/zoneinfo/America/New_York":    "America/New_York",
		"  /var/db/timezone/zoneinfo/Europe/London  \n": "Europe/London",
		// Linux common path
		"/usr/share/zoneinfo/Asia/Seoul":       "Asia/Seoul",
		"/usr/share/zoneinfo/America/New_York": "America/New_York",
		"/etc/zoneinfo/UTC":                    "UTC",
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

func TestParseTzutilOutput(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "korea", in: "Korea Standard Time\n", want: "Korea Standard Time"},
		{name: "padded", in: "  Pacific Standard Time  \r\n", want: "Pacific Standard Time"},
		{name: "empty", in: "\n", want: ""},
		{name: "blank", in: "", want: ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseTzutilOutput(tc.in); got != tc.want {
				t.Errorf("ParseTzutilOutput() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestParseTimedatectlShow(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "asia_seoul",
			in: `Timezone=Asia/Seoul
LocalRTC=no
CanNTP=yes
NTP=yes
NTPSynchronized=yes
TimeUSec=Fri 2026-07-17 12:00:00 KST
RTCTimeUSec=Fri 2026-07-17 03:00:00 UTC
`,
			want: "Asia/Seoul",
		},
		{
			name: "utc",
			in:   "Timezone=UTC\nLocalRTC=no\n",
			want: "UTC",
		},
		{
			name: "empty",
			in:   "LocalRTC=no\n",
			want: "",
		},
		{
			name: "blank",
			in:   "",
			want: "",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := ParseTimedatectlShow(tc.in); got != tc.want {
				t.Errorf("ParseTimedatectlShow() = %q, want %q", got, tc.want)
			}
		})
	}
}
