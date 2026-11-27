package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
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
		var prefill string
		if !writeOwn {
			pick, edit, err := selectSubject(subjects)
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
				prefill = edit
			default:
				return subjects[pick], nil
			}
		}
		writeOwn = false

		subject, err := inputSubject(prefill)
		if err != nil {
			return "", err
		}
		if subject != "" {
			return subject, nil
		}
	}
}

// selectSubject shows the subject list. Pressing e returns pickWriteOwn with the
// highlighted suggestion (empty for the other rows) to edit; r regenerates.
func selectSubject(subjects []string) (int, string, error) {
	options := make([]huh.Option[int], 0, len(subjects)+2)
	for i, s := range subjects {
		options = append(options, huh.NewOption(s, i))
	}
	options = append(options,
		huh.NewOption("↻ Regenerate", pickRegenerate),
		huh.NewOption("✎ Write my own", pickWriteOwn),
	)

	var pick int
	var edit string
	field := huh.NewSelect[int]().
		Title("Commit subject").
		Description("e edit · r regenerate").
		Options(options...).
		Value(&pick)
	filter := func(_ tea.Model, msg tea.Msg) tea.Msg {
		k, ok := msg.(tea.KeyPressMsg)
		if !ok {
			return msg
		}
		switch k.String() {
		case "e":
			pick = pickWriteOwn
			if i, ok := field.Hovered(); ok && i >= 0 {
				edit = subjects[i]
			}
		case "r":
			pick = pickRegenerate
		default:
			return msg
		}
		return tea.QuitMsg{}
	}
	if err := runField(field, tea.WithFilter(filter)); err != nil {
		return 0, "", err
	}
	return pick, edit, nil
}

// inputSubject prompts for a custom subject, starting from initial. An empty
// result means the user wants to go back to the list.
func inputSubject(initial string) (string, error) {
	value := initial
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

func runField(field huh.Field, opts ...tea.ProgramOption) error {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	// Filtering would make Esc ambiguous; the list is short enough without it.
	keymap.Select.Filter = key.NewBinding()
	return huh.NewForm(huh.NewGroup(field)).WithKeyMap(keymap).WithProgramOptions(opts...).Run()
}
