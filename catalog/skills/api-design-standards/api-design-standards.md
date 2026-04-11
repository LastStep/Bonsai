# API Design Standards

> [!important]
> These rules apply to every HTTP API endpoint. Framework-agnostic — enforce regardless of language or stack.

---

## URLs

- Use plural nouns for collections: `/users`, `/orders` — never `/user` or `/getUsers`
- Use kebab-case for multi-word segments: `/line-items`, not `/lineItems`
- Nest at most two levels: `/users/{id}/orders` — promote deeper resources to top-level with a filter
- Path parameters for identity, query parameters for everything else
- No verbs in URLs — the HTTP method is the verb
- Lowercase only — no underscores, no uppercase in paths

## HTTP Methods

- **GET** — read, idempotent, cacheable, never mutates state
- **POST** — create or trigger action; return `201 Created` + `Location` header for new resources
- **PUT** — full replace, idempotent; client sends the complete resource
- **PATCH** — partial update; send only changed fields
- **DELETE** — remove, idempotent; return `204`; deleting an already-deleted resource returns `204`, not `404`
- POST is the only non-idempotent standard method

## Responses

- Top-level response is always a JSON object, never a bare array — `{ "data": [...] }` for collections
- Pick one field naming convention (snake_case or camelCase) and enforce it everywhere — never mix
- Use ISO 8601 for all dates: `"2024-11-15T09:30:00Z"`
- Use strings for all identifiers, even if numeric
- Omit null fields unless null carries distinct semantic meaning
- Return the full resource after creation and update

## Errors

- Use RFC 9457 Problem Details format with `Content-Type: application/problem+json`
- Required fields: `type` (URI), `title` (short summary), `status` (HTTP code), `detail` (specific explanation)
- Add `validation_errors: [{ "field": "email", "message": "must be valid" }]` for `422` responses
- Never expose stack traces, file paths, SQL, or infrastructure details
- Status code mapping: `400` bad syntax, `401` not authenticated, `403` not authorized, `404` not found, `409` conflict, `422` invalid input, `429` rate limited

## Pagination

- Cursor-based by default — return opaque `next_cursor` token; avoid offset for large or changing datasets
- Accept `limit` and `after`/`before` as query parameters; default `limit=20`, max `limit=100`
- Return `has_more` (boolean) and `next_cursor` in response metadata
- Do not return `total_count` by default — offer as opt-in (`?include_total=true`) if needed

## Idempotency

- Require `Idempotency-Key` header on POST endpoints that create resources or trigger side effects
- Store key + response for 24–48 hours; replay stored response on duplicate key
- Return `409 Conflict` if a key is reused with a different request body

## Versioning

- Version in URL path: `/v1/users` — increment only for breaking changes
- Breaking: removing/renaming a field, changing a type, adding a required request field
- Not breaking: adding optional response fields, new endpoints, new optional parameters
- Signal deprecation with `Sunset` header and document the migration path

## API Evolution

- Be conservative in what you send, liberal in what you accept (Postel's Law)
- Clients must ignore unknown fields in responses — servers should ignore unknown fields in requests (or reject with `400`, but pick one and document it)
- Never remove or rename a field in an existing version — add the new field alongside, deprecate the old
- Treat error codes and enum values as part of the contract — add new ones, never change existing ones

## Security Basics

- HTTPS only — reject plain HTTP or redirect with `301`
- Return `X-Request-Id` (server-generated UUID) on every response; log it for traceability
- CORS: explicit origin allowlist, never `Access-Control-Allow-Origin: *` on authenticated endpoints
- Enforce request body size limits; return `413` if exceeded
- Validate `Content-Type` header on requests with bodies; return `415` if wrong

## Rate Limiting

- Return `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset` headers on every response
- Return `429` with `Retry-After` header when limit is exceeded
- Apply limits per-client (by token/key), not per-IP alone
