---
name: golang-naming
description: "Golang naming conventions — packages, constructors, interfaces, enums, errors, receivers, getters, acronyms, and test names. Use when choosing or reviewing an identifier, package, or error name, or settling a convention debate (New vs NewTypeName, Get prefixes, ALL_CAPS constants, ErrNotFound vs NotFoundError, error-string casing). Not for implementation questions that don't involve a naming decision."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "🏷"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
---

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-naming` skill takes precedence.

# Go Naming Conventions

> "Clear is better than clever." — Go Proverbs
>
> "Design the architecture, name the components, document the details." — Go Proverbs

To ignore a rule, just add a comment to the code.

## MixedCaps

Use `MixedCaps` for every identifier, constants included (`MaxPacketSize`, not `MAX_PACKET_SIZE` or `kMaxBufferSize`) — capitalization is Go's export mechanism, not emphasis, and tooling assumes it throughout. Underscores belong only in test subcase functions (`TestFoo_InvalidInput`), generated code, and OS/cgo interop.

## Avoid Stuttering

Go call sites always include the package name, so repeating it in the identifier wastes the reader's time — `http.HTTPClient` forces parsing "HTTP" twice. A name MUST NOT repeat information already present in the package name, type name, or surrounding context.

```go
// Good — clean at the call site
http.Client       // not http.HTTPClient
json.Decoder      // not json.JSONDecoder
user.New()        // not user.NewUser()
config.Parse()    // not config.ParseConfig()

// In package sqldb:
type Connection struct{}  // not DBConnection — "db" is already in the package name

// Anti-stutter applies to ALL exported types, not just the primary struct:
// In package dbpool:
type Pool struct{}        // not DBPool
type Status struct{}      // not PoolStatus — callers write dbpool.Status
type Option func(*Pool)   // not PoolOption
```

## Frequently Missed Conventions

These conventions are correct but non-obvious — they are the most common source of naming mistakes:

**Constructor naming:** When a package exports a single primary type, the constructor is `New()`, not `NewTypeName()`. This avoids stuttering — callers write `apiclient.New()` not `apiclient.NewClient()`. Use `NewTypeName()` only when a package has multiple constructible types (like `http.NewRequest`, `http.NewServeMux`).

**Boolean struct fields:** Unexported boolean fields MUST use `is`/`has`/`can` prefix — `isConnected`, `hasPermission`, not bare `connected` or `permission`. The exported getter keeps the prefix: `IsConnected() bool`. This reads naturally as a question and distinguishes booleans from other types.

**Error strings are fully lowercase — including acronyms.** Write `"invalid message id"` not `"invalid message ID"`, because error strings are often concatenated with other context (`fmt.Errorf("parsing token: %w", err)`) and mixed case looks wrong mid-sentence. Sentinel errors should include the package name as prefix: `errors.New("apiclient: not found")`.

**Enum zero values:** Prefix values with the type name (`StatusReady`) and place an explicit `Unknown`/`Invalid` sentinel at iota position 0, or start at `iota + 1`. A `var s Status` silently becomes 0 — if that maps to a real state like `StatusReady`, code can behave as if a status was deliberately chosen when it wasn't.

**Subtest names:** Table-driven test case names in `t.Run()` should be fully lowercase descriptive phrases: `"valid id"`, `"empty input"` — not `"valid ID"` or `"Valid Input"`.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| `GetName()` getter | Go omits `Get` because `user.Name()` reads naturally at call sites. But `Is`/`Has`/`Can` prefixes are kept for boolean predicates: `IsHealthy() bool` not `Healthy() bool` |
| `Url`, `Http`, `Json` acronyms | Mixed-case acronyms create ambiguity (`HttpsUrl` — is it `Https+Url`?). Use all caps or all lower |
| `this` or `self` receiver | Go methods are called frequently — use 1-2 letter abbreviation (`s` for `Server`) to reduce visual noise |
| Inconsistent receiver names | Switching names across methods of the same type confuses readers — use one name consistently |
| `util`, `helper` packages | These names say nothing about content — use specific names that describe the abstraction |
| Plural, underscored or MixedCaps package names | Go convention is a singular, lowercase single word (`net/url`, not `net/urls` or `url_parser`) — keeps import paths consistent |
| `ErrAPIResponse` error type | The `Err` prefix marks sentinel error variables (`ErrNotFound`); error types take the `Error` suffix (`APIError`, `PathError`) |
| `userSlice` type-in-name | Types encode implementation detail — `users` describes what it holds, not how |
| Long names for short scopes | Name length should match scope — `i` is fine for a 3-line loop, `userIndex` is noise; a package-level `t` is too cryptic |
| Naming constants by value | Values change, roles don't — `DefaultPort` survives a port change, `Port8080` doesn't |
| `FetchCtx()` context variant | `WithContext` is the standard Go suffix — `FetchWithContext()` is instantly recognizable |
| `sort()` in-place but no `In` | Readers assume functions return new values. `SortIn()` signals mutation |
| `parse()` panicking on error | `MustParse()` warns callers that failure panics — surprises belong in the name |
| Mixing `With*`, `Set*`, `Use*` | Consistency across the codebase — `With*` is the Go convention for functional options |
| `Wrapf` without `f` suffix | The `f` suffix signals format-string semantics — `Wrapf`, `Errorf` tell callers to pass format args |
| Unnecessary import aliases | Aliases add cognitive load. Only alias on collision — `mrand "math/rand"` |
| Inconsistent concept names | Using `user`/`account`/`person` for the same concept forces readers to track synonyms — pick one name |

Applying these fixes means renaming existing identifiers — → See `samber/cc-skills-golang@golang-gopls` skill to do it safely: its rename updates every call site across the workspace and refuses a rename that would break interface satisfaction, which a grep/sed or manual find-and-replace rename silently misses.

## References

Read the reference that matches the name you are choosing:

- [references/packages-files.md](./references/packages-files.md) — when naming a package or file, or adding an import alias
- [references/identifiers.md](./references/identifiers.md) — when naming variables, boolean fields, receivers, or identifiers containing acronyms
- [references/functions-methods.md](./references/functions-methods.md) — when naming functions, getters/setters, constructors, functional options, or deciding on named returns
- [references/types-errors.md](./references/types-errors.md) — when naming an interface, struct, constant, enum, or error
- [references/testing.md](./references/testing.md) — when naming tests, subtests, table fields, or test helpers

## Enforce with Linters

Many naming convention issues are caught automatically by linters: `revive`, `predeclared`, `misspell`, `errname`. See `samber/cc-skills-golang@golang-lint` skill for configuration and usage.

## Cross-References

- → See `samber/cc-skills-golang@golang-code-style` skill for broader formatting and style decisions
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for interface naming depth and receiver design
- → See `samber/cc-skills-golang@golang-lint` skill for automated enforcement (revive, predeclared, misspell, errname)
- → See `samber/cc-skills-golang@golang-gopls` skill for safe rename when applying a naming fix
- → See `samber/cc-skills-golang@golang-refactoring` skill for how to apply a rename safely at scale (gopls Rename/Inline, blast-radius mapping, staged PR workflow) once you've decided what to rename identifiers to
