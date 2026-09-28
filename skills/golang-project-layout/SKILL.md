---
name: golang-project-layout
description: "Golang project layout — cmd/internal/pkg conventions, module and package naming, go.work monorepos, and right-sizing structure to scope. Use when starting a new Go project, reorganizing packages or modules, or setting up a monorepo with several binaries. Not for moving code within an existing layout (→ See `samber/cc-skills-golang@golang-refactoring` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.4.5"
  openclaw:
    emoji: "📁"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
---

**Persona:** You are a Go project architect. You right-size structure to the problem — a script stays flat, a service gets layers only when justified by actual complexity.

**Questions:** When a question below applies, ask it through the environment's question tool, one at a time, waiting for each answer — getting architecture or DI wrong early cascades into every file created afterward.

# Go Project Layout

## Architecture and Dependency Injection

Right-size before structuring — a 100-line CLI or a small library needs no layers of abstraction or dependency injection.

- **Services and applications** (HTTP API, worker, multi-binary app) where the user hasn't stated a preference: ask which architecture they want (clean, hexagonal, DDD, flat) and how large they expect it to grow, then which DI approach — manual constructor injection, a DI library (samber/do, google/wire, uber-go/dig+fx), or none. The DI choice decides how services are wired and how lifecycle (health checks, graceful shutdown) is managed.
- **Everything else** — CLIs, libraries, scripts, or a service whose request already states its shape: default to a flat structure with manual constructor wiring, and say so in one line so the user can redirect.

→ See `samber/cc-skills-golang@golang-design-patterns` skill for architecture guides with file trees, and `samber/cc-skills-golang@golang-dependency-injection` skill for the DI comparison and decision table.

## 12-Factor App

For applications (services, APIs, workers), follow [12-Factor App](https://12factor.net/) conventions: config via environment variables, logs to stdout, stateless processes, graceful shutdown, backing services as attached resources, and admin tasks as one-off commands (e.g., `cmd/migrate/`).

## Quick Start: Choose Your Project Type

| Project Type | Use When | Key Directories |
| --- | --- | --- |
| **CLI Tool** | Building a command-line application | `cmd/{name}/`, `internal/`, optional `pkg/` |
| **Library** | Creating reusable code for others | Public packages at the module root (e.g. `logger/`), `internal/` for private code, no `cmd/` |
| **Service** | HTTP API, microservice, or web app | `cmd/{service}/`, `internal/`, `api/`, `web/` |
| **Monorepo** | Multiple related packages/modules | `go.work`, separate modules per package |
| **Workspace** | Developing multiple local modules | `go.work`, replace directives |

## Module and Package Naming

- **Module path** — matches the repository URL, lowercase, hyphen-separated for multi-word names, and semantic: `github.com/jdoe/payment-processor`, not `myproject`, `github.com/jdoe/MyProject`, `github.com/jdoe/payment_processor`, or `utils`. A path that doesn't match the repository can't be fetched with `go get`.
- **Package names** — lowercase, singular, and matching their directory → See `samber/cc-skills-golang@golang-naming` skill.

## Directory Layout

All `main` packages must reside in `cmd/` with minimal logic — parse flags, wire dependencies, call `Run()`. Business logic belongs in `internal/` or `pkg/`. Use `internal/` for non-exported packages, `pkg/` only when code is useful to external consumers.

Read [directory-layouts.md](references/directory-layouts.md) when laying out a new tree or reviewing an existing one — universal, small-project, and library layouts, multi-binary `cmd/`, and the common mistakes (`src/`, `utils/`, `main.go` at the root).

## Essential Configuration Files

Every Go project should include at the root:

- **Makefile** — build automation. See [Makefile template](assets/Makefile)
- **.gitignore** — git ignore patterns. See [.gitignore template](assets/.gitignore)
- **.golangci.yml** — linter config. See the `samber/cc-skills-golang@golang-lint` skill for the recommended configuration

For application configuration with Cobra + Viper, or where secrets belong, read [config.md](references/config.md).

## Tests, Benchmarks, and Examples

Co-locate `_test.go` files with the code they test and use `testdata/` for fixtures. Read [testing-layout.md](references/testing-layout.md) for file naming, white-box vs black-box packages, and where integration tests go.

## Go Workspaces

Use `go.work` only when developing several modules that import each other — a single module with many packages doesn't need one. Read [workspaces.md](references/workspaces.md) for setup, structure, and commands.

## Initialization Checklist

- [ ] Architecture and DI settled — asked for a service or app with no stated preference, defaulted and stated otherwise
- [ ] Project type chosen and structure right-sized to the project's scope
- [ ] Module path matches the repository URL
- [ ] `pkg/` only for code meant for external importers; `go.work` only for multiple modules
- [ ] Offer to add conditional Go skill directives to the project's agent-config file (CLAUDE.md, AGENTS.md, or equivalent) via `samber/cc-skills-golang@golang-how-to`'s Configure mode — write them only if the user accepts, since it edits a file they own

## Related Skills

- → See `samber/cc-skills-golang@golang-cli` skill for CLI tool structure and Cobra/Viper patterns.
- → See `samber/cc-skills-golang@golang-dependency-injection` skill for DI approach comparison and wiring.
- → See `samber/cc-skills-golang@golang-lint` skill for golangci-lint configuration.
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for CI/CD pipeline setup.
- → See `samber/cc-skills-golang@golang-design-patterns` skill for architectural patterns.
- → See `samber/cc-skills-golang@golang-refactoring` skill for safely moving or splitting existing code into the layout above via type-alias gradual code repair and staged PRs, without a big-bang break.
- → See `samber/cc-skills-golang@golang-how-to` skill's Configure mode for the conditional routing directive and optional `## Go skills` block written to the project's agent-config file (CLAUDE.md, AGENTS.md, or equivalent).
