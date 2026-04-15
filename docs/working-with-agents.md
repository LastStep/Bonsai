# Working With Agents

A guide to communicating effectively with Bonsai-powered agents. This isn't about what buttons to press — it's about the **language patterns, framing techniques, and collaboration rhythms** that produce the best results.

If you haven't set up a workspace yet, start with the [README](../README.md). If you want to understand the system architecture, read the [Handbook](../HANDBOOK.md).

---

**Table of Contents**

- [The Core Rhythm](#the-core-rhythm)
- [Starting a Session](#starting-a-session)
- [How to Frame Requests](#how-to-frame-requests)
- [The Power of Terse Communication](#the-power-of-terse-communication)
- [Quality Gates](#quality-gates)
- [Trust and Delegation](#trust-and-delegation)
- [Correcting Course](#correcting-course)
- [Advanced Patterns](#advanced-patterns)
- [What to Avoid](#what-to-avoid)

---

## The Core Rhythm

Every productive session follows a four-beat pattern:

```
1. ORIENT   — "Hi, get started"
2. COMMIT   — "Let's work on [task]"
3. VERIFY   — "Verify everything"
4. SHIP     — "Commit and push. That's all for the session"
```

This rhythm works because it maps to how the agent is designed to operate. The orient phase triggers startup protocols. The commit phase gives the agent a clear objective. The verify phase activates self-review. The ship phase triggers session wrap-up.

You don't need to follow this rigidly, but sessions that skip phases — especially orient and verify — consistently produce worse results.

---

## Starting a Session

**Always open with a greeting before your task.**

```
You:    Hi, get started
Agent:  [reads identity, checks memory, loads protocols, reviews status...]
```

This isn't politeness — it's functional. The greeting triggers the agent's full startup sequence: identity loading, memory recall, protocol enforcement, status review, and routine checks. Jumping straight to a task risks the agent working without its full context.

**What works:**
- `"Hi, get started"`
- `"Hey, let's go"`
- Any short greeting followed by a pause

**What doesn't work:**
- Immediately pasting a task with no greeting
- Long preambles explaining what you want before the agent has oriented

Wait for the agent to report its status before giving your first task. The status report tells you what it remembers, what's pending, and what's overdue — context that might change what you ask for.

---

## How to Frame Requests

### Lead with What, Not How

State the outcome you want. Leave implementation to the agent.

| Instead of this | Say this |
|:---------------|:---------|
| _"Create a function called validateUser that takes a JWT token, decodes it, checks the expiry..."_ | _"I need JWT authentication for the API"_ |
| _"Open Status.md and move the auth item from Pending to In Progress"_ | _"Start working on the auth task"_ |
| _"Run go test ./... and check for failures in the config package"_ | _"Run the tests, focus on config"_ |

The agent has the full codebase, the project context, and domain knowledge. When you over-specify, you constrain it to your mental model — which may miss better approaches.

### Lead with Principles, Not Specifications

The most powerful thing you can say to an agent isn't a detailed spec — it's a guiding principle.

```
"Every file should be editable by the user, because that's the whole point"
```

A sentence like this reshapes an entire feature. The agent will derive dozens of implementation decisions from a single principle — and they'll be consistent with each other because they share a root.

**More examples:**

| Principle | What it produces |
|:----------|:----------------|
| _"Bonsai is a scaffolder, not a cage"_ | Files are generated but user-owned, never locked |
| _"The agent should be able to run without the user present"_ | Autonomous workflows, self-checks, robust error handling |
| _"Keep the CLI interactive — no silent defaults"_ | Huh forms for every input, confirmation prompts, previews |

One good principle outweighs a page of specifications.

### Use Compound Directives

Chain related steps in a single message to create a complete workflow pipeline:

```
"Plan the feature, verify the plan, implement it, verify the implementation"
```

```
"Research the options, present a comparison, recommend one"
```

```
"Fix the bug, add a test for it, verify nothing else broke"
```

This works because the agent treats each clause as a phase. It won't skip verification just because you also asked for implementation — the explicit sequence creates accountability.

### Brain-Dump When You Need To

You don't need to pre-structure your thoughts. Raw, messy input works:

```
"fix core files. should they be general, have separate directory in catalog?
and whats the tmpl extension... self update skills and workflows...
currently we remove complete agents, we should be able to remove individual
skills etc. also research https://github.com/example/tool"
```

The agent will parse this into structured action items. It's better at organizing your thoughts than you might be at pre-organizing them — and the time you'd spend formatting a clean request is time wasted.

---

## The Power of Terse Communication

Brevity isn't laziness — it's precision. Short, well-timed messages consistently outperform verbose instructions.

### Why Short Messages Work

1. **Less ambiguity.** `"Fix the login bug"` has one interpretation. A paragraph about the login bug has many.
2. **Faster feedback loops.** You can course-correct with another short message if the agent misinterprets.
3. **The agent fills in context.** It has the codebase, the project state, and its memory. You don't need to repeat what it already knows.

### Effective Short Patterns

| Pattern | Example | When to use |
|:--------|:--------|:-----------|
| **Selection** | `"B"` or `"option 2"` or `"the second one"` | When the agent presents options |
| **Approval** | `"yup"`, `"do it"`, `"go ahead"` | When the agent proposes a plan |
| **Direction** | `"push"`, `"commit"`, `"continue"` | When the agent pauses for confirmation |
| **Correction** | `"no, the other file"`, `"not that — the backlog"` | When the agent goes the wrong direction |
| **Scope** | `"do all"`, `"just the first three"` | When the agent asks how much to do |

### Silence Is Approval

When things are going well, **don't interrupt**. Approving tool calls without adding messages is the strongest signal that the agent is on track. The longest productive streaks happen during silence — 50+ consecutive approvals spanning complete feature implementations.

If you find yourself wanting to say "looks good, keep going" — don't. Just keep approving. The message adds nothing; the silence says everything.

---

## Quality Gates

### The "Verify Everything" Instruction

The single most effective quality practice:

```
"Verify everything and check if you missed something or made a mistake"
```

Say this **before committing or shipping work**. It consistently catches real bugs:

- Off-by-one errors and edge cases
- Stale references in documentation
- Missing test coverage
- Variable ordering bugs
- Non-deterministic iteration issues
- Broken cross-references

This isn't paranoia — it's structural. The agent's first-pass implementation focuses on making things work. The verification pass focuses on making things correct. These are different cognitive modes, and separating them produces better results than trying to do both at once.

**Variations that work equally well:**
- `"Verify everything"`
- `"Check if you missed something"`
- `"Review your work before we commit"`
- `"Does everything look right?"`

**Variations that work less well:**
- `"Are you sure?"` — signals doubt rather than process
- `"Double-check that"` — too vague about what to check
- `"Is that correct?"` — invites a yes/no rather than a re-examination

### The Verification Phase

Make verification a **phase**, not an afterthought. In the four-beat rhythm:

```
1. ORIENT   — get started
2. COMMIT   — implement the feature
3. VERIFY   — "verify everything"     ← this is a full phase
4. SHIP     — commit and push
```

Don't rush from implementation to shipping. The verification phase regularly takes 10-20% of the session and catches issues that would cost 10x more to fix after shipping.

---

## Trust and Delegation

### Delegate Fully

The most productive interactions treat the agent as a capable team member, not a tool to be micromanaged.

**Delegation that works:**

| What you say | What it signals |
|:-------------|:---------------|
| _"You can design this yourself"_ | Full creative authority within the domain |
| _"Which one would you like to work on?"_ | Agent's judgment is valued |
| _"If you're happy with everything, commit and push"_ | Quality judgment is delegated |
| _"Create a backlog system for yourself if you don't have one"_ | Agent can self-provision |

**The rule:** control at the philosophical level, delegate at the implementation level. You set the principles and the destination. The agent figures out the route.

### Don't Inspect Intermediate Work

Resist the urge to review every file change, every function signature, every variable name. The "verify everything" gate at the end is more effective than continuous oversight, and it gives the agent room to make coherent design decisions without interruption.

If you inspect intermediate work, you'll be tempted to micro-correct — and each micro-correction resets the agent's momentum.

### Increasing Autonomy Over Time

Trust is earned through results, not granted upfront. A natural progression:

1. **Early sessions:** Review plans before dispatching, verify frequently
2. **After a few sessions:** Skip plan review for familiar patterns, verify at the end
3. **Established rhythm:** State the goal, stay silent, verify, ship

The agent's memory and your shared working patterns compound over time. What required 10 messages in session 1 might take 2 messages by session 10.

---

## Correcting Course

### Correct Through Observation

When the agent does something wrong, state what you observe — not what you feel.

| Instead of this | Say this |
|:---------------|:---------|
| _"That's wrong, you should have..."_ | _"It's the backlog.md"_ |
| _"No, I don't like that approach"_ | _"No, it should be more real-world themed"_ |
| _"You forgot to..."_ | _"Don't see the tests"_ |

Corrections delivered as data produce faster recovery. The agent doesn't need to process criticism or reassurance — it needs the right information to adjust.

### Compound Corrections

You can correct and redirect in a single message:

```
"No, don't commit that. And yes, add the other two to gitignore"
```

This is efficient because it handles the error and provides the next instruction simultaneously. No need for a separate correction message followed by a separate instruction.

### When the Agent Doesn't Understand

If the agent misinterprets after one attempt, don't rephrase the same instruction with more words. Instead:

1. **State the answer directly** — `"It's in station/Playbook/Backlog.md"`
2. **Give a concrete example** — `"Like this: [example]"`
3. **Narrow the scope** — `"Just fix the function on line 42"`

Repeating a vague instruction louder doesn't help. Switching from vague to specific does.

---

## Advanced Patterns

### Hypothesis-Then-Validate

Assert your position, then invite pushback:

```
"I think we should use SQLite for this. Thoughts?"
```

This preserves your authority (you're proposing, not asking) while creating space for the agent to contribute expertise. The agent might confirm your hypothesis, refine it, or flag a problem you missed.

### Co-Creative Framing

Sometimes the agent's *description* of a problem is more valuable than its solution. Listen for framing:

```
Agent:  "...this is essentially a status bar for the agent's self-awareness..."
You:    "You said a critical word — status bar. Let's build that as a sensor."
```

The agent's analysis can become your creative input. When you hear the agent name something well, grab it.

### Session Type Awareness

Match your communication style to the kind of session you're running:

| Session type | Duration | Your role | Communication |
|:------------|:---------|:----------|:-------------|
| **Sprint** | 15-45 min | Direct, decisive | 4-8 messages, mostly approvals |
| **Marathon** | 1-5 hours | Guide, checkpoint | 10-15 messages, verify between phases |
| **Design** | 30-60 min | Collaborative, exploratory | Richer messages, more back-and-forth |
| **Micro** | 1-15 min | Quick directive | 1-3 messages, in-and-out |

Don't design-session when you're sprinting, and don't sprint when you're designing.

### The Philosophical Guard Rail

When you see the agent heading in a direction that technically works but philosophically conflicts with your vision, a single reframing sentence is your most powerful tool:

```
"Every file should be editable by the user, because that's the whole point"
```

```
"We're building a scaffolder, not a framework"
```

```
"The agent should never need the user present to finish its work"
```

These sentences propagate. The agent will apply the principle to decisions you haven't thought of yet.

---

## What to Avoid

| Anti-pattern | Why it fails | Do this instead |
|:------------|:-------------|:---------------|
| **Elaborate prompts** | Over-specification constrains the agent and adds noise | State the goal in 1-2 sentences |
| **Reviewing every diff** | Kills momentum, tempts micro-correction | Use "verify everything" at the end |
| **Brainstorming when you know the answer** | Wastes rounds on options you'll reject | State your choice directly |
| **Asking "are you sure?"** | Signals doubt, invites hedging | Ask "verify everything" instead |
| **Explaining why** | The agent doesn't need motivation, it needs direction | State what you want, skip the justification |
| **Positive reinforcement** | "Good job!" doesn't improve output | Forward motion is your approval |
| **Intervening during autonomous runs** | Breaks flow, resets momentum | Wait until the agent pauses or finishes |
| **Asking the same question twice** | If the agent didn't understand, repetition won't help | Give the answer directly |
| **Pre-structuring your input** | Time spent formatting is time wasted | Brain-dump and let the agent organize |
| **Using the agent for pair programming** | The dispatch model works better than back-and-forth | Delegate fully, review at the end |

---

## Quick Reference

### The Five Essential Phrases

| Phrase | When | What it does |
|:-------|:-----|:-------------|
| `"Hi, get started"` | Session open | Triggers full startup sequence |
| `"Verify everything"` | Before shipping | Activates self-review — catches real bugs |
| `"You can design this yourself"` | Complex features | Delegates creative authority |
| `"Thoughts?"` | After stating a position | Invites expertise without yielding authority |
| `"That's all for the session"` | Session close | Triggers memory update, logging, cleanup |

### Session Cheat Sheet

```
Start:     "Hi, get started"
           [wait for status report]

Task:      "I need [outcome]"              — features
           "Fix [problem]"                 — bugs
           "[principle]. Design it."       — architecture

Approve:   [silence]                       — keep approving tools
           "yup" / "do it" / "go ahead"   — explicit approval

Correct:   "No, [what's actually true]"    — factual correction
           "Not that — [alternative]"      — redirection

Verify:    "Verify everything"
           [wait for review to complete]

Ship:      "Commit and push"
           "That's all for the session"
```
