# Research: Catalog Expansion — Agents, Skills, Workflows, Sensors, Routines

> Date: 2026-04-11
> Status: Research complete, ready for implementation decisions

---

## Current Inventory

| Category | Count | Items |
|----------|-------|-------|
| Agents | 3 | tech-lead, backend, frontend |
| Skills | 5 | coding-standards, design-guide, planning-template, database-conventions, testing |
| Workflows | 5 | session-logging, planning, code-review, plan-execution, reporting |
| Protocols | 4 | session-start, scope-boundaries, memory, security |
| Sensors | 5 | routine-check, scope-guard-files, agent-review, scope-guard-commands, session-context |
| Routines | 4 | roadmap-accuracy, doc-freshness-check, status-hygiene, memory-consolidation |
| Scaffolding | 4 | index, playbook, logs, reports |

---

## Part 1: New Agent Types

### Tier 1 — High Priority (broad audience, distinct from existing agents)

#### 1. `fullstack` — Full-Stack Agent

**Why:** Most real-world projects don't split frontend/backend teams. Next.js, SvelteKit, Nuxt blur the line. Highest demand for solo developers and small teams. The frontend+backend split works for big teams; fullstack is the natural fit for everyone else.

**Description:** Implements full-stack features end-to-end — UI, API routes, database, auth, tests.

**Defaults:**
- Skills: `coding-standards`, `testing`, `design-guide`, `database-conventions`, `api-design-standards`, `auth-patterns`
- Workflows: `plan-execution`, `reporting`, `session-logging`
- Protocols: (all required — auto-installed)
- Sensors: `session-context`, `scope-guard-files`

**Scope owns:** `src/`, `app/`, `prisma/`, `drizzle/`, test files
**Scope excludes:** `infrastructure/`, `terraform/`, `.github/workflows/`, `Dockerfile`

---

#### 2. `devops` — DevOps / Infrastructure Agent

**Why:** Infrastructure-as-code is highly structured, pattern-heavy, and one of the strongest AI agent use cases. Every non-trivial project needs CI/CD, containers, and deployment config. Universal need.

**Description:** Manages infrastructure-as-code, CI/CD pipelines, containers, and deployment automation.

**Defaults:**
- Skills: `coding-standards`, `iac-conventions`, `container-standards`
- Workflows: `plan-execution`, `reporting`, `session-logging`, `security-audit`
- Protocols: (all required)
- Sensors: `session-context`, `scope-guard-files`, `iac-safety-guard`
- Routines: `dependency-audit`, `infra-drift-check`

**Scope owns:** `infrastructure/`, `terraform/`, `pulumi/`, `deploy/`, `k8s/`, `helm/`, `Dockerfile`, `docker-compose*.yml`, `.github/workflows/`, `Makefile`, `scripts/`
**Scope excludes:** `src/`, `app/`, `lib/`, `components/`, `pages/`, application test code

**Blocked commands:** `terraform destroy` (without `-target`), `terraform apply -auto-approve`, `kubectl delete namespace`, `kubectl delete deployment` (without `--dry-run`), `docker system prune -a`, `docker volume rm`

---

#### 3. `security` — Security Engineer Agent

**Why:** Cross-cutting, every project needs it. "Shift left" is the 2025 trend — validating code against security policies before CI/CD. A dedicated security agent understands OWASP, dependency scanning, and auth patterns.

**Description:** Audits code for vulnerabilities, reviews auth patterns, scans dependencies, enforces security standards.

**Defaults:**
- Skills: `coding-standards`, `testing`, `auth-patterns`, `review-checklist`
- Workflows: `code-review`, `reporting`, `session-logging`, `security-audit`
- Protocols: (all required)
- Sensors: `session-context`, `scope-guard-files`, `scope-guard-commands`, `api-security-check`
- Routines: `vulnerability-scan`

