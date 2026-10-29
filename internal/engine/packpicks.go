package engine

// The pack command's second stage (T-6009's plan, steps 9-16; the director's eight picks of 28 Oct, task T-6011): what
// the engine works out for bonsai pack beyond the core, each read from the home's clones or the lock, never the
// network unless named:
//
//   - PackChanges (pick 1): a moved pack's commit titles between its locked commit and the new one, limited to the
//     pack's folder, at most MaxPackChanges and a count of the rest, from the clone fetch already made (which keeps
//     full history). A read that fails says why and never refuses. PackChangesTo fetches a tag first (pack outdated
//     --changes: the network).
//   - PacksBehindLine (pick 2): the start hook's one line naming the packs the home's cache says are behind; offline.
//   - Downgrade (pick 4): a move to a release tag whose version sorts below the locked one; a commit, a branch or a
//     tag with no version has no order and is never a downgrade.
//   - Needs (pick 5, the director's "Refuse, change the rule", 28 Oct): before any write by pack set, add or set
//     --all, each pack the plan moves or adds must be written in a format this Bonsai reads (bonsai.pack/1) and need
//     no newer Claude Code than this machine runs; refused pack-needs. A Claude Code version that cannot be read is
//     said in the preview, never refused. Only bonsai pack refuses for a need: status, check and update keep warning.
//   - OldMarketHint (pick 6, the director's "Hint, don't remove"): after a write that moved the marketplace name, the
//     old marketplace still registered on this machine and the command that removes it, a person's to run once no
//     other checkout uses it. Nothing removes it.
//   - CommitLines (pick 7): after a write, the two commands that commit the files written, one per line, never
//     joined by && (PowerShell 5.1 has none).

