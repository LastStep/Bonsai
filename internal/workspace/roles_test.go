package workspace

// Step 5.5.0's reads of roles: the roles in force with each limit's default and max from the pack, from bonsai.yaml
// and from both, the messages thresholds, and the settings record's compare, which reads neither roles nor messages.

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/schema"
)

func limit(n int64) *int64 { return &n }

// rolesLock is a lock whose pack declares three roles: the session's, a lead and a helper, as workflow ships them.
func rolesLock(t *testing.T) *Lock {
	t.Helper()
	l := func(d, m *int64) format.RoleLimit { return format.RoleLimit{Default: d, Max: m} }
	d := &format.Declares{Roles: []format.PackRole{
		{Name: "orchestrator", Kind: "lead", Session: true, Writes: true, Minutes: l(nil, nil), Actions: l(nil, nil), RedRuns: l(nil, nil)},
		{Name: "builder", Kind: "lead", Writes: true, MaxTurns: limit(2400), Minutes: l(limit(180), limit(180)),
			Actions: l(limit(600), limit(600)), RedRuns: l(limit(4), limit(4))},
		{Name: "auditor", Kind: "helper", MaxTurns: limit(600), Minutes: l(limit(30), limit(30)), Actions: l(limit(150), limit(150)),
			RedRuns: l(nil, nil)},
	}}
	o, err := d.Object()
	if err != nil {
		t.Fatal(err)
	}
	return &Lock{Packs: []LockedPack{{ID: "workflow", Declares: o, PathSet: true}, {ID: "plain", Declares: schema.Object{}, PathSet: true}}}
}

// configWith is a bonsai.yaml read in full whose workflow entry's roles are the YAML given (indented under roles:),
// or null for "".
func configWith(t *testing.T, roles string) *format.Workspace {
	t.Helper()
	c, _ := configFileWith(t, roles)
	return c.Full
}

// configFileWith is configWith's Config, every key with its line, and the file's text.
func configFileWith(t *testing.T, roles string) (*Config, string) {
	t.Helper()
	y := strings.Replace(checksYAML, "packs: []\n", "packs:\n  - id: workflow\n    source: \"https://example.com/workflow.git\"\n"+
		"    path: null\n    ref: v1\n    roles: null\n", 1)
	if roles != "" {
		y = strings.Replace(y, "    roles: null\n", "    roles:\n"+roles, 1)
	}
	c, err := ReadConfigFull([]byte(y))
	if err != nil {
		t.Fatalf("%v\n%s", err, y)
	}
	return c, y
}

// show writes a limit in force as "default(from)/max(from)", "-" for no limit, and "!" after it when the default set
// was above the max.
func show(l LimitInForce) string {
	n := func(p *int64) string {
		if p == nil {
			return "-"
		}
		return schema.Show(*p)
	}
	s := n(l.Default) + "(" + l.DefaultFrom + ")/" + n(l.Max) + "(" + l.MaxFrom + ")"
	if l.OverMax {
		s += "!" + n(l.SetDefault) + "(" + l.SetDefaultFrom + ")"
	}
	return s
}

