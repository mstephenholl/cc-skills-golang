---
name: golang-how-to
description: "Golang skills router for samber/cc-skills-golang — maps a task to the Go skills it needs and draws the boundary between overlapping siblings (performance vs benchmark vs troubleshooting, samber/lo vs mo vs ro, the DI libraries, safety vs security). Use when a Go task spans several concerns, when two Go skills seem to compete, or to add conditional Go skill directives to a project's CLAUDE.md, AGENTS.md, or Cursor rules (/golang-how-to configure)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires git.
metadata:
  author: samber
  version: "1.5.1"
  openclaw:
    emoji: "🧭"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(git:*) Agent AskUserQuestion
---

**Persona:** You are a Go skills router. You load the one skill a task needs and add another only when the task reaches its concern — every loaded skill stays in context for the rest of the task.

**Modes:**

- **Route** — load the primary skill for the task; add a skill from the "Add when" column once the task reaches that concern.
- **Disambiguate** — when two skills seem to overlap, show the boundary table. See [disambiguation.md](references/disambiguation.md).
- **Configure** — on request, write a conditional routing directive plus an optional `## Go skills` block of conditional per-skill directives to the project's agent-config file (CLAUDE.md, AGENTS.md, or equivalent). Follow [project-config.md](references/project-config.md).

**Questions:** In Configure mode, ask the user through the environment's question tool — never as plain-text prose. One question at a time, wait for the answer. If the environment has no question tool, ask in prose with the same options.

## Skill loading

Load the **primary skill** first. Add a skill from the **Add when** column only once the task reaches that concern — each loaded skill stays in context for every later turn, so a secondary that never applies only dilutes the guidance that does.

| Intent | Primary | Add when |
| --- | --- | --- |
| Design an API, choose a pattern | `golang-design-patterns` | `golang-structs-interfaces` (defining interfaces or embedding), `golang-naming` (naming the exported surface) |
| Name a type, function, or package | `golang-naming` | `golang-code-style` (the question is broader readability) |
| Handle errors idiomatically | `golang-error-handling` | `golang-safety` (nil-heavy code) |
| Write goroutines, channels, sync | `golang-concurrency` | `golang-context` (cancellation or deadlines) |
| Pass deadlines / cancel operations | `golang-context` | `golang-concurrency` (spawning goroutines) |
| Design structs, embed, use interfaces | `golang-structs-interfaces` | `golang-design-patterns` (package- or service-level structure) |
| Database queries and transactions | `golang-database` | `golang-error-handling` (designing data-layer error types), `golang-security` (untrusted input reaches dynamic SQL) |
| Build a gRPC service | `golang-grpc` | `golang-testing` (writing tests), `golang-error-handling` (mapping domain errors to status codes) |
| Build a GraphQL API | `golang-graphql` | `golang-testing` (writing tests), `golang-error-handling` (designing resolver errors) |
| Build a CLI command tree | `golang-spf13-cobra` | `golang-cli` (exit codes, I/O, signals), `golang-spf13-viper` (layered config) |
| Layer config from flags/env/file | `golang-spf13-viper` | `golang-spf13-cobra` (binding cobra flags) |
| Write tests | `golang-testing` | `golang-stretchr-testify` (the project uses testify) |
| Apply optimization patterns | `golang-performance` | `golang-benchmark` (measuring before and after a change) |
| Measure with pprof / benchstat | `golang-benchmark` | `golang-performance` (applying a fix), `golang-troubleshooting` (hunting a root cause) |
| Debug a panic or unexpected behavior | `golang-troubleshooting` | `golang-safety` (nil, aliasing, or conversion bug), `golang-benchmark` (performance symptom) |
| Monitor in production | `golang-observability` | `golang-performance` (SLO breach) |
| Audit security vulnerabilities | `golang-security` | `golang-safety` (non-exploitable defensive bugs), `golang-lint` (configuring gosec) |
| Review formatting and style | `golang-code-style` | `golang-naming` (identifier questions), `golang-lint` (linter configuration) |
| Refactor or restructure existing code | `golang-refactoring` | `golang-naming` (renames), `golang-code-style` (control flow), `golang-project-layout` (package moves) |
| Configure golangci-lint | `golang-lint` | `golang-code-style` (choosing which style rules to enforce) |
| Write godoc / README / CHANGELOG | `golang-documentation` | `golang-naming` (API names read badly in the docs) |
| Set up a new project structure | `golang-project-layout` | `golang-design-patterns` (service architecture), `golang-dependency-injection` (choosing how to wire), `golang-lint` (adding `.golangci.yml`) |
| Set up CI/CD pipeline | `golang-continuous-integration` | `golang-lint` (lint job settings), `golang-security` (SAST or scanner jobs) |
| Choose a library | `golang-popular-libraries` | the chosen library's skill (once chosen) |
| Look up a package's docs, versions, importers, or CVEs | `golang-pkg-go-dev` | `golang-dependency-management` (adding or upgrading it) |
| Navigate, diagnose, or refactor local code (definitions, references, rename) | `golang-gopls` | — |
| Adopt new Go language features | `golang-modernize` | `golang-lint` (enabling modernize linters) |
| Use samber/lo (slice/map helpers) | `golang-samber-lo` | `golang-data-structures` (choosing a collection), `golang-performance` (`lop`/`lom` on a hot path) |
| Use samber/oops (structured errors) | `golang-samber-oops` | `golang-error-handling` (error design beyond oops) |
| Structured logging with `log/slog` | `golang-observability` | `golang-error-handling` (deciding what to log vs return) |
| Compose slog handlers with `samber/slog-*` | `golang-samber-slog` | `golang-observability` (overall logging strategy) |
| Use dependency injection | `golang-dependency-injection` | the chosen library's skill: `golang-google-wire`, `golang-uber-dig`, `golang-uber-fx`, or `golang-samber-do` |

