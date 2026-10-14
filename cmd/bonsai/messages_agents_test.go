package main

// bonsai messages' agent column (note 5.5.11, 9): each thread's recipient, the agent tied to its newest brief, with its
// counts and limits as the guard recorded them and spent or not, read from the guard's state under
// .bonsai/local/agents/; budget-spent raised while that agent is spent; and unanswered, once an agent tied the brief,
// read from its last call, not the brief's age.

import (
	"strings"
	"testing"
	"time"

	"github.com/LastStep/Bonsai/internal/agents"
	"github.com/LastStep/Bonsai/internal/asks"
	"github.com/LastStep/Bonsai/internal/schema"
)

func TestMessagesAgentColumn(t *testing.T) {
	c := newCLI(t)
	t.Setenv(asks.SessionEnv, "")
	start := time.Date(2026, 10, 16, 9, 0, 0, 0, time.UTC)
	clock := start
	old := messageNow
	t.Cleanup(func() { messageNow = old })
	messageNow = func() time.Time { return clock }
	c.link()
	c.commit("link")
	_, out, _ := c.run("", "message", "--from", "orchestrator", "--to", "builder@T-0901", "--kind", "brief", "--body", "Go.")
	id := strings.Split(out, "\n")[0]
	thread := func() schema.Object {
		th := threadsIn(c.messagesDoc(0, "messages"))
		if len(th) != 1 {
			t.Fatalf("%d threads", len(th))
		}
		return th[0]
	}
	flags := func(th schema.Object) string {
		f, _ := th.Get("flags")
		var out []string
		for _, x := range f.([]any) {
			out = append(out, x.(string))
		}
		return strings.Join(out, ",")
	}
	// No agent tied it: the recipient null, unanswered by the brief's age alone.
	if r, _ := thread().Get("recipient"); r != nil {
		t.Errorf("a recipient before any tie: %v", r)
	}
	clock = start.Add(60 * time.Minute)
	if got := flags(thread()); got != "unanswered" {
		t.Errorf("an hour, no agent: %q", got)
	}
	// An agent ties it (as the guard writes its state): its brief's 3 actions, two calls made.
	s := agents.Store{Main: c.root}
	n := func(v int64) *int64 { return &v }
	h := agents.Header{Agent: "a1", Session: "s1", Type: "workflow:builder", Role: "builder", Pack: "workflow", Kind: "lead",
		Writes: true, Limits: agents.Limits{Minutes: agents.Limit{Value: n(240), Max: n(480)}, Actions: agents.Limit{Value: n(100), Max: n(200)}},
		First: agents.Stamp(start)}
	if _, err := s.Create(h); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.Tie(id, "s1", "a1"); !ok || err != nil {
		t.Fatal(ok, err)
	}
	lim := h.Limits
	lim.Actions.Value = n(3)
	if err := s.Append("s1", "a1", agents.Event{E: agents.EventTie, At: agents.StampMilli(start), Brief: id, To: "builder@T-0901",
		Thread: id, From: "orchestrator", Limits: &lim}); err != nil {
		t.Fatal(err)
	}
	for i := int64(1); i <= 2; i++ {
		if _, err := s.Call("s1", "a1", start.Add(time.Duration(i)*time.Minute), i); err != nil {
			t.Fatal(err)
		}
	}
	clock = start.Add(2*time.Minute + 59*time.Minute + 59*time.Second)
	th := thread()
	r, _ := th.Get("recipient")
	want := `{"agent":"a1","counts":{"minutes":2,"actions":2,"red_runs":0},"limits":{"minutes":{"value":240,"max":480},` +
		`"actions":{"value":3,"max":200},"red_runs":{"value":null,"max":null}},"spent":false}`
	if got := schema.Show(r); got != want {
		t.Errorf("the recipient:\n got %s\nwant %s", got, want)
	}
	if got := flags(th); got != "" {
		t.Errorf("the brief three hours old, its agent's last call a second under an hour: %q", got)
	}
	clock = start.Add(2*time.Minute + 60*time.Minute)
	if got := flags(thread()); got != "unanswered" {
		t.Errorf("its agent's last call an hour old: %q", got)
	}
	_, text, _ := c.run("", "messages")
	if !strings.Contains(text, "  flags: unanswered  agent a1: minutes 2 of 240, actions 2 of 3, red runs 0 of no limit\n") {
		t.Errorf("the line's agent column:\n%s", text)
	}
	// Spent: budget-spent, and the recipient says so.
	if err := s.Append("s1", "a1", agents.Event{E: agents.EventSpent, At: agents.StampMilli(start.Add(3 * time.Minute)),
		Measure: "actions", Count: 4, Of: 3}); err != nil {
		t.Fatal(err)
	}
	th = thread()
	if got := flags(th); got != "unanswered,budget-spent" {
		t.Errorf("spent: %q", got)
	}
	if r, _ := th.Get("recipient"); !strings.HasSuffix(schema.Show(r), `"spent":true}`) {
		t.Errorf("spent: %s", schema.Show(r))
	}
	if _, text, _ = c.run("", "messages"); !strings.Contains(text, "red runs 0 of no limit, spent\n") {
		t.Errorf("the line's agent column, spent:\n%s", text)
	}
}
