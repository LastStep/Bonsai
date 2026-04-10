# Bonsai Research — Core Concept & Landscape Analysis

> Research conducted 2026-04-02. Context for upcoming design discussions.

---

## 1. What Bonsai Is

Bonsai is a CLI scaffolding tool (`pip install bonsai-agents`) that generates structured agent workspaces for multi-agent Claude Code projects.

**The core problem it solves:** When you run multiple Claude Code agents on the same project, they have no inherent awareness of each other. They'll overwrite files, make conflicting architectural decisions, lose context between sessions, and drift from the plan. Bonsai creates the scaffolding that prevents this — structured isolation, mandatory planning, cross-session memory, and accountability through reports.

**The core mechanism:** A catalog of composable instruction templates (agents, skills, workflows, protocols) that get rendered via Jinja2 into a target project's workspace directories, giving each agent a complete instruction set for its role.

**The flow:**
```
bonsai init   → project scaffolding (docs, status tracking, routing)
bonsai add    → agent workspace (identity, memory, skills, workflows, protocols)
agents work   → plan → execute → report → review cycle
```

**Three agent types today:** Tech Lead (architects, plans, reviews), Backend (executes backend plans), Frontend (executes frontend plans). Each gets a tailored identity, memory template, and compatible catalog items.

---

## 2. The Landscape — What Others Are Doing

### 2.1 Get Shit Done (GSD)

**Repo:** github.com/gsd-build/get-shit-done  
**Core thesis:** Context rot kills quality. Fresh context windows per agent stage solve it.

GSD is a **lifecycle orchestrator**. It prescribes a rigid pipeline: Discuss → Plan → Execute → Verify → Ship. At each stage, it spawns fresh agents with clean 200k-token context windows, avoiding the accumulated garbage problem that degrades output quality.

**Key innovations:**
- **Wave-based parallelization** — independent tasks run simultaneously, dependent ones sequentially, each in a fresh context
- **XML-structured plans** with built-in verification (`<task>`, `<files>`, `<action>`, `<verify>`, `<done>`)
- **Atomic commits** — one git commit per completed task
- **Model profiles** — quality/balanced/budget modes assigning Opus/Sonnet/Haiku to different stages
- **`.planning/` directory** — persistent knowledge structure (PROJECT.md, REQUIREMENTS.md, ROADMAP.md, STATE.md, CONTEXT.md) kept small to avoid context degradation

**What GSD is NOT:** It has no concept of agent identity or memory files. Agents are ephemeral by design — they get fresh context, do their job, and die. The system doesn't care who the agent "is", only what stage of the pipeline it's in.

**Philosophy:** Rejects enterprise ceremony (sprints, story points, Jira). Embeds sophistication in the system so the user workflow stays simple. *"The complexity is in the system, not in your workflow."*

### 2.2 Everything Claude Code (ECC)

**Repo:** github.com/affaan-m/everything-claude-code  
**Core thesis:** Agent performance is an engineering discipline. Tokens are currency.

ECC is a **performance harness** — a plugin delivering 36 specialized subagents, 150+ skills, 68 command shims, rules across 12 language ecosystems, and hook-based automation. Born from an Anthropic hackathon win. 50K+ stars.

**Key innovations:**
- **Bounded subagents** — each agent has a defined role, limited tool scope, and specific model assignment (not just "do everything")
- **Continuous learning pipeline** — sessions generate patterns → `evaluate-session` hook extracts "instincts" with confidence scores → `/evolve` clusters instincts into reusable skills → skills become permanent knowledge
- **Token consciousness** — Sonnet for 80% of tasks, Opus for deep reasoning, Haiku for subagent work; max thinking tokens capped; auto-compact threshold lowered from 95% to 50%
- **Cross-platform** — works with Claude Code, Cursor, Codex, OpenCode

**What ECC is NOT:** It's a plugin that distributes pre-built content, not a generator that creates project-specific content. You install ECC and get the same 150 skills regardless of your project. There's no workspace isolation concept — it's about making one agent better, not coordinating many.

### 2.3 Other Notable Projects

