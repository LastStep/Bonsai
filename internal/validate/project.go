package validate

// Project-level validation pass (Plan 40 Phase 2).
//
// This pass lints two repo-resident formats that the downstream hub
// consumes: the project manifest at the repo root (.bonsai/project.yaml)
// and the in-repo memory note tree under {memory_dir} (default
// station/Memory/). It is strictly read-only — never writes or mutates a
// single byte — and runs regardless of any --agent filter because these
// findings are project-scoped, not agent-scoped. Every Issue it appends
// carries an empty AgentName and the pass never touches AgentsScanned.
//
// Security posture (frozen by the plan's Input-validation row):
//   - manifest memory_dir traversal is accidental-grade — reuse
//     wsvalidate.InvalidReason after trimming Normalise's trailing slash.
//   - note-target resolution is adversarial-grade — every note file and
//     every directory inside the tree is resolved via filepath.EvalSymlinks
//     and checked to remain under the resolved memory_dir; symlinks that
//     escape are rejected (never read, never indexed). Relations and
//     superseded_by resolve by sanitized permalink against a trivial
//     in-memory map — link text is NEVER treated as a filesystem path.
//
// Parsing uses typed structs + yaml.Unmarshal only (no eval) and the walk
// is bounded on both file count and per-file size so a hostile or
// accidentally-huge tree cannot OOM or hang the audit.

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/LastStep/Bonsai/internal/wsvalidate"
	"gopkg.in/yaml.v3"
)

const (
	// manifestRelPath is the repo-relative location of the project manifest.
	manifestRelPath = ".bonsai/project.yaml"
	// defaultMemoryDir is the manifest's default memory_dir (frozen v1).
	defaultMemoryDir = "station/Memory"
	// memoryIndexName is the index file whose line budget we enforce. It
	// lives one level above memory_dir (scaffolded under docs_path next to
	// the Memory/ tree).
	memoryIndexName = "MEMORY.md"
	// memoryIndexMaxLines is the frozen MEMORY.md budget — over this is a
	// warning, not an error.
	memoryIndexMaxLines = 200

	// maxNoteFiles bounds the recursive walk so a hostile tree with an
	// enormous number of files cannot make the audit run unbounded. When the
	// cap is hit the walk stops and a warning is emitted; the cap is far
	// above any realistic hand-authored memory graph.
	maxNoteFiles = 10000
	// maxNoteSizeBytes bounds the bytes read from any single note. A note is
	// a small markdown file; anything larger is either a mistake or hostile,
	// so it is flagged (invalid_note) and its frontmatter is not parsed.
	maxNoteSizeBytes = 1 << 20 // 1 MiB
)

// slugCharset is the frozen [a-z0-9-] rule shared by slug and permalink.
var slugCharset = regexp.MustCompile(`^[a-z0-9-]+$`)

// dateCharset matches a YYYY-MM-DD calendar date shape (content not range-checked).
var dateCharset = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// wikilinkPattern extracts [[target]] link bodies from a note's Relations
// section. The captured text is treated as a permalink reference and is
// sanitized + looked up in the in-memory index — it is NEVER used as a path.
var wikilinkPattern = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

// noteTypes is the frozen set of legal note `type` values.
var noteTypes = map[string]bool{"decision": true, "note": true, "fact": true, "log": true}

// manifestStatuses is the frozen set of legal manifest `status` values.
var manifestStatuses = map[string]bool{"idea": true, "active": true, "paused": true, "done": true, "archived": true}

