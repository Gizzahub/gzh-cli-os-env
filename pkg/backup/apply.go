// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

package backup

import (
	"context"
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/system"
)

// DomainWriter applies supported snapshot domain values to the host.
// Inject a fake in unit tests; nil Writer in ApplyOptions uses the real
// system package backends.
type DomainWriter interface {
	SetLocale(ctx context.Context, locale string) error
	SetTimezone(ctx context.Context, tz string) error
}

// ApplyOptions controls Apply behavior.
type ApplyOptions struct {
	// DryRun plans supported domains without writing.
	DryRun bool
	// Writer is optional; nil uses system.SetLocale / system.SetTimezone.
	Writer DomainWriter
}

// Skip records a domain that was not applied and why.
type Skip struct {
	Domain string
	Reason string
}

// ApplyResult summarizes what Apply did (or would do in dry-run).
type ApplyResult struct {
	// Applied lists domains written successfully (empty when DryRun).
	Applied []string
	// Skipped lists domains not written with explicit reasons.
	Skipped []Skip
	// Planned lists domains that would be written (DryRun only).
	Planned []string
}

// systemWriter is the default DomainWriter backed by pkg/system.
type systemWriter struct{}

func (systemWriter) SetLocale(ctx context.Context, locale string) error {
	return system.SetLocale(ctx, locale)
}

func (systemWriter) SetTimezone(ctx context.Context, tz string) error {
	return system.SetTimezone(ctx, tz)
}

// Apply writes supported domains from snapshot onto the host.
//
// Supported write domains today: locale, timezone (macOS/Linux via pkg/system).
// keyboard, displays, and shortcuts are always skipped with an explicit reason
// (no set API yet). Empty supported-domain values are skipped as empty.
//
// On write failure the partial ApplyResult is returned with the error so
// callers can report what already applied. DryRun never calls Writer.
func Apply(ctx context.Context, snapshot *Snapshot, opts ApplyOptions) (*ApplyResult, error) {
	if snapshot == nil {
		return nil, fmt.Errorf("nil snapshot")
	}
	w := opts.Writer
	if w == nil {
		w = systemWriter{}
	}
	res := &ApplyResult{}

	if err := applyLocale(ctx, w, snapshot.Locale, opts.DryRun, res); err != nil {
		return res, err
	}
	if err := applyTimezone(ctx, w, snapshot.Timezone, opts.DryRun, res); err != nil {
		return res, err
	}

	// Unsupported domains: never silent — always report skip reason.
	skipUnsupported(res, "keyboard", snapshot.Keyboard != nil)
	skipUnsupported(res, "displays", len(snapshot.Displays) > 0)
	skipUnsupported(res, "shortcuts", len(snapshot.Shortcuts) > 0)

	return res, nil
}

func applyLocale(ctx context.Context, w DomainWriter, locale string, dryRun bool, res *ApplyResult) error {
	if locale == "" {
		res.Skipped = append(res.Skipped, Skip{Domain: "locale", Reason: "empty in snapshot"})
		return nil
	}
	if dryRun {
		res.Planned = append(res.Planned, "locale")
		return nil
	}
	if err := w.SetLocale(ctx, locale); err != nil {
		return fmt.Errorf("apply locale: %w", err)
	}
	res.Applied = append(res.Applied, "locale")
	return nil
}

func applyTimezone(ctx context.Context, w DomainWriter, tz string, dryRun bool, res *ApplyResult) error {
	if tz == "" {
		res.Skipped = append(res.Skipped, Skip{Domain: "timezone", Reason: "empty in snapshot"})
		return nil
	}
	if dryRun {
		res.Planned = append(res.Planned, "timezone")
		return nil
	}
	if err := w.SetTimezone(ctx, tz); err != nil {
		return fmt.Errorf("apply timezone: %w", err)
	}
	res.Applied = append(res.Applied, "timezone")
	return nil
}

func skipUnsupported(res *ApplyResult, domain string, present bool) {
	reason := "no set API"
	if !present {
		reason = "empty in snapshot; no set API"
	}
	res.Skipped = append(res.Skipped, Skip{Domain: domain, Reason: reason})
}