// Each limit's default and max in force: from the pack (roles: null, or a role or limit bonsai.yaml leaves out), from
// bonsai.yaml (it sets both), and from both (it sets one); a default above its max has the max hold; null is no limit.
func TestRolesInForce(t *testing.T) {
	lock := rolesLock(t)
	get := func(t *testing.T, roles string) map[string]RoleInForce {
		t.Helper()
		got, err := RolesInForce(lock, configWith(t, roles))
		if err != nil {
			t.Fatal(err)
		}
		out := map[string]RoleInForce{}
		var order []string
		for _, r := range got {
			out[r.Role.Name] = r
			order = append(order, r.Pack+":"+r.Role.Name)
		}
		if strings.Join(order, ",") != "workflow:orchestrator,workflow:builder,workflow:auditor" {
			t.Errorf("roles, in the lock's order: %v", order)
		}
		return out
	}
	// From the pack.
	r := get(t, "")
	for _, c := range []struct{ role, limit, want string }{
		{"builder", "minutes", "180(pack)/180(pack)"}, {"builder", "actions", "600(pack)/600(pack)"}, {"builder", "red_runs", "4(pack)/4(pack)"},
		{"auditor", "red_runs", "-(pack)/-(pack)"}, {"orchestrator", "minutes", "-(pack)/-(pack)"},
	} {
		l, ok := r[c.role].Limit(c.limit)
		if !ok || show(l) != c.want {
			t.Errorf("from the pack, %s's %s: %s, want %s", c.role, c.limit, show(l), c.want)
		}
	}
	if r["builder"].Role.MaxTurns == nil || *r["builder"].Role.MaxTurns != 2400 || r["auditor"].Role.Kind != "helper" || !r["orchestrator"].Role.Session {
		t.Errorf("the roles as declared: %+v", r)
	}
	if _, ok := r["builder"].Limit("hours"); ok {
		t.Error("a limit that is not one of format.RoleLimits")
	}
	// From the project: both set (one raised, one lowered, one to no limit); and from both: one of the two set.
	r = get(t, "      builder:\n        minutes:\n          default: 240\n          max: 480\n        actions:\n          max: 900\n"+
		"        red_runs:\n          default: null\n          max: null\n      auditor:\n        actions:\n          default: 100\n")
	for _, c := range []struct{ role, limit, want string }{
		{"builder", "minutes", "240(bonsai.yaml)/480(bonsai.yaml)"},
		{"builder", "actions", "600(pack)/900(bonsai.yaml)"},
		{"builder", "red_runs", "-(bonsai.yaml)/-(bonsai.yaml)"},
		{"auditor", "actions", "100(bonsai.yaml)/150(pack)"},
		{"auditor", "minutes", "30(pack)/30(pack)"},
	} {
		l, _ := r[c.role].Limit(c.limit)
		if show(l) != c.want {
			t.Errorf("%s's %s: %s, want %s", c.role, c.limit, show(l), c.want)
		}
	}
	// A default above its max: the max holds, from where the max came; the default as set is kept for check to name.
	r = get(t, "      builder:\n        minutes:\n          default: 240\n        actions:\n          default: null\n"+
		"      orchestrator:\n        minutes:\n          max: 60\n")
	for _, c := range []struct{ role, limit, want string }{
		{"builder", "minutes", "180(pack)/180(pack)!240(bonsai.yaml)"},
		{"builder", "actions", "600(pack)/600(pack)!-(bonsai.yaml)"},
		{"orchestrator", "minutes", "60(bonsai.yaml)/60(bonsai.yaml)!-(pack)"},
	} {
		l, _ := r[c.role].Limit(c.limit)
		if show(l) != c.want {
			t.Errorf("over its max, %s's %s: %s, want %s", c.role, c.limit, show(l), c.want)
		}
	}
	// A role bonsai.yaml sets that the pack does not declare is left out (check names it); no lock, no roles.
	r = get(t, "      tester:\n        minutes:\n          default: 5\n")
	if _, ok := r["tester"]; ok {
		t.Error("an undeclared role is in force")
	}
	if got, err := RolesInForce(nil, nil); got != nil || err != nil {
		t.Errorf("no lock: %v, %v", got, err)
	}
	// A lock written before set 8 (no roles in its declares) declares none.
	old := &Lock{Packs: []LockedPack{{ID: "workflow", Declares: schema.Object{{Key: "protected", Value: []any{"x/**"}}}, PathSet: true}}}
	if got, err := RolesInForce(old, configWith(t, "")); len(got) != 0 || err != nil {
		t.Errorf("a lock from before set 8: %v, %v", got, err)
	}
	// A declares that does not read is an error naming the pack.
	bad := &Lock{Packs: []LockedPack{{ID: "workflow", Declares: schema.Object{{Key: "roles", Value: "builder"}}}}}
	if _, err := RolesInForce(bad, nil); err == nil || !strings.Contains(err.Error(), "the pack workflow") {
		t.Errorf("a declares that does not read: %v", err)
	}
}

