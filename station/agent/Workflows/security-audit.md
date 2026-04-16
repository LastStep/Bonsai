---
tags: [workflow, security]
description: Security audit — secrets scan, dependency audit, SAST, config review, access control, infrastructure.
---

## Triggers

**Slash command:** `/security-audit`
**Activate when:**
- Running a security audit on the codebase or recent changes
- Checking for secrets, vulnerable dependencies, or unsafe patterns

**Examples:**
> **User:** "Check the project for security issues"
> **Action:** Load security-audit workflow, scan secrets, audit deps, review config

---

# Workflow: Security Audit

---

## When to Use

When performing a security review of the project — scheduled, pre-release, or after a significant change.

---

## Steps

### 1. Secrets Scan

- Run `gitleaks` or `trufflehog` against the repository
- Verify `.env` files are in `.gitignore`
- Check for hardcoded secrets in source code: API keys, passwords, tokens, connection strings
- Check CI/CD configs for exposed secrets (should use secret variables, not inline)

### 2. Dependency Audit

- Run the appropriate scanner for each package manager: `npm audit`, `pip-audit`, `govulncheck`, `cargo audit`
- Triage by severity:
  - **Critical/High** — immediate remediation required
  - **Medium** — plan remediation within current sprint
  - **Low** — track, remediate opportunistically
- Flag dependencies with no maintenance activity in 12+ months

### 3. Static Analysis (SAST)

- Run `semgrep` with OWASP rulesets against the codebase
- Focus areas: injection (SQL, command, SSRF), broken authentication, sensitive data exposure, XSS
- Filter false positives — only report findings with actionable remediation

### 4. Configuration Review

- CORS: verify explicit origin allowlist, no `*` on authenticated endpoints
- Security headers: HSTS, CSP, X-Frame-Options, X-Content-Type-Options
- Token expiry: access tokens <= 15 min, refresh token rotation enabled
- Rate limiting: configured on auth endpoints and public APIs
- TLS: version 1.2+ enforced, no weak cipher suites

### 5. Access Control Review

- Map all endpoints to their required permissions
- Verify BOLA protections: users can only access their own resources
- Check for mass assignment: ORM create/update must not accept raw request bodies
- Verify RBAC completeness: no endpoints missing authorization checks

### 6. Infrastructure Review

- Run `tfsec` or `checkov` on all IaC configuration
- Check container security: non-root user, read-only filesystem, dropped capabilities
- Verify network policies restrict inter-service communication to required paths
- Check secrets management: no secrets in environment variables or config files in plain text

### 7. Report

- Write findings to `Reports/Pending/` using the report template
- Each finding: severity, location (file + line), description, remediation steps
- Group by category (secrets, dependencies, SAST, config, access, infrastructure)
- Include a summary with counts by severity

> [!note]
> The paths above are relative to the project docs location. Check your workspace CLAUDE.md for exact paths.