// projectManifest mirrors the frozen v1 manifest schema (.bonsai/project.yaml).
// Typed struct only — we never extend config.CustomItemMeta. Unknown keys in
// `links` (and elsewhere) are ignored by yaml.Unmarshal rather than erroring,
// matching the frozen "unknown keys ignored" rule. SchemaVersion is a pointer
// so a missing key is distinguishable from an explicit 0.
type projectManifest struct {
	SchemaVersion *int              `yaml:"schema_version"`
	Name          string            `yaml:"name"`
	Slug          string            `yaml:"slug"`
	Status        string            `yaml:"status"`
	Tags          []string          `yaml:"tags"`
	Description   string            `yaml:"description"`
	Links         map[string]string `yaml:"links"`
	Created       string            `yaml:"created"`
	MemoryDir     string            `yaml:"memory_dir"`

	// memoryDirInvalid is set by loadManifest when an explicit memory_dir
	// failed validation (traversal/absolute). It is unexported and untagged
	// so yaml.Unmarshal never populates it. When true, auditProject skips the
	// memory-tree walk entirely instead of defaulting to station/Memory — the
	// manifest is broken and walking a sanitized/default tree would mislead.
	memoryDirInvalid bool
}

// noteFrontmatter mirrors the frozen v1 memory-note frontmatter. Pointer
// fields capture present-vs-absent where the distinction matters:
// SchemaVersion (absent is a missing-required-field error, not a 0-value
// error) and SupersededBy (absent ≡ null ≡ not-superseded — no missing-key
// error; only a non-null value is resolved). Typed struct only.
type noteFrontmatter struct {
	SchemaVersion *int     `yaml:"schema_version"`
	Title         string   `yaml:"title"`
	Type          string   `yaml:"type"`
	Permalink     string   `yaml:"permalink"`
	Tags          []string `yaml:"tags"`
	Scope         string   `yaml:"scope"`
	ValidFrom     string   `yaml:"valid_from"`
	SupersededBy  *string  `yaml:"superseded_by"`
}

// noteRecord is a parsed, in-tree note retained for the relation-resolution
// second pass. Only notes that survived frontmatter linting (and so have a
// usable permalink) feed the index; relations are resolved against that index.
type noteRecord struct {
	relPath      string
	permalink    string
	supersededBy string   // sanitized; "" when absent/null
	relations    []string // sanitized [[target]] bodies
}

// auditProject runs the manifest + memory-tree lint and appends any Issues
// to report. projectRoot is the repo root (the same root auditAgent uses).
func auditProject(projectRoot string, report *Report) {
	manifest, manifestPresent := loadManifest(projectRoot, report)

	// Resolve memory_dir. When the manifest carries a non-empty, *validated*
	// memory_dir we honour it; when the key is simply absent we fall back to
	// the frozen default so an existing tree is still linted. But when the
	// manifest carried an explicit memory_dir that FAILED validation
	// (traversal/absolute), loadManifest has already recorded the error and
	// blanked the field — we must NOT walk anything for this run (neither the
	// out-of-tree target nor the default), so we bail before the stat/walk.
	memDirRel := defaultMemoryDir
	slug := ""
	slugKnown := false
	if manifestPresent && manifest != nil {
		if manifest.memoryDirInvalid {
			return
		}
		if manifest.MemoryDir != "" {
			memDirRel = manifest.MemoryDir
		}
		if manifest.Slug != "" {
			slug = manifest.Slug
			slugKnown = true
		}
	}

	memDirAbs := filepath.Join(projectRoot, filepath.FromSlash(memDirRel))
	memInfo, statErr := os.Stat(memDirAbs)
	memoryPresent := statErr == nil && memInfo.IsDir()

	// Manifest absent but a memory tree present → scope is unverifiable.
	// Warn, skip the scope-match check (slugKnown stays false), but still
	// lint the rest of the frontmatter below.
	if memoryPresent && !manifestPresent {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryMissingManifest,
			Severity: SeverityWarning,
			Path:     manifestRelPath,
			Detail:   "memory tree present but " + manifestRelPath + " is missing — note scope cannot be verified; run `bonsai add` to install the project-manifest item",
		})
	}

	if memoryPresent {
		auditMemoryTree(projectRoot, memDirRel, memDirAbs, slug, slugKnown, report)
	}

	// MEMORY.md line budget. The index sits one level above the memory tree
	// (scaffolded under docs_path beside Memory/). Checked independently of
	// whether the tree itself has notes.
	auditMemoryIndex(projectRoot, memDirRel, report)
}

