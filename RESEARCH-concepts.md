# Bonsai — Core Concept Decisions

> Written 2026-04-02. Captures conceptual positions reached during design discussion.

---

## 1. What Bonsai Is

Bonsai is a **methodology for multi-agent Claude Code development**, delivered through a generator (CLI) that scaffolds project-specific agent workspaces.

The CLI and catalog are the delivery mechanism. The real product is the methodology — a system of identity, memory, protocols, workflows, and skills that makes agents predictable, accountable, and coordinated.

---

## 2. Ambient vs. Command-Driven (Bonsai's Key Differentiator)

Other systems (GSD, ECC) are **command-driven** — you trigger a command, the agent enters a mode, does the thing, exits. Without the trigger, it's default Claude.

Bonsai is **ambient** — the agent reads its instruction set at session start and *is* that role. No commands needed for baseline behavior. You open Claude Code in a workspace and start talking. The agent already knows who it is, what it should do, and where its boundaries are.

**Commands are optional accelerators, not the operating model.** The routing in CLAUDE.md and the protocol/workflow structure serve as an always-loaded command registry. Explicit triggers (slash commands, direct references like "run the code-review workflow") skip the inference step and go straight to execution — but the agent should find its way there naturally from its instruction set.

The spectrum:
- **Always on:** Identity + protocols (the agent is always this role, always follows hard rules)
- **Internalized but activatable:** Workflows + skills (agent knows them, reaches for them naturally, but explicit triggers accelerate)

---

## 3. Catalog Ownership — Three Layers

The catalog operates at three levels:

1. **Bonsai upstream** — ships with the package, continuously improved. Encodes methodology best practices. This is the reference implementation.
2. **Project-generated** — `bonsai add` renders catalog items into local workspace files. Files become project-owned at this point.
3. **Project-customized** — user edits local files for project-specific needs (domain security rules, custom workflows, identity tweaks).

**Update model:** Generator never overwrites. Once generated, files are yours. When upstream improves, you can look at new catalog versions and decide whether to adopt. No forced updates — local edits reflect project-specific reality that upstream can't know.

**Implication:** Catalog quality is critical. It's the first impression of the methodology. Weak defaults = weak projects. The eval system is the mechanism for ensuring upstream catalog quality.

---

## 4. Authority Hierarchy

```
Human (decides)
  ↑ reports to
Meta-layer (observes the system)
  ↑ operates within
Tech Lead (architects within the system)
  ↑ executes plans from
Backend / Frontend (implements)
```

Each layer has one verb:
- **Implements** — Backend/Frontend execute plans, don't architect
- **Architects** — Tech Lead plans and reviews, doesn't execute
- **Observes** — Meta-layer watches system health, doesn't modify
- **Decides** — Human has final authority on all changes

No layer does the job of another. Authority flows upward. Problems escalate, never get silently fixed by a lower layer.

---

## 5. The Meta-Layer (Outside the Paradigm)

### What it is

A layer that exists **outside** the agent paradigm. It doesn't have a workspace, doesn't follow session-start protocol, doesn't have an identity in the agent sense. It observes the system itself rather than operating within it.

### What it watches

- Are protocols contradicting each other across agents?
- Is an agent's behavior drifting from its identity?
- Are reports being produced but never reviewed?
- Is the routing working — are agents reading the right files?
- Is shared infrastructure (status, plans, logs) staying coherent?

### Why it's needed

Agents inside the system can only see their own workspace. Self-audit (via memory protocol) is introspection — an agent checking itself. The meta-layer provides **system-wide observation** that no individual agent can perform.

The tech-lead comes closest (it reviews agent output) but it's still a participant bound by scope rules. The meta-layer is not a participant — it's infrastructure.

### Authority

**Reports to the human only.** Does not modify the system autonomously. Watches, diagnoses, recommends. The human decides what to act on.

**Rationale:** Autonomous modification creates unpredictable behavior — the exact problem Bonsai exists to prevent. The meta-layer follows the same escalation principle as everything else in the methodology.

### Where it lives

