package generate

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

func copyOrRender(fsys fs.FS, srcPath, destPath string, ctx interface{}) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	if strings.HasSuffix(srcPath, ".tmpl") {
		content, err := renderTemplate(fsys, srcPath, ctx)
		if err != nil {
			return err
		}
		outPath := strings.TrimSuffix(destPath, ".tmpl")
		return os.WriteFile(outPath, []byte(content), 0644)
	}
	data, err := fs.ReadFile(fsys, srcPath)
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
}

func descFor(names []string, cat *catalog.Catalog, category string) map[string]string {
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
		if desc == "" {
			desc = titleCase(strings.ReplaceAll(name, "-", " "))
		}
		result[name] = desc
	}
	return result
}

// Scaffolding generates project management infrastructure files.
// Returns a list of created file paths relative to projectRoot.
func Scaffolding(projectRoot string, cfg *config.ProjectConfig, catFS fs.FS) ([]string, error) {
	docsRoot := projectRoot
	if cfg.DocsPath != "" {
		docsRoot = filepath.Join(projectRoot, cfg.DocsPath)
	}
	ctx := &TemplateContext{
		ProjectName:        cfg.ProjectName,
		ProjectDescription: cfg.Description,
	}

	var created []string

	err := fs.WalkDir(catFS, "scaffolding", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel := strings.TrimPrefix(path, "scaffolding/")
		dest := filepath.Join(docsRoot, rel)

		finalDest := dest
		if strings.HasSuffix(finalDest, ".tmpl") {
			finalDest = strings.TrimSuffix(finalDest, ".tmpl")
		}
		if _, statErr := os.Stat(finalDest); statErr == nil {
			return nil // don't overwrite
		}

		if err := copyOrRender(catFS, path, dest, ctx); err != nil {
			return err
		}

		relToProject, _ := filepath.Rel(projectRoot, finalDest)
		created = append(created, relToProject)
		return nil
	})

	return created, err
}