// loadManifest reads + lints .bonsai/project.yaml. Returns (manifest, true)
// when the file exists (manifest may still have carried errors — they are
// appended to report). Returns (nil, false) when the file is simply absent
// (not an error on its own — only a warning if a memory tree exists, handled
// by the caller).
func loadManifest(projectRoot string, report *Report) (*projectManifest, bool) {
	abs := filepath.Join(projectRoot, filepath.FromSlash(manifestRelPath))
	data, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false
		}
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidManifest,
			Severity: SeverityError,
			Path:     manifestRelPath,
			Detail:   "could not read manifest: " + err.Error(),
		})
		return nil, true
	}

	var m projectManifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidManifest,
			Severity: SeverityError,
			Path:     manifestRelPath,
			Detail:   "manifest is not valid YAML: " + err.Error(),
		})
		return nil, true
	}

	addManifest := func(detail string) {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidManifest,
			Severity: SeverityError,
			Path:     manifestRelPath,
			Detail:   detail,
		})
	}

	// Required fields + value rules (all errors on violation).
	if m.SchemaVersion == nil {
		addManifest("missing required field: schema_version")
	} else if *m.SchemaVersion != 1 {
		addManifest(fmt.Sprintf("schema_version must be 1, got %d", *m.SchemaVersion))
	}
	if m.Name == "" {
		addManifest("missing required field: name")
	}
	if m.Slug == "" {
		addManifest("missing required field: slug")
	} else if !slugCharset.MatchString(m.Slug) {
		addManifest(fmt.Sprintf("slug %q is out of charset — only [a-z0-9-] allowed", m.Slug))
	}
	if m.Status == "" {
		addManifest("missing required field: status")
	} else if !manifestStatuses[m.Status] {
		addManifest(fmt.Sprintf("status %q is not one of idea|active|paused|done|archived", m.Status))
	}
	if m.Created == "" {
		addManifest("missing required field: created")
	} else if !dateCharset.MatchString(m.Created) {
		addManifest(fmt.Sprintf("created %q must be YYYY-MM-DD", m.Created))
	}

	// memory_dir is optional + present-optional, but when present it must be
	// repo-relative and non-traversing. Accidental-grade per the plan: reuse
	// wsvalidate.InvalidReason on the Normalise'd value, trimming the
	// trailing slash Normalise appends so the reason text reads cleanly.
	//
	// On an invalid (traversing/absolute) value we blank m.MemoryDir AFTER
	// recording the error so downstream code never joins it onto projectRoot
	// and walks an out-of-tree directory. A blanked-because-invalid value is
	// NOT the same as an absent key: auditProject must skip the walk entirely
	// for this run rather than falling back to the default tree (the manifest
	// is broken — the error already tells the user to fix it). See the
	// memoryDirInvalid flag threaded through auditProject.
	if m.MemoryDir != "" {
		norm := strings.TrimRight(wsvalidate.Normalise(m.MemoryDir), "/")
		if reason := wsvalidate.InvalidReason(norm); reason != "" {
			addManifest(fmt.Sprintf("memory_dir %q invalid: %s", m.MemoryDir, reason))
			m.MemoryDir = ""
			m.memoryDirInvalid = true
		}
	}

	// links is present-optional; unknown keys are ignored (not an error) and
	// URL content is intentionally not validated — nothing to do here.

	return &m, true
}

