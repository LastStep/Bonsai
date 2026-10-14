package guard

// The roles' rules (design/plan-5.md, step 5.5, "Role rules and stall stops, written once"; 5.5.11 notes 3 to 6), in a
// project whose lock declares roles; elsewhere the guard is as before. Claude Code only: they read its payload's
// agent_id and agent_type, as measured on Claude Code 2.1.300 (WSL, 15 Oct: the run report's payload table).
//
// When they apply (rolesMayApply): the main checkout's .bonsai/local/agents/ is there (the guard has kept agents'
// state in it), or its lock as it stands names roles; then every role fact is read as committed (committedRoles), and
// where the committed lock declares no role the guard is as before. A plain session (no agent_id, no agent_type, no
// BONSAI_ROLE) is the person's own, and reads nothing new until it starts or messages an agent, or runs bonsai
// message (its address is its role's name).
//
// The caller's role (caller), in this order: its agent_type where that names a declared role (a subagent, or a
// top-level --agent session; a subagent naming the session's role is read as an undeclared type); else, inside a
// subagent (agent_id set), an undeclared type, held as a helper; else, for a top-level session, BONSAI_ROLE where it
// names a declared role other than the session's (a named session, counted by its session id); else the person's own
// session, which plays the session's role. A type names a role by its scoped name (<pack>:<role>, as Claude Code names
// a plugin's agent), or by its bare name where one pack alone declares it. An agent's role, its facts and its limits
// are read once, at its first call, and kept in its state (internal/agents), so a later edit of the lock or of
// bonsai.yaml, committed or not, changes nothing for it; a new brief re-reads them.
//
// Who may call whom (mayCall), on Agent and Task (the type asked for), SendMessage (its to, resolved) and Workflow:
// the person's session any declared role but its own, any undeclared type, Workflow, and by SendMessage main or
// another session; a lead (a declared role of kind lead) a declared helper or an undeclared type, and by SendMessage
// one of them it may resume; a helper or an undeclared type no one. Each allowed Agent, Task or Workflow call is kept
// for the session (agents.Spawn), so the agent it starts is joined to it at its first call (link): by Claude Code's
// own meta file beside the agent's transcript, which names the starting call's tool_use_id (hookin.SpawnCall: present
// at 36 of 36 first calls measured, three calls side by side among them), else the oldest call of its type not yet
// joined (an agent a Workflow script starts: the newest Workflow call, since the guard sees no Workflow call end); the
// join writes a spawn record (bonsai.log/1) naming its parent, so the tree survives a cleaned state file.
//
// Known limits, said once (step 5.5's "What it cannot do"): an agent gone quiet makes no call, so nothing stops it; a
// session started where the shell judge cannot read the command reads as the person's own; Bonsai's records are
// attributed, not authenticated; a role that changes nothing can still write through what the guard cannot read (an
// interpreter, a script, an MCP tool's server); a named session can spend its budget on purpose to be let stop;
// agent teams' teammates are not covered; with bonsai missing or broken, the second line lets reads through with no
// budget; Claude Code's own internal agents (prompt suggestions, /btw) fire no tool call and are never seen; one that
// did would run as an undeclared type. A SendMessage resuming an agent that never made a tool call (so no state names
// it) does not resolve, and is refused from every caller but the person's session.

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/format"
	"github.com/LastStep/Bonsai/internal/hookin"
	"github.com/LastStep/Bonsai/internal/vergit"
	"github.com/LastStep/Bonsai/internal/workspace"
)

// RoleEnv names the variable a named session's role is read from (step 5.4's named sessions; never in a payload, as
// measured: only the hook's environment holds it).
const RoleEnv = "BONSAI_ROLE"

// HandBackTools are Claude Code's tools that hand a subagent's result back to its caller, which an agent past its
// budget may still call (budget-exempt): their one home.
var HandBackTools = []string{"SubagentHandback"}

// generalType is the type Claude Code gives an agent started with no subagent_type (measured: a fork reads so too).
const generalType = "general-purpose"

// workflowType is the type Claude Code gives every agent a Workflow script starts (measured).
const workflowType = "workflow-subagent"

