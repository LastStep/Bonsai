package engine

// check's findings on the limits a project sets for its packs' roles (set 8; plan-5 5.5.0 note 8), each a word of
// format.RoleCheckWords, their one home, and each a finding whose next step is an agent's (one whose task grants
// bonsai.yaml). They read the lock's declares (the roles each pack declares, offline) and bonsai.yaml as it stands, through
// workspace.RolesInForce, the one function every reader of a role's limits uses:
//
//   - role-undeclared: bonsai.yaml sets limits under a pack's entry (packs[].roles) for a role that pack does not
//     declare; the finding names the pack's roles;
//   - role-default-over-max: a role's limit in force has its default above its max (no limit is above any number),
//     whichever set each; the max then holds, and the finding names both and where each came from;
//   - role-actions-over-turns: a role's max of actions in force is above the max_turns its pack declares (no limit is
//     above any number), so Claude Code would stop the agent at its maxTurns before the guard's budget does.
//
// A pack bonsai.yaml lists that the lock does not hold is the packs finding's (check.go), and a lock whose declares do
// not read is the lock's: neither gives a role finding.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/LastStep/Bonsai/internal/workspace"
)

// checkRoles adds the role findings.
func (r *CheckResult) checkRoles() {
	if r.Config == nil || r.Config.Full == nil || r.Lock == nil {
		return
	}
	locked := map[string]workspace.LockedPack{}
	for _, lp := range r.Lock.Packs {
		locked[lp.ID] = lp
	}
	entry := map[string]int{} // a pack's id to its index in bonsai.yaml's packs
	for i, p := range r.Config.Full.Packs {
		entry[p.ID] = i
		lp, ok := locked[p.ID]
		if !ok || len(p.Roles) == 0 {
			continue
		}
		d, err := lp.Declared()
		if err != nil {
			continue
		}
		declared := map[string]bool{}
		var names []string
		for _, role := range d.Roles {
			declared[role.Name] = true
			names = append(names, role.Name)
		}
		roles := "it declares none"
		if len(names) > 0 {
			roles = "its roles: " + strings.Join(names, ", ")
		}
		for _, m := range p.Roles {
			if declared[m.Key] {
				continue
			}
			at := fmt.Sprintf("packs[%d].roles.%s", i, m.Key)
			next := fmt.Sprintf("take %s out of bonsai.yaml", at)
			if len(names) > 0 {
				next = fmt.Sprintf("in bonsai.yaml, name one of the pack %s's roles (%s) in place of %s, or take %s out", p.ID,
					strings.Join(names, ", "), m.Key, at)
			}
			r.add("role-undeclared", workspace.ConfigFile, "", fmt.Sprintf("bonsai.yaml sets limits for the role %s under the pack %s "+
				"(%s), which that pack does not declare (%s)", quote(m.Key), p.ID, at, roles), next)
		}
	}
	inForce, err := workspace.RolesInForce(r.Lock, r.Config.Full)
	if err != nil {
		return // the lock's or bonsai.yaml's own finding names what does not read
	}
	for _, ri := range inForce {
		name := ri.Pack + ":" + ri.Role.Name
		at := "packs[].roles." + ri.Role.Name
		if i, ok := entry[ri.Pack]; ok {
			at = fmt.Sprintf("packs[%d].roles.%s", i, ri.Role.Name)
		}
		for _, limit := range []string{"minutes", "actions", "red_runs"} {
			l, _ := ri.Limit(limit)
			if !l.OverMax {
				continue
			}
			r.add("role-default-over-max", workspace.ConfigFile, "", fmt.Sprintf("the role %s's %s default in force, %s (%s), is above "+
				"its max, %s (%s): the max holds", name, limit, limitWords(l.SetDefault), fromWords(l.SetDefaultFrom),
				limitWords(l.Max), fromWords(l.MaxFrom)), fmt.Sprintf("in bonsai.yaml, under %s.%s, set a default at or below the max, or "+
				"raise the max (raising a limit is never a loosening of the checks)", at, limit))
		}
		if ri.Role.MaxTurns == nil {
			continue
		}
		if max := ri.Actions.Max; max == nil || *max > *ri.Role.MaxTurns {
			r.add("role-actions-over-turns", workspace.ConfigFile, "", fmt.Sprintf("the role %s's max of actions in force, %s (%s), is "+
				"above its max_turns, %d (its pack's agents/%s.md), so Claude Code would stop the agent before the guard's budget does",
				name, limitWords(max), fromWords(ri.Actions.MaxFrom), *ri.Role.MaxTurns, ri.Role.Name),
				fmt.Sprintf("in bonsai.yaml, set %s.actions.max to %d or less; or move to a pack whose role allows more", at, *ri.Role.MaxTurns))
		}
	}
}

// limitWords is a limit's value for a person: its number, or "no limit".
func limitWords(n *int64) string {
	if n == nil {
		return "no limit"
	}
	return strconv.FormatInt(*n, 10)
}

// fromWords says where a limit's value came from.
func fromWords(from string) string {
	if from == workspace.LimitFromConfig {
		return "bonsai.yaml's"
	}
	return "the pack's"
}
