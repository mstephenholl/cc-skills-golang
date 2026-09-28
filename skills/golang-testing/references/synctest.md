# testing/synctest — Deterministic Time and Goroutines

`testing/synctest` (Go 1.25+) runs a test inside a _bubble_ with a fake clock: time advances only when every goroutine in the bubble is durably blocked, so timers, tickers, deadlines, and context cancellation fire in a fixed order and the test runs in microseconds instead of real seconds.

## When to use it

| Situation | Use |
| --- | --- |
| Goroutines coordinating through `time.Sleep`, `time.After`, `time.Ticker`, or `context.WithTimeout` | `synctest.Test` |
| A test that is flaky because it races a real timer against a goroutine | `synctest.Test` — the ordering becomes deterministic, so the flake either disappears or reproduces every run |
| Single-goroutine logic that reads `time.Now()` (rate limiters, expiry, schedulers) | Inject a clock instead (→ [Time Mocking](./mocking.md#time-mocking)) — it works on every Go version and keeps the dependency visible in the API |

## Example

```go
func TestContextTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        const timeout = 5 * time.Second

        ctx, cancel := context.WithTimeout(t.Context(), timeout)
        defer cancel()

        time.Sleep(timeout - time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != nil {
            t.Fatalf("before timeout: %v", err)
        }

        time.Sleep(time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("after timeout: got %v, want DeadlineExceeded", err)
        }
    })
}
```

## Semantics inside the bubble

- `time.Sleep` returns as soon as every other goroutine in the bubble is blocked — the fake clock jumps forward instead of waiting.
- `time.After` and timers fire when the fake clock reaches their deadline.
- `synctest.Wait()` blocks until every other goroutine in the bubble is durably blocked; call it before asserting on state that background goroutines update.
- Goroutines blocked on I/O, a real network socket, or a mutex are not durably blocked, so the clock won't advance past them — keep real I/O out of the bubble.

## Version rules

- **Go 1.25+:** use `synctest.Test`. Do not write the Go 1.24 experimental `synctest.Run` API in Go 1.25+ code — it was an experiment, and `Test` replaced it.
- **Go 1.24 only:** a module that explicitly targets Go 1.24 and opts into `GOEXPERIMENT=synctest` may use `synctest.Run` as a compatibility fallback.
- **Go 1.27+:** `synctest.Sleep(d)` advances the bubble's clock; it is exactly `time.Sleep(d)` followed by `synctest.Wait()`.
- **Go 1.27+:** `httptest.NewTestServer(t, handler)` serves over an in-memory network, so HTTP server tests can run inside a bubble. It registers its own cleanup; send requests through `server.Client()` rather than a default client, and don't call `Start`/`StartTLS`.