// gitWait limits each of the committed reads' git calls.
const gitWait = 3 * time.Second

// stateWait is how long a call waits for an agent's header that another call side by side is writing.
const stateWait = 300 * time.Millisecond

// briefPattern finds a message's id in a prompt or a message's text (bonsai.message/1's id).
var briefPattern = sync.OnceValue(func() *regexp.Regexp {
	return regexp.MustCompile(`M-[0-9]{4}-[0-9]{2}-[0-9]{2}-[0-9a-f]{8}`)
})

// briefIDs gives the message ids a text names, in order, each once, at most eight.
func briefIDs(text string) []string {
	var out []string
	for _, id := range briefPattern().FindAllString(text, -1) {
		if !listHas(out, id) {
			out = append(out, id)
		}
		if len(out) == 8 {
			break
		}
	}
	return out
}

// roleSet is the roles in force, as committed at the main checkout's HEAD (or the files, where HEAD holds neither).
type roleSet struct {
	roles []workspace.RoleInForce
	files string // why they were read from the files; "" when read as committed
}

// rolesMayApply is the cheap test of whether a project's calls may meet the roles' rules (this file's comment):
// .bonsai/local/agents/ in the main checkout, or a lock that names roles as it stands. Neither costs a git call.
func rolesMayApply(main string) bool {
	if (agents.Store{Main: main}).Exists() {
		return true
	}
	raw, err := os.ReadFile(filepath.Join(main, filepath.FromSlash(workspace.LockFile)))
	return err == nil && strings.Contains(string(raw), `"roles"`)
}

// committedRoles reads the roles in force from the lock and bonsai.yaml as committed at the main checkout's HEAD:
// one git call each (vergit.Committed, its fixed environment; the lock first, and bonsai.yaml only where it declares a
// role); where HEAD holds neither (no commit yet, or neither committed), the files as they stand, saying so; where
// HEAD holds bonsai.yaml but no lock, no role. Any other error (git cannot run, HEAD cannot be read, a file that does
// not read) is the caller's to answer.
var committedRoles = func(main string) (*roleSet, error) {
	lockRaw, lerr := vergit.Committed(main, workspace.LockFile, gitWait)
	absent := func(err error) bool { return errors.Is(err, vergit.ErrNoCommit) || errors.Is(err, vergit.ErrNotInHead) }
	rs := &roleSet{}
	var cfgRaw []byte
	switch {
	case lerr != nil && !absent(lerr):
		return nil, fmt.Errorf("the lock cannot be read as committed at the main checkout's HEAD: %v", lerr)
	case absent(lerr):
		var cerr error
		cfgRaw, cerr = vergit.Committed(main, workspace.ConfigFile, gitWait)
		switch {
		case cerr == nil:
			return rs, nil // HEAD holds bonsai.yaml but no lock: no role is committed
		case !absent(cerr):
			return nil, fmt.Errorf("bonsai.yaml cannot be read as committed at the main checkout's HEAD: %v", cerr)
		}
		rs.files = "the main checkout's HEAD holds neither bonsai.yaml nor the lock"
		var err error
		if cfgRaw, err = os.ReadFile(filepath.Join(main, workspace.ConfigFile)); err != nil {
			return nil, fmt.Errorf("bonsai.yaml cannot be read: %v", err)
		}
		if lockRaw, err = os.ReadFile(filepath.Join(main, filepath.FromSlash(workspace.LockFile))); err != nil {
			return rs, nil // no lock: no role
		}
	}
	lock, err := workspace.ReadLock(lockRaw)
	if err != nil {
		return nil, fmt.Errorf("the lock as committed does not read: %v", err)
	}
	if !lockRoles(lock) {
		return rs, nil // no role declared: bonsai.yaml's limits are not needed
	}
	if cfgRaw == nil {
		var cerr error
		cfgRaw, cerr = vergit.Committed(main, workspace.ConfigFile, gitWait)
		if cerr != nil && !absent(cerr) {
			return nil, fmt.Errorf("bonsai.yaml cannot be read as committed at the main checkout's HEAD: %v", cerr)
		}
	}
	var full *format.Workspace
	if cfgRaw != nil {
		cfg, err := workspace.ReadConfigFull(cfgRaw)
		if err != nil {
			return nil, fmt.Errorf("bonsai.yaml as committed does not read: %v", err)
		}
		full = cfg.Full
	}
	if rs.roles, err = workspace.RolesInForce(lock, full); err != nil {
		return nil, err
	}
	return rs, nil
}

