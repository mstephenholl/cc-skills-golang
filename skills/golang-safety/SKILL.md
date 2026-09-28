---
name: golang-safety
description: "Defensive Golang coding against accidental bugs — nil maps and pointers, typed-nil interfaces, append aliasing, integer truncation, float comparison, defer in loops, and safe zero values. Use when a Go program panics on nil, when reviewing code for nil-safety or numeric conversions, or when designing a type whose zero value must work. Not for concurrency design (→ See `samber/cc-skills-golang@golang-concurrency` skill) or exploitable vulnerabilities (→ See `samber/cc-skills-golang@golang-security` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.4"
  openclaw:
    emoji: "🛡"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent
paths:
  - "**/*.go"
---

**Persona:** You are a defensive Go engineer. You treat every untested assumption about nil, capacity, and numeric range as a latent crash waiting to happen.

# Go Safety: Correctness & Defensive Coding

Prevents programmer mistakes — bugs, panics, and silent data corruption in normal (non-adversarial) code. Security handles attackers; safety handles ourselves.

## Best Practices Summary

1. **Prefer generics over `any`** when the type set is known — compiler catches mismatches instead of runtime panics
2. **Use comma-ok type assertions** (`v, ok := x.(T)`) unless a mismatch is a programming error that should panic; for reflection in Go 1.25+, prefer `reflect.TypeAssert[T](value)` over `value.Interface().(T)`
3. **Typed nil pointer in an interface is not `== nil`** — the type descriptor makes it non-nil
4. **Writing to a nil map panics** — initialize before the first write
5. **`append` may reuse the backing array** — both slices share memory if capacity allows, silently corrupting each other
6. **Copy slices and maps at API boundaries** — on the way in (constructor arguments) and out (getters), behind an unexported field — otherwise callers mutate your internals through the shared backing array
7. **`defer` runs at function exit, not loop iteration** — extract the loop body to a function
8. **Integer conversions truncate silently** — `int64` to `int32` wraps without error
9. **Float arithmetic is not exact** — use epsilon comparison or `math/big`
10. **Design useful zero values**, as `sync.Mutex` and `bytes.Buffer` do — nil map fields panic on first write, so lazy-init them in methods
11. **Use `sync.Once` for lazy init** — it guarantees exactly-once even under concurrency; use `sync.OnceValues` (Go 1.21+) when the init can fail, so every caller gets the error instead of it being discarded

## Nil Safety

### The nil interface trap

Interfaces store (type, value). An interface is `nil` only when both are nil. Returning a typed nil pointer sets the type descriptor, making it non-nil:

```go
// ✗ Dangerous — interface{type: *MyHandler, value: nil} is not == nil
func getHandler() http.Handler {
    var h *MyHandler // nil pointer
    if !enabled {
        return h // interface{type: *MyHandler, value: nil} != nil
    }
    return h
}

// ✓ Good — return nil explicitly
func getHandler() http.Handler {
    if !enabled {
        return nil // interface{type: nil, value: nil} == nil
    }
    return &MyHandler{}
}
```

### Nil map, slice, and channel behavior

| Type    | Index into nil | Write to nil   | Len/Cap of nil | Range over nil |
| ------- | -------------- | -------------- | -------------- | -------------- |
| Map     | Zero value     | **panic**      | 0              | 0 iterations   |
| Slice   | **panic**      | **panic**      | 0              | 0 iterations   |
| Channel | Blocks forever | Blocks forever | 0              | Blocks forever |

Read [nil-safety.md](./references/nil-safety.md) when a type has optional func fields or callbacks, methods that may run on a nil receiver, an error return built from a typed pointer, or generic code that compares against nil.

## Slice & Map Safety

`append` reuses the backing array if capacity allows, so both slices then share memory:

```go
// ✗ Dangerous — a and b share backing array
a := make([]int, 3, 5)
b := append(a, 4)
b[0] = 99 // also modifies a[0]

// ✓ Good — full slice expression forces new allocation
b := append(a[:len(a):len(a)], 4)
```

Concurrent map access with any writer crashes the process with a fatal error that `recover` cannot catch — → See `samber/cc-skills-golang@golang-concurrency` for sync primitives.

Read [slice-map-safety.md](./references/slice-map-safety.md) when returning part of a larger buffer, producing output from map iteration, deleting while iterating, or storing pointers to loop variables.

## Enforce with Linters

Many safety pitfalls are caught automatically by linters: `errcheck`, `forcetypeassert`, `nilerr`, `govet`, `staticcheck`. See the `samber/cc-skills-golang@golang-lint` skill for configuration and usage.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Bare type assertion `v := x.(T)` | Panics on type mismatch, crashing the program. Use `v, ok := x.(T)` unless a mismatch is a bug that should crash |
| Returning typed nil in interface function | Interface holds (type, nil) which is != nil. Return untyped `nil` for the nil case |
| Writing to a nil map | Nil maps have no backing storage — write panics. Initialize with `make(map[K]V)` or lazy-init |
| Calling an optional func field | Calling a nil func panics. Check `if t.OnDone != nil` before calling, or default it in the constructor |
| Assuming `append` always copies | If capacity allows, both slices share the backing array. Use `s[:len(s):len(s)]` to force a copy |
| Returning a small subslice of a large buffer | The subslice keeps the whole backing array alive, so the GC cannot free it. Return `slices.Clone(buf[:n])` |
| `defer` in a loop | `defer` runs at function exit, not loop iteration — resources accumulate. Extract body to a separate function |
| Narrowing an integer without a bounds check | Values wrap silently (3B → -1.29B as `int32`; -1 → 255 as `byte`). Check both bounds against the target's `math.Min*`/`math.Max*`, and reject negatives for unsigned targets |
| Comparing floats with `==` | IEEE 754 representation is not exact (`0.1+0.2 != 0.3`). Use `math.Abs(a-b) < epsilon` — `1e-9` for general precision, `1e-2` for cents |
| Dividing without a zero check | Integer division by zero panics; float division yields `±Inf` or `NaN`, which propagates silently and fails every comparison. Guard `divisor == 0` first |
| Returning internal slice/map reference | Callers can mutate your struct's internals through the shared backing array. Return a defensive copy |
| Multiple `init()` with ordering assumptions | `init()` execution order across files is unspecified. → See `samber/cc-skills-golang@golang-design-patterns` — use explicit constructors |
| Blocking forever on nil channel | Nil channels block on both send and receive. Initialize before use — except when deliberately setting a channel to nil to disable its `select` case |

## Cross-References

- → See `samber/cc-skills-golang@golang-concurrency` skill for concurrent access patterns and sync primitives
- → See `samber/cc-skills-golang@golang-data-structures` skill for slice/map internals, capacity growth, and container/ packages
- → See `samber/cc-skills-golang@golang-error-handling` skill for nil error interface trap
- → See `samber/cc-skills-golang@golang-security` skill for security-relevant safety issues (memory safety, integer overflow)
- → See `samber/cc-skills-golang@golang-troubleshooting` skill for debugging panics and race conditions
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for automated AI-driven code review in CI using these guidelines
