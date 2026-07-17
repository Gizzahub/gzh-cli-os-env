// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package power reports power and battery status. Phase 2 covers battery
// status on macOS via `pmset -g batt`; Phase 4 adds Linux via sysfs
// (/sys/class/power_supply/BAT*) with optional upower fallback; Phase 5
// adds Windows via WMIC Win32_Battery with PowerShell CIM fallback.
// Other platforms return ErrUnsupported.
package power

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// ErrUnsupported is returned on platforms without a battery backend.
var ErrUnsupported = errors.New("power: battery query unsupported on this platform")

// ErrNoBattery is returned when no battery is present (desktop Linux, etc.).
var ErrNoBattery = errors.New("power: no battery")

const (
	sourceAC      = "AC"
	sourceBattery = "Battery"
)

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
	source := sourceAC
	if strings.Contains(output, "Battery Power") {
		source = sourceBattery
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
// Canceling ctx aborts the underlying platform query.
func GetBattery(ctx context.Context) (BatteryStatus, error) {
	switch runtime.GOOS {
	case "darwin":
		return getBatteryMacOS(ctx)
	case "linux":
		return getBatteryLinux(ctx)
	case "windows":
		return getBatteryWindows(ctx)
	default:
		return BatteryStatus{}, ErrUnsupported
	}
}

// getBatteryMacOS shells out to `pmset -g batt`.
func getBatteryMacOS(ctx context.Context) (BatteryStatus, error) {
	out, err := exec.CommandContext(ctx, "pmset", "-g", "batt").Output()
	if err != nil {
		return BatteryStatus{}, err
	}
	return ParseBatteryOutput(string(out))
}

// sysfsPowerSupply is the Linux power_supply root (overridable in tests).
var sysfsPowerSupply = "/sys/class/power_supply"

// ParseSysfsCapacity parses a BAT*/capacity file (0-100 integer).
func ParseSysfsCapacity(content string) (int, error) {
	pct, err := strconv.Atoi(strings.TrimSpace(content))
	if err != nil {
		return 0, errors.New("power: invalid sysfs capacity")
	}
	if pct < 0 || pct > 100 {
		return 0, errors.New("power: capacity out of range")
	}
	return pct, nil
}

// ParseSysfsStatus maps a BAT*/status value to Source and Charging.
func ParseSysfsStatus(content string) (source string, charging bool) {
	s := strings.TrimSpace(strings.ToLower(content))
	switch s {
	case "charging":
		return sourceAC, true
	case "full", "not charging", "fully-charged", "fully charged":
		return sourceAC, false
	case "discharging":
		return sourceBattery, false
	default:
		return sourceBattery, false
	}
}

// ParseUPowerDump extracts percent and state from `upower -i` text output.
func ParseUPowerDump(output string) (BatteryStatus, error) {
	var pct int
	var havePct bool
	state := ""
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "percentage:") {
			f := strings.TrimSpace(strings.TrimPrefix(line, "percentage:"))
			f = strings.TrimSuffix(f, "%")
			f = strings.TrimSpace(f)
			n, err := strconv.Atoi(f)
			if err == nil {
				pct = n
				havePct = true
			}
		}
		if strings.HasPrefix(line, "state:") {
			state = strings.TrimSpace(strings.TrimPrefix(line, "state:"))
		}
	}
	if !havePct {
		return BatteryStatus{}, errors.New("power: could not parse upower percentage")
	}
	source, charging := ParseSysfsStatus(state)
	return BatteryStatus{Source: source, Percent: pct, Charging: charging}, nil
}

func readSysfsBattery() (BatteryStatus, error) {
	entries, err := os.ReadDir(sysfsPowerSupply)
	if err != nil {
		return BatteryStatus{}, ErrNoBattery
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, "BAT") {
			continue
		}
		base := filepath.Join(sysfsPowerSupply, name)
		// #nosec G304 -- path is under fixed sysfsPowerSupply + BAT* entry name
		capB, err := os.ReadFile(filepath.Join(base, "capacity"))
		if err != nil {
			continue
		}
		pct, err := ParseSysfsCapacity(string(capB))
		if err != nil {
			continue
		}
		// #nosec G304 -- status file under same BAT* directory
		stB, err := os.ReadFile(filepath.Join(base, "status"))
		if err != nil {
			stB = []byte("Unknown")
		}
		source, charging := ParseSysfsStatus(string(stB))
		return BatteryStatus{Source: source, Percent: pct, Charging: charging}, nil
	}
	return BatteryStatus{}, ErrNoBattery
}