**Scope owns:** `security/`, auth/middleware modules, security config files
**Read access:** entire codebase (for audit)
**Scope excludes:** application business logic, UI components, infrastructure (writes only to security-owned files)

**Blocked commands:** Application servers (`npm start`, `npm run dev`, `flask run`). Allowed: security scanners (`npm audit`, `semgrep`, `gitleaks`, `trivy`).

---

#### 4. `qa` — QA / Testing Agent

**Why:** Testing is the most natural AI agent use case. The current `testing` skill is a passive reference; a QA agent actively writes and maintains tests. Adaptive test generation from implementation plans is a growing pattern.

**Description:** Writes tests, maintains test suites, analyzes coverage, produces test plans for new features.

**Defaults:**
- Skills: `coding-standards`, `testing`, `test-strategy`
- Workflows: `plan-execution`, `reporting`, `session-logging`, `test-plan`
- Protocols: (all required)
- Sensors: `session-context`, `scope-guard-files`, `test-integrity-guard`
- Routines: `test-coverage-check`

**Scope owns:** `tests/`, `test/`, `__tests__/`, `spec/`, `e2e/`, `cypress/`, `playwright/`, test config files, test fixtures/data
**Scope excludes:** application source code (reads it, never modifies)

---

#### 5. `reviewer` — Code Review Agent

**Why:** Microsoft's AI code review handles 90%+ of PRs (600K+/month), 85% satisfaction. A dedicated reviewer distinct from tech-lead (which plans and orchestrates) fills the "quality gate" role.

**Description:** Reviews code changes against standards, security, performance, and architectural compliance. Read-only — never writes application code.

**Defaults:**
- Skills: `coding-standards`, `testing`, `review-checklist`
- Workflows: `code-review`, `pr-review`, `session-logging`
- Protocols: (all required)
- Sensors: `session-context`, `scope-guard-files`, `scope-guard-commands`
- Routines: `standards-drift`

**Scope:** Read access to entire codebase. Write access ONLY to own workspace directory.
**Blocked commands:** ALL execution commands (same as tech-lead). Allowed: `git log`, `git diff`, `git show`, `wc`, `ls`.

---

#### 6. `docs` — Documentation Agent

**Why:** Documentation is increasingly first-class. ADR-writing agents are a key 2025 pattern. Docs drift is one of the most common complaints in codebases. An agent that actively maintains docs is high-value.

**Description:** Writes and maintains technical documentation — API docs, ADRs, changelogs, READMEs.

**Defaults:**
- Skills: `coding-standards`, `documentation-standards`
- Workflows: `session-logging`, `reporting`, `api-docs-generation`
- Protocols: (all required)
- Sensors: `session-context`, `scope-guard-files`, `scope-guard-commands`
- Routines: `changelog-maintenance`, `api-docs-drift`

**Scope owns:** `docs/`, `documentation/`, `ADRs/`, `README.md`, `CHANGELOG.md`, `CONTRIBUTING.md`, OpenAPI specs
**Scope excludes:** application logic, test code, infrastructure (reads them for context)

**Blocked commands:** Build/test/server commands. Allowed: doc generation tools (`typedoc`, `godoc`, `swagger-codegen`).

---

### Tier 2 — Future (distinct audience, build later)

| Agent | Why | Notes |
|-------|-----|-------|
| `data` | Data pipelines, ML, notebooks | Growing rapidly but niche audience |
| `mobile` | React Native, Flutter, native | Large audience, distinct patterns |
| `gamedev` | Godot, Unity, Unreal | Very niche but unique patterns AI handles poorly without specialization |
| `sre` | Observability, incident response | Overlaps heavily with devops |
| `embedded` | Firmware, microcontrollers, RTOS | Very niche, underserved by AI tooling |
| `architect` | System-level design across repos | Overlaps with tech-lead; "senior variant" |
| `release` | Versioning, changelogs, pipelines | Better as devops skills than standalone agent |

### Not Recommended

