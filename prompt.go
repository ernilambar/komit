package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/huh/v2"
)

const (
	pickRegenerate = -1
	pickWriteOwn   = -2
)

func stdinIsTTY() bool {
	info, err := os.Stdin.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

// chooseSubject runs the interactive selection loop and returns the subject to
// commit. It returns huh.ErrUserAborted when the user cancels.
func chooseSubject(cfg config, input string, n int) (string, error) {
	subjects, err := suggestSubjects(cfg, input, n)
	writeOwn := errors.Is(err, errNoSubjects)
	if err != nil && !writeOwn {
		return "", err
	}

	for {
		if !writeOwn {
			pick, err := selectSubject(subjects)
			if err != nil {
				return "", err
			}
			switch pick {
			case pickRegenerate:
				subjects, err = suggestSubjects(cfg, input, n)
				writeOwn = errors.Is(err, errNoSubjects)
				if err != nil && !writeOwn {
					return "", err
				}
				continue
			case pickWriteOwn:
			default:
				return subjects[pick], nil
			}
		}
		writeOwn = false

		prefill := ""
		if len(subjects) > 0 {
			prefill = subjects[0]
		}
		subject, err := inputSubject(prefill)
		if err != nil {
			return "", err
		}
		if subject != "" {
			return subject, nil
		}
	}
}

func selectSubject(subjects []string) (int, error) {
	options := make([]huh.Option[int], 0, len(subjects)+2)
	for i, s := range subjects {
		options = append(options, huh.NewOption(s, i))
	}
	options = append(options,
		huh.NewOption("↻ Regenerate", pickRegenerate),
		huh.NewOption("✎ Write my own", pickWriteOwn),
	)

	var pick int
	field := huh.NewSelect[int]().Title("Commit subject").Options(options...).Value(&pick)
	if err := runField(field); err != nil {
		return 0, err
	}
	return pick, nil
}

// inputSubject prompts for a custom subject. An empty result means the user
// wants to go back to the list.
func inputSubject(prefill string) (string, error) {
	value := prefill
	field := huh.NewInput().Title("Commit subject").Value(&value).Validate(func(s string) error {
		if strings.TrimSpace(s) == "" {
			return nil
		}
		if _, ok := cleanLine(s); !ok {
			return fmt.Errorf("subject must be 1-%d characters after cleanup", maxSubjectLen)
		}
		return nil
	})
	if err := runField(field); err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	subject, _ := cleanLine(value)
	return subject, nil
}

func runField(field huh.Field) error {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	// Filtering would make Esc ambiguous; the list is short enough without it.
	keymap.Select.Filter = key.NewBinding()
	return huh.NewForm(huh.NewGroup(field)).WithKeyMap(keymap).Run()
}
