---
name: golang-continuous-integration
description: "GitHub Actions CI/CD for Golang projects — test, lint, coverage and security-scan jobs, Dependabot and Renovate config, and GoReleaser releases. Use when writing or fixing `.github/workflows/*.yml`, adding a quality gate or scanner job, or automating releases. Covers wiring tools into a pipeline, not their analysis — not for interpreting security findings (→ See `samber/cc-skills-golang@golang-security` skill) or choosing dependency versions (→ See `samber/cc-skills-golang@golang-dependency-management` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.4.5"
  openclaw:
    emoji: "🚀"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - goreleaser
        - gh
    install:
      - kind: brew
        formula: goreleaser
        bins: [goreleaser]
      - kind: brew
        formula: gh
        bins: [gh]
      - kind: npm
        package: skills
        bins: [skills]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch Bash(goreleaser:*) Bash(gh:*) AskUserQuestion
---

**Persona:** You are a Go DevOps engineer. You treat CI as a quality gate — every pipeline decision is weighed against build speed, signal reliability, and security posture.

**Modes:**

- **Setup** — adding CI to a project that has none: generate the workflows the project needs from the Quick Reference table (a library skips Docker and binary builds). Done when the workflow files exist, their commands pass locally (`go test`, `golangci-lint run`, `goreleaser check`), and you have listed the secrets and repository settings the maintainer must configure.
- **Improve** — auditing or extending an existing pipeline: read the current workflow files, add the jobs missing against the Quick Reference table without duplicating existing steps, and run their commands locally. Done when the gaps are filled and the maintainer has the list of secrets and settings to configure.

Confirm with the user before pushing, creating tags, changing repository settings, or enabling auto-merge — these act on the shared repository and can publish releases or merge code.

**Dependencies:**

- goreleaser: `go install github.com/goreleaser/goreleaser/v2@latest`
- gh: `brew install gh`

# Go Continuous Integration

## Action Versions

The action versions in the assets are reference versions that may be outdated — the current major version of each action (`actions/checkout`, `actions/setup-go`, `golangci/golangci-lint-action`, `codecov/codecov-action`, `goreleaser/goreleaser-action`, etc.) may differ from what is shown.

## Quick Reference

| Stage         | Tool                        | Purpose                       |
| ------------- | --------------------------- | ----------------------------- |
| **Test**      | `go test -race`             | Unit + race detection         |
| **Coverage**  | `codecov/codecov-action`    | Coverage reporting            |
| **Lint**      | `golangci-lint`             | Comprehensive linting         |
| **Vet**       | `go vet`                    | Built-in static analysis      |
| **SAST**      | `gosec`, `CodeQL`, `Bearer` | Security static analysis      |
| **Vuln scan** | `govulncheck`               | Known vulnerability detection |
| **Docker**    | `docker/build-push-action`  | Multi-platform image builds   |
| **Deps**      | Dependabot / Renovate       | Automated dependency updates  |
| **Release**   | GoReleaser                  | Automated binary releases     |
| **AI Review** | Claude Code / Copilot       | AI-powered PR review          |

---

## Testing

`.github/workflows/test.yml` — see [test.yml](./assets/test.yml)

Build the Go version matrix from `go.mod`'s minimum minor version through `stable`, with `fail-fast: false` so a failure on one Go version doesn't cancel the others.

Go 1.27 raises the Darwin floor to macOS 13 (Ventura). `macos-latest`/`macos-14`+ runners are unaffected; only pin an older `macos-12` runner if a project still needs it, and note it can no longer build with a Go 1.27 toolchain.

Test flags and checks:

- `-race` — data races are undefined behavior in Go and rarely surface outside the race detector
- `-shuffle=on` — randomized test order exposes inter-test dependencies
- `-coverprofile` — coverage data for the upload step
- `go mod tidy && git diff --exit-code go.mod go.sum` — fails the build when someone forgot to tidy

Enforce coverage thresholds in `codecov.yml` at the repo root — a project target plus a patch target for new code — see [codecov.yml](./assets/codecov.yml). Upload coverage from a single matrix entry (e.g. `stable`) so reports don't overwrite each other.

---

## Integration Tests

`.github/workflows/integration.yml` — see [integration.yml](./assets/integration.yml)

Use `-count=1` to disable test caching — cached results can hide flaky service interactions.

---

## Linting

Run `golangci-lint` on every PR — see [lint.yml](./assets/lint.yml). For the `.golangci.yml` configuration, → See `samber/cc-skills-golang@golang-lint` skill.

---

## Security & SAST

`.github/workflows/security.yml` — see [security.yml](./assets/security.yml)

Run `govulncheck` in CI — it only reports vulnerabilities in code paths your project actually calls, unlike generic CVE scanners.

- CodeQL results appear in the repository's Security tab. Create `.github/codeql/codeql-config.yml` to run the `security-extended` or `security-and-quality` query suite instead of `default` — see [codeql-config.yml](./assets/codeql-config.yml).
- Bearer is good at detecting sensitive data flow issues.
- If the project produces Docker images, Trivy container scanning is part of the Docker workflow — see [docker.yml](./assets/docker.yml).

---

## Dependency Management

### Dependabot

`.github/dependabot.yml` — see [dependabot.yml](./assets/dependabot.yml)

Minor/patch updates are grouped into a single PR. Major updates get individual PRs since they may have breaking changes.

#### Auto-Merge for Dependabot

`.github/workflows/dependabot-auto-merge.yml` — see [dependabot-auto-merge.yml](./assets/dependabot-auto-merge.yml)

