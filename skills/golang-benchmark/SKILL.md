---
name: golang-benchmark
description: "Golang benchmark measurement and profile interpretation — `testing.B` benchmarks, pprof CPU/memory/trace analysis, benchstat comparisons, and CI regression detection. Use when writing or comparing Go benchmarks, reading a profile, or judging whether a performance change is statistically real. Not for choosing the optimization to apply (→ See `samber/cc-skills-golang@golang-performance` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.6"
  openclaw:
    emoji: "📊"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - benchstat
    install:
      - kind: go
        package: golang.org/x/perf/cmd/benchstat@latest
        bins: [benchstat]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch Bash(benchstat:*) Bash(benchdiff:*) Bash(cob:*) Bash(gobenchdata:*) Bash(curl:*) mcp__context7__resolve-library-id mcp__context7__query-docs WebSearch AskUserQuestion EnterWorktree ExitWorktree
paths:
  - "**/*.go"
---

**Persona:** You are a Go performance measurement engineer. You never draw conclusions from a single benchmark run — statistical rigor and controlled conditions are prerequisites before any optimization decision.

**Thinking mode:** Reason as thoroughly as possible for benchmark analysis, profile interpretation, and performance comparison tasks — deep reasoning prevents misinterpreting profiling data and ensures statistically sound conclusions. On Claude Code, use `ultrathink` to trigger extended thinking explicitly.

**Dependencies:**

- benchstat: `go install golang.org/x/perf/cmd/benchstat@latest`

# Go Benchmarking & Performance Measurement

Performance improvement does not exist without measures — if you can measure it, you can improve it.

## Writing Benchmarks

### File and Ordering Conventions

Benchmark functions live in a `_bench_test.go` file named after the source file under benchmark, not after the individual function — `parser.go` -> `parser_bench_test.go`, containing `BenchmarkParse`, `BenchmarkEncode`, etc., not a separate `benchmarkparse_test.go` per function.

- Keeping benchmarks in their own file (instead of mixed into `parser_test.go`) keeps `go test -bench=. ./pkg/parser` output free of unrelated `Test*` noise.
- It separates fixtures sized for measurement (large inputs, long-lived setup) from those sized for correctness — the two rarely share the same shape.
- The file still follows Go's one-test-file-per-source-file convention (→ See `samber/cc-skills-golang@golang-testing` skill), just with the `_bench` suffix marking its narrower purpose.

Order `Benchmark*` functions inside `parser_bench_test.go` to mirror the order of the functions/methods they measure in `parser.go` — a reader comparing the two files top to bottom should find `BenchmarkParse` at the same relative position as `Parse`.

### `b.Loop()` (Go 1.24+) — preferred

For Go 1.24+, prefer `b.Loop()` for new benchmarks. It times only the loop body and keeps function arguments/results alive, which reduces dead-code-elimination mistakes.

```go
func BenchmarkParse(b *testing.B) {
    data := loadFixture("large.json") // setup — excluded from timing
    for b.Loop() {
        Parse(data)  // compiler cannot eliminate this call
    }
}
```

Legacy `b.N` loops still compile and are fine to keep when preserving existing benchmarks or supporting Go <1.24. They are easier to get wrong: setup may need `b.ResetTimer()`, and results may need a sink if the compiler can eliminate the work. Go 1.26 fixed an earlier `b.Loop()` inlining limitation — benchmarks on 1.24–1.25 already benefit from `b.Loop()` but may miss inlining optimizations that 1.26 delivers.

Go 1.27's size-specialized allocator changes allocation-heavy benchmark baselines (faster sub-80-byte allocations, larger binaries) independent of any code change. Treat a `benchstat` comparison that straddles the Go 1.26→1.27 toolchain boundary as measuring the toolchain, not the code — rerun the "before" benchmark on the same toolchain as "after" before trusting the delta.

### Custom metrics

```go
for b.Loop() {
    Encode(data)
}
// after the loop: b.Loop() has stopped the timer, so b.Elapsed() covers only the timed iterations
b.ReportMetric(float64(b.N*len(data))/b.Elapsed().Seconds(), "bytes/s")
```

## Running Benchmarks

```bash
go test -run='^$' -bench=BenchmarkEncode -benchmem -count=10 ./pkg/... | tee bench.txt
```

Read [go-test-bench.md](./references/go-test-bench.md) when writing sub-benchmarks, choosing other `go test` flags (`-benchtime`, `-cpu`, profile outputs), or decoding a raw `go test -bench` result line.

## Comparing Optimization Variants in Parallel

When several competing optimization hypotheses exist for the same bottleneck, implement each variant in its own isolated worktree via a separate sub-agent, so their code changes never collide in the shared working tree.

