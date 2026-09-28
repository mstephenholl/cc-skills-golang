---
name: golang-dependency-injection
description: "Golang dependency injection choices — manual constructor injection vs a container (google/wire, uber-go/dig, uber-go/fx, samber/do). Use when deciding whether or which DI library to adopt, or when replacing globals, init() wiring, or service locators with injection. For a chosen library's API → See `samber/cc-skills-golang@golang-google-wire`, `samber/cc-skills-golang@golang-uber-dig`, `samber/cc-skills-golang@golang-uber-fx`, or `samber/cc-skills-golang@golang-samber-do` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.5"
  openclaw:
    emoji: "🔌"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go software architect. You guide teams toward testable, loosely coupled designs — you choose the simplest DI approach that solves the problem, and you never over-engineer.

**Modes:**

- **Design** (new project, new service, or adding a service to an existing DI setup): follow the existing DI setup if there is one; otherwise pick manual injection or a library from the decision table, then write the wiring. Done when the graph builds and each service receives its dependencies through its constructor.
- **Refactor** (existing coupled code): find `init()`/globals, concrete deps tests can't swap, and container-passing service locators. Produce a staged plan and, when asked to refactor, execute it stage by stage.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-dependency-injection` skill takes precedence.

# Dependency Injection in Go

This skill is not exhaustive. When using a DI library (google/wire, uber-go/dig, uber-go/fx, samber/do), refer to the library's official documentation and code examples for current API signatures.

## Rules

1. **Inject through constructors, not package-level vars or `init()`** — a global `db` is a hidden dependency that tests can't swap and that couples every importer to one instance. Declare interfaces where they're consumed and return concrete structs (→ See `samber/cc-skills-golang@golang-structs-interfaces` skill).
2. **The container lives only at the composition root** (`main()` or app startup) — passing the injector into services makes it a service locator: dependencies vanish from constructor signatures and missing ones fail at call time instead of at wiring time. Resolve dependencies inside providers and hand services plain constructor arguments:

   ```go
   do.Provide(injector, func(i do.Injector) (*UserService, error) {
       return NewUserService(do.MustInvoke[*Database](i), do.MustInvoke[Mailer](i)), nil
   })
   ```

3. **One container per application; scopes for request-level isolation** — building a container per request re-creates every singleton on each call, so the DB pool and caches are never reused.
4. **Singletons for stateful services, transients for stateless ones** — DB pools and cache clients must be shared; stateless request processors can be built fresh through a factory. Prefer lazy registration so unused services are never constructed and startup stays fast — samber/do and dig/fx build on first use, while wire builds everything eagerly.
5. **Keep the dependency graph shallow** — in a chain like `Repo → UserService → NotificationService → OrderService → PaymentService`, every change ripples down the chain and every test must build all of it. Have most services depend on repositories and config directly, not reach them through other services.

## Manual Injection First

Default to manual constructor injection, wired in `main()` in layer order (config → infrastructure → repositories → services → transport) — it is visible and compile-checked. Read [references/manual-di.md](references/manual-di.md) when writing that wiring by hand. Manual DI breaks down when:

- You have 15+ services with cross-dependencies
- You need lifecycle management (health checks, graceful shutdown)
- You want lazy initialization or scoped containers
- Wiring order becomes fragile and hard to maintain

| Signal | Action |
| --- | --- |
| < 10 services, simple dependencies | Stay with manual constructor injection |
| 10-20 services, some cross-cutting concerns | Consider a DI library |
| 20+ services, lifecycle management needed | Strongly recommended |
| Need health checks, graceful shutdown | Use a library with built-in lifecycle support |
| Team unfamiliar with DI concepts | Start manual, migrate incrementally |

## Decision Table

| Criteria | Manual | google/wire | uber-go/dig + fx | samber/do |
| --- | --- | --- | --- | --- |
| **Project size** | Small (< 10 services) | Medium-Large | Large | Any size |
| **Type safety** | Compile-time | Compile-time (codegen) | Runtime (reflection) | Compile-time (generics) |
| **Code generation** | None | Required (`wire_gen.go`) | None | None |
| **Reflection** | None | None | Yes | None |
| **API style** | N/A | Provider sets + build tags | Struct tags + decorators | Simple, generic functions |
| **Lazy loading** | Manual | N/A (all eager) | Built-in | Built-in |
| **Singletons** | Manual | Built-in | Built-in | Built-in |
| **Transient/factory** | Manual | Manual | Built-in | Built-in |
| **Scopes/modules** | Manual | Provider sets | Module system (fx) | Built-in (hierarchical) |
| **Health checks** | Manual | Manual | Manual | Built-in interface |
| **Graceful shutdown** | Manual | Manual | Built-in (fx) | Built-in interface |
| **Container cloning** | N/A | N/A | N/A | Built-in |
| **Debugging** | Print statements | Compile errors | `fx.Visualize()` | `ExplainInjector()`, web interface |
| **Go version** | Any | Any | Any | 1.18+ (generics) |
| **Learning curve** | None | Medium | High | Low |

Once a library is chosen, its skill owns the API:

| Library | → See | What sets it apart |
| --- | --- | --- |
| google/wire | `samber/cc-skills-golang@golang-google-wire` skill | Injector files carry `//go:build wireinject` and list every provider in `wire.Build(...)`; `wire.Bind` maps interfaces; `wire_gen.go` is generated as plain constructor calls — regenerate it, never edit it |
| uber-go/dig | `samber/cc-skills-golang@golang-uber-dig` skill | Reflection container without lifecycle; errors surface at `Invoke`, not compile time |
| uber-go/fx | `samber/cc-skills-golang@golang-uber-fx` skill | Lifecycle through an `lc fx.Lifecycle` parameter injected into the provider, then `lc.Append(fx.Hook{OnStart, OnStop})`, both taking a `context.Context` |
| samber/do | `samber/cc-skills-golang@golang-samber-do` skill | Generic providers, scopes, health checks, shutdown, and clone-and-override testing |

## Testing

Test a service by passing a hand-written fake of its consumer-side interface to the constructor — declare the fake in the `_test.go` file, with no container and no database. For samber/do integration tests that keep most services real, → See `samber/cc-skills-golang@golang-samber-do` skill (clone the container, override one boundary).

## Cross-References

- → See `samber/cc-skills-golang@golang-testing` skill for test structure and fakes
- → See `samber/cc-skills-golang@golang-project-layout` skill for where the composition root lives

## References

- [samber/do/v2 documentation](https://do.samber.dev) | [github.com/samber/do/v2](https://github.com/samber/do)
- [google/wire user guide](https://github.com/google/wire/blob/main/docs/guide.md)
- [uber-go/fx documentation](https://uber-go.github.io/fx/)
- [uber-go/dig](https://github.com/uber-go/dig)
