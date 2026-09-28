# Viper Unmarshal and Struct Mapping

## Table of Contents

- [Basic Unmarshal](#basic-unmarshal)
- [mapstructure tags](#mapstructure-tags)
- [UnmarshalKey — extracting a sub-tree](#unmarshalkey--extracting-a-sub-tree)
- [Env-only keys are invisible to Unmarshal](#env-only-keys-are-invisible-to-unmarshal)
- [Default decode hooks](#default-decode-hooks)
- [net.IP and custom types](#netip-and-custom-types)
- [Squash for embedded structs](#squash-for-embedded-structs)
- [Remain for unknown keys](#remain-for-unknown-keys)
- [Weak decoding](#weak-decoding)

## Basic Unmarshal

```go
type Config struct {
    Port     int    `mapstructure:"port"`
    Host     string `mapstructure:"host"`
    LogLevel string `mapstructure:"log_level"`
    Database struct {
        DSN     string `mapstructure:"dsn"`
        MaxConn int    `mapstructure:"max_conn"`
    } `mapstructure:"database"`
}

var cfg Config
if err := viper.Unmarshal(&cfg); err != nil {
    return fmt.Errorf("decoding config: %w", err)
}
```

## mapstructure tags

Always use `mapstructure` tags. Without them, mapstructure falls back to case-insensitive field name matching, which works for simple cases but silently fails for:

- Nested structs where the outer key uses an underscore (`max_conn` → `MaxConn`)
- Unexported fields
- Fields where the Go name does not match the config key

```go
// ✓ Good — explicit and immune to rename surprises
type TLSConfig struct {
    CertFile string `mapstructure:"cert_file"`
    KeyFile  string `mapstructure:"key_file"`
    Enabled  bool   `mapstructure:"enabled"`
}

// ✗ Fragile — relies on case-folding; breaks when config key uses underscores
type TLSConfig struct {
    CertFile string  // viper key "certfile" or "CertFile", not "cert_file"
    KeyFile  string
    Enabled  bool
}
```

## UnmarshalKey — extracting a sub-tree

```go
type DatabaseConfig struct {
    DSN     string `mapstructure:"dsn"`
    MaxConn int    `mapstructure:"max_conn"`
}

var dbCfg DatabaseConfig
if err := viper.UnmarshalKey("database", &dbCfg); err != nil {
    return fmt.Errorf("decoding database config: %w", err)
}
```

Prefer `UnmarshalKey` over `viper.Sub` + `Unmarshal` — fewer nil checks and less boilerplate.

## Env-only keys are invisible to Unmarshal

`Unmarshal` decodes only the keys viper already knows — keys from a config file, `SetDefault`, `BindEnv` or a bound flag. `AutomaticEnv` makes `Get` consult the environment but registers no keys, so with no config entry for `enabled`, `MYAPP_ENABLED=true` gives `viper.GetBool("enabled") == true` while `Unmarshal` leaves `cfg.Enabled` false.

```go
// ✓ Register every key that may come only from env
viper.SetDefault("enabled", false)
viper.BindEnv("database.password") // honours prefix and replacer

// ✓ Or let viper derive keys from the struct (experimental, v1.20+)
v := viper.NewWithOptions(viper.ExperimentalBindStruct())
```

## Default decode hooks

Viper's default decoder config already includes `StringToTimeDurationHookFunc()` (`"1h30m"` → `time.Duration`), a comma-splitting string → slice hook (`"a,b"` → `[]string{"a", "b"}`), and `WeaklyTypedInput: true` (`"true"` → `bool`, `"8080"` → `int`). Tagged `time.Duration`, slice and bool fields therefore decode without extra configuration.

Passing `viper.DecodeHook(h)` to `Unmarshal`, `viper.WithDecodeHook(h)` to `NewWithOptions`, or assigning `dc.DecodeHook` replaces those defaults instead of appending — with only an IP hook, `timeout: 30s` fails with `'timeout' cannot parse value as 'time.Duration'`. Compose with the existing `dc.DecodeHook` inside a decoder option (below), or list `StringToTimeDurationHookFunc()` and `StringToSliceHookFunc(",")` yourself when using `DecodeHook` / `WithDecodeHook`. Since viper v1.20 the package is `github.com/go-viper/mapstructure/v2`; the archived `github.com/mitchellh/mapstructure` types don't match viper's `DecoderConfigOption` and fail to compile.

```go
import "github.com/go-viper/mapstructure/v2"

err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
        dc.DecodeHook,                         // keep viper's duration/slice hooks
        mapstructure.StringToIPHookFunc(),     // add net.IP
    )
})
```

## net.IP and custom types

mapstructure v2 ships hooks for `net.IP`, `net.IPNet`, `netip.Addr`, `netip.AddrPort` and `netip.Prefix`. Write your own only for domain types:

```go
import "github.com/go-viper/mapstructure/v2"

func stringToIPHookFunc() mapstructure.DecodeHookFunc {
    return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
        if f.Kind() != reflect.String || t != reflect.TypeOf(net.IP{}) {
            return data, nil
        }
        ip := net.ParseIP(data.(string))
        if ip == nil {
            return nil, fmt.Errorf("invalid IP address: %s", data)
        }
        return ip, nil
    }
}
```

## Squash for embedded structs

```go
type BaseConfig struct {
    LogLevel string `mapstructure:"log_level"`
    Debug    bool   `mapstructure:"debug"`
}

type ServerConfig struct {
    BaseConfig `mapstructure:",squash"`  // merge BaseConfig fields at this level
    Port       int    `mapstructure:"port"`
}
```

Without `,squash`, the base config must be nested under a `baseconfig` key in the config file.

## Remain for unknown keys

```go
type Config struct {
    Port     int                    `mapstructure:"port"`
    Remain   map[string]interface{} `mapstructure:",remain"`
}
```

Extra keys from the config file are collected in `Remain` instead of being silently dropped. Useful for forward compatibility.

## Weak decoding

Viper enables `WeaklyTypedInput` by default, so env strings decode into bools and numbers. Turn it off for strict decoding when a silently converted value would hide a config typo:

```go
err := viper.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
    dc.WeaklyTypedInput = false
})
```
