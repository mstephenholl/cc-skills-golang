---
name: golang-data-structures
description: "Golang data structure choice and internals — slice capacity and aliasing, map growth (Swiss tables), container/list/heap/ring, generic containers, strings.Builder vs bytes.Buffer, and unsafe/weak pointers. Use when choosing between collection types, writing a generic container, or reasoning about slice or map memory. Not for applying optimization patterns after profiling (→ See `samber/cc-skills-golang@golang-performance` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.5"
  openclaw:
    emoji: "🗃"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__* mcp__context7__resolve-library-id mcp__context7__query-docs
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who understands data structure internals. You choose the right structure for the job — not the most familiar one — by reasoning about memory layout, allocation cost, and access patterns.

# Go Data Structures

Built-in and standard library data structures: internals, correct usage, and selection guidance.

- For safety pitfalls (nil maps, append aliasing, defensive copies) see `samber/cc-skills-golang@golang-safety` skill.
- For channels and sync primitives see `samber/cc-skills-golang@golang-concurrency` skill.
- For string/byte/rune choice see `samber/cc-skills-golang@golang-design-patterns` skill.

## Best Practices Summary

1. Preallocate when the size is known or estimable — `make([]T, 0, n)`, `make(map[K]V, n)`, `slices.Grow`, `strings.Builder.Grow` — each growth copies the whole backing array.
2. Never depend on when `append` reallocates — the growth algorithm is a runtime detail, not a spec guarantee; it changed in Go 1.18 and may change again, so preallocate instead of predicting it.
3. Prefer arrays only for fixed, compile-time sizes (`[32]byte` digests, IPv4 addresses, `[2]int` coordinates) — they are value types and comparable, so they work as allocation-free map keys where a formatted string key would allocate on every lookup.
4. Store large structs in maps as `map[K]*V` — map access copies the whole value; below ~128 bytes a value map is usually faster because pointers add GC pressure.
5. Generic data structures should use the **tightest constraint** possible — `comparable` for keys, `cmp.Ordered` for sorting, custom interfaces for domain-specific ordering.
6. **`unsafe.Pointer`** MUST only follow the 6 valid conversion patterns from the `unsafe` package docs — NEVER store it in a `uintptr` variable across statements, because a `uintptr` is not a reference: the GC may free the object, and stack growth may move it, before you convert back.
7. Consider **`weak.Pointer[T]`** (Go 1.24+) for canonicalization maps where the GC may reclaim entries, paired with `runtime.AddCleanup` to delete the dead map entry — not for caches that must retain entries, since a weak entry vanishes as soon as no strong reference remains.

## Slices and Maps

A slice is a 3-word header: pointer, length, capacity. Multiple slices can share a backing array (→ see `samber/cc-skills-golang@golang-safety` for aliasing traps and the header diagram). Read [slice-internals.md](./references/slice-internals.md) when using the `slices` package (`Clip`, `Grow`, `Compact`…), reasoning about `len` vs `cap`, or copying and splitting slices.

Since Go 1.24, maps are Swiss tables — 8-slot groups probed through a control word of hash fragments, with no overflow chains. They are reference types — assigning a map copies the pointer, not the data. Read [map-internals.md](./references/map-internals.md) when reasoning about map memory or growth: Swiss-table layout, load factor, why maps never shrink (and what to do about it), pointer vs value elements.

## Containers and Buffers

Prefer a slice unless a container's specific operation earns it: `container/list` only when you hold `*list.Element` handles for O(1) removal from the middle (LRU caches, ordered maps) — linked lists have poor cache locality — `container/heap` for priority queues, `container/ring` for fixed-size rotation. Assemble strings with `strings.Builder` — `bytes.Buffer.String()` copies the bytes — and use `bytes.Buffer` when you need an `io.Reader` or buffer reuse. Read [containers.md](./references/containers.md) when implementing one of these containers, choosing between the two buffers, or reading lines with `bufio.Scanner` (64 KB default token limit).

## Generic Collections (Go 1.18+)

Read [generics.md](./references/generics.md) when writing a generic container or deciding whether a type parameter earns its place at all — one constrained by `any` and never used for type-specific behavior is `interface{}` with extra syntax.

## Pointer Types

Read [pointers.md](./references/pointers.md) when using `unsafe.Pointer` (the 6 valid patterns, `unsafe.Add`/`unsafe.Slice`) or `weak.Pointer[T]` with `runtime.AddCleanup`.

## Copy Semantics Quick Reference

| Type | Copy Behavior | Independence |
| --- | --- | --- |
| `int`, `float`, `bool`, `string` | Value (deep copy) | Fully independent |
| `array`, `struct` | Value (deep copy) | Fully independent |
| `slice` | Header copied, backing array shared | Use `slices.Clone` |
| `map` | Reference copied | Use `maps.Clone` |
| `channel` | Reference copied | Same channel |
| `*T` (pointer) | Address copied | Same underlying value |
| `interface` | Value copied (type + value pair) | Depends on held type |

## Third-Party Libraries

For advanced data structures (trees, sets, queues, stacks) beyond the standard library:

- **`emirpasic/gods`** — comprehensive collection library (trees, sets, lists, stacks, maps, queues)
- **`deckarep/golang-set`** — thread-safe and non-thread-safe set implementations
- **`gammazero/deque`** — fast double-ended queue

When using third-party libraries, refer to their official documentation and code examples for current API signatures.

## Cross-References

- → See `samber/cc-skills-golang@golang-performance` skill for struct field alignment, memory layout optimization, and cache locality
- → See `samber/cc-skills-golang@golang-safety` skill for nil map/slice pitfalls, append aliasing, defensive copying, `slices.Clone`/`Equal`
- → See `samber/cc-skills-golang@golang-concurrency` skill for channels, `sync.Map`, `sync.Pool`, and all sync primitives
- → See `samber/cc-skills-golang@golang-design-patterns` skill for `string` vs `[]byte` vs `[]rune`, iterators, streaming
- → See `samber/cc-skills-golang@golang-structs-interfaces` skill for struct composition, embedding, and generics vs `any`
- → See `samber/cc-skills-golang@golang-code-style` skill for slice/map initialization style

## References

- [Go Data Structures (Russ Cox)](https://research.swtch.com/godata)
- [The Go Memory Model](https://go.dev/ref/mem)
- [Effective Go](https://go.dev/doc/effective_go)
