// Package validate audits a Bonsai project for inconsistencies between
// .bonsai.yaml, .bonsai-lock.yaml, and the agent/ workspace files. It is
// strictly read-only — fixes happen through bonsai update.
//
// The audit runs six detection categories per installed agent. See the
// Category constants for the list. Run returns a Report containing all
// issues; callers translate that into exit codes / human / JSON output.
//
// Dependencies are intentionally limited to internal/config,
// internal/catalog, and internal/generate (only ParseFrontmatter). No TUI
// dependency — this package has to stay CI-safe so `bonsai validate
// --json` works in headless contexts.
package validate

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/catalog"
	"github.com/LastStep/Bonsai/internal/config"
	"github.com/LastStep/Bonsai/internal/generate"
)

// Severity classifies an Issue's actionability. Errors block clean state
// (exit 1); warnings flag drift the user should be aware of but doesn't
// strictly break anything (also exit 1 — any issue is non-zero).
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
)

// Category is the machine-readable issue kind. Stable identifiers — agents
// reading --json output match against these strings.
type Category string

const (
	// CategoryOrphanedRegistration: name appears in installed.<Cat>, file
	// exists, but lock entry is missing OR custom_items[name] is missing /
	// has empty Description. Plan 34's repro pattern.
	CategoryOrphanedRegistration Category = "orphaned_registration"
	// CategoryMissingFile: name appears in installed.<Cat> but the
	// expected file under agent/<Dir>/ does not exist on disk.
	CategoryMissingFile Category = "missing_file"
	// CategoryStaleLockEntry: lock.Files[relPath] references a custom
	// source but the file itself was deleted.
	CategoryStaleLockEntry Category = "stale_lock_entry"
	// CategoryUntrackedCustomFile: file under agent/<Dir>/ has valid
	// frontmatter but is not in installed.<Cat> and not in the lock.
	// User dropped the file but never ran bonsai update.
	CategoryUntrackedCustomFile Category = "untracked_custom_file"
	// CategoryInvalidFrontmatter: file under agent/<Dir>/ has missing or
	// malformed frontmatter (e.g. missing description, sensor missing
	// event, routine missing frequency).
	CategoryInvalidFrontmatter Category = "invalid_frontmatter"
	// CategoryWrongExtension: a file is the wrong extension for its
	// category dir (.md in Sensors, .sh in Skills/Workflows/Protocols/
	// Routines). Top-level only — subdirs ignored to match
	// generate.ScanCustomFiles.
	CategoryWrongExtension Category = "wrong_extension_in_category"
)

// Issue is a single audit finding. JSON tag layout is the stable contract
// downstream agents depend on; do not rename without bumping a major.
type Issue struct {
	Category    Category `json:"category"`
	Severity    Severity `json:"severity"`
	AgentName   string   `json:"agent,omitempty"`
	AbilityType string   `json:"ability_type,omitempty"` // skill|workflow|protocol|sensor|routine
	Name        string   `json:"name,omitempty"`
	Path        string   `json:"path,omitempty"`
	Detail      string   `json:"detail"`
}

// Report is the full audit result. Issues is the flat list across all
// scanned agents; AgentsScanned records the agent set the audit covered
// so callers can verify --agent filtering applied.
type Report struct {
	Issues        []Issue  `json:"issues"`
	AgentsScanned []string `json:"agents_scanned"`
}

// HasIssues reports whether at least one Issue was recorded — used by the
// command layer to drive exit code 1.
func (r *Report) HasIssues() bool { return len(r.Issues) > 0 }

// HasErrors reports whether any Issue.Severity == SeverityError. Useful
// for callers that want to distinguish hard breakage from drift warnings.
func (r *Report) HasErrors() bool {
	for _, iss := range r.Issues {
		if iss.Severity == SeverityError {
			return true
		}
	}
	return false
}

// categoryDef captures the per-ability-type wiring shared with
// generate.ScanCustomFiles. Duplicated locally rather than re-exported
// from internal/generate to keep the validate package's import surface
// small. The shape is intentionally tiny so drift is easy to spot.
type categoryDef struct {
	dir   string // workspace subdir under agent/
	ext   string // expected file extension (with leading dot)
	items func(*config.InstalledAgent) []string
}

