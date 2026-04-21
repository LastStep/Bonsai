package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// RenderFileTree draws a rich file-tree widget used by the redesigned
// `bonsai init` cinematic flow (Planted + Observe stages). It is separate
// from the flat FileTree helper at styles.go:462 — that helper is still used
// by `add`/`remove`/`update` panels and must remain untouched.
//
// Rendering rules (locked by Plan 22, Phase 1):
//   - Branch glyphs: "├─ " for intermediate siblings, "└─ " for the last.
//   - Continuation prefixes: "│  " for non-last ancestors, "   " (three spaces)
//     for last ancestors.
//   - Name color:
//     NodeDir  + NodeCurrent  → Leaf bold
//     NodeDir  + *            → Bark bold
//     NodeFile + NodeNew      → Leaf
//     NodeFile + NodeRequired → Subtle bold
//     else                    → Subtle
//   - Status badges sit flush-right with 1 column of padding from the note
//     (or name). Each badge reserves 9 visible columns including its trailing
//     space — "NEW      " (Leaf) and "REQUIRED " (Bark).
//   - Note text is muted, separated from the name by two spaces. When the
//     full line would exceed opts.MaxWidth, the note is truncated with "…"
//     using the same budget that reserves room for the trailing badge.
//   - NodeCurrent additionally receives a 2-col left border "│ " in Leaf, and
//     the full line (border + tree content) is wrapped in a leaf-tinted
//     background.
//   - Dense mode removes blank lines between siblings at every depth. Default
//     (non-dense) mode inserts one blank line between each sibling.
//   - When opts.Root is non-nil, a root label line renders above the children
//     with the same name/note/badge conventions and the children hang below
//     it using "│  " continuation.

// NodeKind discriminates directories from files for rendering purposes.
type NodeKind int

const (
	// NodeFile is a leaf (document).
	NodeFile NodeKind = iota
	// NodeDir is an internal directory node whose Children may be nested.
	NodeDir
)

// NodeStatus styles a node and optionally attaches a right-aligned badge.
type NodeStatus int

const (
	// NodeNormal renders the node with the default foreground colors; no badge.
	NodeNormal NodeStatus = iota
	// NodeNew marks newly-written files: Leaf-colored name + "NEW" badge.
	NodeNew
	// NodeRequired marks required files: bold Subtle name + "REQUIRED" badge.
	NodeRequired
	// NodeCurrent tints a whole subtree with a leaf border + background.
	NodeCurrent
)

// TreeNode is one entry in the rendered tree. Kind=NodeFile entries ignore
// Children. Status controls coloring and badge. Note is a one-line caption
// rendered after the name in muted text.
type TreeNode struct {
	Name     string
	Kind     NodeKind
	Status   NodeStatus
	Note     string
	Children []TreeNode
}

// FileTreeOpts tunes the renderer.
type FileTreeOpts struct {
	// Root, when non-nil, renders a prefix line with the root's name and note
	// above the main tree. The root itself is not counted as a sibling of the
	// top-level nodes; its children (if any) are ignored.
	Root *TreeNode
	// Dense collapses sibling spacing to a single line each (no blank line
	// between siblings at any depth).
	Dense bool
	// MaxWidth caps the rendered width of each row. 0 means auto-detect via
	// the terminal width (falling back to 80).
	MaxWidth int
}

// Badge column width — the label plus one trailing space, rendered to the
// right of the line so all badges of the same kind stack. Kept as a constant
// so tests can check alignment.
const (
	badgeNew      = "NEW"
	badgeRequired = "REQUIRED"
	badgeColWidth = 9 // longest badge ("REQUIRED") + one trailing space
)