// lockRoles reports a lock whose packs declare a role (their declares' roles key).
func lockRoles(l *workspace.Lock) bool {
	for _, p := range l.Packs {
		if v, ok := p.Declares.Get("roles"); ok {
			if list, ok := v.([]any); ok && len(list) > 0 {
				return true
			}
		}
	}
	return false
}

// byType gives the declared role an agent type names: <pack>:<role>, or a bare <role> one pack alone declares; nil
// for none.
func (rs *roleSet) byType(t string) *workspace.RoleInForce {
	if t == "" {
		return nil
	}
	if pack, name, scoped := strings.Cut(t, ":"); scoped {
		for i := range rs.roles {
			if rs.roles[i].Pack == pack && rs.roles[i].Role.Name == name {
				return &rs.roles[i]
			}
		}
		return nil
	}
	var found *workspace.RoleInForce
	for i := range rs.roles {
		if rs.roles[i].Role.Name == t {
			if found != nil {
				return nil // two packs declare it: a bare name names neither
			}
			found = &rs.roles[i]
		}
	}
	return found
}

// sessionRole gives the role the person's own session plays (its session: true), nil when no pack declares one.
func (rs *roleSet) sessionRole() *workspace.RoleInForce {
	for i := range rs.roles {
		if rs.roles[i].Role.Session {
			return &rs.roles[i]
		}
	}
	return nil
}

// limitsOf gives a role's limits in force as an agent's: each default as its value, with its max.
func limitsOf(r *workspace.RoleInForce) agents.Limits {
	return agents.Limits{Minutes: agents.Limit{Value: r.Minutes.Default, Max: r.Minutes.Max},
		Actions: agents.Limit{Value: r.Actions.Default, Max: r.Actions.Max},
		RedRuns: agents.Limit{Value: r.RedRuns.Default, Max: r.RedRuns.Max}}
}

// undeclaredLimits are an undeclared type's limits: those in force of the declared helper with the smallest default
// actions (no limit counts as the largest; a tie goes by the pack's id, then the role's name), else Bonsai's own fixed
// helper limits (format.HelperLimits).
func (rs *roleSet) undeclaredLimits() agents.Limits {
	var helpers []*workspace.RoleInForce
	for i := range rs.roles {
		if rs.roles[i].Role.Kind == "helper" && !rs.roles[i].Role.Session {
			helpers = append(helpers, &rs.roles[i])
		}
	}
	if len(helpers) == 0 {
		var l agents.Limits
		for _, h := range format.HelperLimits {
			v := agents.Limit{Value: h.Default, Max: h.Max}
			switch h.Limit {
			case "minutes":
				l.Minutes = v
			case "actions":
				l.Actions = v
			case "red_runs":
				l.RedRuns = v
			}
		}
		return l
	}
	sort.SliceStable(helpers, func(i, j int) bool {
		a, b := helpers[i], helpers[j]
		switch {
		case a.Actions.Default == nil && b.Actions.Default != nil:
			return false
		case a.Actions.Default != nil && b.Actions.Default == nil:
			return true
		case a.Actions.Default != nil && *a.Actions.Default != *b.Actions.Default:
			return *a.Actions.Default < *b.Actions.Default
		case a.Pack != b.Pack:
			return a.Pack < b.Pack
		}
		return a.Role.Name < b.Role.Name
	})
	return limitsOf(helpers[0])
}

// callee is what a call asks for: a declared role (role set), an undeclared type, the session's role, or (a
// SendMessage) main or another session, or nothing the guard recorded.
type callee struct {
	role       *workspace.RoleInForce // a declared role; nil for an undeclared type
	session    bool                   // the session's role, asked for as an agent's type
	main       bool                   // a SendMessage's to naming main or another session
	unresolved bool                   // a SendMessage's to that names no agent the guard recorded
	shown      string                 // as a refusal names it
}

