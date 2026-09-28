# samber/oops — API Catalogue

Builder methods, terminal methods, per-layer examples, context propagation, accessors and output formats.

## Table of Contents

- [Builder methods](#builder-methods)
- [Terminal methods](#terminal-methods)
- [Per-layer examples](#per-layer-examples)
  - [Repository layer](#repository-layer)
  - [HTTP handler layer](#http-handler-layer)
  - [Context at each layer](#context-at-each-layer)
- [Context propagation](#context-propagation)
- [Panic recovery](#panic-recovery)
- [Reading error information](#reading-error-information)
- [Output formats](#output-formats)

## Builder methods

Every method returns a new `OopsErrorBuilder`, so a base builder can be extended in several branches without the branches affecting each other.

| Methods | Use case |
| --- | --- |
| `.With("key", value)` | Add custom key-value attribute (lazy `func() any` values supported) |
| `.WithContext(ctx, "key1", "key2")` | Extract values from Go context into attributes (lazy values supported) |
| `.In("domain")` | Set the feature/service/domain |
| `.Tags("auth", "sql")` | Add categorization tags (query with `err.HasTag("tag")`) |
| `.Code("iam_authz_missing_permission")` | Set machine-readable error identifier/slug |
| `.Public("Could not fetch user.")` | Set user-safe message (separate from technical details) |
| `.Hint("Runbook: https://doc.acme.org/doc/abcd.md")` | Add debugging hint for developers |
| `.Owner("team/slack")` | Identify responsible team/owner |
| `.User(id, "k", "v")` | Add user identifier and attributes |
| `.Tenant(id, "k", "v")` | Add tenant/organization context and attributes |
| `.Trace(id)` | Add trace / correlation ID (default: generated ULID, controlled by `oops.AutoTraceID`) |
| `.Span(id)` | Add span ID representing a unit of work/operation (default: ULID) |
| `.Time(t)` | Override error timestamp (default: `time.Now()`) |
| `.Since(t)` | Set duration based on time since `t` (exposed via `err.Duration()`) |
| `.Duration(d)` | Set explicit error duration |
| `.Request(req, includeBody)` | Attach `*http.Request` (optionally including body) |
| `.Response(res, includeBody)` | Attach `*http.Response` (optionally including body) |
| `oops.FromContext(ctx)` | Start from an `OopsErrorBuilder` stored in a Go context |

## Terminal methods

- `.Errorf(format, args...)` / `.New(message)` — create a new error
- `.Wrap(err)` — wrap an existing error; returns nil when `err` is nil
- `.Wrapf(err, format, args...)` — wrap with a message; returns nil when `err` is nil
- `.Join(err1, err2, ...)` — combine multiple errors
- `.Recover(fn)` / `.Recoverf(fn, format, args...)` — run `fn` and convert a panic into an error

## Per-layer examples

### Repository layer

```go
func (r *UserRepository) FetchUser(id string) (*User, error) {
    query := "SELECT * FROM users WHERE id = $1"
    row, err := r.db.Query(query, id)
    if err != nil {
        return nil, oops.
            In("user-repository").
            Tags("database", "postgres").
            With("query", query).
            With("user_id", id).
            Wrapf(err, "failed to fetch user from database")
    }
    // ...
}
```

### HTTP handler layer

```go
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
    userID := getUserID(r)

    err := h.service.CreateUser(r.Context(), userID)
    if err != nil {
        err = oops.
            In("http-handler").
            Tags("endpoint", "/users").
            Request(r, false).
            User(userID).
            Wrapf(err, "create user failed")
        http.Error(w, oops.GetPublic(err, "Internal server error"), http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
```

### Context at each layer

Each layer wraps with the context only it knows — the handler adds the request, the service the operation, the repository the query:

```go
func Controller() error {
    return oops.In("controller").Trace(traceID).Wrapf(Service(), "user request failed")
}

func Service() error {
    return oops.In("service").With("op", "create_user").Wrapf(Repository(), "db operation failed")
}

func Repository() error {
    return oops.In("repository").Tags("database", "postgres").Errorf("connection timeout")
}
```

## Context propagation

Store a pre-configured builder in the request context once, in middleware, so downstream code inherits trace ID, request and user without passing them through every signature:

```go
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        builder := oops.
            In("http").
            Request(r, false).
            Trace(r.Header.Get("X-Trace-ID"))

        ctx := oops.WithBuilder(r.Context(), builder)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

func handler(ctx context.Context) error {
    return oops.FromContext(ctx).Tags("handler", "users").Errorf("something failed")
}
```

## Panic recovery

`Recover` runs the callback and returns the panic as a structured error (nil when nothing panicked):

```go
func ProcessData(data string) (err error) {
    return oops.
        In("data-processor").
        Code("panic_recovered").
        Hint("Check input data format and dependencies").
        With("input_data", data).
        Recover(func() {
            riskyOperation(data)
        })
}
```

For a goroutine, call it inside the goroutine body — a recover in the parent never sees a child's panic:

```go
go func(item Item) {
    if err := oops.In("worker").With("item_id", item.ID).Recover(func() {
        process(item)
    }); err != nil {
        errCh <- err
    }
}(item)
```

## Reading error information

`oops.AsOops` unwraps like `errors.As`, whereas a bare `err.(oops.OopsError)` assertion misses an oops error wrapped by `fmt.Errorf("…: %w", err)`:

```go
if oopsErr, ok := oops.AsOops(err); ok {
    fmt.Println("Code:", oopsErr.Code())
    fmt.Println("Domain:", oopsErr.Domain())
    fmt.Println("Tags:", oopsErr.Tags())
    fmt.Println("Context:", oopsErr.Context())
    fmt.Println("Stacktrace:", oopsErr.Stacktrace())

    userID, userData := oopsErr.User()       // (string, map[string]any)
    tenantID, tenantData := oopsErr.Tenant() // (string, map[string]any)
}

// Get public-facing message with fallback
publicMsg := oops.GetPublic(err, "Something went wrong")
```

## Output formats

```go
fmt.Printf("%+v\n", err)       // verbose with stack trace
bytes, _ := json.Marshal(err)  // JSON for logging
slog.Error(err.Error(), slog.Any("error", err))  // slog integration via LogValuer
```
