---
name: golang-stretchr-testify
description: "Golang testing with stretchr/testify — assert vs require, mocks and argument matchers, suites, and Eventually/JSONEq-style assertions. Apply when writing tests with testify, or when the codebase imports `github.com/stretchr/testify`. For general Go testing patterns → See `samber/cc-skills-golang@golang-testing` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.5"
  openclaw:
    emoji: "✅"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - gotests
    install:
      - kind: go
        package: github.com/cweill/gotests/...@latest
        bins: [gotests]
    skill-library-version: "1.11.1"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(gotests:*) AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who treats tests as executable specifications. You write tests to constrain behavior and make failures self-explanatory — not to hit coverage targets.

**Modes:**

- **Write mode** — adding tests or mocks. Done when the tests pass, every mock's expectations are asserted, and each failure message names the case that broke.
- **Review mode** — auditing test code for testify misuse. Deliverable: findings with file:line, drawn from the traps below; fixes applied if the user asked for them.

# stretchr/testify

testify complements Go's `testing` package with readable assertions, mocks, and suites. It does not replace `testing` — `*testing.T` stays the entry point.

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

## assert vs require

Both packages offer identical assertions and differ only on failure: `assert` records it and continues, so one run shows every failure; `require` calls `t.FailNow()`. Use `require` for preconditions (setup, error checks, nil guards), where continuing would panic or mislead, and `assert` for the verifications that follow.

Bind them once per test as `is := assert.New(t)` and `must := require.New(t)`:

```go
func TestParseConfig(t *testing.T) {
    is := assert.New(t)
    must := require.New(t)

    cfg, err := ParseConfig("testdata/valid.yaml")
    must.NoError(err)    // stop if parsing fails — cfg would be nil
    must.NotNil(cfg)

    is.Equal("production", cfg.Environment)
    is.Equal(8080, cfg.Port)
}
```

### Trap: a parent-scoped `is` inside subtests

`assert.New(t)` captures the exact `*testing.T` it was built with. Reused inside a `t.Run` closure, a parent-scoped instance reports every subtest failure against the _parent_, while the failing subtest itself prints `--- PASS` — whether or not the subtest calls `t.Parallel()`.

```go
// ✗ Bad — is is bound to the parent's t; failures are misattributed
func TestCalculatePrice(t *testing.T) {
    is := assert.New(t)
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            is.Equal(tt.expected, CalculatePrice(tt.quantity, tt.unitPrice))
        })
    }
}

// ✓ Good — each subtest builds its own instance from its own t
func TestCalculatePrice(t *testing.T) {
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            is := assert.New(t)
            is.Equal(tt.expected, CalculatePrice(tt.quantity, tt.unitPrice))
        })
    }
}
```

**Diagnose:** 1- `go test -v -run '^TestName$'` with a deliberately broken case — `--- FAIL: TestName` above a `--- PASS: TestName/<case>` line for the broken case means the assert scope is leaking

## Equality semantics

| Assertion | Compares | Use when |
| --- | --- | --- |
| `Equal` | `reflect.DeepEqual`, types must match exactly; pointers compare the values they point to, not addresses | Default |
| `EqualValues` | Converts to a common type first (`int32(1)` equals `int64(1)`) | Comparing across numeric or named types |
| `EqualExportedValues` | Exported fields only | Structs whose unexported fields (caches, mutexes, timestamps) legitimately differ |
| `Same` / `NotSame` | Pointer identity | You need "the same object", not "an equal object" |

Prefer the assertion that states the intent over a hand-rolled `Equal`, since its failure message then names the actual problem: `JSONEq`/`YAMLEq` (documents equivalent despite key order and whitespace), `ElementsMatch` (same elements, any order), `InDelta` (floats within a tolerance), `WithinDuration` (times within a window), `Regexp` (string matches a pattern), `Eventually`/`EventuallyWithT` (async condition, polled until a timeout) and `Same`/`NotSame` (pointer identity).