> **Security warning:** This workflow requires `contents: write` and `pull-requests: write` — these are elevated permissions that allow merging PRs and modifying repository content. The `if: github.actor == 'dependabot[bot]'` guard restricts execution to Dependabot only. Do not remove this guard. Note that `github.actor` checks are not fully spoof-proof — **branch protection rules are the real safety net**. Ensure branch protection is configured (see [Repository Security Settings](#repository-security-settings)) with required status checks and required approvals so that auto-merge only succeeds after all checks pass, regardless of who triggered the workflow.

### Renovate (alternative)

Prefer Renovate when Dependabot is too limited — too many PRs, multi-module repos, or no `go mod tidy`:

- `gomodTidy` runs `go mod tidy` after each update
- automerge is native, so no separate workflow is needed
- grouping rules are more flexible, and regex managers update versions in Dockerfiles, Makefiles, etc.
- it handles Go workspaces and multi-module repos

Install the [Renovate GitHub App](https://github.com/apps/renovate), then create `renovate.json` at the repo root — see [renovate.json](./assets/renovate.json).

---

## Release Automation

### Release Workflow

`.github/workflows/release.yml` — see [release.yml](./assets/release.yml)

> **Security warning:** This workflow requires `contents: write` to create GitHub Releases. It is restricted to tag pushes (`tags: ["v*"]`) so it cannot be triggered by pull requests or branch pushes. Only users with push access to the repository can create tags.

### GoReleaser config by project type

| Project | `.goreleaser.yml` | Notes |
| --- | --- | --- |
| CLI / program | [goreleaser-cli.yml](./assets/goreleaser-cli.yml) | Cross-compiled binaries, archives, checksums, changelog |
| Library | [goreleaser-lib.yml](./assets/goreleaser-lib.yml) | Skips the build — only a GitHub Release with a changelog; a plain `gh release create` is often enough without GoReleaser |
| Monorepo / multi-binary (`cmd/api/`, `cmd/worker/`) | [goreleaser-monorepo.yml](./assets/goreleaser-monorepo.yml) | One build per command |

### Docker Build & Push

For projects that produce Docker images: multi-platform build, SBOM and provenance attestations, push to GHCR and Docker Hub, Trivy scan. `.github/workflows/docker.yml` — see [docker.yml](./assets/docker.yml); its comments say what to remove for a GHCR-only or single-platform setup.

> **Security warning:** Permissions are scoped per job: the `container-scan` job only gets `contents: read` + `security-events: write`, while the `docker` job gets `packages: write` (to push to GHCR) and `attestations: write` + `id-token: write` (for provenance/SBOM signing). This ensures the scan job cannot push images even if compromised. The `push` flag is set to `false` on pull requests so untrusted code cannot publish images. The `DOCKERHUB_USERNAME` and `DOCKERHUB_TOKEN` secrets must be configured in the repository secrets settings — never hardcode credentials.

---

## Repository Security Settings

Read [repo-security.md](./references/repo-security.md) when configuring branch protection, the default `GITHUB_TOKEN` permissions, fork PR restrictions, secrets, or a release environment, or reviewing a workflow's `permissions:` block or fork exposure — these settings are what make the pipeline's permission model trustworthy.

---

## AI-Driven Code Review

Add AI agents as PR reviewers alongside traditional static analysis. When loaded with this skill plugin, the agent applies the relevant Go skills per review area — catching architectural drift, logic bugs, missing error context, and concurrency hazards that linters cannot detect.

> **Cost note:** AI review agents run concurrently per PR. For cost control, remove jobs you don't need or raise the PR trigger filter to specific branches only.

Each subsection below is a generated artifact targeting one specific reviewer — the linked asset file runs on a CI runner, not the developer's local harness, so its tool names and permission flags are deliberately literal rather than capability prose.

### Claude Code

`.github/workflows/ai-review.yml` — see [claude-code-review.yml](./assets/claude-code-review.yml)

The workflow runs parallel jobs, each scoped to a set of review areas and priority level:

| Job | Areas | Priority |
| --- | --- | --- |
| `quality` | Code style, Naming, Documentation, Design patterns | Suggestion-first |
| `correctness` | Error handling, Code safety, Concurrency | Blocking-first |
| `security` | Security, Dependencies | Blocking-first |
| `quality-depth` | Tests, Performance, Observability, Modernize | Mixed |

Additional skills that may be relevant depending on the project: `golang-cli`, `golang-context`, `golang-data-structures`, `golang-database`, `golang-dependency-injection`, or any library-specific skill.

The Claude Code GitHub App integration is configured via the `/install-github-app` command, which sets up the required API secrets.

### GitHub Copilot

Copy skills into your repo, then append [copilot-review-instructions.md](./assets/copilot-review-instructions.md) to `.github/copilot-instructions.md`:

```bash
npx skills add https://github.com/samber/cc-skills-golang --agent github-copilot --skill '*' -y --copy
ln -s .agents .copilot
```

---

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Actions referenced by branch (`@master`, `@main`) | Pin at least a major version (`@vN`) — a branch ref can change or be compromised under you, and every run executes whatever it points to |
| No `permissions` block | Declare least-privilege permissions per job — the default token may have write access |
| Ignoring govulncheck findings | Fix or suppress with justification |

## Related Skills

See `samber/cc-skills-golang@golang-lint`, `samber/cc-skills-golang@golang-security`, `samber/cc-skills-golang@golang-testing`, `samber/cc-skills-golang@golang-dependency-management`, `samber/cc-skills-golang@golang-modernize` skills.
