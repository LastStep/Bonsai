package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// BonsaiTheme returns a custom Huh form theme matching the Zen Garden palette.
func BonsaiTheme() *huh.Theme {
	t := huh.ThemeBase()

	t.Focused.Base = t.Focused.Base.BorderForeground(lipgloss.Color("#3A5F4A"))
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(Bark).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(Bark).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(Stone)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(Ember)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(Ember)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(Petal)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(Petal)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(Petal)
	t.Focused.Option = t.Focused.Option.Foreground(Sand)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(Petal)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(Moss)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(Moss).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(Stone).SetString("· ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(Sand)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("#FFFFFF")).Background(Leaf)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(Sand).Background(lipgloss.Color("#2D2D3D"))
	t.Focused.Next = t.Focused.FocusedButton

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(Leaf)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(Stone)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(Petal)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}

// ItemOption describes an item for multi-select prompts.
type ItemOption struct {
	Name string
	Desc string
}

// AskText prompts for text input.
func AskText(title, defaultVal string, required bool) (string, error) {
	var value string
	input := huh.NewInput().
		Title(title).
		Value(&value)

	if defaultVal != "" {
		input.Placeholder(defaultVal)
	}

	if required {
		input.Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return fmt.Errorf("required")
			}
			return nil
		})
	}

	err := huh.NewForm(huh.NewGroup(input)).WithTheme(BonsaiTheme()).Run()
	if err != nil {
		return "", err
	}
	if value == "" && defaultVal != "" {
		value = defaultVal
	}
	return value, nil
}

// AskSelect prompts for a single selection.
func AskSelect(title string, options []huh.Option[string]) (string, error) {
	var value string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(title).
				Options(options...).
				Value(&value),
		),
	).WithTheme(BonsaiTheme()).Run()
	return value, err
}

// AskMultiSelect prompts for multiple selections.
func AskMultiSelect(title string, options []huh.Option[string]) ([]string, error) {
	var values []string
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title(title).
				Options(options...).
				Value(&values),
		),
	).WithTheme(BonsaiTheme()).Run()
	return values, err
}

// AskConfirm prompts for yes/no confirmation.
func AskConfirm(title string, defaultVal bool) (bool, error) {
	value := defaultVal
	err := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Affirmative("Yes").
				Negative("No").
				Value(&value),
		),
	).WithTheme(BonsaiTheme()).Run()
	return value, err
}

// PickItems shows a section heading and multi-select for catalog items.
// Items whose names appear in defaults are pre-selected.
func PickItems(label string, available []ItemOption, defaults []string) ([]string, error) {
	if len(available) == 0 {
		return nil, nil
	}

	defaultSet := make(map[string]bool)
	for _, d := range defaults {
		defaultSet[d] = true
	}

	Section(label)

	var options []huh.Option[string]
	for _, item := range available {
		opt := huh.NewOption(item.Name+" "+StyleMuted.Render(GlyphDash+" "+item.Desc), item.Name)
		if defaultSet[item.Name] {
			opt = opt.Selected(true)
		}
		options = append(options, opt)
	}

	return AskMultiSelect("", options)
}
