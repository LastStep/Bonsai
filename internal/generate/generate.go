package generate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
)

// TemplateContext holds all variables available to Go templates.
type TemplateContext struct {
	ProjectName        string
	ProjectDescription string
	AgentName          string
	AgentDisplayName   string
	AgentDescription   string
	OtherAgents        []OtherAgent
	Workspace          string
	DocsPath           string
	Protocols          []string
	Skills             []string
	Workflows          []string
	Routines           []string
}

// OtherAgent describes a sibling agent for template rendering.
type OtherAgent struct {
	AgentType string
	Workspace string
}

var funcMap = template.FuncMap{
	"title": titleCase,
}

func titleCase(s string) string {
	var result strings.Builder
	capitalize := true
	for _, r := range s {
		if capitalize {
			result.WriteRune(unicode.ToUpper(r))
			capitalize = false
		} else {
			result.WriteRune(r)
		}
		if r == '-' || r == ' ' {
			capitalize = true
		}
	}
	return result.String()
}

func renderTemplate(fsys fs.FS, tmplPath string, ctx interface{}) (string, error) {
	data, err := fs.ReadFile(fsys, tmplPath)
	if err != nil {
		return "", err
	}
	tmpl, err := template.New(filepath.Base(tmplPath)).Funcs(funcMap).Parse(string(data))
	if err != nil {
		return "", fmt.Errorf("parsing template %s: %w", tmplPath, err)
	}
	var buf strings.Builder
	if err := tmpl.Execute(&buf, ctx); err != nil {
		return "", fmt.Errorf("executing template %s: %w", tmplPath, err)
	}
	return buf.String(), nil
}

func descFor(names []string, cat *catalog.Catalog, category string, customItems map[string]*config.CustomItemMeta) map[string]string {
	result := make(map[string]string)
	for _, name := range names {
		var desc string
		switch category {
		case "skill":
			if item := cat.GetSkill(name); item != nil {
				desc = item.Description
			}
		case "workflow":
			if item := cat.GetWorkflow(name); item != nil {
				desc = item.Description
			}
		case "protocol":
			if item := cat.GetProtocol(name); item != nil {
				desc = item.Description
			}
		case "routine":
			if item := cat.GetRoutine(name); item != nil {
				desc = item.Description
			}
		}
		// Fall back to custom item metadata
		if desc == "" && customItems != nil {
			if meta, ok := customItems[name]; ok && meta.Description != "" {
				desc = meta.Description
			}
		}
		if desc == "" {
			desc = catalog.DisplayNameFrom(name)
		}
		result[name] = desc
	}
	return result
}

// scenariosDesc returns a trigger-aware description for CLAUDE.md tables.
// Uses triggers.scenarios (joined with "; ") if available, falls back to description.
func scenariosDesc(item *catalog.CatalogItem) string {
	if item != nil && item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
		return strings.Join(item.Triggers.Scenarios, "; ")
	}
	if item != nil {
		return item.Description
	}
	return ""
}

// CuratedSlashWorkflows is the set of workflows that get .claude/skills/ files
// and appear in the Quick Triggers table. Keep this small — each entry consumes
// ~1,536 chars of context budget at session start.
var CuratedSlashWorkflows = map[string]bool{
	"planning": true, "code-review": true, "pr-review": true,
	"security-audit": true, "issue-to-implementation": true,
	"test-plan": true, "plan-execution": true,
}

// ─── Write result types ────────────────────────────────────────────────

// FileAction describes what happened to a single file during generation.
type FileAction int

const (
	ActionCreated   FileAction = iota // new file written
	ActionUpdated                     // existing unmodified file overwritten with new content
	ActionUnchanged                   // existing file identical to rendered content — no write
	ActionSkipped                     // file already exists (scaffolding write-once)
	ActionConflict                    // file modified by user, not overwritten
	ActionForced                      // conflict overridden by user, overwritten
)

// FileResult describes the outcome for one file.
type FileResult struct {
	RelPath string
	Action  FileAction
	Source  string
	content []byte      // stashed for force-retry on conflicts
	perm    os.FileMode // stashed for chmod on force-retry
}

// WriteResult collects all file outcomes from a generation operation.
type WriteResult struct {
	Files []FileResult
}

// Add appends a file result.
func (wr *WriteResult) Add(r FileResult) {
	wr.Files = append(wr.Files, r)
}

// Conflicts returns only the conflict entries.
func (wr *WriteResult) Conflicts() []FileResult {
	var out []FileResult
	for _, f := range wr.Files {
		if f.Action == ActionConflict {
			out = append(out, f)
		}
	}
	return out
}

// HasConflicts returns true if any files had conflicts.
func (wr *WriteResult) HasConflicts() bool {
	for _, f := range wr.Files {
		if f.Action == ActionConflict {
			return true
		}
	}
	return false
}

// Summary returns counts by action type.
func (wr *WriteResult) Summary() (created, updated, unchanged, skipped, conflicts int) {
	for _, f := range wr.Files {
		switch f.Action {
		case ActionCreated:
			created++
		case ActionUpdated, ActionForced:
			updated++
		case ActionUnchanged:
			unchanged++
		case ActionSkipped:
			skipped++
		case ActionConflict:
			conflicts++
		}
	}
	return
}

// ForceConflicts overwrites all conflict files using stashed content.
// Call after user confirmation.
func (wr *WriteResult) ForceConflicts(projectRoot string, lock *config.LockFile) {
	for i, f := range wr.Files {
		if f.Action != ActionConflict || f.content == nil {
			continue
		}
		absPath := filepath.Join(projectRoot, f.RelPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			continue
		}
		if err := os.WriteFile(absPath, f.content, 0644); err != nil {
			continue
		}
		if f.perm != 0 {
			_ = os.Chmod(absPath, f.perm)
		}
		lock.Track(f.RelPath, f.content, f.Source)
		wr.Files[i].Action = ActionForced
	}
}

// ForceSelected overwrites only the conflict files whose RelPath appears in paths.
// Unmatched conflicts remain as ActionConflict.
func (wr *WriteResult) ForceSelected(paths []string, projectRoot string, lock *config.LockFile) {
	selected := make(map[string]bool, len(paths))
	for _, p := range paths {
		selected[p] = true
	}
	for i, f := range wr.Files {
		if f.Action != ActionConflict || f.content == nil || !selected[f.RelPath] {
			continue
		}
		absPath := filepath.Join(projectRoot, f.RelPath)
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			continue
		}
		if err := os.WriteFile(absPath, f.content, 0644); err != nil {
			continue
		}
		if f.perm != 0 {
			_ = os.Chmod(absPath, f.perm)
		}
		lock.Track(f.RelPath, f.content, f.Source)
		wr.Files[i].Action = ActionForced
	}
}

// ─── Lock-aware write primitives ───────────────────────────────────────

// normalizeShellLF strips carriage returns from shell script content. It is a
// belt-and-braces defence against CRLF sneaking through git clients that
// ignore `.gitattributes` — bash refuses scripts whose shebang ends in `\r`.
// No-op for non-`.sh` paths.
func normalizeShellLF(data []byte, rel string) []byte {
	if !strings.HasSuffix(rel, ".sh") {
		return data
	}
	// Replace CRLF with LF first, then drop any remaining standalone CR.
	data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
	data = bytes.ReplaceAll(data, []byte("\r"), nil)
	return data
}

