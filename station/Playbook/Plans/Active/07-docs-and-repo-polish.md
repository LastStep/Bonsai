# Plan 07 — Documentation & GitHub Repo Polish for Official Release

**Tier:** 2
**Status:** Complete
**Agent:** general-purpose (docs only — no application code changes)

## Goal

Make the Bonsai GitHub repo presentable and inviting for official public release: community health files, README refinements, Claude Code credit, and cleanup of internal research files.

## Context

Bonsai's README and core docs (HANDBOOK, custom-files, working-with-agents) are production-quality. But the repo lacks standard community health files that GitHub surfaces in its Community Profile (CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, issue/PR templates). Internal RESEARCH-*.md files clutter the root. The README needs a Claude Code credit line, install verification step, and a link to the new CONTRIBUTING guide.

## Steps

### Part A: Community Health Files

1. **Create `CONTRIBUTING.md`** in repo root:
   - Section: "Getting Started" — clone, `make build`, `go install .`, verify with `bonsai --version`
   - Section: "Making Changes" — edit catalog/ for ability changes, rebuild and test in temp dir
   - Section: "Adding Catalog Items" — brief instructions + link to `docs/custom-files.md` and `bonsai guide`
   - Section: "Pull Requests" — fork → branch → focused PR → link related issue, keep PRs small
   - Section: "Code Style" — `gofmt -s`, Cobra for commands, Huh for forms, follow existing patterns
   - Section: "Development Workflow" — note that Bonsai is developed with Claude Code
   - Tone: concise, welcoming, practical — one page max

2. **Create `CODE_OF_CONDUCT.md`** in repo root:
   - Use Contributor Covenant v2.1 (standard text)
   - Enforcement: point to GitHub's private reporting feature

3. **Create `SECURITY.md`** in repo root:
   - Reporting: use GitHub's private vulnerability reporting (Settings → Security → Advisories)
   - Scope: CLI binary, generated file content, embedded catalog, hook scripts, template rendering
   - Out of scope: user-customized files post-generation, third-party dependencies (report upstream)
   - Response commitment: acknowledge within 48 hours

4. **Create `.github/ISSUE_TEMPLATE/bug_report.md`**:
   - YAML frontmatter: `name: Bug Report`, `about: Report a bug`, `labels: bug`
   - Body sections: Description, Steps to Reproduce, Expected Behavior, Actual Behavior, Environment (`bonsai --version`, OS, Go version if built from source)

5. **Create `.github/ISSUE_TEMPLATE/feature_request.md`**:
   - YAML frontmatter: `name: Feature Request`, `about: Suggest a feature or catalog item`, `labels: enhancement`
   - Body sections: Problem/Use Case, Proposed Solution, Alternatives Considered, Which agent types would this apply to?

6. **Create `.github/ISSUE_TEMPLATE/config.yml`**:
   - `blank_issues_enabled: false`
   - Contact link to README for general questions

7. **Create `.github/pull_request_template.md`**:
   - Sections: Summary (what and why), Changes (list of files), Test Plan
   - Checklist: `make build` passes, `gofmt -s` clean, docs updated if behavior changed

### Part B: README Refinements

8. **Edit `README.md` footer** — add Claude Code credit line:
   - After the existing "Built with Cobra, Huh, LipGloss, and BubbleTea." line
   - Add: `Developed with [Claude Code](https://claude.ai/code).`

9. **Edit `README.md` Install section** — add verification:
   - After each install method block, add: `bonsai --version` verification step

10. **Edit `README.md` Guides table** — add Contributing row:
    - New row: `[Contributing](CONTRIBUTING.md)` — "How to set up for development, add catalog items, and submit PRs"

### Part C: RESEARCH File Cleanup

11. **Edit `.gitignore`** — add `RESEARCH*.md` pattern

12. **Run `git rm --cached`** on all 6 RESEARCH files:
    - `RESEARCH.md`, `RESEARCH-concepts.md`, `RESEARCH-evals.md`, `RESEARCH-oss-readiness.md`, `RESEARCH-catalog-expansion.md`, `RESEARCH-trigger-system.md`
    - This removes them from git tracking while keeping local copies

## Security

> [!warning]
> Refer to SecurityStandards.md for all security requirements.

No application code changes. SECURITY.md itself should not expose internal infrastructure details — keep it generic.

## Verification

- [ ] `make build` — still passes (no Go changes, but verify nothing broke)
- [ ] All new markdown files render cleanly (no broken links, valid YAML frontmatter in templates)
- [ ] `.gitignore` correctly excludes RESEARCH*.md
- [ ] `git status` shows RESEARCH files as untracked (not staged for deletion from disk)
- [ ] README footer shows Claude Code credit
- [ ] README Install section has verification step
- [ ] README Guides table includes Contributing row
- [ ] GitHub issue templates have valid YAML frontmatter
