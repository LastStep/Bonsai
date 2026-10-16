package engine

// The pack cache, cleaned in update (spec section 6, "Generated files": the cache in update; plan-5 5.6.1, 6). With the
// home's cache_keep_days set (bonsai settings set cache_keep_days=<days> --machine), update --yes removes, after its
// own run, each pack clone in <home>/cache/git/ whose last use is older than that many days, never one the project's
// lock or bonsai.yaml names. A clone is a refillable copy outside every project: the next fetch that needs it clones
// it again. The --adopt copies in <home>/cache/adopted/ are a person's saved edits and are never cleaned. Nothing
// else cleans the cache: never a preview, never check or status.
//
// A clone's last use is a mark, the file bonsai-used in the clone's folder, whose time every fetch and every read of
// the clone sets (markUsed), so a clone another project is using now is never older than a day. A clone with no mark
// (fetched by a Bonsai before step 5.6.1) counts its newest top-level entry's time.

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// usedMark is the file in a clone's folder whose time is the clone's last use.
const usedMark = "bonsai-used"

// markUsed sets a clone's last use to now; a clone that is not there yet, or a mark that cannot be written, is left
// as it is (the clone then counts its own entries' times).
func markUsed(dir string) {
	if _, err := os.Stat(filepath.Join(dir, "HEAD")); err != nil {
		return
	}
	p := filepath.Join(dir, usedMark)
	now := cacheClock()
	if err := os.Chtimes(p, now, now); err == nil {
		return
	}
	if err := os.WriteFile(p, []byte("the time of this file is when Bonsai last fetched or read this clone\n"), 0o644); err == nil {
		_ = os.Chtimes(p, now, now)
	}
}

// cacheClock is the time a clone is marked and judged by (a variable, so a test can move it).
var cacheClock = time.Now

// cloneName is a clone's folder name in cache/git/ (fetch.go's cache.dir): only such folders are ever removed.
var cloneName = regexp.MustCompile(`^[0-9a-f]{16}\.git$`)

// lastUse is a clone's last use: its mark's time, else its newest top-level entry's.
func lastUse(dir string) (time.Time, bool) {
	if fi, err := os.Stat(filepath.Join(dir, usedMark)); err == nil {
		return fi.ModTime(), true
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return time.Time{}, false
	}
	var newest time.Time
	for _, e := range entries {
		if fi, err := e.Info(); err == nil && fi.ModTime().After(newest) {
			newest = fi.ModTime()
		}
	}
	if fi, err := os.Stat(dir); err == nil && fi.ModTime().After(newest) {
		newest = fi.ModTime()
	}
	return newest, !newest.IsZero()
}

// CacheCleaned is what CleanCache did: the clones removed (their folder names, sorted), and those it could not remove.
type CacheCleaned struct {
	Days    int
	Removed []string
	Failed  []string // "<name>: <why>"
	Problem string   // the home's settings could not be read: nothing was cleaned
}

// CleanCache removes the clones in <home>/cache/git/ unused for longer than the home's cache_keep_days, keeping each
// clone of a source in keep (the project's lock's and bonsai.yaml's). With cache_keep_days none (the default) it does
// nothing and gives nil. It never touches cache/adopted/, never fails the command that calls it, and gives what it
// did for that command's text.
func CleanCache(home string, keep []string) *CacheCleaned {
	if home == "" {
		return nil
	}
	hs, err := workspace.LoadHomeSettings(home)
	if err != nil {
		return &CacheCleaned{Problem: oneLine(err.Error())}
	}
	if hs.CacheKeepDays == 0 {
		return nil
	}
	out := &CacheCleaned{Days: hs.CacheKeepDays}
	c := cache{home: home}
	kept := map[string]bool{}
	for _, s := range keep {
		kept[filepath.Base(c.dir(s))] = true
	}
	root := filepath.Join(home, "cache", "git")
	entries, err := os.ReadDir(root)
	if err != nil {
		return out
	}
	cutoff := cacheClock().Add(-time.Duration(hs.CacheKeepDays) * 24 * time.Hour)
	for _, e := range entries {
		name := e.Name()
		if !e.IsDir() || !cloneName.MatchString(name) || kept[name] {
			continue
		}
		dir := filepath.Join(root, name)
		used, ok := lastUse(dir)
		if !ok || !used.Before(cutoff) {
			continue
		}
		if err := removeTree(dir); err != nil {
			out.Failed = append(out.Failed, name+": "+oneLine(err.Error()))
			continue
		}
		out.Removed = append(out.Removed, name)
	}
	sort.Strings(out.Removed)
	sort.Strings(out.Failed)
	return out
}

// removeTree removes a folder and everything in it; a file git made read-only (its packs, on Windows) is made
// writable and the removal tried again.
func removeTree(dir string) error {
	if err := os.RemoveAll(dir); err == nil {
		return nil
	}
	_ = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err == nil {
			_ = os.Chmod(p, 0o755)
		}
		return nil
	})
	return os.RemoveAll(dir)
}

// Line is what the clean did, in one line for update's text; "" when there is nothing to say.
func (cc *CacheCleaned) Line() string {
	switch {
	case cc == nil:
		return ""
	case cc.Problem != "":
		return "the pack cache was not cleaned: " + ascii(cc.Problem)
	case len(cc.Removed) == 0 && len(cc.Failed) == 0:
		return ""
	}
	line := ""
	if n := len(cc.Removed); n > 0 {
		line = plural(n, "removed 1 pack clone", "removed "+strconv.Itoa(n)+" pack clones") + " unused for over " +
			plural(cc.Days, "1 day", strconv.Itoa(cc.Days)+" days") + " from the cache (cache_keep_days)"
	}
	if len(cc.Failed) > 0 {
		if line != "" {
			line += "; "
		}
		line += "could not remove " + ascii(strings.Join(cc.Failed, ", "))
	}
	return line
}