// headerOf gives the header an agent of a role (nil: an undeclared type) starts with.
func headerOf(rs *roleSet, r *workspace.RoleInForce) agents.Header {
	if r == nil {
		return agents.Header{Kind: "helper", Limits: rs.undeclaredLimits()}
	}
	return agents.Header{Role: r.Role.Name, Pack: r.Pack, Kind: r.Role.Kind, SessRole: r.Role.Session, Writes: r.Role.Writes,
		MaxTurns: r.Role.MaxTurns, Limits: limitsOf(r)}
}

// who is the caller of one call, where the roles apply.
type who struct {
	key     string        // its state's name: its agent id, or a top-level session's own id; "" for a plain session
	h       agents.Header // its role and facts (a plain session's: SessRole)
	st      *agents.State // its state; nil for the person's session
	counted bool          // an agent or a named session: counted and held to its budget
}

// lead reports a caller of kind lead (a declared lead, a named session of one).
func (w *who) lead() bool { return w.h.Role != "" && w.h.Kind == "lead" && !w.h.SessRole }

// shown names the caller in a refusal.
func (w *who) shown() string {
	switch {
	case w.h.SessRole:
		return "the person's session"
	case w.h.Role == "":
		t := w.h.Type
		if t == "" {
			t = "an unnamed type"
		}
		return "an agent of the type " + quote(t, 60) + ", which no pack declares (held as a helper)"
	case w.h.Named:
		return "a named session playing the " + w.h.Kind + " role " + quote(w.h.Role, 40)
	}
	return "an agent of the " + w.h.Kind + " role " + quote(w.h.Role, 40)
}

// roleRun is the roles' part of one guard run: what it read, once.
type roleRun struct {
	j       *judge
	in      *hookin.Payload
	main    string
	store   agents.Store
	rs      *roleSet // nil until read
	rsErr   error
	caller  *who
	now     time.Time
	climbs  []string // the tasks of the bonsai ladder --task calls this call's shell line runs, as passed
	records []extraRecord
}

// extraRecord is a guard record written beside the call's own (a brief that ties nothing, a state that cannot be
// read, roles read from the files) or a spawn record.
type extraRecord struct {
	d      Decision
	spawn  *spawnRecord
	in     *hookin.Payload
	atTime time.Time
}

// roles reads, once, the roles in force.
func (rr *roleRun) roles() (*roleSet, error) {
	if rr.rs == nil && rr.rsErr == nil {
		rr.rs, rr.rsErr = committedRoles(rr.main)
		if rr.rsErr == nil && rr.rs.files != "" {
			rr.note(RuleRolesFromFiles, rr.rs.files+", so the role facts and limits were read from the files as they stand")
		}
	}
	return rr.rs, rr.rsErr
}

// note keeps a guard record to write beside the call's own.
func (rr *roleRun) note(rule, why string) {
	rr.records = append(rr.records, extraRecord{d: Decision{Allow: true, Rule: rule, Why: why}, in: rr.in, atTime: rr.now})
}

// rolesFault is the answer when the roles cannot be read: the first line refuses (a guard that cannot read its rules
// blocks), the second lets the call through (second-line-fault, Main).
func rolesFault(err error) Decision {
	return deny(RuleBadLock, "the roles in force cannot be read ("+oneLine(err.Error())+"), so the guard cannot tell what "+
		"this agent may do, and it blocks", "tell the person: a person checks the lock and bonsai.yaml as committed (bonsai check), "+
		"and that git runs in the main checkout")
}

// newRoleRun starts the roles' part of a run for a call in the session's project, or gives nil where the roles do not
// apply (rolesMayApply false, or no role is committed and the caller is no agent known to the state).
func (j *judge) newRoleRun(in *hookin.Payload) *roleRun {
	if j.session == nil {
		return nil
	}
	main := j.local(j.session).Main
	if !rolesMayApply(main) {
		return nil
	}
	return &roleRun{j: j, in: in, main: main, store: agents.Store{Main: main}, now: j.clock()}
}

