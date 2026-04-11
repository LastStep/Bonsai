# Auth Patterns

> [!important]
> These rules are non-negotiable. All authentication and authorization code must conform. Framework-agnostic.

---

## Flow Selection

- User-facing apps (web, SPA, mobile): **Authorization Code + PKCE** — always require PKCE, even for confidential clients
- Machine-to-machine (services, CI/CD, cron): **Client Credentials**
- Developer-facing APIs where simplicity matters: **API keys**
- Server-rendered web apps: **session-based auth** with server-side session store
- Never implement the Implicit flow — formally deprecated by OAuth 2.1
- Never implement Resource Owner Password Credentials — it exposes passwords to the client
- Enforce exact redirect URI matching — never allow wildcards or prefix matching

## Token Management

- Sign JWTs with **ES256** (preferred) or **RS256** — never accept `none`, never use HS256 with public keys
- Always whitelist accepted algorithms in verification — never let the token's `alg` header dictate
- Access tokens: **15 minutes max** for standard apps, **5 minutes** for high-security
- Refresh tokens: **7–14 days** standard, **24 hours** for SPAs, rotate on every use
- Implement refresh token reuse detection — if a rotated token is replayed, revoke the entire token family
- Store access tokens in memory, refresh tokens in HTTP-only Secure SameSite=Strict cookies
- Never store tokens in localStorage — XSS exposes them
- Validate claims in order: `iss` → `aud` → `exp`/`nbf` → `sub` → signature
- Clock skew tolerance: no more than 30 seconds

## Password Handling

- Hash with **Argon2id** (preferred) or **bcrypt** (work factor 12+) — never MD5, SHA-1, or plain SHA-256
- Minimum **15 characters** for single-factor passwords (NIST SP 800-63B-4); allow up to at least 64
- Do not require character composition rules (uppercase, special chars) — NIST advises against them
- Do not require periodic rotation — force change only on evidence of compromise
- Screen against a blocklist: common passwords, dictionary words, context-specific terms, known breaches
- Check passwords against HaveIBeenPwned using the k-anonymity range endpoint — never send the full hash
- Return generic errors ("invalid credentials") — never reveal whether username or password was wrong

## Session Management

- Generate session IDs with a CSPRNG — at least 128 bits of entropy
- Regenerate session ID on every privilege change (login, role escalation)
- Idle timeout: 15–30 minutes standard, 2–5 minutes high-security
- Absolute timeout: 4–8 hours maximum regardless of activity
- Enforce timeouts server-side — never rely on client timers alone
- On logout, invalidate the session server-side — never rely solely on clearing the cookie

## API Keys

- Prefix with service + environment identifier: `myapp_live_`, `myapp_test_`
- Generate with CSPRNG, at least 256 bits of entropy
- Show the full key exactly once at creation; display only a truncated form afterward
- Hash keys with SHA-256 before storing — never store plaintext
- Support two active keys during rotation (24–48 hour overlap window); rotate every 90 days standard, 30 days high-security
- Transmit in `Authorization` header or custom header — never in URL query parameters
- Scope keys to minimum necessary permissions; apply per-key rate limits

## Authorization

- Default to RBAC; add ReBAC when permissions depend on object relationships; add ABAC for runtime context
- Deny by default — if no rule explicitly grants access, deny
- Enforce at the middleware/guard layer — never scatter checks inside business logic
- Separate authentication ("who?") from authorization ("allowed?") into distinct stages
- Check object-level authorization on every data access using a user-supplied identifier (BOLA prevention)
- Never rely on ID obscurity (UUIDs) as a security control — always verify access rights
- Write tests that attempt cross-user access (user A accessing user B's resources)

## Common Vulnerabilities

- Use constant-time comparison for all secret comparisons (tokens, keys, hashes)
- Whitelist fields on every endpoint — never bind raw request bodies to internal models (mass assignment)
- Use separate DTOs/schemas for input validation — never reuse domain models for request binding
- Rate-limit all auth endpoints aggressively: login, registration, password reset, token refresh
