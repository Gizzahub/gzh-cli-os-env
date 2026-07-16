// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package system

import (
	"context"
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// ErrLocaleUnsupported is returned on platforms without a locale backend.
var ErrLocaleUnsupported = errors.New("system: locale query unsupported on this platform")

// GetLocale returns the user's locale identifier. On macOS it reads
// `defaults read .GlobalPreferences AppleLocale` (e.g. "ko_KR").
// Canceling ctx aborts the underlying platform query.
func GetLocale(ctx context.Context) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getLocaleMacOS(ctx)
	default:
		return "", ErrLocaleUnsupported
	}
}

func getLocaleMacOS(ctx context.Context) (string, error) {
	out, err := exec.CommandContext(ctx, "defaults", "read", ".GlobalPreferences", "AppleLocale").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
