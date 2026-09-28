---
name: golang-samber-oops
description: "Structured error handling in Golang with samber/oops — error builders, stack traces, error codes, error context, error wrapping, error attributes, user-facing vs developer messages, panic recovery, and logger integration. Apply when using or adopting samber/oops, or when the codebase already imports github.com/samber/oops."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.3"
  openclaw:
    emoji: "💥"
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

**Persona:** You are a Go engineer who treats errors as structured data. Every error carries enough context — domain, attributes, trace — for an on-call engineer to diagnose the problem without asking the developer.

# samber/oops Structured Error Handling

**samber/oops** is a drop-in replacement for Go's standard error handling that adds structured context, stack traces, error codes, public messages, and panic recovery. Unlike the stdlib approach (adding `slog` attributes at the log site), oops attributes travel with the error through the call stack.

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

## Build errors with a reusable builder

Chain context onto an `oops` builder, then end with a terminal method (`Errorf`, `New`, `Wrap`, `Wrapf`, `Join`, `Recover`). Each builder method returns a new builder, so shared context set once at the top of a function can be extended per error path:

```go
func (s *UserService) CreateOrder(ctx context.Context, req CreateOrderRequest) error {
    builder := oops.
        In("order-service").
        Tags("orders", "checkout").
        Tenant(req.TenantID, "plan", req.Plan).
        User(req.UserID, "email", req.UserEmail)

    product, err := s.catalog.GetProduct(ctx, req.ProductID)
    if err != nil {
        return builder.
            With("product_id", req.ProductID).
            Wrapf(err, "product lookup failed")
    }

    if product.Stock < req.Quantity {
        return builder.
            Code("insufficient_stock").
            Public("Not enough items in stock.").
            With("product_id", req.ProductID).
            With("requested", req.Quantity).
            With("available", product.Stock).
            Errorf("insufficient stock")
    }

    return nil
}
```

Prefer the dedicated methods over generic `.With()` where they exist — `.User()`, `.Tenant()`, `.Request()`, `.Trace()` — because oops emits them as dedicated `user`, `tenant`, `request` and `trace` fields in logs and JSON.

## Rules

- **Keep error messages low-cardinality** — interpolating IDs or values into the message makes every occurrence unique, which breaks grouping in Datadog, Loki and Sentry; put variable data in `.With()` and keep the message static.

  ```go
  // ✗ Bad — high-cardinality, breaks APM grouping
  oops.Errorf("failed to process user %s in tenant %s", userID, tenantID)

  // ✓ Good — static message + structured attributes
  oops.With("user_id", userID).With("tenant_id", tenantID).Errorf("failed to process user")
  ```

- **Wrap without a nil check** — `Wrap` and `Wrapf` return nil when `err` is nil, so `return oops.In("processor").Wrapf(err, "fetch failed")` replaces the `if err != nil { … } return nil` block.
- **Add context once per package boundary** — each layer wraps with what only it knows (handler: `.Request()`; service: the operation; repository: the query), not at every function call.
- **Set request-wide context once in middleware** — store a builder with `oops.WithBuilder(ctx, builder)` and start downstream errors from `oops.FromContext(ctx)`, so trace ID, request and user reach every error without extra parameters.
- **Wrap goroutine bodies with `oops.Recover` where a panic would crash the process** — call it inside the goroutine, since a recover in the parent never sees a child's panic, and attach `.In()`, `.Code()` or `.Hint()` so the recovered error is diagnosable.
- **Read attributes with `oops.AsOops(err)`** — it unwraps like `errors.As`, whereas a bare `err.(oops.OopsError)` misses an oops error wrapped by `fmt.Errorf("%w")`; use `oops.GetPublic(err, fallback)` for the user-facing message.

## References

- [references/api.md](references/api.md) — when you need the full builder-method table, terminal methods, per-layer examples (repository, handler, middleware), a panic-recovery example, accessors such as `User()`/`Tenant()`, or output formats.
- [references/advanced.md](references/advanced.md) — when checking invariants with `oops.Assert`, tuning stack depth, source fragments, timestamps or trace IDs, or wiring a zerolog/logrus/zap formatter.

Library documentation: [github.com/samber/oops](https://github.com/samber/oops), [pkg.go.dev/github.com/samber/oops](https://pkg.go.dev/github.com/samber/oops).

## Cross-References

- → See `samber/cc-skills-golang@golang-error-handling` skill for general error handling patterns
- → See `samber/cc-skills-golang@golang-observability` skill for logger integration and structured logging

If you encounter a bug or unexpected behavior in samber/oops, open an issue at <https://github.com/samber/oops/issues>.
