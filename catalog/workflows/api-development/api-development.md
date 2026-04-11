---
tags: [workflow, development]
description: Spec-first API development — define contract, generate scaffolding, implement, test, document.
---

# Workflow: API Development

---

## When to Use

When implementing a new API endpoint or modifying an existing one. Always start from the contract (OpenAPI spec), not from the code.

---

## Steps

### 1. Define Contract

- Write or update the OpenAPI 3.1 spec FIRST — before any implementation
- Define for each endpoint: method, path, description, parameters (with types and constraints), request/response schemas, auth requirements, rate limits
- Include error response schemas (RFC 9457 Problem Details format)
- Validate the spec with a linter (Spectral, Redocly) if available
- Get alignment on the contract before writing code

### 2. Generate Scaffolding

- Generate types/structs from the OpenAPI schemas:
  - **TypeScript:** interfaces or Zod schemas from components
  - **Go:** structs from schema definitions
  - **Python:** Pydantic models from schemas
- Generate request validation middleware from parameter constraints
- Set up the route/handler skeleton matching the spec paths

### 3. Implement

- [ ] Service layer first — business logic, isolated from HTTP concerns
- [ ] Handler/controller layer — HTTP request/response, delegates to service
- [ ] Apply auth middleware to protected endpoints
- [ ] Error handling: catch service errors, map to RFC 9457 responses with correct status codes
- [ ] Pagination: cursor-based for collections, return `data[]` + `meta.has_more` + `meta.next_cursor`
- [ ] Idempotency: support `Idempotency-Key` header on POST endpoints that create resources
- [ ] Request ID: generate UUID, set `X-Request-Id` header, pass through to logs

### 4. Write Tests

- [ ] **Contract tests** — response matches OpenAPI spec (status codes, required fields, types)
- [ ] **Unit tests** — service layer logic in isolation (mock data layer)
- [ ] **Integration tests** — handler + service + real database (test container or in-memory)
- [ ] **Error path tests** — 400 (bad input), 401 (no auth), 403 (forbidden), 404 (not found), 422 (validation), 409 (conflict)
- [ ] **Edge cases** — empty collections, max pagination limits, concurrent requests with same idempotency key

### 5. Document

- Verify implementation matches the OpenAPI spec — no drift between spec and code
- Add real-world request/response examples to the spec
- Update CHANGELOG with the new endpoint
- If breaking an existing endpoint: bump API version, add `Sunset` header to old version, document migration

### 6. Self-Check

- [ ] Every response matches the OpenAPI spec exactly (fields, types, status codes)
- [ ] Rate limiting headers present: `RateLimit-Limit`, `RateLimit-Remaining`, `RateLimit-Reset`
- [ ] Auth tested: valid token, expired token, missing token, wrong scope
- [ ] No N+1 queries — check query count for list endpoints
- [ ] Request body size limits enforced
- [ ] CORS configured with explicit origin allowlist
