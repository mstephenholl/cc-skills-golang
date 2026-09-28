---
name: golang-documentation
description: "Golang documentation — godoc comments, Example tests, README, CONTRIBUTING, CHANGELOG, Go Playground links, and llms.txt. Use when writing or reviewing doc comments or project docs for a Go library, application, or CLI. Not for inline comment style (→ See `samber/cc-skills-golang@golang-code-style` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.6"
  openclaw:
    emoji: "📝"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch
paths:
  - "**/*.go"
---

**Persona:** You are a Go technical writer and API designer. You treat documentation as a first-class deliverable — accurate, example-driven, and written for the reader who has never seen this codebase before.

**Orchestration mode:** For documenting or reviewing a large codebase where many packages lack docs, fan out parallel sub-agents split by package — doc comments live in disjoint files, so packages edit without conflicts — and consolidate into one set of per-file findings or applied doc changes. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Write mode** — produce the documentation the user asked for. Compare the project against the [checklist](#documentation-checklist) and list the missing items in your reply without generating them unasked. Done when the requested docs exist and follow the Writing Principles.
- **Review mode** — the deliverable is per-file findings: file:line, what is missing or wrong, and a suggested rewrite; if the user asked for fixes, apply them. Parallelize by package only when many packages lack docs.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-documentation` skill takes precedence.

# Go Documentation

Write documentation that serves both humans and AI agents. Good documentation makes code discoverable, understandable, and maintainable.

## Cross-References

- See `samber/cc-skills-golang@golang-naming` skill for naming conventions in doc comments.
- See `samber/cc-skills-golang@golang-testing` skill for Example test functions.
- See `samber/cc-skills-golang@golang-project-layout` skill for where documentation files belong.
- See `samber/cc-skills@humanizer-en-asd-ste100` skill for strict, controlled English prose (ASD-STE100) when regulated or safety-critical documentation demands maximal clarity and unambiguity.

## Writing Principles

Apply to every piece of documentation you write or review:

**Concision** — write the shortest version that carries the idea. Remove ornament and hollow transitions. Never drop facts, warnings, or user-requested depth.

**Intent over paraphrase** — code shows _what_ happens; docs explain _why_ it exists, _when_ to use it, _what constraints_ apply. A comment that only restates the signature wastes the reader's time.

**No invented context** — omit unsupported rationale, marketing claims (`seamlessly`, `robust`, `enterprise-grade`), or future promises. Leave gaps visible rather than filling with speculation.

**Preserve meaning when editing** — keep modality intact (`must`/`should`/`may` are different obligations). Preserve conditions, warnings, required actions. A cleaner sentence that changes obligations is wrong.

**Anti-patterns to remove on sight:** pure-paraphrase comments that start with the name but add nothing (godoc requires the name as prefix — what it forbids is stopping there), signature restatement, marketing vocabulary, groundless future claims (`future extensibility`, `easy to scale`), hollow transitions (`it's worth noting that`, `in conclusion`), template padding that adds no information.

## Doc Comments

- **Document every exported identifier**, plus complex internal functions; skip test functions — their names are the documentation.
- **Start with the identifier's name and a verb phrase**, then cover why it exists, when to use it, its constraints (including concurrency safety), and the errors it returns.
- **Include parameters, return values, error cases, and a usage example for exported functions** — the doc comment is the API's only contract on pkg.go.dev; keep each section to what the signature doesn't already say.

Read [references/code-comments.md](./references/code-comments.md) when writing or reviewing a doc comment, package comment, file-level description, `Deprecated:` marker or `// Play:` link.

## Project Type

A **library** has no `main` package and is imported by others; an **application/CLI** has a `main` package or `cmd/` directory and ships a binary or image. A module can be both — importable `pkg/` packages get library docs, `cmd/` binaries get application docs, and `internal/` packages get doc comments only, since external users cannot import them. Application/CLI docs center on installation methods (prebuilt binaries, `go install`, Docker, Homebrew), `--help` text and configuration — users run the binary and never read its package API.

## Documentation Checklist

| Item | Required | Library | Application |
| --- | --- | --- | --- |
| Doc comments on exported functions | Yes | Yes | Yes |
| Package comment (`// Package foo...`) — MUST exist | Yes | Yes | Yes |
| README.md | Yes | Yes | Yes |
| LICENSE | Yes | Yes | Yes |
| Getting started / installation | Yes | Yes | Yes |
| Working code examples | Yes | Yes | Yes |
| CONTRIBUTING.md | Recommended | Yes | Yes |
| CHANGELOG.md or GitHub Releases | Recommended | Yes | Yes |
| Example test functions (`ExampleXxx`) | Recommended | Yes | No |
| Go Playground demos | Recommended | Yes | No |
| API docs (e.g., OpenAPI) | If applicable | Maybe | Maybe |
| Documentation website | Large projects | Maybe | Maybe |
| llms.txt | Recommended | Yes | Yes |

A private project might not need a documentation website, llms.txt, Go Playground demos...

## References

- [references/project-docs.md](./references/project-docs.md) — when writing or reordering a README, CONTRIBUTING.md or CHANGELOG, or documenting installation and distribution. Templates: [README](./assets/templates/README.md), [CONTRIBUTING](./assets/templates/CONTRIBUTING.md), [CHANGELOG](./assets/templates/CHANGELOG.md).
- [references/library.md](./references/library.md) — when documenting a library: `ExampleXxx` tests, Go Playground demos, pkg.go.dev rendering, a documentation website, llms.txt ([template](./assets/templates/llms.txt)), or discoverability registries.
- [references/application.md](./references/application.md) — when documenting an application or CLI: `--help` text, configuration (env vars, files, flags), architecture decision records, or REST/event/gRPC API docs.

To inspect how a published package renders its docs, symbols, and examples on pkg.go.dev, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill.
