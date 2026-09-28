---
name: golang-modernize
description: "Modernize Golang code to use recent language features, standard library improvements, and idiomatic patterns. Use when reviewing Go code with old-style patterns, when encountering a deprecation warning, or when the user asks for modernization, a Go version upgrade (e.g. to Go 1.27), or a CI/tooling refresh. Not for structural refactors, extracting functions, or moving code between packages (→ See `samber/cc-skills-golang@golang-refactoring` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.5.2"
  openclaw:
    emoji: "🔄"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.27"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch WebSearch AskUserQuestion EnterWorktree ExitWorktree
paths:
  - "**/*.go"
---

<!-- markdownlint-disable ol-prefix -->

**Persona:** You are a Go modernization engineer. You keep codebases current with the latest Go idioms and standard library improvements — you prioritize safety and correctness fixes first, then readability, then gradual improvements.

**Orchestration mode:** For a full-codebase modernization scan, fan out parallel read-only sub-agents split by category — deprecated APIs, language features, standard library, testing patterns, tooling and CI — skipping categories the codebase doesn't touch, and consolidate into one list ranked by the Migration Priority Guide. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Inline mode** (the developer is working on something else): modernize only code you're already changing; list other opportunities in one closing line; drop them if declined. A broad rewrite during someone else's task buries their change under unrelated churn.
- **Full-scan mode** (explicit `/golang-modernize` invocation or CI): scan read-only (see Orchestration mode), rank by the Migration Priority Guide, then apply the rewrite in an isolated worktree so a sweeping multi-file change never touches the developer's main tree until reviewed. Done when the ranked list is delivered and, if the user asked for changes, they are applied in the worktree with `go test ./...` green.

# Go Code Modernization Guide

