---
name: golang-how-to
description: "Golang skills router for samber/cc-skills-golang — maps a task to the Go skills it needs and draws the boundary between overlapping siblings (performance vs benchmark vs troubleshooting, samber/lo vs mo vs ro, the DI libraries, safety vs security). Use when a Go task spans several concerns, when two Go skills seem to compete, or to add conditional Go skill directives to a project's CLAUDE.md, AGENTS.md, or Cursor rules (/golang-how-to configure)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires git.
metadata:
  author: samber
  version: "1.5.0"
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

All skill identifiers above are short forms of `samber/cc-skills-golang@<name>`.

## Code navigation with gopls

`gopls` gives semantic code intelligence for Go — go-to-definition, find references, diagnostics, package API, symbol search, refactoring. → See `samber/cc-skills-golang@golang-gopls` skill for the three ways to reach it (its own MCP server, the native `LSP` tool, and its CLI), the full capability matrix, and efficient read/edit workflows.

`gopls` only reasons about code that is present and resolvable in the local build: your workspace plus every dependency exactly as pinned in `go.sum` (including `replace` directives). For any fact that isn't tied to your local build — version history, licenses, ecosystem-wide importers, a package you haven't added yet — use `golang-pkg-go-dev` (`godig`). See the `godig` vs gopls vs Context7 vs govulncheck section below for the full boundary.

## `godig` vs gopls vs Context7 vs govulncheck

Four tools can answer "is this dependency OK to use," and they don't overlap as much as they look:

- **Context7** is a general-purpose, cross-language documentation fetcher — useful when no more specific source exists. For a Go package or module, `godig` is almost always the better choice: it pulls **structured, Go-specific data** straight from pkg.go.dev — exact versions, exported symbols with signatures, runnable examples, `imported-by`, and known vulnerabilities — rather than Context7's generic scraped/curated docs, which don't expose that structure and can lag or miss lesser-known Go modules. Reach for Context7 only when a dependency's documentation genuinely doesn't exist or isn't indexed on pkg.go.dev.
- **`godig`** answers questions about the **published ecosystem**: any Go package or module, whether or not it's in your `go.mod` yet — it calls the remote pkg.go.dev API and never touches your local checkout. Its `vulns` command reports CVEs known for a package/version in isolation, regardless of whether your build actually reaches the vulnerable code path.
- **`gopls`** (→ `samber/cc-skills-golang@golang-gopls`, via its MCP server, the native `LSP` tool, or its CLI) answers questions about **your specific build**: your code plus every dependency exactly as pinned in `go.sum`, including `replace` directives pointing at forks or local paths — neither `godig` nor Context7 can see that. Its `go_vulncheck` operation runs a single, on-demand reachability check against the workspace as it stands right now.
- **`govulncheck`** (the standalone CLI, wrapped by the `samber/cc-skills-golang@golang-security` skill) is the whole-tree audit: it walks the entire module's call graph to confirm which known vulnerabilities are actually reachable, and is the tool of record for CI gates and periodic security sweeps — `gopls`'s `go_vulncheck` is a lighter-weight, single-shot version of the same analysis for use mid-edit.

Pick by task:

| Task | Tool | How |
| --- | --- | --- |
| Find where a symbol is defined in your own repo | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_search`, then `go_file_context` |
| Understand a file's intra-package dependencies | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_file_context` |
| Jump into a dependency's exact resolved source (incl. forks/`replace`d versions) | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_package_api`, or the native `LSP` tool's `goToDefinition` |
| Find every call site in your own code that references a dependency's symbol | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_symbol_references` — `godig`'s `imported-by` only lists public _packages_, not call sites in your repo |
| Get compiler diagnostics right after an edit | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_diagnostics` (MCP), or automatic with the native `LSP` tool |
| Check whether your current build can reach a known vulnerability, mid-edit | `gopls` | `samber/cc-skills-golang@golang-gopls` — `go_vulncheck` |
| Rename, extract, inline, or otherwise refactor local code | `gopls` | `samber/cc-skills-golang@golang-gopls` — safe rename, `refactor.*` code actions |
| Whole-tree vulnerability audit across the module (CI, periodic sweep) | `govulncheck` | `samber/cc-skills-golang@golang-security` skill — `govulncheck ./...` |
| List available versions of a published package | `godig` | `godig versions <path>` |
| Check known CVEs for a package/version you haven't added yet | `godig` | `godig vulns <path>` |
| See exported symbols/signatures of a published package | `godig` | `godig symbols` / `symbol doc` |
| Get runnable code examples for a symbol | `godig` | `godig symbol examples` |
| Read a package's rendered README/docs | `godig` | `godig module readme` / `package doc` |
| See who imports a package across the whole public ecosystem | `godig` | `godig imported-by` |
| Search for a package or library candidate | `godig` | `godig search` |
| Check a package's or module's license | `godig` | `godig package licenses` / `module licenses` |
| Get docs for a non-Go library, or a Go module not indexed on pkg.go.dev | Context7 | library docs lookup (resolve the library, then query its docs) |

See the `samber/cc-skills-golang@golang-pkg-go-dev` skill for the full `godig` command reference, and the `samber/cc-skills-golang@golang-security` skill for the whole-tree `govulncheck` remediation workflow.

## Categories at a glance

Full catalog with "use when" hooks: [by-category.md](references/by-category.md)

| Category | Skills |
| --- | --- |
| Code Quality | `golang-code-style` `golang-documentation` `golang-error-handling` `golang-lint` `golang-naming` `golang-safety` `golang-security` `golang-structs-interfaces` |
| Architecture & Design | `golang-concurrency` `golang-context` `golang-data-structures` `golang-database` `golang-dependency-injection` `golang-design-patterns` `golang-modernize` `golang-refactoring` |
| QA & Performance | `golang-benchmark` `golang-observability` `golang-performance` `golang-testing` `golang-troubleshooting` |
| Project Setup | `golang-cli` `golang-continuous-integration` `golang-dependency-management` `golang-gopls` `golang-pkg-go-dev` `golang-popular-libraries` `golang-project-layout` `golang-stay-updated` |
| APIs | `golang-graphql` `golang-grpc` `golang-swagger` |
| Dependency Injection | `golang-dependency-injection` `golang-google-wire` `golang-uber-dig` `golang-uber-fx` `golang-samber-do` |
| Frameworks | `golang-spf13-cobra` `golang-spf13-viper` |
| samber/\* | `golang-samber-do` `golang-samber-hot` `golang-samber-lo` `golang-samber-mo` `golang-samber-oops` `golang-samber-ro` `golang-samber-slog` |
| Testing | `golang-stretchr-testify` `golang-testing` |

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
