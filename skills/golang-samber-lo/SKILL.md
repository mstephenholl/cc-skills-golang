---
name: golang-samber-lo
description: "Functional helpers for Golang with samber/lo — Map, Filter, GroupBy and friends, plus choosing between lo, lop (parallel), lom (in-place), loi (iterators), and the standard library. Apply when using or adopting samber/lo, or when the codebase imports `github.com/samber/lo`. Not for streaming pipelines (→ See `samber/cc-skills-golang@golang-samber-ro` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "🧰"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.53.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) mcp__context7__resolve-library-id mcp__context7__query-docs AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who prefers declarative collection transforms over manual loops. You reach for `lo` to eliminate boilerplate, but you know when the stdlib is enough and when to upgrade to `lop`, `lom`, or `loi`.

# samber/lo — Functional Utilities for Go

Generics-first helpers for slices, maps, strings, channels and tuples; the core `lo` package never mutates its input. The library is v1 with strict semver — only `exp/` packages may break, there is no v2, so install with `go get github.com/samber/lo@v1`. Its one non-stdlib dependency is the Go project's `golang.org/x/text`, used only by `lo.Capitalize`; every other helper runs on the standard library alone.

**Official Resources:**

- [github.com/samber/lo](https://github.com/samber/lo)
- [lo.samber.dev](https://lo.samber.dev)
- [pkg.go.dev/github.com/samber/lo](https://pkg.go.dev/github.com/samber/lo)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

## Choose the Package

Start with `lo`; move a call site to another package only when a profile or a lazy-evaluation need justifies it.

| Alias | Import | Go | Use when | Cost |
| --- | --- | --- | --- | --- |
| `lo` | `github.com/samber/lo` | 1.18+ | Default for every transform | Allocates a new slice/map per call |
| `lop` | `github.com/samber/lo/parallel` | 1.18+ | Expensive per-item CPU work (parsing, hashing, image ops), or ~1000+ items | One goroutine per element, no limit, no context, no error return |
| `lom` | `github.com/samber/lo/mutable` | 1.18+ | Hot path where `pprof -alloc_objects` blames `lo.Filter`/`lo.Map` | Mutates the input; unsafe while other goroutines read it |
| `loi` | `github.com/samber/lo/it` | 1.23+ | 3+ chained transforms on large data, or only the first N results needed | Lazy `iter.Seq` pipelines built on range-over-func |
| `simd` | `github.com/samber/lo/exp/simd` | see its `go.mod` | Numeric bulk ops (Sum, Mean, Min, Max, Contains) after a benchmark | Experimental, outside semver |

- **`lop` is CPU parallelism, not I/O concurrency** — it has no concurrency limit, no context cancellation and no error path, so for HTTP or database fan-out use `errgroup` with `SetLimit` and a context.
- **`lom` only after the profiler says so** — in-place mutation trades away the immutability callers rely on, so measure allocation pressure first and keep it off slices shared across goroutines.
- **`loi` removes intermediate slices** — `Filter → Map → Take` runs as one lazy pass; below Go 1.23, compose `lo` calls or write a single loop instead.
- **`exp/simd` needs `GOEXPERIMENT=simd` and publishes only pseudo-versions** — pin the exact version and benchmark against plain `lo`, since without hardware support it falls back to the scalar `lo` functions.
- **Infinite or event-driven streams belong to `samber/ro`** — `lo` transforms finite collections (→ See `samber/cc-skills-golang@golang-samber-ro` skill).

**Diagnose:** 1- `go tool pprof -alloc_objects` — confirm `lo.*` calls top the allocation profile before moving them to `lom` 2- `go tool pprof` on a CPU profile — confirm the transform callback dominates before moving to `lop`

Read [references/package-guide.md](references/package-guide.md) when switching a call site to `lop`, `lom`, `loi` or `simd`.

## Rules

- **Prefer the stdlib when it covers the operation** — `slices.Contains` and `slices.Sort` (Go 1.21+) and `slices.Collect(maps.Keys(m))` (Go 1.23+; `maps.Keys` returns an iterator, not a slice) need no third-party call. Use `lo` for what the stdlib lacks: Map, Filter, Reduce, GroupBy, Chunk, Flatten and the error variants.
- **Pick the error variant by failure policy** — `MapErr`, `FilterErr`, `ReduceErr` and the other `…Err` helpers stop at the first error and return `(result, error)`; to keep the successes and drop the failures in one pass, return `(value, ok)` from `lo.FilterMap`.
- **Keep `lo.Must` to tests and startup** — a panic is the right failure for a missing config in `main`, but in a request handler a malformed body then panics instead of returning a 400.

## Common Mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| Discarding `lo.Filter`'s return value | `lo` returns a new slice; the input is untouched | Assign the result, or use `lom.Filter` when in-place mutation is intended |
| `lop.Map` over 50 cheap items | Goroutine setup costs more than the transform | `lo.Map`; keep `lop` for expensive per-item CPU work |
| `lo.Ternary(cond, expensive(), cheap)` | Go evaluates both arguments before the call, so `expensive()` always runs | `lo.TernaryF` with closures, or a plain `if` |
| Chaining eager transforms over large data | Each step allocates an intermediate slice | `loi` on Go 1.23+ |

## References

- [references/api-reference.md](references/api-reference.md) — when looking up a helper by domain (maps, strings, tuples, channels, retry, conditionals, pointers) or checking its signature.
- [references/advanced-patterns.md](references/advanced-patterns.md) — when composing multi-step transforms, building lookup maps from slices, combining `lo` with `samber/mo`, or writing `loi` pipelines.

If you encounter a bug or unexpected behavior in samber/lo, open an issue at [github.com/samber/lo/issues](https://github.com/samber/lo/issues).

## Cross-References

- → See `samber/cc-skills-golang@golang-samber-mo` skill for monadic types (Option, Result, Either) that compose with lo transforms
- → See `samber/cc-skills-golang@golang-data-structures` skill for choosing the right underlying data structure
- → See `samber/cc-skills-golang@golang-performance` skill for profiling methodology before switching to `lom`/`lop`
