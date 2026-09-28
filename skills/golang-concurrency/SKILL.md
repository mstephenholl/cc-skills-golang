---
name: golang-concurrency
description: "Golang concurrency design — goroutine lifecycle and leaks, channel ownership, channels vs mutexes vs atomics, errgroup, singleflight, worker pools, and pipelines. Use when writing or reviewing code that spawns goroutines or shares state across them. Not for nil or aliasing bugs (→ See `samber/cc-skills-golang@golang-safety` skill), nor for debugging an already hung or racing program (→ See `samber/cc-skills-golang@golang-troubleshooting` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.3"
  openclaw:
    emoji: "⚡"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go concurrency engineer. You assume every goroutine is a liability until proven necessary — correctness and leak-freedom come before performance.

**Orchestration mode:** For a codebase-wide concurrency audit, fan out parallel sub-agents split by package — goroutine spawns and shared state are scattered, and each package can be scanned independently — and consolidate into the Audit-mode report. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Write** — implement concurrent code (goroutines, channels, sync primitives, worker pools, pipelines). Done when every goroutine you add passes the checklist below and `go test -race` passes where tests exist.
- **Review** — a PR diff. Start from the goroutines, channels and shared state the diff touches, then trace their callers — a leak often lives outside the diff. Deliverable: leaks and races with file:line; fixes applied if the user asked for them.
- **Audit** — inventory each goroutine's exit path, each shared mutable value's guard, and each channel's owner and closer; report leaks and races with file:line; confirm with `-race` where tests exist. Scans are read-only and parallelize freely; if the user asked for fixes, give each agent disjoint files, apply them, and re-run `-race`.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-concurrency` skill takes precedence.

# Go Concurrency Best Practices

Goroutines are cheap but not free — every one you spawn is a resource you must manage. Aim for structured concurrency: every goroutine has a clear owner, a predictable exit, and a path for its errors.

## Before Spawning a Goroutine

- [ ] **How will it exit, and who tells it to?** — a `context.Context`, done channel or closed input; without one it leaks and accumulates until the process runs out of memory
- [ ] **Who waits for it?** — `sync.WaitGroup` or `errgroup`; call `wg.Add` before `go`, because an `Add` inside the goroutine races with `Wait`, which can return early
- [ ] **Who owns the channels?** — the creator/sender owns and closes them
- [ ] **Should this be synchronous instead?** — don't add concurrency without measured need

## Rules

| Rule | Why |
| --- | --- |
| Bound the number of goroutines — `errgroup.SetLimit(n)` or a semaphore | One `go` per item exhausts memory and downstream connections under load |
| Only the sender closes a channel; a receiver that wants to stop signals through a context or done channel | A send on a closed channel panics |
| Specify channel direction in signatures (`chan<-`, `<-chan`) | The compiler rejects misuse at build time |
| Default to unbuffered channels; size a buffer only from measurement | Large buffers mask backpressure and hide slow consumers |
| Send values, not pointers, on channels | A pointer keeps memory shared with the sender, defeating the ownership transfer |
| Include `ctx.Done()` in any `select` that blocks on a caller's behalf — sends as well as receives | Without it the goroutine outlives the caller's cancellation |
| Reuse one `time.NewTimer` + `Reset` in long-running loops instead of `time.After` | Each `time.After` call allocates a new timer — churn in hot loops |
| Keep mutex critical sections short; never hold a lock across I/O | A slow call under the lock stalls every waiter |
| Test with `go test -race` and `go.uber.org/goleak` | Races and leaks rarely show up in a plain test run |

## Channel vs Mutex vs Atomic

| Scenario | Use | Why |
| --- | --- | --- |
| Passing data between goroutines | Channel | Communicates ownership transfer |
| Coordinating goroutine lifecycle | Channel + context | Clean shutdown with select |
| Protecting shared struct fields | `sync.Mutex` / `sync.RWMutex` | Simple critical sections |
| Simple counters, flags | `sync/atomic` | Lock-free, lower overhead |
| Map whose keys are written once, or whose goroutines touch disjoint keys | `sync.Map` | Otherwise `RWMutex` + map is faster. **Unsynchronized concurrent map read/write is a fatal runtime error, not just a race** |
| Caching expensive computations | `sync.Once` / `singleflight` | Execute once or deduplicate |

## WaitGroup vs errgroup

| Need | Use | Why |
| --- | --- | --- |
| Wait for goroutines, errors not needed | `sync.WaitGroup` | Fire-and-forget |
| Wait + collect first error | `errgroup.Group` | Error propagation |
| Wait + cancel siblings on first error | `errgroup.WithContext` | Context cancellation on error |
| Wait + limit concurrency | `errgroup.SetLimit(n)` | Built-in worker pool |

## Sync Primitive Gotchas

| Primitive | Key notes |
| --- | --- |
| `sync.RWMutex` | Never upgrade `RLock` to `Lock` — it deadlocks |
| `sync/atomic` | Prefer typed atomics (Go 1.19+): `atomic.Int64`, `atomic.Bool` |
| `sync.Pool` | `Reset()` before `Put()`; never assume `Get()` returns a clean object |
| `sync.Once` | Go 1.21+: `OnceFunc`, `OnceValue`, `OnceValues` |
| `sync.WaitGroup` | Go 1.25+: prefer `wg.Go(func(){ ... })` for fire-and-wait tasks that do not panic and do not need error propagation. For Go <1.25 use `Add`/`Done`. For errors/cancellation/limits, use `errgroup` with context. |
| `x/sync/singleflight` | Deduplicates concurrent calls for the same key — prevents cache stampedes |

## Deep Dives

- Read [channels-and-select.md](references/channels-and-select.md) when writing a goroutine's lifecycle or panic recovery, closing or buffering a channel, or writing a `select` loop with timers.
- Read [sync-primitives.md](references/sync-primitives.md) when using a mutex, atomic, `sync.Map`, `sync.Pool`, `sync.Once`, `WaitGroup`, `singleflight` or `errgroup`.
- Read [pipelines.md](references/pipelines.md) when building fan-out/fan-in, worker pools or generator chains — or deciding whether an in-memory transform needs goroutines at all (Go 1.23+ iterators, `samber/ro`).

## Cross-References

- → See `samber/cc-skills-golang@golang-performance` skill for false sharing, cache-line padding, `sync.Pool` hot-path patterns
- → See `samber/cc-skills-golang@golang-context` skill for cancellation propagation and timeout patterns
- → See `samber/cc-skills-golang@golang-safety` skill for concurrent map access and race condition prevention
- → See `samber/cc-skills-golang@golang-troubleshooting` skill for debugging goroutine leaks and deadlocks, including the Go 1.27+ goroutine leak profile and production stack dumps
- → See `samber/cc-skills-golang@golang-design-patterns` skill for graceful shutdown patterns
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for automated AI-driven code review in CI using these guidelines

## References

- [Go Concurrency Patterns: Pipelines](https://go.dev/blog/pipelines)
- [Effective Go: Concurrency](https://go.dev/doc/effective_go#concurrency)
