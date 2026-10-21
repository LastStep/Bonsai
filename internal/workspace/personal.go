package workspace

// The personal memory layer (spec section 10; contract section 7.4; plan-5 5.6.3): notes about the person, kept on
// this machine in the Bonsai home, never in a project and never forwarded (contract section 2.6).
//
//	<home>/personal/INDEX.md            the index, bonsai.memory/1 with kind index: one line per note
//	<home>/personal/notes/<name>.md     one note a fact, bonsai.memory/1
//	<home>/cache/personal-copy.json     on Windows, the record of the last refresh of the copy (below)
//
// A person writes the layer: the guard and the engine's deny rule Edit(~/.bonsai/**) refuse an agent's file-tool
// write anywhere in the home, since the index loads into every session in every project on the machine. An agent
// drafts a note with base's memory skill and hands it over. init --yes and update --yes write an empty index when
// there is none (EnsurePersonalIndex), never over one, never in a preview, never from check or status; unlink leaves
// the layer, which is the machine's.
//
// The Windows copy. WSL's layer is the canonical one; on Windows, once a person has named WSL's folder in the home
// setting personal_from (such as //wsl.localhost/<distro>/home/<user>/.bonsai/personal), init --yes and update --yes
// refresh the home's personal/ from it when it differs (RefreshPersonal): only INDEX.md and notes/*.md, regular files,
// following no link (a link is left out and named), each at most PersonalFileMax bytes and at most PersonalNotesMax
// notes (past either nothing is copied, and the outcome says why). Each read of the source has PersonalReadWait;
// past it the refresh stops, recorded as a failure, never a wait. Each refresh records its outcome in
// cache/personal-copy.json, which check reads without reaching WSL. Example of the record:
//
//	{
//	  "source": "//wsl.localhost/Ubuntu/home/someone/.bonsai/personal",
//	  "refreshed": "2026-10-18T10:00:00Z",
//	  "files": [
//	    {"path": "INDEX.md", "sha256": "<64 hex>"},
//	    {"path": "notes/M-short-replies.md", "sha256": "<64 hex>"}
//	  ],
//	  "failed": null,
//	  "why": null
//	}
//
// source is the folder read; refreshed when the copy last matched it (RFC 3339, UTC), or null when it never has;
// files each copied file, its path in personal/ with forward slashes and the SHA-256 of its bytes, in path order;
// failed when the last refresh failed (null when it did not) and why, one line.

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The personal layer's places in the home.
const (
	PersonalDir    = "personal" // <home>/personal/
	PersonalIndex  = "INDEX.md" // its index
	PersonalNotes  = "notes"    // its notes' folder
	PersonalRecord = "cache/personal-copy.json"
)

// The Windows copy's limits.
const (
	PersonalFileMax  = 64 * 1024 // the most bytes a copied file may hold
	PersonalNotesMax = 500       // the most notes copied
)

// PersonalReadWait is how long a refresh waits on each read of the source (a variable for the tests): a read through
// \\wsl.localhost can start WSL.
var PersonalReadWait = 2 * time.Second

// PersonalPath is <home>/personal/.
func PersonalPath(home string) string { return filepath.Join(home, PersonalDir) }

// PersonalIndexText is the empty index init and update write: bonsai.memory/1 with kind index, today's date, and one
// body line saying what goes there and who writes it.
func PersonalIndexText(today time.Time) []byte {
	return []byte("---\n" +
		"format: bonsai.memory/1   # fields: bonsai check --schema bonsai.memory; how to fill: skill base:memory\n" +
		"id: null\n" +
		"title: Personal memory\n" +
		"kind: index\n" +
		"updated: " + today.UTC().Format("2006-01-02") + "\n" +
		"source: null\n" +
		"labels: {}\n" +
		"---\n\n" +
		"Notes about the person, for every project on this machine: one line per note in notes/, written by the person " +
		"(an agent drafts a note with base's memory skill and hands it over).\n")
}

