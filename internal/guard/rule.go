package guard

// What the guard judges a file tool's path by, and where it reads that from (design/plan-5.md, 5.3.1).
//
// The projects holding a path (contract §13: "the project is the one holding the path, for the guard"; 5.3.1 note 1):
// for each form of the path (paths.go), every folder at or above it that holds bonsai.yaml, past any .git, to the
// volume's root. The nearest one judges alone when it is a verified checkout (workspace.VerifyMain; Claude Code's own
// worktrees sit inside the main checkout, under .claude/worktrees/, where the main checkout's .claude/** would refuse
// every edit, and are verified there). When it is not (a folder with no .git of its own, a bonsai.yaml someone
// planted, a nested repository never linked on this machine), the lists of every project holding the path apply,
// each read against its own folder: so a bonsai.yaml with empty lists in docs/, or a .git placed there, narrows
// nothing. A path in another linked project is judged by that project's lists. A path in no linked project is
// allowed (outside-project).
//
// The lists a project judges by: the person_only and protected globs of its bonsai.yaml, and each pack's declared
// protected paths in its lock (workspace.ReadConfig and workspace.ReadDeclared, both lean: 5.3.1 note 3). A worktree
// whose main checkout is verified judges by the main checkout's bonsai.yaml and lock: its own copies change nothing
// (rung 0 judges a branch's change to them, step 5.4). Any other checkout judges by its own. The person_only list is
// a part of the protected one (spec §6), so a path on it alone is protected as well.
//
// Grants come with step 5.3.2, and only from the nearest verified project holding the path; this build grants none.
// When no project holding a refused path is verified, the refusal is main-not-verified and names the person's step
// that verifies it (a person runs bonsai update in the main checkout), since no task's grant could count there.
//
// Where a record goes (5.3.1 note 4): the nearest verified project holding the path; else the session's own project
// when it holds the path; else the nearest that is a checkout (never a folder above it someone placed a bonsai.yaml
// in). A verified worktree's records go to its main checkout's .bonsai/local/ (local). A call with no path to judge (a
// shell or another tool) is recorded in the session's own project (CLAUDE_PROJECT_DIR's), as is a path outside every
// project. When neither the path's project nor the session's is linked, the guard allows and records nothing (spec §7).

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// judge is one guard run's view of the linked projects its call touches: each project's files read once, and its
// main checkout verified only when an answer needs it (5.3.1 note 5: this machine's records are read only then).
type judge struct {
	home     string                    // the Bonsai home, "" when there is none
	session  *linkedProject            // the session's own project, nil when CLAUDE_PROJECT_DIR is in none
	projects map[string]*linkedProject // by the folder holding bonsai.yaml, as found
	configs  map[string]*config        // bonsai.yaml by folder: a project's own, or a worktree's main checkout's
	holding  map[string]bool           // whether a folder holds bonsai.yaml, looked up once (hookin.Holding)
}

// linkedProject is one folder holding bonsai.yaml: a project, linked or planted.
type linkedProject struct {
	root   string              // the folder, absolute and clean
	layout workspace.GitLayout // its .git, read with no git process
	own    *config             // its own bonsai.yaml
	trust  *workspace.Trust    // nil until asked (verified)
	lists  *lists              // nil until asked (listsOf)
}

// config is one bonsai.yaml as the guard reads it (workspace.LoadConfig, lean), or why it does not read.
type config struct {
	cfg *workspace.Config
	err error
}

// lists are what a project judges a path by: a bonsai.yaml's lists and its lock's declared protected paths. err is
// set, with rule, when they cannot be read: every path they would judge is then refused.
type lists struct {
	from     string // the folder they were read from: the project's own, or its verified main checkout
	cfg      *workspace.Config
	declared []workspace.PackDeclared
	rule     string // bad-config or bad-lock, with err
	err      error
}

func newJudge(home string) *judge {
	return &judge{home: home, projects: map[string]*linkedProject{}, configs: map[string]*config{}, holding: map[string]bool{}}
}

// project gives the project whose bonsai.yaml lies in root, reading its layout and its own bonsai.yaml once.
func (j *judge) project(root string) *linkedProject {
	if p, ok := j.projects[root]; ok {
		return p
	}
	p := &linkedProject{root: root, layout: workspace.ReadGitLayout(root), own: j.config(root)}
	j.projects[root] = p
	return p
}

// config reads the bonsai.yaml in folder once.
func (j *judge) config(folder string) *config {
	if c, ok := j.configs[folder]; ok {
		return c
	}
	cfg, err := workspace.LoadConfig(folder)
	c := &config{cfg: cfg, err: err}
	j.configs[folder] = c
	return c
}

// id is the project's own workspace id, "" when its bonsai.yaml does not read.
func (p *linkedProject) id() string {
	if p.own.cfg == nil {
		return ""
	}
	return p.own.cfg.ID
}

