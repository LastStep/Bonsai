package workspace

// The roles in force and their limits (set 8; plan-5 5.5.0 note 8; step 5.5, "Role rules and stall stops, written
// once"): the one function every reader of a role's limits uses (the guard from 5.5.11, bonsai message from 5.5.10,
// status --json's roles, check's role findings), so each reads the same numbers. It reads only the copies it is
// handed, never a file: the guard hands it the lock and bonsai.yaml as committed at the main checkout's HEAD, status
// the same (as committed where HEAD holds them), check the files as they stand.
//
// Where the numbers come from, for each of a role's three limits (format.RoleLimits: minutes, actions, red_runs):
//   - the pack's default and max, as its pack.yaml declares them, copied into the lock's declares (roles);
//   - the project's own, under the pack's entry in bonsai.yaml (packs[].roles: a role's name to any of its limits,
//     each with its default, its max or both; what is left out keeps the pack's; null keeps the pack's for every role).
//
// In force: the default bonsai.yaml's where it sets one, else the pack's; the max bonsai.yaml's where it sets one,
// else the pack's; null (nil here) is no limit. A default above its max (no limit is above any number) has the max
// hold: the default in force is then the max, from where the max came (check names it, role-default-over-max). A
// brief's value, up to the max in force, is the brief's reader's (bonsai message, the guard), never this function's.
// Raising a default or a max is never a loosening of the checks (the director, 14 Oct): the settings record's compare
// (checks.go) reads neither roles nor messages.
//
// Two packs declaring one role's name each keep their own (a role is <pack>:<name>, as Claude Code names it), so a
// reader looks a role up by its pack and name. A lock written before set 8 declares no role.
//
// A role's limits are a closed list (format.RoleLimits: minutes, actions, red_runs; contract §2.2: a closed list never
// grows inside a major), so a name under packs[].roles.<role> that is none of them (minutez) is no newer field to keep
// and skip but a mistake: bonsai.yaml with one does not read for the roles, and every reader says so naming the file,
// its line and the step that fixes it (RolesError), never the lock. check finds each such name (UnknownLimits) and
// still raises every other role finding, on the limits with those names left out (KnownLimits). A field inside a limit
// other than default and max is an unknown field of a known major (contract §2.2): kept in the file, never read.

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/reader"
	"github.com/LastStep/Bonsai/internal/schema"
)

// Where a limit's default or max came from: the closed list of bonsai.status/1's roles[].<limit>.default_from and
// max_from (its schema is the list's home).
const (
	LimitFromPack   = "pack"        // the pack's pack.yaml, through the lock's declares
	LimitFromConfig = "bonsai.yaml" // the project's own, packs[].roles
)

// LimitInForce is one limit of a role in force.
type LimitInForce struct {
	Default     *int64 // what an agent of the role gets when its brief names none; nil for no limit
	DefaultFrom string // LimitFromPack or LimitFromConfig: where Default came from
	Max         *int64 // the most a brief may give it; nil for no limit
	MaxFrom     string // where Max came from

	// OverMax: the default as set (SetDefault, from SetDefaultFrom) is above the max, which then holds, so Default is
	// the max and DefaultFrom the max's source. check names it (role-default-over-max).
	OverMax        bool
	SetDefault     *int64
	SetDefaultFrom string
}

// RoleInForce is one role a linked pack declares, as the lock holds it, with its three limits in force.
type RoleInForce struct {
	Pack    string          // the declaring pack's id
	Role    format.PackRole // the role as the pack declares it (name, kind, session, writes, max_turns, the pack's limits)
	Minutes LimitInForce
	Actions LimitInForce
	RedRuns LimitInForce
}

// Limit gives one of the role's limits in force by its name in format.RoleLimits (minutes, actions, red_runs); ok
// false for another name.
func (r RoleInForce) Limit(name string) (LimitInForce, bool) {
	switch name {
	case "minutes":
		return r.Minutes, true
	case "actions":
		return r.Actions, true
	case "red_runs":
		return r.RedRuns, true
	}
	return LimitInForce{}, false
}

// SetLimit is one limit as bonsai.yaml sets it for a role: each of its default and max set or not (left out keeps the
// pack's), and when set, a whole number or nil for no limit.
type SetLimit struct {
	HasDefault, HasMax bool
	Default, Max       *int64
}