// writeFile implements the lock-aware write policy.
// If force is true, modified files are overwritten.
func writeFile(projectRoot, relPath string, content []byte, source string, lock *config.LockFile, force bool) FileResult {
	content = normalizeShellLF(content, relPath)
	absPath := filepath.Join(projectRoot, relPath)
	exists, modified := lock.IsModified(projectRoot, relPath)

	if !exists {
		if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
			return FileResult{RelPath: relPath, Action: ActionConflict, Source: source}
		}
		if err := os.WriteFile(absPath, content, 0644); err != nil {
			return FileResult{RelPath: relPath, Action: ActionConflict, Source: source}
		}
		lock.Track(relPath, content, source)
		return FileResult{RelPath: relPath, Action: ActionCreated, Source: source}
	}

	if modified && !force {
		return FileResult{RelPath: relPath, Action: ActionConflict, Source: source, content: content}
	}

	// If file exists, is not user-modified, and content matches rendered output, skip the write.
	if !modified {
		if existing, err := os.ReadFile(absPath); err == nil && bytes.Equal(existing, content) {
			return FileResult{RelPath: relPath, Action: ActionUnchanged, Source: source}
		}
	}

	if err := os.WriteFile(absPath, content, 0644); err != nil {
		return FileResult{RelPath: relPath, Action: ActionConflict, Source: source, content: content}
	}
	lock.Track(relPath, content, source)

	action := ActionUpdated
	if force && modified {
		action = ActionForced
	}
	return FileResult{RelPath: relPath, Action: action, Source: source}
}

// writeFileChmod is like writeFile but also sets file permissions (for sensor scripts).
func writeFileChmod(projectRoot, relPath string, content []byte, source string, lock *config.LockFile, force bool, perm os.FileMode) FileResult {
	content = normalizeShellLF(content, relPath)
	result := writeFile(projectRoot, relPath, content, source, lock, force)
	if result.Action == ActionCreated || result.Action == ActionUpdated || result.Action == ActionForced || result.Action == ActionUnchanged {
		absPath := filepath.Join(projectRoot, relPath)
		_ = os.Chmod(absPath, perm)
	}
	// Stash perm for force-retry
	result.perm = perm
	return result
}

// renderContent renders a template or reads a raw file from the catalog FS,
// returning the content bytes without writing to disk.
func renderContent(fsys fs.FS, srcPath string, ctx interface{}) ([]byte, error) {
	if strings.HasSuffix(srcPath, ".tmpl") {
		content, err := renderTemplate(fsys, srcPath, ctx)
		if err != nil {
			return nil, err
		}
		return []byte(content), nil
	}
	return fs.ReadFile(fsys, srcPath)
}

// ─── Helpers ───────────────────────────────────────────────────────────

// hasScaffolding checks if a scaffolding item is selected in the project config.
// Returns true if scaffolding list is empty (backward compat: old configs without the field).
func hasScaffolding(cfg *config.ProjectConfig, name string) bool {
	if len(cfg.Scaffolding) == 0 {
		return true
	}
	for _, s := range cfg.Scaffolding {
		if s == name {
			return true
		}
	}
	return false
}

// Scaffolding generates project management infrastructure files for selected items.
// Scaffolding files are write-once: if a file already exists, it is skipped (not conflicted).
func Scaffolding(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	catFS := cat.FS()
	docsRoot := projectRoot
	if cfg.DocsPath != "" {
		docsRoot = filepath.Join(projectRoot, cfg.DocsPath)
	}
	ctx := &TemplateContext{
		ProjectName:        cfg.ProjectName,
		ProjectDescription: cfg.Description,
	}

	// Build set of allowed file prefixes from selected scaffolding items
	allowedFiles := make(map[string]bool)
	for _, name := range cfg.Scaffolding {
		item := cat.GetScaffolding(name)
		if item == nil {
			continue
		}
		for _, f := range item.Files {
			allowedFiles[f] = true
		}
	}

	err := fs.WalkDir(catFS, "scaffolding", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(path, "scaffolding/")
		if rel == "manifest.yaml" {
			return nil // skip the manifest itself
		}

		if !isAllowedScaffoldingFile(rel, allowedFiles) {
			return nil
		}

		content, err := renderContent(catFS, path, ctx)
		if err != nil {
			return err
		}

		relToProject := rel
		if strings.HasSuffix(relToProject, ".tmpl") {
			relToProject = strings.TrimSuffix(relToProject, ".tmpl")
		}
		if cfg.DocsPath != "" {
			relToProject = filepath.Join(cfg.DocsPath, relToProject)
		}

		// Scaffolding is write-once: if file exists, skip (don't conflict)
		absPath := filepath.Join(projectRoot, relToProject)
		if _, statErr := os.Stat(absPath); statErr == nil {
			result.Add(FileResult{RelPath: relToProject, Action: ActionSkipped, Source: "scaffolding:" + rel})
			return nil
		}

		r := writeFile(projectRoot, relToProject, content, "scaffolding:"+rel, lock, force)
		result.Add(r)
		return nil
	})

	// Create empty directories listed in selected items (e.g. Plans/Active/, Reports/Pending/)
	for _, f := range sortedKeys(allowedFiles) {
		if strings.HasSuffix(f, "/") {
			dirPath := filepath.Join(docsRoot, f)
			_ = os.MkdirAll(dirPath, 0755)
		}
	}

	return err
}

