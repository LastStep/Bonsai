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
// Raising a default or a max is never a loosening of the checks (Rohan, 14 Oct): the settings record's compare
// (checks.go) reads neither roles nor messages.
//
// Two packs declaring one role's name each keep their own (a role is <pack>:<name>, as Claude Code names it), so a
// reader looks a role up by its pack and name. A lock written before set 8 declares no role.

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/LastStep/Bonsai/internal/format"
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
// read in full; nil reads as setting none). Its error names a pack whose declares or whose roles in bonsai.yaml do
// not read; a role bonsai.yaml sets that its pack does not declare is left out here (check names it,
// role-undeclared). A nil lock gives none.
func RolesInForce(lock *Lock, cfg *format.Workspace) ([]RoleInForce, error) {
	if lock == nil {
		return nil, nil
	}
	set := map[string]schema.Object{} // a pack's id to its roles in bonsai.yaml
	if cfg != nil {
		for _, p := range cfg.Packs {
			set[p.ID] = p.Roles
		}
	}
	var out []RoleInForce
	for _, lp := range lock.Packs {
		d, err := lp.Declared()
		if err != nil {
			return nil, fmt.Errorf("the lock's declares for the pack %s cannot be read: %v", lp.ID, err)
		}
		own, err := ConfigRoles(set[lp.ID])
		if err != nil {
			return nil, fmt.Errorf("bonsai.yaml's roles for the pack %s: %v", lp.ID, err)
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
// error names a value that is not of the schema's shape (the full read holds the file to the schema first, so only a
// value built by hand gets one), or a limit's name the format does not have.
func ConfigRoles(roles schema.Object) (map[string]map[string]SetLimit, error) {
	out := map[string]map[string]SetLimit{}
	for _, m := range roles {
		limits, ok := m.Value.(schema.Object)
		if !ok {
			return nil, fmt.Errorf("the role %s is %s, not a mapping of its limits", m.Key, schema.Show(m.Value))
		}
		byName := map[string]SetLimit{}
		for _, l := range limits {
			if !roleLimit(l.Key) {
				return nil, fmt.Errorf("the role %s sets %s, which is not one of a role's limits (minutes, actions, red_runs)", m.Key, l.Key)
			}
			fields, ok := l.Value.(schema.Object)
			if !ok {
				return nil, fmt.Errorf("the role %s's %s is %s, not a mapping of default and max", m.Key, l.Key, schema.Show(l.Value))
			}
			var s SetLimit
			for _, f := range fields {
				n, err := limitValue(f.Value)
				if err != nil {
					return nil, fmt.Errorf("the role %s's %s.%s %v", m.Key, l.Key, f.Key, err)
				}
				switch f.Key {
				case "default":
					s.HasDefault, s.Default = true, n
				case "max":
					s.HasMax, s.Max = true, n
				}
			}
			byName[l.Key] = s
		}
		out[m.Key] = byName
	}
	return out, nil
}

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