| Agent | Why not |
|-------|---------|
| `blockchain` | Narrow audience; better as backend skills |
| `project-manager` | Overlaps completely with tech-lead |
| `designer` | Agent can't view/produce visuals; better as frontend skills |
| `performance` | Only 3-20% success on perf bugs (PerfBench); better as a skill/workflow |

---

## Part 2: New Skills (9)

### `api-design-standards`
**Agents:** backend, fullstack, tech-lead

Core rules:
- Plural nouns for collections: `/users`, `/orders` — never `/user`, `/getUsers`
- Kebab-case for multi-word paths: `/order-items`
- Nest max 2 levels: `/users/{id}/orders` — never deeper
- No verbs in URLs
- HTTP methods: GET=read, POST=create (201+Location), PUT=full replace, PATCH=partial, DELETE=remove (204)
- Error format: RFC 9457 Problem Details (`application/problem+json`) — `type`, `title`, `status`, `detail`, `instance`
- Pagination: cursor-based for large/changing datasets, offset OK for small/static. Always return `data[]`, `meta.has_more`, `meta.next_cursor`. Default `limit=20`, max `limit=100`. Opaque Base64 cursors.
- Idempotency: POST/PATCH support `Idempotency-Key` header (UUID v4), server stores 24-48hrs, dupes return original response
- Versioning: URL path (`/v1/users`), never break existing versions, `Sunset` header for deprecation
- Status codes: 200, 201, 204, 400, 401, 403, 404, 409, 422, 429, 500, 503
- Rate limiting: `X-RateLimit-Limit`, `X-RateLimit-Remaining`, `X-RateLimit-Reset` headers
- CORS: explicit allowlist, never `*` in production
- Request IDs: UUID per request, `X-Request-Id` header, include in all logs

---

### `auth-patterns`
**Agents:** backend, fullstack, security

Core rules:
- Web apps: OAuth 2.0 + OIDC with PKCE (Authorization Code flow)
- SPAs/mobile: OAuth 2.0 with PKCE — NEVER Implicit flow (deprecated by IETF)
- Machine-to-machine: Client Credentials or mTLS
- Internal services: API keys with rotation policy
- JWT: RS256 or ES256 only — NEVER `none`, never HS256 with public keys
- Access tokens: 15 min max. Refresh tokens: server-side or HTTP-only Secure SameSite=Strict cookie, rotate on every use
- Always validate: `exp`, `iss`, `aud`, `iat` claims
- Never store JWTs in localStorage
- Passwords: bcrypt or argon2id — NEVER MD5/SHA-1/SHA-256. Min 12 chars. Check against Have I Been Pwned.
- Login rate limiting: 5 failures = 15 min lockout
- Sessions: regenerate ID after login, absolute timeout 24h, idle timeout 30min
- API keys: prefix with service identifier (`sk_live_`, `pk_test_`), hash server-side (SHA-256), support 2-key rotation window

---

### `iac-conventions`
**Agents:** devops

Core rules:
- Terraform resource identifiers: `snake_case` — never repeat type in name
- Cloud resource names: `kebab-case` pattern `{env}-{project}-{component}`
- Module naming: `terraform-{provider}-{purpose}`
- Module files: `main.tf`, `variables.tf`, `outputs.tf`, `versions.tf`
- State: always remote (S3+DynamoDB / GCS / Azure Blob), locking enabled, separate per environment
- Never store secrets in state — use `sensitive = true`
- File organization: `main.tf`, `variables.tf`, `outputs.tf`, `providers.tf`, `locals.tf`, `data.tf`, `versions.tf`
- Mandatory tags: `Environment`, `Project`, `Team`, `ManagedBy` ("terraform"), `CostCenter`
- `prevent_destroy = true` on stateful resources (databases, S3, encryption keys)
- Never `-auto-approve` in production
- Always `terraform plan` before `terraform apply`
- Pin provider versions: `version = "~> 5.0"`, never unversioned
- Run `tfsec`/`checkov` in pre-commit hooks