// isAllowedScaffoldingFile checks if a file path matches any allowed file entry.
// Handles both exact file matches and directory prefix matches (entries ending with /).
func isAllowedScaffoldingFile(rel string, allowed map[string]bool) bool {
	// Exact match (with or without .tmpl suffix)
	if allowed[rel] {
		return true
	}
	if allowed[rel+".tmpl"] {
		return true
	}
	if strings.HasSuffix(rel, ".tmpl") {
		if allowed[strings.TrimSuffix(rel, ".tmpl")] {
			return true
		}
	}
	// Directory prefix match — file is under an allowed directory
	for f := range allowed {
		if strings.HasSuffix(f, "/") && strings.HasPrefix(rel, f) {
			return true
		}
	}
	return false
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// SettingsJSON generates or updates .claude/settings.json with sensor hooks
// for EVERY installed agent. Used by `bonsai update` which regenerates all
// agents in one pass. `bonsai add` should use SettingsJSONForAgent to scope
// regeneration to a single agent — otherwise a concurrent `bonsai add` would
// flag user edits in unrelated agent workspaces as conflicts. Plan 27 §B2.
//
// Settings are written per-workspace so users can launch Claude Code from
// there directly.
func SettingsJSON(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	for _, installed := range cfg.Agents {
		if err := writeSettingsJSONForAgent(projectRoot, installed, cat, lock, result, force); err != nil {
			return err
		}
	}
	return nil
}

// SettingsJSONForAgent is the single-agent scoped variant of SettingsJSON.
// Used by `bonsai add` so the regeneration pass only touches the agent being
// added — prior `bonsai update` leaves SettingsJSON unchanged (still iterates
// cfg.Agents). Plan 27 §B2 fix for the cross-agent conflict leak that tripped
// dogfood finding #7b.
func SettingsJSONForAgent(projectRoot string, agent *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	_ = cfg // retained for signature symmetry with the all-agents variant
	if agent == nil {
		return nil
	}
	return writeSettingsJSONForAgent(projectRoot, agent, cat, lock, result, force)
}

// writeSettingsJSONForAgent renders .claude/settings.json under the given
// installed agent's workspace. Shared helper between SettingsJSON (all
// agents) and SettingsJSONForAgent (single agent).
func writeSettingsJSONForAgent(projectRoot string, installed *config.InstalledAgent, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	type hookEntry struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	type hookGroup struct {
		Hooks   []hookEntry `json:"hooks"`
		Matcher string      `json:"matcher,omitempty"`
	}
	type groupKey struct{ event, matcher string }

	settingsPath := filepath.Join(projectRoot, installed.Workspace, ".claude", "settings.json")

	existing := make(map[string]interface{})
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	groups := make(map[groupKey][]string)

	for _, sensorName := range installed.Sensors {
		var event, matcher string
		if sensor := cat.GetSensor(sensorName); sensor != nil {
			event = sensor.Event
			matcher = sensor.Matcher
		} else if installed.CustomItems != nil {
			if meta, ok := installed.CustomItems[sensorName]; ok && meta.Event != "" {
				event = meta.Event
				matcher = meta.Matcher
			}
		}
		if event == "" {
			continue
		}
		k := groupKey{event, matcher}
		scriptPath := installed.Workspace + "agent/Sensors/" + sensorName + ".sh"
		cmd := fmt.Sprintf(
			`bash -c 'r="$PWD"; while [ "$r" != "/" ] && [ ! -f "$r/.bonsai.yaml" ]; do r=$(dirname "$r"); done; bash "$r/%s" "$r"'`,
			scriptPath,
		)
		groups[k] = append(groups[k], cmd)
	}

	hooksConfig := make(map[string][]hookGroup)
	for k, commands := range groups {
		var hooks []hookEntry
		for _, cmd := range commands {
			hooks = append(hooks, hookEntry{Type: "command", Command: cmd})
		}
		g := hookGroup{Hooks: hooks}
		if k.matcher != "" {
			g.Matcher = k.matcher
		}
		hooksConfig[k.event] = append(hooksConfig[k.event], g)
	}

	if len(hooksConfig) > 0 {
		existing["hooks"] = hooksConfig
	} else {
		delete(existing, "hooks")
	}

	if err := os.MkdirAll(filepath.Dir(settingsPath), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}
	content := append(data, '\n')
	relPath := filepath.Join(installed.Workspace, ".claude", "settings.json")
	r := writeFile(projectRoot, relPath, content, "generated:settings-json", lock, force)
	result.Add(r)
	return nil
}

// howToWorkLines generates the compact "How to Work" heuristics section for the workspace CLAUDE.md.
func howToWorkLines(agentName string, docsPrefix string, hasRoutines bool, hasWorkspaceGuide bool) []string {
	var lines []string
	lines = append(lines,
		"### How to Work", "",
		"> Decision heuristics — how to use this workspace effectively.", "",
	)

	// Shared heuristics (all agents)
	lines = append(lines,
		fmt.Sprintf("- **Before starting work:** Check `%sPlaybook/Status.md` for assigned tasks and `%sPlaybook/Plans/Active/` for your current plan.", docsPrefix, docsPrefix),
		"- **When to load a Workflow:** You are starting a multi-step activity (planning, reviewing, auditing). Load the matching workflow from the table above and follow it end-to-end.",
		"- **When to load a Skill:** You need reference standards for a specific domain (coding style, API design, test strategy). Load it, use it, move on.",
		fmt.Sprintf("- **Decision logging:** When you make or observe a significant architectural decision, append it to `%sLogs/KeyDecisionLog.md`.", docsPrefix),
		fmt.Sprintf("- **Out-of-scope findings:** Don't fix bugs, debt, or improvements outside your current task. Add them to `%sPlaybook/Backlog.md`.", docsPrefix),
		"- **Workspace evolution:** `bonsai add` (new abilities), `bonsai remove` (uninstall), `bonsai update` (sync custom files), `bonsai list` (see installed), `bonsai catalog` (browse available).",
	)

	// Agent-type-specific heuristics
	switch agentName {
	case "tech-lead":
		lines = append(lines,
			"- **You orchestrate, not implement.** Plan features, dispatch to code agents via worktrees, review their output. Never write application code directly.",
			fmt.Sprintf("- **Check Backlog first:** Before creating new work items, check `%sPlaybook/Backlog.md` for existing entries.", docsPrefix),
			fmt.Sprintf("- **After completing work:** Update `%sPlaybook/Status.md` and log results.", docsPrefix),
		)
	case "backend", "frontend", "fullstack":
		lines = append(lines,
			fmt.Sprintf("- **Plan first:** Read your assigned plan in `%sPlaybook/Plans/Active/` before writing any code. Follow it exactly.", docsPrefix),
			"- **When stuck:** If the plan is ambiguous, stop and report — don't guess or make design decisions.",
			"- **Stay in scope:** Only modify files within your workspace boundary.",
		)
	case "devops":
		lines = append(lines,
			"- **Plan before apply:** Never auto-approve destructive infrastructure operations. Require explicit user confirmation.",
			"- **Stay in scope:** Only modify infrastructure and deployment files within your workspace boundary.",
		)
	case "security":
		lines = append(lines,
			"- **Audit and report.** You read the entire codebase but only modify security-owned files.",
			"- **Evidence-based findings:** Every finding must reference a specific file, line, and standard (OWASP, CWE, CVE).",
		)
	}

	// Workspace guide pointer
	if hasWorkspaceGuide {
		lines = append(lines,
			"- **New to this workspace?** Load `agent/Skills/workspace-guide.md` for a full operational reference.",
		)
	}

	lines = append(lines, "")

	return lines
}

// bonsaiReferenceLines builds the "Bonsai Reference" block rendered after
// the Core navigation table in every agent's workspace CLAUDE.md. Points to:
//
//   - bonsai-model.md — the mental-model skill (tech-lead-only; for non-
//     tech-lead agents, path resolves into tech-lead's workspace via
//     cfg.DocsPath, which is tech-lead's workspace by convention).
//   - .bonsai/catalog.json — filesystem-discoverable catalog snapshot
//     (written by WriteCatalogSnapshot on every init/add/update).
//   - .bonsai.yaml — project config (installed-state source of truth).
//
// All paths are computed relative to workspaceRoot using filepath.Rel so the
// block renders correctly from any workspace depth (e.g. "station/" one
// level deep, or "services/backend/" two levels deep). Plan 31 Phase D.
func bonsaiReferenceLines(projectRoot, workspaceRoot string, cfg *config.ProjectConfig) []string {
	// relFromWorkspace resolves a project-root-relative path into one that's
	// relative to the workspace root (where CLAUDE.md lives). Falls back to
	// the input if Rel fails — no broken link, just a less-friendly path.
	relFromWorkspace := func(fromProjectRoot string) string {
		abs := filepath.Join(projectRoot, fromProjectRoot)
		rel, err := filepath.Rel(workspaceRoot, abs)
		if err != nil {
			return fromProjectRoot
		}
		return filepath.ToSlash(rel)
	}

	// bonsai-model.md lives in the tech-lead workspace (cfg.DocsPath).
	// For tech-lead itself, DocsPath == installed.Workspace so the path
	// resolves to "agent/Skills/bonsai-model.md" (same as quick-triggers
	// refs). For non-tech-lead agents, it resolves to something like
	// "../station/agent/Skills/bonsai-model.md".
	docsPath := cfg.DocsPath
	if docsPath == "" {
		docsPath = "station/"
	}
	bonsaiModelRel := relFromWorkspace(filepath.Join(docsPath, "agent", "Skills", "bonsai-model.md"))
	catalogJSONRel := relFromWorkspace(filepath.Join(".bonsai", "catalog.json"))
	bonsaiYAMLRel := relFromWorkspace(".bonsai.yaml")

	return []string{
		"---", "",
		"## Bonsai Reference", "",
		"> Read these when reasoning about Bonsai itself — what catalog items exist, how to customize, what `bonsai add`/`remove`/`update` do.", "",
		"| Need | Read |",
		"|------|------|",
		fmt.Sprintf("| Bonsai mental model — catalog shape, customization, decisions | [%s](%s) |", bonsaiModelRel, bonsaiModelRel),
		fmt.Sprintf("| Available abilities (all catalog items) | [%s](%s) |", catalogJSONRel, catalogJSONRel),
		fmt.Sprintf("| Current installed state | [%s](%s) |", bonsaiYAMLRel, bonsaiYAMLRel),
		"",
	}
}

func quickTriggersLines(installed *config.InstalledAgent, cat *catalog.Catalog) []string {
	var lines []string
	lines = append(lines,
		"### Quick Triggers", "",
		"> Common phrases and commands that activate specific behaviors.", "",
		"| You want to... | Say or do this |",
		"|----------------|---------------|",
		"| Start a session | \"Hi, get started\" |",
	)

	// Add workflow-based triggers from curated set — derive descriptions from catalog
	for _, w := range installed.Workflows {
		if !CuratedSlashWorkflows[w] {
			continue
		}
		desc := catalog.DisplayNameFrom(w)
		if item := cat.GetWorkflow(w); item != nil {
			if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
				desc = item.Triggers.Scenarios[0]
			} else {
				desc = item.Description
			}
		}
		lines = append(lines, fmt.Sprintf("| %s | \"[describe task]\" or `/%s` |", desc, w))
	}

	lines = append(lines,
		"| Self-review before shipping | \"Verify everything\" |",
		"| End session | \"That's all\" |",
		"", "",
	)
	return lines
}