// auditMemoryTree walks {memory_dir}/** in two passes: first it collects and
// frontmatter-lints every note (building the permalink index), then it
// resolves relations + superseded_by against that index. Symlinked files or
// directories whose resolved target escapes memory_dir are rejected.
func auditMemoryTree(projectRoot, memDirRel, memDirAbs, slug string, slugKnown bool, report *Report) {
	// Resolve the memory_dir boundary once. EvalSymlinks canonicalises the
	// real on-disk root so per-entry resolved paths can be prefix-checked
	// against it. If the root itself can't be resolved, treat the tree as
	// unwalkable rather than guessing.
	rootResolved, err := filepath.EvalSymlinks(memDirAbs)
	if err != nil {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     filepath.ToSlash(memDirRel),
			Detail:   "could not resolve memory_dir: " + err.Error(),
		})
		return
	}
	boundary := rootResolved + string(os.PathSeparator)

	// underBoundary reports whether resolved is rootResolved itself or a
	// descendant of it — the adversarial-grade prefix check.
	underBoundary := func(resolved string) bool {
		return resolved == rootResolved || strings.HasPrefix(resolved, boundary)
	}

	var records []noteRecord
	index := map[string]bool{} // sanitized permalink → present
	fileCount := 0
	capHit := false

	walkErr := filepath.WalkDir(memDirAbs, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			// A per-entry stat/read error: record it and keep walking the
			// rest of the tree rather than aborting the whole pass.
			report.Issues = append(report.Issues, Issue{
				Category: CategoryInvalidNote,
				Severity: SeverityError,
				Path:     relToProject(projectRoot, path),
				Detail:   "could not walk entry: " + walkErr.Error(),
			})
			return nil
		}

		// Symlink rejection — adversarial-grade. Any symlink (file or dir)
		// whose resolved target escapes memory_dir is refused: not read, not
		// indexed, and (for dirs) not descended into. We conservatively skip
		// in-bounds symlinks too — they are an unusual shape in a
		// hand-authored note tree and WalkDir does not follow them anyway.
		if d.Type()&fs.ModeSymlink != 0 {
			resolved, rerr := filepath.EvalSymlinks(path)
			if rerr != nil || !underBoundary(resolved) {
				report.Issues = append(report.Issues, Issue{
					Category: CategorySymlinkEscape,
					Severity: SeverityError,
					Path:     relToProject(projectRoot, path),
					Detail:   "symlink target escapes memory_dir (or is unresolvable) — refused",
				})
			}
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		if d.IsDir() {
			return nil
		}
		// Only *.md files are notes. .gitkeep and anything else is ignored.
		if !strings.EqualFold(filepath.Ext(path), ".md") {
			return nil
		}

		fileCount++
		if fileCount > maxNoteFiles {
			capHit = true
			return fs.SkipAll
		}

		rec, ok := lintNoteFile(projectRoot, path, slug, slugKnown, report)
		if ok {
			records = append(records, rec)
			index[rec.permalink] = true
		}
		return nil
	})

	if walkErr != nil && walkErr != fs.SkipAll {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     filepath.ToSlash(memDirRel),
			Detail:   "memory tree walk failed: " + walkErr.Error(),
		})
	}

	if capHit {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityWarning,
			Path:     filepath.ToSlash(memDirRel),
			Detail:   fmt.Sprintf("memory tree exceeds the %d-note cap — walk stopped early; some notes were not linted", maxNoteFiles),
		})
	}

	// Second pass: resolve relations + non-null superseded_by against the
	// index. Sort records for deterministic issue ordering.
	sort.Slice(records, func(i, j int) bool { return records[i].relPath < records[j].relPath })
	for _, rec := range records {
		if rec.supersededBy != "" && !index[rec.supersededBy] {
			report.Issues = append(report.Issues, Issue{
				Category: CategoryInvalidNote,
				Severity: SeverityError,
				Name:     rec.permalink,
				Path:     rec.relPath,
				Detail:   fmt.Sprintf("superseded_by %q does not resolve to an existing note permalink", rec.supersededBy),
			})
		}
		for _, target := range rec.relations {
			if !index[target] {
				report.Issues = append(report.Issues, Issue{
					Category: CategoryUnresolvedRelation,
					Severity: SeverityWarning,
					Name:     rec.permalink,
					Path:     rec.relPath,
					Detail:   fmt.Sprintf("relation [[%s]] points at a not-yet-existing note (forward reference)", target),
				})
			}
		}
	}
}

