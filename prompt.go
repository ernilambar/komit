package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
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
	subjects, err := generateWithSpinner(cfg, input, n, nil)
	writeOwn := errors.Is(err, errNoSubjects)
	if err != nil && !writeOwn {
		return "", err
	}
	// rejected collects every list the user regenerated away from.
	var rejected []string
	var notice string
	if writeOwn {
		notice = capitalize(err.Error())
	}

	for {
		var prefill string
		if !writeOwn {
			pick, edit, err := selectSubject(subjects, notice)
			notice = ""
			if err != nil {
				return "", err
			}
			switch pick {
			case pickRegenerate:
				// A failed regenerate keeps the current list and says why.
				avoid := append(rejected, subjects...)
				fresh, err := generateWithSpinner(cfg, input, n, avoid)
				switch {
				case errors.Is(err, huh.ErrUserAborted):
					return "", err
				case err != nil:
					notice = "Regenerate failed: " + err.Error()
				default:
					rejected = avoid
					subjects = fresh
				}
				continue
			case pickWriteOwn:
				prefill = edit
			default:
				return subjects[pick], nil
			}
		}
		writeOwn = false

		subject, err := inputSubject(prefill, notice)
		notice = ""
		if err != nil {
			return "", err
		}
		if subject != "" {
			return subject, nil
		}
	}
}

// generateWithSpinner runs suggestSubjects behind a spinner. Ctrl+C returns
// huh.ErrUserAborted.
func generateWithSpinner(cfg config, input string, n int, rejected []string) ([]string, error) {
	var subjects []string
	err := spinner.New().
		Title(" Generating subjects…").
		ActionWithErr(func(context.Context) error {
			var err error
			subjects, err = suggestSubjects(cfg, input, n, rejected)
			return err
		}).
		Run()
	if errors.Is(err, tea.ErrInterrupted) {
		return nil, huh.ErrUserAborted
	}
	return subjects, err
}

// selectSubject shows the subject list. Pressing e returns pickWriteOwn with the
// highlighted suggestion (empty for the other rows) to edit; r regenerates.
func selectSubject(subjects []string, notice string) (int, string, error) {
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
		Description(withNotice(notice, "e edit · r regenerate")).
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
func inputSubject(initial, notice string) (string, error) {
	value := initial
	field := huh.NewInput().
		Title("Commit subject").
		Description(withNotice(notice, "Submit empty to go back")).
		Value(&value).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return nil
			}
			_, err := cleanInput(s)
			return err
		})
	if err := runField(field); err != nil {
		return "", err
	}
	if strings.TrimSpace(value) == "" {
		return "", nil
	}
	subject, _ := cleanInput(value)
	return subject, nil
}

// withNotice puts a one-off notice above the usual hint.
func withNotice(notice, hint string) string {
	if notice == "" {
		return hint
	}
	return notice + "\n" + hint
}

func runField(field huh.Field, opts ...tea.ProgramOption) error {
	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = key.NewBinding(key.WithKeys("ctrl+c", "esc"))
	// Filtering would make Esc ambiguous; the list is short enough without it.
	keymap.Select.Filter = key.NewBinding()
	return huh.NewForm(huh.NewGroup(field)).WithKeyMap(keymap).WithProgramOptions(opts...).Run()
}