const (
	bonsaiStartMarker = "<!-- BONSAI_START -->"
	bonsaiEndMarker   = "<!-- BONSAI_END -->"
)

// WorkspaceClaudeMD generates the workspace CLAUDE.md with navigation tables.
// Includes section markers for safe partial updates. Custom items from
// installed.CustomItems are included in nav tables alongside abilities.
func WorkspaceClaudeMD(projectRoot string, workspaceRoot string, agentDef *catalog.AgentDef, installed *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	docsPrefix := cfg.DocsPath
	custom := installed.CustomItems // may be nil

	var lines []string
	lines = append(lines,
		fmt.Sprintf("# %s — %s", cfg.ProjectName, agentDef.DisplayName), "",
		fmt.Sprintf("**Working directory:** `%s`", installed.Workspace), "",
		"> [!warning]",
		"> **FIRST:** Read [agent/Core/identity.md](agent/Core/identity.md), then [agent/Core/memory.md](agent/Core/memory.md).", "",
		"---", "",
		"## Navigation", "",
		"> All agent instruction files live in `agent/`.", "",
		"### Core (load first, every session)", "",
		"| File | Purpose |",
		"|------|---------|",
		"| [agent/Core/identity.md](agent/Core/identity.md) | Who I am, relationships, mindset |",
		"| [agent/Core/memory.md](agent/Core/memory.md) | Working memory — flags, work state, notes |",
		"| [agent/Core/self-awareness.md](agent/Core/self-awareness.md) | Context monitoring, hard thresholds |", "",
	)

	// Bonsai Reference block (Plan 31 Phase D) — pi-style pointer so the
	// user's agent knows where to find the Bonsai mental model + catalog
	// snapshot without being hand-fed docs. Every agent points to tech-lead's
	// copy of bonsai-model.md (the skill is tech-lead-only; project-level
	// knowledge, not per-agent). Paths are computed relative to the workspace
	// root (where CLAUDE.md lives) so they resolve correctly from any
	// workspace depth.
	bonsaiRefLines := bonsaiReferenceLines(projectRoot, workspaceRoot, cfg)
	lines = append(lines, bonsaiRefLines...)

	// Quick Triggers reference
	qt := quickTriggersLines(installed, cat)
	lines = append(lines, qt...)

	if len(installed.Protocols) > 0 {
		protoDescs := descFor(installed.Protocols, cat, "protocol", custom)
		lines = append(lines,
			"### Protocols (load after Core, every session)", "",
			"| File | Purpose |",
			"|------|---------|",
		)
		for _, p := range installed.Protocols {
			lines = append(lines, fmt.Sprintf("| [agent/Protocols/%s.md](agent/Protocols/%s.md) | %s |", p, p, protoDescs[p]))
		}
		lines = append(lines, "")
	}

	if len(installed.Workflows) > 0 {
		wfDescs := descFor(installed.Workflows, cat, "workflow", custom)
		lines = append(lines,
			"### Workflows (load when starting an activity)", "",
			"| Activate when... | Read this |",
			"|------------------|-----------|",
		)
		for _, w := range installed.Workflows {
			desc := wfDescs[w]
			if item := cat.GetWorkflow(w); item != nil {
				if sd := scenariosDesc(item); sd != "" {
					desc = sd
				}
			}
			lines = append(lines, fmt.Sprintf("| %s | [agent/Workflows/%s.md](agent/Workflows/%s.md) |", desc, w, w))
		}
		lines = append(lines, "")
	}

	if len(installed.Skills) > 0 {
		skillDescs := descFor(installed.Skills, cat, "skill", custom)
		lines = append(lines,
			"### Skills (load when doing specific work)", "",
			"| Activate when... | Read this |",
			"|------------------|-----------|",
		)
		for _, s := range installed.Skills {
			desc := skillDescs[s]
			if item := cat.GetSkill(s); item != nil {
				if sd := scenariosDesc(item); sd != "" {
					desc = sd
				}
			}
			lines = append(lines, fmt.Sprintf("| %s | [agent/Skills/%s.md](agent/Skills/%s.md) |", desc, s, s))
		}
		lines = append(lines, "")
	}

	if len(installed.Routines) > 0 {
		lines = append(lines,
			"### Routines (periodic self-maintenance)", "",
			"| Routine | Frequency | File |",
			"|---------|-----------|------|",
		)
		for _, r := range installed.Routines {
			freq := ""
			displayName := catalog.DisplayNameFrom(r)
			if routine := cat.GetRoutine(r); routine != nil {
				freq = routine.Frequency
				displayName = routine.DisplayName
			} else if custom != nil {
				if meta, ok := custom[r]; ok {
					if meta.Frequency != "" {
						freq = meta.Frequency
					}
					if meta.DisplayName != "" {
						displayName = meta.DisplayName
					}
				}
			}
			lines = append(lines, fmt.Sprintf("| %s | %s | [agent/Routines/%s.md](agent/Routines/%s.md) |",
				displayName, freq, r, r))
		}
		lines = append(lines, "",
			"> Routines are opt-in — check [agent/Core/routines.md](agent/Core/routines.md) for the dashboard and procedures.", "")
	}

	if len(installed.Sensors) > 0 {
		lines = append(lines,
			"### Sensors (auto-enforced via hooks)", "",
			"| Sensor | Event | What it does |",
			"|--------|-------|-------------|",
		)
		for _, sensorName := range installed.Sensors {
			sensor := cat.GetSensor(sensorName)
			if sensor != nil {
				eventStr := sensor.Event
				if sensor.Matcher != "" {
					eventStr += fmt.Sprintf(" (%s)", sensor.Matcher)
				}
				lines = append(lines, fmt.Sprintf("| [agent/Sensors/%s.sh](agent/Sensors/%s.sh) | %s | %s |", sensorName, sensorName, eventStr, sensor.Description))
			} else if custom != nil {
				if meta, ok := custom[sensorName]; ok && meta.Event != "" {
					eventStr := meta.Event
					if meta.Matcher != "" {
						eventStr += fmt.Sprintf(" (%s)", meta.Matcher)
					}
					desc := meta.Description
					if desc == "" {
						desc = catalog.DisplayNameFrom(sensorName)
					}
					lines = append(lines, fmt.Sprintf("| [agent/Sensors/%s.sh](agent/Sensors/%s.sh) | %s | %s |", sensorName, sensorName, eventStr, desc))
				}
			}
		}
		lines = append(lines, "",
			"> Sensors run automatically — they are configured in `.claude/settings.json`.", "")
	}

	// How to Work section
	hasWorkspaceGuide := false
	for _, s := range installed.Skills {
		if s == "workspace-guide" {
			hasWorkspaceGuide = true
			break
		}
	}
	htw := howToWorkLines(agentDef.Name, docsPrefix, len(installed.Routines) > 0, hasWorkspaceGuide)
	lines = append(lines, htw...)

	lines = append(lines,
		"---", "",
		"## Memory", "",
		"> [!warning]",
		"> **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in [agent/Core/memory.md](agent/Core/memory.md) — version-controlled, auditable, inside the project.", "",
		"When you would normally write to auto-memory (feedback, references, project context, flags), write to the appropriate section in [agent/Core/memory.md](agent/Core/memory.md) instead.", "",
		"---", "",
		"### External References", "",
		"| Need | Read this |",
		"|------|-----------|",
	)

	// extRef computes the relative path from the workspace root (where CLAUDE.md lives)
	// to a file inside DocsPath. Display text keeps docsPrefix for agent context.
	extRef := func(target string) string {
		full := filepath.Join(cfg.DocsPath, target)
		rel, err := filepath.Rel(installed.Workspace, full)
		if err != nil {
			return target
		}
		return filepath.ToSlash(rel)
	}

	lines = append(lines,
		fmt.Sprintf("| Project snapshot | [%sINDEX.md](%s) |", docsPrefix, extRef("INDEX.md")),
		fmt.Sprintf("| Current work status | [%sPlaybook/Status.md](%s) |", docsPrefix, extRef("Playbook/Status.md")),
		fmt.Sprintf("| Long-term direction | [%sPlaybook/Roadmap.md](%s) |", docsPrefix, extRef("Playbook/Roadmap.md")),
		fmt.Sprintf("| Security standards | [%sPlaybook/Standards/SecurityStandards.md](%s) |", docsPrefix, extRef("Playbook/Standards/SecurityStandards.md")),
		fmt.Sprintf("| Your assigned plan | [%sPlaybook/Plans/Active/](%s) |", docsPrefix, extRef("Playbook/Plans/Active/")),
		fmt.Sprintf("| Backlog | [%sPlaybook/Backlog.md](%s) |", docsPrefix, extRef("Playbook/Backlog.md")),
		fmt.Sprintf("| Prior decisions | [%sLogs/KeyDecisionLog.md](%s) |", docsPrefix, extRef("Logs/KeyDecisionLog.md")),
	)
	if hasScaffolding(cfg, "reports") {
		lines = append(lines, fmt.Sprintf("| Submit report | [%sReports/Pending/](%s) |", docsPrefix, extRef("Reports/Pending/")))
	}
	lines = append(lines, "")

	generatedContent := strings.Join(lines, "\n")
	relPath, _ := filepath.Rel(projectRoot, filepath.Join(workspaceRoot, "CLAUDE.md"))
	absPath := filepath.Join(projectRoot, relPath)

	// Check for existing file with markers — preserve user content outside markers
	if existing, err := os.ReadFile(absPath); err == nil {
		existingStr := string(existing)
		startIdx := strings.Index(existingStr, bonsaiStartMarker)
		endIdx := strings.Index(existingStr, bonsaiEndMarker)

		if startIdx >= 0 && endIdx >= 0 && endIdx > startIdx {
			// Markers found — splice in new content, preserve content outside markers
			beforeMarkers := existingStr[:startIdx]
			afterMarkers := existingStr[endIdx+len(bonsaiEndMarker):]

			fullContent := beforeMarkers + bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + afterMarkers
			contentBytes := []byte(fullContent)

			// Short-circuit if content is already identical
			if bytes.Equal(existing, contentBytes) {
				result.Add(FileResult{RelPath: relPath, Action: ActionUnchanged, Source: "generated:workspace-claude-md"})
				return nil
			}

			if err := os.WriteFile(absPath, contentBytes, 0644); err != nil {
				return err
			}
			lock.Track(relPath, contentBytes, "generated:workspace-claude-md")
			result.Add(FileResult{RelPath: relPath, Action: ActionUpdated, Source: "generated:workspace-claude-md"})
			return nil
		}

		// File exists but no markers — migrate: backup + overwrite with markers
		_ = os.WriteFile(absPath+".bak", existing, 0644)
		fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")
		if err := os.WriteFile(absPath, fullContent, 0644); err != nil {
			return err
		}
		lock.Track(relPath, fullContent, "generated:workspace-claude-md")
		result.Add(FileResult{RelPath: relPath, Action: ActionUpdated, Source: "generated:workspace-claude-md"})
		return nil
	}

	// No existing file — create with markers via lock-aware write
	fullContent := []byte(bonsaiStartMarker + "\n" + generatedContent + bonsaiEndMarker + "\n")
	r := writeFile(projectRoot, relPath, fullContent, "generated:workspace-claude-md", lock, force)
	result.Add(r)
	return nil
}