// resolve finds the caller (this file's comment), making an agent's state at its first call. It gives nil with no
// error where no role is committed, and the caller is not one the state already holds.
func (rr *roleRun) resolve() (*who, error) {
	in := rr.in
	envRole := rr.j.getenv(RoleEnv)
	if in.AgentID == "" && in.AgentType == "" && envRole == "" {
		return &who{h: agents.Header{SessRole: true}}, nil // the person's own session
	}
	key := in.AgentID
	if key == "" {
		key = in.Session
	}
	keep := in.Session != "" && hookin.ID(key) != "" // a state can be kept for it (Claude Code names every session)
	if keep {
		st, err := rr.store.ReadWait(in.Session, key, stateWait)
		switch {
		case err == nil:
			return &who{key: key, h: st.Header, st: st, counted: !st.Header.SessRole}, nil
		case !errors.Is(err, agents.ErrNoState):
			rr.note(RuleStateUnreadable, "the state of this agent cannot be read ("+oneLine(err.Error())+"), so its count starts afresh")
		}
	}
	rs, err := rr.roles()
	if err != nil {
		return nil, err
	}
	if len(rs.roles) == 0 {
		return nil, nil
	}
	h := agents.Header{SessRole: true}
	switch r := rs.byType(in.AgentType); {
	case in.AgentID != "" && (r == nil || r.Role.Session):
		h = headerOf(rs, nil) // a subagent of a type no pack declares, or naming the session's role
	case r != nil:
		h = headerOf(rs, r)
		h.Named = in.AgentID == "" && !r.Role.Session
	default:
		if r := rs.byType(envRole); r != nil && !r.Role.Session {
			h = headerOf(rs, r)
			h.Named = true
		}
	}
	h.Agent, h.Session, h.Type, h.First, h.ReadFrom, h.Note = key, in.Session, in.AgentType, agents.Stamp(rr.now), "HEAD", ""
	if rs.files != "" {
		h.ReadFrom, h.Note = "files", rs.files
	}
	if h.Named {
		h.Task = rr.j.getenv(workspace.TaskEnv)
	}
	if !keep {
		rr.note(RuleStateUnreadable, "the call names no session or agent id a state can be kept under, so it is not counted")
		return &who{key: key, h: h}, nil
	}
	made, err := rr.store.Create(h)
	if err != nil {
		rr.note(RuleStateUnreadable, "the state of this agent cannot be written ("+oneLine(err.Error())+"), so it is not counted")
		return &who{key: key, h: h, counted: false}, nil
	}
	st, err := rr.store.ReadWait(in.Session, key, stateWait)
	if err != nil {
		rr.note(RuleStateUnreadable, "the state of this agent cannot be read back ("+oneLine(err.Error())+"), so it is not counted")
		return &who{key: key, h: h, counted: false}, nil
	}
	w := &who{key: key, h: st.Header, st: st, counted: !st.Header.SessRole}
	if made && in.AgentID != "" {
		rr.link(w)
	}
	return w, nil
}

