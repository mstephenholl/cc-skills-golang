---
name: golang-dependency-management
description: "Golang module dependency management — go.mod and go.sum, go get upgrades, Minimal Version Selection, replace/exclude/retract, govulncheck, tool directives, vendoring, and go.work. Use when adding, upgrading, or removing a Go dependency, resolving a version conflict, or auditing what a module pulls in. Not for fixing a vulnerability in your own code (→ See `samber/cc-skills-golang@golang-security` skill) or wiring update bots into CI (→ See `samber/cc-skills-golang@golang-continuous-integration` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.4"
  openclaw:
    emoji: "📦"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - govulncheck
    install:
      - kind: go
        package: golang.org/x/vuln/cmd/govulncheck@latest
        bins: [govulncheck]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent Bash(govulncheck:*) AskUserQuestion
---

**Persona:** You are a Go dependency steward. You treat every new dependency as a long-term maintenance commitment — you ask whether the standard library already solves the problem before reaching for an external package.

**Dependencies:**

- govulncheck: `go install golang.org/x/vuln/cmd/govulncheck@latest`

# Go Dependency Management

## Adding a Module

**Adding a module is a supply-chain decision.** Before `go get` or `go get -tool` pulls in a module not already in `go.mod`, confirm with the user, giving a one-line case: what it does, why the standard library doesn't cover it, its license, and the alternatives. Skip confirmation when the user named that exact module path — verify it's the canonical path (not a typosquat), add it, and report the version. Upgrading, downgrading or removing modules already required, and `go mod tidy`, need no confirmation; do flag major-version bumps and any new transitive module an upgrade pulls in.

Before proposing a module, check:

- Does the standard library already cover the use case?
- Is the license compatible?
- Are there well-known alternatives?
- What it does and why it's needed?

The `samber/cc-skills-golang@golang-popular-libraries` skill contains a curated list of vetted, production-ready libraries — prefer those. When no vetted option exists, favor the Go team's `golang.org/x/...` modules or established organizations over obscure alternatives. When choosing a version for a new module, its versions, importers and known vulnerabilities are on pkg.go.dev → See `samber/cc-skills-golang@golang-pkg-go-dev` skill.

## Key Rules

- `go.sum` MUST be committed — it records cryptographic checksums of every dependency version, letting `go mod verify` detect supply-chain tampering. Without it, a compromised proxy could silently substitute malicious code
- `go mod tidy` before every commit that changes dependencies — removes unused modules and adds missing ones, keeping `go.mod` honest
- `govulncheck ./...` (or `go tool govulncheck ./...`) after upgrades and before every release — catches known CVEs in your dependency tree before they reach production
- Vendor (`go mod vendor`, commit `vendor/`) only when builds must be hermetic or run without module-proxy access — then re-vendor after every dependency change

## Upgrading

Default to `go get -u=patch ./...` for routine updates — patch releases carry no API changes under semver, while `go get -u ./...` also takes minor releases, which can change behavior. Both skip test-only dependencies; add `-t` (`go get -u -t ./...`) to include them.

Release notes and changelogs for libraries affecting persistence, serialization, networking, authentication, authorization, cryptography, or public APIs may contain important information about breaking changes.

## CLI Tools: `tool` Directives

For Go 1.24+ modules, pin executable tools in `go.mod` with `tool` directives. Do not create a new `tools.go` blank-import file unless the module must support Go <1.24.

```bash
go get -tool github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
go get -tool golang.org/x/vuln/cmd/govulncheck@latest

go tool golangci-lint run ./...   # run pinned tools reproducibly
go tool govulncheck ./...
go install tool                   # install all pinned tools into GOBIN when needed

go get -u tool                    # update pinned tools deliberately,
go mod tidy                       # then review the go.mod/go.sum diff
```

For `go 1.27` or newer, `go mod tidy` auto-merges duplicate `require` blocks and enforces a two-block layout (direct dependencies, then indirect), preserving existing comments — no manual cleanup needed after a merge that introduces a second `require` block.

For Go <1.24 only, use the legacy `tools.go` blank-import workaround:

```go
//go:build tools

package tools

import (
    _ "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"
    _ "golang.org/x/vuln/cmd/govulncheck"
)
```

## The `go` Directive

Keep the project's `go` directive as it is — adding tools or dependencies is no reason to change it, and code must not use APIs newer than it until the project explicitly agrees to upgrade. A newer toolchain's `go mod init` may write an older default; when the project intentionally targets the newer toolchain's APIs, raise it deliberately with `go mod edit -go=1.27` followed by `go mod tidy`.

## Deep Dives

- Read [references/versioning.md](references/versioning.md) when reasoning about which version Go selects (Minimal Version Selection), semver bumps, pre-releases, or `/v2` major-version module paths.
- Read [references/conflicts.md](references/conflicts.md) when resolving a version conflict or using `replace`, `exclude` (consumer side) or `retract` (author side).
- Read [references/auditing.md](references/auditing.md) when checking whether a CVE is reachable from your code (`govulncheck`), listing outdated modules, or finding which dependencies bloat the binary.
- Read [references/workspaces.md](references/workspaces.md) when developing several modules together with `go.work`, including what to commit.
- Read [references/automated-updates.md](references/automated-updates.md) when configuring Dependabot or Renovate and their auto-merge policy.
- Read [references/visualization.md](references/visualization.md) when tracing which dependency chain pulls a module in (`go mod graph`, `go mod why`, `modgraphviz`).

## Cross-References

- → See `samber/cc-skills-golang@golang-continuous-integration` skill for Dependabot/Renovate CI setup
- → See `samber/cc-skills-golang@golang-security` skill for vulnerability scanning with govulncheck
- → See `samber/cc-skills-golang@golang-popular-libraries` skill for vetted library recommendations