// EnsureRoutineCheckSensor adds or removes the routine-check sensor based on whether routines are installed.
// Call this before generating settings.json.
func EnsureRoutineCheckSensor(installed *config.InstalledAgent) {
	const sensorName = "routine-check"
	hasRoutines := len(installed.Routines) > 0

	hasSensor := false
	for _, s := range installed.Sensors {
		if s == sensorName {
			hasSensor = true
			break
		}
	}

	if hasRoutines && !hasSensor {
		installed.Sensors = append(installed.Sensors, sensorName)
	} else if !hasRoutines && hasSensor {
		filtered := installed.Sensors[:0]
		for _, s := range installed.Sensors {
			if s != sensorName {
				filtered = append(filtered, s)
			}
		}
		installed.Sensors = filtered
	}
}

// parseFrequencyDays extracts the number of days from a frequency string like "5 days" or "14 days".
func parseFrequencyDays(freq string) int {
	parts := strings.Fields(freq)
	if len(parts) >= 1 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			return n
		}
	}
	return 7 // default fallback
}

// RoutineDashboard generates or updates agent/Core/routines.md with the current routine configuration.
// Preserves last_ran dates from the existing dashboard.
func RoutineDashboard(projectRoot string, workspaceRoot string, installed *config.InstalledAgent, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	dashPath := filepath.Join(workspaceRoot, "agent", "Core", "routines.md")

	// Parse existing dashboard to preserve last_ran dates
	existing := make(map[string]string) // routine name → last_ran date
	if data, err := os.ReadFile(dashPath); err == nil {
		inDashboard := false
		for _, line := range strings.Split(string(data), "\n") {
			if strings.Contains(line, "ROUTINE_DASHBOARD_START") {
				inDashboard = true
				continue
			}
			if strings.Contains(line, "ROUTINE_DASHBOARD_END") {
				break
			}
			if !inDashboard || !strings.HasPrefix(line, "|") {
				continue
			}
			// Skip header and separator rows
			if strings.Contains(line, "---|") {
				continue
			}
			fields := strings.Split(line, "|")
			if len(fields) >= 5 {
				name := strings.TrimSpace(fields[1])
				lastRan := strings.TrimSpace(fields[3])
				if name != "" && lastRan != "" && name != "Routine" {
					existing[name] = lastRan
				}
			}
		}
	}

	var lines []string
	lines = append(lines,
		"---",
		"tags: [core, routines]",
		"description: Periodic self-maintenance routines — schedules, dashboard, execution tracking.",
		"---",
		"",
		"# Routines",
		"",
		"> [!note]",
		"> Routines are periodic maintenance tasks the agent checks at session start. The session-start hook flags overdue routines — the user decides whether to run them now or defer. Routines are **opt-in per session**, never automatic.",
		"",
		"---",
		"",
		"## How Routines Work",
		"",
		"1. **Session start:** Hook parses this file, compares `last_ran` against `frequency`, flags overdue routines.",
		"2. **User decides:** Run now, defer, or skip. Agent never runs a routine without user approval.",
		"3. **Execution:** Read the routine's definition file in `agent/Routines/`, follow the procedure step by step.",
		"4. **Log:** Append results to `Logs/RoutineLog.md` (date, routine, outcome, notes).",
		"5. **Update:** Set `last_ran` to today's date in this file.",
		"",
		"### Rules",
		"",
		"- Every routine must be **idempotent** — safe to re-run if interrupted mid-session.",
		"- When validating facts against codebase, **mark stale entries as outdated** rather than deleting — preserves audit trail.",
		"- Consolidation decisions follow four options: **keep** (still accurate), **update** (merge new info), **archive** (outdated but historically useful), **insert_new** (truly unique fact).",
		"",
		"---",
		"",
		"## Dashboard",
		"",
		"<!-- ROUTINE_DASHBOARD_START — session-start hook parses this table -->",
		"",
		"| Routine | Frequency | Last Ran | Next Due | Status |",
		"|---------|-----------|----------|----------|--------|",
	)

	for _, routineName := range installed.Routines {
		displayName := catalog.DisplayNameFrom(routineName)
		freq := ""

		routine := cat.GetRoutine(routineName)
		if routine != nil {
			displayName = routine.DisplayName
			freq = routine.Frequency
		} else if installed.CustomItems != nil {
			if meta, ok := installed.CustomItems[routineName]; ok {
				if meta.DisplayName != "" {
					displayName = meta.DisplayName
				}
				freq = meta.Frequency
			}
		}

		if freq == "" {
			continue // can't compute dashboard without frequency
		}

		lastRan := "_never_"
		nextDue := "_overdue_"
		status := "pending"

		if prev, ok := existing[displayName]; ok && prev != "_never_" {
			lastRan = prev
			// Compute nextDue from last_ran + frequency
			if t, err := time.Parse("2006-01-02", lastRan); err == nil {
				freqDays := parseFrequencyDays(freq)
				due := t.AddDate(0, 0, freqDays)
				nextDue = due.Format("2006-01-02")
				if time.Now().Before(due) {
					status = "done"
				}
			}
		}

		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
			displayName, freq, lastRan, nextDue, status))
	}

	lines = append(lines,
		"",
		"<!-- ROUTINE_DASHBOARD_END -->",
		"",
		"---",
		"",
		"## Routine Definitions",
		"",
		"| Routine | File |",
		"|---------|------|",
	)

	for _, routineName := range installed.Routines {
		displayName := catalog.DisplayNameFrom(routineName)
		if routine := cat.GetRoutine(routineName); routine != nil {
			displayName = routine.DisplayName
		} else if installed.CustomItems != nil {
			if meta, ok := installed.CustomItems[routineName]; ok && meta.DisplayName != "" {
				displayName = meta.DisplayName
			}
		}
		lines = append(lines, fmt.Sprintf("| %s | `agent/Routines/%s.md` |", displayName, routineName))
	}

	lines = append(lines, "")

	content := []byte(strings.Join(lines, "\n"))
	relPath, _ := filepath.Rel(projectRoot, dashPath)
	r := writeFile(projectRoot, relPath, content, "generated:routine-dashboard", lock, force)
	result.Add(r)
	return nil
}