// bonsai.yaml's packs[].roles read by hand (a value the full read would refuse is an error, never a guess).
func TestConfigRoles(t *testing.T) {
	got, err := ConfigRoles(nil)
	if err != nil || len(got) != 0 {
		t.Errorf("roles: null: %v, %v", got, err)
	}
	for _, c := range []struct {
		roles schema.Object
		want  string
	}{
		{schema.Object{{Key: "builder", Value: "x"}}, "not a mapping of its limits"},
		{schema.Object{{Key: "builder", Value: schema.Object{{Key: "hours", Value: schema.Object{}}}}}, "not one of a role's limits"},
		{schema.Object{{Key: "builder", Value: schema.Object{{Key: "minutes", Value: 3}}}}, "not a mapping of default and max"},
		{schema.Object{{Key: "builder", Value: schema.Object{{Key: "minutes", Value: schema.Object{{Key: "max", Value: "x"}}}}}}, "not a whole number"},
	} {
		if _, err := ConfigRoles(c.roles); err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s: %v, want %q", schema.Show(c.roles), err, c.want)
		}
	}
	// A field of a limit other than default and max: kept in the file, never read, whatever its value.
	got, err = ConfigRoles(schema.Object{{Key: "builder", Value: schema.Object{{Key: "minutes", Value: schema.Object{{Key: "maxx", Value: "x"},
		{Key: "max", Value: int64(5)}}}}}})
	if err != nil || !got["builder"]["minutes"].HasMax || got["builder"]["minutes"].HasDefault || *got["builder"]["minutes"].Max != 5 {
		t.Errorf("an unknown field in a limit: %v, %v", got, err)
	}
}

// An unknown limit name under a role (VM's D2: minutez): a role's limits are a closed list, so bonsai.yaml does not
// read for the roles. RolesInForceOf's error names bonsai.yaml, its line and the step that fixes it, never the lock;
// UnknownLimits names each such name with its line, in any packs entry; KnownLimits leaves them out so the rest is
// read (the copy alone: the file's own is kept). A field inside a limit other than default and max is kept and never
// read, whatever its value. The lock's declares that do not read are the lock's.
func TestUnknownLimitNames(t *testing.T) {
	roles := "      builder:\n        minutes:\n          default: 240\n          maxx: soon\n        minutez:\n          max: 5\n" +
		"      auditor:\n        hours:\n          max: 1\n"
	c, text := configFileWith(t, roles)
	line := func(prefix string) int {
		for i, l := range strings.Split(text, "\n") {
			if strings.HasPrefix(l, prefix) {
				return i + 1
			}
		}
		return 0
	}
	_, err := RolesInForceOf(rolesLock(t), c)
	var re *RolesError
	if !errors.As(err, &re) || re.File != ConfigFile || re.Line != line("        minutez:") || !re.Unknown {
		t.Fatalf("an unknown name: %#v", err)
	}
	at := "bonsai.yaml line " + strconv.Itoa(re.Line)
	if re.Error() != at+" sets minutez for the role builder under the pack workflow (packs[0].roles.builder.minutez), which is not "+
		"one of a role's limits (minutes, actions, red_runs)" || re.Next() != "in "+at+", write minutes, actions or red_runs in place "+
		"of minutez, or take it out" || strings.Contains(re.Error()+re.Next(), "lock") {
		t.Errorf("the error: %s; next: %s", re.Error(), re.Next())
	}
	unknown := UnknownLimits(c)
	if len(unknown) != 2 || unknown[0].Line != line("        minutez:") || unknown[1].Line != line("        hours:") ||
		!strings.Contains(unknown[1].Error(), "sets hours for the role auditor") {
		t.Errorf("UnknownLimits: %v", unknown)
	}
	known := KnownLimits(c.Full)
	in, err := RolesInForce(rolesLock(t), known)
	if err != nil || len(in) != 3 || show(in[1].Minutes) != "180(pack)/180(pack)!240(bonsai.yaml)" {
		t.Errorf("the known limits read: %v %v", in, err)
	}
	if len(UnknownLimits(c)) != 2 || len(UnknownLimits(&Config{Full: known})) != 0 {
		t.Error("KnownLimits changed the file's own roles, or kept an unknown name")
	}
	// A role whose value is no mapping, built by hand, names its place; the lock's declares are the lock's.
	_, err = RolesInForce(rolesLock(t), &format.Workspace{Packs: []format.WorkspacePack{{ID: "workflow", Roles: schema.Object{{Key: "builder", Value: "x"}}}}})
	if !errors.As(err, &re) || re.File != ConfigFile || re.Unknown || !strings.Contains(re.Error(), "bonsai.yaml: the pack workflow's roles (packs[0].roles.builder)") {
		t.Errorf("a role built by hand: %v", err)
	}
	bad := &Lock{Packs: []LockedPack{{ID: "workflow", Declares: schema.Object{{Key: "roles", Value: "builder"}}}}}
	if _, err := RolesInForceOf(bad, c); !errors.As(err, &re) || re.File != LockFile || !strings.Contains(re.Next(), "bonsai update") {
		t.Errorf("a declares that does not read: %v", err)
	}
}

