# Container Standards

> [!important]
> These rules apply to all Dockerfiles, container images, and Kubernetes manifests.

---

## Dockerfiles

- Always multi-stage builds: builder stage (compile/install) -> runtime stage (minimal image)
- Pin base image versions to specific tags: `node:20.11.1-alpine`, not `node:latest` or `node:20`
- Prefer distroless or Alpine for runtime stages — minimize attack surface
- Run as non-root: `RUN addgroup -S app && adduser -S app -G app` then `USER app`
- Set `WORKDIR /app` — never operate in the filesystem root
- Layer ordering for cache efficiency: OS deps -> language deps -> copy source -> build
- One process per container — no supervisor daemons
- Use `COPY`, not `ADD` (unless extracting archives)
- Set `HEALTHCHECK`: `HEALTHCHECK --interval=30s --timeout=3s --retries=3 CMD <health-check-command>`
- Use `ARG` for build-time values, `ENV` for runtime — never bake secrets into either

## .dockerignore

- Always include: `.git`, `node_modules`, `__pycache__`, `.env*`, `*.log`, `coverage/`, `.terraform/`, `*.tfstate`, `.venv`, `dist/`, `build/`
- Mirror `.gitignore` as a starting point, then add build artifacts

## Image Security

- Scan images with `trivy` or `grype` in CI — fail on critical/high CVEs
- Never run as root in production — `USER` directive is mandatory
- No secrets in image layers — use build secrets (`--mount=type=secret`) or runtime injection
- Pin package manager versions in `RUN` commands where possible
- Remove package manager caches in the same `RUN` layer: `rm -rf /var/cache/apk/*`

## Kubernetes Manifests

- Always set resource `requests` AND `limits` for CPU and memory
- Security context on every pod:
  - `runAsNonRoot: true`
  - `readOnlyRootFilesystem: true`
  - `allowPrivilegeEscalation: false`
  - `capabilities: { drop: ["ALL"] }`
- Never use `latest` tag — always pinned image versions
- Set both liveness AND readiness probes with appropriate thresholds
- Use `PodDisruptionBudget` for production workloads
- Namespace isolation — never deploy to `default` namespace
- Use `ConfigMap` for non-sensitive config, `Secret` for sensitive values (with external secret management)

## Compose Files

- Pin image versions — same rule as Kubernetes
- Use named volumes for persistent data, not bind mounts in production
- Set `restart: unless-stopped` for services that should survive host reboots
- Define `healthcheck` for every service
- Use `.env` file for variable substitution, never hardcode environment-specific values