// PathScopedRules generates .claude/rules/skill-{name}.md files for EVERY
// installed agent's skills that have triggers.paths defined. Used by
// `bonsai update` which regenerates all agents. `bonsai add` should call
// PathScopedRulesForAgent instead to scope regeneration to the agent being
// added and avoid flagging unrelated agents' user-edited rule files as
// conflicts. Plan 27 §B2.
//
// These auto-load when Claude reads matching files.
func PathScopedRules(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	for _, installed := range cfg.Agents {
		writePathScopedRulesForAgent(projectRoot, installed, cat, lock, result, force)
	}
	return nil
}

// PathScopedRulesForAgent is the single-agent scoped variant of
// PathScopedRules. Used by `bonsai add` so the regeneration pass only touches
// the agent being added. Plan 27 §B2 fix for the cross-agent conflict leak
// that tripped dogfood finding #7b — iterating cfg.Agents here regenerates
// rule files in unrelated workspaces, and `writeFile` → `lock.IsModified`
// then flags any user-edited rule as a conflict.
func PathScopedRulesForAgent(projectRoot string, agent *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	_ = cfg // retained for signature symmetry with the all-agents variant
	if agent == nil {
		return nil
	}
	writePathScopedRulesForAgent(projectRoot, agent, cat, lock, result, force)
	return nil
}

// writePathScopedRulesForAgent renders .claude/rules/skill-*.md for the skills
// belonging to a single installed agent. Shared helper between PathScopedRules
// and PathScopedRulesForAgent.
func writePathScopedRulesForAgent(projectRoot string, installed *config.InstalledAgent, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) {
	for _, skillName := range installed.Skills {
		item := cat.GetSkill(skillName)
		if item == nil || item.Triggers == nil || len(item.Triggers.Paths) == 0 {
			continue
		}

		// Build content
		var lines []string
		lines = append(lines, "---")
		lines = append(lines, "paths:")
		for _, p := range item.Triggers.Paths {
			lines = append(lines, fmt.Sprintf("  - \"%s\"", p))
		}
		lines = append(lines, "---", "")
		lines = append(lines, fmt.Sprintf("When working with files matching these paths, load and follow the **%s** skill at `%sagent/Skills/%s.md`.",
			item.DisplayName, installed.Workspace, skillName))

		if len(item.Triggers.Scenarios) > 0 {
			lines = append(lines, "", "Activate when:")
			for _, s := range item.Triggers.Scenarios {
				lines = append(lines, "- "+s)
			}
		}
		lines = append(lines, "")

		content := []byte(strings.Join(lines, "\n"))
		relPath := filepath.Join(installed.Workspace, ".claude", "rules", "skill-"+skillName+".md")
		r := writeFile(projectRoot, relPath, content, "generated:rule-skill-"+skillName, lock, force)
		result.Add(r)
	}
}

// WorkflowSkills generates .claude/skills/{name}/SKILL.md files for curated
// workflows across EVERY installed agent, enabling /slash-command invocation
// and description-based auto-invocation. Used by `bonsai update` which
// regenerates all agents. `bonsai add` should call WorkflowSkillsForAgent
// instead so its regeneration pass does not touch unrelated agent workspaces
// and flag user-edited SKILL.md files as conflicts. Plan 27 §B2.
func WorkflowSkills(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	for _, installed := range cfg.Agents {
		writeWorkflowSkillsForAgent(projectRoot, installed, cat, lock, result, force)
	}
	return nil
}

// WorkflowSkillsForAgent is the single-agent scoped variant of WorkflowSkills.
// Used by `bonsai add`. Plan 27 §B2 fix for the cross-agent conflict leak.
func WorkflowSkillsForAgent(projectRoot string, agent *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	_ = cfg // retained for signature symmetry with the all-agents variant
	if agent == nil {
		return nil
	}
	writeWorkflowSkillsForAgent(projectRoot, agent, cat, lock, result, force)
	return nil
}

// writeWorkflowSkillsForAgent renders .claude/skills/<workflow>/SKILL.md for
// the curated workflows belonging to a single installed agent. Shared helper
// between WorkflowSkills and WorkflowSkillsForAgent.
func writeWorkflowSkillsForAgent(projectRoot string, installed *config.InstalledAgent, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) {
	for _, wfName := range installed.Workflows {
		if !CuratedSlashWorkflows[wfName] {
			continue
		}
		item := cat.GetWorkflow(wfName)
		if item == nil {
			continue
		}

		var lines []string
		lines = append(lines, "---")
		lines = append(lines, fmt.Sprintf("name: %s", wfName))

		// Description from scenarios or fallback
		desc := item.Description
		if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
			desc = strings.Join(item.Triggers.Scenarios, ". ")
		}
		lines = append(lines, fmt.Sprintf("description: %s", desc))

		// when_to_use from scenarios
		if item.Triggers != nil && len(item.Triggers.Scenarios) > 0 {
			lines = append(lines, fmt.Sprintf("when_to_use: %s", strings.Join(item.Triggers.Scenarios, "; ")))
		}

		lines = append(lines, "---", "")
		lines = append(lines, fmt.Sprintf("Load and follow the **%s** workflow at `%sagent/Workflows/%s.md`.",
			item.DisplayName, installed.Workspace, wfName))

		// Examples section
		if item.Triggers != nil && len(item.Triggers.Examples) > 0 {
			lines = append(lines, "", "## Examples", "")
			for _, ex := range item.Triggers.Examples {
				lines = append(lines, fmt.Sprintf("> **User:** \"%s\"", ex.Prompt))
				lines = append(lines, fmt.Sprintf("> **Action:** %s", ex.Action))
				lines = append(lines, "")
			}
		}
		lines = append(lines, "")

		content := []byte(strings.Join(lines, "\n"))
		relPath := filepath.Join(installed.Workspace, ".claude", "skills", wfName, "SKILL.md")
		r := writeFile(projectRoot, relPath, content, "generated:skill-workflow-"+wfName, lock, force)
		result.Add(r)
	}
}