// link joins a subagent, at its first call, to the call that started it (this file's comment), notes the link in its
// state, writes the spawn record, and ties the first sound brief that call's prompt names.
func (rr *roleRun) link(w *who) {
	in := rr.in
	spawns, err := rr.store.Spawns(in.Session)
	if err != nil {
		rr.note(RuleStateUnreadable, "the session's spawn calls cannot be read ("+oneLine(err.Error())+"), so this agent is "+
			"joined to none")
		return
	}
	var picked *agents.Spawn
	if w.h.Type == workflowType {
		for i := len(spawns) - 1; i >= 0 && picked == nil; i-- {
			if spawns[i].Tool == "Workflow" {
				picked = &spawns[i]
			}
		}
	} else {
		if use, _, err := hookin.SpawnCall(in.TranscriptPath, in.Session, in.AgentID); err == nil {
			for i := range spawns {
				if spawns[i].Use == use && spawns[i].Tool != "Workflow" && childOf(&spawns[i], w.h.Type) {
					if ok, _ := rr.store.Claim(in.Session, use, w.key); ok {
						picked = &spawns[i]
					}
					break
				}
			}
		}
		for i := range spawns {
			if picked != nil {
				break
			}
			sp := &spawns[i]
			if sp.Tool == "Workflow" || !childOf(sp, w.h.Type) || rr.store.Claimed(in.Session, sp.Use) {
				continue
			}
			if ok, _ := rr.store.Claim(in.Session, sp.Use, w.key); ok {
				picked = sp
			}
		}
	}
	if picked == nil {
		return
	}
	e := agents.Event{E: agents.EventLink, At: agents.Stamp(rr.now), Parent: picked.Parent, Call: picked.Use, Briefs: picked.Briefs}
	if err := rr.store.Append(in.Session, w.key, e); err != nil {
		rr.note(RuleStateUnreadable, "this agent's link to its starting call cannot be written ("+oneLine(err.Error())+")")
		return
	}
	rr.records = append(rr.records, extraRecord{spawn: &spawnRecord{child: in.AgentID, childType: in.AgentType,
		use: picked.Use, tool: picked.Tool, parent: picked.Parent}, in: in, atTime: rr.now})
	if w.h.Type == workflowType {
		return // a Workflow script's agents have no brief of their own: no role, so no address
	}
	for _, id := range picked.Briefs {
		if rr.tie(w, id, "the call that started it") {
			break
		}
	}
	if st, err := rr.store.Read(in.Session, w.key); err == nil {
		w.st = st
	}
}

// childOf reports whether an agent of type t is what a spawn call asked for: the same type; Claude Code's
// general-purpose agent for a call naming none; or, for a bare name one pack declares, that pack's scoped type.
func childOf(sp *agents.Spawn, t string) bool {
	switch {
	case sp.Type == t:
		return true
	case sp.Type == "" && t == generalType:
		return true
	case sp.Role != "" && !strings.Contains(sp.Type, ":"):
		_, name, scoped := strings.Cut(t, ":")
		return scoped && name == sp.Type
	}
	return false
}

// resolveCallee reads what an Agent or Task call asks for.
func (rr *roleRun) resolveCallee(rs *roleSet, asked string) callee {
	t := asked
	if t == "" {
		t = generalType
	}
	r := rs.byType(t)
	switch {
	case r != nil && r.Role.Session:
		return callee{session: true, shown: "the session's role " + quote(r.Pack+":"+r.Role.Name, 60)}
	case r != nil:
		return callee{role: r, shown: "the " + r.Role.Kind + " role " + quote(r.Pack+":"+r.Role.Name, 60)}
	}
	return callee{shown: "the type " + quote(t, 60) + ", which no pack declares (held as a helper)"}
}

// mayCall says why the caller may not call the callee, "" when it may (this file's comment).
func mayCall(w *who, c callee) string {
	switch {
	case c.session:
		return "the person's session is never a subagent"
	case (c.main || c.unresolved) && w.h.SessRole:
		return "" // the person's session to main, another session, or a name it gave (Rohan's "A. Only your sessions")
	case c.main || c.unresolved:
		return "only the person's session may write to main or another session, or to an agent the guard never saw started"
	case w.h.SessRole:
		return ""
	case !w.lead():
		return "a helper, and every type no pack declares, calls no one"
	case c.role != nil && c.role.Role.Kind == "lead":
		return "a lead may call only helpers and types no pack declares, never another lead"
	}
	return ""
}

// rolesCalled names who a caller may call, for a refusal.
func rolesCalled(w *who) string {
	switch {
	case w.h.SessRole:
		return "any declared lead or helper, a type no pack declares, Workflow, main and another session"
	case w.lead():
		return "declared helpers and types no pack declares, by Agent or Task, and by SendMessage one of them it started"
	}
	return "no one"
}