// orderedCategories returns the ability-type wiring in a deterministic
// order. Iteration order matters for test stability — Go map iteration
// is intentionally randomised, so a slice is used instead.
func orderedCategories() []struct {
	itemType string
	def      categoryDef
} {
	return []struct {
		itemType string
		def      categoryDef
	}{
		{"skill", categoryDef{dir: "Skills", ext: ".md", items: func(ia *config.InstalledAgent) []string { return ia.Skills }}},
		{"workflow", categoryDef{dir: "Workflows", ext: ".md", items: func(ia *config.InstalledAgent) []string { return ia.Workflows }}},
		{"protocol", categoryDef{dir: "Protocols", ext: ".md", items: func(ia *config.InstalledAgent) []string { return ia.Protocols }}},
		{"sensor", categoryDef{dir: "Sensors", ext: ".sh", items: func(ia *config.InstalledAgent) []string { return ia.Sensors }}},
		{"routine", categoryDef{dir: "Routines", ext: ".md", items: func(ia *config.InstalledAgent) []string { return ia.Routines }}},
	}
}

// wrongExtFor returns the "wrong" extension for a category dir — the one
// that should never appear there. Used by the wrong-extension scan.
// Empty string means no specific wrong extension to flag.
func wrongExtFor(itemType string) string {
	switch itemType {
	case "sensor":
		return ".md" // sensors are .sh; .md here is misplaced
	case "skill", "workflow", "protocol", "routine":
		return ".sh" // these are .md; .sh here is misplaced
	}
	return ""
}

// Run audits the project at projectRoot. cfg + cat + lock should already
// be loaded by the caller (cmd/validate.go calls config.Load,
// catalog.New, and config.LoadLockFile). When agentFilter is non-empty
// only that agent is scanned; an unknown agent name returns an error so
// callers can surface a clear message.
func Run(projectRoot string, cfg *config.ProjectConfig, cat *catalog.Catalog, lock *config.LockFile, agentFilter string) (*Report, error) {
	_ = cat // catalog isn't required for any current detection — kept in
	// the signature so future categories (e.g. catalog drift) can use it
	// without breaking the public Run signature.

	if cfg == nil {
		return nil, errors.New("validate: nil project config")
	}
	if lock == nil {
		// LoadLockFile already returns an empty lock when the file is
		// absent; defend against direct callers that pass nil.
		lock = config.NewLockFile()
	}

	report := &Report{Issues: []Issue{}, AgentsScanned: []string{}}

	// Build deterministic agent list (filter to one if requested).
	names := make([]string, 0, len(cfg.Agents))
	for n := range cfg.Agents {
		names = append(names, n)
	}
	sort.Strings(names)

	if agentFilter != "" {
		if _, ok := cfg.Agents[agentFilter]; !ok {
			return nil, fmt.Errorf("validate: agent %q not installed", agentFilter)
		}
		names = []string{agentFilter}
	}

	for _, agentName := range names {
		installed := cfg.Agents[agentName]
		if installed == nil {
			continue
		}
		report.AgentsScanned = append(report.AgentsScanned, agentName)
		auditAgent(projectRoot, agentName, installed, lock, report)
	}

	return report, nil
}