// triggerSection generates a markdown trigger header from catalog triggers metadata.
func triggerSection(item *catalog.CatalogItem, workspace string, category string, isSlashCommand bool) string {
	if item.Triggers == nil {
		return ""
	}
	t := item.Triggers

	var lines []string
	lines = append(lines, "## Triggers", "")

	if isSlashCommand {
		lines = append(lines, fmt.Sprintf("**Slash command:** `/%s`", item.Name))
	}

	if len(t.Paths) > 0 {
		lines = append(lines, fmt.Sprintf("**Auto-loads when reading:** %s", strings.Join(t.Paths, ", ")))
	}

	if len(t.Scenarios) > 0 {
		lines = append(lines, "**Activate when:**")
		for _, s := range t.Scenarios {
			lines = append(lines, "- "+s)
		}
	}

	if len(t.Examples) > 0 {
		lines = append(lines, "", "**Examples:**")
		for _, ex := range t.Examples {
			lines = append(lines, fmt.Sprintf("> **User:** \"%s\"", ex.Prompt))
			lines = append(lines, fmt.Sprintf("> **Action:** %s", ex.Action))
		}
	}

	lines = append(lines, "", "---", "")
	return strings.Join(lines, "\n")
}

// injectTriggerSection inserts the trigger section after YAML frontmatter
// when present, otherwise prepends it. A YAML frontmatter block is
// recognized when content starts with "---\n" and contains a closing
// "\n---\n" before any other content. Returns the original content
// unchanged when ts is empty.
func injectTriggerSection(ts string, content []byte) []byte {
	if ts == "" {
		return content
	}
	const open = "---\n"
	if bytes.HasPrefix(content, []byte(open)) {
		rest := content[len(open):]
		if idx := bytes.Index(rest, []byte("\n---\n")); idx >= 0 {
			end := len(open) + idx + len("\n---\n")
			out := make([]byte, 0, len(content)+len(ts))
			out = append(out, content[:end]...)
			out = append(out, []byte(ts)...)
			out = append(out, content[end:]...)
			return out
		}
	}
	return append([]byte(ts), content...)
}

// AgentWorkspace generates the full agent/ directory in a workspace.
func AgentWorkspace(projectRoot string, agentDef *catalog.AgentDef, installed *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	workspaceRoot := filepath.Join(projectRoot, installed.Workspace)
	agentDir := filepath.Join(workspaceRoot, "agent")
	catFS := cat.FS()

	ctx := &TemplateContext{
		ProjectName:        cfg.ProjectName,
		ProjectDescription: cfg.Description,
		AgentName:          agentDef.Name,
		AgentDisplayName:   agentDef.DisplayName,
		AgentDescription:   agentDef.Description,
	}

	for name, a := range cfg.Agents {
		if name != agentDef.Name {
			ctx.OtherAgents = append(ctx.OtherAgents, OtherAgent{
				AgentType: a.AgentType,
				Workspace: a.Workspace,
			})
		}
	}
	// Sort for deterministic template output. Map iteration order is randomised
	// per process, so without this every re-render of templates that range over
	// OtherAgents (agent identities, scope-guard, dispatch-guard) emits a
	// different byte order and the lockfile flags every file as Updated.
	sort.Slice(ctx.OtherAgents, func(i, j int) bool {
		return ctx.OtherAgents[i].AgentType < ctx.OtherAgents[j].AgentType
	})

	// 1. Core files (layered: shared defaults from catalog/core/, agent overrides from agent core/)
	coreDir := filepath.Join(agentDir, "Core")

	// Build set of agent-specific core files (for override detection)
	agentCoreFiles := make(map[string]bool)
	agentEntries, _ := fs.ReadDir(catFS, agentDef.CoreDir)
	for _, e := range agentEntries {
		if !e.IsDir() {
			agentCoreFiles[e.Name()] = true
		}
	}

	// Shared core files — use agent override if present, otherwise shared
	sharedEntries, err := fs.ReadDir(catFS, catalog.SharedCoreDir)
	if err == nil {
		for _, e := range sharedEntries {
			if e.IsDir() {
				continue
			}
			src := catalog.SharedCoreDir + "/" + e.Name()
			if agentCoreFiles[e.Name()] {
				src = agentDef.CoreDir + "/" + e.Name() // agent override
			}
			content, err := renderContent(catFS, src, ctx)
			if err != nil {
				return fmt.Errorf("core file %s: %w", e.Name(), err)
			}
			destName := strings.TrimSuffix(e.Name(), ".tmpl")
			relPath, _ := filepath.Rel(projectRoot, filepath.Join(coreDir, destName))
			r := writeFile(projectRoot, relPath, content, "catalog:core/"+destName, lock, force)
			result.Add(r)
		}
	}

	// Agent-specific core files not in shared set (e.g. identity.md.tmpl)
	for _, e := range agentEntries {
		if e.IsDir() {
			continue
		}
		// Skip files already handled by shared loop
		if sharedEntries != nil {
			found := false
			for _, se := range sharedEntries {
				if se.Name() == e.Name() {
					found = true
					break
				}
			}
			if found {
				continue
			}
		}
		src := agentDef.CoreDir + "/" + e.Name()
		content, err := renderContent(catFS, src, ctx)
		if err != nil {
			return fmt.Errorf("core file %s: %w", e.Name(), err)
		}
		destName := strings.TrimSuffix(e.Name(), ".tmpl")
		relPath, _ := filepath.Rel(projectRoot, filepath.Join(coreDir, destName))
		r := writeFile(projectRoot, relPath, content, "catalog:core/"+destName, lock, force)
		result.Add(r)
	}

	// Full template context — used by skills, sensors, and routines that have .tmpl files
	fullCtx := &TemplateContext{
		ProjectName:        cfg.ProjectName,
		ProjectDescription: cfg.Description,
		AgentName:          agentDef.Name,
		AgentDisplayName:   agentDef.DisplayName,
		AgentDescription:   agentDef.Description,
		Workspace:          installed.Workspace,
		DocsPath:           cfg.DocsPath,
		Protocols:          installed.Protocols,
		Skills:             installed.Skills,
		Workflows:          installed.Workflows,
		Routines:           installed.Routines,
		OtherAgents:        ctx.OtherAgents,
	}

	// 2. Skills (rendered through templates if .tmpl, otherwise copied as-is)
	for _, skillName := range installed.Skills {
		item := cat.GetSkill(skillName)
		if item == nil {
			continue
		}
		content, err := renderContent(catFS, item.ContentPath, fullCtx)
		if err != nil {
			return fmt.Errorf("skill %s: %w", skillName, err)
		}

		// Prepend trigger section if triggers exist
		content = injectTriggerSection(triggerSection(item, installed.Workspace, "skill", false), content)

		relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Skills", skillName+".md"))
		r := writeFile(projectRoot, relPath, content, "catalog:skills/"+skillName, lock, force)
		result.Add(r)
	}

	// 3. Workflows
	for _, wfName := range installed.Workflows {
		item := cat.GetWorkflow(wfName)
		if item == nil {
			continue
		}
		data, err := fs.ReadFile(catFS, item.ContentPath)
		if err != nil {
			return err
		}

		// Prepend trigger section if triggers exist
		data = injectTriggerSection(triggerSection(item, installed.Workspace, "workflow", CuratedSlashWorkflows[wfName]), data)

		relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Workflows", wfName+".md"))
		r := writeFile(projectRoot, relPath, data, "catalog:workflows/"+wfName, lock, force)
		result.Add(r)
	}

	// 4. Protocols
	for _, protoName := range installed.Protocols {
		item := cat.GetProtocol(protoName)
		if item == nil {
			continue
		}
		data, err := fs.ReadFile(catFS, item.ContentPath)
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Protocols", protoName+".md"))
		r := writeFile(projectRoot, relPath, data, "catalog:protocols/"+protoName, lock, force)
		result.Add(r)
	}

	// 5. Sensors (rendered through templates)
	for _, sensorName := range installed.Sensors {
		sensor := cat.GetSensor(sensorName)
		if sensor == nil {
			continue
		}
		content, err := renderContent(catFS, sensor.ContentPath, fullCtx)
		if err != nil {
			return fmt.Errorf("sensor %s: %w", sensorName, err)
		}
		destName := strings.TrimSuffix(filepath.Base(sensor.ContentPath), ".tmpl")
		relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Sensors", destName))
		r := writeFileChmod(projectRoot, relPath, content, "catalog:sensors/"+sensorName, lock, force, 0755)
		result.Add(r)
	}

	// 6. Routines (rendered through templates, like sensors)
	for _, routineName := range installed.Routines {
		routine := cat.GetRoutine(routineName)
		if routine == nil {
			continue
		}
		content, err := renderContent(catFS, routine.ContentPath, fullCtx)
		if err != nil {
			return fmt.Errorf("routine %s: %w", routineName, err)
		}
		relPath, _ := filepath.Rel(projectRoot, filepath.Join(agentDir, "Routines", routineName+".md"))
		r := writeFile(projectRoot, relPath, content, "catalog:routines/"+routineName, lock, force)
		result.Add(r)
	}

	// 7. Routine dashboard
	if len(installed.Routines) > 0 {
		if err := RoutineDashboard(projectRoot, workspaceRoot, installed, cat, lock, result, force); err != nil {
			return fmt.Errorf("routine dashboard: %w", err)
		}
	}

	// 8. Workspace CLAUDE.md
	return WorkspaceClaudeMD(projectRoot, workspaceRoot, agentDef, installed, cfg, cat, lock, result, force)
}

