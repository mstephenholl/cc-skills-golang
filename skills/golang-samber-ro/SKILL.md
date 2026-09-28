---
name: golang-samber-ro
description: "Reactive streams in Golang with samber/ro — observables, operators, subjects, backpressure, and plugins for event-driven pipelines. Apply when using or adopting samber/ro, or when the codebase imports `github.com/samber/ro`. Not for finite slice transforms (→ See `samber/cc-skills-golang@golang-samber-lo` skill) or plain channel pipelines (→ See `samber/cc-skills-golang@golang-concurrency` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "👁"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "0.3.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent mcp__context7__resolve-library-id mcp__context7__query-docs AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who reaches for reactive streams when data flows asynchronously or infinitely. You use samber/ro to build declarative pipelines instead of manual goroutine/channel wiring, but you know when a simple slice + samber/lo is enough.

# samber/ro — Reactive Streams for Go

Go implementation of [ReactiveX](https://reactivex.io/): generics-first, type-safe pipelines for asynchronous streams with backpressure, error propagation and context integration. The library is v0 — it follows semver, but exported APIs may still break before v1.0.0, so pin the version rather than tracking the latest minor.

**Official Resources:**

- [github.com/samber/ro](https://github.com/samber/ro)
- [ro.samber.dev](https://ro.samber.dev)
- [pkg.go.dev/github.com/samber/ro](https://pkg.go.dev/github.com/samber/ro)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

Reach for `ro` when values arrive over time, come from several sources, or need time-aware operators (retry, timeout, throttle, buffer). A finite slice belongs to `samber/lo` (→ See `samber/cc-skills-golang@golang-samber-lo` skill), and bounded goroutine fan-out to `errgroup` — both skip the subscription and goroutine overhead.

## Core Pattern

```go
observable := ro.Pipe2( // typed PipeN checks each operator's types at compile time
    ro.RangeWithInterval(0, 5, 1*time.Second),
    ro.Filter(func(x int) bool { return x%2 == 0 }),
    ro.Map(func(x int) string { return fmt.Sprintf("even-%d", x) }),
)

sub := observable.Subscribe(ro.NewObserver(
    func(s string) { fmt.Println(s) },   // onNext
    func(err error) { log.Println(err) }, // onError — without it, errors vanish
    func() { fmt.Println("Done!") },      // onComplete
))
sub.Wait() // block until complete or error; sub.Unsubscribe() cancels

// Finite stream: Collect blocks until completion and returns ([]T, error)
values, err := ro.Collect(observable)
```

## Cold vs Hot Observables

**Cold** (default): each `.Subscribe()` starts a new independent execution. Safe and predictable — use by default.

**Hot**: multiple subscribers share a single execution. Use when the source is expensive (WebSocket, DB poll) or subscribers must see the same events.

| Convert with | Behavior |
| --- | --- |
| `Share()` | Cold → hot with reference counting; the source is torn down when the last subscriber leaves |
| `ShareReplay(n)` | Same as Share + buffers last N values for late subscribers |
| `Connectable()` | Cold → hot, but waits for an explicit `.Connect()` so every subscriber can attach first |
| Subjects | Natively hot — push with `.Next()`, `.Error()`, `.Complete()` |

| Subject | Constructor | Replay behavior |
| --- | --- | --- |
| `PublishSubject` | `NewPublishSubject[T]()` | None — late subscribers miss past events |
| `BehaviorSubject` | `NewBehaviorSubject[T](initial)` | Replays last value to new subscribers |
| `ReplaySubject` | `NewReplaySubject[T](bufferSize)` | Replays last N values |
| `AsyncSubject` | `NewAsyncSubject[T]()` | Emits only last value, only on complete |
| `UnicastSubject` | `NewUnicastSubject[T](bufferSize)` | Single subscriber only; buffers until it attaches |

## Common Mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| Subscribing with `ro.OnNext()` alone | Errors are silently dropped — bugs hide in production | `ro.NewObserver(onNext, onError, onComplete)` |
| Untyped `ro.Pipe()` | Operators are `any`, so type mismatches surface at runtime | `Pipe2` … `Pipe25` |
| Leaving an infinite stream unbounded | Goroutines leak; `FromChannel` completes only when its channel closes | Bound it with `Take(n)`, `TakeUntil(signal)`, `Timeout(d)` or context cancellation, or call `Unsubscribe()` |
| Passing `ctx` only to `Subscribe` | The pipeline ignores cancellation and keeps running on shutdown | Chain `ContextWithTimeout` or `ContextReset` with `ThrowOnContextCancel`, and handle the error in `onError` |
| `RetryConfig` without `MaxRetries` | `0` means retry forever; there is no backoff field, only a fixed `Delay` | Set `MaxRetries`, and put exponential backoff inside the source if needed |
| Fallback before retry | `Catch`/`OnErrorReturn` swallows the error, so the retry never fires | Order: `RetryWithConfig` → `Catch`/`OnErrorResumeNextWith` → `OnErrorReturn` |
| Logging inside `Map` | Mixes side effects into the transform | `Tap`, `TapOnNext`, `TapOnError` observe without altering the stream |
| `Share()` when one consumer suffices | Extra lifecycle to reason about | Stay cold until several consumers need the same execution |

## References

- [references/operators-guide.md](references/operators-guide.md) — when choosing an operator (combining with `CombineLatest`/`Zip`/`Merge`, `FlatMap`, `MapErr`, `Scan`, buffering, error handling, context, terminal) or checking its signature.
- [references/subjects-guide.md](references/subjects-guide.md) — when pushing values imperatively into a stream, choosing a Subject, or converting cold to hot with `Share`, `ShareReplay` or `Connectable`.
- [references/patterns.md](references/patterns.md) — when building retry + timeout + fallback, fan-in, dependent-stream combination, running aggregation, a file watcher, or graceful shutdown.
- [references/plugin-ecosystem.md](references/plugin-ecosystem.md) — when the source or sink is HTTP, files, cron schedules, OS signals, JSON/CSV, a logger, or a rate limiter.

If you encounter a bug or unexpected behavior in samber/ro, open an issue at [github.com/samber/ro/issues](https://github.com/samber/ro/issues).

## Cross-References

- → See `samber/cc-skills-golang@golang-samber-mo` skill for monadic types (Option, Result, Either) that compose with ro pipelines
- → See `samber/cc-skills-golang@golang-samber-hot` skill for in-memory caching (also available as an ro plugin)
- → See `samber/cc-skills-golang@golang-concurrency` skill for goroutine/channel patterns when reactive streams are overkill
- → See `samber/cc-skills-golang@golang-observability` skill for monitoring reactive pipelines in production