// verified says whether the project's main checkout is verified (workspace.VerifyMain), asking once.
func (j *judge) verified(p *linkedProject) workspace.Trust {
	if p.trust != nil {
		return *p.trust
	}
	t := workspace.Trust{Why: "its bonsai.yaml cannot be read", Next: "a person fixes bonsai.yaml (bonsai check names the line)"}
	if p.own.err == nil {
		mainID := ""
		if p.layout.Worktree() {
			if c := j.config(p.layout.Main); c.cfg != nil {
				mainID = c.cfg.ID
			}
		}
		t = workspace.VerifyMain(j.home, p.layout, p.id(), mainID)
	}
	p.trust = &t
	return t
}

// listsOf reads the lists a project judges by, once: its verified main checkout's for a worktree, else its own.
func (j *judge) listsOf(p *linkedProject) *lists {
	if p.lists != nil {
		return p.lists
	}
	l := &lists{from: p.root}
	c := p.own
	if p.layout.Worktree() && j.verified(p).Verified {
		l.from, c = p.layout.Main, j.config(p.layout.Main)
	}
	if c.err != nil {
		l.rule, l.err = RuleBadConfig, c.err
	} else {
		l.cfg = c.cfg
		declared, err := workspace.LoadDeclared(l.from)
		switch {
		case err == nil:
			l.declared = declared
		case errors.Is(err, fs.ErrNotExist) && len(c.cfg.Packs) == 0:
			// No packs and no lock: nothing is declared.
		default:
			l.rule, l.err = RuleBadLock, err
		}
	}
	p.lists = l
	return l
}

// local is where a record decided in this project goes: a verified worktree's main checkout's .bonsai/local/, else its
// own (as workspace.FindLocal answers every other writer).
func (j *judge) local(p *linkedProject) workspace.Local {
	l := workspace.Local{Root: p.root, Main: p.root, Branch: p.layout.Branch()}
	if p.layout.Worktree() && j.verified(p).Verified {
		l.Main = p.layout.Main
	}
	return l
}

// chain gives the projects holding an absolute path, nearest first.
func (j *judge) chain(abs string) []*linkedProject {
	var out []*linkedProject
	for _, root := range hookin.Holding(filepath.Dir(abs), j.holding) {
		out = append(out, j.project(root))
	}
	return out
}

// judges are the projects whose lists judge a path held by chain: the nearest alone when it is a verified checkout,
// else every one.
func (j *judge) judges(chain []*linkedProject) []*linkedProject {
	if len(chain) > 1 && j.verified(chain[0]).Verified {
		return chain[:1]
	}
	return chain
}

// grantor is the nearest verified project holding a path, the only one a grant may come from (step 5.3.2); nil when
// none is verified.
func (j *judge) grantor(chain []*linkedProject) *linkedProject {
	for _, p := range chain {
		if j.verified(p).Verified {
			return p
		}
	}
	return nil
}

// recordIn is the project a decision about a path held by chain is recorded in (this file's comment): the nearest
// verified one; else the session's own project when it holds the path; else the nearest checkout whose bonsai.yaml
// reads, else the nearest whose bonsai.yaml reads; else the session's project (nil when it has none).
func (j *judge) recordIn(chain []*linkedProject) *linkedProject {
	if len(chain) == 1 && chain[0].own.cfg != nil {
		return chain[0] // verified or not, the same project: local says which local/ it writes
	}
	if g := j.grantor(chain); g != nil {
		return g
	}
	for _, p := range chain {
		if p == j.session {
			return p
		}
	}
	var readable *linkedProject
	for _, p := range chain {
		switch {
		case p.own.cfg == nil:
		case p.layout.GitDir != "":
			return p
		case readable == nil:
			readable = p
		}
	}
	if readable != nil {
		return readable
	}
	return j.session
}

// list is one of the lists a path is judged against.
type list struct {
	name  string // the list as a refusal names it
	rule  string
	globs []string
}

// all gives a project's lists in the order they are tried: person_only first, so a path on both is refused as
// person-only; then protected; then each pack's declared protected paths, in the lock's order.
func (l *lists) all() []list {
	out := []list{
		{"bonsai.yaml's person_only list", RulePersonOnly, l.cfg.PersonOnly},
		{"bonsai.yaml's protected list", RuleProtected, l.cfg.Protected},
	}
	for _, d := range l.declared {
		out = append(out, list{"the protected paths the pack " + strconv.QuoteToASCII(d.Pack) + " declares", RuleProtected, d.Protected})
	}
	return out
}

// judged is one form of a path, a project judging it, and the path relative to that project's folder.
type judged struct {
	abs   string
	chain []*linkedProject
	p     *linkedProject
	rel   string
}

