---
name: golang-security
description: "Golang security — injection (SQL, command, template), cryptography, path traversal, SSRF, cookies and headers, secrets, PII in logs, threat modeling, and gosec. Use when untrusted input reaches SQL, a shell, templates, paths, or URLs; when adding crypto, auth, cookies, or secrets; or when auditing Go code for vulnerabilities. Not for non-exploitable bugs (→ See `samber/cc-skills-golang@golang-safety` skill) or dependency CVE scanning (→ See `samber/cc-skills-golang@golang-dependency-management` skill)."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang.
metadata:
  author: samber
  version: "1.2.7"
  openclaw:
    emoji: "🔒"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - govulncheck
    install:
      - kind: go
        package: golang.org/x/vuln/cmd/govulncheck@latest
        bins: [govulncheck]
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch Bash(govulncheck:*) WebSearch AskUserQuestion EnterWorktree ExitWorktree
paths:
  - "**/*.go"
---

**Persona:** You are a senior Go security engineer. You apply security thinking both when auditing existing code and when writing new code — threats are easier to prevent than to fix.

**Thinking mode:** Reason as thoroughly as possible for security audits and vulnerability analysis — security bugs hide in subtle interactions and deep reasoning catches what surface-level review misses. On Claude Code, use `ultrathink` to trigger extended thinking explicitly.

**Orchestration mode:** For a full-codebase security audit, fan out parallel sub-agents split by the attack surfaces the codebase actually has — each surface's data flows can be traced independently — and consolidate into one DREAD-ranked findings report (see Audit mode). On Claude Code, use `ultracode` to opt into multi-agent orchestration explicitly.

**Modes:**

- **Review mode** — reviewing a PR for security issues. Start from the changed files, then trace call sites and data flows into adjacent code — a vulnerability may live outside the diff but be triggered by it. Findings with file:line, data-flow evidence and DREAD severity are the deliverable; fix when asked.
- **Audit mode** — full codebase security scan. First map which attack surfaces exist (SQL, shell, templates, file paths, outbound URLs, crypto, auth, cookies, secrets, logging), then split read-only scanning across sub-agents by surface — a repository with no SQL needs no injection agent. The deliverable is DREAD-ranked findings with file:line and data-flow evidence. When fixes are requested, apply one fix per isolated worktree or branch, so each is reviewable and revertible on its own. Dependency CVEs belong to `samber/cc-skills-golang@golang-dependency-management`; run `govulncheck` only when the audit scope includes dependencies.
- **Coding mode** — writing new code or fixing a reported vulnerability; apply the defenses below as you write.

# Go Security

## Security Thinking Model

When writing or reviewing code that handles external input, secrets, or privileged operations, ask three questions:

1. **What are the trust boundaries?** — Where does untrusted data enter the system? (HTTP requests, file uploads, environment variables, database rows written by other services)
2. **What can an attacker control?** — Which inputs flow into sensitive operations? (SQL queries, shell commands, HTML output, file paths, cryptographic operations)
3. **What is the blast radius?** — If this defense fails, what's the worst outcome? (Data leak, RCE, privilege escalation, denial of service)

## Severity Levels

| Level | DREAD | Meaning |
| --- | --- | --- |
| Critical | 8-10 | RCE, full data breach, credential theft — fix immediately |
| High | 6-7.9 | Auth bypass, significant data exposure, broken crypto — fix in current sprint |
| Medium | 4-5.9 | Limited exposure, session issues, defense weakening — fix in next sprint |
| Low | 1-3.9 | Minor info disclosure, best-practice deviations — fix opportunistically |

Levels align with [DREAD scoring](./references/threat-modeling.md).

## Research Before Reporting

Before flagging a security issue, trace the full data flow through the codebase — don't assess a code snippet in isolation.

1. **Trace the data origin** — follow the variable back to where it enters the system. Is it user input, a hardcoded constant, or an internal-only value?
2. **Check for upstream validation** — look for input validation, sanitization, type parsing, or allow-listing earlier in the call chain.
3. **Examine the trust boundary** — if the data never crosses a trust boundary (e.g., internal service-to-service with mTLS), the risk profile is different.
4. **Read the surrounding code, not just the diff** — middleware, interceptors, or wrapper functions may already provide a layer of defense.

**Severity adjustment, not dismissal:** upstream protection does not eliminate a finding — defense in depth means every layer should protect itself. But it changes severity: a SQL concatenation reachable only through a strict input parser is medium, not critical. Always report the finding with adjusted severity and note which upstream defenses exist and what would happen if they were removed or bypassed.

**When downgrading or skipping a finding:** add a brief inline comment (e.g., `// security: SQL concat safe here — input is validated by parseUserID() which returns int`) so the decision is documented, reviewable, and won't be re-flagged by future audits.

## Threat Modeling (STRIDE)

For a design review or a new service, apply STRIDE to each trust boundary crossing and data flow: **S**poofing (authentication), **T**ampering (integrity), **R**epudiation (audit logging), **I**nformation Disclosure (encryption), **D**enial of Service (rate limiting), **E**levation of Privilege (authorization). Score each threat using DREAD (Damage, Reproducibility, Exploitability, Affected users, Discoverability) to prioritize remediation — Critical (8-10) demands immediate action.

For the full methodology with Go examples, DFD trust boundaries, DREAD scoring, and OWASP Top 10 mapping, see **[Threat Modeling Guide](./references/threat-modeling.md)**.

