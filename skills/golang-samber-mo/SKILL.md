---
name: golang-samber-mo
description: "Monadic types for Golang with samber/mo — Option, Result, Either, Future, and typed pipelines. Apply when using or adopting samber/mo, or when the codebase imports `github.com/samber/mo`. Not for nil-safety without this library (→ See `samber/cc-skills-golang@golang-safety` skill) or native error wrapping (→ See `samber/cc-skills-golang@golang-error-handling` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.1.6"
  openclaw:
    emoji: "🎭"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.16.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer bringing functional programming safety to Go. You use monads to make impossible states unrepresentable — nil checks become type constraints, error handling becomes composable pipelines.

# samber/mo — Monads and Functional Abstractions for Go

Go 1.18+ library of type-safe monadic types, inspired by Scala, Rust, and fp-ts.

**Official Resources:**

- [pkg.go.dev/github.com/samber/mo](https://pkg.go.dev/github.com/samber/mo)
- [github.com/samber/mo](https://github.com/samber/mo)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get github.com/samber/mo
```

## Choose the Type

| Type | Use for | Note |
| --- | --- | --- |
| `Option[T]` | A value that may be absent, where absence differs from the zero value | Maps to JSON `null` and SQL `NULL` |
| `Result[T]` | A step that may fail, chained with other fallible steps | Equivalent to `Either[error, T]`; `ToEither()` maps Ok → Right, Err → Left |
| `Either[L, R]` | One of two valid alternatives, neither an error (cached vs fresh user) | Build with `mo.Left` / `mo.Right`, branch with `Match` |
| `Either3`–`Either5` | One of 3–5 types | `mo.NewEither3Arg1` … constructors, one `Match` handler per type |
| `Future[T]` | An async value | Eager — starts running when constructed |
| `Task[T]` / `IO[T]` | A deferred async / sync side effect | Lazy — runs only on `Run()`; `Task.Run()` returns a `*Future[T]` |
| `State[S, A]` | A computation that threads state (parser position) | `Run(initial)` returns `(result, newState)` |

- **Compose `IO`/`IOEither` by wrapping them** — they have no `Map`/`FlatMap`, so build a larger program as a new `mo.NewIOEither(func() (R, error) { … })` whose function runs the smaller ones; the composed value stays lazy, so constructing it performs no I/O and tests can build it with stubbed effects.
- **Reach for mo when steps chain** — for a single fallible call with no follow-up, plain `if err != nil` is clearer than a `Result`.
- **Use `Option` only when absence carries meaning** — a `count` where `0` means "no items" stays `int`; a `nickname` where "not set" differs from `""` becomes `Option[string]`.
- **Keep `(T, error)` in exported signatures** — convert with `mo.TupleToResult(f())` on entry and `.Get()` on exit, so callers never need mo to call your API.
- **Keep `MustGet` inside `mo.Do`** — `Do` turns the panic into an `Err`; elsewhere `MustGet` crashes on `None`/`Err`. `Do` recovers every panic, not only `MustGet`'s, so a nil dereference inside it silently becomes an `Err`. Reach for it when combining several Options and Results — the block reads as straight-line code instead of nested `FlatMap` chains.

## Methods Cannot Change the Type Parameter

Go methods cannot introduce new type parameters, so the transform methods (`Map`, `FlatMap`, Either's `MapLeft`/`MapRight`) keep the same type parameters:

- `Option.Map` takes `func(T) (T, bool)` — returning `false` turns `Some` into `None`, so it doubles as a filter.
- `Result.Map` takes `func(T) (T, error)`; `MapValue` takes `func(T) T` for infallible steps; `FlatMap` takes `func(T) Result[T]` for steps that already return a `Result`.
- When a step changes the type, use the curried functions in the `option`, `result` and `either` sub-packages — `option.Map(strconv.Itoa)(opt)` — and `PipeN` to chain several:

```go
import "github.com/samber/mo/result"

// []byte -> Config -> ValidConfig: each step changes the type
parsed := result.Pipe2(
    mo.TupleToResult(os.ReadFile("config.yaml")),
    result.Map(func(data []byte) Config { return parseConfig(data) }),
    result.FlatMap(func(cfg Config) mo.Result[ValidConfig] { return validate(cfg) }),
)
```

- To collapse any Option, Result or Either into another type in one call, use `mo.Fold(x, onSuccess, onFailure)`.

## Common Mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| `mo.TupleToOption(m[key])` | Inside a call argument a map index yields one value, so it does not compile | `v, ok := m[key]; opt := mo.TupleToOption(v, ok)` |
| `if s == "" { return mo.None[string]() }` to build an Option | Hand-rolls what the library already does | `mo.EmptyableToOption(s)` — None for any type's zero value (`""`, `0`, nil); it checks via reflection, so on a hot path use `mo.TupleToOption(s, s != "")`, still without an `if` |
| `json:"x,omitempty"` on an `Option` field | `omitempty` ignores struct types; `None` still marshals as `null` | `omitzero` (Go 1.24+), which uses `Option.IsZero` |
| `mo.Try` around a call that can panic | `Try` only converts the returned error; the panic propagates | Call it inside `mo.Do`, which recovers panics into `Err` |
| Two structs (DB row and JSON response) for nullable columns | `Option` implements `sql.Scanner`, `driver.Valuer` and `json.Marshaler`/`Unmarshaler` | One struct with `mo.Option[T]` fields |

## References

- [references/option.md](references/option.md) — when building Options from pointers, zero values or lookups (`PointerToOption`, `EmptyableToOption`, `TupleToOption`), or checking Option methods and encodings.
- [references/result.md](references/result.md) — when wrapping `(T, error)` calls, choosing among `Map`, `MapValue`, `FlatMap` and `MapErr`, or serializing a Result.
- [references/either.md](references/either.md) — when modelling Either or Either3–5 values.
- [references/pipelines.md](references/pipelines.md) — when a step changes the type: sub-package `Map`/`FlatMap`/`Match` signatures, `PipeN`, and `Fold`.
- [references/advanced-types.md](references/advanced-types.md) — when using Future, IO, IOEither, Task, TaskEither or State.
- [references/monads-guide.md](references/monads-guide.md) — when explaining monads to someone new to them, or arguing for or against adopting mo.

If you encounter a bug or unexpected behavior in samber/mo, open an issue at <https://github.com/samber/mo/issues>.

## Cross-References

- → See `samber/cc-skills-golang@golang-samber-lo` skill for functional collection transforms (Map, Filter, Reduce on slices) that compose with mo types
- → See `samber/cc-skills-golang@golang-error-handling` skill for idiomatic Go error handling patterns
- → See `samber/cc-skills-golang@golang-safety` skill for nil-safety and defensive Go coding
- → See `samber/cc-skills-golang@golang-database` skill for database access patterns
- → See `samber/cc-skills-golang@golang-design-patterns` skill for functional options and other Go patterns