---

### `test-strategy`
**Agents:** qa, backend, frontend, fullstack

Core rules:
- Testing pyramid: 70% unit, 20% integration, 10% E2E
- Unit tests: isolated, no I/O/network/DB, mock everything external, <100ms per test, AAA pattern (Arrange/Act/Assert)
- Integration tests: real database (test container or in-memory), real HTTP, one DB per suite, reset between tests, test success AND error paths
- E2E tests: P0 critical paths only (login, core flow, payment, signup), Page Object Model, never `sleep()` — explicit waits, headless in CI, max 20-30 E2E tests
- Coverage targets: 80% overall minimum, 95%+ critical paths (auth/payments/mutations), 90% new code
- What NOT to test: third-party library internals, getter/setter boilerplate, framework-generated code, private methods directly
- Test naming: `test_{method}_{scenario}_{expected}` or `describe/it` equivalent

---

### `documentation-standards`
**Agents:** docs, tech-lead

Core rules:
- README mandatory sections: project name + one-liner, quick start (<=3 commands), prerequisites, installation, usage/config, architecture overview (if >5 source files), contributing, license
- API docs: OpenAPI 3.1 as single source of truth. Every endpoint: method, path, description, params with types+constraints, request/response schemas, auth requirements. Include real examples.
- ADR format (MADR): `# ADR-{NNN}: {Title}`, Status, Date, Context, Decision, Consequences. Write for: tech choices, architectural patterns, dependency additions, API breaking changes. Keep <5min read. Append-only — supersede, never edit.
- Changelog: Keep a Changelog format. Categories: Added, Changed, Deprecated, Removed, Fixed, Security. Newest first. Unreleased section at top. Link each version to git diff.
- Docstrings: all public functions/methods/classes. Google style (Python), JSDoc (JS/TS), GoDoc (Go). Include: description, params with types, return value, exceptions.

---

### `container-standards`
**Agents:** devops

Core rules:
- Always multi-stage builds: builder (compile) -> runtime (minimal)
- Base images: specific digest pins, not mutable tags (`node:20.11.1-alpine` not `node:latest`)
- Prefer distroless or Alpine for runtime
- Non-root user: `RUN addgroup -S app && adduser -S app -G app` then `USER app`
- `WORKDIR /app`, never operate in root
- Layer ordering: OS deps -> language deps -> copy source -> build
- One process per container
- `COPY` not `ADD`
- Set `HEALTHCHECK`
- `.dockerignore`: `.git`, `node_modules`, `__pycache__`, `.env*`, `*.log`, `coverage/`, `.terraform/`, `*.tfstate`
- K8s manifests: always set resource requests AND limits, `runAsNonRoot: true`, `readOnlyRootFilesystem: true`, `allowPrivilegeEscalation: false`, `capabilities: { drop: ["ALL"] }`, never `latest` tag, liveness AND readiness probes

---

### `review-checklist`
**Agents:** reviewer, tech-lead

Checklist areas:
- **Correctness:** matches plan/ticket, edge cases handled, no swallowed exceptions
- **Security (OWASP):** no hardcoded secrets, parameterized SQL, input validation at boundaries, auth checks on every protected endpoint, no internal leaks in errors, SSRF allowlist for user URLs
- **Performance:** no N+1 queries, large collections paginated/streamed, no sync blocking in async paths, appropriate caching
- **Maintainability:** functions <50 lines / files <300 lines (guidelines), no code duplication, clear naming, no dead code, no TODOs without ticket reference

---

### `cli-conventions`
**Agents:** backend, fullstack

