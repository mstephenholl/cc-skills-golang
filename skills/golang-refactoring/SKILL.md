---
name: golang-refactoring
description: "Golang refactoring at scale — a coverage-adaptive safety net, behavior-preserving transforms (gopls rename/extract, gofmt -r, gopatch), breaking import cycles, and staged PRs. Use when a Go function, type, or package has outgrown its shape, when planning a multi-step refactor, or when renaming or moving code across packages. For the target shape → See `samber/cc-skills-golang@golang-naming`, `samber/cc-skills-golang@golang-project-layout`, or `samber/cc-skills-golang@golang-design-patterns` skill."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness, and for projects using Golang. Requires gopls and git.
metadata:
  author: samber
  version: "1.1.3"
  openclaw:
    emoji: "♻️"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - gopls
    install:
      - kind: go
        package: golang.org/x/tools/gopls@latest
        bins: [gopls]
      - kind: go
        package: golang.org/x/perf/cmd/benchstat@latest
        bins: [benchstat]
    skill-library-version: "0.20.0"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Bash(gh:*) Bash(gopls:*) Bash(benchstat:*) LSP mcp__gopls__* Agent AskUserQuestion EnterWorktree ExitWorktree WebFetch WebSearch
paths:
  - "**/*.go"
---

> **Community default.** A company skill that explicitly supersedes `samber/cc-skills-golang@golang-refactoring` skill takes precedence.

**Persona:** You are a Go refactoring engineer. You never change structure and behavior in the same step — you keep a green test net, prefer behavior-preserving tools over hand-edits, and land changes as small, reviewable PRs.

**Thinking mode:** Reason as thoroughly as possible when planning a multi-step refactor — mapping blast radius, sequencing PRs to avoid merge conflicts, and deciding where a refactor can safely go parallel all punish shallow reasoning, since a wrong ordering call surfaces as a broken build or a conflict-riddled merge, not as an obviously wrong plan. On Claude Code, use `ultrathink` to trigger extended thinking explicitly.

**Modes:**

- **Plan mode** — applies to multi-step or High-risk refactors; Low/Medium single steps proceed with the verify loop below. Map structure and blast radius with gopls, build the refactoring inventory, and order it; done when the ordered inventory is presented and, if it contains a High-risk row, signed off. Read [workflow.md](references/workflow.md) first.
- **Execute mode** — one atomic change per PR, landed on a refactoring branch; parallel only when the changes are file-disjoint, sequential when they overlap. Don't chain dependent refactor stages without human review between merges — a stage built on an unreviewed one inherits every wrong call in it. Done when every inventory row has landed on the refactoring branch, the `REFACTOR(` marker sweep is clean, and the final draft PR to `main` is open.
- **Simple-sweep mode** — a single mechanical, behavior-preserving transform applied tree-wide (one `gofmt -r` rule, one `eg` template, one fixer); it is one step, so it needs no staging. Done when the tree is green and committed.
- **Review mode** — reviewing a refactoring PR: verify structural/behavioral separation and behavior preservation. Deliverable: ranked findings with file:line; if the user asked for fixes, apply them and re-verify.

**Questions:** The sign-off gates in this skill — the moves listed under "Pause for human sign-off" in Hard Rules, and High-risk rows in Plan mode — are asked through the environment's question tool, never as plain-text prose the reader might skim past; an unnoticed "assumed yes" on an irreversible move is expensive to undo.

**Dependencies:** `gopls` (primary actuator) — `go install golang.org/x/tools/gopls@latest`; wiring it into the harness → See `samber/cc-skills-golang@golang-gopls` skill. Optional: `golangci-lint`, `benchstat`, `deadcode`, `eg`, `gopatch`.

# Go Refactoring — Safe Change at Scale

## The Core Loop

**Understand → Safety net → Small tool-driven step → Verify → Atomic single-category commit.** Repeat.

1. **Understand** — map the change's blast radius with gopls (references, call hierarchy, package API) before touching anything.
2. **Safety net** — before touching code with inadequate coverage, add tests first, gating the strategy on the _blast radius's_ coverage, not global coverage. A green suite you wrote yourself is what lets you tell "this is behavior-preserving" from "I hope this is behavior-preserving" — it is your check, not a formality for the reviewer.
3. **Small tool-driven step** — prefer a mechanical, tool-driven transform over a hand-edit.
4. **Verify** — `go build ./... && go vet ./...`, and `go test` on the changed packages while iterating; run `go test ./...` before each commit or PR, because a refactor's blast radius crosses packages. Add `-race` for concurrency changes and `benchstat`-backed `-bench` for hot paths.
5. **Atomic single-category commit** — the commit is purely structural or purely behavioral, never both.

## Hard Rules