// auditAgent runs the six detection categories against a single agent
// and appends any issues to report.Issues. Operates only on top-level
// files under agent/<Dir>/ — subdirs are ignored, matching ScanCustomFiles.
func auditAgent(projectRoot, agentName string, installed *config.InstalledAgent, lock *config.LockFile, report *Report) {
	workspaceRoot := filepath.Join(projectRoot, installed.Workspace)
	agentDir := filepath.Join(workspaceRoot, "agent")

	cats := orderedCategories()

	// Track per-category seen-on-disk paths so the stale-lock pass can
	// distinguish "file gone" from "file present, just in another category".
	for _, c := range cats {
		dirPath := filepath.Join(agentDir, c.def.dir)
		items := c.def.items(installed)

		// Sort installed names so issue output is deterministic — random
		// map order from CustomItems shouldn't leak into the report.
		sortedItems := append([]string(nil), items...)
		sort.Strings(sortedItems)

		// 1. Orphan + missing checks for each name in installed.<Cat>.
		for _, name := range sortedItems {
			rel := relPath(projectRoot, dirPath, name+c.def.ext)
			abs := filepath.Join(dirPath, name+c.def.ext)

			fi, err := os.Stat(abs)
			fileExists := err == nil && !fi.IsDir()

			if !fileExists {
				report.Issues = append(report.Issues, Issue{
					Category:    CategoryMissingFile,
					Severity:    SeverityError,
					AgentName:   agentName,
					AbilityType: c.itemType,
					Name:        name,
					Path:        rel,
					Detail:      fmt.Sprintf("registered in installed.%s but file does not exist on disk", strings.ToLower(c.def.dir)),
				})
				continue
			}

			// File exists — check tracking.
			_, tracked := lock.Files[rel]
			meta, hasMeta := installed.CustomItems[name]
			isCustomMissing := !tracked || !hasMeta || meta == nil || meta.Description == ""
			// Catalog-shipped items have lock entries with Source like
			// "skills/foo" (not "custom:..."). Custom items are tagged
			// "custom:<type>s/<name>". An orphan only applies to custom
			// items that should have lock + custom_items both populated.
			// Heuristic: if the lock entry exists but is non-custom, the
			// item is catalog-tracked and orphan logic doesn't apply
			// (catalog drift is a separate, out-of-scope concern).
			if tracked {
				entry := lock.Files[rel]
				if entry != nil && !strings.HasPrefix(entry.Source, "custom:") {
					// Catalog-tracked. Skip orphan check — even an empty
					// custom_items entry is expected for catalog items.
					continue
				}
			}
			if isCustomMissing {
				detail := "registered but lock entry missing"
				if tracked && (!hasMeta || meta == nil || meta.Description == "") {
					detail = "registered but custom_items[name] missing or has empty description"
				}
				report.Issues = append(report.Issues, Issue{
					Category:    CategoryOrphanedRegistration,
					Severity:    SeverityError,
					AgentName:   agentName,
					AbilityType: c.itemType,
					Name:        name,
					Path:        rel,
					Detail:      detail + " — run `bonsai update` to recover",
				})
			}
		}

		// 2. Disk scan: untracked custom files, invalid frontmatter, wrong extension.
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			// directory might not exist — that's fine, just skip
			continue
		}

		known := make(map[string]bool, len(items))
		for _, n := range items {
			known[n] = true
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue // top-level only
			}
			fname := entry.Name()
			ext := filepath.Ext(fname)
			rel := relPath(projectRoot, dirPath, fname)
			abs := filepath.Join(dirPath, fname)

			// Wrong-extension check fires regardless of tracking — a
			// stray .md in Sensors/ is always misplaced.
			if wrong := wrongExtFor(c.itemType); wrong != "" && ext == wrong {
				report.Issues = append(report.Issues, Issue{
					Category:    CategoryWrongExtension,
					Severity:    SeverityWarning,
					AgentName:   agentName,
					AbilityType: c.itemType,
					Name:        strings.TrimSuffix(fname, ext),
					Path:        rel,
					Detail:      fmt.Sprintf("%s file in %s/ — expected %s", ext, c.def.dir, c.def.ext),
				})
				continue
			}

			if ext != c.def.ext {
				continue // not a candidate for this category — ignore
			}

			name := strings.TrimSuffix(fname, ext)
			_, tracked := lock.Files[rel]
			inInstalled := known[name]

			// Fully tracked — already covered by the orphan loop above.
			// Don't re-flag.
			if inInstalled && tracked {
				continue
			}
			// Orphaned (in installed, not in lock) — already flagged in
			// the orphan loop. Don't re-flag here.
			if inInstalled {
				continue
			}

			// Untracked candidate. Read + parse frontmatter to decide
			// between invalid_frontmatter (error) and untracked_custom_file
			// (warning).
			data, readErr := os.ReadFile(abs)
			if readErr != nil {
				// Read failure is rare — treat as invalid since we
				// can't validate it.
				report.Issues = append(report.Issues, Issue{
					Category:    CategoryInvalidFrontmatter,
					Severity:    SeverityError,
					AgentName:   agentName,
					AbilityType: c.itemType,
					Name:        name,
					Path:        rel,
					Detail:      "could not read file: " + readErr.Error(),
				})
				continue
			}

			meta, _ := generate.ParseFrontmatter(data)
			problem := frontmatterProblem(c.itemType, meta)
			if problem != "" {
				report.Issues = append(report.Issues, Issue{
					Category:    CategoryInvalidFrontmatter,
					Severity:    SeverityError,
					AgentName:   agentName,
					AbilityType: c.itemType,
					Name:        name,
					Path:        rel,
					Detail:      problem,
				})
				continue
			}

			// Valid frontmatter, not registered — warn.
			report.Issues = append(report.Issues, Issue{
				Category:    CategoryUntrackedCustomFile,
				Severity:    SeverityWarning,
				AgentName:   agentName,
				AbilityType: c.itemType,
				Name:        name,
				Path:        rel,
				Detail:      "valid custom file present but not registered — run `bonsai update`",
			})
		}
	}

	// 3. Stale lock check: lock entries that reference a custom file
	// belonging to this agent's workspace, where the file is gone.
	auditStaleLockEntries(projectRoot, agentName, installed, lock, report)
}

