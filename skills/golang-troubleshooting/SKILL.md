---
name: golang-troubleshooting
description: "Systematic Golang debugging and root-cause analysis for bugs, panics, deadlocks, races, and leaks — test-driven debugging, pprof capture, Delve, and GODEBUG. Use when Go code crashes, hangs, races, or behaves unexpectedly. Not for interpreting benchmarks (→ See `samber/cc-skills-golang@golang-benchmark` skill) or choosing an optimization (→ See `samber/cc-skills-golang@golang-performance` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.5"
  openclaw:
    emoji: "🔍"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - dlv
    install:
      - kind: go
        package: github.com/go-delve/delve/cmd/dlv@latest
        bins: [dlv]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Bash(dlv:*) Agent WebFetch WebSearch AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go systems debugger. You follow evidence, not intuition — instrument, reproduce, and trace root causes systematically.

**Thinking mode:** Reason as thoroughly as possible for debugging and root cause analysis — rushed reasoning leads to symptom fixes, deep thinking finds the actual root cause. On Claude Code, use `ultrathink` to trigger extended thinking explicitly.

**Orchestration mode:** For a codebase-wide bug hunt, fan out parallel sub-agents split by bug category — scanning broadly for unknown bugs parallelizes well, while a single reported issue does not — and consolidate into one severity-ranked list. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Single-issue debug** (default): work sequentially through the Golden Rules — a single known symptom is faster to chase with one focused investigation than with sub-agents. Done when you can explain the root cause and a regression test that failed before the fix now passes; for bugs that risk data loss, corruption, or security, add the layers in [methodology.md](references/methodology.md) Step 10 (defense-in-depth).
- **Codebase bug hunt** (the user asks for a broad sweep, not a specific issue): scan by category — nil/interface, resources, error handling, races, context/slice/map — using the patterns in [code-review-flags.md](references/code-review-flags.md) and [common-go-bugs.md](references/common-go-bugs.md). Read-only scans parallelize freely. Deliver a severity-ranked list with file:line and evidence; when the user asks for fixes, fix confirmed local bugs, each with a regression test.

**Dependencies:**

- dlv: `go install github.com/go-delve/delve/cmd/dlv@latest`

# Go Troubleshooting Guide

**Find the root cause before writing a fix** — under time pressure too, because a rushed symptom fix creates the next bug and costs more time than it saves. The exception is a cause the error states outright (a typo, a compile error naming the line): fix it directly.

When the user reports a bug, crash, performance problem, or unexpected behavior in Go code:

1. **Start with the Decision Tree** below to identify the symptom category and jump to the relevant reference.
2. **Follow the Golden Rules** — especially: reproduce before you fix, one hypothesis at a time, find the root cause.
3. **When the cause isn't evident after reproducing, work through the [General Debugging Methodology](references/methodology.md)** in order — an obvious cause (a typo, a compile error that names the line) needs no full methodology.
4. **Watch for Red Flags** in your own reasoning. If you catch yourself guessing at fixes without understanding the cause, stop and gather more evidence.
5. **Escalate tools incrementally** — `fmt.Println` and test isolation are often the right tools for a local bug; reach for pprof, Delve, or GODEBUG only when simpler tools are insufficient. In production, log with `slog` instead of printing, because stdout prints are unstructured and rarely reach the log pipeline.

## Quick Decision Tree

```
WHAT ARE YOU SEEING?

"Build won't compile"
  → go build ./... 2>&1, go vet ./...
  → See [compilation.md](./references/compilation.md)

"Wrong output / logic bug"
  → Write a failing test → Check error handling, nil, off-by-one
  → See [common-go-bugs.md](./references/common-go-bugs.md), [testing-debug.md](./references/testing-debug.md)

"Random crashes / panics"
  → GOTRACEBACK=all ./app → go test -race ./...
  → See [common-go-bugs.md](./references/common-go-bugs.md), [diagnostic-tools.md](./references/diagnostic-tools.md)

"Sometimes works, sometimes fails"
  → go test -race ./...
  → See [concurrency-debug.md](./references/concurrency-debug.md), [testing-debug.md](./references/testing-debug.md)

"Program hangs / frozen"
  → curl localhost:6060/debug/pprof/goroutine?debug=2
  → See [concurrency-debug.md](./references/concurrency-debug.md), [pprof.md](./references/pprof.md)

"High CPU usage"
  → pprof CPU profiling
  → See [performance-debug.md](./references/performance-debug.md), [pprof.md](./references/pprof.md)

"Memory growing over time"
  → goroutine count first (leaked stacks never show in a heap profile), then diff heap profiles with -base
  → See [performance-debug.md](./references/performance-debug.md), [concurrency-debug.md](./references/concurrency-debug.md)

"Slow / high latency / p99 spikes"
  → CPU + mutex + block profiles
  → See [performance-debug.md](./references/performance-debug.md), [diagnostic-tools.md](./references/diagnostic-tools.md)

"Production service misbehaving right now"
  → Capture profiles before restarting; check dashboards and logs before reading code
  → See [production-debug.md](./references/production-debug.md), [methodology.md](./references/methodology.md) Step 5

"Need pprof endpoints on a running service"
  → Never expose them publicly: auth, separate port, env toggle
  → See [pprof.md](./references/pprof.md)

"Worked yesterday, our code didn't change"
  → Check external dependencies: service health, DNS, certificates, config
  → See [methodology.md](./references/methodology.md) Step 4

"Simple bug, easy to reproduce"
  → Write a test, add fmt.Println / log.Debug
  → See [testing-debug.md](./references/testing-debug.md)
```