**Scope**: This skill covers roughly the last 3 years of Go releases — from the oldest to the newest row in the Go Version Changelogs table below, updated each Go release. Projects targeting an older `go.mod` still get suggestions, with narrower coverage; older modernizations that are still commonly missed (`any`, `errors.Is`/`errors.As`, `min`/`max`, `slices`) live in the baseline section of [versions.md](./references/versions.md#baseline-and-tooling-priorities).

## Ground rules

- **Gate every suggestion on the module's `go` directive** (`go.mod` or `go.work`) — an API newer than the directive breaks the build, and since Go 1.27 `go test` runs the `stdversion` vet check that flags it; bump the directive or revert the suggestion, don't silence the check. When the directive lags the newest row of the changelog table, suggest the upgrade and the features it unlocks.
- **Respect `.modernize`** in the project root — it lists suggestions the developer declined, so never re-suggest them. When the developer explicitly declines one, append a line so it isn't raised again.
- **Verify a dependency update** with `go mod tidy` and the test suite before suggesting it. Major version upgrades may contain breaking changes — the dependency's changelog documents them.
- **Rename or replace APIs with gopls** (e.g. `reflect.PtrTo` → `PointerTo`, `math/rand` → `math/rand/v2`) → See `samber/cc-skills-golang@golang-gopls` skill — safe rename updates every call site and refuses a rename that would break interface satisfaction, and post-edit diagnostics catch compile errors a grep/sed sweep would leave behind.

### `.modernize` file format

```
# Ignored modernization suggestions
# Format: <date> <category> <description>
2026-01-15 slog-migration Team decided to keep zap for now
2026-02-01 math-rand-v2 Legacy module requires math/rand compatibility
```

## Go Version Changelogs

Reference the relevant changelog when suggesting a modernization:

| Version | Release       | Changelog                   |
| ------- | ------------- | --------------------------- |
| Go 1.21 | August 2023   | <https://go.dev/doc/go1.21> |
| Go 1.22 | February 2024 | <https://go.dev/doc/go1.22> |
| Go 1.23 | August 2024   | <https://go.dev/doc/go1.23> |
| Go 1.24 | February 2025 | <https://go.dev/doc/go1.24> |
| Go 1.25 | August 2025   | <https://go.dev/doc/go1.25> |
| Go 1.26 | February 2026 | <https://go.dev/doc/go1.26> |
| Go 1.27 | August 2026   | <https://go.dev/doc/go1.27> |

Versions newer than Go 1.27 are documented in the official Go release notes.

## Using the modernize linter

The `modernize` linter (available since **golangci-lint v2.6.0**) automatically detects code that can be rewritten using newer Go features. It originates from `golang.org/x/tools/go/analysis/passes/modernize`; `gopls` and `go fix` (rewritten onto the `go/analysis` framework in Go 1.26, with fixer coverage still growing in Go 1.27 — see [Tooling modernization](./references/tooling.md) for the exact fixer list) cover overlapping modernization checks, but exact coverage differs by tool version. See the `samber/cc-skills-golang@golang-lint` skill for configuration.

## Version-specific and tooling modernizations

- Read [versions.md](./references/versions.md) before suggesting or applying a modernization — the sections up to the module's `go` directive hold what is already available, and the ones past it hold what an upgrade would unlock; each gives the before/after for the items in the priority guide below.
- Read [tooling.md](./references/tooling.md) when the task touches CI, golangci-lint, govulncheck, PGO, `go fix`, or the toolchain version.

## Deprecated Packages Migration

| Deprecated | Replacement | Since |
| --- | --- | --- |
| `math/rand` | `math/rand/v2` | Go 1.22 |
| `crypto/elliptic` (most functions) | `crypto/ecdh` | Go 1.21 |
| `reflect.SliceHeader`, `StringHeader` | `unsafe.Slice`, `unsafe.String` | Go 1.21 |
| `reflect.PtrTo` | `reflect.PointerTo` | Go 1.22 |
| `runtime.GOROOT()` | `go env GOROOT` | Go 1.24 |
| `runtime.SetFinalizer` | `runtime.AddCleanup` | Go 1.24 |
| `crypto/cipher.NewOFB`, `NewCFB*` | AEAD modes or `NewCTR` | Go 1.24 |
| `golang.org/x/crypto/sha3` | `crypto/sha3` | Go 1.24 |
| `golang.org/x/crypto/hkdf` | `crypto/hkdf` | Go 1.24 |
| `golang.org/x/crypto/pbkdf2` | `crypto/pbkdf2` | Go 1.24 |
| `testing/synctest.Run` | `testing/synctest.Test` | Go 1.25 |
| `crypto/rsa.EncryptPKCS1v15` for new encryption use | RSA-OAEP (`rsa.EncryptOAEP` / `rsa.EncryptOAEPWithOptions`) or HPKE/KEM design | Go 1.26 |
| `net/http/httputil.ReverseProxy.Director` | `ReverseProxy.Rewrite` | Go 1.26 |
| `crypto/tls.Config.Rand` | `testing/cryptotest.SetGlobalRandom()` | Go 1.27 |
| `github.com/google/uuid` (simple cases) | `uuid` (stdlib) | Go 1.27 |

## Go 1.27+ version-bump risk checklist

Several Go 1.27 changes need **verification, not a rewrite**, before a `go.mod` bump ships. Most notably: a `godebug` line in `go.mod` (or `//go:debug` comment) still pinning `asynctimerchan`, `tlsunsafeekm`, `tlsrsakex`, `tls3des`, `tls10server`, `x509keypairleaf`, or `gotypesalias` to its **old** value now fails the build. Full checklist in [Go version modernizations](./references/versions.md#go-127-version-bump-risk-checklist-verify-dont-rewrite).

## Migration Priority Guide

When modernizing a codebase, prioritize changes by impact. Pre-Go 1.22 items and CI tooling sit in the [baseline section](./references/versions.md#baseline-and-tooling-priorities) of versions.md.

### High priority (safety and correctness)

1. Remove loop variable shadow copies _(Go 1.22+)_ — prevents subtle bugs
2. Replace `math/rand` with `math/rand/v2` _(Go 1.22+)_ — remove `rand.Seed` calls
3. Use `os.Root` for user-supplied file paths _(Go 1.24+)_ — prevents path traversal
4. Run `govulncheck` _(Go 1.22+)_ — catch known vulnerabilities
5. Migrate deprecated crypto packages _(Go 1.24+)_ — security critical
6. Before bumping to `go 1.27`, resolve removed `GODEBUG` keys and `crypto/tls.Config.Rand` callers _(Go 1.27+)_ — see the risk checklist above; a stale `GODEBUG` value now fails the build

### Medium priority (readability and maintainability)

7. Use `range` over int _(Go 1.22+)_
8. Use `cmp.Or` for default values _(Go 1.22+)_
9. Use `sync.WaitGroup.Go` _(Go 1.25+)_
10. Use `t.Context()` in tests _(Go 1.24+)_
11. Use `b.Loop()` in benchmarks _(Go 1.24+)_
12. Use generic methods for helpers scoped to one type, and `strings.CutLast`/`bytes.CutLast` instead of `LastIndex` slicing _(Go 1.27+)_
13. Migrate to the `encoding/json/v2` API — the new default since Go 1.27; review its duplicate-key and invalid-UTF-8 strictness against real payloads first _(Go 1.27+)_

### Lower priority (gradual improvement)

14. Adopt iterators where they simplify code _(Go 1.23+)_
15. Use `strings.SplitSeq` and iterator variants _(Go 1.24+)_
16. Move tool deps to `go.mod` tool directives _(Go 1.24+)_
17. Replace `google/uuid`/`gofrs/uuid` with the stdlib `uuid` package, after checking for RFC-variant features the stdlib doesn't cover _(Go 1.27+)_
18. Run `go fix ./...` after a toolchain upgrade to apply the safe automated transformations _(Go 1.26+, more fixers in Go 1.27)_

## Related Skills

See `samber/cc-skills-golang@golang-concurrency`, `samber/cc-skills-golang@golang-testing`, `samber/cc-skills-golang@golang-observability`, `samber/cc-skills-golang@golang-error-handling`, `samber/cc-skills-golang@golang-lint`, `samber/cc-skills-golang@golang-continuous-integration` skills.

- → See `samber/cc-skills-golang@golang-refactoring` skill for staging a large modernization sweep as small human-reviewed PRs instead of one big worktree sweep.