// spawnCall judges an Agent, Task, SendMessage or Workflow call where the roles apply.
func (rr *roleRun) spawnCall(w *who) Decision {
	in := rr.in
	rs, err := rr.roles()
	if err != nil {
		return rolesFault(err)
	}
	refuse := func(what, why string) Decision {
		d := deny(RuleCallRefused, fmt.Sprintf("%s refused: %s may not call %s: %s (who may call whom: it may call %s)",
			in.Tool, w.shown(), what, why, rolesCalled(w)),
			"hand the work back to your caller in your report (bonsai message --kind report), naming what is left")
		return d
	}
	switch in.Tool {
	case "Workflow":
		if !w.h.SessRole {
			return refuse("Workflow", "only the person's session runs a Workflow")
		}
		rr.addSpawn(w, agents.Spawn{Tool: "Workflow", Type: workflowType, Briefs: briefIDs(in.Script)})
		return allow(RuleCallAllowed, "Workflow")
	case "SendMessage":
		c, target := rr.resolveTo(rs, in.To)
		if why := mayCall(w, c); why != "" {
			return refuse(c.shown, why)
		}
		if target != "" {
			from := w.h.Role
			if w.h.SessRole {
				if r := rs.sessionRole(); r != nil {
					from = r.Role.Name
				}
			}
			for _, id := range briefIDs(in.MessageText) {
				_ = rr.store.Append(in.Session, target, agents.Event{E: agents.EventOffer, At: agents.StampMilli(rr.now), Brief: id, From: from})
			}
		}
		return allow(RuleCallAllowed, redactTo(in.To))
	}
	c := rr.resolveCallee(rs, in.SubagentType)
	if why := mayCall(w, c); why != "" {
		return refuse(c.shown, why)
	}
	t := in.SubagentType
	role := ""
	if c.role != nil {
		role = c.role.Role.Name
	}
	rr.addSpawn(w, agents.Spawn{Tool: in.Tool, Type: t, Role: role, Name: in.AgentName, Briefs: briefIDs(in.AgentPrompt)})
	shownType := t
	if shownType == "" {
		shownType = generalType
	}
	return allow(RuleCallAllowed, shownType)
}

// redactTo is a SendMessage's to as a record's target: an id or name as written (ids and names are no secret), cut.
func redactTo(to string) string {
	if to == "" {
		return ""
	}
	return oneLine(to)
}

// addSpawn keeps an allowed Agent, Task or Workflow call for the session; one that cannot be kept is noted.
func (rr *roleRun) addSpawn(w *who, sp agents.Spawn) {
	sp.At, sp.Use = agents.StampMilli(rr.now), rr.in.ToolUseID
	switch {
	case w.h.SessRole:
		sp.Parent = "session"
	case rr.in.AgentID != "":
		sp.Caller, sp.Parent = rr.in.AgentID, rr.in.AgentID
	default:
		sp.Caller, sp.Parent = w.key, w.key // a named session: its own id
	}
	if hookin.ID(sp.Use) == "" {
		rr.note(RuleStateUnreadable, "the call has no tool_use_id the guard can keep, so the agent it starts is joined to none")
		return
	}
	if err := rr.store.AddSpawn(rr.in.Session, sp); err != nil {
		rr.note(RuleStateUnreadable, "the call cannot be kept for the session ("+oneLine(err.Error())+"), so the agent it starts "+
			"is joined to none")
	}
}

// resolveTo resolves a SendMessage's to (step 5.5's "SendMessage's target is resolved, or refused"): an agent whose
// state this session holds, by its id; else the newest Agent or Task call the guard allowed that gave that name; else
// nothing. main, and another session's id, resolve to no agent. target is the agent's state name, "" for none.
func (rr *roleRun) resolveTo(rs *roleSet, to string) (callee, string) {
	session := rr.in.Session
	if to == "main" {
		return callee{main: true, shown: "main (the person's session)"}, ""
	}
	if hookin.ID(to) != "" && session != "" {
		if st, err := rr.store.Read(session, to); err == nil {
			h := st.Header
			switch {
			case h.SessRole || h.Named:
				return callee{main: true, shown: "another session"}, ""
			case h.Role == "":
				return callee{shown: "the agent " + quote(to, 40) + " of a type no pack declares"}, to
			}
			r := rs.byType(h.Pack + ":" + h.Role)
			if r == nil {
				// The role went from the lock since it started: judged by the facts it started with.
				r = &workspace.RoleInForce{Pack: h.Pack, Role: format.PackRole{Name: h.Role, Kind: h.Kind}}
			}
			return callee{role: r, shown: "the agent " + quote(to, 40) + " of the " + h.Kind + " role " + quote(h.Role, 40)}, to
		}
	}
	if spawns, err := rr.store.Spawns(session); err == nil && to != "" {
		for i := len(spawns) - 1; i >= 0; i-- {
			sp := spawns[i]
			if sp.Name != to {
				continue
			}
			c := rr.resolveCallee(rs, sp.Type)
			c.shown = "the agent named " + quote(to, 40) + " (" + c.shown + ")"
			return c, rr.claimedBy(sp.Use)
		}
	}
	return callee{unresolved: true, shown: quote(to, 60) + ", which names no agent the guard saw started"}, ""
}

