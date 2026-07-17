// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package power

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestParseBatteryOutput_Discharging(t *testing.T) {
	in := "Now drawing from 'Battery Power'\n -InternalBattery-0 (id=1)\t45%; discharging; 4:15 remaining\n"
	s, err := ParseBatteryOutput(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Source != "Battery" {
		t.Errorf("Source = %q, want %q", s.Source, "Battery")
	}
	if s.Percent != 45 {
		t.Errorf("Percent = %d, want 45", s.Percent)
	}
	if s.Charging {
		t.Error("Charging = true, want false")
	}
}

func TestParseBatteryOutput_Charging(t *testing.T) {
	in := "Now drawing from 'AC Power'\n -InternalBattery-0 (id=1)\t80%; charging; 0:30 remaining\n"
	s, err := ParseBatteryOutput(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Source != "AC" {
		t.Errorf("Source = %q, want %q", s.Source, "AC")
	}
	if s.Percent != 80 {
		t.Errorf("Percent = %d, want 80", s.Percent)
	}
	if !s.Charging {
		t.Error("Charging = false, want true")
	}
}

func TestParseBatteryOutput_ACAttached(t *testing.T) {
	in := "Now drawing from 'AC Power'\n -InternalBattery-0 (id=1)\t100%; AC attached; (charged)\n"
	s, err := ParseBatteryOutput(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 100 {
		t.Errorf("Percent = %d, want 100", s.Percent)
	}
	if !s.Charging {
		t.Error("AC attached should report Charging = true")
	}
}

func TestParseBatteryOutput_Invalid(t *testing.T) {
	if _, err := ParseBatteryOutput("no battery info here"); err == nil {
		t.Fatal("expected error for unparseable input, got nil")
	}
}

func TestParseSysfsCapacity(t *testing.T) {
	cases := []struct {
		in      string
		want    int
		wantErr bool
	}{
		{"85\n", 85, false},
		{"  100  ", 100, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"150", 0, true},
		{"-1", 0, true},
	}
	for _, tc := range cases {
		got, err := ParseSysfsCapacity(tc.in)
		if tc.wantErr {
			if err == nil {
				t.Errorf("ParseSysfsCapacity(%q) expected error", tc.in)
			}
			continue
		}
		if err != nil {
			t.Errorf("ParseSysfsCapacity(%q) unexpected error: %v", tc.in, err)
			continue
		}
		if got != tc.want {
			t.Errorf("ParseSysfsCapacity(%q) = %d, want %d", tc.in, got, tc.want)
		}
	}
}

func TestParseSysfsStatus(t *testing.T) {
	cases := []struct {
		in       string
		source   string
		charging bool
	}{
		{"Charging\n", "AC", true},
		{"Full", "AC", false},
		{"Discharging", "Battery", false},
		{"Not charging", "AC", false},
		{"Unknown", "Battery", false},
		{"", "Battery", false},
	}
	for _, tc := range cases {
		src, ch := ParseSysfsStatus(tc.in)
		if src != tc.source || ch != tc.charging {
			t.Errorf("ParseSysfsStatus(%q) = (%q, %v), want (%q, %v)",
				tc.in, src, ch, tc.source, tc.charging)
		}
	}
}

func TestParseUPowerDump(t *testing.T) {
	in := `
  native-path:          BAT0
  power supply:         yes
  updated:              Fri 17 Jul 2026 12:00:00 PM KST (1 seconds ago)
  has history:          yes
  has statistics:       yes
  battery
    present:             yes
    rechargeable:        yes
    state:               discharging
    warning-level:       none
    energy:              40.5 Wh
    energy-empty:        0 Wh
    energy-full:         54.0 Wh
    energy-full-design:  54.0 Wh
    energy-rate:         12.0 W
    voltage:             11.4 V
    percentage:          75%
    capacity:            100%
    technology:          lithium-ion
`
	s, err := ParseUPowerDump(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 75 {
		t.Errorf("Percent = %d, want 75", s.Percent)
	}
	if s.Source != "Battery" {
		t.Errorf("Source = %q, want Battery", s.Source)
	}
	if s.Charging {
		t.Error("Charging = true, want false")
	}
}

func TestParseUPowerDump_Charging(t *testing.T) {
	in := "  state:               charging\n  percentage:          42%\n"
	s, err := ParseUPowerDump(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 42 || s.Source != "AC" || !s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParseUPowerDump_FullyCharged(t *testing.T) {
	in := "  state:               fully-charged\n  percentage:          100%\n"
	s, err := ParseUPowerDump(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 100 || s.Source != "AC" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParseUPowerDump_Invalid(t *testing.T) {
	if _, err := ParseUPowerDump("no percentage here"); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestReadSysfsBattery(t *testing.T) {
	root := t.TempDir()
	bat0 := filepath.Join(root, "BAT0")
	if err := os.MkdirAll(bat0, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bat0, "capacity"), []byte("67\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bat0, "status"), []byte("Discharging\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	old := sysfsPowerSupply
	sysfsPowerSupply = root
	t.Cleanup(func() { sysfsPowerSupply = old })

	s, err := readSysfsBattery()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 67 || s.Source != "Battery" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestReadSysfsBattery_NoBattery(t *testing.T) {
	old := sysfsPowerSupply
	sysfsPowerSupply = t.TempDir()
	t.Cleanup(func() { sysfsPowerSupply = old })

	if _, err := readSysfsBattery(); !errors.Is(err, ErrNoBattery) {
		t.Fatalf("got err=%v, want ErrNoBattery", err)
	}
}

func TestParseWmicBatteryList_AC(t *testing.T) {
	in := "EstimatedChargeRemaining=85\nBatteryStatus=2\n"
	s, err := ParseWmicBatteryList(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 85 || s.Source != "AC" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParseWmicBatteryList_Discharging(t *testing.T) {
	in := `
EstimatedChargeRemaining=45
BatteryStatus=1
`
	s, err := ParseWmicBatteryList(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 45 || s.Source != "Battery" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParseWmicBatteryList_Charging(t *testing.T) {
	for _, code := range []int{6, 7, 8, 9} {
		in := "EstimatedChargeRemaining=60\nBatteryStatus=" + strconv.Itoa(code) + "\n"
		s, err := ParseWmicBatteryList(in)
		if err != nil {
			t.Fatalf("code %d: unexpected error: %v", code, err)
		}
		if s.Percent != 60 || s.Source != "AC" || !s.Charging {
			t.Errorf("code %d: got %+v", code, s)
		}
	}
}

func TestParseWmicBatteryList_Invalid(t *testing.T) {
	if _, err := ParseWmicBatteryList("no battery fields"); err == nil {
		t.Fatal("expected error, got nil")
	}
	if _, err := ParseWmicBatteryList("EstimatedChargeRemaining=abc\nBatteryStatus=2\n"); err == nil {
		t.Fatal("expected error for invalid percent")
	}
}

func TestParsePowerShellBattery(t *testing.T) {
	in := `
EstimatedChargeRemaining : 85
BatteryStatus            : 2
`
	s, err := ParsePowerShellBattery(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 85 || s.Source != "AC" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParsePowerShellBattery_Charging(t *testing.T) {
	in := "EstimatedChargeRemaining : 42\nBatteryStatus            : 6\n"
	s, err := ParsePowerShellBattery(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 42 || s.Source != "AC" || !s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParsePowerShellBattery_Discharging(t *testing.T) {
	in := "EstimatedChargeRemaining : 30\nBatteryStatus            : 1\n"
	s, err := ParsePowerShellBattery(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.Percent != 30 || s.Source != "Battery" || s.Charging {
		t.Errorf("got %+v", s)
	}
}

func TestParsePowerShellBattery_Invalid(t *testing.T) {
	if _, err := ParsePowerShellBattery(""); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestParseWmicBatteryList(t *testing.T) {
	in := "\nEstimatedChargeRemaining=85\nBatteryStatus=6\n\n"
	s, err := ParseWmicBatteryList(in)
	if err != nil {
		t.Fatal(err)
	}
	if s.Percent != 85 || s.Source != "AC" || !s.Charging {
		t.Fatalf("%+v", s)
	}
}