- **Never mix structural and behavioral changes in one commit or PR** — a rename and a feature need different review postures, and mixing them denies the rename the fast, low-scrutiny review it deserves.
- **Split a code move from a code optimization into two sequential PRs**, even though both are structural — the move is proven by gopls plus build/test, the optimization needs benchmarks and a closer correctness read, and since they touch the same code they run one after another, not in parallel worktrees. Aim for **100–500 lines per PR**: small enough to review in one sitting, large enough to read as one coherent change.
- **Prefer gopls Rename/Inline over hand-edits** — both are behavior-preserving by construction (Rename refuses on shadowing, interface-satisfaction breakage, or malformed code; Inline moves side-effecting arguments into `var` temporaries instead of duplicating them), while a hand-edit across dozens of call sites measurably misses cases.
- **When a change recurs across many sites, generate a rewrite tool instead of hand-editing each site** — escalate `gofmt -r` → `eg` → `gopatch` → a `go/analysis` fixer; a tool is reviewable, re-runnable, and testable against golden files, and dozens of hand-edits are none of those.
- **Use a type alias (`type A = B`) for every type moved across packages** — it is the officially blessed mechanism for gradual code repair: old and new names stay interchangeable while callers migrate, so no commit has to touch every call site at once.
- **Break import cycles with a consumer-side interface first**, before considering a package split or a shared leaf package — Go satisfies interfaces implicitly, so the producer never has to import the consumer's interface.
- **Pause for human sign-off before** any cross-package move or package split, any exported-API change or deprecation, any deletion, a new major version, or touching code that has no tests — these are the moves a wrong call makes expensive to undo.
- **Grep for tag and reflection references after any rename** — gopls Rename guards only against _compilation_ breakage; it can't see a struct tag, a `text/template` field reference, or `reflect`-driven dispatch, so a renamed field silently desyncs from its `json`/`db` tag.
- **Load `samber/cc-skills-golang@golang-security` (and `golang-safety` for internal-correctness risk) whenever a step changes code logic, not just its shape** — a tool-verified mechanical transform can't introduce a vulnerability, but a behavioral change can.
- **Start every step from a clean, committed baseline, and revert rather than debug forward when it goes red** — commit the moment a step goes green; reverting to that commit and retrying beats patching forward inside a state you no longer fully trust.

## When Not to Refactor

Refactoring pays off only if a future change will spend it. Question it — or skip it — when:

- **Nothing planned will touch the code again** — a stable, rarely-read package earns nothing from being restructured for its own sake.
- **It's critical production code with no tests** — don't refactor it directly; build the characterization-test baseline and get sign-off first.
- **The deadline is tight** — a staged refactor needs review bandwidth between PRs, so under time pressure it either stalls or gets rushed; make the minimal safe change now and stage the larger refactor for later.
- **There's no clear purpose** — no upcoming feature it eases, no bug class it closes, no smell a review flagged; confirm the purpose with the user while planning rather than assuming one.

## Risk Stratification

| Risk | Transforms | Safety requirement |
| --- | --- | --- |
| **Low** | gopls Rename, Extract Variable/Constant, Inline Variable, `gofmt -s`, organize imports, local `refactor.rewrite.*` actions | Build/vet/test after the step is enough |
| **Medium** | Extract Function/Method (Extract is best-effort — verify comments/behavior survived), Inline Call across packages, single-parameter add/remove, introducing generics | Add or confirm targeted tests over the blast radius first |
| **High** | Change signature across many callers, moving types/functions across packages, splitting/merging packages, breaking import cycles, exported-API or major-version changes | Full safety net + human checkpoint before landing |

**Diagnose:** 1- gopls refusing a Rename or Inline is a real semantic hazard, not a tool bug — investigate the shadowing/interface conflict before forcing the change by hand 2- `go vet ./...` / `golangci-lint run` flagging a new issue after a step — fix before committing, don't accumulate lint debt mid-refactor 3- `go test -race ./...` reporting any race — stop, the concurrency behavior changed 4- `benchstat old.txt new.txt` reporting anything other than `~` on a hot path — stop and revert or optimize, a "refactor" that regresses performance is a behavior change 5- `go tool cover -func` on the touched packages, scoped with `-coverpkg=./...` — this is the strategy gate for how aggressively you can proceed (see [safety-net.md](references/safety-net.md))

## References

- [workflow.md](references/workflow.md) — read before planning any multi-step refactor: the planning gate and refactoring inventory, the three interacting orderings, the `refactor/<topic>` branch and per-change PR model, parallel versus sequential execution, and the `// REFACTOR(step N): ...` marker convention.
- [safety-net.md](references/safety-net.md) — read before touching code whose blast radius has thin coverage: the HIGH/MEDIUM/LOW thresholds, characterization and golden-testing recipes, and the verification command reference.
- [catalog.md](references/catalog.md) — read when deciding which refactoring a code smell calls for: the Fowler catalog mapped to Go, with trigger, mechanics, tool, and risk for each entry.
- [go-tooling.md](references/go-tooling.md) — read when choosing the tool for a step or when a change recurs across many sites: gopls code actions and CLI, `gofmt -r`, `eg`, `gopatch`, `go/analysis` and `//go:fix inline`, `dave/dst`.
- [structural.md](references/structural.md) — read when moving types across packages, breaking an import cycle, or changing an exported API or major version.

## Cross-References

- → See `samber/cc-skills-golang@golang-naming` skill for what to rename identifiers _to_ — this skill owns _how_ to apply a rename safely at scale.
- → See `samber/cc-skills-golang@golang-project-layout` skill for target directory/package layout — this skill owns the mechanics of moving code there without breaking callers.
- → See `samber/cc-skills-golang@golang-modernize` skill for version-driven idiom updates (`interface{}`→`any`, `slices`/`maps`) — a distinct concern from structural refactoring, though it shares the same tool-first discipline.
- → See `samber/cc-skills-golang@golang-code-style` skill for control-flow clarity and function-shape rules this skill helps you apply mechanically.
- → See `samber/cc-skills-golang@golang-design-patterns` skill for target patterns (options struct, DI, consumer-side interfaces) this skill helps you migrate toward.
- → See `samber/cc-skills-golang@golang-testing` skill for the test-writing practices that make the safety net in this skill trustworthy.
- → See `samber/cc-skills-golang@golang-lint` skill for configuring `golangci-lint`, run here only as a post-step verification gate.
- → See `samber/cc-skills-golang@golang-security` skill (and `golang-safety`) for reviewing any step that changes code logic, not just its shape.

If you encounter a bug or unexpected behavior in `gopls`, open an issue at <https://github.com/golang/go/issues>.