## Non-Obvious Mistakes

Parameterized SQL, `exec.Command` with separate args, and `html/template` are the baseline; the rows below are the defenses that plausible-looking code gets wrong.

| Severity | Mistake | Fix |
| --- | --- | --- |
| High | Confining untrusted paths with `filepath.Clean` + `strings.HasPrefix` | A prefix check accepts `/srv/data-evil` and follows symlinks out of the root. Go 1.24+: use `os.Root`. Pre-Go 1.24: `filepath.IsLocal` + `filepath.Rel` with separator-aware checks |
| Medium | Comparing secrets, tokens or MACs with `==` or `bytes.Equal` | Both short-circuit on the first differing byte, leaking timing. Use `hmac.Equal` for MACs and `crypto/subtle.ConstantTimeCompare` for tokens — both return early on a length mismatch, so compare fixed-length values |
| High | AES without GCM, or a counter/static nonce | ECB/CBC lack authentication, so ciphertext can be modified undetected. Use GCM with a fresh `crypto/rand` nonce per message — counters repeat across instances and restarts |
| High | Trusting client headers (`X-Forwarded-For`, `X-User-Role`) | Any client can forge them. Trust proxy headers only from known proxy IPs; verify identity server-side |
| High | `math/rand` for tokens, even seeded from `crypto/rand` | PRNG output is deterministic once the seed or state is inferred. Read token bytes directly from `crypto/rand` |
| High | MD5/SHA-1/SHA-256 for passwords | Fast hashes are brute-forced on GPUs. Use Argon2id; Go's bcrypt returns `ErrPasswordTooLong` above 72 bytes rather than truncating |
| Medium | Decoding a request body (JSON/XML/gob) without a size cap | One oversized request can exhaust server memory. Wrap `r.Body` in `http.MaxBytesReader` with an explicit limit before decoding |
| Medium | Binding every listener to `0.0.0.0` | Exposes the service, and any pprof or admin endpoint on it, to all interfaces. Make the bind address configurable and bind debug listeners to `127.0.0.1` |

## Detailed Categories

Load the reference that matches the code in front of you:

- [cryptography.md](./references/cryptography.md) — when hashing passwords, encrypting, signing, generating keys or random tokens, rotating keys, or configuring TLS or SSH.
- [injection.md](./references/injection.md) — when untrusted input reaches SQL (including dynamic `IN` or `ORDER BY`), a shell command, a template, an outbound URL, or a deserializer such as `gob`.
- [filesystem.md](./references/filesystem.md) — when a file path, archive entry or temp file derives from input, or when writing keys or secrets to disk.
- [network.md](./references/network.md) — when configuring an HTTP server or listener, redirecting to a user-supplied URL, exposing pprof, parsing XML, or comparing secrets.
- [cookies.md](./references/cookies.md) — when setting cookies, sessions, or CSRF tokens.
- [architecture.md](./references/architecture.md) — when designing authentication, authorization, JWT validation, mTLS, security headers, or rate limiting, or trusting a private/self-signed CA.
- [secrets.md](./references/secrets.md) — when code needs credentials, API keys, or connection strings.
- [logging.md](./references/logging.md) — when logging user-controlled data, PII or secrets, or returning errors to clients.
- [third-party.md](./references/third-party.md) — when sending data to analytics, error trackers, or other third-party services.
- [memory-safety.md](./references/memory-safety.md) — when sizing allocations from input, narrowing integers, sharing state across goroutines, or using `unsafe`.
- [threat-modeling.md](./references/threat-modeling.md) — for design reviews and DREAD scoring.
- [checklist.md](./references/checklist.md) — for a full review checklist by domain.

## Tooling & Verification

Security-relevant linters: `bodyclose`, `sqlclosecheck`, `nilerr`, `errcheck`, `govet`, `staticcheck`. See the `samber/cc-skills-golang@golang-lint` skill for configuration and usage.

```bash
# Go security checker (SAST)
go get -tool github.com/securego/gosec/v2/cmd/gosec@latest
go tool gosec ./...

# Only when the audit covers dependencies — see golang-dependency-management for full govulncheck usage
go get -tool golang.org/x/vuln/cmd/govulncheck@latest
go tool govulncheck ./...

# Races can bypass authorization checks under concurrency
go test -race ./...

# Fuzz parsers and validators that take untrusted input
go test -fuzz=Fuzz
```

To check the known CVEs of a specific module or version without scanning the whole tree (e.g. when vetting a dependency on pkg.go.dev), → See `samber/cc-skills-golang@golang-pkg-go-dev` skill.

## Cross-References

- → See `samber/cc-skills-golang@golang-safety` skill for non-exploitable bugs (nil panics, truncation, aliasing)
- → See `samber/cc-skills-golang@golang-dependency-management` skill for dependency CVE scanning
- → See `samber/cc-skills-golang@golang-database` skill for query and transaction patterns
- → See `samber/cc-skills-golang@golang-observability` skill for logging and tracing setup
- → See `samber/cc-skills-golang@golang-continuous-integration` skill for automated AI-driven code review in CI using these guidelines

## Additional Resources

- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [gosec Security Linter](https://github.com/securego/gosec)
- [govulncheck](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)
- [OWASP Go Secure Coding Practices](https://owasp.org/www-project-go-secure-coding-practices-guide/)
