// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package osenv

import (
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-os-env/pkg/input"
)

func TestDescribeRepeat(t *testing.T) {
	tests := []struct {
		name    string
		setting input.Setting
		want    string
	}{
		{"set value carries its millisecond conversion", input.Setting{Value: 2, Set: true}, "2 (30ms)"},
		{"initial delay", input.Setting{Value: 25, Set: true}, "25 (375ms)"},
		{"explicit zero is still a reading", input.Setting{Set: true}, "0 (0ms)"},
		{"unset is reported as unset", input.Setting{}, "not set (system default)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describeRepeat(tt.setting); got != tt.want {
				t.Errorf("describeRepeat() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestDescribeRepeatUnsetNamesNoNumber is the load-bearing rule: macOS omits a
// key until it is changed, and printing Apple's default here would present a
// guess as a measurement.
func TestDescribeRepeatUnsetNamesNoNumber(t *testing.T) {
	got := describeRepeat(input.Setting{})
	if strings.ContainsAny(got, "0123456789") {
		t.Errorf("describeRepeat(unset) = %q, must not report any number", got)
	}
}

func TestDescribePressAndHold(t *testing.T) {
	tests := []struct {
		name    string
		setting input.Setting
		want    string
	}{
		{"zero means key repeat", input.Setting{Set: true}, "0 (key repeat)"},
		{"one means accent menu", input.Setting{Value: 1, Set: true}, "1 (accent menu)"},
		{"unset", input.Setting{}, "not set (system default)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := describePressAndHold(tt.setting); got != tt.want {
				t.Errorf("describePressAndHold() = %q, want %q", got, tt.want)
			}
		})
	}
}