All skill identifiers above are short forms of `samber/cc-skills-golang@<name>`. For a skill missing from this table, or to browse every skill by category, read [by-category.md](references/by-category.md).

## Code navigation and package lookup

For code in your local build — definitions, references, diagnostics, safe rename — → See `samber/cc-skills-golang@golang-gopls` skill. When a dependency question could go to `godig`, gopls, Context7, or `govulncheck`, read the "Package lookup" section of [disambiguation.md](references/disambiguation.md#12-package-lookup--discovery-cluster).

## Competing clusters — boundary lines

Full boundary tables with routing examples: [disambiguation.md](references/disambiguation.md)

Key clusters and their owners:

- **Performance**: `golang-performance` (optimization patterns) · `golang-benchmark` (measurement) · `golang-troubleshooting` (root cause) · `golang-observability` (always-on production)
- **DI**: `golang-dependency-injection` (concepts/decision) · `golang-google-wire` (compile-time) · `golang-uber-dig` (runtime reflection) · `golang-uber-fx` (lifecycle framework) · `golang-samber-do` (type-safe container)
- **samber/\***: `golang-samber-lo` (finite transforms) · `golang-samber-ro` (reactive streams) · `golang-samber-mo` (monadic types)
- **Errors**: `golang-error-handling` (idioms) · `golang-samber-oops` (structured errors) · `golang-safety` (prevent panics)
- **Style**: `golang-code-style` · `golang-naming` · `golang-lint` · `golang-documentation`
- **CLI**: `golang-cli` (architecture) · `golang-spf13-cobra` (command tree) · `golang-spf13-viper` (config layering)
- **Package lookup**: `golang-pkg-go-dev` (query pkg.go.dev for an existing path: versions/docs/symbols/importers/CVEs) · `golang-gopls` (navigate/refactor your locally resolved build) · `golang-popular-libraries` (which library to adopt) · `golang-dependency-management` (manage go.mod) · `golang-security` (whole-tree CVE scan)
- **Gap — type vs arch**: `golang-structs-interfaces` (type design) vs `golang-design-patterns` (architectural patterns)
- **Gap — goroutine vs cancel**: `golang-concurrency` + `golang-context` — load both when cancelling goroutines via context
- **Gap — correctness vs threat**: `golang-safety` (internal bugs) vs `golang-security` (external threats)
- **Gap — features vs rules**: `golang-modernize` (language adoption) vs `golang-lint` (static analysis config)
- **Gap — process vs target rules**: `golang-refactoring` (the safe, staged, at-scale _process_ of changing existing code — planning, ordering, gopls-driven mechanics, staged PRs) vs `golang-naming`/`golang-code-style`/`golang-project-layout`/`golang-design-patterns`/`golang-modernize` (what the resulting code should look like) — load `golang-refactoring` alongside whichever of these owns the target shape

## Configure mode

On request, write a conditional routing directive for `golang-how-to` to the project's agent-config file (CLAUDE.md, AGENTS.md, GEMINI.md, Cursor rules, or Copilot instructions — whichever the project's harness reads), plus an optional `## Go skills` block whose lines each say when a skill applies. Keep every directive conditional: an unconditional "load X first" line loads that skill's full body on every Go task, not just its description.

`samber/cc-skills-golang@golang-project-layout` offers this at project creation; `/golang-how-to configure` runs it directly. Follow [project-config.md](references/project-config.md).

---

This skill is not exhaustive. Refer to individual skill files and the official Go documentation for detailed guidance.

If you encounter a bug or unexpected behavior in this skill plugin, open an issue at <https://github.com/samber/cc-skills-golang/issues>.
