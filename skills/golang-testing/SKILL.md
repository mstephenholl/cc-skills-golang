---
name: golang-testing
description: "Golang tests — table-driven tests, parallel tests, fuzzing, fixtures, goroutine leak checks, synctest, coverage, and integration tests. Use when writing or reviewing Go tests, choosing a testing approach, or fixing flaky or slow tests. For testify APIs → See `samber/cc-skills-golang@golang-stretchr-testify` skill; for benchmarks → See `samber/cc-skills-golang@golang-benchmark` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.4.3"
  openclaw:
    emoji: "🧪"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - gotests
    install:
      - kind: go
        package: github.com/cweill/gotests/gotests@latest
        bins: [gotests]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent Bash(gotests:*) AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who treats tests as executable specifications. You write tests to constrain behavior, not to hit coverage targets.

**Thinking mode:** Reason as thoroughly as possible when diagnosing a failing or flaky test — flakiness usually hides a race, shared state, or a real-time dependency that a shallow read writes off as "CI noise". On Claude Code, use `ultrathink` to trigger extended thinking explicitly.

**Orchestration mode:** For auditing a large test suite, fan out parallel sub-agents split by the Audit mode concerns — each is an independent read-only scan — and consolidate into one ranked gap report. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Write mode** — new tests for new or existing code. `gotests` can scaffold the table if installed; the value is in the edge cases and error paths you add. Done when the tests pass and each one would fail if the behavior it pins broke.
- **Review mode** — a PR's test changes. Check that new behavior is covered, assertions pin outcomes rather than internals, and no flakiness pattern slipped in. Deliverable: findings with file:line; fixes applied if the user asked for them.
- **Audit mode** — an existing suite. Concerns: coverage gaps and assertion quality; integration isolation and build tags; goroutine leaks and races; flakiness (order dependence, real sleeps, shared state). Deliverable: one severity-ranked gap report with file:line; if the user asked for fixes, apply them across disjoint files and re-run with `-race -shuffle=on`.
- **Debug mode** — a failing or flaky test. Reproduce with `-run '^TestX$' -count=N -race`, isolate the failing assertion, trace the cause into production code or setup. Done when the cause is explained and the repro passes repeatedly.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-testing` skill takes precedence.

**Dependencies:**

- gotests (optional, table scaffolding): `go install github.com/cweill/gotests/gotests@latest`

# Go Testing

## Rules

1. **Name every table case** — a `name` field passed to `t.Run` makes failures and `-run 'TestX/case'` point at one case; an index doesn't survive reordering.
2. **Call `t.Parallel()` in the top-level test and in each subtest** of independent cases — the parent call overlaps the test with the package's other parallel tests, the subtest call overlaps its cases. Skip it for tests touching process-global state: `t.Setenv` and `t.Chdir` panic under a parallel test or ancestor.
3. **No order dependence** — each test builds its own fixtures and passes alone under `-run` and shuffled under `-shuffle=on`; a test that only passes after another one is reading leftover state.
4. **Test through the public API, from `package foo_test`** — a test that inspects unexported fields, caches, or memoization turns every refactor into a rewrite while proving nothing about the contract. Assert what callers observe (same input → same output, hit vs miss, `Len()`), even when a teammate suggests peeking at internals; use `package foo` only for an unexported function with real logic of its own.
5. **Mock interfaces defined at the consumer, not concrete types** — extract the small interface the code under test calls and inject it; a mock that embeds or wraps the concrete struct still drags the real dependency in. Read [mocking.md](./references/mocking.md) when writing mocks or shared fixtures.
6. **Separate integration tests with a `//go:build integration` file tag** and run them with `go test -tags=integration ./...` — unlike `testing.Short()`, which still compiles them and dials the database on every plain `go test ./...` unless someone remembers `-short`, a tagged file is excluded from the build. Read [integration-testing.md](./references/integration-testing.md) when writing integration tests (Docker Compose fixtures, schema setup, suite lifecycle).
7. **Check for goroutine leaks** in packages that start goroutines: `goleak.VerifyTestMain(m)` in `TestMain`, or `defer goleak.VerifyNone(t)` in one test — a `Stop()` that returns while workers still run passes every functional assertion. Add `goleak.IgnoreCurrent()` to exclude goroutines that existed before the test. → See `samber/cc-skills-golang@golang-concurrency` skill for leak causes and fixes.
8. **Don't wait on the real clock** — `time.Sleep` in a test makes it slow and flaky. For logic that reads `time.Now()`, inject a clock and advance a fake one (see Time Mocking in [mocking.md](./references/mocking.md)). Read [synctest.md](./references/synctest.md) when testing timers, tickers, deadlines, or timing-flaky code; on Go 1.25+ that means `synctest.Test`, never the Go 1.24 experimental `synctest.Run`.
9. **Run `go test -race` in CI** — a data race rarely fails a normal run, so concurrent code that passes without the detector proves little.
10. **Fuzz functions that take untrusted input** (parsers, decoders, sanitizers) alongside their table test: seed with `f.Add` and assert properties that hold for every input — round-trip (`Decode(Encode(x)) == x`), idempotence, invariants such as "no `<` in the output" — because the fuzzer generates inputs with no expected output to compare against.
11. **Write `ExampleXxx` functions for exported APIs** — `go test` checks their stdout against `// Output:` and pkg.go.dev renders them beside the symbol, so they are documentation that fails the build when it drifts. Read [examples.md](./references/examples.md) for naming rules and `// Unordered output:`.
12. **With testify, build `is := assert.New(t)` inside each `t.Run` closure** — a parent-scoped instance reports subtest failures against the parent while the failing subtest prints PASS. → See `samber/cc-skills-golang@golang-stretchr-testify` skill for the trap and its diagnostic.

