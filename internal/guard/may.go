package guard

// The one function (design/plan-5.md, 5.3.2 note 1): may this agent change this path now? From the path's projects
// (rule.go: the projects holding it, the nearest alone when it is a verified checkout), their lists, the packs'
// declares, the floor (floor.go), Bonsai's own files (own.go), the active task's grants (grant.go) and command mode's
// tripwires (tripwire.go), one decision. The file tools call it for their path (rule.go's decide); the delete check
// (step 5.3.3) for each operand of a delete or move, read against the folder the command runs in (a change with del
// set); rung 0 (step 5.4) with the named task's grants at any status (the judge's taskGrants), so the guard and rung 0
// judge alike.
//
// A path is judged in every form it takes (paths.go) and refused when any form is refused. In each form, in order:
//
//  1. Bonsai's home: bonsai-files, never granted (own.go).
//  2. No linked project holds it: a stand-in for Bonsai's program is refused (bonsai-stand-in); anything else is
//     outside-project.
//  3. A delete or move (del) in a linked project that is not the session's own (5.3.2 note 8): only that project's
//     top folder and its floor folders (.git, .bonsai, .claude) are judged, when the project is verified on this
//     machine's record (other-project); anything below them is the shell write question A leaves to rung 0 and the
//     Desk.
//  4. Each project judging it: its lists must read and its globs be ones the guard reads (bad-config, bad-lock,
//     bad-glob). Then the floor's entry for the path and the lists' globs that match it: bonsai-files and
//     never-granted refuse whatever a task grants; person-only and protected refuse unless the grantor's running task
//     grants the path (granted), and refuse as main-not-verified when no project holding the path is verified.
//  5. In command mode, the tripwires of contract §10.1 for a declared document (tripwire.go).
//
// Every refusal says why first and names the next step (spec §3, §7).

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// change is one change the guard judges.
type change struct {
	tool string          // the tool or command, as a refusal names it: "Edit", "rm"; never ""
	path string          // the path as written
	cwd  string          // the folder a relative path is read against (the payload's cwd, a command's folder)
	base string          // the folder used when cwd is not absolute: the session's project
	del  bool            // a delete or a move (step 5.3.3): note 3 of the comment above, and the floor's NotForDeletes
	in   *hookin.Payload // the file tool's call, which says what the file holds after it; nil when that is unknown
}

// caseFoldSystem reports a system whose file names ignore letter case everywhere (Windows, macOS).
func caseFoldSystem() bool {
	switch runtime.GOOS {
	case "windows", "darwin", "ios":
		return true
	}
	return false
}

// may judges one change (this file's comment): the decision, and the project its record goes in (nil: the session's
// project is nil too, and nothing is recorded).
func (j *judge) may(c change) (Decision, *linkedProject) {
	forms, err := pathForms(c.base, c.cwd, c.path)
	if err != nil {
		return deny(RuleBadPath, fmt.Sprintf("%s of %s refused: on Windows the guard cannot tell where the path leads (%s), "+
			"so it cannot tell whether it is guarded", c.tool, quote(filepath.ToSlash(c.path), 120), oneLine(err.Error())),
			"write to the file by its plain path inside the project (a drive letter and folders), or tell the person"), j.session
	}
	var first []*linkedProject // the chain of the first form a project holds: an allow is recorded there
	granted := false
	for _, abs := range forms {
		if devicePath(abs) || adminShare(abs) != "" {
			continue // a \\?\ form, or one through this machine's admin share, names no project folder of its own: its
			// drive or share form is among the forms
		}
		chain := j.chain(abs)
		if home := j.inHome(abs); home != "" {
			rec := j.recordIn(chain)
			d := deny(RuleBonsaiFiles, fmt.Sprintf("%s of %s refused: it is in Bonsai's home (%s), whose files only Bonsai "+
				"writes: this machine's records and settings, which the guard itself trusts", c.tool, quote(filepath.ToSlash(abs), 120),
				quote(filepath.ToSlash(home), 120)), "leave Bonsai's home as it is; a person runs the bonsai command that "+
				"changes it, in a terminal of their own")
			d.Target = target(rec, []string{abs})
			return d, rec
		}
		if len(chain) == 0 {
			if standIn(abs) {
				return deny(RuleStandIn, fmt.Sprintf("%s of %s refused: outside every linked project, a file named as "+
					"Bonsai's program could stand in for it in a folder on the PATH", c.tool, quote(filepath.ToSlash(abs), 120)),
					"give the file another name; a person installs Bonsai"), j.session
			}
			continue
		}
		if first == nil {
			first = chain
		}
		d, rec := j.form(c, abs, chain)
		if !d.Allow {
			return d, rec
		}
		granted = granted || d.Rule == RuleGranted
	}
	if first == nil {
		return allow(RuleOutside, ""), j.session
	}
	rec := j.recordIn(first)
	if granted {
		return allow(RuleGranted, target(rec, forms)), rec
	}
	return allow(RuleNotProtected, target(rec, forms)), rec
}

