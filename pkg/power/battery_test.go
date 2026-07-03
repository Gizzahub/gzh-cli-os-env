// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package power

import "testing"

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
