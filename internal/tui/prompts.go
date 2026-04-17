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

	t.Focused.Base = t.Focused.Base.BorderForeground(
		lipgloss.AdaptiveColor{Dark: "#3A5F4A", Light: "#5A7F6A"},
	)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(ColorSecondary).Bold(true)
	t.Focused.Title = t.Focused.Title.MarginBottom(0)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(ColorSecondary).Bold(true).MarginBottom(1)
	t.Focused.Description = t.Focused.Description.Foreground(ColorMuted)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(ColorDanger)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(ColorDanger)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(ColorAccent)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(ColorAccent)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(ColorAccent)
	t.Focused.Option = t.Focused.Option.Foreground(ColorSubtle)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(ColorAccent)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(ColorSuccess)
	t.Focused.SelectedPrefix = lipgloss.NewStyle().Foreground(ColorSuccess).SetString("✓ ")
	t.Focused.UnselectedPrefix = lipgloss.NewStyle().Foreground(ColorMuted).SetString("· ")
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(ColorSubtle)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(lipgloss.Color("#FFFFFF")).Background(ColorPrimary)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(ColorSubtle).Background(
		lipgloss.AdaptiveColor{Dark: "#2D2D3D", Light: "#E5E7EB"},
	)
	t.Focused.Next = t.Focused.FocusedButton

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.
		Foreground(ColorPrimary).
		Bold(true)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(ColorMuted)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(ColorAccent)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description
	t.Focused.Base = t.Focused.Base.PaddingLeft(1)

	return t
}

// ItemOption describes an item for multi-select prompts.
// Name is the human-readable display label. Value is the machine identifier
// returned as the selection result; if empty, Name is used.
type ItemOption struct {
	Name     string
	Value    string
	Desc     string
	Required bool
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

// PickItems shows a section heading and multi-select for abilities.
// Items whose names appear in defaults are pre-selected.
// Items marked Required are auto-included and shown as a locked info line.
// Returns machine identifiers (Value fields), not display names.
func PickItems(label string, available []ItemOption, defaults []string) ([]string, error) {
	if len(available) == 0 {
		return nil, nil
	}

	valueOf := func(item ItemOption) string {
		if item.Value != "" {
			return item.Value
		}
		return item.Name
	}

	// Split into required and optional
	var required []ItemOption
	var optional []ItemOption
	for _, item := range available {
		if item.Required {
			required = append(required, item)
		} else {
			optional = append(optional, item)
		}
	}

	Section(label)

	// When every item in the section is required, the user has no decision to
	// make — collapse to a single chip line so the flow doesn't read as a wall
	// of descriptions. When some are optional, show required items in full so
	// the user knows what's already in the bundle before they pick the rest.
	if len(optional) == 0 && len(required) > 0 {
		names := make([]string, len(required))
		for i, r := range required {
			names[i] = r.Name
		}
		plural := "s"
		if len(required) == 1 {
			plural = ""
		}
		head := StyleSuccess.Render(GlyphCheck) + " " +
			StyleSand.Render(fmt.Sprintf("%d item%s auto-included", len(required), plural))
		chips := StyleMuted.Render(strings.Join(names, "  "+GlyphDot+"  "))
		fmt.Println("    " + head)
		fmt.Println("    " + chips)
	} else if len(required) > 0 {
		for _, r := range required {
			line := "    " + StyleSuccess.Render(GlyphCheck) + " " + r.Name
			if r.Desc != "" {
				line += " " + StyleMuted.Render(GlyphDash+" "+r.Desc)
			}
			line += " " + StyleAccent.Render("(required)")
			fmt.Println(line)
		}
	}

	// Collect required values (machine identifiers)
	var result []string
	for _, r := range required {
		result = append(result, valueOf(r))
	}

	// If there are optional items, show multi-select
	if len(optional) > 0 {
		defaultSet := make(map[string]bool)
		for _, d := range defaults {
			defaultSet[d] = true
		}

		var options []huh.Option[string]
		for _, item := range optional {
			opt := huh.NewOption(item.Name+" "+StyleMuted.Render(GlyphDash+" "+item.Desc), valueOf(item))
			if defaultSet[valueOf(item)] {
				opt = opt.Selected(true)
			}
			options = append(options, opt)
		}

		selected, err := AskMultiSelect("", options)
		if err != nil {
			return nil, err
		}
		result = append(result, selected...)
	}

	return result, nil
}
