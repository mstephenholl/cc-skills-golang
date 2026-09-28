# Configure mode — conditional Go skill directives for a project

This workflow writes two independent things to the project's agent-config file(s) (CLAUDE.md, AGENTS.md, or equivalent):

1. A **routing directive** for `golang-how-to` itself — one sentence saying when to consult the router.
2. An optional `## Go skills` block of **conditional** directives — one line per skill, each naming the situation in which that skill applies.

Both are conditional by design. An unconditional "load X before every task" line pulls the skill's whole SKILL.md body into context on every Go task — typically 2,000–4,000 tokens per skill (see the README's SKILL.md (tok) column), not the ~100-token description — and that body stays in context for every later turn, diluting the guidance the task actually needs.

## Table of Contents

- [When to use](#when-to-use)
- [Routing directive](#routing-directive)
- [Step 1 — Detect the project config file(s)](#step-1--detect-the-project-config-files)
- [Step 2 — Idempotency check](#step-2--idempotency-check)
- [Step 3 — Confirm the skill set with the user](#step-3--confirm-the-skill-set-with-the-user)
- [Step 4 — Write the block](#step-4--write-the-block)
  - [Markdown targets (CLAUDE.md, AGENTS.md, GEMINI.md, copilot-instructions.md)](#markdown-targets-claudemd-agentsmd-geminimd-copilot-instructionsmd)
  - [Cursor target (`.cursor/rules/*.mdc`)](#cursor-target-cursorrulesmdc)
- [Step 5 — Confirm to the user](#step-5--confirm-to-the-user)
- [Notes on company overrides (⚙️ skills)](#notes-on-company-overrides-%EF%B8%8F-skills)

## When to use

- The user runs `/golang-how-to configure`, or accepts an offer to configure (e.g. from `samber/cc-skills-golang@golang-project-layout` at project creation). Writing to someone's agent-config file is a visible change to their project, so it happens only on request or with consent.
- The project has hard requirements on specific skills (e.g., `golang-security` must apply whenever untrusted input is handled, not only when the user says "security").
- A company skill overrides a community default (⚙️ skills) and must always win.

## Routing directive

### Template

```markdown
For Go tasks that span several concerns (a new service, a refactor, an audit), or when two Go skills seem to overlap, consult the `samber/cc-skills-golang@golang-how-to` skill to choose which Go skills to load.
```

### Insertion point

- If a `## Go skills` block already exists or is being created in the same pass, insert the directive as its own line directly above that heading, separated by a blank line.
- Otherwise, append it under a `## Go development` heading (create the heading if the file has no such section).

### Idempotency

Grep for the exact phrase before writing, and for the older unconditional wording this workflow used to write:

```bash
grep -n 'consult the `samber/cc-skills-golang@golang-how-to` skill' CLAUDE.md
grep -n 'load the `samber/cc-skills-golang@golang-how-to` skill first' CLAUDE.md
```

Skip writing if the current directive is present. If only the older "load … first" line is present, offer to replace it with the conditional one.

## Step 1 — Detect the project config file(s)

Every harness reads its own agent-config file or directory. None is more "primary" than another — detect and write to whichever exist, and write to all of them if more than one does:

| File / directory | Harness(es) | Format |
| --- | --- | --- |
| `CLAUDE.md` | Claude Code | Markdown, single file, appended to |
| `AGENTS.md` | Codex, OpenCode, and other multi-agent harnesses | Markdown, single file, appended to |
| `GEMINI.md` | Gemini CLI, Antigravity | Markdown, single file, appended to |
| `.cursor/rules/*.mdc` | Cursor | **Directory** of `.mdc` files, each with its own YAML frontmatter — not a single markdown file to append to |
| `.github/copilot-instructions.md` | GitHub Copilot | Markdown, single file, appended to |

Check which of these exist at the project root. If multiple exist, write to all of them — different harnesses read different files, and a project may support several. If none exist, ask the user which one(s) to create.

## Step 2 — Idempotency check

For the markdown files (`CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`), grep each one for the routing directive and an existing skills block before writing:

```bash
grep -n 'consult the `samber/cc-skills-golang@golang-how-to` skill' CLAUDE.md
grep -n -E "^## (Go skills|Required Go skills)" CLAUDE.md
```

For Cursor, check whether `.cursor/rules/golang-skills.mdc` already exists instead — its presence itself is the idempotency signal, since it's a dedicated file rather than a shared section inside a larger document.

If a skills block (or, for Cursor, the rule file) already exists, read it and confirm with the user whether to update it in place or skip. A legacy `## Required Go skills` block with "MUST always be applied… at the start of every Go-related task" wording should be offered for conversion to the conditional form below.

## Step 3 — Confirm the skill set with the user

Confirm which skills to add — one question, one round of confirmation, not a running back-and-forth. Present the ⭐️ recommended skills as the default selection, each with its condition from the table, and note the cost: a skill whose condition matches loads its full SKILL.md body for the rest of that task.

Recommended ⭐️ set for most projects, with the condition written for each:

| Skill | Condition line |
| --- | --- |
| `golang-code-style` | when reviewing Go code for readability or settling a style question |
| `golang-data-structures` | when choosing a collection type or sizing slices and maps on a hot path |
| `golang-design-patterns` | when designing a new API, constructor, package, or service lifecycle |
| `golang-documentation` | when writing doc comments, a README, or a CHANGELOG |
| `golang-error-handling` | when defining error types, or deciding whether to wrap, return, or log an error |
| `golang-modernize` | when editing code written for an older Go version than the module's `go` directive |
| `golang-naming` | when introducing or renaming exported identifiers, packages, or errors |
| `golang-safety` | when code dereferences pointers from callers, narrows numeric types, or shares slices and maps across boundaries |
| `golang-security` | when untrusted input reaches SQL, shell, templates, paths, or URLs, or when touching crypto, auth, cookies, or secrets |
| `golang-testing` | when writing or changing tests |
| `golang-troubleshooting` | when a test fails, a program panics, or behavior is unexpected |

Additional skills to suggest based on codebase context, with the condition `when working on code that imports <path>`:

- Database layer detected (`database/sql`, `sqlx`, `pgx`) → suggest `golang-database`
- CI config detected (`.github/workflows/`) → suggest `golang-continuous-integration` (condition: `when editing .github/workflows/`)
- Cobra imports detected → suggest `golang-spf13-cobra`
- Viper imports detected → suggest `golang-spf13-viper`
- samber/lo imports detected → suggest `golang-samber-lo`
- Any other library-specific import → suggest the matching library skill

If the user insists a skill must apply to every Go task regardless of context, honor it — write that one line without a condition — but state the per-task cost once.

## Step 4 — Write the block

### Markdown targets (CLAUDE.md, AGENTS.md, GEMINI.md, copilot-instructions.md)

Template:

```markdown
For Go tasks that span several concerns (a new service, a refactor, an audit), or when two Go skills seem to overlap, consult the `samber/cc-skills-golang@golang-how-to` skill to choose which Go skills to load.

## Go skills

Apply these `samber/cc-skills-golang` skills when their condition matches the task:

- `samber/cc-skills-golang@golang-error-handling` — when defining error types, or deciding whether to wrap, return, or log an error
- `samber/cc-skills-golang@golang-security` — when untrusted input reaches SQL, shell, templates, paths, or URLs, or when touching crypto, auth, cookies, or secrets
- `samber/cc-skills-golang@golang-testing` — when writing or changing tests
```

Replace the skill list with the confirmed set from Step 3. Use the fully-qualified `samber/cc-skills-golang@<name>` identifier for each skill. If Step 2 found the routing directive already present elsewhere in the file, don't duplicate it — write only the `## Go skills` block.

Insertion point:

- If the file is empty: write the block at the top.
- If the file has existing content: append after the last section, separated by a blank line.
- If a `## Go skills` (or legacy `## Required Go skills`) block already exists: replace only the bullet list inside it, preserving surrounding content.

Edit the file directly, rather than shelling out to a script — that keeps the change reviewable as a normal diff. Perform an idempotency check after writing: re-read the file and verify the block appears exactly once.

### Cursor target (`.cursor/rules/*.mdc`)

`.cursor/rules` is a directory, not a file — each rule lives in its own `.mdc` file with YAML frontmatter (`description`, `globs`, `alwaysApply`). Do not try to append to it as if it were a single markdown document; the append-and-replace logic above does not apply here.

Create `.cursor/rules/golang-skills.mdc` (create the `.cursor/rules/` directory first if it doesn't exist) using [cursor-go-skills.mdc](../assets/cursor-go-skills.mdc) as the starting template. Keep its `alwaysApply: false` plus Go `globs`, so Cursor attaches the rule only when Go files are in context rather than in every conversation. Replace the placeholder `## Go skills` list with the confirmed set from Step 3, same as the markdown targets. If the file already exists, replace only the bullet list, preserving its frontmatter and surrounding content.

## Step 5 — Confirm to the user

After writing, summarize:

- Which file(s) or rule(s) were updated
- Whether the routing directive for `golang-how-to` was added or was already present
- Which skills were added, with their conditions
- Cost: each skill loads its full SKILL.md body (README's SKILL.md (tok) column) only on tasks where its condition matches
- Note: skills marked ⚙️ (overridable) will be superseded if a company skill explicitly declares the override in its body

## Notes on company overrides (⚙️ skills)

Skills marked ⚙️ in the README support company overrides. If the project has a company skill that supersedes a community default (e.g., `acme/cc-skills@golang-error-handling-acme` supersedes `samber/cc-skills-golang@golang-error-handling`), use the company skill FQN in the block instead — do NOT list both.

To declare an override in a company skill body, add near the top:

```
> This skill supersedes `samber/cc-skills-golang@golang-error-handling` for [Company] projects.
```

Overridable skills: `golang-code-style`, `golang-concurrency`, `golang-context`, `golang-database`, `golang-dependency-injection`, `golang-design-patterns`, `golang-documentation`, `golang-error-handling`, `golang-naming`, `golang-observability`, `golang-structs-interfaces`, `golang-testing`.
