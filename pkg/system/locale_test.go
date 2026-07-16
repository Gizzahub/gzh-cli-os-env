// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package system

import "testing"

func TestParseLocalectlStatus(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "simple",
			in: `   System Locale: LANG=ko_KR.UTF-8
       VC Keymap: us
      X11 Layout: us
`,
			want: "ko_KR.UTF-8",
		},
		{
			name: "multi_value",
			in: `   System Locale: LANG=en_US.UTF-8
                     LC_TIME=C.UTF-8
       VC Keymap: us
`,
			want: "en_US.UTF-8",
		},
		{
			name: "c_utf8",
			in:   "System Locale: LANG=C.UTF-8\n",
			want: "C.UTF-8",
		},
		{
			name: "empty",
			in:   "VC Keymap: us\n",
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
			if got := ParseLocalectlStatus(tc.in); got != tc.want {
				t.Errorf("ParseLocalectlStatus() = %q, want %q", got, tc.want)
			}
		})
	}
}