Core rules:
- Structure: `<program> <command> <subcommand> [flags] [arguments]`
- Verbs for commands, nouns for resources
- Consistent flags: `-h`/`--help`, `-v`/`--verbose`, `-V`/`--version`, `-o`/`--output`, `-f`/`--force`
- Default output for humans, `--output json` for machines
- Stdout for primary output, stderr for progress/errors
- Exit codes: 0=success, 1=error, 2=usage error
- Destructive ops require `--force` or interactive confirm
- `--dry-run` for destructive commands
- Never silently overwrite files
- Progress for ops >2s
- Respect `NO_COLOR` env var
- Shell completions for bash/zsh/fish
- Config: `$XDG_CONFIG_HOME/<program>/config.yaml`

---

### `mobile-patterns`
**Agents:** mobile, fullstack

Core rules:
- Formal state management: BLoC (Flutter), Redux/Zustand (React Native)
- Separate presentation from business logic — no API calls in widgets/components
- Repository pattern for data access
- Offline-first: local DB is primary, network is optimization layer
- Conflict resolution: last-write-wins for simple data, CRDT or manual for collaborative
- Background sync (WorkManager/Background Fetch)
- Lazy load screens, windowed lists (FlatList/ListView), image caching
- Target: <3s cold start, <16ms frame time (60fps)
- Follow Material Design (Android) / HIG (iOS)
- Handle safe areas, notches, dynamic type/font scaling
- Request permissions just-in-time with context explanation

---

## Part 3: New Sensors (3)

### `api-security-check`
**Event:** PreToolUse | **Matcher:** Edit|Write | **Agents:** backend, fullstack, security

Detects in code changes:
1. SQL string concatenation (`"SELECT.*" +` or f-strings with SQL keywords)
2. `eval()`, `exec()`, `Function()` with dynamic input
3. Hardcoded secrets: `password = "`, `api_key = "`, `secret = "`, `token = "`
4. `cors({ origin: '*' })` or `Access-Control-Allow-Origin: *`
5. User input passed directly to filesystem ops (`fs.readFile(req.params.path)`)
6. SSRF: user-controlled URLs to `fetch()`/`axios()`/`http.get()` without validation
7. Mass assignment: ORM create/update with raw request body
8. Sensitive data in logs: `console.log(password)`, `logger.info(token)`
9. Disabled security headers: removal of helmet, HSTS, CSP

---

### `iac-safety-guard`
**Event:** PreToolUse | **Matcher:** Bash | **Agents:** devops

Blocks (exit 2):
- `terraform destroy` (without `-target`)
- `terraform apply -auto-approve`
- `terraform state rm`, `terraform force-unlock`
- `kubectl delete namespace`, `kubectl delete deployment` (without `--dry-run`), `kubectl delete pv`, `kubectl delete --all`
- `docker system prune -a`, `docker volume rm`
- `aws s3 rb --force`, `gcloud projects delete`, `az group delete`

---

### `test-integrity-guard`
**Event:** PreToolUse | **Matcher:** Edit|Write | **Agents:** qa, backend, frontend, fullstack

Blocks:
- Removing/commenting out assertions (`assert`, `expect`, `.toBe`, `.toEqual`)
- Adding `.skip`/`.only` without documented reason
- Empty test bodies (no assertions)
- Catch-all error swallowing in tests: `try { ... } catch(e) {}`

Warns:
- `sleep()`/`setTimeout` with hardcoded delays >100ms in tests
- Test files with no assertions at all
- `console.log`/`print()` in test code (debug artifacts)
- Commented-out test code (>3 consecutive lines)

---

## Part 4: New Workflows (4)

### `security-audit`
**Agents:** security, devops, tech-lead

Steps:
1. **Secrets scan** — `gitleaks`/`trufflehog` on repo, check `.env` not in `.gitignore`
2. **Dependency audit** — `npm audit`/`pip-audit`/`govulncheck`. Critical/high = immediate, medium = plan, low = track
3. **Static analysis (SAST)** — `semgrep` with OWASP rulesets. Injection, broken auth, data exposure, XSS
4. **Config review** — CORS config, security headers (HSTS/CSP/X-Frame), token expiry, rate limiting, TLS 1.2+
5. **Access control review** — Map endpoints to permissions, BOLA protections, mass assignment, RBAC completeness
6. **Infrastructure review** — `tfsec`/`checkov` on IaC, container security, network policies
7. **Report** — Findings with severity, location, remediation to `Reports/Pending/`

