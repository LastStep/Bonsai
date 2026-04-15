package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(guideCmd)
}

var guideCmd = &cobra.Command{
	Use:   "guide",
	Short: "View the custom files guide.",
	Long:  "Display the guide for creating custom skills, workflows, protocols, sensors, and routines.",
	RunE:  runGuide,
}

func runGuide(cmd *cobra.Command, args []string) error {
	content := guideMarkdown

	// Strip YAML frontmatter if present
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
		return fmt.Errorf("failed to create renderer: %w", err)
	}

	out, err := renderer.Render(content)
	if err != nil {
		return fmt.Errorf("failed to render guide: %w", err)
	}

	fmt.Print(out)
	return nil
}
