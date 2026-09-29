# Viper Env Binding and Flag Binding

## Table of Contents

- [The binding interaction model](#the-binding-interaction-model)
- [AutomaticEnv vs BindEnv](#automaticenv-vs-bindenv)
- [SetEnvKeyReplacer in depth](#setenvkeyreplacer-in-depth)
- [AllowEmptyEnv](#allowemptyenv)
- [Flag binding](#flag-binding)
- [How pflag binding interacts with precedence](#how-pflag-binding-interacts-with-precedence)
- [Debugging binding](#debugging-binding)

## The binding interaction model

Three settings control how viper maps environment variables to keys. They must be set together:

```go
viper.SetEnvPrefix("MYAPP")                           // adds MYAPP_ prefix
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // database.host → MYAPP_DATABASE_HOST
viper.AutomaticEnv()                                   // activates auto-binding
```

Call these before any `ReadInConfig` or `viper.Get*` call — typically in a root command's `PersistentPreRunE` or in `init()`.

## AutomaticEnv vs BindEnv

| Method | Behavior |
| --- | --- |
| `AutomaticEnv()` | Every key is automatically mapped to its env equivalent (with prefix and replacer applied) |
| `BindEnv(key, envVars...)` | Only the specified key is bound, to the specified env var name(s) |

Use `AutomaticEnv` for the common case. Use `BindEnv` when you need to bind to an env var with a name that doesn't follow your prefix/replacer convention (e.g., third-party env vars like `GOOGLE_APPLICATION_CREDENTIALS`).

```go
// Bind a specific non-prefixed env var
viper.BindEnv("google.credentials", "GOOGLE_APPLICATION_CREDENTIALS")
```

## SetEnvKeyReplacer in depth

Viper keys use `.` as separator for nested values. Env vars cannot contain dots. The replacer maps between them.

```go
// Config file:
// database:
//   host: localhost
//   max_conn: 25

// Without replacer:
viper.SetEnvPrefix("MYAPP")
viper.AutomaticEnv()
viper.GetString("database.host")  // looks for MYAPP_DATABASE.HOST — no match

// With replacer:
viper.SetEnvPrefix("MYAPP")
viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
viper.AutomaticEnv()
viper.GetString("database.host")  // looks for MYAPP_DATABASE_HOST — matches
```

Viper prepends the prefix and uppercases first, then applies the replacer to the whole name: `database.host` → `MYAPP_DATABASE.HOST` → `MYAPP_DATABASE_HOST`. The replacer therefore also rewrites the prefix — a `-` → `_` replacer turns prefix `MY-APP` into `MY_APP_…`.

## AllowEmptyEnv

By default, viper ignores env vars set to the empty string — the empty string is treated as "not set" and viper continues down the precedence stack. Override this behavior:

```go
viper.AllowEmptyEnv(true)
// now MYAPP_PORT="" → viper.GetInt("port") == 0, not the default
```

Use `AllowEmptyEnv`, not an `os.LookupEnv` + `viper.Set` workaround — `Set` writes to the override register, which sits above flags in viper's precedence, so the workaround would also silently beat an explicit `--flag`.

## Flag binding

Bind a pflag after defining it:

```go
func init() {
    rootCmd.PersistentFlags().Int("port", 8080, "listen port")
    viper.BindPFlag("port", rootCmd.PersistentFlags().Lookup("port"))
}
```

Bind an entire flag set:

```go
viper.BindPFlags(rootCmd.PersistentFlags())
```

**Timing rule:** Viper stores the `*pflag.Flag` and reads its `Changed` state and value lazily on every `Get`, so a binding made after parsing still works for later reads. What fails is reading before the binding exists — typically a `viper.Unmarshal` in the root's `PersistentPreRunE` while the subcommand binds in `RunE`. Bind in `init()` or `PersistentPreRunE`.

**One binding per key:** `BindPFlag` overwrites any earlier binding for the same key. When two subcommands each define a local `--port` and bind it to `"port"` in `init()`, the last `init()` wins for every command — running the other subcommand with `--port 9090` reads an unchanged flag and falls through to env, file or defaults. Bind such flags in the root's `PersistentPreRunE` via `cmd.Flags().Lookup(...)` — `cmd` is the command actually running — or give them distinct keys; a command's own `PreRunE` works only when nothing has read or unmarshaled the key before it runs.

## How pflag binding interacts with precedence

Viper checks `flag.Changed` (whether the user explicitly passed the flag). This is how it distinguishes between "flag default" (low priority) and "flag explicitly set" (high priority):

- `flag.Changed == false` (flag has its default): viper treats the flag as not present and falls through to env/file/default. The flag's own default is used only as a last resort — it loses even to `SetDefault`.
- `flag.Changed == true` (flag was provided on the command line): viper treats the flag value as the highest-priority source.

This means `viper.GetInt("port")` correctly returns the flag value when `--port 9090` is passed, and falls back to env `MYAPP_PORT` or config file `port: 8080` otherwise.

## Debugging binding

Print all resolved values to verify your binding is correct:

```go
fmt.Println(viper.AllSettings())
// map[database:map[host:localhost max_conn:25] port:8080]
```

Check env var resolution:

```go
os.Setenv("MYAPP_PORT", "9090")
viper.AutomaticEnv()
fmt.Println(viper.GetInt("port"))  // 9090
```
