# samber/oops — Advanced Patterns

## Assertions

Use assertions for invariant checks (carefully — assertions panic):

```go
func ProcessPayment(amount int) error {
    return oops.
        In("payment-service").
        Recover(func() {
            oops.Assertf(amount > 0, "amount must be positive, got %d", amount)
            oops.Assert(amount < 1_000_000)
            // ... payment logic
        })
}
```

Assertions should be rare in Go. Use them only for truly impossible states that indicate a bug.

## Configuration

```go
oops.StackTraceMaxDepth = 20          // adjust stack trace depth (default 10)
oops.SourceFragmentsHidden = false    // show source code fragments (default true: hidden)
loc, _ := time.LoadLocation("America/New_York")
oops.Local = loc                      // set timezone for error timestamps (default UTC)
oops.AutoTraceID = false              // skip ULID trace generation when an external tracer always sets .Trace(id)
```

These are package-level variables: set them once at program start, before any error is built, rather than post-processing errors in a wrapper.

## Logger integration

`samber/oops` works with any logger. The error struct provides methods for extracting structured data:

```go
if oopsErr, ok := oops.AsOops(err); ok {
    userID, _ := oopsErr.User() // User() returns (id, attributes)
    logger.Error("operation failed",
        "code", oopsErr.Code(),
        "domain", oopsErr.Domain(),
        "user_id", userID,
        "error", oopsErr,
    )
}

// With slog
slog.Error(err.Error(), slog.Any("error", err))

// With zerolog (formatter in github.com/samber/oops/loggers/zerolog)
log.Error().Err(err).Msg("operation failed")

// With logrus (formatter in github.com/samber/oops/loggers/logrus)
log.WithError(err).Error("operation failed")
```
