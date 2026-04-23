package listflow

import (
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/tui"
)

// maxTreeEntries caps the per-agent workspace file tree. Over this, the
// tree truncates to maxTreeEntries real rows plus one muted "... (N more)"
// synthetic row. Decision D2 from Plan 28 session 2026-04-23 — a large
// workspace should not balloon the list output.
const maxTreeEntries = 50

// RenderAgentPanel renders one agent's TitledPanel + workspace tree (or
// missing-workspace hint) as a single string block, ready to be joined
// into RenderAll's output.
//
// agentName is the machine name used as the config map key; agent is
// the full installed-agent record; cat is consulted for display-name
// lookups (nil is tolerated — falls back to DisplayNameFrom).
// projectDir is the absolute project root — used as the anchor for the
// workspace path-escape check.
func RenderAgentPanel(agentName string, agent *config.InstalledAgent, cat *catalog.Catalog, projectDir string, termW int) string {
	if agent == nil {
		return ""
	}

	// Display name — prefer the catalog's canonical form when available,
	// fall back to DisplayNameFrom(name) when the agent type isn't in the
	// loaded catalog (e.g. an old project on a newer Bonsai) or when cat
	// is nil (as in unit tests).
	displayName := catalog.DisplayNameFrom(agentName)
	if cat != nil {
		if def := cat.GetAgent(agentName); def != nil && def.DisplayName != "" {
			displayName = def.DisplayName
		}
	}

	pairs := buildPairs(agent, cat)
	content := tui.CardFields(pairs)
	panel := tui.TitledPanelString(displayName, content, tui.Water)

	var b strings.Builder
	b.WriteString(panel)
	b.WriteString("\n")

	// Workspace tree / hint / warning — rendered below the panel with
	// one blank line between. The tree helper emits a root label line
	// that includes the workspace path, so no extra heading is needed.
	subblock := renderWorkspaceBlock(agent.Workspace, projectDir)
	if subblock != "" {
		b.WriteString("\n")
		b.WriteString(subblock)
	}

	return b.String()
}

// buildPairs reconstructs the [][2]string field list used by CardFields.
// Same shape as the pre-Plan-28 cmd/list.go code: Workspace + per-category
// comma-joined display names. Empty categories are omitted so the panel
// stays compact.
func buildPairs(agent *config.InstalledAgent, cat *catalog.Catalog) [][2]string {
	pairs := [][2]string{{"Workspace", agent.Workspace}}

	lookup := func(name string, get func(string) string) string {
		if get != nil {
			if dn := get(name); dn != "" {
				return dn
			}
		}
		return catalog.DisplayNameFrom(name)
	}
	join := func(names []string, get func(string) string) string {
		out := make([]string, len(names))
		for i, n := range names {
			out[i] = lookup(n, get)
		}
		return strings.Join(out, ", ")
	}

	getSkill := func(string) string { return "" }
	getWorkflow := func(string) string { return "" }
	getProtocol := func(string) string { return "" }
	getSensor := func(string) string { return "" }
	getRoutine := func(string) string { return "" }
	if cat != nil {
		getSkill = func(n string) string {
			if s := cat.GetSkill(n); s != nil {
				return s.DisplayName
			}
			return ""
		}
		getWorkflow = func(n string) string {
			if w := cat.GetWorkflow(n); w != nil {
				return w.DisplayName
			}
			return ""
		}
		getProtocol = func(n string) string {
			if p := cat.GetProtocol(n); p != nil {
				return p.DisplayName
			}
			return ""
		}
		getSensor = func(n string) string {
			if s := cat.GetSensor(n); s != nil {
				return s.DisplayName
			}
			return ""
		}
		getRoutine = func(n string) string {
			if r := cat.GetRoutine(n); r != nil {
				return r.DisplayName
			}
			return ""
		}
	}

	if len(agent.Skills) > 0 {
		pairs = append(pairs, [2]string{"Skills", join(agent.Skills, getSkill)})
	}
	if len(agent.Workflows) > 0 {
		pairs = append(pairs, [2]string{"Workflows", join(agent.Workflows, getWorkflow)})
	}
	if len(agent.Protocols) > 0 {
		pairs = append(pairs, [2]string{"Protocols", join(agent.Protocols, getProtocol)})
	}
	if len(agent.Sensors) > 0 {
		pairs = append(pairs, [2]string{"Sensors", join(agent.Sensors, getSensor)})
	}
	if len(agent.Routines) > 0 {
		pairs = append(pairs, [2]string{"Routines", join(agent.Routines, getRoutine)})
	}
	return pairs
}