## Test file layout

Name the test file after the source file it tests, not after the function or method under test: `helloworld.go` → `helloworld_test.go`, holding `TestHelloWorld`, `TestAbcd`, and the rest. Tools (`go test`, coverage reports, IDE "jump to test", `gotests`) and reviewers resolve tests by source file, so tests split by symbol name (`abcd_test.go`) scatter across files and break that mapping.

A very large source file may split its tests by concern (`foo_test.go` + `foo_edgecases_test.go`), but every split name still derives from the source file, never from a function name. Prefer one `_test.go` per source file even when it grows, since each split adds navigation overhead.

Order test functions to match the order of the functions they test in the source file, so a reader scrolling `foo.go` beside `foo_test.go` finds the matching test by position; drift between the two orderings compounds as either file grows.

## HTTP handlers

Test handlers with `httptest.NewRecorder()` and `handler.ServeHTTP(w, req)` — no listener, no port; reserve a real server for testing an HTTP client. Read [http-testing.md](./references/http-testing.md) for table-driven handler tests with bodies, query parameters, and headers.

## Benchmarks

Write each variant as a sub-benchmark (`b.Run`), so each gets its own name for comparison tooling to diff, and use `b.Loop()` rather than a `b.N` loop on Go 1.24+. Read [benchmarks.md](./references/benchmarks.md) for the code shape and size-parameterized benchmarks. → See `samber/cc-skills-golang@golang-benchmark` skill for `benchstat`, profiling, and CI regression detection.

## Coverage

Coverage locates untested paths; it does not measure assertion quality, so read the uncovered lines (`go tool cover -html=coverage.out`) and treat the percentage as a gap finder, not a target. Read [coverage.md](./references/coverage.md) for modes, `-coverpkg`, and reporting pitfalls.

## Recent Go versions

- **Go 1.26+: `t.ArtifactDir()`** (also on `*testing.B` and `*testing.F`) — write files a test wants to keep for inspection there instead of ad-hoc paths or the repo. They survive only when the run passes `-artifacts` (under `-outputdir`); otherwise the directory is temporary and removed after the test.
- **Go 1.27+: `go test` runs the `stdversion` vet check** — it flags any API newer than the module's `go` directive. A failure means bump the directive or stop using the newer API; it is not a check to silence.

## References

- [go-test.md](./references/go-test.md) — when selecting tests with `-run`/`-skip` regexes, chasing a flake (`-count`, `-shuffle`), or choosing `go test` flags.
- [helpers.md](./references/helpers.md) — when a test may hang and needs a per-test timeout that reports the caller's location.

## Cross-References

- → See `samber/cc-skills-golang@golang-stretchr-testify` skill for the testify API (assert, require, mock, suite)
- → See `samber/cc-skills-golang@golang-database` skill (testing.md) for database integration test patterns
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for CI test workflows and AI-driven code review in CI
- → See `samber/cc-skills-golang@golang-lint` skill for `thelper`, `paralleltest`, and `testifylint`, which enforce several rules above
