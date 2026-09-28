---
name: golang-graphql
description: "GraphQL APIs in Golang with gqlgen or graphql-go — schema design, resolvers, DataLoaders against N+1, subscriptions, and query complexity limits. Apply when building or reviewing a Go GraphQL server, or when the codebase imports `github.com/99designs/gqlgen` or `github.com/graph-gophers/graphql-go`."
user-invocable: false
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "0.2.5"
  openclaw:
    emoji: "🔮"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
    skill-library-version: "0.17.89"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(curl:*) Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go GraphQL engineer. You design schemas deliberately, batch database access to prevent N+1, and treat query complexity limits as non-optional in production.

**Modes:**

- **Build** — new schemas, resolvers, or server setup. Match the existing resolver layout and `gqlgen.yml`; done when the schema regenerates (gqlgen) or parses (graph-gophers) cleanly and every child-list resolver loads through a per-request DataLoader.
- **Review** — prioritize N+1 resolvers, global DataLoaders, missing complexity caps and introspection enabled in production; deliver ranked findings with file:line. If the user asked for fixes, apply them and re-verify.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-graphql` skill takes precedence.

# Go GraphQL Best Practices

Both major libraries are schema-first: write SDL (`.graphql` files), bind Go resolvers. Choose based on project size and team preferences.

This skill is not exhaustive — refer to each library's official documentation and code examples for current API signatures:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

## Library Choice

| Library | Approach | Type safety | Build step | Best for |
| --- | --- | --- | --- | --- |
| `github.com/99designs/gqlgen` | Codegen | Compile-time | `go generate` | Large schemas, federation, strict types |
| `github.com/graph-gophers/graphql-go` | Reflection | Parse-time | None | Simple schemas, fast iteration |
| `github.com/graphql-go/graphql` | Code-first | Runtime | None | **Avoid** — verbose, no SDL |

Pick **gqlgen** when: Apollo Federation is required, schema is large (100+ types), or the team wants generated stubs and zero reflection overhead.

Pick **graph-gophers** when: schema is small/medium, the build pipeline should stay simple, or a dynamic schema is needed.

For deep-dive on each library, see [gqlgen reference](./references/gqlgen.md) and [graphql-go reference](./references/graphql-go.md).

## Schema Design

```graphql
# ✓ Good — explicit nullability; ID scalar for opaque identifiers
type User {
  id: ID!
  email: String! # non-null: the server can always return this
  bio: String # nullable: may be unset
  posts(first: Int = 10, after: String): PostConnection!
}