// lintNoteFile reads + lints a single note. It returns (record, true) only
// when the note has a usable permalink to seed the index; on a fatal
// frontmatter problem it appends the error and returns ok=false so the broken
// note is neither indexed nor considered a valid relation target. Note that a
// note can still produce errors AND return ok=true (e.g. a bad scope) — the
// permalink alone gates indexing.
func lintNoteFile(projectRoot, path, slug string, slugKnown bool, report *Report) (noteRecord, bool) {
	rel := relToProject(projectRoot, path)

	fi, err := os.Lstat(path)
	if err != nil {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     rel,
			Detail:   "could not stat note: " + err.Error(),
		})
		return noteRecord{}, false
	}
	if fi.Size() > maxNoteSizeBytes {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     rel,
			Detail:   fmt.Sprintf("note exceeds the %d-byte size cap — not parsed", int64(maxNoteSizeBytes)),
		})
		return noteRecord{}, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     rel,
			Detail:   "could not read note: " + err.Error(),
		})
		return noteRecord{}, false
	}

	fm, body, ok := splitFrontmatter(data)
	if !ok {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     rel,
			Detail:   "missing or unterminated YAML frontmatter block",
		})
		return noteRecord{}, false
	}

	var note noteFrontmatter
	if err := yaml.Unmarshal(fm, &note); err != nil {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Path:     rel,
			Detail:   "frontmatter is not valid YAML: " + err.Error(),
		})
		return noteRecord{}, false
	}

	add := func(detail string) {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryInvalidNote,
			Severity: SeverityError,
			Name:     note.Permalink,
			Path:     rel,
			Detail:   detail,
		})
	}

	// Required-field checks (all errors).
	if note.SchemaVersion == nil {
		add("missing required frontmatter field: schema_version")
	} else if *note.SchemaVersion != 1 {
		add(fmt.Sprintf("schema_version must be 1, got %d", *note.SchemaVersion))
	}
	if note.Title == "" {
		add("missing required frontmatter field: title")
	}
	if note.Type == "" {
		add("missing required frontmatter field: type")
	} else if !noteTypes[note.Type] {
		add(fmt.Sprintf("type %q is not one of decision|note|fact|log", note.Type))
	}
	if note.Scope == "" {
		add("missing required frontmatter field: scope")
	} else if slugKnown && note.Scope != "project/"+slug {
		add(fmt.Sprintf("scope %q does not match manifest slug — expected %q", note.Scope, "project/"+slug))
	}

	// Permalink is required AND charset-checked. An out-of-charset permalink
	// is an error and must never be indexed (so it cannot satisfy a
	// relation), so we don't seed the index from it.
	permalinkOK := false
	if note.Permalink == "" {
		add("missing required frontmatter field: permalink")
	} else if !slugCharset.MatchString(note.Permalink) {
		add(fmt.Sprintf("permalink %q is out of charset — only [a-z0-9-] allowed", note.Permalink))
	} else {
		permalinkOK = true
	}

	if !permalinkOK {
		return noteRecord{}, false
	}

	rec := noteRecord{
		relPath:   rel,
		permalink: note.Permalink,
	}
	// superseded_by: absent ≡ null ≡ not-superseded (no error). Only a
	// non-null value is recorded for second-pass resolution. Sanitize before
	// indexing — link text is never a path.
	if note.SupersededBy != nil {
		if s := sanitizePermalink(*note.SupersededBy); s != "" {
			rec.supersededBy = s
		}
	}
	rec.relations = extractRelations(body)
	return rec, true
}