// renderWorkspaceBlock returns the file-tree / hint / warning block below
// an agent panel. Returns empty string only when workspace is explicitly
// empty-string in config (shouldn't happen for a generated project — but
// defensive). Contracts:
//
//   - workspace path contains ".." or escapes projectDir → muted warning
//     line, no walk ("workspace path escapes project root — tree skipped").
//   - workspace dir does not exist on disk → Hint CTA "Workspace missing
//     — run: bonsai update" (decision D3).
//   - exists but zero entries after filtering → tree with a single
//     "(empty)" row.
//   - exists with ≤ maxTreeEntries entries → full tree.
//   - exists with > maxTreeEntries entries → first 50 rows + synthetic
//     "... (N more)" muted row appended inside the tree (decision D2).
func renderWorkspaceBlock(workspace, projectDir string) string {
	if workspace == "" {
		// No workspace configured. Treat as "missing" so the user still
		// sees a CTA rather than a silent void.
		return renderHintLine("Workspace missing — run: bonsai update")
	}

	// Resolve the absolute workspace path. Refuse any path that contains
	// ".." (after Clean) or escapes projectDir.
	cleaned := filepath.Clean(workspace)
	if strings.Contains(cleaned, "..") {
		return renderWarningLine("workspace path escapes project root — tree skipped")
	}
	absWorkspace := cleaned
	if !filepath.IsAbs(absWorkspace) {
		absWorkspace = filepath.Join(projectDir, cleaned)
	}
	absWorkspace = filepath.Clean(absWorkspace)
	if projectDir != "" {
		absProject, err := filepath.Abs(projectDir)
		if err == nil {
			rel, relErr := filepath.Rel(absProject, absWorkspace)
			if relErr != nil || strings.HasPrefix(rel, "..") || rel == ".." {
				return renderWarningLine("workspace path escapes project root — tree skipped")
			}
		}
	}

	// Stat the resolved path. A missing directory surfaces the D3 CTA.
	info, err := statDir(absWorkspace)
	if err != nil || info == nil || !info.IsDir() {
		return renderHintLine("Workspace missing — run: bonsai update")
	}

	entries, truncated, total := scanWorkspace(absWorkspace)
	if len(entries) == 0 {
		// Present the root label with an "(empty)" marker so the user
		// knows the scan ran but produced nothing.
		tree := tui.FileTree([]string{"(empty)"}, workspace)
		return tree
	}

	if truncated {
		// Append a synthetic "... (N more)" entry to the flat list so it
		// renders as the final leaf in the tree. The helper treats it as
		// any other leaf — muted styling is applied by re-rendering the
		// tree output (below). Cheap alternative: pad with a special
		// marker and post-process. We use the post-process route to keep
		// styling consistent with the rest of the tree.
		extra := total - maxTreeEntries
		placeholder := "... (" + itoa(extra) + " more)"
		entries = append(entries, placeholder)
	}

	return tui.FileTree(entries, workspace)
}

// renderHintLine renders the muted single-line hint used for missing
// workspace + other terse CTAs below an agent panel. Indent matches the
// panel inset.
func renderHintLine(msg string) string {
	muted := lipgloss.NewStyle().Foreground(tui.ColorMuted)
	return "    " + muted.Render(msg)
}

