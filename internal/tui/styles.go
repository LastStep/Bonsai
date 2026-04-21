package tui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-isatty"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

// termWidth returns the terminal width in columns, or 80 if stdout isn't a tty
// or the size can't be determined.
func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 80
	}
	return w
}

// ─── Color Profile ────────────────────────────────────────────────────────

func init() {
	if !isatty.IsTerminal(os.Stdout.Fd()) || termenv.EnvNoColor() {
		DisableColor()
	}
}

// DisableColor forces all output to plain text (no ANSI escapes).
func DisableColor() {
	lipgloss.SetColorProfile(termenv.Ascii)
}

// ─── Zen Garden Palette ───────────────────────────────────────────────────

var (
	Leaf  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#4A9E6F", Light: "#2D7A4B"}
	Bark  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#D4A76A", Light: "#8B6914"}
	Stone lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#6B7280", Light: "#4B5563"}
	Water lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#7EC8E3", Light: "#1A7FA0"}
	Moss  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#73D677", Light: "#2D8A3E"}
	Ember lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#E36F6F", Light: "#C53030"}
	Amber lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#E3C16F", Light: "#B7791F"}
	Sand  lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#C4B7A6", Light: "#6B5E4F"}
	Petal lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#D4A0C0", Light: "#9B4D8A"}
)

// ─── Semantic Tokens ──────────────────────────────────────────────────────
//
// Semantic aliases backed by the Zen Garden palette. Prefer these in new code
// and migrate existing callsites on touch. Swap the palette value here to
// re-theme the whole TUI in one place.

var (
	ColorPrimary   = Leaf  // Brand accent — headings, primary action, banner title
	ColorSecondary = Bark  // Field labels, category headers
	ColorAccent    = Petal // Interactive chrome — cursor, selectors, next/prev
	ColorSubtle    = Sand  // Body text, option labels
	ColorMuted     = Stone // Hints, descriptions, at-rest borders
	ColorSuccess   = Moss  // Success states
	ColorDanger    = Ember // Errors
	ColorWarning   = Amber // Warnings
	ColorInfo      = Water // Info panels, review box
)

// Enso / rule chrome tokens — used by the init-flow chrome. Dimmer shades of
// Leaf/Stone for at-rest rail segments and thin dividers. Kept separate from
// the primary semantic tokens above so other commands keep their current
// palette untouched.
var (
	ColorLeafDim lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#3A7253", Light: "#3D6D53"}
	ColorRule    lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#3B4049", Light: "#D4D0CA"}
	ColorRule2   lipgloss.TerminalColor = lipgloss.AdaptiveColor{Dark: "#4A4F58", Light: "#B9B5AF"}
)

// ─── Styles ───────────────────────────────────────────────────────────────

var (
	StyleTitle   = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	StyleLabel   = lipgloss.NewStyle().Bold(true).Foreground(ColorSecondary)
	StyleMuted   = lipgloss.NewStyle().Foreground(ColorMuted)
	StyleSuccess = lipgloss.NewStyle().Foreground(ColorSuccess)
	StyleError   = lipgloss.NewStyle().Foreground(ColorDanger)
	StyleWarning = lipgloss.NewStyle().Foreground(ColorWarning)
	StyleAccent  = lipgloss.NewStyle().Foreground(ColorInfo)
	StyleSand    = lipgloss.NewStyle().Foreground(ColorSubtle)
)

// ─── Panels ───────────────────────────────────────────────────────────────

var (
	PanelSuccess = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorSuccess).
			Padding(1, 2)
	PanelError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorDanger).
			Padding(1, 2)
	PanelWarning = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorWarning).
			Padding(1, 2)
	PanelInfo = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorInfo).
			Padding(1, 2)
	PanelEmpty = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(1, 2)
)

// ─── Harness Styles ──────────────────────────────────────────────────────
//
// Used exclusively by the long-lived `internal/tui/harness` package to render
// the persistent header/footer that frames the active step inside AltScreen.
// Kept here (not in the harness package) so palette changes ripple through one
// file and so non-harness callsites can compose with the same tokens if needed.

var (
	HarnessHeader = lipgloss.NewStyle().Padding(0, 2).Foreground(ColorMuted)
	HarnessCrumb  = lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	HarnessFooter = lipgloss.NewStyle().Padding(0, 2).Foreground(ColorMuted)
)

