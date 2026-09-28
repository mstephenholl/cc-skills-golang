# godig setup and command examples

Read this when `godig` isn't installed or wired as an MCP server, or when you want a worked invocation for a command.

## Table of Contents

- [Install](#install)
- [Register the MCP server (optional)](#register-the-mcp-server-optional)
- [Command examples](#command-examples)

## Install

```bash
go install github.com/samber/godig/cmd/godig@latest
```

## Register the MCP server (optional)

`godig mcp` runs over **stdio** by default, or **streamable HTTP** with `--transport http`. The command is harness-agnostic — any MCP-capable host can point at it. Claude Code registers it via its own CLI:

stdio (the client launches godig on demand):

```bash
claude mcp add pkg-go-dev -- godig mcp
```

streamable HTTP (shared server at `/mcp`, default `:8080`):

```bash
godig mcp --transport http --addr :8080
claude mcp add --transport http pkg-go-dev http://localhost:8080/mcp
```

Hosted instance (no install needed) — a public server runs at `https://godig.samber.dev/mcp`:

```bash
claude mcp add --transport http pkg-go-dev https://godig.samber.dev/mcp
```

Other MCP-capable harnesses (Cursor, Windsurf, and others) each have their own MCP server registration — an entry in their respective settings file pointing at the same `godig mcp` command or hosted URL, not a shared config format.

The CLI and the MCP server expose the **same** operations under matching names. Prefer the CLI when `godig` is installed; the hosted instance is a fallback when it is not.

## Command examples

Always request Markdown output (`-o md`):

```bash
# Overview — start here (compact, one call)
godig overview github.com/samber/ro -o md

# Search
godig search "result option monad" --limit 5 -o md

# Package facets
godig package info github.com/samber/ro -o md
godig package imports github.com/samber/ro -o md
godig package doc github.com/samber/ro --format md -o md
godig package examples github.com/samber/ro --symbol Map -o md
godig package licenses github.com/samber/ro -o md

# Single symbol (token-efficient vs package-wide doc/examples)
godig symbol doc github.com/samber/lo Map -o md
godig symbol examples github.com/samber/oops OopsError.Error -o md

# Module facets
godig module info github.com/samber/ro -o md
godig module readme github.com/samber/ro -o raw
godig dependencies github.com/samber/ro -o md

# Lists (auto-paginated; --limit to cap)
godig versions github.com/samber/ro -o md
godig major-versions github.com/samber/lo -o md
godig packages github.com/samber/ro -o md
godig imported-by github.com/samber/ro --limit 20 -o md
godig symbols github.com/samber/ro --filter 'kind=="Function"' -o md

# Pin a version / set the build context
godig versions github.com/samber/ro --filter 'hasPrefix(version,"v0.3")' -o md
godig package doc github.com/samber/lo --version v1.50.0 -o md
godig symbols github.com/samber/ro --goos linux --goarch amd64 -o md

# Vulnerabilities
godig vulns github.com/samber/ro -o md
```