# ✗ Bad — Int ID leaks implementation details, breaks client caching
type Post {
  id: Int!
}
```

**Nullability rule:** mark a field `!` only when the server can _always_ return a value. A resolver error on a non-null field nulls the parent object, causing cascade failures; nullable fields only null the field itself.

**Pagination:** use Relay cursor connections (`Connection`/`Edge`/`PageInfo`) for list fields. Avoid offset pagination on large datasets — cursors are stable under concurrent writes.

**Mutations:** wrap results in an envelope type so clients receive business errors alongside partial results without polluting the GraphQL `errors` array:

```graphql
type CreateUserPayload {
  user: User
  errors: [UserError!]!
}
```

## Resolver Patterns

Keep resolvers thin — translate GraphQL input into a service call and the result back into GraphQL types, with per-type resolver structs (`userResolver`, `postResolver`) rather than one monolithic resolver; SQL inside a resolver bypasses both the service layer and the DataLoaders below.

## N+1 Prevention (DataLoaders)

Each `User.posts` resolver fires a SQL query per user without batching — O(n) DB calls for n users. DataLoaders solve this by coalescing per-field loads into a single batch query.

**Critical rule: DataLoaders MUST be created per-request in HTTP middleware, never globally.** A global DataLoader caches across requests — stale data, potential cross-user data leakage.

```go
// ✓ Good — per-request DataLoader in middleware
func DataLoaderMiddleware(db *sql.DB, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        loaders := &Loaders{
            PostsByUserID: newPostsByUserIDLoader(r.Context(), db),
        }
        ctx := context.WithValue(r.Context(), loadersKey, loaders)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// ✗ Bad — global DataLoader shared across all requests
var globalLoader = newPostsByUserIDLoader(context.Background(), db)
```

In gqlgen, mark batched fields with `resolver: true` in `gqlgen.yml` to force a dedicated resolver method. See [gqlgen reference](./references/gqlgen.md) for full DataLoader wiring.

## Authentication and Authorization

Authenticate in HTTP middleware that stores the caller's identity in `context.Context`, then authorize per field — `@hasRole`-style schema directives in gqlgen keep the policy in the schema instead of scattered across resolvers, and graph-gophers resolvers check the context identity explicitly. Authenticate subscriptions in the WebSocket `InitFunc` — browsers can't set an `Authorization` header on the upgrade request, so the token arrives in the `connection_init` payload ([gqlgen reference](./references/gqlgen.md)).

## Error Handling

Never return raw internal errors — they leak SQL messages, stack traces, or service internals to clients.

```go
// gqlgen — custom ErrorPresenter strips internal details
srv.SetErrorPresenter(func(ctx context.Context, err error) *gqlerror.Error {
    var gqlErr *gqlerror.Error
    if errors.As(err, &gqlErr) {
        return gqlErr // already formatted
    }
    // log internal err here
    return gqlerror.Errorf("internal error") // safe client message
})

// Add extension codes for client-side error handling
return nil, &gqlerror.Error{
    Message: "user not found",
    Extensions: map[string]any{"code": "NOT_FOUND"},
}
```

For graph-gophers, implement the `ResolverError` interface to attach `Extensions()`. See [graphql-go reference](./references/graphql-go.md).

Use `graphql.AddError(ctx, err)` in gqlgen for non-fatal field errors where the resolver can still return partial data.

For error wrapping patterns, see the `samber/cc-skills-golang@golang-error-handling` skill.

## Subscriptions

Subscriptions use long-lived WebSocket connections. The critical discipline: **always respect context cancellation** — a leaked goroutine per disconnected client exhausts resources silently.

```go
// ✓ Good — closes channel when client disconnects
func (r *subscriptionResolver) MessageAdded(ctx context.Context, room string) (<-chan *model.Message, error) {
    ch := make(chan *model.Message, 1)
    sub := r.pubsub.Subscribe(room) // subscribe once before the goroutine
    go func() {
        defer close(ch)                       // always close; signals iteration to stop
        defer r.pubsub.Unsubscribe(room, sub) // release the broker side too, or it keeps publishing into sub
        for {
            select {
            case <-ctx.Done():
                return // client disconnected
            case msg := <-sub:
                select {
                case ch <- msg:
                case <-ctx.Done():
                    return
                }
            }
        }
    }()
    return ch, nil
}

// ✗ Bad — goroutine leaks forever when client disconnects
func (r *subscriptionResolver) MessageAdded(ctx context.Context, room string) (<-chan *model.Message, error) {
    ch := make(chan *model.Message, 1)
    go func() {
        for msg := range r.pubsub.Subscribe(room) {
            ch <- msg // blocks forever after client gone
        }
    }()
    return ch, nil
}
```

## Performance and Safety

Production GraphQL servers require explicit limits. Without them, a single deeply nested query exhausts CPU and memory.

```go
// gqlgen — handler.New, not the deprecated NewDefaultServer: the latter
// registers Introspection unconditionally, so the gate below would be a no-op
srv := handler.New(es)
srv.AddTransport(transport.POST{})
srv.Use(extension.FixedComplexityLimit(200)) // max cost per query

// Gate introspection — only in non-production environments
if os.Getenv("ENV") != "production" {
    srv.Use(extension.Introspection{})
}
```

Full transport, query-cache and APQ setup: [Production Handler Setup](references/gqlgen.md#production-handler-setup).

For graph-gophers: `graphql.MaxDepth(10)` and `graphql.MaxParallelism(10)` options at `ParseSchema` time.

**Query allow-listing:** in production, consider persisted queries (gqlgen APQ extension) to reject arbitrary query strings.

## Common Mistakes

| Mistake | Why it matters | Fix |
| --- | --- | --- |
| N+1 queries in child resolvers | One SQL per parent row → O(n) DB calls | Use per-request DataLoader |
| Global DataLoader | Cross-request cache — stale data, data leaks | Create DataLoader in request middleware |
| Editing `models_gen.go` directly | Next `go generate` wipes hand edits | Use `autobind` or `models.<T>.model` in `gqlgen.yml` |
| Forgetting `go generate` after schema change | Resolver interface mismatch at compile time | Re-run `go tool gqlgen generate` |
| `int` field in graph-gophers resolver | Library requires `int32` for `Int` scalar | Use `int32` (or `float64` for `Float`) |
| Introspection enabled in production | Exposes full schema to attackers | Gate with `ENV` check |
| No complexity cap | Deeply nested query → CPU/memory DoS | `extension.FixedComplexityLimit(N)` |
| Leaking DB errors from resolvers | Exposes SQL internals to clients | Wrap in `ErrorPresenter` / `ResolverError` |
| Subscription goroutine leak | Client disconnect → goroutine runs forever | `defer close(ch)` + `select ctx.Done()` |
| Nullable field for always-required data | Clients must null-check everywhere | Mark `!` in schema; return error from resolver |

## Deep Dives

- **[gqlgen reference](./references/gqlgen.md)** — read when working in a gqlgen project: `gqlgen.yml` model binding, DataLoader wiring and batch-function shape, auth directives, WebSocket transport, file uploads, Federation v2, production handler setup
- **[graphql-go reference](./references/graphql-go.md)** — read when working with graph-gophers: Go type mapping (`int32`, pointers for nullable), resolver and args structs, custom scalars, OpenTelemetry tracing
- **[Testing](./references/testing.md)** — read when writing GraphQL tests: gqlgen client harness, gqltesting, subscriptions, auth directives

## Cross-References

- → See `samber/cc-skills-golang@golang-context` skill for context propagation in resolvers and subscriptions
- → See `samber/cc-skills-golang@golang-error-handling` skill for error wrapping and sentinel patterns
- → See `samber/cc-skills-golang@golang-testing` skill for table-driven and integration test patterns
- → See `samber/cc-skills-golang@golang-observability` skill for tracing and metrics in resolvers
- → See `samber/cc-skills-golang@golang-security` skill for input validation and injection prevention
- → See `samber/cc-skills-golang@golang-database` skill for N+1 query patterns and DataLoader database batching

## References

- [gqlgen](https://github.com/99designs/gqlgen)
- [graph-gophers/graphql-go](https://github.com/graph-gophers/graphql-go)
- [Relay cursor connections spec](https://relay.dev/graphql/connections.htm)

If you encounter a bug or unexpected behavior in gqlgen, open an issue at <https://github.com/99designs/gqlgen/issues>.

If you encounter a bug or unexpected behavior in graph-gophers/graphql-go, open an issue at <https://github.com/graph-gophers/graphql-go/issues>.
