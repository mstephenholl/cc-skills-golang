---
name: golang-design-patterns
description: "Idiomatic Golang design patterns — functional options, constructors, avoiding init() and globals, enums, panic vs error, resource lifecycle, graceful shutdown, retries, and architecture styles (clean, hexagonal, flat). Use when designing a Go API or package, choosing between patterns, or hardening a service for production. Not for DI containers (→ See `samber/cc-skills-golang@golang-dependency-injection` skill) or error wrapping and logging (→ See `samber/cc-skills-golang@golang-error-handling` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.5"
  openclaw:
    emoji: "🏗"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go architect who values simplicity and explicitness. You apply patterns only when they solve a real problem — not to demonstrate sophistication — and you push back on premature abstraction.

**Modes:**

- **Design mode** — creating new APIs, packages, or application structure: favor the smallest pattern that satisfies the requirement and follow the repo's existing structure. Ask only when picking an application-level architecture (clean/hexagonal/DDD) for a repo with no established layout. Done when the design builds against its callers and the rules below hold.
- **Review mode** — auditing existing code for design issues: look for `init()` abuse, implicit global state, unbounded resources, missing timeouts, and dropped close/flush errors on write paths. Rank findings by impact with file:line; if the user asked for fixes, apply them.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-design-patterns` skill takes precedence.

# Go Design Patterns & Idioms

## Rules

1. **Enum zero value is a sentinel** — name `iota` 0 `Unknown` (or skip it) and start real values at 1, because Go's zero value otherwise passes silently as the first business value: an unset `OrderStatus` reads as `Pending`.
2. **Avoid `init()` and mutable globals** — `init()` cannot return errors (it must panic or `log.Fatal`), runs before `main()` and before every test so its side effects make tests unpredictable, and multiple `init()` functions run across files in filename order. Build dependencies in explicit constructors (`NewUserRepository(db)`) called from `main`.
3. **Explicit defaults** — set defaults in the constructor, not through `default:"8080"` struct tags read by reflection, because readers can't see behavior hidden in tags without knowing the library.
4. **`runtime.AddCleanup` over `runtime.SetFinalizer`** (Go 1.24+) — it allows several cleanups per object, runs even when the object sits in a cycle, and the cleanup receives a copy of a value rather than the object, so nothing can resurrect it.
5. **Surface close/flush errors on write resources** — `defer f.Close()` is fine for reads, but for files, `bufio.Writer`, gzip and network writers the final `Close`/`Flush` is where buffered data reaches disk or wire; dropping its error reports success on a truncated write.
6. **Retries check the context between attempts** — wait with `select` on `ctx.Done()` and a timer rather than `time.Sleep`, and return `ctx.Err()` once cancelled, because a sleeping retry loop keeps working for a caller that already gave up.
7. **Panic is for bugs, not expected errors** — return errors for anything a caller can handle (bad input, a missing config field, I/O failure); panic on violated invariants (nil where the contract forbids it) and in `Must*` constructors called at init time with constant input.
8. **Compile regexps once, at package level** with `regexp.MustCompile` — compilation is O(n) in the pattern and allocates, so per-call compilation dominates a hot handler.
9. **Bound everything** — give every external call a timeout and every pool, queue and buffer a maximum size, because a slow upstream or an unbounded queue grows until the process hangs or runs out of memory.
10. **Stream large transfers** — iterate rows (`rows.Next()` or an `iter.Seq2`) and encode each record straight to the writer instead of collecting millions into a slice, which keeps memory constant instead of risking OOM.

## Constructors: functional options vs builder

Default to functional options when optional config will grow — each option is one `With*` function, so adding one never breaks callers. Use a builder only when configuration steps must validate against each other. When any option can fail validation, give the option type an error return so bad config fails at construction, not at first use:

```go
type Option func(*Server) error

func WithMaxConns(n int) Option {
    return func(s *Server) error {
        if n <= 0 {
            return fmt.Errorf("max conns must be positive, got %d", n)
        }
        s.maxConns = n
        return nil
    }
}

func NewServer(addr string, opts ...Option) (*Server, error) {
    s := &Server{addr: addr, readTimeout: 5 * time.Second, maxConns: 100} // defaults before options
    for _, opt := range opts {
        if err := opt(s); err != nil {
            return nil, err
        }
    }
    return s, nil
}
```

## Architecture

Match the architecture to the project's size — a 200-line CLI is a flat `main.go` plus a few files, with no layers and no DI framework, and complexity gets added when the code demands it. Principles that hold regardless of style:

- **Keep the domain pure** — no framework or infrastructure imports in the domain layer, so business rules test without a database.
- **Validate at boundaries, trust inside** — check input in handlers, CLI parsing and consumers; re-validating the same data in service and repository layers clutters code without adding safety. Business rules still belong in the domain.
- **Make illegal states unrepresentable** — a type with unexported fields and a validating constructor (`NewEmail`) means invalid values cannot reach the functions that accept it.

| Read | When |
| --- | --- |
| [references/architecture.md](references/architecture.md) | Choosing a structure for a new project or sizing one to its scope |
| [references/clean-architecture.md](references/clean-architecture.md) | Applying the dependency rule, use cases and layered adapters |
| [references/hexagonal-architecture.md](references/hexagonal-architecture.md) | A service with several entry points (HTTP, gRPC, consumers) or several driven systems |
| [references/ddd.md](references/ddd.md) | Modeling aggregates, value objects such as money, domain repositories, or bounded-context communication |
| [references/resource-management.md](references/resource-management.md) | Writing a resource pool, GC-driven cleanup of native handles, or graceful shutdown |
| [references/data-handling.md](references/data-handling.md) | Streaming a large result set or response with iterators |

## Cross-References

- → See `samber/cc-skills-golang@golang-error-handling` skill for error wrapping, sentinel errors, and the single handling rule
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for interface design and composition
- → See `samber/cc-skills-golang@golang-context` skill for timeout and cancellation patterns
- → See `samber/cc-skills-golang@golang-concurrency` skill for goroutine lifecycle
- → See `samber/cc-skills-golang@golang-database` skill for queries, transactions, and connection pools
- → See `samber/cc-skills-golang@golang-data-structures` skill for data structure selection and container/ packages
- → See `samber/cc-skills-golang@golang-project-layout` skill for directory structure and 12-factor conventions
- → See `samber/cc-skills-golang@golang-refactoring` skill for staging a migration toward one of these patterns (options struct, DI, consumer-side interfaces) across an existing codebase
