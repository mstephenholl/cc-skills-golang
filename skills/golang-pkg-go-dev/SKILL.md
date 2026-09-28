---
name: golang-pkg-go-dev
description: "Golang module lookup on pkg.go.dev via godig — versions, docs and symbols, examples, licenses, known CVEs, and importers of any published Go package; preferred over Context7 for Go modules. Use when looking up a module not yet in go.mod, or a published package's versions, license, vulnerabilities, or importers. Not for upgrading (→ See `samber/cc-skills-golang@golang-dependency-management` skill), choosing a library (→ See `samber/cc-skills-golang@golang-popular-libraries` skill), or code in your build (→ See `samber/cc-skills-golang@golang-gopls` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires the godig CLI (go install github.com/samber/godig/cmd/godig@latest) or access to a godig MCP server, and internet access to reach the pkg.go.dev API.
metadata:
  author: samber
  version: "1.4.4"
  openclaw:
    emoji: "🔎"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - godig
    install:
      - kind: go
        package: github.com/samber/godig/cmd/godig@latest
        bins: [godig]
    skill-library-version: "0.2.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Bash(godig:*) Agent
---

# golang-pkg-go-dev

**Dependencies:** `godig` — `go install github.com/samber/godig/cmd/godig@latest`, or a registered godig MCP server. When neither is available, read [setup.md](references/setup.md) for install and MCP registration (including a hosted instance).