import (
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// MaxPackChanges is the most commit titles a moved pack lists; the rest are counted.
const MaxPackChanges = 30

// packChanges reads a pack's commit titles from..to in the home's clone of source, newest first, limited to the
// pack's folder p ("" for the repository's top): at most MaxPackChanges, and how many more there are.
func (c cache) packChanges(source, p, from, to string) (titles []string, more int, err error) {
	for _, commit := range []string{from, to} {
		if !commitPattern().MatchString(commit) || !c.has(source, commit) {
			return nil, 0, fmt.Errorf("the commit %s is not in this machine's clone of %s", short(commit), source)
		}
	}
	args := []string{"log", "--no-color", "--format=%s", from + ".." + to}
	if p != "" {
		args = append(args, "--", path.Clean(p))
	}
	out, err := git(c.dir(source), nil, args...)
	if err != nil {
		return nil, 0, err
	}
	titles = []string{}
	for _, l := range strings.Split(strings.ReplaceAll(string(out), "\r\n", "\n"), "\n") {
		if l = strings.TrimSpace(l); l == "" {
			continue
		}
		if len(titles) == MaxPackChanges {
			more++
			continue
		}
		titles = append(titles, ascii(l))
	}
	return titles, more, nil
}

// packChange is one pack's changes as bonsai.changes/1's pack_changes holds them: titles [] and why set when the
// read failed.
func packChange(id string, titles []string, more int, err error) format.ChangesPackChange {
	pc := format.ChangesPackChange{Pack: id, Titles: titles, More: int64(more)}
	if err != nil {
		why := ascii(err.Error())
		pc.Titles, pc.More, pc.Why = []string{}, 0, &why
	}
	if pc.Titles == nil {
		pc.Titles = []string{}
	}
	return pc
}

// FillPackChanges sets the plan's PackChanges: for each pack it moves from one commit to another, in the plan's
// order, its commit titles between the two (packChanges), from the home's clone alone. A pack added or taken out has
// none. A read that fails is kept with its why; it never refuses.
func (p *Plan) FillPackChanges() {
	p.PackChanges = []format.ChangesPackChange{}
	if p.Config == nil {
		return
	}
	folders := map[string]string{}
	for _, r := range p.Config.Packs {
		folders[r.ID] = r.Path
	}
	c := cache{home: p.Home}
	for _, m := range p.Packs {
		if m.From == "" || m.To == "" || m.From == m.To {
			continue
		}
		titles, more, err := c.packChanges(m.Source, folders[m.ID], m.From, m.To)
		p.PackChanges = append(p.PackChanges, packChange(m.ID, titles, more, err))
	}
}

// PackChangesTo fetches source at ref into the home's cache (the network) and gives the pack's commit titles from
// the locked commit from to it, limited to its folder p, as pack_changes holds them.
func PackChangesTo(home, id, source, p, from, ref string) format.ChangesPackChange {
	c := cache{home: home}
	to, err := c.fetch(source, ref)
	if err != nil {
		return packChange(id, nil, 0, errors.New(Unexpected(err).What))
	}
	titles, more, err := c.packChanges(source, p, from, to)
	return packChange(id, titles, more, err)
}

// Downgrade reports whether moving a pack locked at version to ref goes back: ref is a release tag (a prefix and
// vX.Y.Z, NewerTags' form) whose version sorts below the locked one. ordered is false when either has no version
// (a 40-character commit, a branch, a tag with no version number): such a move has no order and is never a
// downgrade.
func Downgrade(version, ref string) (down, ordered bool) {
	locked, _, ok := ParseVersion(version)
	m := releaseTag().FindStringSubmatch(ref)
	if !ok || m == nil || commitPattern().MatchString(ref) {
		return false, false
	}
	v, _, ok := ParseVersion(m[2])
	if !ok {
		return false, false
	}
	return older(v, locked), true
}

// Needs is what bonsai pack holds each pack it moves or adds to before anything is written (Request.Needs; this
// file's comment): Claude gives claude --version's first line (ErrNoClaude when it is not on the PATH); it is asked
// once, and only when a pack declares needs.claude_code. Code is the error word a refusal carries, the caller's
// (cmd/bonsai's pack word gives pack-needs, format.ErrorWords): the engine's own pack-<word> literals are check
// --pack's words (format.PackCheckWords).
type Needs struct {
	Claude func() (string, error)
	Code   string

	asked   bool
	version [3]int
	text    string
	why     string // why the version could not be read; "" when it was
}

// claude reads Claude Code's version once.
func (n *Needs) claude() {
	if n.asked {
		return
	}
	n.asked = true
	if n.Claude == nil {
		n.why = "Claude Code's version was not read"
		return
	}
	st := JudgeClaude(nil, n.Claude)
	if st.State == "unknown" {
		n.why = st.Why
		return
	}
	n.version, _, _ = ParseVersion(st.Version)
	n.text = st.Version
}

// format refuses a pack whose bonsai/pack.yaml at commit names a format this Bonsai does not read. A pack.yaml that
// is missing or does not read is left to packAt, which refuses it as bad-pack.
func (n *Needs) format(c cache, ref workspace.PackRef, commit string) error {
	name := workspace.PackFile
	if ref.Path != "" {
		name = path.Clean(ref.Path) + "/" + name
	}
	raw, err := git(c.dir(ref.Source), nil, "cat-file", "blob", commit+":"+name)
	if err != nil {
		return nil
	}
	r := reader.ReadYAML(raw)
	var s string
	switch {
	case r.Outcome == reader.Refused && r.Refusal.Code == reader.CodeFormatTooNew:
		// A newer major of the format (bonsai.pack/2): the reader stops at its format line, which its message quotes.
		s = r.Refusal.Message
		if _, rest, ok := strings.Cut(s, `"`); ok {
			if name, _, ok := strings.Cut(rest, `"`); ok {
				s = name
			}
		}
	case r.Outcome == reader.Accepted && r.Value != nil:
		f, ok := r.Value.Get("format")
		text, isText := f.(string)
		if !ok || !isText || text == workspace.PackFormat {
			return nil
		}
		s = text
	default:
		return nil
	}
	return errorf(n.Code, ExitState, "a person installs a Bonsai that reads "+ascii(s)+" (bonsai status --full names the newest release), "+
		"or picks a version of the pack written in "+workspace.PackFormat+"; then run the command again",
		"the pack %s at %s is written in %s, a pack format this Bonsai does not read (it reads %s): nothing was written",
		ref.ID, short(commit), ascii(s), workspace.PackFormat).whose("person")
}

// need refuses a pack whose needs.claude_code is newer than the Claude Code this machine runs; a version that cannot
// be read refuses nothing (Plan.NeedsUnread says so).
func (n *Needs) need(pd *PackData, p *Plan) error {
	if pd.Manifest == nil || pd.Manifest.ClaudeCode == "" {
		return nil
	}
	want, text, ok := ParseVersion(pd.Manifest.ClaudeCode)
	if !ok {
		return nil
	}
	n.claude()
	if n.why != "" {
		p.NeedsUnread = append(p.NeedsUnread, fmt.Sprintf("the pack %s needs Claude Code %s, and %s, so it was not compared", pd.Ref.ID, text, n.why))
		return nil
	}
	if !older(n.version, want) {
		return nil
	}
	return errorf(n.Code, ExitState, "a person updates Claude Code on this machine (claude update), or picks a version of the pack "+
		"that needs no newer one; then run the command again",
		"the pack %s at %s needs Claude Code %s, and this machine runs %s: nothing was written", pd.Ref.ID, short(pd.Commit), text, n.text).whose("person")
}

// PacksBehindLine is the start hook's line naming the packs behind (pick 2): each locked pack bonsai.yaml names
// whose newer state from the home's cache is newer, as "<id> <ref> -> <tag>", in the lock's order; "" when none is
// behind, or there is no lock or no cache. It never reads a source.
func PacksBehindLine(home, root string, cfg *workspace.Config) string {
	if cfg == nil || home == "" {
		return ""
	}
	lock, err := workspace.LoadLock(root)
	if err != nil || lock == nil {
		return ""
	}
	cache, err := LoadPackCache(home)
	if err != nil {
		return ""
	}
	var behind []string
	for _, lp := range lock.Packs {
		for _, r := range cfg.Packs {
			if r.ID != lp.ID {
				continue
			}
			if n := PackNewerOf(cache, lp.Source, lp.ID, r.Ref, r.Path, lp.Version); n != nil && n.State == PackNewer {
				behind = append(behind, ascii(lp.ID+" "+r.Ref+" -> "+n.Tag))
			}
		}
	}
	if len(behind) == 0 {
		return ""
	}
	return "Packs behind: " + strings.Join(behind, ", ") + " (bonsai pack outdated)"
}

// OldMarket is the result word of the old marketplace's hint among the plugin step's results (bonsai.changes/1's
// plugins[].result, an open list).
const OldMarket = "old-marketplace"

// OldMarketHint is the plugin step's hint after a write that moved the workspace's marketplace name (pick 6): a result
// naming the old one (the marketplace of the lock as read, OldMarket), which a session in the checkout registered on
// this machine and Bonsai leaves registered, and the command that removes it, a person's step to take once no other
// checkout of the project, still on the old lock, uses it. nil when the name did not move, or the plugin step does not
// run (no PluginCLI). It asks Claude Code nothing (the step's calls stay as they were) and removes nothing.
func OldMarketHint(p *Plan, cli PluginCLI) *PluginResult {
	if cli == nil || p.OldMarket == "" || p.oldLock == nil || len(p.oldLock.Packs) == 0 || p.lock == nil || len(p.lock.Packs) == 0 ||
		p.OldMarket == lockMarket(p.Config, p.lock) {
		return nil
	}
	lp := p.oldLock.Packs[0]
	return &PluginResult{Pack: lp.ID, Plugin: PluginID(lp.ID, p.OldMarket), Commit: lp.Commit, Result: OldMarket,
		Message: ascii("the workspace's plugin marketplace is now " + lockMarket(p.Config, p.lock) + "; the old one, " + p.OldMarket +
			", stays registered on this machine (claude plugin marketplace list shows it) and nothing removed it, since a checkout of " +
			"this project still on the old lock uses it"),
		Next: "once no other checkout of this project uses it, run: claude plugin marketplace remove " + p.OldMarket, Who: "person"}
}

// CommitLines are the commands that commit what the plan wrote (pick 7): git add with every file written, in the
// plan's order and the lock last, forward slashes (and bonsai.yaml first when update wrote the lock from a hand edit
// of it), then git commit with a message naming each pack moved; one per line, never joined by && (PowerShell 5.1
// has none). nil when nothing was written.
func (p *Plan) CommitLines() []string {
	var paths []string
	config := false
	for _, f := range p.Files {
		if f.Writes() && f.Result != Unchanged {
			paths = append(paths, ShellArg(f.Path))
			config = config || f.Path == workspace.ConfigFile
		}
	}
	if p.LockWrite || p.LockRemove {
		paths = append(paths, workspace.LockFile)
	}
	if len(paths) == 0 {
		return nil
	}
	// The lock written follows bonsai.yaml as it stands, a hand edit's among them (update after one): the two are
	// committed together, bonsai.yaml first.
	if !config && p.LockWrite && p.Command == "update" {
		paths = append([]string{workspace.ConfigFile}, paths...)
	}
	refs := map[string]string{}
	if p.Config != nil {
		for _, r := range p.Config.Packs {
			refs[r.ID] = r.Ref
		}
	}
	var moves []string
	for _, m := range p.Packs {
		switch {
		case m.To == "":
			moves = append(moves, m.ID+" taken out")
		case m.From != m.To:
			moves = append(moves, m.ID+" to "+refs[m.ID])
		}
	}
	msg := "bonsai: " + p.Command
	if len(moves) > 0 {
		msg = "bonsai: " + strings.Join(moves, ", ")
	}
	msg = ascii(msg)
	quoted := `"` + msg + `"`
	if strings.ContainsAny(msg, "\"$`\\") {
		quoted = "'" + strings.ReplaceAll(msg, "'", "''") + "'" // nothing in it either shell would read
	}
	return []string{"git add " + strings.Join(paths, " "), "git commit -m " + quoted}
}