// RolesInForce gives every role the lock's packs declare, in the lock's order of packs and each pack's order of
// roles, with each of its limits in force from the pack's declaration (the lock) and the project's own (cfg, bonsai.yaml
// read in full; nil reads as setting none). Its error, a *RolesError, names the file that does not read for a pack:
// the lock's declares, or the pack's roles in bonsai.yaml (a name that is not one of a role's limits among them); a
// role bonsai.yaml sets that its pack does not declare is left out here (check names it, role-undeclared). A nil lock
// gives none. RolesInForceOf gives the same with bonsai.yaml's line in its error.
func RolesInForce(lock *Lock, cfg *format.Workspace) ([]RoleInForce, error) {
	if lock == nil {
		return nil, nil
	}
	set := map[string]schema.Object{} // a pack's id to its roles in bonsai.yaml
	index := map[string]int{}         // and to its entry's index there
	if cfg != nil {
		for i, p := range cfg.Packs {
			set[p.ID], index[p.ID] = p.Roles, i
		}
	}
	var out []RoleInForce
	for _, lp := range lock.Packs {
		d, err := lp.Declared()
		if err != nil {
			return nil, &RolesError{File: LockFile, Pack: lp.ID, Err: err}
		}
		own, err := ConfigRoles(set[lp.ID])
		if err != nil {
			re := &RolesError{File: ConfigFile, Pack: lp.ID, Index: index[lp.ID], Err: err}
			var pe *rolePathError
			if errors.As(err, &pe) {
				re.Path, re.Unknown = pe.path, pe.unknown
			}
			return nil, re
		}
		for _, r := range d.Roles {
			ri := RoleInForce{Pack: lp.ID, Role: r}
			mine := own[r.Name]
			ri.Minutes = inForce(r.Minutes, mine["minutes"])
			ri.Actions = inForce(r.Actions, mine["actions"])
			ri.RedRuns = inForce(r.RedRuns, mine["red_runs"])
			out = append(out, ri)
		}
	}
	return out, nil
}

// RolesInForceOf is RolesInForce from bonsai.yaml as a Config holds it (c, nil for none), its error's Line filled from
// the file's map (c.Doc).
func RolesInForceOf(lock *Lock, c *Config) ([]RoleInForce, error) {
	var full *format.Workspace
	if c != nil {
		full = c.Full
	}
	in, err := RolesInForce(lock, full)
	var re *RolesError
	if errors.As(err, &re) && re.File == ConfigFile && c != nil {
		re.Line = ConfigLine(c.Doc, re.Index, re.Path)
	}
	return in, err
}

// RolesError is RolesInForce's error: the file that does not read for a pack's roles, and where.
type RolesError struct {
	File    string   // LockFile: the lock's declares for the pack; ConfigFile: the pack's entry's roles in bonsai.yaml
	Pack    string   // the pack's id
	Index   int      // ConfigFile: the pack's entry's index in bonsai.yaml's packs
	Path    []string // ConfigFile: the keys under the entry's roles down to the value that does not read (a role, a limit)
	Unknown bool     // ConfigFile: Path's last key is not one of a role's limits (format.RoleLimits, a closed list)
	Line    int      // ConfigFile: that key's line in bonsai.yaml where the file's map was at hand (RolesInForceOf), else 0
	Err     error    // what does not read
}

func (e *RolesError) Unwrap() error { return e.Err }

func (e *RolesError) Error() string {
	if e.File == LockFile {
		return fmt.Sprintf("the lock's declares for the pack %s cannot be read: %v", e.Pack, e.Err)
	}
	if e.Unknown {
		return fmt.Sprintf("%s sets %s for the role %s under the pack %s (%s), which is not one of a role's limits (%s)", e.where(),
			e.Path[len(e.Path)-1], e.Path[0], e.Pack, e.at(), limitNames(", "))
	}
	return fmt.Sprintf("%s: the pack %s's roles (%s): %v", e.where(), e.Pack, e.at(), e.Err)
}

// Next is the step that fixes it: the pack's declares written again (a person's update), or bonsai.yaml's line
// changed (by whoever may change bonsai.yaml).
func (e *RolesError) Next() string {
	if e.File == LockFile {
		return "the lock is Bonsai's to write: a person runs bonsai update in the main checkout, which writes each pack's " +
			"declares again, and commits the lock"
	}
	if e.Unknown {
		return fmt.Sprintf("in %s, write %s in place of %s, or take it out", e.where(), limitNames(" or "), e.Path[len(e.Path)-1])
	}
	return fmt.Sprintf("in %s, give %s as bonsai check --schema bonsai.workspace describes a role's limits", e.where(), e.at())
}

