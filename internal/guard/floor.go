package guard

// The floor (design/plan-5.md, 5.3.2 note 2): the paths the guard guards whatever a project's lists say, because
// Bonsai's own rules and lines live in them, creation included. A project may leave its lists short, and a guard whose
// own files an agent may unguard through the same guard guards nothing.
//
// Floor below is the list's one home, printed on the reference page. Each entry is a project-relative glob, matched as
// the lists' globs are (matchGlob: a glob naming a folder holds what is in it; ** any number of whole folders), against
// the path relative to each project judging it (rule.go's judges), letter case folded where the project's folder
// ignores it. The first entry that matches says what the path is:
//
//   - person-only: guarded, and grantable as rule 3 says (a running task's bonsai.allows; in command mode only a
//     person's grant should put it there, which tripwire.go's grant-person-only refusal stands behind);
//   - never-granted: no grant opens it, so an agent's own task cannot open the settings files or .git in agents mode,
//     where two file-tool edits (a grant, then the edit) would otherwise take the guard's line out; a person, or bonsai
//     update run by one, changes them;
//   - bonsai-files: Bonsai's alone (rule 5): Bonsai writes them as a program, never through the file tools;
//   - free: the floor says nothing more (.bonsai/STATE.md, which agents rewrite: spec §6); the lists may still guard it.
//
// At a checkout's top the entries are anchored; the any-depth entries (**/) keep a project or a checkout from being
// planted inside another (5.3.1 note 1: a planted bonsai.yaml or .git narrows nothing, and the floor refuses an
// agent's tools writing either). For a delete or move (step 5.3.3), a .git below the top is not guarded (5.3.3 note 4:
// a nested repository's .git inside node_modules, say): NotForDeletes.

import "strings"

// The classes of a floor entry: the rule a guarded one is refused under, or free.
const floorFree = "free"

// FloorPath is one entry of the floor.
type FloorPath struct {
	Glob          string // project-relative, forward slashes, as the lists' globs
	At            string // "top" (a checkout's top) or "any depth"
	Class         string // RulePersonOnly, RuleNeverGranted, RuleBonsaiFiles, or "free"
	NotForDeletes bool   // a delete or move does not meet it (5.3.3 note 4)
	Why           string // what it is, ASCII: said in a refusal and on the reference page
}

// Floor is the floor's table, its one home, in the order it is matched: the first entry that matches a path decides.
var Floor = []FloorPath{
	{".bonsai/STATE.md", "top", floorFree, false, "where the project stands; agents rewrite it (spec section 6)"},
	{".bonsai/local/", "top", RuleBonsaiFiles, false, "the log, the asks and the ladder results, which only Bonsai writes"},
	{".bonsai/.gitignore", "top", RuleBonsaiFiles, false, "written by Bonsai, so .bonsai/local/ is never committed"},
	{".bonsai/tasks.md", "top", RuleBonsaiFiles, false, "a generated table: bonsai check --write rebuilds it"},
	{".bonsai/sessions.md", "top", RuleBonsaiFiles, false, "a generated table: bonsai check --write rebuilds it"},
	{".bonsai/lock.json", "top", RulePersonOnly, false, "the lock: what each pack wrote and declares"},
	{".bonsai", "top", RulePersonOnly, false, "Bonsai's folder in the project; nothing else of an agent's belongs in it"},
	{"bonsai.yaml", "top", RulePersonOnly, false, "the project's Bonsai settings, its lists among them"},
	{".claude/settings.json", "top", RuleNeverGranted, false, "Claude Code's project settings, which hold Bonsai's hook lines and deny rules"},
	{".claude/settings.local.json", "top", RuleNeverGranted, false, "Claude Code's local settings, which beat the project's"},
	{".git", "top", RuleNeverGranted, false, "the checkout's git folder or file, which says where its main checkout is"},
	{"**/.git", "any depth", RuleNeverGranted, true, "a .git below the top would plant a checkout inside the project"},
	{"**/bonsai.yaml", "any depth", RulePersonOnly, false, "a bonsai.yaml below the top would plant a project inside this one"},
	{"**/.bonsai", "any depth", RulePersonOnly, false, "a .bonsai folder below the top is a planted project's"},
}

// floorOf gives the floor's entry for a project-relative path, nil when none matches; with del, the entries a delete
// does not meet are passed over. fold ignores letter case.
func floorOf(rel string, fold, del bool) *FloorPath {
	for i := range Floor {
		f := &Floor[i]
		if del && f.NotForDeletes {
			continue
		}
		if matchGlob(f.Glob, rel, fold) {
			return f
		}
	}
	return nil
}

// floorGlobs are the floor's guarded globs (every entry but the free ones), for a delete (del) or not: what the delete
// check meets a folder against (judgedRoots).
func floorGlobs(del bool) []string {
	var out []string
	for _, f := range Floor {
		if f.Class == floorFree || (del && f.NotForDeletes) {
			continue
		}
		out = append(out, f.Glob)
	}
	return out
}

// floorPersonOnly are the floor's person-only globs: with bonsai.yaml's person_only list, what tripwire.go's
// grant-person-only judges a new grant against.
func floorPersonOnly() []string {
	var out []string
	for _, f := range Floor {
		if f.Class == RulePersonOnly {
			out = append(out, f.Glob)
		}
	}
	return out
}

// floorFolders are the folders at a checkout's top that hold the floor's top paths: with the top folder itself, all
// a delete or move in another linked project is judged by (5.3.2 note 8).
var floorFolders = []string{".git", ".bonsai", ".claude"}

// isFloorFolder reports whether a project-relative path is one of floorFolders itself, letter case folded with fold.
func isFloorFolder(rel string, fold bool) bool {
	for _, f := range floorFolders {
		if rel == f || (fold && strings.EqualFold(rel, f)) {
			return true
		}
	}
	return false
}