**Run the benchmarks serially, not concurrently** — one variant at a time, back in the main tree or sequentially per worktree. Concurrent runs share the same CPU, so the noisy-neighbor effect contaminates `ns/op` and reintroduces the exact noise `-count` and `benchstat` exist to eliminate. Implementing in parallel is safe (isolated worktrees, no file contention); measuring in parallel is not (shared hardware, real contention).

Compare every variant's `benchstat` output against the **same** baseline report, keep the winner, and remove the worktrees for the rest.

## Documenting Results in Commits

Paste benchstat output in the commit body when the change has a measurable performance impact. This documents _why_ an optimization was made, prevents future readers from reverting it, and lets reviewers verify the claim without re-running benchmarks.

Commit body example:

```
perf(parser): reduce Parse allocations 50% with sync.Pool

Replace per-call []byte allocation with a pooled buffer.

goos: linux / goarch: amd64 / cpu: AMD Ryzen 9 5950X
          │    old     │              new               │
          │  sec/op    │  sec/op     vs base            │
Parse-32    4.592µ ± 2%  3.041µ ± 1%  -33.78% (p=0.000 n=10)

          │   old    │             new              │
          │   B/op   │   B/op     vs base           │
Parse-32   1.024Ki ± 0%  0.512Ki ± 0%  -50.00% (p=0.000 n=10)

          │ old  │            new             │
          │ allocs/op │ allocs/op  vs base    │
Parse-32   12.00 ± 0%   6.000 ± 0%  -50.00% (p=0.000 n=10)
```

**Rules:**

- Only include benchmarks directly affected by the change — strip unrelated rows
- Never paste results with `~` (no statistical significance) — the improvement cannot be claimed
- Include the hardware context line (`goos/goarch/cpu`) so results are reproducible
- Follow the repo's commit convention; if it uses Conventional Commits, use `perf(scope):` for performance-only changes

## Profiling from Benchmarks

Generate profiles directly from benchmark runs — no HTTP server needed:

```bash
# CPU profile
go test -bench=BenchmarkParse -cpuprofile=cpu.prof ./pkg/parser
go tool pprof cpu.prof

# Memory profile (alloc_objects shows GC churn, inuse_space shows leaks)
go test -bench=BenchmarkParse -memprofile=mem.prof ./pkg/parser
go tool pprof -alloc_objects mem.prof

# Execution trace
go test -bench=BenchmarkParse -trace=trace.out ./pkg/parser
go tool trace trace.out
```

## Reference Files

- [go-test-bench.md](./references/go-test-bench.md) — when writing sub-benchmarks, picking `go test` bench flags, or decoding a raw result line
- [benchstat.md](./references/benchstat.md) — when comparing runs, reading `~`/p-values/±%, choosing `-count`, interleaving runs, or filtering and projecting results
- [pprof.md](./references/pprof.md) — when reading a CPU, heap, goroutine, mutex or block profile, or filtering, labeling, diffing or exporting one
- [trace.md](./references/trace.md) — when latency is high but CPU is low, or when investigating scheduling, GC phases, annotations or the flight recorder
- [compiler-analysis.md](./references/compiler-analysis.md) — when a benchmark shows allocations or call overhead you didn't expect (escape analysis, inlining, bounds checks, SSA, assembly), or to verify a compiler decision (receiver choice, inlining, escape) before relying on it
- [tools.md](./references/tools.md) — when a narrower diagnostic answers the question (GODEBUG, `runtime/metrics`, `expvar`, fieldalignment, fgprof, perf)
- [ci-regression.md](./references/ci-regression.md) — when adding benchmark gating to a pipeline, choosing benchdiff/cob/gobenchdata, or tuning runners for stable results
- [investigation-session.md](./references/investigation-session.md) — when production behaves differently from benchmarks and you need a temporary instrumented session (PromQL, host correlation, cost warnings), or when you need alert thresholds for GC pressure, goroutine leaks, non-heap RSS growth, or `go_memstats_*` scrape overhead
- [prometheus-go-metrics.md](./references/prometheus-go-metrics.md) — when writing PromQL on Go runtime metrics or enabling the `runtime/metrics` collectors

## Cross-References

- → See `samber/cc-skills-golang@golang-performance` skill for optimization patterns to apply after measuring ("if X bottleneck, apply Y")
- → See `samber/cc-skills-golang@golang-troubleshooting` skill for pprof setup on running services (enable, secure, capture), Delve debugger, GODEBUG flags, root cause methodology
- → See `samber/cc-skills-golang@golang-observability` skill for everyday always-on monitoring, continuous profiling (Pyroscope), distributed tracing (OpenTelemetry)
- → See `samber/cc-skills-golang@golang-testing` skill for general testing practices
- → See `samber/cc-skills@promql-cli` skill for querying Prometheus runtime metrics in production to validate benchmark findings
