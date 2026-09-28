---
name: golang-uber-fx
description: "Golang application framework with uber-go/fx — modules, lifecycle hooks, annotations, decorators, and signal-aware startup and shutdown. Apply when using or adopting uber-go/fx, or when the codebase imports `go.uber.org/fx`. For raw DI without lifecycle → See `samber/cc-skills-golang@golang-uber-dig` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "🏭"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.24.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go architect building a long-running service with fx. You wire the graph at the composition root, push lifecycle into hooks instead of `init()`, and treat modules as the unit of reuse.

# Using uber-go/fx for Application Wiring in Go

Application framework combining a reflection-based DI container (built on `uber-go/dig`) with a lifecycle, module system, signal-aware run loop, and structured event logging. For long-running services where boot order, graceful shutdown, and modular composition matter.

**Official Resources:**

- [pkg.go.dev/go.uber.org/fx](https://pkg.go.dev/go.uber.org/fx)
- [uber-go.github.io/fx](https://uber-go.github.io/fx/)
- [github.com/uber-go/fx](https://github.com/uber-go/fx)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get go.uber.org/fx
```

## fx vs. dig

fx is built on top of dig and shares the same reflection-based container engine. The DI primitives (`Provide`, `Invoke`, `In`/`Out` structs, named values, value groups) are identical — `fx.In`/`fx.Out` are re-exports of `dig.In`/`dig.Out`.

What fx adds on top:

| Concern | dig | fx |
| --- | --- | --- |
| DI container | ✅ `dig.New()` | ✅ (embedded) |
| Lifecycle hooks | ❌ | ✅ `fx.Lifecycle` OnStart/OnStop |
| Module system | ❌ | ✅ `fx.Module` with scoped decorators |
| Signal-aware run loop | ❌ | ✅ `app.Run()` blocks on SIGINT/SIGTERM |
| Structured event logging | ❌ | ✅ `fx.WithLogger` / `fxevent` |
| Startup/shutdown timeout | ❌ | ✅ `fx.StartTimeout` / `fx.StopTimeout` |

**Choose fx** for long-running services (HTTP servers, workers, daemons) — lifecycle and signal handling are mandatory there, and modules make large service graphs manageable.

**Choose raw dig** when you need wiring without a framework: one-shot CLI commands, libraries that expose a container to callers, test harnesses, or embedding DI into an existing app that manages its own lifecycle — fx's lifecycle and modules add nothing to a program that builds a graph, runs once, and exits. See `samber/cc-skills-golang@golang-uber-dig` skill.

## The Application

```go
app := fx.New(
    fx.Supply(cfg),                                  // pre-built values (flags, config, secrets)
    fx.Provide(NewLogger, NewDatabase, NewServer),   // lazy constructors
    fx.Invoke(RegisterRoutes, StartMetricsExporter), // always run during Start
)
app.Run() // blocks until SIGINT/SIGTERM, then runs OnStop hooks
```

Boot stages: `fx.New` validates types (constructors do not run); `app.Start(ctx)` runs each `fx.Invoke` and fires OnStart hooks in topological order; main blocks on `app.Done()`; `app.Stop(ctx)` fires OnStop hooks in reverse order. Default timeout is **15 seconds** — override with `fx.StartTimeout` / `fx.StopTimeout`.

- **`fx.Invoke` is the trigger** — a constructor runs only if an Invoke references its type directly or transitively, so an app with no Invoke builds nothing.
- **`fx.Supply` for pre-built values** — wrapping a parsed config in `fx.Provide(func() *Config { return cfg })` adds a no-op constructor and hides that the value already exists.
- **Validate the graph in CI** with `fx.New(...).Err()` — it reports missing providers and cycles without starting anything.

## Lifecycle Hooks

Inject `fx.Lifecycle` and append hooks rather than starting work in constructors or `init()` — Start/Stop ordering follows graph topology, which `init()` goroutines and constructor side effects ignore, leading to races and leaks.

```go
func NewHTTPServer(lc fx.Lifecycle, cfg *Config) *http.Server {
    srv := &http.Server{Addr: cfg.Addr}

    lc.Append(fx.Hook{
        OnStart: func(ctx context.Context) error {
            ln, err := net.Listen("tcp", srv.Addr)
            if err != nil { return err }
            go srv.Serve(ln) // blocking work in a goroutine
            return nil
        },
        OnStop: func(ctx context.Context) error {
            return srv.Shutdown(ctx)
        },
    })
    return srv
}
```

- **OnStart must return quickly** — run blocking work (a server, a queue consumer) in a goroutine and signal it to stop and drain in OnStop; a blocking OnStart hangs the boot and dependent hooks never fire.
- **Respect `ctx.Done()` in hooks** — the context is bounded by `StartTimeout`/`StopTimeout`; a hook that ignores it is reported as a timeout failure while its goroutine keeps running, leaking resources.
- `fx.StartHook` / `fx.StopHook` / `fx.StartStopHook` adapt simpler signatures (no context, no error, or both): `lc.Append(fx.StartStopHook(srv.Start, srv.Stop))`.

## Parameters, Results and Value Groups

`fx.In`/`fx.Out`, the `name`/`optional`/`group` tags, and value groups behave exactly as in dig — a consumer takes `[]http.Handler` tagged `group:"routes"` in an `fx.In` struct, and group order is unspecified. → See `samber/cc-skills-golang@golang-uber-dig` skill.

## fx.Annotate

Prefer `fx.Annotate` over an `fx.Out` wrapper struct to add tags or interface bindings — the constructor stays untouched and reusable outside fx:

```go
fx.Provide(
    fx.Annotate(NewPostgresDB, fx.As(new(Database)), fx.ResultTags(`name:"primary"`)),
    fx.Annotate(NewUserHandler,
        fx.As(new(http.Handler)),
        fx.ResultTags(`group:"routes"`),
    ),
)
```

## fx.Module

`fx.Module` groups providers, invokes, and decorators for one concern (HTTP, DB, metrics) — group by concern, not by layer, and keep `main()` to modules plus `Run()`. Modules **scope decorators** to themselves and their children: a logger renamed in `fx.Module("db", ...)` only appears renamed inside that module, whereas a top-level `fx.Decorate` applies to the whole app.

```go
var DatabaseModule = fx.Module("database",
    fx.Provide(NewConnection, NewUserRepository),
    fx.Decorate(func(log *zap.Logger) *zap.Logger {
        return log.Named("db")
    }),
)

func main() {
    fx.New(fx.Provide(NewConfig, NewLogger), DatabaseModule, HTTPModule).Run()
}
```

## Testing

Use `go.uber.org/fx/fxtest` to integrate fx with `*testing.T` (failures call `t.Fatal`, `RequireStop` registers as `t.Cleanup`). `fx.Populate(&target)` pulls values out of the graph; `fx.Replace` swaps real dependencies for fakes without editing the module. Read [references/testing.md](references/testing.md) when writing an fx test.

## Further Reading

- [references/advanced.md](references/advanced.md) — read for `fx.Supply`/`fx.Replace`/`fx.Decorate` details, optional deps, routing fx events to your logger (`fx.WithLogger`), manual `Start`/`Stop` instead of `Run()`, and the Quick Reference
- [references/recipes.md](references/recipes.md) — read for a full HTTP service with database and metrics, a background worker with graceful drain, multiple implementations of one interface, or embedding fx in a CLI sub-command

## Cross-References

- → See `samber/cc-skills-golang@golang-uber-dig` skill for the underlying container, `dig.In`/`dig.Out`, and DI without lifecycle
- → See `samber/cc-skills-golang@golang-dependency-injection` skill for DI concepts and library comparison
- → See `samber/cc-skills-golang@golang-samber-do` skill for a generics-based alternative without reflection
- → See `samber/cc-skills-golang@golang-google-wire` skill for compile-time DI (no runtime container)
- → See `samber/cc-skills-golang@golang-context` skill for context propagation in OnStart/OnStop hooks
- → See `samber/cc-skills-golang@golang-testing` skill for general testing patterns

If you encounter a bug or unexpected behavior in uber-go/fx, open an issue at <https://github.com/uber-go/fx/issues>.