Open question. It can't be in a workspace (needs to see all workspaces). It can't be in the catalog (it's not a scaffolded role). It's likely something Bonsai itself provides rather than something the generator produces.

### Relationship to evals

- **Evals** = development-time testing of the methodology (before deployment)
- **Meta-layer** = runtime observation of the methodology (during use)

Same concern (is the methodology working?), different timescales.

---

## 6. Talents — A New Catalog Category

### The problem that surfaced it

During this research conversation, the question arose: when should an agent stop and commit accumulated knowledge to a document? This is a judgment call — not a hard rule (protocol), not a step-by-step procedure (workflow), not domain knowledge (skill). It's a natural behavioral pattern that some agents need and others don't.

### What talents are

Talents are **innate behavioral aptitudes** — things the agent *is naturally good at*, not things it's told to do. They have no checklist, no trigger, no hard rules. They describe a quality of how the agent works, and the agent internalizes them as part of its nature.

The human analogy: a talented researcher doesn't follow a "when to take notes" checklist. They just feel when accumulated knowledge needs to be captured. It's not a process — it's an aptitude.

### How talents differ from everything else

| Category | Nature | Loading | Example |
|----------|--------|---------|---------|
| **Protocols** | Hard rules, constraints | Always loaded, session start | "Never cross workspace boundaries" |
| **Workflows** | Step-by-step procedures | Internalized, triggered by situation | "When assigned a plan, follow these steps" |
| **Skills** | Domain knowledge | On-demand, reached for when relevant | "Here's how we write tests" |
| **Talents** | Innate behavioral aptitudes | Always active, ambient from session start | "Naturally sense when to crystallize knowledge" |

### Key properties

- **Ambient** — loaded at session start like protocols, active throughout the session
- **Soft** — not constraints or checklists. Described as natural tendencies, not imperatives
- **Composable** — assigned per agent type via catalog meta.yaml, just like skills/workflows/protocols
- **Closer to core than to protocols** — talents shape *who the agent is*, not what rules it follows. They're almost an extension of identity, but modular and composable from the catalog

### Example talents (conceptual, not yet designed)

- **Knowledge commitment** — naturally sensing when accumulated decisions/discoveries should be captured before moving on. Good for: researcher, tech-lead during discovery
- **Architectural intuition** — sensing when a design is getting too complex or when components should be split. Good for: tech-lead
- **Regression sensing** — noticing when a change is likely to break something elsewhere. Good for: backend, frontend executors
- **Simplification instinct** — gravitating toward simpler solutions, resisting unnecessary complexity. Good for: all agents, but especially executors

### Where talents live (leaning, not decided)

Closer to core than to the other categories. Tentative structure:

```
agent/
├── Core/
│   ├── identity.md       ← who you are
│   ├── memory.md         ← what you remember
│   └── self-awareness.md ← monitoring yourself
├── Talents/              ← what you're naturally good at (closer to core)
├── Protocols/            ← what you must always do
├── Workflows/            ← how you do specific activities
└── Skills/               ← what you know
```

### Open questions

- Should talents be in `Core/Talents/` (nested under core) or `Talents/` (peer to protocols)?
- How is the "soft" language of talents written so the agent internalizes it as nature rather than instruction?
- What's the right number of talents per agent? Too many dilutes the concept.
- Can users define custom talents, or are they upstream-only?

---

## 7. The Instruction Taxonomy (Complete Picture)

Everything an agent loads, organized by what it is and how it works:

```
CORE (who you are — always loaded, defines the agent)
├── Identity      — role, mindset, relationships, authority
├── Memory        — cross-session state (flags, work state, notes)
├── Self-awareness — context monitoring, hard thresholds
└── Talents       — innate aptitudes, ambient behavioral patterns

PROTOCOLS (what you must do — always loaded, non-negotiable)
├── Session start  — boot sequence
├── Scope boundaries — workspace isolation
├── Security       — hard security rules
└── Memory protocol — how to read/write/clean memory

WORKFLOWS (how you do things — internalized, activated by situation)
├── Plan execution — follow assigned plan
├── Planning       — create plans for executors
├── Code review    — review agent output
├── Reporting      — write completion reports
└── Session logging — end-of-session capture

SKILLS (what you know — on-demand, reached for when relevant)
├── Coding standards — language conventions
├── Testing         — test patterns and requirements
├── Design guide    — UI/UX conventions
├── Database conventions — schema and query patterns
└── Planning template — plan structure and rules
```

The loading gradient: Core (always, defines identity) → Talents (always, shapes behavior) → Protocols (always, enforces rules) → Workflows (internalized, situationally activated) → Skills (available, pulled when needed).

---

## 8. Positions Taken

| Question | Position |
|----------|----------|
| Is Bonsai a generator or a methodology? | Methodology, delivered through a generator |
| Command-driven or ambient? | Ambient baseline, commands as optional accelerators |
| Can the meta-layer modify the system? | No — reports to human only |
| Catalog: starter kit or upstream source? | Both — starter kit that upstream continuously improves |
| Do generated files get auto-updated? | No — generated once, then user-owned |
| Is knowledge-commitment a protocol, skill, or new category? | New category: talent |
| Are talents always loaded? | Yes — ambient from session start, like protocols but soft |
| Where do talents sit in the hierarchy? | Closer to core than to protocols |

---

## 9. Open Questions (Not Yet Decided)

- How generic should agent types be? (Currently: tech-lead/backend/frontend. Should it support arbitrary roles?)
- How does the meta-layer get implemented? (Part of Bonsai CLI? Separate tool? A special agent definition?)
- How do agents discover and respond to each other's work beyond shared artifacts?
- What's the right balance of routing detail to ensure agents find workflows without explicit triggers?
- Should the catalog include meta-layer observation rules, or are those separate from the methodology?
- Where exactly do talents live in the directory structure? (Core/Talents/ vs Talents/)
- How is talent language written to feel innate rather than instructional?
- What's the right number of talents per agent type?
- Can users define custom talents or are they upstream-only?
