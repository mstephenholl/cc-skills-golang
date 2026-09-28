---
name: golang-samber-do
description: "Dependency injection in Golang using samber/do — service containers, lifecycle management, scopes, health checks, graceful shutdown, and module organization. Apply when using or adopting samber/do, when the codebase imports github.com/samber/do or github.com/samber/do/v2, or when moving manual constructor injection into a samber/do container."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.6"
  openclaw:
    emoji: "💉"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "2.0.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go architect setting up dependency injection with samber/do. You keep the container at the composition root and let provider errors surface there.

# Using samber/do for Dependency Injection in Go

Type-safe dependency injection toolkit for Go based on Go 1.18+ generics.

**Official Resources:**

- [pkg.go.dev/github.com/samber/do/v2](https://pkg.go.dev/github.com/samber/do/v2)
- [do.samber.dev](https://do.samber.dev)
- [github.com/samber/do/v2](https://github.com/samber/do)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

Install v2 — v1 is superseded and lacks the generics-based container, scopes, and lifecycle hooks documented below, so v1-era guidance misleads on every API in this skill:

```bash
go get github.com/samber/do/v2
```

## Registering Services

| Kind | Built | Register with |
| --- | --- | --- |
| **Lazy** (default) | On first invocation, then shared | `do.Provide` / `do.Lazy` in a package |
| **Eager** | Before registration — you pass the built value | `do.ProvideValue` / `do.Eager` in a package |
| **Transient** | On every invocation | `do.ProvideTransient` / `do.Transient` in a package |

Lazy and transient services take a provider, `func(i do.Injector) (T, error)`:

```go
import "github.com/samber/do/v2"

injector := do.New()

do.Provide(injector, func(i do.Injector) (*UserRepository, error) { // lazy
    return NewUserRepository(do.MustInvoke[*sql.DB](i)), nil
})
do.ProvideValue(injector, cfg) // eager: already built
do.ProvideTransient(injector, func(i do.Injector) (*RequestLogger, error) { // fresh per invocation
    return NewRequestLogger(do.MustInvoke[*slog.Logger](i)), nil
})
```

To make a provider-built service fail fast at boot (a database that must be reachable before serving), invoke it once in `main()` right after registration — or build it yourself and register the value.

Return construction errors from providers rather than swallowing them — a provider that returns a half-built service crashes later, far from the cause.

### Use `do.MustInvoke` inside providers

Inside a provider, use `do.MustInvoke` (or `MustInvokeAs`/`MustInvokeNamed`/`MustInvokeStruct`) rather than the error-returning variant:

- A provider already returns `(T, error)`, so propagating a dependency failure with `do.Invoke` costs an extra `if err != nil { return nil, err }` on every call.
- `do.MustInvoke` panics instead, but samber/do correctly catches and recovers that panic at the enclosing `Invoke` call and converts it back into a regular error — this recover happens inside the library itself, not in caller code, so `MustInvoke` is safe to use inside providers.
- The failure still surfaces as an error at the composition root, just without the manual boilerplate in every provider.

Reserve `do.Invoke` for call sites outside the graph that must degrade gracefully instead of crashing.

### Keep the injector at the composition root

Invoke services in `main()` and in providers only. Passing `do.Injector` into a handler or service turns it into a service locator: its dependencies disappear from the constructor signature and a missing one fails at call time. Resolve dependencies in the provider and pass them to a plain constructor.

### Implicit aliasing (preferred)

Register the concrete type — the provider returns `*PostgreSQLDatabase`, not the interface — and invoke it as the interface, with no alias registration:

```go
do.Provide(injector, func(i do.Injector) (*PostgreSQLDatabase, error) {
    return &PostgreSQLDatabase{}, nil
})

db := do.MustInvokeAs[Database](injector) // first registered service implementing Database
```

### Named services

Register several services of the same type under distinct names instead of wrapper types — a second `do.Provide` for the same type panics with "service has already been declared":

```go
do.ProvideNamed(injector, "primary-db", func(i do.Injector) (*sql.DB, error) { /* ... */ })
do.ProvideNamed(injector, "replica-db", func(i do.Injector) (*sql.DB, error) { /* ... */ })

primary := do.MustInvokeNamed[*sql.DB](injector, "primary-db")
```

## Package Organization

Group registrations per layer with `do.Package()` and compose them in `do.New()`:

```go
// infrastructure/package.go
var Package = do.Package(
    do.Lazy(func(i do.Injector) (*postgres.DB, error) {
        cfg := do.MustInvoke[*Config](i)
        return postgres.Connect(cfg.DatabaseURL)
    }),
)

// main.go
func main() {
    injector := do.New(infrastructure.Package, repository.Package, service.Package, transport.Package)

    server := do.MustInvoke[*http.Server](injector)
    go server.ListenAndServe()

    // Blocks until a signal arrives, then shuts down built services that implement a Shutdown method
    injector.ShutdownOnSignalsWithContext(context.Background(), syscall.SIGTERM, os.Interrupt)
}
```

## Going Further

- Read [references/advanced.md](references/advanced.md) when adding per-request or per-module scopes, health checks, `Shutdown` methods, struct injection with `do:""` tags, explicit aliases, or debugging the graph — it also holds the full registration and invocation Quick Reference.
- Read [references/testing.md](references/testing.md) when testing a service against mocks — clone the production container and override one boundary.

## Cross-References

- → See `samber/cc-skills-golang@golang-dependency-injection` skill for DI concepts, comparison, and when to adopt a DI library
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for interface design patterns
- → See `samber/cc-skills-golang@golang-testing` skill for general testing patterns

If you encounter a bug or unexpected behavior in samber/do, open an issue at <https://github.com/samber/do/issues>.