// ─── Glyphs ───────────────────────────────────────────────────────────────

const (
	GlyphCheck = "✓"
	GlyphCross = "✗"
	GlyphWarn  = "⚠"
	GlyphArrow = "→"
	GlyphDash  = "—"
	GlyphDot   = "·"
)

// ─── Banner ───────────────────────────────────────────────────────────────

// Banner prints the Bonsai welcome banner.
// version is the build version (pass "" or "dev" to hide).
// action is an optional contextual sub-line (e.g., "Initializing new project").
// Pass "" for no action line.
func Banner(version, action string) {
	title := lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary).Render("BONSAI")
	tagline := StyleMuted.Render("agent scaffolder")
	header := title + "  " + tagline

	var lines []string
	lines = append(lines, header)

	if version != "" && version != "dev" {
		ver := StyleMuted.Render("v" + version)
		lines = append(lines, ver)
	}

	if action != "" {
		lines = append(lines, "")
		lines = append(lines, lipgloss.NewStyle().Foreground(ColorInfo).Render(action))
	}

	content := strings.Join(lines, "\n")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(1, 3).
		Render(content)

	fmt.Println("\n" + indent(box, 2))
}

// ─── Display Helpers ──────────────────────────────────────────────────────

// Blank prints an empty line.
func Blank() { fmt.Println() }

// Success prints a green checkmark message.
func Success(msg string) {
	fmt.Println("\n  " + StyleSuccess.Render(GlyphCheck) + " " + msg)
}

// Error prints a red cross message.
func Error(msg string) {
	fmt.Println("\n  " + StyleError.Render(GlyphCross) + " " + msg)
}

// Warning prints a yellow warning message.
func Warning(msg string) {
	fmt.Println("\n  " + StyleWarning.Render(GlyphWarn) + " " + msg)
}

// Hint prints dimmed helper text, aligned with message text.
func Hint(msg string) {
	fmt.Println("    " + StyleMuted.Render(msg))
}

// Answer prints a compact styled summary of an answered prompt so prior answers
// stay visible as the user advances through a multi-step flow.
// Example output:   ▸ Project name   my-project
func Answer(label, value string) {
	key := StyleLabel.Render(label)
	val := value
	if strings.TrimSpace(value) == "" {
		val = StyleMuted.Render("(skipped)")
	} else {
		val = StyleSuccess.Render(value)
	}
	fmt.Println("  " + StyleMuted.Render(GlyphArrow) + " " + key + "  " + val)
}

// Info prints dimmed informational text.
func Info(msg string) {
	fmt.Println("  " + StyleMuted.Render(msg))
}

// Heading prints a primary section heading.
func Heading(title string) {
	fmt.Println("\n  " + StyleTitle.Render("▸ "+title))
}

// Section prints a secondary section heading.
func Section(title string) {
	fmt.Println("  " + StyleLabel.Render("▸ "+title))
}

// SectionHeader prints a section title with a trailing rule line.
func SectionHeader(title string) {
	label := StyleLabel.Render(title)
	rule := StyleMuted.Render(" " + strings.Repeat("─", 34))
	fmt.Println("\n  " + label + rule)
}

// ─── Panel Functions ──────────────────────────────────────────────────────

// SuccessPanel renders a green-bordered panel.
func SuccessPanel(content, title string) {
	if title != "" {
		fmt.Println("\n  " + StyleSuccess.Bold(true).Render(title))
	}
	fmt.Println(indent(PanelSuccess.Render(content), 2))
}

// ErrorPanel renders a red-bordered panel.
func ErrorPanel(msg string) {
	fmt.Println("\n" + indent(PanelError.Render(msg), 2))
}

// FatalPanel renders a structured error and exits. title: what happened. detail: why. hint: how to fix.
func FatalPanel(title, detail, hint string) {
	content := StyleError.Bold(true).Render(title)
	if detail != "" {
		content += "\n" + detail
	}
	if hint != "" {
		content += "\n" + StyleMuted.Render(hint)
	}
	fmt.Println("\n" + indent(PanelError.Render(content), 2))
	os.Exit(1)
}

// ErrorDetail renders a structured non-fatal error. Same shape as FatalPanel but does not exit.
// title: what happened (bold, error color). detail: why. hint: how to fix (muted).
func ErrorDetail(title, detail, hint string) {
	content := StyleError.Bold(true).Render(title)
	if detail != "" {
		content += "\n" + detail
	}
	if hint != "" {
		content += "\n" + StyleMuted.Render(hint)
	}
	fmt.Println("\n" + indent(PanelError.Render(content), 2))
}