// RenderFileTree renders nodes (and the optional Root) into a string safe to
// embed inside AltScreen views.
func RenderFileTree(nodes []TreeNode, opts FileTreeOpts) string {
	width := opts.MaxWidth
	if width <= 0 {
		width = termWidth()
	}

	var buf strings.Builder

	// Styles used across nodes. Foreground-only so they compose cleanly with
	// the NodeCurrent background wrap below.
	nameStyles := nameStyleMap()
	noteStyle := lipgloss.NewStyle().Foreground(ColorMuted)
	badgeNewStyle := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	badgeReqStyle := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)

	// Root label (if present).
	if opts.Root != nil {
		line := renderNodeLine(*opts.Root, "", "", width, nameStyles, noteStyle, badgeNewStyle, badgeReqStyle)
		buf.WriteString(line)
		buf.WriteString("\n")
	}

	// Children indent under root when a root is provided.
	childPrefix := ""
	if opts.Root != nil {
		childPrefix = "│  "
	}

	for i, n := range nodes {
		last := i == len(nodes)-1
		renderSubtree(&buf, n, childPrefix, last, opts, width, nameStyles, noteStyle, badgeNewStyle, badgeReqStyle)
		// Insert blank-line separator between siblings at the top level in
		// non-dense mode. The separator is emitted after every sibling
		// except the final one so trailing whitespace stays bounded.
		if !opts.Dense && !last {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// renderSubtree writes one TreeNode (plus its descendants) into buf. prefix
// is the accumulated continuation string drawn before the branch glyph for
// this node. last controls whether this node uses "└─ " / "   " instead of
// "├─ " / "│  " for its descendants.
func renderSubtree(
	buf *strings.Builder,
	n TreeNode,
	prefix string,
	last bool,
	opts FileTreeOpts,
	width int,
	nameStyles map[nameKey]lipgloss.Style,
	noteStyle, badgeNewStyle, badgeReqStyle lipgloss.Style,
) {
	branch := "├─ "
	if last {
		branch = "└─ "
	}

	line := renderNodeLine(n, prefix, branch, width, nameStyles, noteStyle, badgeNewStyle, badgeReqStyle)
	buf.WriteString(line)
	buf.WriteString("\n")

	if n.Kind != NodeDir || len(n.Children) == 0 {
		return
	}

	nextPrefix := prefix + "│  "
	if last {
		nextPrefix = prefix + "   "
	}
	for i, c := range n.Children {
		childLast := i == len(n.Children)-1
		renderSubtree(buf, c, nextPrefix, childLast, opts, width, nameStyles, noteStyle, badgeNewStyle, badgeReqStyle)
		if !opts.Dense && !childLast {
			buf.WriteString("\n")
		}
	}
}

// renderNodeLine builds a single row: [current-border] prefix + branch + name
// + note + right-aligned badge, optionally wrapped in a leaf-tint background
// when NodeCurrent is set.
func renderNodeLine(
	n TreeNode,
	prefix, branch string,
	width int,
	nameStyles map[nameKey]lipgloss.Style,
	noteStyle, badgeNewStyle, badgeReqStyle lipgloss.Style,
) string {
	// Stable plain-text tree scaffold — used both as the rendered leader and
	// as the basis for width math (styles do not affect cell width).
	leader := prefix + branch

	name := n.Name
	if n.Kind == NodeDir && !strings.HasSuffix(name, "/") {
		name += "/"
	}
	styledName := nameStyles[nameKey{kind: n.Kind, status: n.Status}].Render(name)

	// Badge.
	var badgeRaw, badgeStyled string
	switch n.Status {
	case NodeNew:
		badgeRaw = badgeNew
		badgeStyled = badgeNewStyle.Render(badgeNew)
	case NodeRequired:
		badgeRaw = badgeRequired
		badgeStyled = badgeReqStyle.Render(badgeRequired)
	}

	// Compute budgets in visible columns.
	leftCols := lipgloss.Width(leader) + lipgloss.Width(name)
	noteGap := 0
	note := n.Note
	if note != "" {
		noteGap = 2 // two-space gap before the note
	}

	badgeCols := 0
	if badgeRaw != "" {
		badgeCols = badgeColWidth // badge + trailing space
	}

	// NodeCurrent consumes two extra visible columns for the "│ " left border.
	currentCols := 0
	if n.Status == NodeCurrent {
		currentCols = 2
	}

	// Decide how much of the note we can show without overflowing the row.
	// Budget = width - current-border - left-content - note-gap - badge-col.
	noteBudget := width - currentCols - leftCols - noteGap - badgeCols
	if noteBudget < 0 {
		noteBudget = 0
	}
	if note != "" && lipgloss.Width(note) > noteBudget {
		note = ansi.Truncate(note, noteBudget, "…")
	}
	styledNote := ""
	if note != "" {
		styledNote = strings.Repeat(" ", noteGap) + noteStyle.Render(note)
	}

	// Compose the visible line (without the optional current-border wrap).
	row := leader + styledName + styledNote

	if badgeRaw != "" {
		// Badge block = label + trailing spaces = exactly badgeColWidth cols.
		trailing := badgeColWidth - lipgloss.Width(badgeRaw)
		if trailing < 1 {
			trailing = 1
		}
		badgeBlock := badgeStyled + strings.Repeat(" ", trailing)

		// Right-align the block so its trailing space sits at column
		// (width - currentCols). Pad is whatever space remains between the
		// row content and the badge block; always keep at least 1 column of
		// visual breathing room.
		rowVis := lipgloss.Width(row)
		pad := width - currentCols - badgeColWidth - rowVis
		if pad < 1 {
			pad = 1
		}
		row += strings.Repeat(" ", pad) + badgeBlock
	}

	if n.Status == NodeCurrent {
		// Prepend leaf-colored left border. The trailing row already carries
		// no background; wrap the whole line in a Leaf-tint background so the
		// highlight spans the width.
		borderStyle := lipgloss.NewStyle().Foreground(ColorPrimary)
		border := borderStyle.Render("│ ")
		// Pad trailing spaces so the background stretches to width.
		currentBg := lipgloss.NewStyle().Background(ColorLeafDim)
		rowVis := lipgloss.Width(row)
		pad := width - currentCols - rowVis
		if pad < 0 {
			pad = 0
		}
		highlight := currentBg.Render(row + strings.Repeat(" ", pad))
		return border + highlight
	}

	return row
}

// nameKey indexes the precomputed name styles.
type nameKey struct {
	kind   NodeKind
	status NodeStatus
}

// nameStyleMap returns the lookup used by renderNodeLine for name coloring.
// Built on each Render call so that DisableColor() and palette swaps take
// effect without package-init ordering surprises.
func nameStyleMap() map[nameKey]lipgloss.Style {
	leafBold := lipgloss.NewStyle().Foreground(ColorPrimary).Bold(true)
	barkBold := lipgloss.NewStyle().Foreground(ColorSecondary).Bold(true)
	leaf := lipgloss.NewStyle().Foreground(ColorPrimary)
	subtleBold := lipgloss.NewStyle().Foreground(ColorSubtle).Bold(true)
	subtle := lipgloss.NewStyle().Foreground(ColorSubtle)

	m := map[nameKey]lipgloss.Style{}
	// Files
	for _, st := range []NodeStatus{NodeNormal, NodeNew, NodeRequired, NodeCurrent} {
		switch st {
		case NodeNew:
			m[nameKey{NodeFile, st}] = leaf
		case NodeRequired:
			m[nameKey{NodeFile, st}] = subtleBold
		default:
			m[nameKey{NodeFile, st}] = subtle
		}
	}
	// Directories
	for _, st := range []NodeStatus{NodeNormal, NodeNew, NodeRequired, NodeCurrent} {
		if st == NodeCurrent {
			m[nameKey{NodeDir, st}] = leafBold
		} else {
			m[nameKey{NodeDir, st}] = barkBold
		}
	}
	return m
}