`godig` queries the [pkg.go.dev](https://pkg.go.dev) API for docs, symbols, versions, importers, licenses, and vulnerabilities of any published Go package — including ones not yet in your `go.mod`. It works as a CLI and as an MCP server with the same operations under matching names; every operation is **read-only** and needs no authentication.

For code in your locally resolved build (`go.sum`, `replace`d forks, call sites in your repo) → See `samber/cc-skills-golang@golang-gopls` skill. For when to use `godig` over Context7 or `govulncheck` → See `samber/cc-skills-golang@golang-how-to` skill (Package lookup cluster).

## Commands

**Global flags (all commands):** `-o/--output table|json|raw|md` (default `table` — pass `-o md` for chat), `--base-url` (pkg.go.dev API), `--vuln-base-url` (Go vulnerability database, consulted by `vulns` and `overview`), `--timeout`, `--log-level debug|info|warn|error|off`. All are also settable via `GODIG_*` env vars.

| Command | Args | Specific flags | Purpose |
| --- | --- | --- | --- |
| `overview` | `<path>` | `--version` | Compact summary (metadata, versions, licenses, vulns) — start here |
| `search` | `<query>` | `--symbol --limit --filter` | Find packages (optionally exporting a symbol) |
| `package info` | `<path>` | `--module --version` | Package metadata |
| `package imports` | `<path>` | `--module --version` | Packages this package imports (plain list) |
| `package doc` | `<path>` | `--module --version --goos --goarch --format md\|text\|html\|markdown` | Full package doc (LARGE) |
| `package examples` | `<path>` | `--module --version --goos --goarch --symbol` | Runnable examples (LARGE; scope with `--symbol`) |
| `package licenses` | `<path>` | `--module --version` | License files, full text (LARGE) |
| `symbol doc` | `<path> <symbol>` | `--module --version --goos --goarch` | One symbol's signature + doc (token-efficient) |
| `symbol examples` | `<path> <symbol>` | `--module --version --goos --goarch` | One symbol's runnable examples |
| `symbols` | `<path>` | `--module --version --goos --goarch --limit --filter` | List exported symbols |
| `module info` | `<path>` | `--version` | Module metadata |
| `module licenses` | `<path>` | `--version` | Module license files (LARGE) |
| `module readme` | `<path>` | `--version` | Module README, full Markdown (LARGE) |
| `dependencies` | `<path>` | `--version` | go.mod deps: requires / replaces / excludes / go directive |
| `packages` | `<path>` | `--version --limit --filter` | Packages contained in a module |
| `versions` | `<path>` | `--limit --filter` | All versions, newest first |
| `major-versions` | `<path>` | `--limit --filter --exclude-pseudo` | Major versions (v1, v2 …) living as separate modules |
| `imported-by` | `<path>` | `--module --version --limit --filter` | Packages that import this one |
| `vulns` | `<path>` | `--version --limit` | Known vulnerabilities (from the Go vuln DB) |
| `mcp` | — | `--transport stdio\|http --addr --cache-ttl --cache-size` | Run as an MCP server |
| `version` | — | — | Print godig version / commit / build date |

**Exit codes:** `0` success, `1` runtime error (network, package not found), `2` usage error — a missing/invalid argument or flag (e.g. a non-positive `--limit`), or a command group invoked with no subcommand (`godig package`). Check for `2` to tell a malformed call apart from a failed lookup.

Read [sample-output.md](references/sample-output.md) when you need the shape of a command's `-o md` output before parsing it.

### Tips

- **Start with `overview`** — one call returns a compact summary (metadata, latest + recent versions, license types, vulnerabilities). Reach for `doc`/`examples`/`module readme`/`licenses` (LARGE) only when the full text is needed.
- **Always pass `-o md`** so results render as Markdown (tables, or raw doc/README) in the chat. Other formats exist (`table` default, `json`, `raw`) but prefer `md` here.
- `<path>` is a full import path, e.g. `github.com/samber/lo` — pass it as the positional argument.
- `--version` pins a specific module version (`v1.5.0`, `latest`, `master`, `main`); `--module` disambiguates which module a package belongs to.
- `--filter` narrows list results server-side with a Go boolean expression — see [Filter syntax](#filter-syntax).
- `--goos`/`--goarch` set the documentation/symbols build context (e.g. `linux`/`amd64`).
- Prefer `symbol doc`/`symbol examples` over the package-wide `package doc`/`package examples` when you only need one symbol — far fewer tokens.
- **Parallelize independent lookups** — every command is a self-contained, read-only HTTP query, so issue the calls for several symbols, packages, or modules in one turn rather than one after another. For a large fan-out (documenting many symbols, comparing many candidate libraries, auditing CVEs across a dependency set), hand the calls to parallel sub-agents that each return a compact summary, so the raw LARGE output never lands in the main context.
- Listing commands auto-paginate (return all results); use `--limit` to cap.

### Filter syntax

`--filter` (on `search`, `versions`, `major-versions`, `packages`, `imported-by`, `symbols`) takes a **Go boolean expression evaluated server-side, once per result item**. It is not a regex — wrap the whole expression in single quotes for the shell.

- **Identifiers are the item's fields, which differ per command** — a field valid for one list is rejected by another (e.g. `search` exposes `packagePath`, not `path`). An unknown field fails with `undefined identifier: <name>` (HTTP 400), which names the offending field. Fields use the item's lowercase JSON key; the exception is enum-like values such as `kind`, which are capitalized (`Function`, not `func`).
- **Operators**: `==` `!=` `<` `<=` `>` `>=`, boolean `&&` `||` `!`, parentheses for grouping.
- **String functions**: `contains(s, sub)`, `hasPrefix(s, pre)`, `hasSuffix(s, suf)`.
- **Literals**: double-quoted strings (`"Function"`), `true`/`false`, numbers.

Filterable fields per command (string unless noted):

| Command | Fields |
| --- | --- |
| `search` | `modulePath`, `packagePath`, `synopsis`, `version` |
| `versions` | `version`, `modulePath`, `deprecated` (bool), `retracted` (bool), `hasGoMod` (bool), `commitTime` |
| `packages` | `path`, `name`, `synopsis`, `isRedistributable` (bool) |
| `imported-by` | `path` (the importing package path) |
| `symbols` | `name`, `kind` (`Function`/`Method`/`Type`/`Variable`/`Constant`), `synopsis`, `parent` |
| `major-versions` | `modulePath`, `major`, `version`, `isLatest` (bool) |

```bash
godig symbols github.com/samber/lo --filter 'kind=="Function"' -o md
godig symbols github.com/samber/lo --filter 'kind=="Function" && hasPrefix(name,"Map")' -o md
godig versions github.com/samber/lo --filter 'hasPrefix(version,"v1.5")' -o md
godig versions github.com/samber/lo --filter 'deprecated==false && retracted==false' -o md
godig search "result option" --filter 'hasPrefix(packagePath,"github.com/samber/")' -o md
```

---

This skill is not exhaustive. `godig --help` and each sub-command's `--help` list current flags and output formats; the data mirrors what [pkg.go.dev](https://pkg.go.dev) exposes.

If you encounter a bug or unexpected behavior in `godig`, open an issue at <https://github.com/samber/godig/issues>.