// WarningPanel renders a yellow-bordered panel.
func WarningPanel(msg string) {
	fmt.Println("\n" + indent(PanelWarning.Render(msg), 2))
}

// InfoPanel renders a blue-bordered panel.
func InfoPanel(content, title string) {
	if title != "" {
		fmt.Println("\n  " + StyleAccent.Bold(true).Render(title))
	}
	fmt.Println(indent(PanelInfo.Render(content), 2))
}

// EmptyPanel renders a dim panel for empty states.
func EmptyPanel(msg string) {
	fmt.Println("\n" + indent(PanelEmpty.Render(msg), 2))
}

// TitledPanelString renders the same bordered panel as TitledPanel and returns
// the result as a string (with a leading newline, matching the side-effecting
// TitledPanel's fmt.Println output). Use inside AltScreen or when composing
// with other styled content.
func TitledPanelString(title, content string, color lipgloss.TerminalColor) string {
	bdr := lipgloss.NewStyle().Foreground(color)
	ttl := lipgloss.NewStyle().Foreground(color).Bold(true)

	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")

	// Panel layout: "  │  <line><pad>  │"
	// The leading indent is 2, each border is 1 wide, and the line gets 2 columns
	// of left padding inside the box. So the box's outer width is innerW + 2
	// (two border chars), and we reserve 2 columns at the front of the window.
	const leadingIndent = 2
	const borderCols = 2
	const leftPad = 2

	// Compute natural inner width from content
	maxW := 0
	for _, l := range contentLines {
		if w := lipgloss.Width(l); w > maxW {
			maxW = w
		}
	}
	innerW := maxW + 4
	titleVisualW := lipgloss.Width(title) + 2 // spaces around title
	if minW := titleVisualW + 4; innerW < minW {
		innerW = minW
	}

	// Cap against terminal width so right border doesn't fall off-screen
	maxInner := termWidth() - leadingIndent - borderCols
	if maxInner < titleVisualW+4 {
		maxInner = titleVisualW + 4
	}
	if innerW > maxInner {
		innerW = maxInner
	}

	// Max visible width for content text (after the left padding)
	lineBudget := innerW - leftPad

	// Top border: ╭─ Title ──...──╮
	dashesAfter := innerW - 1 - titleVisualW
	if dashesAfter < 1 {
		dashesAfter = 1
	}
	top := bdr.Render("╭─") + ttl.Render(" "+title+" ") + bdr.Render(strings.Repeat("─", dashesAfter)+"╮")
	emptyLine := bdr.Render("│") + strings.Repeat(" ", innerW) + bdr.Render("│")
	bottom := bdr.Render("╰" + strings.Repeat("─", innerW) + "╯")

	var buf strings.Builder
	buf.WriteString("  " + top + "\n")
	buf.WriteString("  " + emptyLine + "\n")
	for _, line := range contentLines {
		if lipgloss.Width(line) > lineBudget {
			line = ansi.Truncate(line, lineBudget, "…")
		}
		w := lipgloss.Width(line)
		pad := lineBudget - w
		if pad < 0 {
			pad = 0
		}
		buf.WriteString("  " + bdr.Render("│") + "  " + line + strings.Repeat(" ", pad) + bdr.Render("│") + "\n")
	}
	buf.WriteString("  " + emptyLine + "\n")
	buf.WriteString("  " + bottom)

	return "\n" + buf.String()
}

// TitledPanel renders a bordered panel with the title embedded in the top border.
// The panel is capped to the terminal width so borders don't break when content
// is wider than the visible columns; over-long lines are truncated with an ellipsis.
func TitledPanel(title, content string, color lipgloss.TerminalColor) {
	fmt.Println(TitledPanelString(title, content, color))
}

// ─── Data Display ─────────────────────────────────────────────────────────

// Fields renders aligned key:value pairs to stdout.
func Fields(pairs [][2]string) {
	maxLen := 0
	for _, p := range pairs {
		if len(p[0]) > maxLen {
			maxLen = len(p[0])
		}
	}
	for _, p := range pairs {
		key := StyleLabel.Render(p[0])
		padding := strings.Repeat(" ", maxLen-len(p[0])+1)
		fmt.Printf("     %s%s%s\n", key, padding, p[1])
	}
}