// EnsurePersonalIndex writes the empty index when the home has none, staged and renamed, and says whether it wrote.
// Anything already at the index's place (a file, a folder, a link) is left as it is.
func EnsurePersonalIndex(home string, today time.Time) (bool, error) {
	p := filepath.Join(PersonalPath(home), PersonalIndex)
	if _, err := os.Lstat(p); err == nil || !errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err := WriteFileAtomic(p, PersonalIndexText(today)); err != nil {
		return false, err
	}
	return true, nil
}

// CopiedFile is one file of the copy: its path in personal/ (forward slashes) and the SHA-256 of its bytes.
type CopiedFile struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

// PersonalCopyRecord is cache/personal-copy.json.
type PersonalCopyRecord struct {
	Source    string       `json:"source"`
	Refreshed *string      `json:"refreshed"`
	Files     []CopiedFile `json:"files"`
	Failed    *string      `json:"failed"`
	Why       *string      `json:"why"`
}

// ReadPersonalRecord reads the record; nil with no error when there is none.
func ReadPersonalRecord(home string) (*PersonalCopyRecord, error) {
	raw, err := os.ReadFile(filepath.Join(home, filepath.FromSlash(PersonalRecord)))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r := &PersonalCopyRecord{}
	if err := json.Unmarshal(raw, r); err != nil {
		return nil, errors.New("is not JSON Bonsai reads: " + oneLine(err))
	}
	return r, nil
}

// writePersonalRecord writes the record whole, byte-stable (two-space indent, LF).
func writePersonalRecord(home string, r *PersonalCopyRecord) error {
	if r.Files == nil {
		r.Files = []CopiedFile{}
	}
	raw, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return WriteFileAtomic(filepath.Join(home, filepath.FromSlash(PersonalRecord)), append(raw, '\n'))
}

