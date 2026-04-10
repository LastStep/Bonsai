package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// ─── Zen Garden Palette ───────────────────────────────────────────────────

var (
	Leaf  = lipgloss.Color("#4A9E6F")
	Bark  = lipgloss.Color("#D4A76A")
	Stone = lipgloss.Color("#6B7280")
	Water = lipgloss.Color("#7EC8E3")
	Moss  = lipgloss.Color("#73D677")
	Ember = lipgloss.Color("#E36F6F")
	Amber = lipgloss.Color("#E3C16F")
	Sand  = lipgloss.Color("#C4B7A6")
	Petal = lipgloss.Color("#D4A0C0")
)

// ─── Styles ───────────────────────────────────────────────────────────────

var (
	StyleTitle   = lipgloss.NewStyle().Bold(true).Foreground(Leaf)
	StyleLabel   = lipgloss.NewStyle().Bold(true).Foreground(Bark)
	StyleMuted   = lipgloss.NewStyle().Foreground(Stone)
	StyleSuccess = lipgloss.NewStyle().Foreground(Moss)
	StyleError   = lipgloss.NewStyle().Foreground(Ember)
	StyleWarning = lipgloss.NewStyle().Foreground(Amber)
	StyleAccent  = lipgloss.NewStyle().Foreground(Water)
	StyleSand    = lipgloss.NewStyle().Foreground(Sand)
)

// ─── Panels ───────────────────────────────────────────────────────────────

var (
	PanelSuccess = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Moss).
			Padding(1, 2)
	PanelError = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Ember).
			Padding(1, 2)
	PanelWarning = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Amber).
			Padding(1, 2)
	PanelInfo = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Water).
			Padding(1, 2)
	PanelEmpty = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Stone).
			Padding(1, 2)
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
func Banner() {
	title := lipgloss.NewStyle().Bold(true).Foreground(Leaf).Render("B O N S A I")
	sub := StyleMuted.Render("agent scaffolder")

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(Leaf).
		Padding(1, 5).
		Align(lipgloss.Center).
		Render(title + "\n" + sub)

	fmt.Println("\n" + indent(box, 3))
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

// TitledPanel renders a bordered panel with the title embedded in the top border.
func TitledPanel(title, content string, color lipgloss.TerminalColor) {
	bdr := lipgloss.NewStyle().Foreground(color)
	ttl := lipgloss.NewStyle().Foreground(color).Bold(true)

	contentLines := strings.Split(strings.TrimRight(content, "\n"), "\n")

	// Calculate inner width from content
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
		w := lipgloss.Width(line)
		pad := innerW - 2 - w
		if pad < 0 {
			pad = 0
		}
		buf.WriteString("  " + bdr.Render("│") + "  " + line + strings.Repeat(" ", pad) + bdr.Render("│") + "\n")
	}
	buf.WriteString("  " + emptyLine + "\n")
	buf.WriteString("  " + bottom)

	fmt.Println("\n" + buf.String())
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

	bc := lipgloss.NewStyle().Foreground(Stone)

	for i, cat := range nonEmpty {
		isLast := i == len(nonEmpty)-1
		branch := "├── "
		if isLast {
			branch = "└── "
		}
		buf.WriteString("  " + bc.Render(branch) + StyleLabel.Render(cat.Name) + "\n")

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

	bc := lipgloss.NewStyle().Foreground(Stone)
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

// CatalogTable renders a styled table for catalog items.
func CatalogTable(headers []string, rows [][]string) {
	if len(rows) == 0 {
		Info("  (none)")
		return
	}

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(Stone)).
		Headers(headers...).
		Rows(rows...).
		BorderRow(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			s := lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
			if row == table.HeaderRow {
				return s.Foreground(Bark).Bold(true)
			}
			if col == 0 {
				return s.Foreground(Leaf)
			}
			return s.Foreground(Stone)
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
