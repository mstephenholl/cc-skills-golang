---
name: golang-code-style
description: "Golang code style judgment calls linters don't catch — line breaking, variable declaration form, control-flow shape, and when a comment helps. Use when reviewing Go code for readability, settling a style question, or setting a project's coding standards. Not for naming (→ See `samber/cc-skills-golang@golang-naming` skill), linter configuration (→ See `samber/cc-skills-golang@golang-lint` skill), or doc comments (→ See `samber/cc-skills-golang@golang-documentation` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.3"
  openclaw:
    emoji: "🎨"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
---

**Orchestration mode:** For a style review across a large codebase, fan out parallel sub-agents split by package — every rule here applies file-locally, so packages review independently — and consolidate into one list of findings with file:line, grouped by rule. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-code-style` skill takes precedence.

# Go Code Style

Style rules that require human judgment — linters handle formatting, this skill handles clarity.

> "Clear is better than clever." — Go Proverbs

When ignoring a rule, add a comment to the code.

## Line Length & Breaking

No rigid line limit, but lines beyond ~120 characters MUST be broken. Break at **semantic boundaries**, not arbitrary column counts. Function calls with 4+ arguments MUST use one argument per line — even when the prompt asks for single-line code:

```go
// Good — each argument on its own line, closing paren separate
mux.HandleFunc("/api/users", func(w http.ResponseWriter, r *http.Request) {
    handleUsers(
        w,
        r,
        serviceName,
        cfg,
        logger,
        authMiddleware,
    )
})
```

When a function signature is too long, the real fix is often **fewer parameters** (use an options struct) rather than better line wrapping. For multi-line signatures, put each parameter on its own line.

## Variable Declarations

SHOULD use `:=` for non-zero values, `var` for zero-value initialization. The form signals intent: `var` means "this starts at zero."

```go
var count int              // zero value, set later
name := "default"          // non-zero, := is appropriate
var buf bytes.Buffer       // zero value is ready to use
```

### Slice & Map Initialization

Initialize a slice explicitly (`[]T{}` or `make`) when it can reach an encoder or an API response — a nil slice encodes as JSON `null`, not `[]`, which surprises consumers — and initialize a map before writing to it, since a nil map panics on write. A function whose callers may write to or encode its result returns `map[K]V{}` or an empty slice, not nil — the caller's first write to a nil map panics far from the function that returned it. A local `var s []T` that only accumulates through `append` is idiomatic; `len`, `range` and map lookups are safe on nil.

```go
users := []User{}                       // returned from an API handler: encodes as []
users := make([]User, 0, len(ids))      // preallocate when capacity is known
m := make(map[string]int, len(items))   // preallocate when size is known
```

Do not preallocate speculatively — `make([]T, 0, 1000)` wastes memory when the common case is 10 items.

### Composite Literals

Composite literals MUST use field names — positional fields break when the type adds or reorders fields.

## Control Flow

- **Errors and edge cases MUST be handled first, then return or `continue`** — the happy path stays at the lowest indentation; nested if-else buries the success path in the innermost block and every new check deepens it.
- **Prefer a `switch` over an if-else chain that compares one value, and give it a `default`** that handles or rejects unexpected values — without it, a newly added constant falls through silently.

### Default-Then-Override Instead of `else`

When the `if` body ends with `return`/`break`/`continue`, the `else` MUST be dropped. For assignments, assign the default first, then override it with independent conditions or a tagless `switch` — an else-if chain hides which branch is the default:

```go
// Good — default-then-override with switch (cleanest for mutually exclusive overrides)
level := slog.LevelInfo
switch {
case debug:
    level = slog.LevelDebug
case verbose:
    level = slog.LevelWarn
}

// Bad — else-if chain hides that there's a default
if debug {
    level = slog.LevelDebug
} else if verbose {
    level = slog.LevelWarn
} else {
    level = slog.LevelInfo
}
```

### Complex Conditions

An `if` or `return` condition with 3+ operands MUST be extracted into named booleans that state the business rule (`isOwner`, `isPublicVerified`) — a wall of `||`/`&&` hides the policy it encodes. Exception: keep an expensive check inline and last, so short-circuit evaluation can skip it. For the before/after example and the short-circuit case, read [references/details.md](./references/details.md).

## Function Design

- Functions SHOULD be **short and focused** — one function, one job.
- Functions SHOULD have **≤4 parameters**. Beyond that, use an options struct (see `samber/cc-skills-golang@golang-design-patterns` skill).
- **Parameter order**: `context.Context` first, then inputs, then output destinations.
- Naked returns help in very short functions (1-3 lines) where return values are obvious, but become confusing when readers must scroll to find what's returned — name returns explicitly in longer functions.

```go
func FetchUser(ctx context.Context, id string) (*User, error)
func SendEmail(ctx context.Context, msg EmailMessage) error  // grouped into struct
```

## Value vs Pointer Arguments

Pass small types (`string`, `int`, `bool`, `time.Time`) by value. Use pointers when mutating, for large structs (~128+ bytes), or when nil is meaningful. Read [references/details.md](./references/details.md) when a parameter's pointer-vs-value choice is unclear or performance-sensitive.

## Code Organization Within Files

- **Group related declarations**: type, constructor, methods together
- **Order**: package doc, imports, constants, types, constructors, methods, helpers
- **One primary type per file** when it has significant methods
- **Blank imports** (`_ "pkg"`) register side effects (init functions). Restricting them to `main` and test packages makes side effects visible at the application root, not hidden in library code
- **Dot imports** pollute the namespace and make it impossible to tell where a name comes from — never use in library code
- **Unexport aggressively** — you can always export later; unexporting is a breaking change, and every exported name is a commitment. → See `samber/cc-skills-golang@golang-gopls` skill to unexport safely — its rename updates every call site atomically and refuses the change when lowercasing a method would break interface satisfaction, a breakage grep/sed silently ships.

## String Handling

Use `strconv` for simple conversions (faster), `fmt.Sprintf` for complex formatting. Use `%q` in error messages to make string boundaries visible. Use `strings.Builder` for loops, `+` for simple concatenation.

## Type Conversions

Prefer explicit, narrow conversions. Use generics over `any` when a concrete type will do:

```go
func Contains[T comparable](slice []T, target T) bool  // not []any
```

## Philosophy

- **"A little copying is better than a little dependency"**
- **Use `slices` and `maps` standard packages**; for filter/group-by/chunk, use `github.com/samber/lo`
- **"Reflection is never clear"** — avoid `reflect` unless necessary
- **Don't abstract prematurely** — extract when the pattern is stable

## Enforce with Linters

Many rules are enforced automatically: `gofmt`, `gofumpt`, `goimports`, `gocritic`, `revive`, `wsl_v5`. → See the `samber/cc-skills-golang@golang-lint` skill.

## Cross-References

- → See the `samber/cc-skills-golang@golang-naming` skill for identifier naming conventions
- → See the `samber/cc-skills-golang@golang-structs-interfaces` skill for pointer vs value receivers, interface design
- → See the `samber/cc-skills-golang@golang-design-patterns` skill for functional options, builders, constructors
- → See the `samber/cc-skills-golang@golang-lint` skill for automated formatting enforcement
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for automated AI-driven code review in CI using these guidelines
- → See `samber/cc-skills-golang@golang-refactoring` skill for mechanically applying guard-clause conversion, function extraction, and options-struct migration safely across many call sites once a review surfaces violations at scale