// PersonalFiles lists the layer's files as the copy counts them, in a personal/ folder: INDEX.md and notes/*.md,
// regular files only, each with its SHA-256, in path order. A folder that is not there lists nothing.
func PersonalFiles(dir string) ([]CopiedFile, error) {
	var out []CopiedFile
	add := func(rel string) error {
		p := filepath.Join(dir, filepath.FromSlash(rel))
		fi, err := os.Lstat(p)
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if !fi.Mode().IsRegular() {
			return nil
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		out = append(out, CopiedFile{Path: rel, SHA256: sum(raw)})
		return nil
	}
	if err := add(PersonalIndex); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(filepath.Join(dir, PersonalNotes))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			if err := add(PersonalNotes + "/" + e.Name()); err != nil {
				return nil, err
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out, nil
}

func sum(raw []byte) string {
	h := sha256.Sum256(raw)
	return hex.EncodeToString(h[:])
}

// PersonalRefresh is what RefreshPersonal did, for init's and update's one line.
type PersonalRefresh struct {
	Source  string
	Copied  int      // files written
	Removed int      // notes taken out, the source no longer having them
	Same    bool     // the copy already matched the source
	Skipped []string // what was left out (a link, a file that is not regular), path and why
	Failed  string   // why nothing was refreshed; "" when it was
}

// sourceFS is how a refresh reads the source: variables, so a test can stand in a source that never answers.
var (
	sourceLstat    = os.Lstat
	sourceReadDir  = os.ReadDir
	sourceReadFile = os.ReadFile
)

// errSlow is a read of the source that took longer than PersonalReadWait.
var errSlow = errors.New("slow")

// within runs one read of the source, waiting PersonalReadWait at most. A read past it is left to end on its own,
// its result dropped.
func within[T any](read func() (T, error)) (T, error) {
	type result struct {
		v   T
		err error
	}
	done := make(chan result, 1)
	go func() {
		v, err := read()
		done <- result{v, err}
	}()
	t := time.NewTimer(PersonalReadWait)
	defer t.Stop()
	select {
	case r := <-done:
		return r.v, r.err
	case <-t.C:
		var zero T
		return zero, errSlow
	}
}

// RefreshPersonal brings the home's personal/ to the source's layer when they differ (the Windows copy, above) and
// records the outcome in cache/personal-copy.json. clock gives the time recorded. It never fails the command that
// calls it: what went wrong is in the result's Failed, and in the record.
func RefreshPersonal(home, source string, clock func() time.Time) *PersonalRefresh {
	out := &PersonalRefresh{Source: source}
	rec, err := ReadPersonalRecord(home)
	if err != nil || rec == nil || rec.Source != source {
		rec = &PersonalCopyRecord{} // another source's record, or none: this one starts afresh
	}
	rec.Source = source
	fail := func(why string) *PersonalRefresh {
		out.Failed = why
		when := clock().UTC().Format(time.RFC3339)
		rec.Failed, rec.Why = &when, &why
		if err := writePersonalRecord(home, rec); err != nil {
			out.Failed += "; and its record could not be written: " + oneLine(err)
		}
		return out
	}
	target := PersonalPath(home)
	if insideFolder(source, target) {
		return fail("personal_from names the home's own personal/ (or a folder in it), which is the copy's target")
	}
	files, skipped, why := readSource(source)
	if why != "" {
		return fail(why)
	}
	out.Skipped = skipped
	have, err := PersonalFiles(target)
	if err != nil {
		return fail("the home's personal/ cannot be read: " + oneLine(err))
	}
	haveSum := map[string]string{}
	for _, f := range have {
		haveSum[f.Path] = f.SHA256
	}
	var copied []CopiedFile
	want := map[string]bool{}
	for _, f := range files {
		s := sum(f.raw)
		want[f.path] = true
		copied = append(copied, CopiedFile{Path: f.path, SHA256: s})
		if haveSum[f.path] == s {
			continue
		}
		if err := WriteFileAtomic(filepath.Join(target, filepath.FromSlash(f.path)), f.raw); err != nil {
			return fail(f.path + " cannot be written: " + oneLine(err))
		}
		out.Copied++
	}
	for _, f := range have {
		if want[f.Path] || !strings.HasPrefix(f.Path, PersonalNotes+"/") {
			continue
		}
		if err := RemoveFile(filepath.Join(target, filepath.FromSlash(f.Path))); err != nil {
			return fail(f.Path + " cannot be removed: " + oneLine(err))
		}
		out.Removed++
	}
	out.Same = out.Copied == 0 && out.Removed == 0
	when := clock().UTC().Format(time.RFC3339)
	sort.Slice(copied, func(i, j int) bool { return copied[i].Path < copied[j].Path })
	rec.Refreshed, rec.Files, rec.Failed, rec.Why = &when, copied, nil, nil
	if err := writePersonalRecord(home, rec); err != nil {
		out.Failed = "the copy was refreshed, but its record cannot be written: " + oneLine(err)
	}
	return out
}

// sourceFile is one file read from the source.
type sourceFile struct {
	path string // in personal/, forward slashes
	raw  []byte
}

// readSource reads the source's INDEX.md and notes/*.md, each read within PersonalReadWait. It gives the files, what
// it left out, and why nothing can be copied ("" when it can).
func readSource(source string) ([]sourceFile, []string, string) {
	slow := func(what string) string {
		wait := strconv.FormatInt(PersonalReadWait.Milliseconds(), 10) + " ms"
		if PersonalReadWait%time.Second == 0 {
			wait = strconv.Itoa(int(PersonalReadWait/time.Second)) + " s"
		}
		return what + " did not answer within " + wait + ": nothing was compared or copied"
	}
	fi, err := within(func() (fs.FileInfo, error) { return sourceLstat(source) })
	switch {
	case errors.Is(err, errSlow):
		return nil, nil, slow("the source " + filepath.ToSlash(source))
	case err != nil:
		return nil, nil, "the source " + filepath.ToSlash(source) + " cannot be read: " + oneLine(err)
	case !fi.IsDir():
		return nil, nil, "the source " + filepath.ToSlash(source) + " is not a folder"
	}
	var files []sourceFile
	var skipped []string
	read := func(rel string, need bool) string {
		p := filepath.Join(source, filepath.FromSlash(rel))
		fi, err := within(func() (fs.FileInfo, error) { return sourceLstat(p) })
		switch {
		case errors.Is(err, errSlow):
			return slow(rel)
		case errors.Is(err, fs.ErrNotExist) && need:
			return "the source holds no " + rel + ": nothing was copied"
		case err != nil:
			return rel + " in the source cannot be read: " + oneLine(err)
		case !fi.Mode().IsRegular():
			skipped = append(skipped, rel+" (not a regular file: a link or a folder is never followed)")
			return ""
		case fi.Size() > PersonalFileMax:
			return rel + " in the source is " + strconv.FormatInt(fi.Size(), 10) + " bytes, over the copy's " +
				strconv.Itoa(PersonalFileMax/1024) + " KB a file: nothing was copied"
		}
		raw, err := within(func() ([]byte, error) { return sourceReadFile(p) })
		switch {
		case errors.Is(err, errSlow):
			return slow(rel)
		case err != nil:
			return rel + " in the source cannot be read: " + oneLine(err)
		case len(raw) > PersonalFileMax:
			return rel + " in the source is over the copy's " + strconv.Itoa(PersonalFileMax/1024) + " KB a file: nothing was copied"
		}
		files = append(files, sourceFile{path: rel, raw: raw})
		return ""
	}
	if why := read(PersonalIndex, true); why != "" {
		return nil, nil, why
	}
	entries, err := within(func() ([]fs.DirEntry, error) { return sourceReadDir(filepath.Join(source, PersonalNotes)) })
	switch {
	case errors.Is(err, errSlow):
		return nil, nil, slow(PersonalNotes + "/")
	case errors.Is(err, fs.ErrNotExist):
		entries = nil
	case err != nil:
		return nil, nil, PersonalNotes + "/ in the source cannot be read: " + oneLine(err)
	}
	var names []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".md") {
			names = append(names, e.Name())
		}
	}
	if len(names) > PersonalNotesMax {
		return nil, nil, "the source holds " + strconv.Itoa(len(names)) + " notes, over the copy's " + strconv.Itoa(PersonalNotesMax) + ": nothing was copied"
	}
	sort.Strings(names)
	for _, n := range names {
		if why := read(PersonalNotes+"/"+n, false); why != "" {
			return nil, nil, why
		}
	}
	return files, skipped, ""
}