// claimedBy gives the agent that a spawn call started, as its claim names it, "" when none has made a call yet.
func (rr *roleRun) claimedBy(use string) string {
	if hookin.ID(use) == "" || rr.in.Session == "" {
		return ""
	}
	raw, err := os.ReadFile(filepath.Join(rr.store.SessionDir(rr.in.Session), agents.ClaimsFolder, use))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

// clock gives the run's time.
func (j *judge) clock() time.Time {
	if j.now != nil {
		return j.now()
	}
	return time.Now()
}

// minutesSince is the whole minutes from start to now (0 when now is before it).
func minutesSince(start, now time.Time) int64 {
	if now.Before(start) {
		return 0
	}
	return int64(now.Sub(start) / time.Minute)
}

// firstLineTool reports a tool the guard's first line judges where the lock declares roles (engine.GuardTools, which a
// test holds this to): the file tools, the shells, and the tools that start or resume an agent.
func firstLineTool(tool string) bool {
	return hookin.FileTool(tool) || shellTool(tool) || hookin.SpawnTool(tool)
}

// judgedTool reports a tool whose input the guard reads (hookin.ReadJudged): every other tool's is passed over, so a
// tool Claude Code adds later, or an input the strict decoder refuses, is counted, not blocked.
func judgedTool(tool string) bool { return firstLineTool(tool) || tool == "Read" }

// roleDecide is decide's part where the roles apply (this file's comment, budget.go's): the caller found and counted,
// its budget judged, the spawn tools judged by who may call whom, the file tools refused to a role that changes
// nothing; done is false when the call goes on to the plain judgment, which then runs with the roles' shell part.
func (j *judge) roleDecide(rr *roleRun, in *hookin.Payload, base string) (d Decision, rec *linkedProject, done bool) {
	w, err := rr.resolve()
	if err != nil {
		return rolesFault(err), j.session, true
	}
	if w == nil {
		return Decision{}, nil, false // no role is committed: as before
	}
	rr.caller = w
	j.sh = &shellRole{rr: rr, w: w}
	reads := ""
	switch {
	case in.Tool == "Read":
		reads = rr.briefPath(in.ReadPath)
	case shellTool(in.Tool):
		reads = briefShow(in.Tool, in.Command)
	}
	if w.counted {
		c := rr.count(w, reads)
		if c.spent != nil {
			ok, report := rr.exempt(w, reads)
			if !ok {
				return rr.spentRefusal(w, c), j.session, true
			}
			d, rec := j.plainDecide(in, base)
			if !d.Allow {
				return d, rec, true
			}
			if report {
				_ = rr.store.Append(in.Session, w.key, agents.Event{E: agents.EventReported, At: agents.StampMilli(rr.now)})
			}
			return allow(RuleBudgetExempt, d.Target), j.session, true
		}
	}
	switch {
	case hookin.SpawnTool(in.Tool):
		return rr.spawnCall(w), j.session, true
	case hookin.FileTool(in.Tool) && !w.h.SessRole && !w.h.Writes:
		d := deny(RuleRoleWritesNothing, fmt.Sprintf("%s refused: %s changes nothing (its role's writes: false), so the file "+
			"tools are refused to it", in.Tool, w.shown()), "put what you would change in your report; your caller, or a role "+
			"that writes, makes the change")
		return d, j.session, true
	}
	d, rec = j.plainDecide(in, base)
	if d.Allow {
		rr.noteClimbs(w)
	}
	return d, rec, true
}
