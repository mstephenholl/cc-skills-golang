---
name: golang-gopls
description: "Golang semantic code navigation and refactoring with gopls — definitions, references, call and implementation hierarchy, symbol search, diagnostics, safe rename, and extract/inline. Use when locating code, finding call sites before a change, checking diagnostics after an edit, or renaming and extracting safely. Not for packages outside your go.mod (→ See `samber/cc-skills-golang@golang-pkg-go-dev` skill) or a whole-tree vulnerability audit (→ See `samber/cc-skills-golang@golang-security` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires the gopls binary (go install golang.org/x/tools/gopls@latest) v0.20+ on PATH.
metadata:
  author: samber
  version: "1.1.5"
  openclaw:
    emoji: "🛰️"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - gopls
    install:
      - kind: go
        package: golang.org/x/tools/gopls@latest
        bins: [gopls]
    skill-library-version: "0.22.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who reaches for semantic code intelligence instead of grep whenever a question is about the resolved build — grep finds text, `gopls` finds meaning (types, call graphs, shadowing, implementation relationships).

**Dependencies:** `gopls` — `go install golang.org/x/tools/gopls@latest` (v0.20+).

`gopls` is the official Go language server. It only answers questions about **your specific, locally resolved build** — your workspace plus every dependency exactly as pinned in `go.sum`, including `replace` directives. For a package outside that build (versions, docs, licenses, CVEs of something you haven't added yet) → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`); for a whole-tree vulnerability audit → See `samber/cc-skills-golang@golang-security` skill (`govulncheck`); for the full boundary with Context7 → See `samber/cc-skills-golang@golang-how-to` skill (Package lookup cluster).

## Three ways to reach gopls

Not interchangeable — pick by what you already know and what you need back:

- **gopls's own MCP server (`gopls mcp`)** — the default for agents: tools take names, file paths, and fuzzy queries instead of cursor positions. It runs headless over stdio and sees only files saved to disk.
- **A built-in LSP integration**, where the harness ships one (Claude Code does, off by default) — operations are keyed by line/character, so they're cheapest once a grep or read has given you a location. Its unique value: compiler diagnostics are pushed into context after every edit.
- **The `gopls` CLI** — `gopls <command> <file:line:col>`. The Go team documents it as experimental and debugging-only, so use it when neither of the above is wired, or for a one-shot scripted check. See [references/cli.md](references/cli.md).

When gopls isn't wired into the harness yet, read [references/mcp.md](references/mcp.md) for registration and the MCP tool list. When a capability you need isn't on the surface you have, look it up in [references/matrix.md](references/matrix.md) (capability → CLI → MCP → LSP).

## Use cases

- **Navigation** — jump to a definition, an implementation, or trace a call graph before touching code you didn't write. Details: [references/features.md](references/features.md#navigation).
- **Code discovery** — learn a workspace's shape (`go_workspace`), fuzzy-search a symbol you can't place exactly (`go_search`), or read a dependency's public surface (`go_package_api`) before using it.
- **Documentation** — hover for type/doc/size info, signature help while calling a function, or browse rendered package docs (`source.doc`, including internal packages pkg.go.dev never sees).
- **Diagnostics & safety** — compiler and analyzer errors after each edit, plus a lightweight `go_vulncheck` reachability check.
- **Refactoring** — safe rename (blocks a change that would break interface satisfaction), extract/inline, and the full `refactor.rewrite.*` family (fill struct/switch, invert if, remove unused parameter, add struct tags, implement interface). Full catalog with gotchas: [references/features.md](references/features.md#transformation).
- **Configuration** — read [references/settings.md](references/settings.md) when a result is missing because of build tags, `GOOS`, or directory filters, or when tuning analyzers and hints.

## Workflow notes

The MCP server ships its own Read and Edit workflows (`gopls mcp -instructions` prints them). These are what they leave out or overstate:

- **Diagnostics aren't pushed over MCP** — call `go_diagnostics` on the changed files after each edit; only a built-in LSP integration pushes them for you.
- **Scope `go test` to the changed packages while iterating** — `./...` slows the loop. A refactor whose blast radius crosses packages still needs a full run before commit (→ See `samber/cc-skills-golang@golang-refactoring` skill).
- **Run `go_vulncheck` only after a `go.mod` change or on a security task** — the server's instructions ask for it at every session start, which costs time on tasks that never touch dependencies.
- **On the CLI or an LSP integration**, keep the same order — references before changing a definition, diagnostics after each edit — and map each MCP tool through [references/matrix.md](references/matrix.md).

**Gotchas worth knowing before you rely on a result:**

- `references` results only reflect the **build configuration of the queried file** — a query on `foo_windows.go` will not surface matches in `bar_linux.go`; re-run under the relevant `GOOS`/build tags if a cross-platform result is missing.
- `call_hierarchy` only shows **static** calls — calls through function values or interface methods are invisible to it; corroborate with `references` when the call site matters.
- Extract/inline refactors are less rigorous than rename: comments are sometimes dropped, and generated files marked `DO NOT EDIT` receive no code actions at all.
- `refactor.rewrite.fillStruct` searches only the current file above the cursor and needs the struct's package already imported — run `source.organizeImports` first if the type was just typed in.

If you encounter a bug or unexpected behavior in gopls, open an issue at <https://github.com/golang/go/issues> with the title prefixed `x/tools/gopls:`.