// decide judges one call. base is the folder a relative path is read against (the session's project, else
// CLAUDE_PROJECT_DIR). It gives the decision and the project its record goes in (nil: no record).
func (j *judge) decide(in *hookin.Payload, base string) (Decision, *linkedProject) {
	if !hookin.FileTool(in.Tool) {
		if in.Tool == "Bash" || in.Tool == "PowerShell" {
			return allow(RuleShell, ""), j.session
		}
		return allow(RuleOtherTool, ""), j.session
	}
	forms, err := pathForms(base, in.Cwd, in.Path)
	if err != nil {
		return deny(RuleBadPath, fmt.Sprintf("%s of %s refused: on Windows the guard cannot tell where the path leads (%s), "+
			"so it cannot tell whether it is protected", in.Tool, quote(filepath.ToSlash(in.Path), 120), oneLine(err.Error())),
			"write to the file by its plain path inside the project (a drive letter and folders), or tell the person"), j.session
	}
	var all []judged
	var first []*linkedProject // the chain of the first form a project holds: an allow is recorded there
	for _, abs := range forms {
		chain := j.chain(abs)
		if len(chain) == 0 {
			continue
		}
		if first == nil {
			first = chain
		}
		for _, p := range j.judges(chain) {
			if rel, ok := relTo(p.root, abs); ok {
				all = append(all, judged{abs: abs, chain: chain, p: p, rel: rel})
			}
		}
	}
	if first == nil {
		return allow(RuleOutside, ""), j.session
	}
	// Every list that judges must read, and every glob on it must be one the guard reads: else nothing can be told.
	for _, x := range all {
		l := j.listsOf(x.p)
		if l.err != nil {
			return j.unread(in, x, l), j.recordIn(x.chain)
		}
		for _, li := range l.all() {
			for _, g := range li.globs {
				if err := checkGlob(g); err != nil {
					return deny(RuleBadGlob, fmt.Sprintf("%s refused: the glob %s on %s is not a glob the guard reads (%s), so it "+
						"cannot tell whether the path is protected", in.Tool, strconv.QuoteToASCII(g), li.name, oneLine(err.Error())),
						"leave the file as it is and tell the person: a person fixes that glob"), j.recordIn(x.chain)
				}
			}
		}
	}
	// The lists, in their order (all), across every form and project: the first glob that matches refuses.
	for i := 0; ; i++ {
		tried := false
		for _, x := range all {
			lists := j.listsOf(x.p).all()
			if i >= len(lists) {
				continue
			}
			tried = true
			for _, g := range lists[i].globs {
				if matchGlob(g, x.rel, caseInsensitive(x.p.root)) {
					rec := j.recordIn(x.chain)
					d := j.refuse(in, x, lists[i], g)
					d.Target = target(rec, []string{x.abs})
					return d, rec
				}
			}
		}
		if !tried {
			break
		}
	}
	rec := j.recordIn(first)
	return allow(RuleNotProtected, target(rec, forms)), rec
}

// refuse is the refusal of a path on a list: protected or person-only when a verified project holds it (a grant could
// come from there, from step 5.3.2), else main-not-verified, naming the person's step that verifies it.
func (j *judge) refuse(in *hookin.Payload, x judged, li list, glob string) Decision {
	shown := quote(x.rel, 120)
	if len(x.chain) > 1 {
		shown += " (in the project " + quote(filepath.Base(x.p.root), 60) + ")"
	}
	on := fmt.Sprintf("%s of %s refused: it is on %s (%s)", in.Tool, shown, li.name, strconv.QuoteToASCII(glob))
	switch {
	case j.grantor(x.chain) == nil:
		t := j.verified(x.chain[0])
		return deny(RuleMainNotVerified, on+", and no task can grant it here: its main checkout is not verified, as "+t.Why,
			t.Next+"; until then, leave the file as it is and tell the person")
	case li.rule == RulePersonOnly:
		return deny(RulePersonOnly, on+", so only a person changes it", "leave the file as it is and ask the person to make this change")
	}
	return deny(RuleProtected, on+", and this build gives no task the right to change it",
		"leave the file as it is and tell the person; a person changes it (a task's grants come with Bonsai's step 5.3)")
}

// unread is the refusal of a path whose project's lists cannot be read.
func (j *judge) unread(in *hookin.Payload, x judged, l *lists) Decision {
	file := "bonsai.yaml"
	if l.rule == RuleBadLock {
		file = workspace.LockFile
	}
	if len(x.chain) > 1 || l.from != x.p.root {
		file += " (in " + quote(filepath.Base(l.from), 60) + ")"
	}
	next := "leave the change and tell the person: a person fixes " + file + " (bonsai check names the problem)"
	var we *workspace.Error
	if errors.As(l.err, &we) && we.Next != "" {
		next = "leave the change and tell the person: " + we.Next
	}
	return deny(l.rule, fmt.Sprintf("%s refused: %s cannot be read (%s), so the guard cannot tell what is protected",
		in.Tool, file, oneLine(l.err.Error())), oneLine(next))
}

// target is the path a record names: relative to the project the record goes in, for the first form it holds; "" for
// none.
func target(rec *linkedProject, forms []string) string {
	if rec == nil {
		return ""
	}
	for _, abs := range forms {
		if rel, ok := relTo(rec.root, abs); ok {
			return rel
		}
	}
	return ""
}

// relTo gives abs relative to root, forward slashes, when root holds it (and it is not root itself).
func relTo(root, abs string) (string, bool) {
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if rel == "." || rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false
	}
	return rel, true
}
