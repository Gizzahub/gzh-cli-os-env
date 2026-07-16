// Copyright (c) 2026 Archmagece
// SPDX-License-Identifier: MIT

package osenv

import (
	"errors"
	"fmt"

	"github.com/gizzahub/gzh-cli-os-env/pkg/input"
	"github.com/spf13/cobra"
)

func newInputCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "input",
		Short: "Input device management",
	}
	cmd.AddCommand(newInputKeyboardCmd())
	return cmd
}

func newInputKeyboardCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "keyboard",
		Short: "Show keyboard configuration",
		Long: `Show keyboard repeat settings and enabled input sources.

On macOS this reads NSGlobalDomain and com.apple.HIToolbox. macOS omits a
repeat setting until it is changed, so an unchanged setting is reported as
"not set" rather than as its default value. Other platforms are not yet
supported.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			kb, err := input.GetKeyboard(cmd.Context())
			if err != nil {
				if errors.Is(err, input.ErrUnsupported) {
					return fmt.Errorf("input keyboard is not supported on this platform yet")
				}
				return err
			}

			out := cmd.OutOrStdout()
			if _, err := fmt.Fprintf(out, "Repeat rate:  %s\n", describeRepeat(kb.RepeatRate)); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "Repeat delay: %s\n", describeRepeat(kb.RepeatDelay)); err != nil {
				return err
			}
			if _, err := fmt.Fprintf(out, "Press-and-hold: %s\n", describePressAndHold(kb.PressAndHold)); err != nil {
				return err
			}

			if len(kb.Sources) == 0 {
				_, err := fmt.Fprintln(out, "\nInput sources: none reported")
				return err
			}
			if _, err := fmt.Fprintln(out, "\nInput sources:"); err != nil {
				return err
			}
			for _, s := range kb.Sources {
				if _, err := fmt.Fprintf(out, "  %s (%s)\n", s.Name, s.Kind); err != nil {
					return err
				}
			}
			return nil
		},
	}
}

// describeRepeat renders a repeat setting, keeping "not set" distinct from a
// measured value so the output never presents a default as a reading.
func describeRepeat(s input.Setting) string {
	if !s.Set {
		return "not set (system default)"
	}
	return fmt.Sprintf("%d (%dms)", s.Value, s.Milliseconds())
}

func describePressAndHold(s input.Setting) string {
	if !s.Set {
		return "not set (system default)"
	}
	if s.Value == 0 {
		return "0 (key repeat)"
	}
	return fmt.Sprintf("%d (accent menu)", s.Value)
}
