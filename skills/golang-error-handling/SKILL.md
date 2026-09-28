---
name: golang-error-handling
description: "Idiomatic Golang error design — sentinel vs custom error types, wrapping with %w, errors.Is/As/Join, panic recovery, and logging an error once with slog. Use when designing error types, deciding whether to wrap, return, or log an error, or fixing duplicate or high-cardinality error logs. For samber/oops → See `samber/cc-skills-golang@golang-samber-oops` skill; for slog handler pipelines → See `samber/cc-skills-golang@golang-samber-slog` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.3"
  openclaw:
    emoji: "⚠"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
---

**Persona:** You are a Go reliability engineer. You treat every error as an event that must either be handled or propagated with context — silent failures and duplicate logs are equally unacceptable.

**Orchestration mode:** For a codebase-wide error-handling audit, fan out parallel sub-agents split by package — each package's error paths can be traced independently — and consolidate into one per-file findings report. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Coding** — writing new error-handling code; apply the rules below as you write.
- **Review** — a PR's error-handling changes: start from the diff, then trace where each returned error is handled upstream, since a log-and-return pair often spans two files. Findings with file:line are the deliverable; fix when asked.
- **Audit** — find swallowed or discarded errors, log-and-return pairs, `%w` leaking past boundaries, and goroutines without `recover`; report by file, and fix when asked.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-error-handling` skill takes precedence.

# Go Error Handling Best Practices

## Rules

1. **Wrap with context** using `fmt.Errorf("{context}: %w", err)` so each layer names what it was doing; keep error strings lowercase with no trailing punctuation, because they are concatenated into longer chains.
2. **Use `%w` internally, `%v` at system boundaries** — `%w` makes the wrapped error part of your API, so at a public boundary it lets callers `errors.As` into backend-internal types and couples them to your implementation.
3. **Match with `errors.Is` and `errors.As`/`errors.AsType`, not `==` or bare type assertions** — wrapping breaks direct comparison. For Go 1.26+, prefer `errors.AsType[T](err)` when `T` implements `error`; use `errors.As(err, &target)` for older Go or non-error interface targets.
4. **Combine independent errors with `errors.Join`** (Go 1.20+) — validating every field, closing every resource, failing batch items — instead of stopping at the first error or hand-rolling a multi-error type.
5. **Log OR return an error, never both** (single handling rule) — every layer that does both adds a duplicate log entry for one failure.
6. **Use sentinel errors** (`errors.New`, package-level `var`) for expected conditions, and custom types when callers need data from the error.
7. **Reserve `panic` for unrecoverable states** — expected conditions such as bad input or a missing row return errors.
8. **Log with `slog`** (Go 1.21+) key-value attributes rather than `fmt.Println` or `log.Printf`, with the level matching severity.
9. **Use `samber/oops`** for production errors needing stack traces, user/tenant context, or structured attributes.
10. **Log HTTP requests** with structured middleware capturing method, path, status, and duration.
11. **Never expose technical errors to users** — translate internal errors to user-friendly messages, log technical details separately.
12. **Keep log grouping low-cardinality** — at logging/APM boundaries, keep message templates stable and attach IDs, paths, line numbers, and counts as structured attributes. Error values may include useful operational context, but avoid putting high-cardinality data into the stable log message used for grouping.

## Detailed Reference

- Read [error-creation.md](./references/error-creation.md) when choosing between a sentinel and a custom error type, or phrasing an error message.
- Read [error-wrapping.md](./references/error-wrapping.md) when inspecting or joining error chains, or deciding between `%w` and `%v`.
- Read [error-handling.md](./references/error-handling.md) when deciding where to log, where to recover from panics, or whether to adopt `samber/oops`.

## Cross-References

- → See `samber/cc-skills-golang@golang-samber-oops` for full samber/oops API, builder patterns, and logger integration
- → See `samber/cc-skills-golang@golang-observability` for structured logging setup, log levels, and request logging middleware
- → See `samber/cc-skills-golang@golang-safety` for nil interface trap and nil error comparison pitfalls
- → See `samber/cc-skills-golang@golang-naming` for error naming conventions (ErrNotFound, PathError)
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for automated AI-driven code review in CI using these guidelines

## References

- [lmittmann/tint](https://github.com/lmittmann/tint)
- [samber/oops](https://github.com/samber/oops)
- [samber/slog-multi](https://github.com/samber/slog-multi)
- [samber/slog-sampling](https://github.com/samber/slog-sampling)
- [samber/slog-formatter](https://github.com/samber/slog-formatter)
- [samber/slog-http](https://github.com/samber/slog-http)
- [samber/slog-sentry](https://github.com/samber/slog-sentry)
- [log/slog package](https://pkg.go.dev/log/slog)