// bonsai messages' thresholds in force: bonsai.yaml's messages where it sets one, else Bonsai's default.
func TestMessageThreshold(t *testing.T) {
	w := configWith(t, "")
	for _, th := range format.MessageThresholds {
		if n, ok := MessageThreshold(w, th.Setting); !ok || n != th.Default {
			t.Errorf("messages null: %s %d, want %d", th.Setting, n, th.Default)
		}
		if n, ok := MessageThreshold(nil, th.Setting); !ok || n != th.Default {
			t.Errorf("no bonsai.yaml: %s %d", th.Setting, n)
		}
	}
	w.Messages = &format.Thresholds{UnansweredMinutes: limit(120)}
	if n, _ := MessageThreshold(w, "unanswered_minutes"); n != 120 {
		t.Errorf("unanswered_minutes set: %d", n)
	}
	if n, _ := MessageThreshold(w, "circling_briefs"); n != 3 {
		t.Errorf("circling_briefs left null: %d", n)
	}
	if _, ok := MessageThreshold(w, "other"); ok {
		t.Error("a setting the table does not have")
	}
}

// The settings record's compare (checks.go) reads neither a packs entry's roles nor messages: a raised default or max,
// a lowered one, and new thresholds are never config-looser (the director, 14 Oct: "raising a budget does not count as
// loosening the checks"), and the record's checks stay the same.
func TestRolesAndMessagesAreNoLoosening(t *testing.T) {
	was := ChecksOf(configWith(t, ""))
	for _, roles := range []string{
		"      builder:\n        minutes:\n          default: 240\n          max: 480\n",
		"      builder:\n        actions:\n          max: 100000\n        red_runs:\n          default: null\n          max: null\n",
		"      builder:\n        minutes:\n          default: 1\n          max: 1\n",
	} {
		now := configWith(t, roles)
		now.Messages = &format.Thresholds{UnansweredMinutes: limit(1), CirclingBriefs: limit(100)}
		c := ChecksOf(now)
		if l := Loosenings(was, c); len(l) != 0 {
			t.Errorf("roles %q and messages read as looser: %q", roles, l)
		}
		if l := Loosenings(c, was); len(l) != 0 {
			t.Errorf("taking roles %q and messages out reads as looser: %q", roles, l)
		}
		if !sameChecks(was, c) {
			t.Errorf("roles %q and messages changed the record's checks:\n%s\n%s", roles, schema.Show(was.JSON()), schema.Show(c.JSON()))
		}
	}
}