// where is bonsai.yaml with its line where it is known.
func (e *RolesError) where() string {
	if e.Line > 0 {
		return ConfigFile + " line " + strconv.Itoa(e.Line)
	}
	return ConfigFile
}

// at is the value's place in bonsai.yaml: packs[<index>].roles, then its path.
func (e *RolesError) at() string {
	return strings.Join(append([]string{fmt.Sprintf("packs[%d].roles", e.Index)}, e.Path...), ".")
}

// limitNames are format.RoleLimits' words, the last two joined by and.
func limitNames(and string) string {
	var w []string
	for _, l := range format.RoleLimits {
		w = append(w, l.Word)
	}
	if len(w) < 2 {
		return strings.Join(w, "")
	}
	return strings.Join(w[:len(w)-1], ", ") + and + w[len(w)-1]
}

// UnknownLimits gives each name bonsai.yaml sets under a role of any packs entry (packs[].roles.<role>) that is not one
// of a role's limits, in the file's order, each a *RolesError naming its line (from c.Doc): check's findings, beside
// the role findings it raises on KnownLimits. A role whose value is not a mapping is the schema's to refuse, not this.
func UnknownLimits(c *Config) []*RolesError {
	if c == nil || c.Full == nil {
		return nil
	}
	var out []*RolesError
	for i, p := range c.Full.Packs {
		for _, m := range p.Roles {
			limits, _ := m.Value.(schema.Object)
			for _, l := range limits {
				if roleLimit(l.Key) {
					continue
				}
				e := &RolesError{File: ConfigFile, Pack: p.ID, Index: i, Path: []string{m.Key, l.Key}, Unknown: true,
					Err: fmt.Errorf("the role %s sets %s, which is not one of a role's limits (%s)", m.Key, l.Key, limitNames(", "))}
				e.Line = ConfigLine(c.Doc, i, e.Path)
				out = append(out, e)
			}
		}
	}
	return out
}

// KnownLimits gives a copy of w with every name that is not one of a role's limits left out of each packs entry's
// roles (UnknownLimits names them), so check raises every other role finding beside them. w itself is not changed.
func KnownLimits(w *format.Workspace) *format.Workspace {
	if w == nil {
		return nil
	}
	cp := *w
	cp.Packs = make([]format.WorkspacePack, len(w.Packs))
	for i, p := range w.Packs {
		if p.Roles != nil {
			roles := schema.Object{}
			for _, m := range p.Roles {
				if limits, ok := m.Value.(schema.Object); ok {
					kept := schema.Object{}
					for _, l := range limits {
						if roleLimit(l.Key) {
							kept = append(kept, l)
						}
					}
					m = schema.Member{Key: m.Key, Value: kept}
				}
				roles = append(roles, m)
			}
			p.Roles = roles
		}
		cp.Packs[i] = p
	}
	return &cp
}

// ConfigLine is the line of a value under bonsai.yaml's packs[index].roles, found by its path (a role, a limit) in the
// file's map: the line of the deepest key of the path that is there, the roles key's own when none is, 0 when the map
// or the entry is not there.
func ConfigLine(doc *reader.Map, index int, path []string) int {
	packs, _ := doc.Get("packs")
	list, _ := packs.([]any)
	if index < 0 || index >= len(list) {
		return 0
	}
	entry, _ := list[index].(*reader.Map)
	e, ok := entry.Entry("roles")
	if !ok {
		return 0
	}
	line := e.Line
	for _, key := range path {
		m, _ := e.Value.(*reader.Map)
		if e, ok = m.Entry(key); !ok {
			break
		}
		line = e.Line
	}
	return line
}

// inForce gives one limit in force from the pack's and the project's.
func inForce(pack format.RoleLimit, own SetLimit) LimitInForce {
	l := LimitInForce{Default: pack.Default, DefaultFrom: LimitFromPack, Max: pack.Max, MaxFrom: LimitFromPack}
	if own.HasDefault {
		l.Default, l.DefaultFrom = own.Default, LimitFromConfig
	}
	if own.HasMax {
		l.Max, l.MaxFrom = own.Max, LimitFromConfig
	}
	if l.Max != nil && (l.Default == nil || *l.Default > *l.Max) {
		l.OverMax, l.SetDefault, l.SetDefaultFrom = true, l.Default, l.DefaultFrom
		l.Default, l.DefaultFrom = l.Max, l.MaxFrom
	}
	return l
}

