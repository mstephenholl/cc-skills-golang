---
name: golang-context
description: "Idiomatic context.Context usage in Golang — propagation through API boundaries, cancellation, timeouts and deadlines, request-scoped values, context.WithoutCancel for background work outliving requests. Apply when designing context propagation across layers, debugging leaked or unexpired contexts, choosing between context.Background/TODO/WithoutCancel, or storing values in context. Not for code that merely accepts ctx as first parameter."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.2"
  openclaw:
    emoji: "🔗"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
---

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-context` skill takes precedence.

# Go context.Context Best Practices

`context.Context` is Go's mechanism for propagating cancellation signals, deadlines, and request-scoped values across API boundaries and between goroutines. Think of it as the "session" of a request — it ties together every operation that belongs to the same unit of work.

## Best Practices Summary

1. Propagate the same context through the entire request lifecycle — `r.Context()` in the handler → service → DB → external APIs, using each call's context-aware variant (`http.NewRequestWithContext`, `QueryContext`, `BeginTx`) — any link that starts a fresh context keeps working after the client is gone.
2. Take `ctx context.Context` as the first parameter (linter-enforced), and pass context through parameters instead of storing it in a struct — the struct outlives the request that filled it, so later calls reuse a context that is already cancelled or belongs to someone else.
3. Call `cancel()` on all control-flow paths for `WithCancel`/`WithTimeout`/`WithDeadline`, unless ownership of the context and cancel function is explicitly returned or transferred — an uncalled `cancel()` keeps the child attached to its parent and leaks its timer until the parent finishes.
4. Create `context.Background()` only at top-level entry points (main, init, tests) — deeper in the call chain, especially mid-request, it detaches the work from the caller's deadline and cancellation.
5. Use `context.TODO()` as the placeholder when a context is needed but none exists yet, never `nil` — `TODO()` marks the gap for a later fix instead of hiding it behind a `Background()` that looks deliberate, and a `nil` context panics on the first `Done()` or `Value()` call, far from the caller that passed it.
6. Declare context value keys as unexported types — with a plain `string` key, two packages using `"user"` silently overwrite each other.
7. Carry only request-scoped metadata in context values, never function parameters — values retrieved through `Value()` lose compile-time typing and disappear from the function signature.
8. Use `context.WithoutCancel` (Go 1.21+) when spawning background work that must outlive the parent request — otherwise the handler returning cancels the audit log or cleanup just started.

## Deep Dives

- Read [cancellation.md](./references/cancellation.md) when adding timeouts or deadlines (nested ones take the shorter), reacting to cancellation (`select`, `AfterFunc` cleanup), or detaching work that must outlive the request.
- Read [values-tracing.md](./references/values-tracing.md) when storing values in a context or propagating trace and correlation IDs across services.
- Read [http-services.md](./references/http-services.md) when wiring HTTP handlers, middleware, HTTP clients or database calls to the request context.

## Cross-References

- → See the `samber/cc-skills-golang@golang-concurrency` skill for goroutine cancellation patterns using context
- → See the `samber/cc-skills-golang@golang-database` skill for context-aware database operations (QueryContext, ExecContext)
- → See the `samber/cc-skills-golang@golang-observability` skill for trace context propagation with OpenTelemetry
- → See the `samber/cc-skills-golang@golang-design-patterns` skill for timeout and resilience patterns

## Enforce with Linters

Many context pitfalls are caught automatically by linters: `govet` (`lostcancel`), `staticcheck` (`SA1012`, nil context), `revive` (`context-as-argument`), `containedctx` (context stored in a struct). → See the `samber/cc-skills-golang@golang-lint` skill for configuration and usage.
