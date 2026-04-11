# Test Strategy

> [!important]
> This guides what tests to write and how to structure them. The `testing` skill covers how to run them. Framework-agnostic.

---

## Test Distribution

- No single model is universal — adapt proportions to where your system's complexity lives
- Monoliths / general-purpose: classic pyramid — many unit, fewer integration, minimal E2E
- Frontend-heavy: testing trophy — most effort in integration tests exercising components together
- Microservices: honeycomb — the service is the unit; integration tests at API/DB boundaries form the largest layer
- Data pipelines: integration-heavy — verify transformations across real boundaries
- Shared libraries: exhaustive unit tests with high coverage — breakage cascades to all consumers

## Test Type Selection

- Pure logic or computation with no dependencies → **unit test**
- Code that crosses a system boundary (database, HTTP, filesystem, queue) → **integration test** with real or containerized dependency
- Critical user journey that drives revenue or prevents data loss → **E2E test** (5–10 flows max)
- Complex input domains or mathematical invariants → **property-based test** to supplement example-based tests
- Contract between two services you own → **contract test** (define consumer expectations, verify on provider)
- Simple delegation (A calls B, returns result unchanged) → **do not test**
- Getters, setters, trivial accessors → **do not test directly**
- Third-party library internals → **do not test**
- Framework-generated boilerplate → **do not test** unless custom logic was added

## Unit Tests

- Mock only at true system boundaries: network, database, filesystem, clock, randomness
- Do not mock types you don't own — wrap in an adapter, mock the adapter
- If setup requires more than two mocks, the design likely needs refactoring
- Prefer state-based assertions ("result equals X") over interaction-based ("method Y was called")
- Follow Arrange / Act / Assert — one blank line separating each phase
- One behavior per test — if you need "and" in the name, split it
- Name tests as behavior descriptions: state/input + expected outcome ("rejects withdrawal when balance is zero", not "testWithdraw3")
- Multiple asserts are fine only when they verify the same behavior
- Never use random data without a fixed seed; never depend on wall-clock time without mocking

## Integration Tests

- Test one real boundary per test — database OR HTTP OR queue, not all three
- Prefer test containers running the actual database engine over in-memory substitutes
- Run each test against clean state — transaction rollback, truncation, or per-test reset
- Set up data through the same public API/repository that production uses, not raw SQL
- Test actual HTTP round-trips (start server, make request, assert response) — not handler function calls
- Tests must not depend on execution order — each sets up its own preconditions

## E2E Tests

- Test only critical user journeys — login, core business flow, payment, signup
- Use stable selectors (data-testid, accessibility roles) — never CSS classes or DOM structure
- Use explicit waits for async operations — never fixed sleep durations
- Use Page Object / Screen Object pattern — one object per page, update one place when UI changes
- Set up data via API calls, not through the UI
- Isolate each test: create its own data, clean up after — never depend on other tests' data
- Keep total E2E suite under 10 minutes wall-clock; if longer, you have too many

## Coverage

- Use coverage as a diagnostic to find untested areas, not as a quality gate
- 70–80% line coverage is a reasonable floor; above 90% likely means testing trivial code
- Critical paths (auth, payments, data mutations): aim for 90%+
- Coverage without assertions is worthless — a test that executes code but asserts nothing catches zero bugs
- Never block a PR solely because coverage decreased — review whether the uncovered code needs tests
- Mutation testing is a stronger quality signal than coverage — run it periodically on critical modules

## Anti-Patterns

- **Implementation coupling**: if a behavior-preserving refactor breaks your test, the test is wrong — assert on outputs, not internals
- **Tautological tests**: if the test re-computes the expected value using the same formula as production, it can never fail independently
- **Flaky tests**: worse than no test — fix or quarantine immediately; never tolerate intermittent failures
- **Over-mocking**: more mock setup than assertions means you're testing wiring, not behavior
- **Ice cream cone**: too many E2E, too few unit tests — push coverage to the lowest level that catches each bug
- **The Liar**: test with no meaningful assertions or catch-all error swallowing — every test needs at least one assertion that can fail
- **Ignoring production bugs**: every bug found in production becomes a regression test before the fix is merged
