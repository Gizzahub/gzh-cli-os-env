// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

// Package power reports power and battery status. Phase 2 covers battery
// status on macOS via `pmset -g batt`; other platforms return
// ErrUnsupported. Power profiles and sleep settings arrive in later phases.
package power

import (
	"errors"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// ErrUnsupported is returned on platforms without a battery backend.
var ErrUnsupported = errors.New("power: battery query unsupported on this platform")

// BatteryStatus is a snapshot of the current battery state.
type BatteryStatus struct {
	Source   string // "AC" (wall power) or "Battery"
	Percent  int    // 0-100
	Charging bool   // true while drawing from AC and not full
}

var batteryLineRe = regexp.MustCompile(`(\d+)%;\s*(charging|discharging|AC attached|AC Power);`)

// ParseBatteryOutput parses the human-readable output of `pmset -g batt`.
// It is a pure function so the parsing logic can be tested without macOS.
//
// Example input:
//
//	Now drawing from 'Battery Power'
//	 -InternalBattery-0 (id=1)	45%; discharging; 4:15 remaining
func ParseBatteryOutput(output string) (BatteryStatus, error) {
	source := "AC"
	if strings.Contains(output, "Battery Power") {
		source = "Battery"
	}

	m := batteryLineRe.FindStringSubmatch(output)
	if m == nil {
		return BatteryStatus{}, errors.New("power: could not parse battery output")
	}

	pct, err := strconv.Atoi(m[1])
	if err != nil {
		return BatteryStatus{}, errors.New("power: invalid percent in battery output")
	}

	state := m[2]
	charging := state == "charging" || strings.Contains(state, "AC")

	return BatteryStatus{Source: source, Percent: pct, Charging: charging}, nil
}

// GetBattery returns the current battery status, dispatching by platform.
func GetBattery() (BatteryStatus, error) {
	switch runtime.GOOS {
	case "darwin":
		return getBatteryMacOS()
	default:
		return BatteryStatus{}, ErrUnsupported
	}
}

// getBatteryMacOS shells out to `pmset -g batt`.
func getBatteryMacOS() (BatteryStatus, error) {
	out, err := exec.Command("pmset", "-g", "batt").Output()
	if err != nil {
		return BatteryStatus{}, err
	}
	return ParseBatteryOutput(string(out))
}
