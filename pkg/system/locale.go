// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package system

import (
	"errors"
	"os/exec"
	"runtime"
	"strings"
)

// ErrLocaleUnsupported is returned on platforms without a locale backend.
var ErrLocaleUnsupported = errors.New("system: locale query unsupported on this platform")

// GetLocale returns the user's locale identifier. On macOS it reads
// `defaults read .GlobalPreferences AppleLocale` (e.g. "ko_KR").
func GetLocale() (string, error) {
	switch runtime.GOOS {
	case "darwin":
		return getLocaleMacOS()
	default:
		return "", ErrLocaleUnsupported
	}
}

func getLocaleMacOS() (string, error) {
	out, err := exec.Command("defaults", "read", ".GlobalPreferences", "AppleLocale").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
