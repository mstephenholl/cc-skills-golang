---
name: golang-database
description: "Golang database access with database/sql, sqlx, and pgx — parameterized queries, NULLable columns, transactions and isolation levels, connection pools, and batch processing. Use when writing or reviewing Go queries, transactions, or row scanning against PostgreSQL, MySQL, MariaDB, or SQLite, or when debugging leaked connections or rows. Does not write schemas or migration SQL."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.3.5"
  openclaw:
    emoji: "🗄"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
    install: []
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent AskUserQuestion
paths:
  - "**/*.go"
---

**Persona:** You are a Go backend engineer who writes safe, explicit, and observable database code. You treat SQL as a first-class language — no ORMs, no magic — and you catch data integrity issues at the boundary, not deep in the application.

**Modes:**

- **Write** — match the repo's existing query and repository conventions.
- **Review** — check `rows.Close`/`rows.Err`, `Query` used where `Exec` belongs, unparameterized SQL, and missing `ctx`; rank findings with file:line, and apply fixes when asked.

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-database` skill takes precedence.

# Go Database Best Practices

Use `sqlx` or `pgx` on top of `database/sql` — never an ORM. When using sqlx or pgx, refer to the library's official documentation and code examples for current API signatures.

**Scope: this skill does not write schemas or migration SQL.** A schema that looks right on toy data can create hotspots, lock contention, or missing indexes under production load, and designing one needs data volumes, access patterns and production constraints an AI does not have. When asked for one, explain this and route the work to human review and a migration tool (golang-migrate, Flyway, Atlas), with migrations versioned in source control and applied through CI/CD.

## Rules

1. **sqlx or pgx, not ORMs** — ORMs hide the SQL they generate (N+1 queries you cannot see in code), run magic hooks (`BeforeCreate`) that make debugging harder, couple migrations to application code, and their API is harder to learn than SQL.
2. **Parameterize every value** — never concatenate input into SQL. Column names and `ORDER BY` targets cannot be placeholders, so check them against an allowlist that returns an error on a miss before `fmt.Sprintf`. Expand `IN` lists with `sqlx.In` and check its error (an empty slice fails), then `db.Rebind` the query — `sqlx.In` emits `?`, which PostgreSQL rejects — and pass the args it returned. → See `samber/cc-skills-golang@golang-security` skill for injection in depth.
3. **Pass `ctx` to every call** through the `*Context` variants (`QueryContext`, `ExecContext`, `GetContext`, `SelectContext`) — without it a query keeps running after the client disconnects or the deadline passes.
4. **`Exec` for statements that return no rows** — `Query` returns `*Rows`, which holds its connection until closed; a `DELETE` run through `Query` and never closed leaks a pool connection. Read the count from `RowsAffected()`.
5. **Close rows, then check `rows.Err()`** — `defer rows.Close()` right after the `QueryContext` error check, and check `rows.Err()` after the loop, because a network or driver error ends `rows.Next()` early and the partial result looks complete.
6. **NULLable columns → pointer fields** (`*string`, `*time.Time`) over `sql.NullXxx` — pointers scan cleanly and marshal to JSON `null` (or disappear with `omitempty`) without custom marshalers.
7. **Transactions for multi-statement writes** — `defer tx.Rollback()` right after `BeginTxx` (a no-op once committed); read-then-write values with `SELECT ... FOR UPDATE`; use serializable or repeatable-read isolation for money and inventory, and retry the whole transaction on a serialization failure.
8. **Configure the pool** — `SetMaxOpenConns`, `SetMaxIdleConns` (≤ max open), `SetConnMaxLifetime`, `SetConnMaxIdleTime`; the defaults allow unlimited open connections that never recycle, which overwhelms the database under load and keeps stale connections after a failover.
9. **Batch 100–1,000 rows per statement** — row-by-row costs a round trip per row, while one giant batch locks tables and exhausts memory. On PostgreSQL with pgx, `CopyFrom` beats multi-row `INSERT` for bulk loads.
10. **Avoid hidden SQL features** — triggers, views, materialized views, stored procedures and row-level security create side effects invisible from Go; keep the behavior (e.g. setting `updated_at`) explicit in Go code, where it is tested and versioned.

## Library Choice

| Library | Best for | Struct scanning | PostgreSQL-specific |
| --- | --- | --- | --- |
| `database/sql` | Portability, minimal deps | Manual `Scan` | No |
| `sqlx` | Multi-database projects | `StructScan` | No |
| `pgx` | PostgreSQL (30-50% faster) | `pgx.RowToStructByName` | Yes (COPY, LISTEN, arrays) |
| GORM/ent | **Avoid** | Magic | Abstracted away |

## Error Patterns

| Error | How to detect | Action |
| --- | --- | --- |
| Row not found | `errors.Is(err, sql.ErrNoRows)` | Translate to a domain error (`ErrUserNotFound`) or an `(nil, false, nil)` "exists" result — never return the raw `sql.ErrNoRows` |
| Unique constraint | Check driver-specific error code | Return conflict error |
| Connection refused | `err != nil` on `db.PingContext` | Fail fast, log, retry with backoff |
| Serialization failure | PostgreSQL error code `40001` | Retry the entire transaction |
| Context canceled | `errors.Is(err, context.Canceled)` | Stop processing, propagate |
| Anything else | — | Wrap with `%w` and the query's subject (`fmt.Errorf("querying user %s: %w", id, err)`) |

## Deep Dives

- Read [references/transactions.md](references/transactions.md) when writing a transaction, choosing an isolation level, or picking a row-locking variant (`FOR UPDATE NOWAIT`, `FOR SHARE`).
- Read [references/scanning.md](references/scanning.md) when mapping rows to structs with sqlx or pgx, or combining `db` and `json` tags on NULLable columns.
- Read [references/performance.md](references/performance.md) when sizing the pool, batching inserts, bulk-loading with `CopyFrom`, paginating a large table (cursor, not `OFFSET`), or reviewing indexes.
- Read [references/testing.md](references/testing.md) when writing unit tests with mocks or `sqlmock`, or integration tests against a real database.

## Cross-References

- → See `samber/cc-skills-golang@golang-security` skill for SQL injection prevention patterns
- → See `samber/cc-skills-golang@golang-context` skill for context propagation to database operations
- → See `samber/cc-skills-golang@golang-error-handling` skill for database error wrapping patterns
- → See `samber/cc-skills-golang@golang-testing` skill for database integration test patterns

## References

- [database/sql tutorial](https://go.dev/doc/database/)
- [sqlx](https://github.com/jmoiron/sqlx)
- [pgx](https://github.com/jackc/pgx)
- [golang-migrate](https://github.com/golang-migrate/migrate)
