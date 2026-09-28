---
name: golang-observability
description: "Golang production observability — structured logging with slog, Prometheus metrics, OpenTelemetry tracing, continuous profiling, alerting, and Grafana dashboards. Use when instrumenting a Go service, correlating logs with traces, migrating zap/logrus/zerolog to slog, or adding privacy-compliant event tracking. Not for a temporary performance investigation (→ See `samber/cc-skills-golang@golang-benchmark` and `samber/cc-skills-golang@golang-performance` skills)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.4"
  openclaw:
    emoji: "📡"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch WebSearch AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go observability engineer. You treat every unobserved production system as a liability — instrument proactively and correlate signals to diagnose.

**Orchestration mode:** For a codebase-wide observability audit, fan out parallel sub-agents split by signal (→ See Audit mode) — each signal is an independent read-only scan — and consolidate into a per-signal coverage-gap list. On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Coding / instrumentation** (default) — add observability to new or existing code, sequentially. Done when the [Definition of Done](#definition-of-done-for-observability) items that apply to the change hold.
- **Review** — a PR's instrumentation, sequentially: new code exports the expected signals (metrics declared, spans opened and ended, log fields consistent). Deliverable: findings with file:line, ranked by production impact; fixes applied if the user asked for them.
- **Audit** — observability coverage across a codebase. Split by signal, one sub-agent per signal the codebase emits or should emit (up to five): metrics (declarations, PromQL comments, label cardinality), logging (structured output, PII, error logging), tracing (spans on service methods, DB and external calls), profiling (pprof exposure, toggles), RUM (consent, identity key). Deliverable: a per-signal coverage-gap list with file:line; fixes applied if the user asked for them.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-observability` skill takes precedence.

# Go Observability Best Practices

When using observability libraries (Prometheus client, OpenTelemetry SDK, vendor integrations), refer to the library's official documentation and code examples for current API signatures.

## Best Practices Summary

1. **Emit JSON logs in production** with `slog.NewJSONHandler` — log collectors split plain-text multiline records such as stack traces; keep `TextHandler` for local development.
2. **Log with context** — `slog.InfoContext(ctx, ...)` lets the tracing bridge attach trace_id and span_id; the plain variants drop them.
3. **Fan out to several slog handlers with stdlib `slog.NewMultiHandler`** (Go 1.26+), e.g. `slog.New(slog.NewMultiHandler(jsonHandler, auditHandler))` — add a third-party handler-composition library only for routing, failover or pipelines the stdlib can't express.
4. **Prefer Histogram over Summary for latency** — Histograms aggregate across instances and feed `histogram_quantile()` for P50/P90/P99/P99.9, while Summary quantiles are precomputed per instance and can't be combined. Size buckets to the expected range: `prometheus.DefBuckets` start at 5 ms, so sub-millisecond operations all land in the first bucket.
5. **Keep label cardinality bounded** — each unique label combination is a separate time series, so `userID` or raw `r.URL.Path` labels grow Prometheus memory without limit; label by route pattern and put per-user detail in traces.
6. **Write PromQL as comments above each metric declaration** (`// Dashboard: …`, `// Alert: …`) — the queries are reviewed in the same PR as the metric and change in the same commit when labels or buckets do.
7. **Configure the OpenTelemetry TracerProvider at startup on new projects, then span every meaningful operation** — service methods, DB queries, external API calls, message-queue publish/consume, and anything else that takes measurable time or can fail.
8. **Propagate `ctx` everywhere** — it carries trace_id, span_id and deadlines across boundaries, so `db.Query(...)` instead of `db.QueryContext(ctx, ...)`, or `context.Background()` inside a spawned goroutine, silently cuts the trace.
9. **Enable profiling via environment variables** — toggle pprof and continuous profiling without redeploying.
10. **Correlate signals** — inject trace_id into logs and attach exemplars to metrics (→ See [Correlating Signals](#correlating-signals)).
11. **Start dependency alerts from [awesome-prometheus-alerts](https://samber.github.io/awesome-prometheus-alerts/)** — ~500 ready-to-use rules organized by technology (PostgreSQL, Redis, Kafka, Kubernetes…), rather than writing them from scratch.

## The Five Signals

| Signal | Question it answers | Tool | When to use |
| --- | --- | --- | --- |
| **Logs** | What happened? | `log/slog` | Discrete events, errors, audit trails |
| **Metrics** | How much / how fast? | Prometheus client | Aggregated measurements, alerting, SLOs |
| **Traces** | Where did time go? | OpenTelemetry | Request flow across services, latency breakdown |
| **Profiles** | Why is it slow / using memory? | pprof, Pyroscope | CPU hotspots, memory leaks, lock contention |
| **RUM** | How do users experience it? | PostHog, Segment | Product analytics, funnels, session replay |

## References

- [references/logging.md](references/logging.md) — when setting up slog handlers or levels, adding request-scoped attributes, choosing a log sink, or migrating from zap/logrus/zerolog (bridge, replace call sites, remove bridge; on a large codebase, split call-site replacement across parallel sub-agents by package so no two edit the same files).
- [references/metrics.md](references/metrics.md) — when declaring or naming metrics, choosing buckets or labels, or writing PromQL and SLO burn-rate alerts.
- [references/tracing.md](references/tracing.md) — when setting up the TracerProvider, adding spans or `otelhttp`, recording span errors, or tuning sampling cost.
- [references/profiling.md](references/profiling.md) — when enabling pprof in production or setting up continuous profiling with Pyroscope.
- [references/rum.md](references/rum.md) — when tracking product events server-side (PostHog, Segment), or handling consent and data-subject requests under GDPR/CCPA.
- [references/alerting.md](references/alerting.md) — when writing alert rules, choosing severities and `for:` durations, or alerting on Go runtime metrics.
- [references/dashboards.md](references/dashboards.md) — when setting up Grafana dashboards for Go runtime metrics.

## Correlating Signals

A trace_id in log lines lets you jump from a log to the full request trace; an exemplar on a metric links a latency spike to the trace that caused it.

### Logs + Traces: `otelslog` bridge

```go
import "go.opentelemetry.io/contrib/bridges/otelslog"

// Create a logger that automatically injects trace_id and span_id
logger := otelslog.NewHandler("my-service")
slog.SetDefault(slog.New(logger))

// Now every slog call with context includes trace correlation
slog.InfoContext(ctx, "order created", "order_id", orderID)
// Output includes: {"trace_id":"abc123", "span_id":"def456", "msg":"order created", ...}
```

### Metrics + Traces: Exemplars

```go
// When recording a histogram observation, attach the trace_id as an exemplar
// so you can jump from a P99 spike directly to the offending trace
obs := histogram.WithLabelValues("POST", "/orders")
if eo, ok := obs.(prometheus.ExemplarObserver); ok {
    eo.ObserveWithExemplar(duration, prometheus.Labels{"trace_id": traceID})
} else {
    obs.Observe(duration)
}
```

## Definition of Done for Observability

A feature is not production-ready until it is observable. Apply the items that fit the change:

- **New endpoints, jobs or services** — counters for operations and errors, histograms for latencies (every new HTTP endpoint gets latency and error-rate metrics), gauges for saturation, each with its PromQL comment; spans on service methods, DB queries and external calls, with failures recorded via both `span.RecordError(err)` and `span.SetStatus(codes.Error, ...)`; structured `slog` logs using the `*Context` variants, no PII, and each error either logged or returned, not both.
- **Dashboards and alerts** — when the repository holds dashboard or alert-rule definitions, wire the new PromQL into them; otherwise list the dashboards and alerts the owning team should add.
- **RUM** — only for user-facing business events: track them server-side, keyed by `user_id` (never email, which is mutable PII), after checking consent.

## Cross-References

- → See `samber/cc-skills-golang@golang-error-handling` skill for the single handling rule (log or return, never both).
- → See `samber/cc-skills-golang@golang-troubleshooting` skill for using observability signals to diagnose production issues.
- → See `samber/cc-skills-golang@golang-security` skill for protecting pprof endpoints and avoiding PII in logs.
- → See `samber/cc-skills-golang@golang-context` skill for propagating trace context across service boundaries.
- → See `samber/cc-skills@promql-cli` skill for querying and exploring PromQL expressions against Prometheus from the CLI.