// CardFields builds aligned key:value field lines as a string for embedding in panels.
func CardFields(pairs [][2]string) string {
	if len(pairs) == 0 {
		return ""
	}
	var buf strings.Builder
	maxLen := 0
	for _, p := range pairs {
		if len(p[0]) > maxLen {
			maxLen = len(p[0])
		}
	}
	for i, p := range pairs {
		key := StyleLabel.Render(p[0])
		padding := strings.Repeat(" ", maxLen-len(p[0])+3)
		buf.WriteString(key + padding + p[1])
		if i < len(pairs)-1 {
			buf.WriteString("\n")
		}
	}
	return buf.String()
}

// ─── Trees ────────────────────────────────────────────────────────────────

// Category is a named group of items for tree rendering.
type Category struct {
	Name  string
	Items []string
}

// ItemTree renders a tree with category branches and item leaves.
func ItemTree(root string, categories []Category, describe func(string) string) string {
	var buf strings.Builder
	buf.WriteString("  " + root + "\n")

	var nonEmpty []Category
	for _, c := range categories {
		if len(c.Items) > 0 {
			nonEmpty = append(nonEmpty, c)
		}
	}

	bc := lipgloss.NewStyle().Foreground(ColorMuted)

	for i, cat := range nonEmpty {
		isLast := i == len(nonEmpty)-1
		branch := "├── "
		if isLast {
			branch = "└── "
		}
		header := StyleLabel.Render(cat.Name) + " " + StyleMuted.Render(fmt.Sprintf("(%d)", len(cat.Items)))
		buf.WriteString("  " + bc.Render(branch) + header + "\n")

		childPrefix := "│   "
		if isLast {
			childPrefix = "    "
		}
		for j, item := range cat.Items {
			isLastItem := j == len(cat.Items)-1
			itemBranch := "├── "
			if isLastItem {
				itemBranch = "└── "
			}
			line := item
			if describe != nil {
				if desc := describe(item); desc != "" {
					line += " " + StyleMuted.Render(GlyphDash+" "+desc)
				}
			}
			buf.WriteString("  " + bc.Render(childPrefix+itemBranch) + line + "\n")
		}
	}

	return buf.String()
}

// FileTree builds a tree display from flat file paths.
func FileTree(files []string, rootLabel string) string {
	type node struct {
		name     string
		children map[string]*node
		order    []string
		isFile   bool
	}

	root := &node{name: rootLabel, children: make(map[string]*node)}
	for _, f := range files {
		parts := strings.Split(f, "/")
		cur := root
		for i, part := range parts {
			if _, exists := cur.children[part]; !exists {
				cur.children[part] = &node{
					name:     part,
					children: make(map[string]*node),
					isFile:   i == len(parts)-1,
				}
				cur.order = append(cur.order, part)
			}
			cur = cur.children[part]
		}
	}

	bc := lipgloss.NewStyle().Foreground(ColorMuted)
	var buf strings.Builder
	buf.WriteString("  " + StyleLabel.Render(rootLabel) + "\n")

	var render func(n *node, prefix string)
	render = func(n *node, prefix string) {
		for i, name := range n.order {
			child := n.children[name]
			isLast := i == len(n.order)-1
			branch := "├── "
			if isLast {
				branch = "└── "
			}
			display := name
			style := StyleSand
			if !child.isFile && len(child.children) > 0 {
				style = StyleLabel
				display += "/"
			}
			buf.WriteString("  " + bc.Render(prefix+branch) + style.Render(display) + "\n")

			nextPrefix := prefix + "│   "
			if isLast {
				nextPrefix = prefix + "    "
			}
			render(child, nextPrefix)
		}
	}
	render(root, "")
	return buf.String()
}

// CatalogTable renders a styled table for abilities.
func CatalogTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		Info("  (none)")
		return
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(ColorMuted)).
		Headers(headers...).
		Rows(rows...).
		BorderRow(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
			if row == table.HeaderRow {
				return s.Foreground(ColorSecondary).Bold(true)
			}
			if col == 0 {
				return s.Foreground(ColorPrimary)
			}
			return s.Foreground(ColorMuted)
		})

	fmt.Println(indent(t.Render(), 2))
}

// ─── Internal Helpers ─────────────────────────────────────────────────────

func indent(s string, n int) string {
	prefix := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