func getBatteryLinux(ctx context.Context) (BatteryStatus, error) {
	if s, err := readSysfsBattery(); err == nil {
		return s, nil
	}

	out, err := exec.CommandContext(ctx, "upower", "-e").Output()
	if err != nil {
		return BatteryStatus{}, ErrNoBattery
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || !strings.Contains(strings.ToLower(line), "battery") {
			continue
		}
		// #nosec G204 -- line is a device path from upower -e, not shell input
		info, err := exec.CommandContext(ctx, "upower", "-i", line).Output()
		if err != nil {
			continue
		}
		return ParseUPowerDump(string(info))
	}
	return BatteryStatus{}, ErrNoBattery
}

// mapWmicBatteryStatus maps Win32_Battery.BatteryStatus codes to Source and Charging.
//
//	1 = Discharging (Battery)
//	2 = On AC, not charging (AC)
//	3–5 = Fully charged / low / critical (AC when plugged, treat as AC not charging)
//	6–9 = Charging variants (AC + Charging)
//	Other = Battery, not charging
func mapWmicBatteryStatus(code int) (source string, charging bool) {
	switch code {
	case 1: // Discharging
		return sourceBattery, false
	case 2: // On AC Power
		return sourceAC, false
	case 3, 4, 5: // Fully Charged / Low / Critical
		return sourceAC, false
	case 6, 7, 8, 9: // Charging / Charging High / Charging Low / Charging Critical
		return sourceAC, true
	default:
		return sourceBattery, false
	}
}

// parseWmicKVList parses WMIC /format:list or PowerShell "Key : Value" lines
// into a flat key→value map (first occurrence wins; keys lowercased).
func parseWmicKVList(output string) map[string]string {
	kv := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Prefer "Key=Value" (WMIC list); fall back to "Key : Value" (PowerShell).
		var key, val string
		if i := strings.Index(line, "="); i > 0 && !strings.Contains(line[:i], " ") {
			key = strings.TrimSpace(line[:i])
			val = strings.TrimSpace(line[i+1:])
		} else if i := strings.Index(line, ":"); i > 0 {
			key = strings.TrimSpace(line[:i])
			val = strings.TrimSpace(line[i+1:])
		} else {
			continue
		}
		if key == "" {
			continue
		}
		lk := strings.ToLower(key)
		if _, exists := kv[lk]; !exists {
			kv[lk] = val
		}
	}
	return kv
}

// batteryStatusFromKV builds BatteryStatus from EstimatedChargeRemaining + BatteryStatus keys.
func batteryStatusFromKV(kv map[string]string) (BatteryStatus, error) {
	pctStr, ok := kv["estimatedchargeremaining"]
	if !ok || pctStr == "" {
		return BatteryStatus{}, errors.New("power: could not parse battery percent")
	}
	pct, err := strconv.Atoi(pctStr)
	if err != nil {
		return BatteryStatus{}, errors.New("power: invalid battery percent")
	}
	if pct < 0 || pct > 100 {
		return BatteryStatus{}, errors.New("power: capacity out of range")
	}

	code := 0
	if s, ok := kv["batterystatus"]; ok && s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			code = n
		}
	}
	source, charging := mapWmicBatteryStatus(code)
	return BatteryStatus{Source: source, Percent: pct, Charging: charging}, nil
}

// ParseWmicBatteryList parses WMIC list-format output for Win32_Battery:
//
//	EstimatedChargeRemaining=85
//	BatteryStatus=2
//
// Pure function for offline unit tests.
func ParseWmicBatteryList(output string) (BatteryStatus, error) {
	return batteryStatusFromKV(parseWmicKVList(output))
}

// ParsePowerShellBattery parses PowerShell Get-CimInstance style output:
//
//	EstimatedChargeRemaining : 85
//	BatteryStatus            : 2
//
// Pure function for offline unit tests.
func ParsePowerShellBattery(output string) (BatteryStatus, error) {
	return batteryStatusFromKV(parseWmicKVList(output))
}

// getBatteryWindows tries WMIC Win32_Battery, then PowerShell Get-CimInstance.
func getBatteryWindows(ctx context.Context) (BatteryStatus, error) {
	// #nosec G204 -- fixed wmic path/args, no user input
	out, err := exec.CommandContext(ctx, "wmic", "path", "Win32_Battery",
		"get", "EstimatedChargeRemaining,BatteryStatus", "/format:list").Output()
	if err == nil {
		if s, perr := ParseWmicBatteryList(string(out)); perr == nil {
			return s, nil
		}
	}

	// PowerShell CIM fallback (WMIC is deprecated on newer Windows).
	ps := "(Get-CimInstance -ClassName Win32_Battery |" +
		" Select-Object EstimatedChargeRemaining,BatteryStatus |" +
		" Format-List) | Out-String"
	// #nosec G204 -- fixed PowerShell command string
	out, err = exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", ps).Output()
	if err != nil {
		return BatteryStatus{}, ErrNoBattery
	}
	s, perr := ParsePowerShellBattery(string(out))
	if perr != nil {
		return BatteryStatus{}, ErrNoBattery
	}
	return s, nil
}
