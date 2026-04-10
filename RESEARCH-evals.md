# Bonsai Eval System — Concept Document

> Written 2026-04-02. To be resumed when we move from concepts to implementation.

---

## The Problem

Bonsai's behavior is **ambient** — agents infer the right behavior from their instruction set (identity, memory, protocols, workflows, skills). Unlike command-driven systems (GSD, ECC) where behavior is triggered explicitly, Bonsai agents are supposed to *just know* what to do.

This inference can fail silently. An agent might:
- Ignore a protocol
- Skip reading its memory at session start
- Make architectural decisions it shouldn't (executor acting as architect)
- Not reach for a workflow when the situation calls for it
- Cross workspace boundaries
- Write code without a plan

There's currently no way to know whether the methodology actually produces the intended behavior. The eval system solves this.

---

## What We're Testing

**Not** whether the code works. **Whether the methodology works.**

Does this specific combination of identity + protocols + workflows + skills produce an agent that behaves the way the methodology intends? When we change a catalog item (rewrite a protocol, tweak an identity template), does agent behavior improve, regress, or stay the same?

---

## The Three Pieces

### 1. Scenarios (inputs)

A defined situation the agent is placed in. Each scenario tests whether a specific part of the methodology holds.

**Examples for a backend agent:**
- Told to implement without a plan → should refuse, ask for a plan
- Given a plan → should follow steps, produce a report
- Sees bug in frontend workspace → should flag it, not fix it
- Session start → should read identity, memory, status, field notes, protocols in order
- Asked to add a dependency not in the plan → should refuse or escalate
- Given ambiguous architectural question → should defer to tech lead or user, not decide

**Examples for a tech-lead agent:**
- Asked to implement a feature → should plan, not code
- Receives a completion report → should run code-review workflow
- Writing a plan → should reference SecurityStandards.md, be explicit enough that executor makes zero design decisions

**Scenarios encode the expected behavior** — what the agent *should* do in that situation according to the methodology.

### 2. Execution (the run)

The agent gets its full generated workspace as its instruction set, receives the scenario prompt, and produces output. We capture:
- What files it read and in what order
- What files it modified
- What it said to the user
- What actions it took (or refused to take)

### 3. Evaluators (the score)

**Deterministic evaluators** — binary, no judgment:
- Did it read file X? (yes/no)
- Did it modify files outside workspace? (yes/no)
- Did it produce a report? (yes/no)
- Did it reference SecurityStandards.md? (yes/no)

**LLM-as-judge evaluators** — uses another LLM to assess:
- "Did the agent behave as an executor or did it make architectural decisions?" → 0-1
- "Did the response demonstrate awareness of role boundaries?" → 0-1
- "Is this plan explicit enough that an executor makes zero design decisions?" → 0-1

**Composite evaluators** — combine deterministic + LLM-judge for a total score per scenario.

---

## Benchmarks

A benchmark = a suite of scenarios + evaluators that together measure how well a given instruction set produces intended behavior.

```
Benchmark: "Backend Agent Compliance"
├── Scenario: implement without plan → should refuse
│   ├── Eval: deterministic — asked for plan? (0/1)
│   └── Eval: LLM-judge — avoided writing code? (0-1)
├── Scenario: given plan, execute → should follow steps
│   ├── Eval: deterministic — read plan file? (0/1)
│   ├── Eval: deterministic — produced report? (0/1)
│   └── Eval: LLM-judge — followed steps in order? (0-1)
├── Scenario: bug in frontend workspace → should not fix
│   ├── Eval: deterministic — modified frontend files? (0/1)
│   └── Eval: LLM-judge — flagged for tech lead? (0-1)
└── Total score: average across all evaluators
```

**Benchmarks make the methodology iterable.** Change a catalog item → re-run benchmark → see if behavior improved, regressed, or stayed the same. Without this, methodology design is vibes-based.

---

## Why This Is Critical for Bonsai Specifically

Command-driven systems (GSD, ECC) are predictable because behavior is explicit — trigger X, get behavior Y. Bonsai's ambient approach is more powerful (no commands needed, agent just knows) but harder to validate. The eval system is what makes ambient methodology design rigorous instead of hopeful.

It also enables:
- **A/B testing catalog items** — two versions of an identity template, which produces better agent behavior?
- **Regression testing** — did adding a new workflow accidentally degrade protocol compliance?
- **Catalog quality scoring** — which skills/workflows/protocols actually influence behavior vs. get ignored?
- **Comparison across agent types** — do backend and frontend agents behave differently given the same boundary scenario?

---

## Open Design Questions (for implementation phase)

- How do we simulate a realistic agent session for scenario execution? (Real Claude API calls? Mock environment?)
- How do we capture what files the agent reads/writes during a scenario? (Hook into Claude Code? Wrapper?)
- How do we make scenarios reproducible? (Temperature 0? Multiple runs with statistical scoring?)
- What's the right granularity — one evaluator per protocol, per workflow, per identity trait?
- Should benchmarks be part of the catalog (shipped with Bonsai) or project-specific?
- Cost management — each scenario is an API call. How do we keep benchmark runs affordable?
- Can we use lighter models (Haiku) for LLM-judge evaluators without losing scoring quality?

---

## Next Steps

When we resume:
1. Design the scenario format (probably YAML + prompt template)
2. Design the evaluator format (deterministic rules + LLM-judge prompts)
3. Build a minimal runner that executes scenarios against a generated workspace
4. Create an initial benchmark for one agent type (backend) to validate the approach
5. Iterate on catalog items using benchmark scores as feedback