---

### `api-development`
**Agents:** backend, fullstack

Steps (spec-first):
1. **Define contract** — Write OpenAPI 3.1 spec FIRST. Endpoints, schemas, errors, auth, rate limits. Validate with Spectral/Redocly.
2. **Generate scaffolding** — Server stubs, TypeScript types / Go structs from schemas, request validation middleware.
3. **Implement endpoint** — Service layer first (business logic), then handler/controller (HTTP), apply auth middleware, RFC 9457 errors.
4. **Write tests** — Contract tests (response matches spec), unit (service logic), integration (with real DB), error paths (400/401/403/404/422).
5. **Document** — Auto-generate from OpenAPI, add real-world examples, update CHANGELOG.
6. **Self-check** — Response matches spec, rate limiting works, auth tested positive+negative, no N+1 queries.

---

### `test-plan`
**Agents:** qa, tech-lead

Steps:
1. **Analyze feature** — Read spec/plan/ticket. Identify inputs, outputs, state changes, dependencies.
2. **Define scope** — In scope, out of scope, entry criteria, exit criteria.
3. **Classify by priority** — P0 (core happy path), P1 (variations + expected errors), P2 (boundary/edge), P3 (failure recovery).
4. **Design test cases** — For each: ID, Priority, Description, Preconditions, Steps, Expected Result, Test Type.
5. **Allocate test types** — Unit: all business logic. Integration: API/DB/service interactions. E2E: only P0 journeys (max 3-5 per feature).
6. **Regression impact** — Identify existing tests touching modified code, flag tests needing updates.
7. **Submit plan** — Structured document ready for execution.

---

### `pr-review`
**Agents:** reviewer, tech-lead

Steps:
1. **Context loading** — Read PR description, linked ticket/plan, full diff.
2. **Scope check** — Does PR match ticket and ONLY ticket? Flag unrelated changes.
3. **Correctness pass** — Logic correct? Edge cases? Error handling?
4. **Security pass** — OWASP Top 10 against changed code, auth on new endpoints, input validation, no secrets/injection/XSS.
5. **Performance pass** — N+1? Unbounded loading? Unnecessary re-renders? Missing indexes?
6. **Maintainability pass** — Naming, complexity, duplication, test coverage.
7. **Standards pass** — Coding standards, API conventions, docs updated.
8. **Verdict** — Pass / Request Changes / Needs Discussion.

---

## Part 5: New Routines (7)

### `dependency-audit` — every 7 days
**Agents:** devops, security

Run `npm audit`/`pip-audit`/`govulncheck`/`cargo audit` for each package manager. Flag critical/high CVEs in direct deps. Flag unmaintained deps (no activity 12+ months). Report: vulnerable deps with severity, CVE ID, recommended fix version.

### `infra-drift-check` — every 7 days
**Agents:** devops

Run `terraform plan` (read-only) for each state file. Compare declared vs actual cloud state. Flag: resources added outside Terraform, manually modified, deleted. Check env vars and secrets resolving.

### `vulnerability-scan` — every 7 days
**Agents:** security

SAST scan (semgrep + OWASP rules) on changed files since last scan. Secrets scan (gitleaks). Dependency audit. Cross-reference against last scan: new, resolved, persistent vulnerabilities. Flag critical/high findings older than 14 days without resolution.

### `test-coverage-check` — every 5 days
**Agents:** qa

Run full test suite with coverage. Compare against targets (80% overall, 95% critical paths). Identify untested files, zero-coverage functions, recently-modified code without tests. Flag coverage regressions.

### `changelog-maintenance` — every 7 days
**Agents:** docs

Compare git log since last changelog entry against CHANGELOG.md. Flag merged PRs with no entry. Check Unreleased section. Report missing entries with suggested categories.