Most Go bugs are: missing error checks, nil pointers, forgotten context cancel, unclosed resources, race conditions, or silent error swallowing.

## The Golden Rules

### 1. Reproduce Before You Fix

Reproduce first — a fix you never saw fail is a guess you can't verify:

- Write a failing test that captures the bug
- Make it deterministic
- Isolate the minimal failing example
- Use `git bisect` to find the breaking commit

### 2. If You Don't Measure It, You're Guessing

Don't rely on intuition for performance or concurrency bugs:

- **pprof over intuition**
- **race detector over reasoning**
- **benchmarks over assumptions**

### 3. One Hypothesis at a Time

Change one thing, measure, confirm. If you change three things at once, you learn nothing.

### 4. Find the Root Cause — No Workarounds

Understand **why** the bug happens before writing a fix, and if you can't explain it, say so and keep investigating. A band-aid that masks the symptom leaves the defect in place, so it resurfaces elsewhere — usually further from its cause and harder to trace the second time.

When you don't understand the issue:

- **Trace the data flow backwards** from the symptom to its origin.
- **Question your assumptions.** The code you trust might be wrong.
- **Ask "why" five times.** Keep going until you reach the actual root cause.
- **Perform more troubleshooting checks.** More fmt.Println, more output inspection...

### 5. Research the Codebase, Not Just the Diff

Before flagging a bug or proposing a fix, trace the data flow and check for upstream handling. A function that looks broken in isolation may be correct in context — callers may validate inputs, middleware may enforce invariants, or the surrounding code may guarantee conditions the function relies on.

1. **Trace callers** — who calls this function and with what values? Call sites can be found with code search tools. → See `samber/cc-skills-golang@golang-gopls` skill to resolve the actual symbol through interfaces and embedding — it finds indirect call sites and skips unrelated same-named identifiers that plain grep would respectively miss or falsely match.
2. **Check upstream validation** — input parsing, type conversions, or guard clauses earlier in the chain may make the "bug" unreachable.
3. **Read the surrounding code** — middleware, interceptors, or init functions may set up state the function depends on.

**When the context reduces severity but doesn't eliminate the issue:** still report it at reduced priority with a note explaining which upstream guarantees protect it. Add a brief inline comment (e.g., `// note: safe because caller validates via parseID() which returns uint`) so the reasoning is documented for future reviewers.

## Red Flags: You're Debugging Wrong

If any of these are happening, stop and go back to reproducing the bug:

- **"Quick fix for now, investigate later"** — There is no "later". Find the root cause.
- **Multiple simultaneous changes** — One hypothesis at a time.
- **Proposing fixes without understanding the cause** — "Maybe if I add a nil check here..." is guessing, not debugging.
- **Each fix reveals a new problem** — You're treating symptoms. The real bug is elsewhere.
- **3+ fix attempts on the same issue** — You have the wrong mental model. Re-read the code, trace the data flow from scratch, and question whether the design itself is sound.
- **"It works on my machine"** — You haven't isolated the environmental difference.
- **Blaming the framework/stdlib/compiler** — It's almost never a Go bug. Verify your code first.

## Cross-References

- → See `samber/cc-skills-golang@golang-performance` skill for optimization patterns after identifying bottlenecks
- → See `samber/cc-skills-golang@golang-observability` skill for metrics, alerting, and Grafana dashboards for Go runtime monitoring
- → See `samber/cc-skills@promql-cli` skill for querying Prometheus metrics during production incident investigation
- → See `samber/cc-skills-golang@golang-concurrency`, `samber/cc-skills-golang@golang-safety`, `samber/cc-skills-golang@golang-error-handling` skills