// buildAgentTemplateContext constructs the TemplateContext used to render
// OtherAgents-dependent files (identity.md, scope-guard-files.sh,
// dispatch-guard.sh). Extracted from AgentWorkspace so RefreshPeerAwareness
// can reuse the exact same build path — the only per-agent variation is the
// installed ability lists, which the caller passes via agentDef + installed.
//
// Deterministic OtherAgents ordering is enforced identically to AgentWorkspace
// so a render from this path produces bit-identical bytes to the full pipeline
// for the same (agentDef, installed, cfg) tuple.
func buildAgentTemplateContext(agentDef *catalog.AgentDef, installed *config.InstalledAgent, cfg *config.ProjectConfig) *TemplateContext {
	ctx := &TemplateContext{
		ProjectName:        cfg.ProjectName,
		ProjectDescription: cfg.Description,
		AgentName:          agentDef.Name,
		AgentDisplayName:   agentDef.DisplayName,
		AgentDescription:   agentDef.Description,
		Workspace:          installed.Workspace,
		DocsPath:           cfg.DocsPath,
		Protocols:          installed.Protocols,
		Skills:             installed.Skills,
		Workflows:          installed.Workflows,
		Routines:           installed.Routines,
	}
	for name, a := range cfg.Agents {
		if name != agentDef.Name {
			ctx.OtherAgents = append(ctx.OtherAgents, OtherAgent{
				AgentType: a.AgentType,
				Workspace: a.Workspace,
			})
		}
	}
	sort.Slice(ctx.OtherAgents, func(i, j int) bool {
		return ctx.OtherAgents[i].AgentType < ctx.OtherAgents[j].AgentType
	})
	return ctx
}

// hasAbility reports whether name appears in the haystack. Small helper used
// by RefreshPeerAwareness to gate per-sensor refreshes on the agent actually
// having the sensor installed.
func hasAbility(haystack []string, name string) bool {
	for _, h := range haystack {
		if h == name {
			return true
		}
	}
	return false
}

// RefreshPeerAwareness re-renders the three OtherAgents-dependent files
// (identity.md, scope-guard-files.sh, dispatch-guard.sh) for every installed
// agent EXCEPT the one specified. Called after `bonsai add` so already-
// installed peers pick up the newly-added agent in their awareness blocks
// (identity table rows, scope-guard per-peer block entries, dispatch-guard
// workspace→agent map).
//
// Lock-aware via the shared writeFile path — user-edited copies trigger the
// standard conflict resolver. Agents without scope-guard-files or
// dispatch-guard installed skip those files silently; identity.md always
// re-renders because every agent has one (shared catalog/core/ layered with
// agents/<type>/core/identity.md.tmpl overrides per AgentWorkspace).
//
// Per peer, a fresh TemplateContext is built via buildAgentTemplateContext —
// no context leaks across peers. See Plan 31 Phase A for motivation; without
// this call sibling peers' OtherAgents lists go stale after `bonsai add` and
// scope-guard silently fails open on the newly-added workspace.
func RefreshPeerAwareness(projectRoot string, excludeAgent string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, result *WriteResult, force bool) error {
	if cfg == nil || cat == nil {
		return nil
	}
	catFS := cat.FS()

	// Deterministic iteration order so test assertions and write-result diffs
	// do not depend on Go map iteration ordering.
	peerNames := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		if name == excludeAgent {
			continue
		}
		peerNames = append(peerNames, name)
	}
	sort.Strings(peerNames)

	for _, name := range peerNames {
		peer := cfg.Agents[name]
		if peer == nil {
			continue
		}
		agentDef := cat.GetAgent(peer.AgentType)
		if agentDef == nil {
			continue
		}

		ctx := buildAgentTemplateContext(agentDef, peer, cfg)
		workspaceRoot := filepath.Join(projectRoot, peer.Workspace)

		// 1. identity.md — always re-render. Resolve override (agent-specific
		//    vs shared core) identically to AgentWorkspace so we don't
		//    accidentally write a stale shared copy when the agent ships its
		//    own identity template.
		identitySrc := ""
		agentIdentity := agentDef.CoreDir + "/identity.md.tmpl"
		sharedIdentity := catalog.SharedCoreDir + "/identity.md.tmpl"
		if _, err := fs.Stat(catFS, agentIdentity); err == nil {
			identitySrc = agentIdentity
		} else if _, err := fs.Stat(catFS, sharedIdentity); err == nil {
			identitySrc = sharedIdentity
		}
		if identitySrc != "" {
			content, err := renderContent(catFS, identitySrc, ctx)
			if err != nil {
				return fmt.Errorf("refresh identity for %s: %w", name, err)
			}
			relPath, _ := filepath.Rel(projectRoot, filepath.Join(workspaceRoot, "agent", "Core", "identity.md"))
			r := writeFile(projectRoot, relPath, content, "catalog:core/identity.md", lock, force)
			result.Add(r)
		}

		// 2. scope-guard-files.sh — only if the peer has it installed.
		if hasAbility(peer.Sensors, "scope-guard-files") {
			if sensor := cat.GetSensor("scope-guard-files"); sensor != nil {
				content, err := renderContent(catFS, sensor.ContentPath, ctx)
				if err != nil {
					return fmt.Errorf("refresh scope-guard-files for %s: %w", name, err)
				}
				destName := strings.TrimSuffix(filepath.Base(sensor.ContentPath), ".tmpl")
				relPath, _ := filepath.Rel(projectRoot, filepath.Join(workspaceRoot, "agent", "Sensors", destName))
				r := writeFileChmod(projectRoot, relPath, content, "catalog:sensors/scope-guard-files", lock, force, 0755)
				result.Add(r)
			}
		}

		// 3. dispatch-guard.sh — only if the peer has it installed.
		if hasAbility(peer.Sensors, "dispatch-guard") {
			if sensor := cat.GetSensor("dispatch-guard"); sensor != nil {
				content, err := renderContent(catFS, sensor.ContentPath, ctx)
				if err != nil {
					return fmt.Errorf("refresh dispatch-guard for %s: %w", name, err)
				}
				destName := strings.TrimSuffix(filepath.Base(sensor.ContentPath), ".tmpl")
				relPath, _ := filepath.Rel(projectRoot, filepath.Join(workspaceRoot, "agent", "Sensors", destName))
				r := writeFileChmod(projectRoot, relPath, content, "catalog:sensors/dispatch-guard", lock, force, 0755)
				result.Add(r)
			}
		}
	}
	return nil
}