### `api-docs-drift` — every 7 days
**Agents:** docs

Compare OpenAPI spec against endpoint implementations. Flag: endpoints in code not in spec, and vice versa. Check schema match. Flag stale docstrings.

### `standards-drift` — every 14 days
**Agents:** reviewer

Scan recent commits for standards violations. Check linter/formatter configs. Flag recurring review feedback themes (same issue 3+ times = needs linter rule). Report compliance summary.

---

## Part 6: Multi-Agent Coordination

### Hub-and-Spoke Model (already implemented)
Tech-lead is orchestrator; execution agents are workers. Research confirms this is the dominant successful pattern.

### Handoff Protocol (tech-lead + fullstack + devops example)
1. User gives task to tech-lead
2. Tech-lead writes plan, identifies executor(s)
3. Tech-lead dispatches to fullstack (app code) or devops (infra)
4. Executor implements, writes report to `Reports/Pending/`
5. Tech-lead reviews report, gives verdict
6. Cross-domain tasks: tech-lead writes TWO plans with dependency order, reviews both

### Scope Boundary Principles
1. **File-system based** — "you own `infrastructure/`" is enforceable; "you handle deployment concerns" is not
2. **Exclusive ownership** — each file has exactly one owner
3. **Read-many, write-one** — any agent can READ, only owner can WRITE
4. **Natural seams** — don't split a Next.js project into frontend/backend agents; use fullstack
5. **Infrastructure always separate** — even solo projects keep infra separate from app code

---

## Part 7: Implementation Summary

### Total new items: 29

| Category | New | Total after |
|----------|-----|-------------|
| Agents | 6 | 9 |
| Skills | 9 | 14 |
| Workflows | 4 | 9 |
| Protocols | 0 | 4 |
| Sensors | 3 | 8 |
| Routines | 7 | 11 |

### Build order recommendation

**Phase 1 — Agents + their required new items:**
1. `fullstack` agent (needs: `api-design-standards`, `auth-patterns` skills)
2. `devops` agent (needs: `iac-conventions`, `container-standards` skills + `iac-safety-guard` sensor)
3. `qa` agent (needs: `test-strategy` skill + `test-integrity-guard` sensor + `test-plan` workflow)

**Phase 2 — Review-oriented agents:**
4. `reviewer` agent (needs: `review-checklist` skill + `pr-review` workflow)
5. `security` agent (needs: `api-security-check` sensor + `security-audit` workflow)
6. `docs` agent (needs: `documentation-standards` skill + `api-docs-generation` workflow)

**Phase 3 — Cross-cutting items:**
7. Remaining skills: `cli-conventions`, `mobile-patterns`
8. All new routines (they enhance existing and new agents)

### Sources
- Agentic Engineering Part 3: Role-Based Agent Personas (sagarmandal.com)
- 5 Types of AI Agent in Software Development (EPAM)
- Coding Agent Teams: The Next Frontier (devops.com)
- Enhancing Code Quality at Scale with AI-Powered Code Reviews (Microsoft)
- 2025 DevOps Stack: Terraform, Kubernetes, and AI-Driven CI/CD
- DevSecOps in 2025: Principles, Technologies & Best Practices
- AI Agent Frameworks for End-to-End Test Automation (mabl.com)
- Game Engine Showdown 2025: Unity vs Godot vs Unreal
- API Design Best Practices in 2025: REST, GraphQL, and gRPC
- Electron vs. Tauri (dolthub.com)
- Full-Stack JS Frameworks 2025: Next.js vs Nuxt.js vs SvelteKit
- Azure SRE Agent, The SRE Report 2025
- PerfBench: Can Agents Resolve Real-World Performance Bugs? (arxiv)
- clig.dev — CLI Guidelines
- RFC 9457 — Problem Details for HTTP APIs
- MADR — Markdown Any Decision Records
- Keep a Changelog (keepachangelog.com)
