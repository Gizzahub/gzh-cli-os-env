// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package backup

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gizzahub/gzh-cli-os-env/pkg/display"
	"github.com/gizzahub/gzh-cli-os-env/pkg/input"
	"github.com/gizzahub/gzh-cli-os-env/pkg/shortcuts"
)

// fakeWriter records Set* calls without touching the host.
type fakeWriter struct {
	locales   []string
	timezones []string
	localeErr error
	tzErr     error
}

func (f *fakeWriter) SetLocale(_ context.Context, locale string) error {
	if f.localeErr != nil {
		return f.localeErr
	}
	f.locales = append(f.locales, locale)
	return nil
}

func (f *fakeWriter) SetTimezone(_ context.Context, tz string) error {
	if f.tzErr != nil {
		return f.tzErr
	}
	f.timezones = append(f.timezones, tz)
	return nil
}

func TestApply_NilSnapshot(t *testing.T) {
	if _, err := Apply(t.Context(), nil, ApplyOptions{}); err == nil {
		t.Fatal("expected error for nil snapshot")
	}
}

func TestApply_DryRun_PlansSupportedDomains(t *testing.T) {
	w := &fakeWriter{}
	s := &Snapshot{
		Version:  1,
		Locale:   "ko_KR.UTF-8",
		Timezone: "Asia/Seoul",
		Keyboard: &input.Keyboard{Sources: []input.Source{{Name: "US"}}},
		Displays: []display.Info{{Name: "Main"}},
		Shortcuts: []shortcuts.Shortcut{
			{Name: "copy", Binding: "Cmd+C"},
		},
	}
	res, err := Apply(t.Context(), s, ApplyOptions{DryRun: true, Writer: w})
	if err != nil {
		t.Fatal(err)
	}
	if len(w.locales) != 0 || len(w.timezones) != 0 {
		t.Fatalf("dry-run must not write: locales=%v timezones=%v", w.locales, w.timezones)
	}
	if len(res.Applied) != 0 {
		t.Fatalf("dry-run Applied should be empty, got %v", res.Applied)
	}
	if got, want := res.Planned, []string{"locale", "timezone"}; !stringSliceEqual(got, want) {
		t.Fatalf("Planned=%v want %v", got, want)
	}
	// Unsupported domains always reported.
	if !hasSkip(res, "keyboard", "no set API") {
		t.Fatalf("expected keyboard skip: %+v", res.Skipped)
	}
	if !hasSkip(res, "displays", "no set API") {
		t.Fatalf("expected displays skip: %+v", res.Skipped)
	}
	if !hasSkip(res, "shortcuts", "no set API") {
		t.Fatalf("expected shortcuts skip: %+v", res.Skipped)
	}
}

func TestApply_DryRun_EmptySupportedSkipped(t *testing.T) {
	w := &fakeWriter{}
	s := &Snapshot{Version: 1}
	res, err := Apply(t.Context(), s, ApplyOptions{DryRun: true, Writer: w})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Planned) != 0 {
		t.Fatalf("Planned should be empty for empty snapshot, got %v", res.Planned)
	}
	if !hasSkip(res, "locale", "empty in snapshot") {
		t.Fatalf("locale skip: %+v", res.Skipped)
	}
	if !hasSkip(res, "timezone", "empty in snapshot") {
		t.Fatalf("timezone skip: %+v", res.Skipped)
	}
	if !hasSkip(res, "keyboard", "empty in snapshot; no set API") {
		t.Fatalf("keyboard empty skip: %+v", res.Skipped)
	}
}

func TestApply_Applied_ViaFakeWriter(t *testing.T) {
	w := &fakeWriter{}
	s := &Snapshot{
		Version:  1,
		Locale:   "en_US.UTF-8",
		Timezone: "UTC",
	}
	res, err := Apply(t.Context(), s, ApplyOptions{Writer: w})
	if err != nil {
		t.Fatal(err)
	}
	if got, want := res.Applied, []string{"locale", "timezone"}; !stringSliceEqual(got, want) {
		t.Fatalf("Applied=%v want %v", got, want)
	}
	if len(res.Planned) != 0 {
		t.Fatalf("Planned should be empty when not dry-run: %v", res.Planned)
	}
	if len(w.locales) != 1 || w.locales[0] != "en_US.UTF-8" {
		t.Fatalf("locales written: %v", w.locales)
	}
	if len(w.timezones) != 1 || w.timezones[0] != "UTC" {
		t.Fatalf("timezones written: %v", w.timezones)
	}
	// Unsupported still skipped even when empty.
	for _, domain := range []string{"keyboard", "displays", "shortcuts"} {
		if !hasSkip(res, domain, "empty in snapshot; no set API") {
			t.Fatalf("expected %s empty skip: %+v", domain, res.Skipped)
		}
	}
}

func TestApply_Skip_UnsupportedNeverSilent(t *testing.T) {
	w := &fakeWriter{}
	s := &Snapshot{
		Version:  1,
		Locale:   "ko_KR",
		Timezone: "Asia/Seoul",
		Keyboard: &input.Keyboard{Sources: []input.Source{{Name: "Korean"}}},
	}
	res, err := Apply(t.Context(), s, ApplyOptions{Writer: w})
	if err != nil {
		t.Fatal(err)
	}
	if !hasSkip(res, "keyboard", "no set API") {
		t.Fatalf("keyboard must not be silent: %+v", res.Skipped)
	}
	// displays/shortcuts empty → explicit empty+no set API
	if !hasSkip(res, "displays", "empty in snapshot; no set API") {
		t.Fatalf("displays: %+v", res.Skipped)
	}
}

func TestApply_WriterError_PartialResult(t *testing.T) {
	w := &fakeWriter{tzErr: errors.New("permission denied")}
	s := &Snapshot{
		Version:  1,
		Locale:   "ko_KR",
		Timezone: "Asia/Seoul",
	}
	res, err := Apply(t.Context(), s, ApplyOptions{Writer: w})
	if err == nil {
		t.Fatal("expected timezone write error")
	}
	if len(res.Applied) != 1 || res.Applied[0] != "locale" {
		t.Fatalf("partial Applied=%v", res.Applied)
	}
	if len(w.locales) != 1 {
		t.Fatalf("locale should have been written before tz fail")
	}
	if len(w.timezones) != 0 {
		t.Fatalf("timezone should not record success on error")
	}
}

func TestBackupApply_DryRunMatchesRestoreContract(t *testing.T) {
	// Guard the Backup/Apply dry-run contract used by CLI restore --apply --dry-run.
	w := &fakeWriter{}
	s := &Snapshot{Version: 1, Locale: "C.UTF-8", Timezone: "Etc/UTC"}
	res, err := Apply(t.Context(), s, ApplyOptions{DryRun: true, Writer: w})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Planned) != 2 {
		t.Fatalf("planned domains: %v", res.Planned)
	}
	if len(res.Applied) != 0 || len(w.locales)+len(w.timezones) != 0 {
		t.Fatal("dry-run must not apply")
	}
}

func hasSkip(res *ApplyResult, domain, reasonSubstr string) bool {
	for _, sk := range res.Skipped {
		if sk.Domain == domain && strings.Contains(sk.Reason, reasonSubstr) {
			return true
		}
	}
	return false
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