- **claude-playbook** — Production-ready `.claude/` scaffolding as a GitHub template (rules, skills, agents, hooks pre-configured)
- **claude-code-templates** — CLI tool with 100+ agents, commands, settings via interactive web interface
- **claude-skills** — 220+ skills across engineering, marketing, product, compliance
- **ruflo** — Multi-agent swarm orchestration with distributed intelligence

---

## 3. The Primitives Everyone Converges On

The ecosystem has settled on a common vocabulary:

| Primitive | What it is | How it loads |
|-----------|-----------|--------------|
| **CLAUDE.md** | Project constitution — always loaded | Auto-read every session |
| **Skills** | On-demand domain knowledge & workflows | Model-invoked via pattern matching |
| **Agents/Subagents** | Bounded executors with limited tool scope | Spawned by orchestrator |
| **Hooks** | Event-driven automation (pre/post tool use) | Fired on tool events |
| **Rules** | Always-follow governance | Loaded from `~/.claude/rules/` |
| **Plugins** | Distributable bundles of the above | Installed via CLI/marketplace |

Everyone agrees: unstructured "vibecoding" breaks at scale. The disagreement is about *what structure* matters most.

---

## 4. Philosophical Comparison — Where Bonsai Sits

| Dimension | GSD | ECC | Bonsai |
|-----------|-----|-----|--------|
| **Core metaphor** | Workflow pipeline | Performance harness | Workspace scaffolder |
| **Primary problem** | Context rot during long tasks | Token waste & inconsistency | Cold-start: how to set up agents that won't collide |
| **Unit of work** | Pipeline stage (discuss/plan/execute/verify) | Individual tool call + skill invocation | Agent workspace (identity + memory + instructions) |
| **Agent identity** | None — agents are ephemeral stage workers | Agent .md files with tool/model boundaries | Rich identity system (role, mindset, relationships) |
| **Memory model** | STATE.md + threads for cross-session persistence | Hook-based session extraction → instinct evolution | memory.md template (flags, work state, notes) |
| **Self-awareness** | Not explicit (fresh context sidesteps the issue) | Token budgets, compaction thresholds | Explicit self-awareness.md with context monitoring |
| **Multi-agent coordination** | Orchestrator spawns fresh agents per stage | Bounded delegation to subagents | Workspace isolation + routing tables + shared docs |
| **What ships** | A lifecycle process | A catalog of pre-built capabilities | A generated, project-specific workspace |

### The Deeper Distinction

**GSD asks:** "How should work flow through agents?" → Answer: a prescribed lifecycle.

**ECC asks:** "How should agents perform their work?" → Answer: bounded tools, token discipline, evolving skills.

**Bonsai asks:** "How should agents understand their role and boundaries?" → Answer: identity, memory, protocols, workspace isolation.

These are three different layers of the same problem. GSD is the process layer. ECC is the capability layer. Bonsai is the identity and coordination layer.

---

## 5. What Bonsai Uniquely Gets Right

### 5.1 Agent Identity as a First-Class Concept

No other project in the ecosystem treats agent identity with this level of care. Bonsai's identity system answers:
- **Who am I?** (role, mindset — executor vs. architect)
- **Who do I answer to?** (user → tech lead → peer agents)
- **What's my relationship to other agents?** (we don't touch each other's code)
- **What takes priority?** (project files override external tools)

This is important because Claude Code agents are stateless by default. Without explicit identity, every session starts from scratch — the agent doesn't know if it should architect or implement, plan or execute, review or code.

### 5.2 Workspace Isolation via Convention

