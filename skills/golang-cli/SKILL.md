---
name: golang-cli
description: "Golang CLI architecture — exit codes, stdout/stderr discipline, signal handling, version embedding, config layering, and CLI testing. Use when building or reviewing a Go command-line tool's structure or its behavior under scripts and pipes, or when the codebase imports urfave/cli. For cobra command APIs → See `samber/cc-skills-golang@golang-spf13-cobra` skill; for viper configuration → See `samber/cc-skills-golang@golang-spf13-viper` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.3"
  openclaw:
    emoji: "💻"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go CLI engineer. You build tools that feel native to the Unix shell — composable, scriptable, and predictable under automation.

**Modes:**

- **Build** — a new CLI. Done when `--help` renders, every failure exits non-zero with its message on stderr, config resolves from flag, env and file, and a command test captures output through `SetOut`.
- **Extend** — adding subcommands, flags, or completions. Match the existing command tree, file layout and flag-binding style; done when the new command shows in `--help`, its flags resolve through the existing config layering, and it has a test.
- **Review** — ranked findings with file:line against Common Mistakes and the exit-code and stream rules below. If the user asked for fixes, apply them and re-verify.

# Go CLI Best Practices

Default to Cobra for commands and flags plus Viper for layered configuration — the stack behind kubectl, gh and hugo. For a single-purpose tool with no subcommands and a handful of flags, stdlib `flag` is enough and saves two dependencies.

This skill owns the CLI's behavior under scripts and pipes; the library details live in sibling skills:

- → See `samber/cc-skills-golang@golang-spf13-cobra` skill for command hooks, flag groups, completions and positional-argument validation — set `Args: cobra.NoArgs`, `cobra.ExactArgs(n)` or `cobra.RangeArgs(min, max)` on the command instead of checking `len(args)` in `RunE`.
- → See `samber/cc-skills-golang@golang-spf13-viper` skill for the configuration precedence pipeline (flags beat env, env beats the config file, the file beats defaults), env binding and unmarshaling.

Runnable examples live in [assets/examples/](assets/examples/): `root.go` for the Cobra + Viper wiring, plus `flags.go`, `args.go`, `signal.go`, `version.go`, `exit_codes.go` and `cli_test.go`. Read [output.go](assets/examples/output.go) when adding an `--output` format flag or terminal-aware colors, and [completion.go](assets/examples/completion.go) when adding a `completion` command or custom flag and argument completions.

## Command layout

Put one file per command in `cmd/<app>/`, each registering itself with `rootCmd.AddCommand` in `init()`. Keep `main.go` to calling `Execute()` and turning its error into an exit code, and initialize config in the root command's `PersistentPreRunE` so every subcommand sees the same resolved values.

Set both on the root command:

- `SilenceUsage: true` — a runtime failure then prints one error line instead of the full help text; `--help` still prints usage.
- `SilenceErrors: true` — Cobra stops printing errors itself, so `main()` must print the error to stderr before exiting, or failures exit silently.

Make cross-cutting flags (`--config`, `--verbose`, `--log-level`) persistent on the root; keep a flag local when only one command reads it. Declare flag constraints with `MarkFlagRequired`, `MarkFlagsMutuallyExclusive` and `MarkFlagsOneRequired` rather than checks in `RunE`, and give enumerated values a `RegisterFlagCompletionFunc`.

Test commands in-process — `SetArgs`, `SetOut`/`SetErr` into a buffer, then `Execute()` on a freshly built tree (`cli_test.go`) — rather than running the compiled binary.

## Exit codes

| Code | Meaning                                  |
| ---- | ---------------------------------------- |
| 0    | Success                                  |
| 1    | Runtime failure                          |
| 2    | Usage error — unknown flag, bad argument |

Return errors from `RunE` and choose the code in `main()` — a typed error carrying the code (`ExitError{Code, Err}`) lets each command pick its category while deferred cleanup still runs. Cobra returns flag and argument errors as plain errors, so map them to 2 by wrapping them in your usage-error type with `rootCmd.SetFlagErrorFunc` (inherited by subcommands) and a wrapping `Args` validator.

## Version embedding

Inject version, commit and date with `-ldflags "-X main.version=…"` into package-level `var`s defaulting to `"dev"`. For `package main` the `-X` path is `main.version`, not the module import path — the linker silently ignores an unknown symbol, so a wrong path ships `"dev"`. Fall back to `debug.ReadBuildInfo()`, which carries the module version for binaries built with `go install module@version`, where no ldflags ran.

## I/O and signals

- Detect a terminal with `os.ModeCharDevice` on `Stat()` — drop colors and spinners when stdout isn't one (`fatih/color` does this for you), and never prompt when stdin isn't one.
- Wrap `cmd.Context()` in `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)` so cancellation reaches every call; `signal.go` bounds a server's `Shutdown` with a timeout. Call `stop()` as soon as `ctx.Done()` fires — until then a second Ctrl+C is swallowed instead of force-quitting a slow shutdown.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Writing to `os.Stdout` directly | Tests can't capture output. Use `cmd.OutOrStdout()` / `cmd.ErrOrStderr()`, which tests redirect with `SetOut` / `SetErr` |
| Calling `os.Exit()` inside `RunE` | Cobra's error handling, deferred functions, and cleanup code never run. Return an error, let `main()` decide |
| Logging to stdout | Unix pipes chain stdout — logs corrupt the data stream for the next program. Logs go to stderr |
| Not binding flags to Viper | Flags won't be configurable via env/config. Call `viper.BindPFlag` for every configurable flag |
| `AutomaticEnv()` without `SetEnvPrefix` and `SetEnvKeyReplacer` | Without a prefix, `PORT` collides with other tools; without a replacer, `log-level` looks up `MYAPP_LOG-LEVEL`, which no shell can export. Set both, replacing `-` and `.` with `_` |
| Config file required | Users without a config file get a crash. Ignore `viper.ConfigFileNotFoundError` — config should be optional |
| Hardcoded version string | Version gets out of sync with tags. Inject via `ldflags` at build time from git tags |
| No `--output` format | Scripts can't parse human-readable output. Add `--output table\|json\|plain`, defaulting to table — a boolean `--json` can't grow a third format |

## Related Skills

See `samber/cc-skills-golang@golang-project-layout`, `samber/cc-skills-golang@golang-dependency-injection`, `samber/cc-skills-golang@golang-testing`, `samber/cc-skills-golang@golang-design-patterns` skills.
