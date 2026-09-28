# Viper WatchConfig and Hot Reload

## Table of Contents

- [Basic setup](#basic-setup)
- [Which file events reload](#which-file-events-reload)
- [Race-safe reload pattern](#race-safe-reload-pattern)
- [Debouncing rapid changes](#debouncing-rapid-changes)
- [Validating config before applying](#validating-config-before-applying)
- [Stopping the watcher](#stopping-the-watcher)

## Basic setup

```go
viper.WatchConfig()
viper.OnConfigChange(func(e fsnotify.Event) {
    log.Printf("config changed: %s (op: %s)", e.Name, e.Op)
    // re-read affected values and apply them
})
```

`WatchConfig` starts a background goroutine that watches the config file using fsnotify. Call it after `ReadInConfig` — it needs the resolved config path.

## Which file events reload

Viper (v1.21) watches the config file's **directory** and filters events by path:

| Event on the config path | Effect |
| --- | --- |
| Write or Create | `ReadInConfig`, then `OnConfigChange` — even if the re-read failed |
| Symlink target changed (Kubernetes ConfigMap `..data` swap) | Reload, as for Write |
| Remove | **The watch loop exits for good** — later writes never reload |
| Rename, Chmod | Ignored |

Consequences:

- **Atomic saves differ by platform.** Renaming a temp file over the config (many editors, `sed -i`, config-management tools) arrives as Create on Linux (inotify) and reloads. On macOS and BSD (kqueue) it arrives as Remove then Create, so the first atomic save silently ends hot reload.
- **Delete-and-recreate deploys end the watch everywhere.** Replace the file with a rename, or mount a directory whose symlink is swapped.
- **Test with in-place writes** — `os.WriteFile("config.yaml", data, 0o644)` produces Write events on every platform.
- **A truncating write can fire twice** — once on the truncate, once on the data — so a reload may briefly see an empty file. Validate before applying (below).

## Race-safe reload pattern

`OnConfigChange` runs on the watcher goroutine, which has just rewritten viper's config map in `ReadInConfig`. Viper is not safe for concurrent use, so request goroutines calling `viper.Get*` race with every reload. Decode into a struct inside the callback and publish it under a lock:

```go
type Config struct {
    mu       sync.RWMutex
    LogLevel string `mapstructure:"log_level"`
    MaxConn  int    `mapstructure:"max_conn"`
}

var cfg Config

viper.OnConfigChange(func(e fsnotify.Event) {
    var newCfg Config
    if err := viper.Unmarshal(&newCfg); err != nil {
        log.Printf("error reloading config: %v", err)
        return  // keep old config on error
    }

    cfg.mu.Lock()
    cfg.LogLevel = newCfg.LogLevel
    cfg.MaxConn = newCfg.MaxConn
    cfg.mu.Unlock()

    log.Printf("config reloaded: log_level=%s", newCfg.LogLevel)
})
```

Readers take `cfg.mu.RLock()` and never call `viper.Get*` directly. An `atomic.Pointer[Config]` swapped in the callback works too and keeps readers lock-free.

## Debouncing rapid changes

Some filesystems fire multiple events per save. Debounce to avoid reloading multiple times:

```go
var reloadTimer *time.Timer
var reloadMu sync.Mutex

viper.OnConfigChange(func(e fsnotify.Event) {
    reloadMu.Lock()
    defer reloadMu.Unlock()
    if reloadTimer != nil {
        reloadTimer.Stop()
    }
    reloadTimer = time.AfterFunc(100*time.Millisecond, func() {
        applyNewConfig()
    })
})
```

## Validating config before applying

Always validate reloaded config before applying it — an invalid config mid-reload should keep the previous working config:

```go
viper.OnConfigChange(func(e fsnotify.Event) {
    var candidate Config
    if err := viper.Unmarshal(&candidate); err != nil {
        log.Printf("reload: invalid config, keeping previous: %v", err)
        return
    }
    if err := validate(candidate); err != nil {
        log.Printf("reload: validation failed, keeping previous: %v", err)
        return
    }
    applyConfig(candidate)
})
```

## Stopping the watcher

`WatchConfig` has no stop function. Its goroutine holds the viper instance and exits only on a Remove event for the config file or a watcher error — so design the watcher's lifetime to match the process. In tests, a watcher started on a file under `t.TempDir()` exits when the directory is cleaned up; without that it leaks for the rest of the test binary and goroutine-leak checkers will report it.
