// Copyright (c) 2026 Gizzahub
// SPDX-License-Identifier: MIT

// Package input reports input device configuration. macOS uses defaults;
// Linux covers GNOME (gsettings) and partial KDE (setxkbmap); Windows uses
// Get-WinUserLanguageList and optional KeyboardDelay from the registry.
// Other platforms return ErrUnsupported.
package input

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// ErrUnsupported is returned when the platform has no input backend.
var ErrUnsupported = errors.New("input: query unsupported on this platform")

// errNoInputSources reports that the input source key is absent, which is a
// legitimate state on a machine that never customized it. It is a distinct
// error rather than an empty result so that a genuine query failure can never
// be mistaken for "this machine has no layouts".
var errNoInputSources = errors.New("input: no input sources configured")

// repeatTickMS is the dwell time of one KeyRepeat/InitialKeyRepeat unit.
// macOS stores these as multiples of one 60 Hz frame.
const repeatTickMS = 15

// globalKey is an NSGlobalDomain preference key. The named type keeps the set
// of readable keys closed to these constants.
type globalKey string

const (
	keyRepeat        globalKey = "KeyRepeat"
	keyInitialRepeat globalKey = "InitialKeyRepeat"
	keyPressAndHold  globalKey = "ApplePressAndHoldEnabled"
)

// Setting is an integer preference that may be unset. macOS omits a key
// entirely until the user changes it from the default, so Set distinguishes
// "read as 2" from "never set" — reporting a default as if it were read would
// make the output a guess wearing a measurement's clothes.
type Setting struct {
	Value int
	Set   bool
}

// Milliseconds converts a key-repeat tick count to milliseconds.
func (s Setting) Milliseconds() int { return s.Value * repeatTickMS }

// Source is one enabled keyboard layout or input method.
type Source struct {
	Name string
	Kind string
}

// Keyboard is the keyboard configuration of the current machine.
type Keyboard struct {
	RepeatRate   Setting // KeyRepeat: delay between repeats
	RepeatDelay  Setting // InitialKeyRepeat: delay before repeating starts
	PressAndHold Setting // ApplePressAndHoldEnabled: 1 accent menu, 0 key repeat
	Sources      []Source
}

// rawSource mirrors one entry of HIToolbox AppleEnabledInputSources.
// The tags carry Apple's key names verbatim, so the project's snake_case tag
// convention cannot apply here.
//
//nolint:tagliatelle // external schema: keys are Apple's, not ours
type rawSource struct {
	InputSourceKind string `json:"InputSourceKind"`
	LayoutName      string `json:"KeyboardLayout Name"`
	BundleID        string `json:"Bundle ID"`
	InputMode       string `json:"Input Mode"`
}

// ParseInputSourcesJSON parses HIToolbox AppleEnabledInputSources as JSON.
// Pure function for offline unit tests.
func ParseInputSourcesJSON(data []byte) ([]Source, error) {
	var raw []rawSource
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("input: parse input sources: %w", err)
	}

	out := make([]Source, 0, len(raw))
	for _, r := range raw {
		name := r.LayoutName
		if name == "" {
			// Input methods carry no layout name; the mode is the specific
			// identity, the bundle only the vendor.
			name = r.InputMode
		}
		if name == "" {
			name = r.BundleID
		}
		if name == "" {
			continue
		}
		out = append(out, Source{Name: name, Kind: r.InputSourceKind})
	}
	return out, nil
}

// ParseDefaultsInt parses the scalar `defaults read` prints for an integer key.
func ParseDefaultsInt(output string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(output))
}

// GetKeyboard returns keyboard configuration for the current platform.
// Canceling ctx aborts the underlying platform queries.
func GetKeyboard(ctx context.Context) (*Keyboard, error) {
	switch runtime.GOOS {
	case "darwin":
		return keyboardMacOS(ctx)
	case "linux":
		return keyboardLinux(ctx)
	case "windows":
		return keyboardWindows(ctx)
	default:
		return nil, ErrUnsupported
	}
}

func keyboardMacOS(ctx context.Context) (*Keyboard, error) {
	kb := &Keyboard{
		RepeatRate:   readGlobalInt(ctx, keyRepeat),
		RepeatDelay:  readGlobalInt(ctx, keyInitialRepeat),
		PressAndHold: readGlobalInt(ctx, keyPressAndHold),
	}

	sources, err := inputSourcesMacOS(ctx)
	if err != nil && !errors.Is(err, errNoInputSources) {
		return nil, err
	}
	kb.Sources = sources
	return kb, nil
}

// readGlobalInt reads one NSGlobalDomain integer. `defaults read` exits
// non-zero for a key that was never set, which is the unset signal rather than
// a failure worth propagating.
func readGlobalInt(ctx context.Context, key globalKey) Setting {
	// #nosec G204 -- key is a package-level constant of type globalKey; the
	// command and all arguments are fixed at compile time, never user input.
	out, err := exec.CommandContext(ctx, "defaults", "read", "NSGlobalDomain", string(key)).Output()
	if err != nil {
		return Setting{}
	}
	v, err := ParseDefaultsInt(string(out))
	if err != nil {
		return Setting{}
	}
	return Setting{Value: v, Set: true}
}

// inputSourcesMacOS extracts one subtree as JSON. Converting the whole
// HIToolbox domain fails because it holds binary blobs that JSON cannot
// represent, so plutil -extract narrows the payload to the array we need.
func inputSourcesMacOS(ctx context.Context) ([]Source, error) {
	export := exec.CommandContext(ctx, "defaults", "export", "com.apple.HIToolbox", "-")
	plist, err := export.Output()
	if err != nil {
		return nil, fmt.Errorf("input: export HIToolbox: %w", err)
	}

	extract := exec.CommandContext(ctx, "plutil", "-extract", "AppleEnabledInputSources", "json", "-o", "-", "-")
	extract.Stdin = strings.NewReader(string(plist))
	data, err := extract.Output()
	if err != nil {
		// plutil exits non-zero both for an absent key and for a real failure.
		// Only the first is a legitimate state, so the two are separated here
		// rather than collapsed into an empty result.
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && strings.Contains(string(exitErr.Stderr), "No value at that key path") {
			return nil, errNoInputSources
		}
		return nil, fmt.Errorf("input: extract input sources: %w", err)
	}
	return ParseInputSourcesJSON(data)
}