// ConfigRoles reads a packs entry's roles in bonsai.yaml (format.WorkspacePack.Roles, as the full read keeps it): a
// role's name to its limits by name (format.RoleLimits), each with what it sets. nil (roles: null) sets none. Its
// error names a limit's name the format does not have (a closed list: never a field to skip), or a value that is not
// of the schema's shape (the full read holds the file to the schema first, so only a value built by hand gets one). A
// field of a limit other than default and max is kept in the file and never read (contract §2.2).
func ConfigRoles(roles schema.Object) (map[string]map[string]SetLimit, error) {
	out := map[string]map[string]SetLimit{}
	for _, m := range roles {
		limits, ok := m.Value.(schema.Object)
		if !ok {
			return nil, &rolePathError{path: []string{m.Key}, msg: fmt.Sprintf("the role %s is %s, not a mapping of its limits", m.Key, schema.Show(m.Value))}
		}
		byName := map[string]SetLimit{}
		for _, l := range limits {
			if !roleLimit(l.Key) {
				return nil, &rolePathError{path: []string{m.Key, l.Key}, unknown: true,
					msg: fmt.Sprintf("the role %s sets %s, which is not one of a role's limits (%s)", m.Key, l.Key, limitNames(", "))}
			}
			fields, ok := l.Value.(schema.Object)
			if !ok {
				return nil, &rolePathError{path: []string{m.Key, l.Key},
					msg: fmt.Sprintf("the role %s's %s is %s, not a mapping of default and max", m.Key, l.Key, schema.Show(l.Value))}
			}
			var s SetLimit
			for _, f := range fields {
				if f.Key != "default" && f.Key != "max" {
					continue // an unknown field: kept, never read
				}
				n, err := limitValue(f.Value)
				if err != nil {
					return nil, &rolePathError{path: []string{m.Key, l.Key, f.Key}, msg: fmt.Sprintf("the role %s's %s.%s %v", m.Key, l.Key, f.Key, err)}
				}
				if f.Key == "default" {
					s.HasDefault, s.Default = true, n
				} else {
					s.HasMax, s.Max = true, n
				}
			}
			byName[l.Key] = s
		}
		out[m.Key] = byName
	}
	return out, nil
}

// rolePathError is ConfigRoles' error: where under the entry's roles the value is, and whether its last key is a name
// that is not one of a role's limits.
type rolePathError struct {
	path    []string
	unknown bool
	msg     string
}

func (e *rolePathError) Error() string { return e.msg }

// roleLimit reports whether a name is one of format.RoleLimits.
func roleLimit(name string) bool {
	for _, w := range format.RoleLimits {
		if w.Word == name {
			return true
		}
	}
	return false
}

// limitValue reads a limit's default or max: a whole number from 0 up, or null for no limit.
func limitValue(v any) (*int64, error) {
	if v == nil {
		return nil, nil
	}
	var n int64
	ok := false
	switch x := v.(type) {
	case json.Number:
		var err error
		n, err = strconv.ParseInt(string(x), 10, 64)
		ok = err == nil
	case int64:
		n, ok = x, true
	case int:
		n, ok = int64(x), true
	}
	if !ok || n < 0 {
		return nil, fmt.Errorf("is %s, not a whole number from 0 up or null", schema.Show(v))
	}
	return &n, nil
}

// MessageThreshold gives one of bonsai messages' thresholds in force (format.MessageThresholds, by its setting's
// name: unanswered_minutes, circling_briefs): bonsai.yaml's messages where it sets it, else Bonsai's default. ok false
// for a name the table does not have. cfg nil, or messages null, gives the default.
func MessageThreshold(cfg *format.Workspace, setting string) (n int64, ok bool) {
	t, ok := format.MessageThresholdOf(setting)
	if !ok {
		return 0, false
	}
	if cfg == nil || cfg.Messages == nil {
		return t.Default, true
	}
	var set *int64
	switch setting {
	case "unanswered_minutes":
		set = cfg.Messages.UnansweredMinutes
	case "circling_briefs":
		set = cfg.Messages.CirclingBriefs
	}
	if set == nil {
		return t.Default, true
	}
	return *set, true
}