// SettingsJSON generates or updates .claude/settings.json with sensor hooks.
func SettingsJSON(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog) error {
	settingsPath := filepath.Join(projectRoot, ".claude", "settings.json")

	existing := make(map[string]interface{})
	if data, err := os.ReadFile(settingsPath); err == nil {
		_ = json.Unmarshal(data, &existing)
	}

	type hookEntry struct {
		Type    string `json:"type"`
		Command string `json:"command"`
	}
	type hookGroup struct {
		Hooks   []hookEntry `json:"hooks"`
		Matcher string      `json:"matcher,omitempty"`
	}

	type groupKey struct{ event, matcher string }
	groups := make(map[groupKey][]string)

	for _, installed := range cfg.Agents {
		for _, sensorName := range installed.Sensors {
			sensor := cat.GetSensor(sensorName)
			if sensor == nil {
				continue
			}
			k := groupKey{sensor.Event, sensor.Matcher}
			scriptPath := installed.Workspace + "agent/Sensors/" + sensorName + ".sh"
			groups[k] = append(groups[k], "bash "+scriptPath)
		}
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
	return os.WriteFile(settingsPath, append(data, '\n'), 0644)
}

// RootClaudeMD generates the root CLAUDE.md routing file.
func RootClaudeMD(projectRoot string, cfg *config.ProjectConfig) error {
	docsPrefix := cfg.DocsPath

	var lines []string
	lines = append(lines, fmt.Sprintf("# %s — Project Router", cfg.ProjectName), "")

	if len(cfg.Agents) > 0 {
		lines = append(lines,
			"## Routing", "",
			"| Working in | Read | Do NOT read |",
			"|------------|------|-------------|",
		)
		for name, agent := range cfg.Agents {
			read := fmt.Sprintf("`%sCLAUDE.md`", agent.Workspace)
			var doNotRead []string
			for otherName, other := range cfg.Agents {
				if otherName != name {
					doNotRead = append(doNotRead, fmt.Sprintf("`%sCLAUDE.md`", other.Workspace))
				}
			}
			dnr := "—"
			if len(doNotRead) > 0 {
				dnr = strings.Join(doNotRead, ", ")
			}
			lines = append(lines, fmt.Sprintf("| `%s` | %s | %s |", agent.Workspace, read, dnr))
		}
		lines = append(lines, "",
			"> Read ONLY the CLAUDE.md for your workspace. Each workspace has its own agent/ directory.", "")
	}

	lines = append(lines,
		"## Universal Rules", "",
		"- **Never touch another workspace's files** — stay in your lane",
		fmt.Sprintf("- **Plans live in `%sPlaybook/Plans/`** — read your assigned plan before writing code", docsPrefix),
		fmt.Sprintf("- **Security rules live in `%sPlaybook/Standards/SecurityStandards.md`** — read every session", docsPrefix),
		fmt.Sprintf("- **Logs go to `%sLogs/`** — write a log after completing any plan", docsPrefix),
		"- **Attribution required** — anything written under the user's name must end with:",
		"  ```",
		"  ---",
		"  Written by **[Agent Name]** · Initiated by [source]",
		"  ```", "",
		"## Triggers", "",
		"| Trigger | Action |",
		"|---------|--------|",
		fmt.Sprintf("| `status` | Read `%sPlaybook/Status.md` and show current In Progress / Pending |", docsPrefix),
		"| `verify` | Run the verification suite for the current workspace |", "",
	)

	return os.WriteFile(filepath.Join(projectRoot, "CLAUDE.md"), []byte(strings.Join(lines, "\n")), 0644)
}

// WorkspaceClaudeMD generates the workspace CLAUDE.md with navigation tables.
func WorkspaceClaudeMD(workspaceRoot string, agentDef *catalog.AgentDef, installed *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog) error {
	docsPrefix := cfg.DocsPath

	var lines []string
	lines = append(lines,
		fmt.Sprintf("# %s — %s", cfg.ProjectName, agentDef.DisplayName), "",
		fmt.Sprintf("**Working directory:** `%s`", installed.Workspace), "",
		"> [!warning]",
		"> **FIRST:** Read `agent/Core/identity.md`, then `agent/Core/memory.md`.", "",
		"---", "",
		"## Navigation", "",
		"> All agent instruction files live in `agent/`.", "",
		"### Core (load first, every session)", "",
		"| File | Purpose |",
		"|------|---------|",
		"| `agent/Core/identity.md` | Who I am, relationships, mindset |",
		"| `agent/Core/memory.md` | Working memory — flags, work state, notes |",
		"| `agent/Core/self-awareness.md` | Context monitoring, hard thresholds |", "",
	)

	if len(installed.Protocols) > 0 {
		protoDescs := descFor(installed.Protocols, cat, "protocol")
		lines = append(lines,
			"### Protocols (load after Core, every session)", "",
			"| File | Purpose |",
			"|------|---------|",
		)
		for _, p := range installed.Protocols {
			lines = append(lines, fmt.Sprintf("| `agent/Protocols/%s.md` | %s |", p, protoDescs[p]))
		}
		lines = append(lines, "")
	}

	if len(installed.Workflows) > 0 {
		wfDescs := descFor(installed.Workflows, cat, "workflow")
		lines = append(lines,
			"### Workflows (load when starting an activity)", "",
			"| Activity | Read this |",
			"|----------|-----------|",
		)
		for _, w := range installed.Workflows {
			lines = append(lines, fmt.Sprintf("| %s | `agent/Workflows/%s.md` |", wfDescs[w], w))
		}
		lines = append(lines, "")
	}

	if len(installed.Skills) > 0 {
		skillDescs := descFor(installed.Skills, cat, "skill")
		lines = append(lines,
			"### Skills (load when doing specific work)", "",
			"| Need | Read this |",
			"|------|-----------|",
		)
		for _, s := range installed.Skills {
			lines = append(lines, fmt.Sprintf("| %s | `agent/Skills/%s.md` |", skillDescs[s], s))
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
			if routine := cat.GetRoutine(r); routine != nil {
				freq = routine.Frequency
			}
			lines = append(lines, fmt.Sprintf("| %s | %s | `agent/Routines/%s.md` |",
				titleCase(strings.ReplaceAll(r, "-", " ")), freq, r))
		}
		lines = append(lines, "",
			"> Routines are opt-in — check `agent/Core/routines.md` for the dashboard and procedures.", "")
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
				lines = append(lines, fmt.Sprintf("| `agent/Sensors/%s.sh` | %s | %s |", sensorName, eventStr, sensor.Description))
			}
		}
		lines = append(lines, "",
			"> Sensors run automatically — they are configured in `.claude/settings.json`.", "")
	}

	lines = append(lines,
		"---", "",
		"## Memory", "",
		"> [!warning]",
		"> **Do NOT use Claude Code's auto-memory system** (`~/.claude/projects/*/memory/`). All persistent memory goes in `agent/Core/memory.md` — version-controlled, auditable, inside the project.", "",
		"When you would normally write to auto-memory (feedback, references, project context, flags), write to the appropriate section in `agent/Core/memory.md` instead.", "",
		"---", "",
		"### External References", "",
		"| Need | Read this |",
		"|------|-----------|",
		fmt.Sprintf("| Project snapshot | `%sINDEX.md` |", docsPrefix),
		fmt.Sprintf("| Current work status | `%sPlaybook/Status.md` |", docsPrefix),
		fmt.Sprintf("| Long-term direction | `%sPlaybook/Roadmap.md` |", docsPrefix),
		fmt.Sprintf("| Security standards | `%sPlaybook/Standards/SecurityStandards.md` |", docsPrefix),
		fmt.Sprintf("| Your assigned plan | `%sPlaybook/Plans/Active/` |", docsPrefix),
		fmt.Sprintf("| Prior decisions | `%sLogs/KeyDecisionLog.md` |", docsPrefix),
		fmt.Sprintf("| Submit report | `%sReports/Pending/` |", docsPrefix), "",
	)

	return os.WriteFile(filepath.Join(workspaceRoot, "CLAUDE.md"), []byte(strings.Join(lines, "\n")), 0644)
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
// This is a managed file — rebuilt when routines are added/removed but preserves last_ran dates.
func RoutineDashboard(workspaceRoot string, installed *config.InstalledAgent, cat *catalog.Catalog) error {
	dashPath := filepath.Join(workspaceRoot, "agent", "Core", "routines.md")

	// Parse existing dashboard to preserve last_ran dates
	existing := make(map[string]string) // routine name → last_ran date
	if data, err := os.ReadFile(dashPath); err == nil {
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(line, "|") || strings.Contains(line, "Routine") || strings.Contains(line, "---") {
				continue
			}
			fields := strings.Split(line, "|")
			if len(fields) >= 5 {
				name := strings.TrimSpace(fields[1])
				lastRan := strings.TrimSpace(fields[3])
				if name != "" && lastRan != "" {
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
		routine := cat.GetRoutine(routineName)
		if routine == nil {
			continue
		}
		displayName := titleCase(strings.ReplaceAll(routineName, "-", " "))
		lastRan := "_never_"
		nextDue := "_overdue_"
		status := "pending"

		if prev, ok := existing[displayName]; ok && prev != "_never_" {
			lastRan = prev
			// Compute nextDue from last_ran + frequency
			if t, err := time.Parse("2006-01-02", lastRan); err == nil {
				freqDays := parseFrequencyDays(routine.Frequency)
				due := t.AddDate(0, 0, freqDays)
				nextDue = due.Format("2006-01-02")
				if time.Now().Before(due) {
					status = "done"
				}
			}
		}

		lines = append(lines, fmt.Sprintf("| %s | %s | %s | %s | %s |",
			displayName, routine.Frequency, lastRan, nextDue, status))
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
		displayName := titleCase(strings.ReplaceAll(routineName, "-", " "))
		lines = append(lines, fmt.Sprintf("| %s | `agent/Routines/%s.md` |", displayName, routineName))
	}

	lines = append(lines, "")

	if err := os.MkdirAll(filepath.Dir(dashPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(dashPath, []byte(strings.Join(lines, "\n")), 0644)
}

// AgentWorkspace generates the full agent/ directory in a workspace.
func AgentWorkspace(projectRoot string, agentDef *catalog.AgentDef, installed *config.InstalledAgent, cfg *config.ProjectConfig, cat *catalog.Catalog) error {
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

	// 1. Core files
	coreDir := filepath.Join(agentDir, "Core")
	entries, err := fs.ReadDir(catFS, agentDef.CoreDir)
	if err == nil {
		for _, e := range entries {
			if !e.IsDir() {
				src := agentDef.CoreDir + "/" + e.Name()
				dest := filepath.Join(coreDir, e.Name())
				if err := copyOrRender(catFS, src, dest, ctx); err != nil {
					return fmt.Errorf("core file %s: %w", e.Name(), err)
				}
			}
		}
	}

	// 2. Skills
	for _, skillName := range installed.Skills {
		item := cat.GetSkill(skillName)
		if item == nil {
			continue
		}
		dest := filepath.Join(agentDir, "Skills", skillName+".md")
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		data, err := fs.ReadFile(catFS, item.ContentPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
	}

	// 3. Workflows
	for _, wfName := range installed.Workflows {
		item := cat.GetWorkflow(wfName)
		if item == nil {
			continue
		}
		dest := filepath.Join(agentDir, "Workflows", wfName+".md")
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		data, err := fs.ReadFile(catFS, item.ContentPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
	}

	// 4. Protocols
	for _, protoName := range installed.Protocols {
		item := cat.GetProtocol(protoName)
		if item == nil {
			continue
		}
		dest := filepath.Join(agentDir, "Protocols", protoName+".md")
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return err
		}
		data, err := fs.ReadFile(catFS, item.ContentPath)
		if err != nil {
			return err
		}
		if err := os.WriteFile(dest, data, 0644); err != nil {
			return err
		}
	}

	// 5. Sensors (rendered through templates)
	sensorCtx := &TemplateContext{
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

	for _, sensorName := range installed.Sensors {
		sensor := cat.GetSensor(sensorName)
		if sensor == nil {
			continue
		}
		destName := filepath.Base(sensor.ContentPath)
		dest := filepath.Join(agentDir, "Sensors", destName)
		if err := copyOrRender(catFS, sensor.ContentPath, dest, sensorCtx); err != nil {
			return fmt.Errorf("sensor %s: %w", sensorName, err)
		}
		finalDest := dest
		if strings.HasSuffix(finalDest, ".tmpl") {
			finalDest = strings.TrimSuffix(finalDest, ".tmpl")
		}
		_ = os.Chmod(finalDest, 0755)
	}

	// 6. Routines (rendered through templates, like sensors)
	for _, routineName := range installed.Routines {
		routine := cat.GetRoutine(routineName)
		if routine == nil {
			continue
		}
		destName := routineName + ".md"
		if strings.HasSuffix(routine.ContentPath, ".tmpl") {
			// Will be rendered, output drops .tmpl
			dest := filepath.Join(agentDir, "Routines", routineName+".md.tmpl")
			if err := copyOrRender(catFS, routine.ContentPath, dest, sensorCtx); err != nil {
				return fmt.Errorf("routine %s: %w", routineName, err)
			}
		} else {
			dest := filepath.Join(agentDir, "Routines", destName)
			if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
				return err
			}
			data, err := fs.ReadFile(catFS, routine.ContentPath)
			if err != nil {
				return err
			}
			if err := os.WriteFile(dest, data, 0644); err != nil {
				return err
			}
		}
	}

	// 7. Routine dashboard (managed file — rebuilt on every generate)
	if len(installed.Routines) > 0 {
		if err := RoutineDashboard(workspaceRoot, installed, cat); err != nil {
			return fmt.Errorf("routine dashboard: %w", err)
		}
	}

	// 8. Workspace CLAUDE.md
	return WorkspaceClaudeMD(workspaceRoot, agentDef, installed, cfg, cat)
}
