---
tags: [workflow, testing]
description: Design a structured test plan for a feature — scope, prioritize, allocate test types.
---

# Workflow: Test Plan

---

## When to Use

When a new feature, significant change, or bug fix needs a structured testing approach before implementation begins. Use this to produce a test plan document that execution agents follow.

---

## Steps

### 1. Analyze Feature

- Read the spec, plan, or ticket for the feature
- Identify: inputs, outputs, state changes, external dependencies
- List all components and modules affected
- Note any non-functional requirements (performance, security, accessibility)

### 2. Define Scope

- **In scope:** what this test plan covers (new code, modified paths, integration points)
- **Out of scope:** what is explicitly NOT tested here (unchanged modules, third-party services)
- **Entry criteria:** what must be true before testing starts (feature implemented, dependencies available)
- **Exit criteria:** what must be true to consider testing complete (all P0/P1 cases pass, coverage targets met)

### 3. Classify by Priority

Assign each test case a priority:

- **P0 — Critical path:** core happy-path scenarios that must work for the feature to ship. If these fail, the feature is broken.
- **P1 — Expected variations:** alternate valid inputs, expected error conditions, common edge cases. If these fail, the feature has significant gaps.
- **P2 — Boundary/edge:** extreme values, empty inputs, max limits, unusual combinations. If these fail, the feature has rough edges.
- **P3 — Failure recovery:** network failures, timeouts, corrupt data, concurrent access. If these fail, the feature lacks resilience.

### 4. Design Test Cases

For each test case, document:

- [ ] **ID** — unique identifier (e.g., `TC-AUTH-001`)
- [ ] **Priority** — P0/P1/P2/P3
- [ ] **Description** — what is being tested, in one sentence
- [ ] **Preconditions** — state required before the test runs
- [ ] **Steps** — concrete actions to perform
- [ ] **Expected result** — what success looks like (specific values, status codes, state changes)
- [ ] **Test type** — Unit / Integration / E2E

### 5. Allocate Test Types

Apply the testing pyramid:

- **Unit tests (70%):** all business logic, transformations, validations. Mock external dependencies. Each test <100ms.
- **Integration tests (20%):** API endpoints with real database, service-to-service interactions, auth flows. Reset state between tests.
- **E2E tests (10%):** P0 critical journeys only — max 3-5 per feature. Page Object Model, explicit waits (never `sleep`).

### 6. Regression Impact

- Identify existing tests that touch modified code
- Flag tests that need updating due to changed behavior
- List any existing tests that might break as a side effect
- Note new test gaps introduced by the change

### 7. Submit Plan

- Write the test plan as a structured document
- Include: scope, priority breakdown, full test case list, type allocation, regression notes
- The plan is ready for execution agents to implement
