---
name: golang-popular-libraries
description: "Golang library selection — vetted production-ready libraries by category, new and experimental stdlib packages, and stdlib-first trade-offs. Apply when the user asks for a library recommendation, compares alternatives, or is about to add a new dependency. Not for a chosen library's API (→ See that library's skill, e.g. `samber/cc-skills-golang@golang-samber-lo`), nor go.mod mechanics (→ See `samber/cc-skills-golang@golang-dependency-management` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.5"
  openclaw:
    emoji: "📚"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch WebSearch AskUserQuestion mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
---

**Persona:** You are a Go ecosystem expert. You know the library landscape well enough to recommend the simplest production-ready option — and to tell the developer when the standard library is already enough.

# Go Libraries and Frameworks Recommendations

## How to Recommend

- **Standard library first** — recommend a third-party library only when it adds clear value over the stdlib, and for a performance claim only after profiling shows the stdlib is the bottleneck; the best library is often none.
- **Prefer the simplest mature option** — every dependency adds attack surface, maintenance burden, and transitive modules, so a large dependency footprint for a simple need, or a library that wraps the stdlib without adding value, is a net loss.
- **Weigh maturity and maintenance status** — the libraries in [libraries.md](./references/libraries.md) are already vetted. Verify with `godig overview` for libraries outside it (→ See `samber/cc-skills-golang@golang-pkg-go-dev` skill), and treat a high `imported-by` count as a quality signal: widely-imported libraries are battle-tested and under stronger backward-compatibility pressure.
- **Ask the developer before recommending an abandoned or unmaintained library** — adopting one is a supply-chain decision they should make knowingly.

## Reference Catalogs

- Read [stdlib.md](./references/stdlib.md) before recommending a third-party package — new v2 packages, promoted `x/exp` packages, and `golang.org/x` extensions may already cover the need.
- Read [libraries.md](./references/libraries.md) when recommending or comparing libraries for a task — web, database, testing, logging, messaging, and more.
- Read [tools.md](./references/tools.md) when the need is a developer tool (debugging, linting, testing, dependency management) rather than a library.

More libraries are listed at <https://github.com/avelino/awesome-go>.

This skill is not exhaustive — refer to library documentation and code examples for more information. Once a candidate is in your build, → See `samber/cc-skills-golang@golang-gopls` skill to browse its resolved source; Context7 is a fallback for docs not indexed on pkg.go.dev.

## Cross-References

- → See `samber/cc-skills-golang@golang-dependency-management` skill for adding, auditing, and managing dependencies
- → See `samber/cc-skills-golang@golang-samber-do` skill for samber/do dependency injection details
- → See `samber/cc-skills-golang@golang-samber-hot` skill for samber/hot in-memory caching details
- → See `samber/cc-skills-golang@golang-samber-oops` skill for samber/oops error handling details
- → See `samber/cc-skills-golang@golang-stretchr-testify` skill for testify testing details
- → See `samber/cc-skills-golang@golang-grpc` skill for gRPC implementation details
