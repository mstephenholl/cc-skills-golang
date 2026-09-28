---
name: golang-grpc
description: "gRPC services in Golang — proto package layout, server and client setup, interceptors, status codes, deadlines, TLS/mTLS, streaming, and bufconn tests. Apply when implementing, reviewing, or debugging Go gRPC servers or clients, or when the codebase imports `google.golang.org/grpc`."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.4"
  openclaw:
    emoji: "🌐"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - protoc
    install:
      - kind: brew
        formula: protobuf
        bins: [protoc]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(protoc:*) AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go distributed systems engineer. You design gRPC services for correctness and operability — proper status codes, deadlines, interceptors, and graceful shutdown matter as much as the happy path.

**Modes:**

- **Build** — implementing a server or client. Done when the proto compiles, health checks and `GracefulStop` are wired, and a bufconn test covering the error codes passes.
- **Review** — findings against Common Mistakes, ranked, with file:line. If the user asked for fixes, apply them and re-run the bufconn tests.

**Dependencies:**

- protoc: `brew install protobuf`
- protoc-gen-go: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
- protoc-gen-go-grpc: `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`

# Go gRPC Best Practices

Treat gRPC as a transport layer — handlers translate protobuf messages into domain calls and domain errors into status codes, nothing more.

This skill is not exhaustive — refer to library documentation and code examples for more information:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

## Protos

Give every RPC its own `Request`/`Response` wrapper messages, even when one field would do — a bare `string` or `google.protobuf.Empty` can never gain a field without breaking clients. Read [protoc-reference.md](references/protoc-reference.md) when writing `.proto` files or wiring code generation — domain/version layout, `buf.gen.yaml`, `go_package`, and embedding `Unimplemented<Service>Server` rather than `Unsafe<Service>Server`.

## Server

```go
srv := grpc.NewServer(grpc.ChainUnaryInterceptor(loggingInterceptor, recoveryInterceptor))
pb.RegisterUserServiceServer(srv, svc)
healthpb.RegisterHealthServer(srv, health.NewServer())
go srv.Serve(lis)

// On shutdown signal: drain in-flight RPCs, but never hang
stopped := make(chan struct{})
go func() { srv.GracefulStop(); close(stopped) }()
select {
case <-stopped:
case <-time.After(15 * time.Second):
    srv.Stop()
}
```

`GracefulStop` waits for every in-flight RPC, and a long-lived stream never finishes on its own — without the `Stop()` fallback, one open stream blocks shutdown until the orchestrator kills the process.

## Client

Behind a Kubernetes headless service, use the `dns:///` scheme with `round_robin` — the default `pick_first` pins every RPC to one replica. Retries and a default timeout come from the service config:

```go
conn, err := grpc.NewClient("dns:///user-service:50051",
    grpc.WithTransportCredentials(creds),
    grpc.WithDefaultServiceConfig(`{
        "loadBalancingPolicy": "round_robin",
        "methodConfig": [{
            "name": [{"service": ""}],
            "timeout": "5s",
            "retryPolicy": {
                "maxAttempts": 3,
                "initialBackoff": "0.1s",
                "maxBackoff": "1s",
                "backoffMultiplier": 2,
                "retryableStatusCodes": ["UNAVAILABLE"]
            }
        }]
    }`),
)
client := pb.NewUserServiceClient(conn)
```

Pass auth tokens and trace IDs with `metadata.NewOutgoingContext`. A client `keepalive.ClientParameters.Time` shorter than the server's `keepalive.EnforcementPolicy.MinTime` (default 5 minutes) gets the connection closed with `too_many_pings`, so lower both together.

## Errors

A plain Go `error` returned from a handler reaches the client as `codes.Unknown`, which tells it nothing about whether to retry. Pick the code by what the caller should do next:

- `InvalidArgument` — the request is malformed; retrying the same request can never succeed.
- `NotFound`, `AlreadyExists`, `PermissionDenied`, `Unauthenticated` — the entity or caller is the problem.
- `FailedPrecondition` — the system must change state first (insufficient stock, non-empty directory); a blind retry fails again.
- `Unavailable` — transient; the only code a retry policy should list.
- `Internal` — a bug. Log the cause server-side and send a generic message, since status messages reach callers verbatim.

```go
if errors.Is(err, ErrNotFound) {
    return nil, status.Errorf(codes.NotFound, "user %q not found", req.UserId)
}
slog.ErrorContext(ctx, "get user", "user_id", req.UserId, "err", err) // details stay server-side
return nil, status.Errorf(codes.Internal, "failed to load user %q", req.UserId)
```

Attach field-level validation errors with `errdetails.BadRequest` via `status.WithDetails`.

## Streaming, testing, security

- Prefer server streaming over one large response — a single message is buffered whole on both sides and hits the 4 MB default receive limit, and raising `MaxRecvMsgSize` only moves the ceiling.
- Test through `bufconn`, which exercises serialization, interceptors and metadata in memory, and assert status codes on every error path. Read [testing.md](references/testing.md) when writing tests — bufconn setup, table-driven code checks, streaming, metadata and deadline tests.
- Enable TLS in production — credentials travel in metadata. Use mTLS or a service mesh for service-to-service auth, and `credentials.PerRPCCredentials` plus an auth interceptor for user tokens.

## Performance

| Setting | Purpose | Typical Value |
| --- | --- | --- |
| `keepalive.ServerParameters.Time` | Ping interval for idle connections | 30s |
| `keepalive.ServerParameters.Timeout` | Ping ack timeout | 10s |
| `grpc.MaxRecvMsgSize` | Override 4 MB default for large payloads | 16 MB |
| Connection pooling | Multiple conns for high-load streaming | 4 connections |

Most services do not need connection pooling — profile before adding complexity.

## Common Mistakes

| Mistake | Fix |
| --- | --- |
| Returning raw `error` | Becomes `codes.Unknown` — client can't decide whether to retry. Use `status.Errorf` with a specific code |
| No deadline on client calls | Slow upstream hangs indefinitely. Always `context.WithTimeout` |
| New connection per request | Wastes TCP/TLS handshakes. Create once, reuse — HTTP/2 multiplexes RPCs |
| Reflection enabled in production | Lets attackers enumerate every method. Enable only in dev/staging |
| `codes.Internal` for all errors | Wrong codes break client retry logic. `Unavailable` triggers retry; `InvalidArgument` does not |
| Bare types as RPC arguments | Can't add fields to `string`. Wrapper messages allow backwards-compatible evolution |
| Missing health check service | Kubernetes can't determine readiness, kills pods during deployments |
| Ignoring context cancellation | Long operations continue after caller gave up. Check `ctx.Err()` |

## Cross-References

- → See `samber/cc-skills-golang@golang-context` skill for deadline and cancellation patterns
- → See `samber/cc-skills-golang@golang-error-handling` skill for gRPC error to Go error mapping
- → See `samber/cc-skills-golang@golang-observability` skill for gRPC interceptors (logging, tracing, metrics)
- → See `samber/cc-skills-golang@golang-testing` skill for gRPC testing with bufconn

If you encounter a bug or unexpected behavior in grpc-go, open an issue at <https://github.com/grpc/grpc-go/issues>.
