package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/tui"
)

// guideTopic pairs a machine key (looked up in guideContents) with a
// human-readable label shown in the interactive picker.
type guideTopic struct {
	Key   string
	Label string
}

// guideTopics is the ordered list of topics shown in the picker. Order here
// drives picker order — keep it deterministic.
var guideTopics = []guideTopic{
	{"quickstart", "Quickstart — 5-step post-install walkthrough"},
	{"concepts", "Concepts — the mental model"},
	{"cli", "CLI — command-by-command reference"},
	{"custom-files", "Custom Files — add your own abilities"},
}

func init() {
	rootCmd.AddCommand(guideCmd)
}

var guideCmd = &cobra.Command{
	Use:   "guide [topic]",
	Short: "View bundled guides in the terminal.",
	Long:  "Render one of the bundled guides as styled terminal output. Run without a topic to pick interactively, or pass one of: quickstart, concepts, cli, custom-files.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runGuide,
}

func runGuide(cmd *cobra.Command, args []string) error {
	var key string
	if len(args) == 1 {
		key = args[0]
		if _, ok := guideContents[key]; !ok {
			return fmt.Errorf("unknown topic %q. Available: quickstart, concepts, cli, custom-files", key)
		}
	} else {
		options := make([]huh.Option[string], 0, len(guideTopics))
		for _, t := range guideTopics {
			options = append(options, huh.NewOption(t.Label, t.Key))
		}
		selected, err := tui.AskSelect("Pick a guide", options)
		if err != nil {
			return err
		}
		key = selected
	}

	content, ok := guideContents[key]
	if !ok {
		return fmt.Errorf("unknown topic %q. Available: quickstart, concepts, cli, custom-files", key)
	}

	out, err := renderMarkdown(content)
	if err != nil {
		return err
	}

	fmt.Print(out)
	return nil
}

// renderMarkdown strips YAML frontmatter (if present) and renders the remaining
// markdown via glamour with the auto-selected terminal style.
func renderMarkdown(content string) (string, error) {
	if strings.HasPrefix(content, "---") {
		if idx := strings.Index(content[3:], "---"); idx >= 0 {
			content = strings.TrimSpace(content[idx+6:])
		}
	}

	renderer, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(100),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create renderer: %w", err)
	}

	out, err := renderer.Render(content)
	if err != nil {
		return "", fmt.Errorf("failed to render guide: %w", err)
	}

	return out, nil
}