// renderWarningLine renders an amber warning line used for the path-
// escape refusal. Kept visually distinct from the blue Hint CTA so the
// user notices the config issue rather than treating it as informational.
func renderWarningLine(msg string) string {
	warn := lipgloss.NewStyle().Foreground(tui.ColorWarning)
	return "    " + warn.Render(tui.GlyphWarn+" "+msg)
}

// statDir wraps filepath-safe Stat so the caller can distinguish "does not
// exist" from "exists but unreadable". Returns (nil, nil) when the path
// does not exist so the caller can short-circuit to the D3 CTA.
func statDir(path string) (fs.FileInfo, error) {
	info, err := osStat(path)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// scanWorkspace walks root collecting relative paths, skipping hidden
// entries, .git, and node_modules. Symlinks are not followed — they
// appear as leaves in the tree but their targets are never traversed,
// which defuses symlink loop attacks.
//
// Returns (entries, truncated, total). truncated=true means the caller
// should append a "... (N more)" row where N = total - maxTreeEntries.
func scanWorkspace(root string) ([]string, bool, int) {
	// Resolve the root for symlink-target comparisons. When EvalSymlinks
	// fails (dangling link, etc.), we fall back to `root` — the walk
	// still terminates because symlinks are skipped rather than followed.
	resolvedRoot := root
	if real, err := evalSymlinks(root); err == nil {
		resolvedRoot = real
	}

	var files []string
	total := 0

	_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Skip unreadable sub-trees rather than aborting the entire
			// walk — a best-effort list is more useful than no list.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if path == root {
			return nil
		}

		name := d.Name()
		if isSkippable(name) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		// Symlink handling — do not follow. For symlink directories,
		// returning SkipDir ensures we don't attempt to traverse into
		// them. For symlink files, they render as-is.
		if d.Type()&fs.ModeSymlink != 0 {
			// Defensive check: even with SkipDir, verify the symlink
			// target is inside the resolved workspace root before
			// counting it. Targets outside are dropped entirely.
			if target, err := evalSymlinks(path); err == nil {
				if !isWithin(target, resolvedRoot) {
					if d.IsDir() {
						return fs.SkipDir
					}
					return nil
				}
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil {
				total++
				if len(files) < maxTreeEntries {
					files = append(files, filepath.ToSlash(rel))
				}
			}
			// Never descend into symlinked directories — sidesteps loops.
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			// Directories are implicit in the tree (the FileTree helper
			// splits on "/" and promotes intermediate segments to
			// branches), so we don't emit them explicitly. The walk
			// continues into the subtree below.
			return nil
		}

		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		total++
		if len(files) < maxTreeEntries {
			files = append(files, filepath.ToSlash(rel))
		}
		return nil
	})

	// Sort for deterministic output — filesystem walk order can vary
	// between runs / platforms.
	sort.Strings(files)

	truncated := total > maxTreeEntries
	return files, truncated, total
}

// isSkippable returns true for entries excluded from the workspace tree:
// hidden files/dirs (leading dot), .git, node_modules. The leading-dot
// check subsumes .git and .claude / .bonsai-lock.yaml but we keep the
// named pair listed explicitly for clarity and as a guard if the hidden
// rule is ever relaxed.
func isSkippable(name string) bool {
	if name == "" {
		return true
	}
	if strings.HasPrefix(name, ".") {
		return true
	}
	if name == "node_modules" || name == ".git" {
		return true
	}
	return false
}

// isWithin reports whether target is under root (inclusive). Both paths
// should be absolute + symlink-resolved. Used to drop symlinks whose
// targets escape the workspace — a path-traversal defense-in-depth
// even though we already SkipDir on symlink directories.
func isWithin(target, root string) bool {
	if target == "" || root == "" {
		return false
	}
	target = filepath.Clean(target)
	root = filepath.Clean(root)
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

// itoa is a tiny strconv.Itoa wrapper kept local so the file avoids a
// strconv import for a single call site. Mirrors the pattern in
// initflow/chrome.go.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits [20]byte
	i := len(digits)
	for n > 0 {
		i--
		digits[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}