## Polling with EventuallyWithT

`Eventually` takes a `func() bool`, so a timeout reports only "condition never satisfied". `EventuallyWithT` hands the callback a `*assert.CollectT`; call the package-level assertions on it, `assert.Equal(c, …)`, not on `t` or a pre-built `is`. An assertion on `t` fails the test on the first unsuccessful poll instead of retrying, and the last tick's failures are what the timeout reports.

```go
is.EventuallyWithT(func(c *assert.CollectT) {
    resp, err := client.GetOrder(orderID)
    assert.NoError(c, err)
    assert.Equal(c, "shipped", resp.Status)
}, 10*time.Second, 500*time.Millisecond)
```

## testify/mock

Mock interfaces to isolate the unit under test: embed `mock.Mock`, implement each method with `m.Called()`, and finish with `m.AssertExpectations(t)`. Key matchers are `mock.Anything`, `mock.AnythingOfType("T")`, and `mock.MatchedBy(func)` for a predicate on some fields of an argument. Call modifiers are `.Once()`, `.Times(n)`, `.Maybe()`, and `.Run(func)`.

Read [mock.md](./references/mock.md) when defining mocks, returning different values per call (retry tests), or verifying specific calls.

## testify/suite

Suites group tests that share setup. `SetupSuite`/`TearDownSuite` run once per suite (shared connections, containers), and `SetupTest`/`TearDownTest` run around every test method (fresh mocks, table truncation).

```go
type TokenServiceSuite struct {
    suite.Suite
    store   *MockTokenStore
    service *TokenService
}

func (s *TokenServiceSuite) SetupTest() {
    s.store = new(MockTokenStore)
    s.service = NewTokenService(s.store)
}

func (s *TokenServiceSuite) TestGenerate_ReturnsValidToken() {
    s.store.On("Save", mock.Anything, mock.Anything).Return(nil)
    token, err := s.service.Generate("user-42")
    s.Require().NoError(err)
    s.NotEmpty(token)
    s.store.AssertExpectations(s.T())
}

// Required launcher — without it, no suite test runs
func TestTokenServiceSuite(t *testing.T) {
    suite.Run(t, new(TokenServiceSuite))
}
```

Suite methods such as `s.Equal()` and `s.NotNil()` behave like `assert` and continue after a failure. For fail-fast, go through `s.Require()`, e.g. `s.Require().NotNil(conn)`.

## Common Mistakes

- **Forgetting `AssertExpectations(t)`** — expectations set with `On()` are never checked, so the test passes even if the method was never called.
- **`is.Equal(ErrNotFound, err)`** — fails once the error is wrapped. Use `is.ErrorIs(err, ErrNotFound)`, which walks the chain, and note its order is `(err, target)`, the reverse of `Equal`.
- **Swapped argument order** — `Equal` and friends take `(expected, actual)`. Swapping them prints backwards diffs.
- **`assert` for guards** — the test continues after the failed guard and panics on the nil dereference. Use `require`.
- **Missing `suite.Run()`** — without the launcher function, zero tests execute, silently.
- **Blaming pointers when `Equal` fails on two `*User`** — `Equal` already dereferences pointers. A mismatch between equal-looking structs comes from unexported fields, `time.Time` monotonic readings or locations, or differing concrete types. Print the diff, then compare with `EqualExportedValues` or assert the fields that matter.

## Linters

`testifylint` catches swapped expected/actual, `assert` where `require` should guard an error, and `Equal(true, x)` instead of `True(x)` — mechanical patterns that are cheaper to lint than to catch in review. → See `samber/cc-skills-golang@golang-lint` skill for configuration.

## Cross-References

- → See `samber/cc-skills-golang@golang-testing` skill for general test patterns, table-driven tests, and CI

If you encounter a bug or unexpected behavior in stretchr/testify, open an issue at <https://github.com/stretchr/testify/issues>.