// mayDelete judges a delete or move of the path p, read against the folder dir (absolute) as a file tool's path is
// read against the payload's cwd, as an edit of that path: step 5.3.3's delete check calls it for each operand (its
// delMay), so the delete check holds no rule of its own about what is guarded.
func (j *judge) mayDelete(dir, p string) Decision {
	d, _ := j.may(change{tool: "a delete or move", path: p, cwd: dir, base: dir, del: true})
	return d
}

// judged is one form of a path, a project judging it, and the path relative to that project's folder.
type judged struct {
	abs   string
	chain []*linkedProject
	p     *linkedProject
	rel   string
	fold  bool // the project's folder ignores letter case
}

// hit is a guard on one judged form: the floor's entry or a list's glob that matched it.
type hit struct {
	x    judged
	rule string // RulePersonOnly, RuleProtected, RuleNeverGranted, RuleBonsaiFiles
	on   string // what holds it, for the refusal: "bonsai.yaml's protected list", "the floor (...)"
	glob string
	next string // the floor's own next step, "" for the rule's
}

// form judges one form of a path that the projects in chain hold (may's steps 3 to 5); an allow is not-protected
// or granted.
func (j *judge) form(c change, abs string, chain []*linkedProject) (Decision, *linkedProject) {
	rec := j.recordIn(chain)
	if c.del && !j.ownChain(chain) {
		for _, p := range chain {
			rel, ok := relTo(p.root, abs)
			if !ok || !isFloorFolder(rel, caseInsensitive(p.root)) || !j.verified(p).Verified {
				continue
			}
			d := deny(RuleOtherProject, fmt.Sprintf("%s of %s refused: it is the %s folder of another linked project (%s), "+
				"whose top folder and floor folders a shell call outside it may not delete or move", c.tool,
				quote(filepath.ToSlash(abs), 120), rel, quote(filepath.Base(p.root), 60)),
				"leave that project as it is; a person deletes or moves it")
			d.Target = target(rec, []string{abs})
			return d, rec
		}
		return allow(RuleNotProtected, ""), rec
	}
	var all []judged
	for _, p := range j.judges(chain) {
		if rel, ok := relTo(p.root, abs); ok {
			all = append(all, judged{abs: abs, chain: chain, p: p, rel: rel, fold: caseInsensitive(p.root)})
		}
	}
	// Every list that judges must read, and every glob on it must be one the guard reads: else nothing can be told.
	for _, x := range all {
		l := j.listsOf(x.p)
		if l.err != nil {
			return j.unread(c.tool, x, l), j.recordIn(x.chain)
		}
		for _, li := range l.all() {
			for _, g := range li.globs {
				if err := checkGlob(g); err != nil {
					return deny(RuleBadGlob, fmt.Sprintf("%s refused: the glob %s on %s is not a glob the guard reads (%s), so it "+
						"cannot tell whether the path is protected", c.tool, strconv.QuoteToASCII(g), li.name, oneLine(err.Error())),
						"leave the file as it is and tell the person: a person fixes that glob"), j.recordIn(x.chain)
				}
			}
		}
	}
	var hits []hit
	for _, x := range all {
		if f := floorOf(x.rel, x.fold, c.del); f != nil && f.Class != floorFree {
			next := ""
			if f.Class == RuleBonsaiFiles && strings.HasSuffix(f.Glob, ".md") {
				next = "leave the table as it is: bonsai check --write, run in the main checkout, rebuilds it from the task files and the log"
			}
			hits = append(hits, hit{x: x, rule: f.Class, on: "the floor of paths Bonsai guards in every project (" + f.Why + ")",
				glob: f.Glob, next: next})
		}
		for _, li := range j.listsOf(x.p).all() {
			for _, g := range li.globs {
				if matchGlob(g, x.rel, x.fold) {
					hits = append(hits, hit{x: x, rule: li.rule, on: li.name, glob: g})
					break
				}
			}
		}
	}
	// What no grant opens refuses first; then what a grant may open, person-only before protected.
	for _, rule := range []string{RuleBonsaiFiles, RuleNeverGranted} {
		for _, h := range hits {
			if h.rule == rule {
				return j.refuseHit(c, h, ""), j.recordIn(h.x.chain)
			}
		}
	}
	outcome := RuleNotProtected
	if len(hits) > 0 {
		h := hits[0]
		for _, x := range hits {
			if x.rule == RulePersonOnly {
				h = x
				break
			}
		}
		g := j.grantor(chain)
		if g == nil {
			t := j.verified(chain[0])
			d := deny(RuleMainNotVerified, hitOn(c, h)+", and no task can grant it here: its main checkout is not verified, as "+
				t.Why, t.Next+"; until then, leave the file as it is and tell the person")
			d.Target = target(j.recordIn(h.x.chain), []string{h.x.abs})
			return d, j.recordIn(h.x.chain)
		}
		rel, ok := relTo(g.root, abs)
		if ok {
			if yes, why := j.grantOf(g).grants(rel, caseInsensitive(g.root)); yes {
				outcome = RuleGranted
			} else {
				return j.refuseHit(c, h, why), j.recordIn(h.x.chain)
			}
		} else {
			return j.refuseHit(c, h, "no task can grant it"), j.recordIn(h.x.chain)
		}
	}
	for _, x := range all {
		if d, ok := j.tripwire(c, x); !ok {
			return d, j.recordIn(x.chain)
		}
	}
	return allow(outcome, ""), rec
}

