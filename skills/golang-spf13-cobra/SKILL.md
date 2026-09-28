---
name: golang-spf13-cobra
description: "Golang CLI command trees with spf13/cobra — command hooks, argument validation, persistent vs local flags, shell completions, help and doc generation, and command tests. Apply when using or adopting spf13/cobra, or when the codebase imports `github.com/spf13/cobra`. For config layering → See `samber/cc-skills-golang@golang-spf13-viper` skill; for CLI architecture (exit codes, I/O, signals) → See `samber/cc-skills-golang@golang-cli` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.1.4"
  openclaw:
    emoji: "🐍"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.10.2"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go CLI engineer building command trees that feel native to the Unix shell. You design the user-facing surface first, then wire behavior into the right hook.

**Modes:**

- **Build** — a new command tree. Done when every command uses `RunE`, positional args are validated by `Args`, output goes through `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`, and a test drives a freshly built tree.
- **Extend** — adding subcommands, flags, or completions. Match the existing tree's constructors, groups and hook chain; done when the new command appears in `--help` and its test passes.
- **Review** — ranked findings with file:line against Common Mistakes and the hook gotchas below. If the user asked for fixes, apply them and re-verify.

# Using spf13/cobra for CLI command trees in Go

Cobra owns the command tree, flag parsing (via `pflag`), args validation, shell completions and doc generation. It does **not** handle configuration layering — that's viper's job.

**Official Resources:**

- [pkg.go.dev/github.com/spf13/cobra](https://pkg.go.dev/github.com/spf13/cobra)
- [github.com/spf13/cobra](https://github.com/spf13/cobra)
- [cobra.dev](https://cobra.dev)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get github.com/spf13/cobra@latest
```

## Cobra vs. viper

| Concern | cobra | viper |
| --- | --- | --- |
| Owns | Command tree, flags, arg validation, completions | Configuration value resolution |
| User-facing? | Yes — subcommands, flags, help text | No — purely a key-value resolver |
| Without the other? | Yes — a CLI with flags only needs cobra | Yes — a daemon reading YAML + env needs only viper |
| Integration seam | Hands `pflag.Flag` to viper via `BindPFlag` | Treats a flag the user set as its flag layer |

Use both only when you need both, binding at the root's `PersistentPreRunE` — → See `samber/cc-skills-golang@golang-spf13-viper` skill for the viper side.

## Hooks and execution order

```
Args validation → PersistentPreRunE → PreRunE → required-flag and flag-group checks → RunE → PostRunE → PersistentPostRunE
```

Execution stops at the first returned error, and only the `*E` variants can return one (see Common Mistakes).

- **A child `PersistentPreRunE` replaces the parent's** — cobra runs only the nearest persistent hook. Call `rootCmd.PersistentPreRunE(cmd, args)` from the child's hook, or set `cobra.EnableTraverseRunHooks = true` (cobra v1.8.0+) to run every ancestor's persistent hooks, root first for pre-run and child first for post-run. It is a package-level global, so it changes every command in the process, tests included.
- **`PostRunE` and `PersistentPostRunE` run only when `RunE` succeeded** — put cleanup that must also run on failure in a `defer` inside `RunE`.
- **Hooks run before required flags are checked** — a `PersistentPreRunE` can't assume `MarkFlagRequired` flags are set.

## Command groups

Register groups with `AddGroup` before the `AddCommand` calls whose commands set `GroupID` — cobra v1.6.0 panicked inside `AddCommand` otherwise, and later versions still panic at `Execute()` when a `GroupID` has no matching group on the parent.

## Common Mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| Using `Run` instead of `RunE` | Cannot return an error — only escape is `os.Exit` or panic, bypassing defers | Use `RunE` — return the error; `Execute()` returns it and `main()` exits non-zero |
| Writing `len(args)` checks in `RunE` | Bypasses cobra's standard error messages ("accepts 1 arg(s), received 2") | Declare `Args: cobra.ExactArgs(1)`; compose rules with `cobra.MatchAll(cobra.MinimumNArgs(1), cobra.OnlyValidArgs)` plus `ValidArgs` |
| Writing to `os.Stdout` / `os.Stderr` or calling `fmt.Println` | Tests cannot capture output — os-level file handles can't be redirected | Use `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`, which tests redirect with `SetOut` / `SetErr` |
| Reusing a root command across tests | Parsed flag values and `Changed` state persist; the second `Execute()` sees flags from the first | Build a fresh command tree per test from a `newRootCmd()` constructor |
| Full usage printed on every runtime error | Cobra prints usage after any returned error | `SilenceUsage: true` on the root — errors print one line, `--help` still prints usage |

## Further Reading

- [commands-and-args.md](references/commands-and-args.md) — read when wiring hooks across a hierarchy, composing `Args` validators, grouping, hiding or deprecating commands
- [flags.md](references/flags.md) — read when defining flags: `StringSlice` vs `StringArray` comma splitting, required/exclusive/one-required/required-together groups, custom `pflag.Value` types, `Changed()` for explicit zero values
- [completions.md](references/completions.md) — read when adding dynamic arg or flag-value completions (`ValidArgsFunction`, `RegisterFlagCompletionFunc`, `ShellCompDirective`) or testing them
- [generators.md](references/generators.md) — read when generating man/markdown docs, customizing help templates, or scaffolding with `cobra-cli`
- [testing.md](references/testing.md) — read when writing command tests: isolation, golden files, error paths, completion tests

## Cross-References

- → See `samber/cc-skills-golang@golang-cli` skill for general CLI architecture — project layout, exit codes, signal handling, I/O patterns
- → See `samber/cc-skills-golang@golang-spf13-viper` skill for configuration layering alongside cobra
- → See `samber/cc-skills-golang@golang-testing` skill for general Go testing patterns

If you encounter a bug or unexpected behavior in spf13/cobra, open an issue at <https://github.com/spf13/cobra/issues>.