// insideFolder reports whether p is folder or a path inside it, read with forward slashes, and without letter case
// on Windows.
func insideFolder(p, folder string) bool {
	norm := func(s string) string {
		s = strings.TrimSuffix(winSlash(filepath.Clean(s)), "/")
		if runtime.GOOS == "windows" {
			s = strings.ToLower(s)
		}
		return s
	}
	a, b := norm(p), norm(folder)
	return a == b || strings.HasPrefix(a, b+"/")
}

// Line is the refresh's one line for init's and update's text.
func (r *PersonalRefresh) Line() string {
	src := filepath.ToSlash(r.Source)
	switch {
	case r.Failed != "":
		return "the personal memory copy was not refreshed from " + src + ": " + r.Failed
	case r.Same:
		return appendSkipped("the personal memory copy is the same as "+src, r.Skipped)
	}
	return appendSkipped("the personal memory copy was refreshed from "+src+": "+strconv.Itoa(r.Copied)+" written, "+
		strconv.Itoa(r.Removed)+" removed", r.Skipped)
}

func appendSkipped(line string, skipped []string) string {
	if len(skipped) == 0 {
		return line
	}
	return line + "; left out: " + strings.Join(skipped, ", ")
}

// SamePersonal reports whether the files of a personal/ folder are those a record lists, path by path and byte by
// byte; differs names the first path that differs.
func SamePersonal(have, recorded []CopiedFile) (same bool, differs string) {
	if len(have) != len(recorded) {
		a, b := map[string]bool{}, map[string]bool{}
		for _, f := range have {
			a[f.Path] = true
		}
		for _, f := range recorded {
			b[f.Path] = true
		}
		for _, f := range have {
			if !b[f.Path] {
				return false, f.Path
			}
		}
		for _, f := range recorded {
			if !a[f.Path] {
				return false, f.Path
			}
		}
	}
	for i := range have {
		if have[i] != recorded[i] {
			return false, have[i].Path
		}
	}
	return true, ""
}
