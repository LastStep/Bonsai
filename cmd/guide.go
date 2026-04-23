package cmd

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"

	"github.com/LastStep/Bonsai/internal/tui/guideflow"
)

// guideTopic pairs a machine key (looked up in guideContents) with a
// human-readable label. The label is informational only — the cinematic
// viewer renders its own tab strip via guideflow.NewTopics.
type guideTopic struct {
	Key   string
	Label string
}

// guideTopics is the canonical topic order used for validation error
// messages and the fallback topic list. The cinematic viewer uses
// guideflow.NewTopics' own canonicalOrder constant which mirrors
// this slice; keep the two in sync when adding a topic.
var guideTopics = []guideTopic{
	{"quickstart", "Quickstart — 5-step post-install walkthrough"},
	{"concepts", "Concepts — the mental model"},
	{"cli", "CLI — command-by-command reference"},
	{"custom-files", "Custom Files — add your own abilities"},
}

// noTTYNoArgErr is the exact error message surfaced when guide is
// invoked with neither an arg nor a TTY (e.g. piped through less
// without specifying a topic). Decision D4 in Plan 28's 2026-04-23
// deltas locks the wording; test asserts verbatim.
const noTTYNoArgErr = "bonsai guide: specify a topic when piping output (quickstart, concepts, cli, custom-files)"

func init() {
	rootCmd.AddCommand(guideCmd)
}

var guideCmd = &cobra.Command{
	Use:   "guide [topic]",
	Short: "View bundled guides in the terminal.",
	Long: "Render one of the bundled guides as styled terminal output. Run without a " +
		"topic to open the cinematic viewer on the first topic; pass one of: " +
		"quickstart, concepts, cli, custom-files.",
	Args: cobra.MaximumNArgs(1),
	RunE: runGuide,
}

// runGuide is the guide command entry point. On a TTY it launches
// the cinematic guideflow viewer (tabbed scroll viewport over the
// four bundled docs). Off a TTY it either renders the specified
// topic as static glamour output (preserves the pre-Plan-28
// behavior for piped consumption) or errors out when no topic was
// passed (decision D4 — piping needs an explicit pick so the
// downstream tool sees a stable single doc).
func runGuide(cmd *cobra.Command, args []string) error {
	var key string
	if len(args) == 1 {
		key = args[0]
		if _, ok := guideContents[key]; !ok {
			return fmt.Errorf("unknown topic %q. Available: quickstart, concepts, cli, custom-files", key)
		}
	}

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		if key == "" {
			return fmt.Errorf("%s", noTTYNoArgErr)
		}
		return renderStatic(guideContents[key])
	}

	topics := guideflow.NewTopics(guideContents)
	if len(topics) == 0 {
		return fmt.Errorf("no guide topics available")
	}

	cwd, _ := os.Getwd()
	stage := guideflow.NewViewer(topics, key, Version, cwd)
	if _, err := tea.NewProgram(stage, tea.WithAltScreen()).Run(); err != nil {
		return fmt.Errorf("guide viewer: %w", err)
	}
	return nil
}

// renderStatic renders the given markdown content through glamour
// with auto-style and a fixed word-wrap of 100 columns — the
// pre-Plan-28 behavior preserved verbatim for piped invocations.
// The cinematic TTY path binds glamour's width to the live
// terminal instead (see guideflow/render.go).
func renderStatic(content string) error {
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