The routing table + scope boundaries protocol is a simple but effective coordination mechanism. Instead of technical sandboxing (which Claude Code doesn't support), Bonsai creates *instructional* boundaries — agents are told their workspace and told never to cross it. This works because LLMs reliably follow clear, well-motivated instructions.

### 5.3 The Plan-Then-Execute Discipline

The separation of Tech Lead (plans) from Backend/Frontend (executes) enforces a critical discipline: **no agent writes code without a plan.** Plans must be explicit enough that "the agent makes zero design decisions." This prevents architectural drift and ensures that the human (or tech lead) retains control over design while delegating implementation.

### 5.4 Cross-Session Continuity

The memory.md system (flags, work state, notes) with explicit cleaning rules gives agents a lightweight but effective way to maintain state across sessions. Combined with the session-start protocol (read identity → memory → status → logs → field notes), every session begins with the agent re-grounding itself in context.

### 5.5 Scaffolding as a Generator, Not a Plugin

Unlike ECC (which distributes the same pre-built content to everyone), Bonsai generates project-specific files rendered with your project name, description, and configuration. The files become *yours* — editable, versionable, and tailored to the specific project.

---

## 6. The Core Idea — What Bonsai Is Really About

Stripping away the features and looking at the essence:

**Bonsai is a context engineering tool for multi-agent Claude Code projects.**

"Context engineering" means: giving an AI agent exactly the right information, at the right time, in the right structure, so it can do its job well. Bonsai does this by:

1. **Establishing identity** — so the agent knows its role and boundaries
2. **Providing memory** — so the agent maintains state across sessions
3. **Encoding protocols** — so hard rules are always loaded and followed
4. **Supplying workflows** — so the agent knows how to do complex multi-step work
5. **Attaching skills** — so domain-specific knowledge is available on demand
6. **Creating shared infrastructure** — so agents can coordinate through artifacts (plans, reports, status, logs)

The CLI and catalog are just the delivery mechanism. The real product is **a well-structured instruction set that makes Claude Code agents predictable, accountable, and coordinated.**

The closest analogy: Bonsai is to Claude Code agents what a well-run engineering org is to software engineers. It provides role definitions (identity), onboarding (session-start protocol), standard operating procedures (workflows), hard rules (protocols), domain knowledge (skills), and project management infrastructure (scaffolding). The agents are the engineers; Bonsai is the organizational structure that makes them effective together.

---

## 7. Open Questions for Discussion

These are not feature requests — they're conceptual tensions worth exploring:

1. **Identity depth vs. flexibility** — The three agent types (tech-lead, backend, frontend) are a good start, but real projects have more roles. How deep should identity customization go? Is Bonsai opinionated about roles, or is it a blank canvas?

2. **Convention vs. enforcement** — Workspace isolation is convention-based (instructions, not sandboxing). This works but is fragile. Is there a way to make boundaries more structural without losing flexibility?

3. **Static scaffolding vs. living system** — Once `bonsai add` generates files, those files are static. The agent evolves its memory.md, but skills/workflows/protocols don't evolve. Should the catalog items be living documents that adapt?

4. **Coordination through artifacts vs. direct communication** — Agents coordinate through shared files (plans, reports, status). Claude Code now supports agent teams with direct peer messaging. Is artifact-based coordination the right model, or should agents talk to each other?

5. **Where does Bonsai end?** — Is Bonsai a scaffolding tool (generates files, steps away), a development methodology (prescribes how agents should work), or a runtime orchestrator (actively manages agent execution)? The answer shapes everything.

6. **Catalog as library vs. catalog as framework** — Current catalog items are opinionated (specific workflows, specific protocols). Should they be more generic/customizable, or is the opinionation the value?

---

## Sources

- Bonsai source code (src/bonsai/, agent/, catalog/)
- [Get Shit Done](https://github.com/gsd-build/get-shit-done/) — README, architecture docs
- [Everything Claude Code](https://github.com/affaan-m/everything-claude-code) — README, plugin structure
- [Writing a Good CLAUDE.md — HumanLayer](https://www.humanlayer.dev/blog/writing-a-good-claude-md)
- [Claude Agent Skills: A First Principles Deep Dive](https://leehanchung.github.io/blogs/2025/10/26/claude-skills-deep-dive/)
- [A Mental Model for Claude Code: Skills, Subagents, and Plugins](https://levelup.gitconnected.com/a-mental-model-for-claude-code-skills-subagents-and-plugins-3dea9924bf05)
- [The Code Agent Orchestra — Addy Osmani](https://addyosmani.com/blog/code-agent-orchestra/)
- [claude-playbook](https://github.com/smartwhale8/claude-playbook), [claude-code-templates](https://github.com/davila7), [claude-skills](https://github.com/alirezarezvani)
