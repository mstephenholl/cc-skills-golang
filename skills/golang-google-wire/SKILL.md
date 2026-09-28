---
name: golang-google-wire
description: "Compile-time dependency injection in Golang with google/wire — provider sets, interface bindings, cleanup functions, injector files, and generated wire_gen.go. Apply when using or adopting google/wire, when the codebase imports `github.com/google/wire`, or when wire_gen.go is stale or fails to generate. For runtime DI → See `samber/cc-skills-golang@golang-uber-dig` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.1.4"
  openclaw:
    emoji: "🪡"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - wire
    install:
      - kind: go
        package: github.com/google/wire/cmd/wire@latest
        bins: [wire]
    skill-library-version: "0.7.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(wire:*) Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go architect using wire for compile-time DI. You let the code generator catch missing dependencies and treat `wire_gen.go` as committed, generated source.

**Dependencies:**

- wire: `go install github.com/google/wire/cmd/wire@latest`

# Using google/wire for Compile-Time Dependency Injection in Go

Code-generation DI toolkit. Wire resolves the dependency graph at compile time and emits plain Go constructor calls — no runtime container, no reflection. Errors appear when you run `wire ./...`, not at first request.

Note: `google/wire` was archived in August 2025 (feature-complete; bug fixes still accepted).

**Official Resources:** [pkg.go.dev](https://pkg.go.dev/github.com/google/wire) · [github.com/google/wire](https://github.com/google/wire) · [User Guide](https://github.com/google/wire/blob/main/docs/guide.md) · [Best Practices](https://github.com/google/wire/blob/main/docs/best-practices.md)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get -tool github.com/google/wire/cmd/wire@latest
go get github.com/google/wire
```

Wire has no lifecycle hooks and no signal handling — shutdown is whatever cleanup functions and `main()` code you write. For a long-running daemon that needs `OnStart`/`OnStop` hooks → See `samber/cc-skills-golang@golang-uber-fx` skill; for the full library comparison → See `samber/cc-skills-golang@golang-dependency-injection` skill.

## Providers

A provider is any Go function — inputs are dependencies, outputs are provided types. Three return forms:

```go
func NewConfig() *Config                          { return &Config{Addr: ":8080"} }
func NewDB(cfg *Config) (*sql.DB, error)          { return sql.Open("postgres", cfg.DSN) }
func NewRedis(cfg *Config) (*redis.Client, func(), error) {
    c := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
    return c, func() { c.Close() }, nil
}
```

Return `(T, func(), error)` from any provider that owns a resource and let wire chain the cleanups — the generated injector runs them in reverse construction order, and if construction fails midway only the cleanups already built run. Read [references/advanced.md](references/advanced.md) for cleanup-chain details.

## Provider Sets

`wire.NewSet` groups providers for reuse; sets can include other sets. Default to one set per package in its own `wire.go`, with that package's `wire.Bind` declarations inside it, so the injector's `wire.Build` lists a handful of sets instead of dozens of providers:

```go
// service/wire.go
var ServiceSet = wire.NewSet(
    NewUserRepo,
    NewUserService,
    wire.Bind(new(UserStore), new(*UserRepo)),
)
```

Keep library sets backward-compatible — adding a required input or removing an output breaks every downstream injector, so introduce only newly created types in the same release.

## Injectors

The injector file declares the initialization function; wire generates its body into `wire_gen.go`:

```go
//go:build wireinject

package main

func InitApp() (*App, func(), error) {
    wire.Build(infra.InfraSet, service.ServiceSet, NewApp)
    return nil, nil, nil // replaced by codegen
}

// main.go
app, cleanup, err := InitApp()
if err != nil {
    log.Fatal(err)
}
defer cleanup()
```

Pass values built before wiring (parsed flags, a pre-configured client) as injector parameters — `func InitApp(cfg *Config) (...)` — because wire treats parameters as pre-built providers; a provider that returns a global is the pattern this replaces.

## Interface Bindings

Wire never resolves interface satisfaction implicitly: add `wire.Bind(new(UserStore), new(*PostgresUserRepo))` to the set. The explicitness keeps the graph unambiguous, so adding another type that implements the same interface elsewhere cannot silently rebind it.

## Struct Providers and Values

```go
wire.Struct(new(Server), "Logger", "DB")      // inject named fields
wire.Struct(new(Server), "*")                 // inject all fields not tagged `wire:"-"`
wire.Value(Foo{X: 42})                        // constant expression (no fn calls / channels)
wire.InterfaceValue(new(io.Reader), os.Stdin) // interface-typed literal
wire.FieldsOf(new(Config), "DSN", "Addr")     // promote struct fields as graph nodes, no extractor funcs
```

## Codegen Workflow

```bash
wire ./...           # regenerate all injectors in the module
wire check ./...     # validate graph without regenerating (fast CI check)
```

Add `//go:generate go run github.com/google/wire/cmd/wire` to injector files so `go generate ./...` also works, and commit `wire_gen.go` so CI builds without running wire.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Editing `wire_gen.go` to fix a build error | Every `wire ./...` overwrites it. Fix the graph instead — add the missing provider (or `wire.Value`) to a set — and regenerate. |
| Missing `//go:build wireinject` on an injector file | Make it the first line. Without it the stub compiles alongside `wire_gen.go`, and both declare the same function (`redeclared in this block`). |
| Two providers for the same type (two `string` DSNs, two `*sql.DB`) | Wire allows one provider per type. Declare distinct named types (`type PrimaryDSN string`, `type ReplicaDSN string`) or struct wrappers (`type PrimaryDB struct{ *sql.DB }`); an alias (`=`) is the same type to wire. |
| Injecting an interface without `wire.Bind` | Add `wire.Bind(new(MyInterface), new(*MyImpl))` to the provider set. |
| Changing a constructor signature without regenerating | Run `wire ./...` (or `go generate ./...`) before `go build`; a stale `wire_gen.go` fails with argument-count errors inside generated code. |
| Deferring `cleanup()` before checking the error | Generated injectors return a nil cleanup on error; check `err` first, or guard with `if cleanup != nil { defer cleanup() }`. |

## Testing

Wire generates plain Go constructors, so unit tests use manual injection — no container to clone or reset. Read [references/testing.md](references/testing.md) when an integration test must wire the full graph with fakes (test-only sets plus a `wireinject` test injector) or when adding a CI stale check for `wire_gen.go`.

## Further Reading

- [references/advanced.md](references/advanced.md) — read for multiple injectors per package, set nesting, codegen errors and flags, `panic(wire.Build(...))` syntax, or injector arguments
- [references/recipes.md](references/recipes.md) — read when building a full HTTP server graph, prod/dev build variants, a cleanup-heavy graph, or a CLI entry point

## Cross-References

- → See `samber/cc-skills-golang@golang-dependency-injection` skill for DI concepts and library comparison
- → See `samber/cc-skills-golang@golang-uber-dig` skill for runtime reflection-based DI without lifecycle
- → See `samber/cc-skills-golang@golang-uber-fx` skill for runtime DI with lifecycle hooks, modules, and signal-aware Run()
- → See `samber/cc-skills-golang@golang-samber-do` skill for generics-based DI without reflection
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for interface design patterns
- → See `samber/cc-skills-golang@golang-testing` skill for general testing patterns

If you encounter a bug or unexpected behavior in google/wire, open an issue at <https://github.com/google/wire/issues>.
