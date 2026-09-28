---
name: golang-structs-interfaces
description: "Golang struct and interface design — interface size and placement, embedding vs named fields, pointer vs value receivers, type assertions and switches, and struct tags (omitempty/omitzero). Use when deciding where an interface belongs or how small it should be, choosing receivers, embedding, or debugging serialization tags. Not for package or service architecture (→ See `samber/cc-skills-golang@golang-design-patterns` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.3"
  openclaw:
    emoji: "🧩"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go type system designer. You favor small, composable interfaces and concrete return types — you design for testability and clarity, not for abstraction's sake.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-structs-interfaces` skill takes precedence.

# Go Structs & Interfaces

## Interface Design

Keep interfaces to 1-3 methods and compose larger contracts from them, because every extra method is one more thing each implementation and mock must provide. Honor canonical signatures — a `String()` method must match `fmt.Stringer`, not `ToString()` — so your types plug into the stdlib. → See `samber/cc-skills-golang@golang-naming` skill for interface naming.

### Define interfaces where they're consumed

Default to declaring the interface in the consumer package, listing only the methods that consumer calls — the consumer stays in control of the contract and never imports a package just for its interface, while the provider exports a concrete type:

```go
// package notification — declares only what it needs; package email exports a concrete Client
type Sender interface {
    Send(to, body string) error
}

type Service struct {
    sender Sender
}
```

The exception is a contract many independent implementations plug into (`io.Reader`, `database/sql/driver.Driver`, a plugin API): the package that owns the extension point defines it.

### Accept interfaces, return structs

Accept interface parameters and return concrete types: `func NewService(store UserStore) *Service`, not `... ServiceInterface`. An interface return hides every other field and method of the concrete type from callers, who can still assign the result to an interface variable themselves.

### Don't create interfaces prematurely

> "Don't design with interfaces, discover them."

An interface written before a second implementation exists is a guess about which methods will vary — usually wrong, so it gets reshaped anyway — and meanwhile it adds indirection that hides the concrete type from readers and tooling. Start with a concrete `UserRepository struct`; extract an interface once a second consumer, a second implementation, or a test double demands it. Testability is a legitimate trigger, but make it a deliberate choice, not a reflex because the type is "a repository".

## Make the Zero Value Useful

Design structs so `var x T` works without a constructor, as `bytes.Buffer` and `sync.Mutex` do — callers who skip `NewT()` otherwise hit nil-map panics. Guard lazily initialized fields in the methods that write them:

```go
func (r *Registry) Register(name string, item Item) {
    if r.items == nil { // zero-value Registry is usable
        r.items = make(map[string]Item)
    }
    r.items[name] = item
}
```

## Generics over `any`

Default to a type parameter (`func Contains[T comparable](s []T, v T) bool`) over `any` parameters, because generics keep type safety that `any` pushes to runtime assertions. Use `any` where the values are genuinely heterogeneous or unknown until runtime — JSON decoding, reflection, a container of mixed types.

## Compile-Time Interface Check

Place `var _ io.ReadWriter = (*MyBuffer)(nil)` next to the type definition — it costs nothing at runtime, and the build fails the moment `MyBuffer` stops satisfying the interface instead of at a distant call site.

## Type Assertions & Type Switches

Use the comma-ok form (`s, ok := val.(string)`) and handle `!ok` with an error — the single-value form panics on mismatch, and "controlled callers" stop being controlled once events are deserialized from outside or a test fixture is wrong. To exploit a richer implementation without widening the parameter type, assert to a small optional interface (`if f, ok := w.(Flusher); ok`).

Read [references/type-assertions.md](references/type-assertions.md) when writing a type switch (case ordering, `nil` cases) or the optional-behavior pattern.

## Embedding vs Named Field

Embedding promotes _all_ of the inner type's methods and fields to the outer type — composition, not inheritance. The receiver of a promoted method is still the _inner_ value, so it cannot see the outer struct's fields; the outer type overrides by declaring a method with the same name.

| Use | When |
| --- | --- |
| **Embed** | The outer type should expose the inner type's full API ("is a" enhanced version), e.g. embedding `http.Handler` to promote `ServeHTTP` |
| **Named field** | The inner type is an internal dependency ("has a") whose methods must not leak to callers — delegate the few you need explicitly |

## Pointer vs Value Receivers

| Use pointer `(s *Server)` | Use value `(s Server)` |
| --- | --- |
| Method modifies the receiver | Receiver is small and immutable |
| Receiver contains `sync.Mutex` or similar | Receiver is a basic type (int, string) |
| Receiver is a large struct | Map, func, or chan types (already references) |

Once one method needs a pointer receiver, make them all pointers — the method set of `T` excludes pointer-receiver methods, so with a mixed set `T` and `*T` satisfy different interfaces and callers discover it at assignment time.

## Struct Field Tags

Tag every exported field of a serialized struct (`json:"created_at"`) — without a tag the encoder uses the Go field name, so a rename silently changes the wire format. Read [references/struct-fields.md](references/struct-fields.md) when choosing `omitempty` vs `omitzero` (an explicit zero becomes indistinguishable from absent) or looking up a tag directive.

## Preventing Struct Copies with `noCopy`

A struct holding a mutex, a channel, or internal pointers breaks when copied: the copy duplicates the lock state, so two goroutines guard two different mutexes and the invariant disappears silently. Embed a `noCopy` sentinel so `go vet` reports every value copy, and pass such structs by pointer — read [references/struct-fields.md](references/struct-fields.md) for the implementation.

**Diagnose:** 1- `go vet ./...` — `copylocks` reports value copies of lock-bearing structs

## Cross-References

- → See `samber/cc-skills-golang@golang-design-patterns` skill for functional options, constructors, and builder patterns
- → See `samber/cc-skills-golang@golang-dependency-injection` skill for wiring interface-typed dependencies
- → See `samber/cc-skills-golang@golang-code-style` skill for value vs pointer function parameters (distinct from receivers)
- → See `samber/cc-skills-golang@golang-gopls` skill for safe rename and the `implementInterface` code action — renaming a method or receiver that participates in interface satisfaction updates every call site and refuses a rename that would silently break the interface, which grep/sed cannot detect
