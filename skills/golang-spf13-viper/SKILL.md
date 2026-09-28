---
name: golang-spf13-viper
description: "Golang layered configuration with spf13/viper — flag and env binding, config files, unmarshaling into structs, hot reload, and test isolation. Apply when using or adopting spf13/viper, or when the codebase imports `github.com/spf13/viper`. For command structure → See `samber/cc-skills-golang@golang-spf13-cobra` skill; for CLI architecture → See `samber/cc-skills-golang@golang-cli` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.1.4"
  openclaw:
    emoji: "🔧"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "1.21.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go engineer who treats configuration as a layered system. Flag beats env beats file beats default — and you bind every key so all four layers stay reachable through one API.

# Using spf13/viper for layered configuration in Go

Viper answers "what is the value of key X right now?" by walking its source layers from highest to lowest priority. It defines no commands or flags — cobra owns those, and a config-file daemon can use viper with no cobra at all.

**Official Resources:**

- [pkg.go.dev/github.com/spf13/viper](https://pkg.go.dev/github.com/spf13/viper)
- [github.com/spf13/viper](https://github.com/spf13/viper)

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

```bash
go get github.com/spf13/viper@latest
```

## The precedence pipeline

Viper resolves a key by walking sources in this order (first set value wins):

```
1. explicit Set()      — viper.Set("key", val)    highest priority
2. flag                — bound pflag.Flag the user actually passed
3. env var             — BindEnv / AutomaticEnv
4. config file         — ReadInConfig / MergeInConfig
5. KV remote           — etcd / Consul
6. default             — viper.SetDefault("key", val)
7. flag default        — a bound flag's default, only when nothing above set the key
```

The pipeline is fixed. A key that "should" come from the config file is usually shadowed by an env var or an explicitly passed flag; a flag the user didn't pass never shadows anything, since its default ranks below even `SetDefault`.

## Config files

```go
if err := viper.ReadInConfig(); err != nil {
    var notFound viper.ConfigFileNotFoundError // value type — a *ConfigFileNotFoundError target never matches
    if !errors.As(err, &notFound) {
        return fmt.Errorf("reading config: %w", err) // bad YAML, permissions: propagate
    }
}
```

Treat a missing config file as normal — a service that runs on flags and env alone shouldn't crash. `ConfigFileNotFoundError` comes only from the `SetConfigName` + `AddConfigPath` search; an explicit `SetConfigFile` path that doesn't exist returns an `fs.ErrNotExist` error instead, which usually should fail because the user asked for that file.

## Env binding

Wire all three together — missing any one breaks nested key resolution:

```go
viper.SetEnvPrefix("MYAPP")                             // PORT → MYAPP_PORT, no collisions
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))  // database.host → MYAPP_DATABASE_HOST
viper.AutomaticEnv()
// ✗ without the replacer, viper looks for MYAPP_DATABASE.HOST (dot preserved)
```

`AutomaticEnv` answers `Get` calls but does not register keys, and `Unmarshal` decodes only keys viper already knows. An env var whose key appears in no config file, `SetDefault`, `BindEnv` or bound flag is invisible to `Unmarshal` even though `GetBool` sees it — register the key, or build the instance with `viper.NewWithOptions(viper.ExperimentalBindStruct())` (v1.20+).

## Flag binding (the cobra seam)

Viper reads a bound flag's value and `Changed` state lazily, at each `Get` or `Unmarshal`, so the binding only has to exist before the first read — which rules out binding in `RunE` when `PersistentPreRunE` already unmarshaled the config. Bind each key once: when several subcommands bind their own local `--port` to the same key in `init()`, the last `BindPFlag` wins for every command, so bind those in the command's own `PreRunE`.

## Unmarshaling into structs

Give every field a `mapstructure` tag — mapstructure matches names case-insensitively but never maps `max_conn` to `MaxConn`. Prefer `UnmarshalKey("database", &dbCfg)` over `Sub("database").Unmarshal(...)`, because `Sub` returns nil (no error) when the key is absent and the method call panics.

Viper's default decoder already parses duration strings, splits comma-separated strings into slices, and decodes weakly typed input (`"true"` → `bool`). Passing `viper.DecodeHook(...)` to `Unmarshal`, `viper.WithDecodeHook(...)` to `NewWithOptions`, or assigning `dc.DecodeHook` replaces that chain rather than appending to it — a lone `net.IP` hook makes every duration fail to decode — so compose yours with `StringToTimeDurationHookFunc()`, importing `github.com/go-viper/mapstructure/v2` (viper v1.20+), not `mitchellh/mapstructure`.

## Hot reload

`WatchConfig` watches the config file's directory and re-reads on Write or Create events for that path, so an in-place write reloads on every platform — test reload with `os.WriteFile`, not an editor save. A Remove event for the file ends the watch for good. On macOS and BSD an atomic save that renames a temp file over the config (many editors, `sed -i`, config-management tools) arrives as Remove then Create, so the first such save silently kills hot reload; on Linux the rename arrives as Create and reloads.

The callback runs on the watcher goroutine, may fire more than once per save, and fires even when the re-read failed and the old values remain. Viper is not safe for concurrent use either — snapshot and validate a config struct inside the callback instead of calling `viper.Get` from request goroutines.

## Test isolation

Use `viper.New()` per test instead of the global — the global instance keeps config files, bindings and `Set` values across tests, so results depend on test order.

## Common Mistakes

| Mistake | Why it fails | Fix |
| --- | --- | --- |
| `AutomaticEnv` without `SetEnvKeyReplacer` | `database.host` looks for `MYAPP_DATABASE.HOST` (dot preserved) — never matches | Add `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))` |
| No `mapstructure` tags on struct fields | Silently misses nested and underscore-named fields | Add `mapstructure:"key_name"` to every field |
| Using global viper in tests | State from one test contaminates the next, causing flaky ordering | Create `viper.New()` per test |
| Missing `ConfigFileNotFoundError` check | Missing config file crashes a service that should run on flags/env alone | `errors.As(err, &notFound)` with `var notFound viper.ConfigFileNotFoundError` — propagate everything else |
| Env-only key missing after `Unmarshal` | `AutomaticEnv` doesn't register keys, so `Unmarshal` skips them | `SetDefault` or `BindEnv` the key, or `ExperimentalBindStruct()` |

## Further Reading

- [sources-and-formats.md](references/sources-and-formats.md) — read when choosing file formats (HCL, INI and properties need a codec since v1.20), searching several paths, merging a base and an override file, embedding defaults, or reading etcd/Consul
- [binding-and-env.md](references/binding-and-env.md) — read when an env var or flag doesn't resolve: `BindEnv` for third-party names, replacer mechanics, `AllowEmptyEnv`, binding timing
- [unmarshal.md](references/unmarshal.md) — read when decoding into structs: `UnmarshalKey`, custom decode hooks (`net.IP`), `squash`, `remain`
- [watch-and-reload.md](references/watch-and-reload.md) — read when adding hot reload: race-safe swap, debouncing, validating before applying
- [testing-and-isolation.md](references/testing-and-isolation.md) — read when testing config code: injecting `*viper.Viper`, `t.Setenv`, `Reset()` limits

## Cross-References

- → See `samber/cc-skills-golang@golang-cli` skill for general CLI architecture — project layout, exit codes, signal handling, cobra+viper integration
- → See `samber/cc-skills-golang@golang-spf13-cobra` skill for the cobra side of this integration (flag definition and binding)
- → See `samber/cc-skills-golang@golang-testing` skill for general Go testing patterns

If you encounter a bug or unexpected behavior in spf13/viper, open an issue at <https://github.com/spf13/viper/issues>.
