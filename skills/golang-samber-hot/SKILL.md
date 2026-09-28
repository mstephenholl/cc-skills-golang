---
name: golang-samber-hot
description: "In-memory caching in Golang with samber/hot — choosing an eviction algorithm (LRU, LFU, W-TinyLFU, S3FIFO, SIEVE…), TTL and jitter, loaders, sharding, and stale-while-revalidate. Apply when using or adopting samber/hot, or when the codebase imports `github.com/samber/hot`. For caching strategy in general → See `samber/cc-skills-golang@golang-performance` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.1.4"
  openclaw:
    emoji: "🔥"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "0.13.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who treats caching as a system design decision. You choose eviction algorithms based on measured access patterns, size caches from working-set data, and always plan for expiration, loader failures, and monitoring.

# Using samber/hot for In-Memory Caching in Go

Generic, type-safe in-memory caching library for Go 1.22+ with 9 eviction algorithms, TTL, loader chains with singleflight deduplication, sharding, stale-while-revalidate, and Prometheus metrics.

**Official Resources:**

- [pkg.go.dev/github.com/samber/hot](https://pkg.go.dev/github.com/samber/hot)
- [github.com/samber/hot](https://github.com/samber/hot)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get -u github.com/samber/hot
```

## Algorithm Selection

Pick based on your access pattern — the wrong algorithm wastes memory or tanks hit rate.

| Algorithm | Constant | Best for | Avoid when |
| --- | --- | --- | --- |
| **W-TinyLFU** | `hot.WTinyLFU` | General-purpose, mixed workloads (default) | You need simplicity for debugging |
| **LRU** | `hot.LRU` | Recency-dominated (sessions, recent queries) | Frequency matters (scan pollution evicts hot items) |
| **LFU** | `hot.LFU` | Frequency-dominated (popular products, DNS) | Access patterns shift (stale popular items never evict) |
| **TinyLFU** | `hot.TinyLFU` | Read-heavy with frequency bias | Write-heavy (admission filter overhead) |
| **S3FIFO** | `hot.S3FIFO` | High throughput, scan-resistant | Small caches (<1000 items) |
| **ARC** | `hot.ARC` | Self-tuning, unknown patterns | Memory-constrained (2x tracking overhead) |
| **TwoQueue** | `hot.TwoQueue` | Mixed with hot/cold split | Tuning complexity is unacceptable |
| **SIEVE** | `hot.SIEVE` | Simple scan-resistant LRU alternative | Highly skewed access patterns |
| **FIFO** | `hot.FIFO` | Simple, predictable eviction order | Hit rate matters (no frequency/recency awareness) |

**Decision shortcut:** Start with `hot.WTinyLFU`. Switch only when profiling shows the miss rate is too high for your SLO.

Read [references/algorithm-guide.md](./references/algorithm-guide.md) when the workload doesn't clearly match one row, or when comparing algorithms by hit rate.

## Core Usage

### Basic Cache with TTL

```go
import "github.com/samber/hot"

cache := hot.NewHotCache[string, *User](hot.WTinyLFU, 10_000).
    WithTTL(5 * time.Minute).
    WithJanitor().
    Build()
defer cache.StopJanitor()

cache.Set("user:123", user)
cache.SetWithTTL("session:abc", session, 30*time.Minute)

value, found, err := cache.Get("user:123")
```

### Loader Pattern (Read-Through)

Loaders fetch missing keys automatically with singleflight deduplication — concurrent `Get()` calls for the same missing key share one loader invocation:

```go
cache := hot.NewHotCache[int, *User](hot.WTinyLFU, 10_000).
    WithTTL(5 * time.Minute).
    WithLoaders(func(ids []int) (map[int]*User, error) {
        return db.GetUsersByIDs(ctx, ids) // batch query
    }).
    WithJanitor().
    Build()
defer cache.StopJanitor()

user, found, err := cache.Get(123) // triggers loader on miss
```

## Capacity Sizing

Capacity is an item count, not bytes, so derive it from a memory budget: `capacity = memoryBudget / bytesPerEntry`, rounded down for headroom. Estimate bytes per entry as the struct size plus its heap-allocated fields (slices, maps, strings), the key, and ~100 bytes of bookkeeping (pointers, expiry timestamps, algorithm metadata).

```
Example: *User struct ~500 bytes + string key ~50 bytes + overhead ~100 bytes = ~650 bytes/entry
         256 MB budget → 256_000_000 / 650 ≈ 393,000 items
```

Expose capacity as configuration rather than a constant, since the right value depends on the production memory limit. If the budget is unknown, choose a conservative default, write a test that fills N entries and compares `runtime.ReadMemStats` before and after to measure bytes per entry, and state the assumption in your reply.

## Common Mistakes

1. **TTL and janitor out of step** — set a TTL when the source data changes (immutable lookups can skip it); when a TTL is set, chain `.WithJanitor()` and `defer cache.StopJanitor()`, or expired entries stay in memory until the algorithm evicts them. `WithJanitor()` without a TTL panics in `Build()` (non-positive ticker interval).
2. **Calling `SetMissing()` without missing cache config** — panics at runtime. Enable `WithMissingCache(algorithm, capacity)` or `WithMissingSharedCache()` in the builder first.
3. **`WithoutLocking()` + `WithJanitor()`** — mutually exclusive, panics. `WithoutLocking()` is only safe for single-goroutine access without background cleanup.
4. **Oversized cache** — a cache holding everything is a map with overhead. Size to your working set (typically 10-20% of total data). Monitor hit rate to validate.
5. **Ignoring loader errors** — `Get()` returns `(zero, false, err)` on loader failure. Always check `err`, not just `found`.

## Best Practices

1. Use `WithJitter(lambda, upperBound)` to spread expirations — without jitter, items created together expire together, causing thundering herd on the loader
2. Monitor with `WithPrometheusMetrics(cacheName)` — hit rate below 80% usually means the cache is undersized or the algorithm is wrong for the workload
3. Use `WithCopyOnRead(fn)` / `WithCopyOnWrite(fn)` for mutable values — without copies, callers mutate cached objects and corrupt shared state

Read [references/production-patterns.md](./references/production-patterns.md) when adding stale-while-revalidate, sharding, negative (missing-key) caching, loader chains, warm-up, or Prometheus alerting.

Read [references/api-reference.md](./references/api-reference.md) when checking a builder option or cache method signature — for example `Peek` (no loader, no expiry check) versus `Get`.

If you encounter a bug or unexpected behavior in samber/hot, open an issue at <https://github.com/samber/hot/issues>.

## Cross-References

- → See `samber/cc-skills-golang@golang-performance` skill for general caching strategy and when to use in-memory cache vs Redis vs CDN
- → See `samber/cc-skills-golang@golang-observability` skill for Prometheus metrics integration and monitoring
- → See `samber/cc-skills-golang@golang-database` skill for database query patterns that pair with cache loaders
- → See `samber/cc-skills@promql-cli` skill for querying Prometheus cache metrics via CLI
