---
name: golang-lint
description: "Golang linting with golangci-lint — .golangci.yml configuration, choosing linters, nolint suppressions, and reading lint output. Use when configuring or running golangci-lint, go vet, staticcheck, or revive, or deciding how to handle a lint warning. Not for adding a lint job to CI (→ See `samber/cc-skills-golang@golang-continuous-integration` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.4.4"
  openclaw:
    emoji: "🧹"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - golangci-lint
    install:
      - kind: brew
        formula: golangci-lint
        bins: [golangci-lint]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
  - ".golangci.yml"
---

**Persona:** You are a Go code quality engineer. You treat linting as a first-class part of the development workflow — not a post-hoc cleanup step.

**Orchestration mode:** For adopting linting on a legacy codebase, fan out parallel sub-agents split by package — after one sequential `--fix` pass, disjoint packages can be fixed concurrently without conflicting edits — and consolidate into a clean run on the touched packages. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Setup mode** — configuring `.golangci.yml` and choosing linters: start from the recommended config below. Done when `golangci-lint config verify` passes and a full run completes.
- **Coding mode** — after an edit batch, run `golangci-lint run` on the changed packages and fix the findings.
- **Interpret/fix mode** — reading lint output, suppressing warnings, fixing issues in existing code: start from "Interpreting Output" and "Suppressing Lint Warnings"; for a large backlog, follow "Adopting Linting on a Legacy Codebase". Done when a run on the touched packages is clean, with no bare `//nolint`.

**Dependencies:**

- golangci-lint v2: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest`

# Go Linting

## Configuration

Keep a `.golangci.yml` at the repository root as the **source of truth** for which linters run and how — without one, golangci-lint falls back to its small `standard` set. Start from the [recommended .golangci.yml](./assets/.golangci.yml) (48 linters, with the rejected ones listed under `disable` and the reason for each); read it before enabling or disabling a linter. For what each linter checks and when it is useful, read the [linter reference](./references/linter-reference.md).

## Quick Reference

```bash
# Run all configured linters
golangci-lint run ./...

# Auto-fix issues where possible
golangci-lint run --fix ./...

# Run formatters (golangci-lint v2+) — separate from `run`
golangci-lint fmt ./...

# Run a single linter only
golangci-lint run --enable-only govet ./...

# List all available linters
golangci-lint linters

# Verbose output with timing info
golangci-lint run --verbose ./...
```

## Suppressing Lint Warnings

Use `//nolint` directives sparingly — fix the root cause first.

```go
// Good: specific linter + justification
//nolint:errcheck // fire-and-forget logging, error is not actionable
_ = logger.Sync()

// Bad: blanket suppression without reason
//nolint
_ = logger.Sync()
```

Rules:

1. **//nolint directives MUST specify the linter name**: `//nolint:errcheck` not `//nolint`
2. **//nolint directives MUST include a justification comment**: `//nolint:errcheck // reason`
3. **The `nolintlint` linter enforces both rules above** — it flags bare `//nolint` and missing reasons
4. **NEVER suppress security linters** (gosec, bodyclose, sqlclosecheck) without a very strong reason

Read **[nolint directives](./references/nolint-directives.md)** when deciding whether a finding deserves a fix or a suppression, or when suppressing several linters, a whole function, or a test file.

## Interpreting Output

Each issue follows this format:

```
path/to/file.go:42:10: message describing the issue (linter-name)
```

The linter name in parentheses tells you which linter flagged it. Use this to:

- Look up the linter in the [reference](./references/linter-reference.md) to understand what it checks
- Suppress with `//nolint:linter-name // reason` if it's a false positive
- Use `golangci-lint run --verbose` for additional context and timing

## Common Issues

| Problem | Solution |
| --- | --- |
| "deadline exceeded" | Set or increase `run.timeout` in `.golangci.yml`; golangci-lint v2 defaults to no timeout (`0`) |
| Linter not found | Check `golangci-lint linters` — linter may need a newer version |
| Conflicts between linters | Disable the less useful one with a comment explaining why |
| v1 config errors after upgrade | Run `golangci-lint migrate` to convert config format |
| Slow on large repos | Reduce `run.concurrency` or exclude paths with `linters.exclusions.paths` / `formatters.exclusions.paths` |

## Adopting Linting on a Legacy Codebase

The order matters — the `--fix` pass rewrites files across the tree, so it must finish before anything else edits:

1. Set `issues.new-from-rev` (e.g. `main` or `HEAD~1`) in `.golangci.yml` so only new and changed code must pass — adding `//nolint` to thousands of existing findings is unmaintainable.
2. Run `golangci-lint run --fix ./...` once, alone, as the mechanical first pass.
3. Split the remaining findings **by package** across parallel sub-agents, so each edits disjoint files — splitting by linter category puts several agents in the same file. Within a package, fix security and resource-leak findings (gosec, bodyclose, sqlclosecheck) before error handling, then style.
4. Move the `new-from-rev` baseline forward as packages come clean.

## Cross-References

- → See `samber/cc-skills-golang@golang-continuous-integration` skill for the CI pipeline with golangci-lint-action, and for automated AI-driven code review in CI
- → See `samber/cc-skills-golang@golang-code-style` skill for style rules that linters enforce
- → See `samber/cc-skills-golang@golang-security` skill for SAST tools beyond linting (gosec, govulncheck)