// splitFrontmatter separates the leading `---`-delimited YAML frontmatter
// from the markdown body. Returns (frontmatterBytes, bodyBytes, true) on a
// well-formed block; (nil, nil, false) when the file does not open with a
// frontmatter fence or the closing fence is missing.
func splitFrontmatter(data []byte) ([]byte, []byte, bool) {
	s := string(data)
	s = strings.TrimPrefix(s, "\ufeff") // tolerate a leading UTF-8 BOM
	if !strings.HasPrefix(s, "---\n") && !strings.HasPrefix(s, "---\r\n") {
		return nil, nil, false
	}
	// Skip the opening fence line, then scan for the closing fence.
	rest := s[strings.IndexByte(s, '\n')+1:]
	lines := strings.SplitAfter(rest, "\n")
	var fmLines []string
	consumed := 0
	closed := false
	for _, ln := range lines {
		trimmed := strings.TrimRight(ln, "\r\n")
		consumed += len(ln)
		if trimmed == "---" {
			closed = true
			break
		}
		fmLines = append(fmLines, ln)
	}
	if !closed {
		return nil, nil, false
	}
	fm := strings.Join(fmLines, "")
	body := rest[consumed:]
	return []byte(fm), []byte(body), true
}

// extractRelations pulls every [[target]] wikilink body out of the note body,
// sanitizes each to a permalink, and returns the deduplicated set. Link text
// is ONLY ever turned into an index key — never a filesystem path.
func extractRelations(body []byte) []string {
	matches := wikilinkPattern.FindAllStringSubmatch(string(body), -1)
	seen := map[string]bool{}
	var out []string
	for _, m := range matches {
		s := sanitizePermalink(m[1])
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// sanitizePermalink reduces arbitrary link text to the [a-z0-9-] permalink
// charset for index lookups: lowercase, then drop every out-of-charset rune.
// This is a pure string transform — the result is used solely as a map key,
// never as a path component. An empty result means "no resolvable target".
func sanitizePermalink(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// auditMemoryIndex flags MEMORY.md when it exceeds the line budget. The index
// is expected one directory above memory_dir (it is scaffolded under
// docs_path beside the Memory/ tree). Absent index → nothing to flag.
func auditMemoryIndex(projectRoot, memDirRel string, report *Report) {
	indexRel := filepath.ToSlash(filepath.Join(filepath.Dir(filepath.FromSlash(memDirRel)), memoryIndexName))
	abs := filepath.Join(projectRoot, filepath.FromSlash(indexRel))

	fi, err := os.Lstat(abs)
	if err != nil || fi.IsDir() {
		return
	}
	if fi.Size() > maxNoteSizeBytes {
		// Pathologically large index — don't read it, just flag the budget
		// breach (a >1MiB MEMORY.md is far past 200 lines by any measure).
		report.Issues = append(report.Issues, Issue{
			Category: CategoryMemoryIndexTooLarge,
			Severity: SeverityWarning,
			Path:     indexRel,
			Detail:   fmt.Sprintf("%s exceeds the %d-line budget", memoryIndexName, memoryIndexMaxLines),
		})
		return
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		return
	}
	lines := strings.Count(string(data), "\n")
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		lines++ // count a final unterminated line
	}
	if lines > memoryIndexMaxLines {
		report.Issues = append(report.Issues, Issue{
			Category: CategoryMemoryIndexTooLarge,
			Severity: SeverityWarning,
			Path:     indexRel,
			Detail:   fmt.Sprintf("%s is %d lines — exceeds the %d-line budget", memoryIndexName, lines, memoryIndexMaxLines),
		})
	}
}

// relToProject returns the slash-form project-relative path for abs, falling
// back to the slash-form abs when Rel fails (it shouldn't here — abs is
// always under projectRoot).
func relToProject(projectRoot, abs string) string {
	rel, err := filepath.Rel(projectRoot, abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}