// hitOn says what refused a guarded path: "<tool> of <path> refused: it is on <what> (<glob>)".
func hitOn(c change, h hit) string {
	shown := quote(h.x.rel, 120)
	if len(h.x.chain) > 1 {
		shown += " (in the project " + quote(filepath.Base(h.x.p.root), 60) + ")"
	}
	return fmt.Sprintf("%s of %s refused: it is on %s (%s)", c.tool, shown, h.on, strconv.QuoteToASCII(h.glob))
}

// refuseHit is the refusal of a guarded path: what guards it, and why nothing grants it (for a path a grant may open).
func (j *judge) refuseHit(c change, h hit, why string) Decision {
	on := hitOn(c, h)
	var d Decision
	switch h.rule {
	case RuleBonsaiFiles:
		next := h.next
		if next == "" {
			next = "leave it as it is: only Bonsai writes it, as a program"
		}
		d = deny(RuleBonsaiFiles, on+", one of Bonsai's own files, which no grant opens", next)
	case RuleNeverGranted:
		d = deny(RuleNeverGranted, on+", which no grant opens: Bonsai's own lines and the checkout's git live there",
			"leave it as it is and tell the person: a person changes it, or runs bonsai update")
	case RulePersonOnly:
		d = deny(RulePersonOnly, on+", which only a person changes or grants to a task, and "+why,
			"leave the file as it is and ask the person to make this change, or to grant it to the running task (its bonsai.allows)")
	default:
		d = deny(RuleProtected, on+", and "+why, "leave the file as it is and tell the person; a task whose bonsai.allows holds "+
			"the path may change it while it reads running")
	}
	d.Target = target(j.recordIn(h.x.chain), []string{h.x.abs})
	return d
}

// ownChain reports whether the session's own project is among the projects holding a path (or the main checkout it
// is a worktree of): rule 8's "the session's own project".
func (j *judge) ownChain(chain []*linkedProject) bool {
	for _, p := range chain {
		if j.sameProject(p, j.session) {
			return true
		}
	}
	return false
}

// judgedRoot is one folder a shell call's deletes and moves are judged in (5.3.2 note 8), with its guarded globs: what
// step 5.3.3's delete check meets a deleted folder against, without reading its tree.
type judgedRoot struct {
	Dir   string   // its top folder, absolute
	Globs []string // its guarded globs, folder-relative: the lists, the packs' declared paths, the floor's for deletes
	Whole bool     // everything in it is guarded (Bonsai's home)
	Own   bool     // the session's own project
}

// judgedRoots gives the folders a shell call's deletes are judged in (5.3.2 note 8): the session's own project with
// every guarded glob, its verified main checkout too when it is a worktree, Bonsai's home folders whole, and each
// other project on this machine's record whose folder still holds a bonsai.yaml, by its floor folders only (which
// may verifies before it refuses: a record that is not verified guards nothing). It reads this machine's records, so
// a caller asks for it only when a delete needs it.
func (j *judge) judgedRoots() []judgedRoot {
	var out []judgedRoot
	seen := []string{}
	add := func(r judgedRoot) {
		for _, s := range seen {
			if samePlace(s, r.Dir) {
				return
			}
		}
		seen = append(seen, r.Dir)
		out = append(out, r)
	}
	if s := j.session; s != nil {
		globs := floorGlobs(true)
		if l := j.listsOf(s); l.err == nil {
			for _, li := range l.all() {
				globs = append(globs, li.globs...)
			}
		}
		add(judgedRoot{Dir: s.root, Globs: globs, Own: true})
		if m := j.local(s).Main; !samePlace(m, s.root) {
			add(judgedRoot{Dir: m, Globs: globs})
		}
	}
	for _, forms := range j.homes() {
		add(judgedRoot{Dir: forms[0], Whole: true})
	}
	if j.home != "" {
		for _, kr := range workspace.MachineRecords(j.home) {
			dir := filepath.Clean(filepath.FromSlash(kr.Record.Path))
			if _, err := os.Lstat(filepath.Join(dir, workspace.ConfigFile)); err == nil {
				add(judgedRoot{Dir: dir, Globs: append([]string(nil), floorFolders...)})
			}
		}
	}
	return out
}
