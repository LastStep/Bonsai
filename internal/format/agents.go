package format

// bonsai agents --json (bonsai.agents/1, set 8; plan-5 5.5.12): the one published view of what Bonsai has recorded
// about each agent, copied as recorded, never interpreted (no derived state, share, cost, estimate or prediction: the
// studio computes those). Printed, never stored. Written by bonsai agents (step 5.5.12), read by the studio and agents. Bonsai's
// own state files under .bonsai/local/agents/ (step 5.5.11) stay internal, with no format.
//
// Beside it, the roles' limits, their one home in Bonsai's code: RoleLimits names the three limits a role declares
// (pack.yaml's roles), a project sets (bonsai.yaml's packs[].roles), a brief gives (a message's budget), status --json
// shows in force and bonsai agents shows as recorded, held by a test to every schema that lists them; HelperLimits are
// Bonsai's fixed limits for an agent of a type no linked pack declares, where no pack declares a helper either.

import (
	"encoding/json"

	"github.com/LastStep/Bonsai/internal/schema"
)

// Agents is the agents document.
type Agents struct {
	Workspace *WorkspaceRef  `json:"workspace"` // null when the command refused before reading one
	Agents    []Agent        `json:"agents"`
	Skipped   []AgentSkipped `json:"skipped"`
	Error     *ErrorObject   `json:"error"`
	Extra     schema.Object  `json:"-"`
}

// Agent is one agent, every field as recorded.
type Agent struct {
	ID       string        `json:"id"`      // Claude Code's agent_id, or a top-level session's id
	Session  string        `json:"session"` // the session it ran in
	Parent   *string       `json:"parent"`  // "session", the starting agent's id, or null
	Role     *string       `json:"role"`    // a declared role's name, or null for an undeclared type
	Type     *string       `json:"type"`    // Claude Code's agent type, or null
	Task     *string       `json:"task"`
	Brief    *string       `json:"brief"` // the tied brief's id
	Model    *string       `json:"model"`
	Started  *string       `json:"started"`
	Ended    *string       `json:"ended"`
	LastCall *AgentCall    `json:"last_call"`
	Counts   *AgentCounts  `json:"counts"` // null for the person's own session
	Limits   *AgentLimits  `json:"limits"` // null for the person's own session
	Spent    bool          `json:"spent"`
	Report   *AgentReport  `json:"report"`
	Log      *string       `json:"log"` // its log file, project-relative, forward slashes
	Thread   *string       `json:"thread"`
	Usage    []AgentUsage  `json:"usage"` // null until its first usage record
	Extra    schema.Object `json:"-"`
}

// AgentCall is an agent's last tool call, as the log records it.
type AgentCall struct {
	At       string        `json:"at"`
	Tool     string        `json:"tool"`
	Category *string       `json:"category"`
	Extra    schema.Object `json:"-"`
}

// AgentCounts are an agent's counts as the guard wrote them.
type AgentCounts struct {
	Minutes *json.Number  `json:"minutes"` // the guard's own figure on the last call line
	Actions int64         `json:"actions"`
	RedRuns int64         `json:"red_runs"`
	Extra   schema.Object `json:"-"`
}

// AgentLimits are the limits that bounded an agent, as the guard wrote them on its start or last tie line.
type AgentLimits struct {
	Minutes AgentLimit    `json:"minutes"`
	Actions AgentLimit    `json:"actions"`
	RedRuns AgentLimit    `json:"red_runs"`
	Extra   schema.Object `json:"-"`
}

// AgentLimit is one limit's value in force and its max; nil for no limit.
type AgentLimit struct {
	Value *int64        `json:"value"`
	Max   *int64        `json:"max"`
	Extra schema.Object `json:"-"`
}

// AgentReport is an agent's report: its id and status, from a report whose from is the agent's own address.
type AgentReport struct {
	ID     string        `json:"id"`
	Status string        `json:"status"` // done, blocked, needs-input
	Extra  schema.Object `json:"-"`
}

// AgentUsage is one model's token counts as the agent's newest usage record holds them, or an unknown item.
type AgentUsage struct {
	Model      *string       `json:"model"`
	Input      *int64        `json:"input"`
	Output     *int64        `json:"output"`
	CacheRead  *int64        `json:"cache_read"`
	CacheWrite *int64        `json:"cache_write"`
	Kind       *string       `json:"kind"`   // null, or unknown
	Reason     *string       `json:"reason"` // why an unknown item's transcript could not be read
	Extra      schema.Object `json:"-"`
}

// AgentSkipped is a state file that could not be read.
type AgentSkipped struct {
	Path   string        `json:"path"`
	Reason string        `json:"reason"`
	Extra  schema.Object `json:"-"`
}

// ReadAgents reads bonsai.agents/1's document.
func ReadAgents(raw []byte) (*Agents, error) {
	a := &Agents{}
	if err := MustLookup("agents").readJSONInto(raw, a); err != nil {
		return nil, err
	}
	return a, nil
}

// Encode writes the document, held to bonsai.agents/1.
func (a *Agents) Encode() ([]byte, error) { return MustLookup("agents").Encode(a) }

// RoleLimits are the three limits of a role, in the order every schema lists them (pack.yaml's roles, bonsai.yaml's
// packs[].roles, a message's budget, status --json's roles, bonsai agents' counts and limits; TestRoleLimitsAreTheSchemas
// holds each to this table): their one home in Bonsai's code. Each limit has a default and a max (pack.yaml, then
// bonsai.yaml), each a whole number or null for none; a brief gives any value up to the max in force.
var RoleLimits = []Word{
	{Word: "minutes", Means: "minutes since the agent's first call, or since it tied its newest brief: its thinking before its first call is not counted, time idle between calls is"},
	{Word: "actions", Means: "tool calls the agent makes, refused ones counted, from the same start"},
	{Word: "red_runs", Means: "failed ladder runs in a row, read from each climb's own result: a green climb sets the count back to 0, a new brief does not"},
}

// HelperLimit is one of Bonsai's fixed helper limits: its default and its max, nil for no limit.
type HelperLimit struct {
	Limit   string // one of RoleLimits
	Default *int64
	Max     *int64
}

func limit(n int64) *int64 { return &n }

// HelperLimits are Bonsai's own limits for an agent of a type no linked pack declares (Explore, Plan, a fork), where
// no linked pack declares a helper either (a project made from the template, whose one role is a lead): 30 minutes,
// 150 actions, no red runs, each max equal to its default (plan-5 step 5.5, "Role rules and stall stops"). Where a
// pack declares helpers, such an agent runs on the limits in force of the declared helper with the smallest default
// actions instead. Their one home, on the reference page; in RoleLimits' order.
var HelperLimits = []HelperLimit{
	{Limit: "minutes", Default: limit(30), Max: limit(30)},
	{Limit: "actions", Default: limit(150), Max: limit(150)},
	{Limit: "red_runs", Default: nil, Max: nil},
}