// auditStaleLockEntries walks lock.Files and flags any custom: entry
// whose path is under this agent's workspace but no longer exists. The
// agent-scoping prevents a stale entry from being reported once per
// scanned agent in a multi-agent project.
func auditStaleLockEntries(projectRoot, agentName string, installed *config.InstalledAgent, lock *config.LockFile, report *Report) {
	// Normalise the workspace prefix once. Use forward slashes since
	// lockfile paths are stored in forward-slash form (filepath.Rel on
	// POSIX produces these; the lock format does the same on Windows
	// elsewhere in the codebase).
	wsPrefix := filepath.ToSlash(filepath.Join(installed.Workspace, "agent")) + "/"

	// Sort relPaths for deterministic output.
	relPaths := make([]string, 0, len(lock.Files))
	for p := range lock.Files {
		relPaths = append(relPaths, p)
	}
	sort.Strings(relPaths)

	for _, rel := range relPaths {
		entry := lock.Files[rel]
		if entry == nil {
			continue
		}
		if !strings.HasPrefix(entry.Source, "custom:") {
			continue // catalog-tracked entries are out of scope
		}
		if !strings.HasPrefix(filepath.ToSlash(rel), wsPrefix) {
			continue // not in this agent's workspace
		}
		abs := filepath.Join(projectRoot, rel)
		if _, err := os.Stat(abs); err == nil {
			continue
		}
		// File is gone — derive the type/name from the source string for
		// cleaner reporting. Source format: "custom:<type>s/<name>".
		abilityType, name := parseCustomSource(entry.Source)
		report.Issues = append(report.Issues, Issue{
			Category:    CategoryStaleLockEntry,
			Severity:    SeverityWarning,
			AgentName:   agentName,
			AbilityType: abilityType,
			Name:        name,
			Path:        rel,
			Detail:      "lock entry references custom file but file is missing — run `bonsai update`",
		})
	}
}

// parseCustomSource extracts (type, name) from a "custom:<type>s/<name>"
// source string. Returns ("", "") if the format doesn't match — defensive
// against future source-string changes.
func parseCustomSource(src string) (string, string) {
	rest := strings.TrimPrefix(src, "custom:")
	slash := strings.Index(rest, "/")
	if slash < 0 {
		return "", ""
	}
	plural := rest[:slash]
	name := rest[slash+1:]
	// Strip trailing "s" to get singular type — matches updateflow's
	// "custom:"+d.Type+"s/" format.
	abilityType := strings.TrimSuffix(plural, "s")
	return abilityType, name
}

// frontmatterProblem returns a non-empty description of the frontmatter
// failure for an ability of the given type, or "" if the meta is valid.
// Mirrors the validation rules in generate.ScanCustomFiles so untracked
// files are flagged consistently between scan and validate.
func frontmatterProblem(itemType string, meta *config.CustomItemMeta) string {
	if meta == nil {
		return "missing or unparseable frontmatter"
	}
	if meta.Description == "" {
		return "missing required frontmatter field: description"
	}
	switch itemType {
	case "sensor":
		if meta.Event == "" {
			return "sensor missing required frontmatter field: event"
		}
	case "routine":
		if meta.Frequency == "" {
			return "routine missing required frontmatter field: frequency"
		}
	}
	return ""
}

// relPath returns the project-relative slash-form path for a file at
// dirPath/name. Errors collapse to the joined path so the report can
// still cite something — RelPath shouldn't fail here in practice since
// projectRoot is the same root that produced dirPath.
func relPath(projectRoot, dirPath, name string) string {
	abs := filepath.Join(dirPath, name)
	rel, err := filepath.Rel(projectRoot, abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}
