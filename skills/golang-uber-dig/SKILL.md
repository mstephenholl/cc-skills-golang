---
name: golang-uber-dig
description: "Runtime dependency injection in Golang with uber-go/dig — Provide/Invoke, dig.In/dig.Out parameter objects, named values, value groups, and scopes. Apply when using or adopting uber-go/dig, or when the codebase imports `go.uber.org/dig`. For lifecycle and modules → See `samber/cc-skills-golang@golang-uber-fx` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "⛏"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.19.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go architect wiring an application graph with dig. You want wiring mistakes to fail at startup or in CI, never at first request.

# Using uber-go/dig for Dependency Injection in Go

Reflection-based DI toolkit, designed to power application frameworks (it is the engine behind `uber-go/fx`) and resolve object graphs during startup.

**Official Resources:**

- [pkg.go.dev/go.uber.org/dig](https://pkg.go.dev/go.uber.org/dig)
- [github.com/uber-go/dig](https://github.com/uber-go/dig)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get go.uber.org/dig
```

## dig or fx

fx is built on dig — `Provide`, `Invoke`, `In`/`Out` structs, named values and value groups are the same engine, so wiring transfers unchanged.

**Choose dig** when you need the wiring graph only: CLI tools, libraries exposing a container to callers, test harnesses, or embedding DI into an existing app that manages its own lifecycle.

**Choose fx** for long-running services (HTTP servers, workers, daemons) — dig has no lifecycle, while fx adds `fx.Lifecycle` OnStart/OnStop hooks, modules, and an `app.Run()` that blocks on SIGINT/SIGTERM, so you don't hand-roll signal handling and start/stop ordering. → See `samber/cc-skills-golang@golang-uber-fx` skill.

## Container, Provide and Invoke

```go
c := dig.New() // options: dig.DeferAcyclicVerification(), dig.RecoverFromPanics(), dig.DryRun(true)

must(c.Provide(func(cfg *Config) (*sql.DB, error) { // lazy: runs only when its output is needed
    return sql.Open("postgres", cfg.DSN)
}))

err := c.Invoke(func(db *sql.DB) error { return db.Ping() })

func must(err error) { if err != nil { panic(err) } }
```

Constructors are **lazy** and **memoized**: each output type is built once and shared (singleton per container). Return an `error` rather than panicking — `Invoke` wraps it with the dependency path that triggered it, which names the failure point. `dig.RecoverFromPanics()` turns panics from third-party constructors into a typed `dig.PanicError` (detect with `errors.As`).

## Parameter Objects with `dig.In`

Once a constructor has 4+ dependencies, embed `dig.In` and take the struct as the single argument — adding a dependency becomes a one-line field instead of a signature break. A plain struct without `dig.In` is treated as one dependency and never filled.

```go
type HandlerParams struct {
    dig.In

    Logger *zap.Logger
    DB     *sql.DB
    Cache  *redis.Client  `optional:"true"` // zero value if not provided
    DBRO   *sql.DB        `name:"readonly"` // named dependency
    Routes []http.Handler `group:"routes"`  // value group
}

func NewHandler(p HandlerParams) *Handler { /* ... */ }
```

## Result Objects with `dig.Out`

Return several values from one constructor and attach `name`/`group` tags to results:

```go
type ConnResult struct {
    dig.Out

    ReadWrite *sql.DB `name:"primary"`
    ReadOnly  *sql.DB `name:"readonly"`
}

func NewConnections(cfg *Config) (ConnResult, error) { /* ... */ }
```

## Named Values

dig rejects two unnamed providers of the same type at `Provide` time. Name them — no wrapper types needed — and consume with `name:"primary"` / `name:"readonly"` on `dig.In` fields:

```go
c.Provide(NewPrimaryDB,  dig.Name("primary"))
c.Provide(NewReadOnlyDB, dig.Name("readonly"))
```

## Value Groups

Many providers, one consumer slice — typical for HTTP handlers, health checks, migrations — so no one assembles the slice by hand:

```go
type RouteResult struct {
    dig.Out
    Handler http.Handler `group:"routes"`
}

func NewUserHandler(db *sql.DB) RouteResult { /* ... */ }

type ServerParams struct {
    dig.In
    Routes []http.Handler `group:"routes"`
}
```

Append `,flatten` (`group:"routes,flatten"`) to a slice-typed result to contribute its elements instead of nesting the slice. Group order is **not guaranteed**; when order matters (middleware chain, migration sequence), provide an explicit ordered slice from a single constructor.

## Provide as Interface (`dig.As`)

Register a concrete constructor and expose it only under interfaces, without an adapter constructor:

```go
c.Provide(NewPostgresDB, dig.As(new(Database), new(io.Closer)))
// Consumers ask for Database or io.Closer; *PostgresDB stays out of the graph.
```

## Organizing and Validating the Graph

- **Keep the container in `main()`** — injecting `*dig.Container` into services turns it into a service locator: dependencies disappear from constructor signatures, and tests must build a real container.
- **Group registration by module** — one file per module that calls `c.Provide` for its types, so a module can later become an `fx.Module` without rewriting wiring.
- **Validate the production graph in CI** — build the container with `dig.DryRun(true)`, register the same Provides, and `Invoke` the composition root; missing providers and cycles surface without running constructors.
- **Keep `init()` empty** — side effects start only inside constructors, after the graph is built.

## Testing

dig containers are cheap — build a fresh one per test, swap one dependency with `Decorate`, and call `Invoke` to drive the system. Read [references/testing.md](references/testing.md) for per-test wiring, shared helpers, CI graph validation, and asserting wire-time errors or recovered panics.

## Further Reading

- [references/advanced.md](references/advanced.md) — read for `Decorate`, child scopes for request-scoped values, optional deps, error helpers, `Visualize`, and the Quick Reference
- [references/recipes.md](references/recipes.md) — read for end-to-end examples: HTTP server with a route group, two databases, request scopes, decorators, dry-run validation

## Cross-References

- → See `samber/cc-skills-golang@golang-uber-fx` skill for application lifecycle, modules, and signal-aware Run() built on top of dig
- → See `samber/cc-skills-golang@golang-dependency-injection` skill for DI concepts and library comparison
- → See `samber/cc-skills-golang@golang-samber-do` skill for a generics-based alternative without reflection
- → See `samber/cc-skills-golang@golang-google-wire` skill for compile-time DI (no runtime container)
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for interface design patterns
- → See `samber/cc-skills-golang@golang-testing` skill for general testing patterns

If you encounter a bug or unexpected behavior in uber-go/dig, open an issue at <https://github.com/uber-go/dig/issues>.
