# LLM Constitution

This projects purpose is to learn and try agentic coding workflows ("vibe coding") and how to build a small scale prototype of a product that can grow into scalable commercial product.

## Agentic workflow

Two actors: Human-PM (me — product decisions) and the main session (the agent driving the work and, where noted, spawning sub-agents).

Goals:
1. Optimize development speed and the feedback loop by right-sizing tasks.
2. Minimize token usage to keep costs in check and be environmentally conscious.
3. Maximize code quality and maintainability.

**When goals conflict, speed and token minimization win over practicing elaborate multi-agent orchestration.** This workflow is deliberately lighter than "full" multi-agent ceremony — at this scale (one family, ~2 users) the app doesn't justify the cost, and the main session does most phases inline. Sub-agents are spawned only at the two points where independent context genuinely pays for itself: red-teaming the spec, and reviewing the diff. (A spawned agent starts cold and re-derives context, so each spawn has a real token cost.)

Version control is part of the workflow. Human-PM makes the initial baseline commit (spec + CLAUDE.md); after that all commits are made by the main session, not the Human-PM — this is a standing authorization to commit without asking each time. Each milestone runs on its own branch off `main`; each vertical slice is one focused commit on that branch once its acceptance test is green. The branch diff is what the code-review loop reviews, and the branch merges to `main` only after Human-PM accepts it.

Conventions:
- Commit messages follow Conventional Commits: `type(scope): summary` (types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`). Per-slice commits on the branch are working checkpoints.
- Branch names follow [Conventional Branch](https://conventionalbranch.org/): `<type>/<description>`, lowercase `a-z0-9` + hyphens, no leading/trailing/consecutive hyphens (e.g. `feature/packing-list`). Use the purpose prefix matching the milestone (`feature`/`fix`/`chore`/`hotfix`/`release`).
- Squash before merge. On acceptance, squash the branch's slice commits into a single Conventional Commit so `main` stays a clean, linear log of one commit per milestone.
- Remote: a private GitHub repo will be added as remote later. Until then merges are local (`git merge --squash`); once the remote exists, use squash-merge PRs.

### Workflow:
1. Human-PM writes a minimal spec: what to build and why.
2. Spawn a red-team agent on the spec — weak points, clarifying questions, usability/feasibility risks.
3. Human-PM decides which findings to accept. *(checkpoint)*
4. Main session refines the spec into milestones that deliver a feature/flow E2E, each with Behaviour-Driven Development (BDD) acceptance conditions. Those conditions become acceptance tests before implementation — written independently of the code that will satisfy them.
5. Main session implements milestone by milestone as vertical slices, sequentially, each slice satisfying its BDD acceptance test. Parallel dev sub-agents only if slices provably share no files. A task is right-sized if one agent can implement and self-verify it in one context without re-loading the whole app. Each slice is committed to the milestone branch once its test is green.
6. Code-review loop (bounded).
   a. Spawn a code-review agent (`/code-review`) on the finished feature's diff. It tags each finding blocking (correctness / bug / security / a failing BDD condition) or non-blocking (style / preference).
   b. Implementer addresses each: fix it, or decline a non-blocking one with a one-line reason.
   c. Disagreements resolve by type — factual (is there a bug?): settle with a test, the test decides; subjective (too complex?): default to the simpler option; product/scope in disguise: escalate to Human-PM as a plain-language trade-off, never as a technical vote.
   d. Max 2 rounds. Blocking findings must be fixed or proven-not-a-bug before the feature is done; anything unresolved escalates to Human-PM.
7. Report to Human-PM: what's done, what to test/verify, and request feedback → continue / complete / discard. On accept, the main session merges the milestone branch to `main`. *(checkpoint)*

Skills vs sub-agents. A skill encodes how to do a repeatable task; a sub-agent provides isolated fresh context — orthogonal, and they compose (a sub-agent can run a skill). Reuse the built-in skills that already fit: `/code-review` (step 6), `/verify` and `/run` (step 7). Do not build custom skills upfront — extract one (e.g. a red-team-the-spec skill) only after running the procedure manually a couple of times, so it's shaped by real use. Refine / implement / orchestrate stay as prose here, not skills.

Spec organization and how specs are written: see [`README.md`](README.md).

## Tech stack

Chosen to maximize interpretability (I'm a junior dev and won't hand-edit much, but want to be able to read what's happening), minimize dependencies and token cost, and deploy simply to the Raspberry Pi. Live collaboration is not a v1 requirement, and a future commercial version would be a separate native mobile build — so neither argues for a JS/reactive stack now.

- Go — backend language; the language I'm most familiar with, compiles to a single static binary for trivial Pi deployment.
- templ — type-safe HTML templating in Go (https://templ.guide/llms.md).
- HTMX — server-driven interactions; responses are literal HTML, the most transparent thing to debug. Live sync, if ever wanted, is a later incremental milestone via SSE.
- TailwindCSS — styling.
- Alpine.js — chosen tool for client-side reactive interactions (live filtering, "select all", counters), added only when a milestone first needs one. Native HTML (`<details>`, `<dialog>`) is preferred for simple toggles/modals; not a baseline dependency.
- SQLite — database; a single file with no separate server process, backs up by copying one file. Its only weakness (many concurrent writers) is irrelevant at ~2 users.
- Goose — versioned database migrations (up/down SQL), same as go-tiimit.
- sqlc — generates type-safe Go from hand-written SQL queries; keeps queries readable (interpretability) while giving compile-time checking. Adds a `sqlc generate` codegen step to the build.

## Milestone 0: walking skeleton

Before any feature, build the thinnest end-to-end slice that exercises the whole stack locally: one HTTP route → renders one templ page → from one SQLite query, styled with Tailwind. It does nothing useful; its only job is to prove the stack integrates and runs on the dev machine before features are stacked on top. Pi deployment is deferred — do not wire up deploy in Milestone 0; it becomes its own later milestone.

Stack constraints this milestone locks in (so agents don't pick heavier defaults):
- BDD conditions map to standard `go test` (table-driven), not a Gherkin framework like `godog`.
- Tailwind via the standalone CLI binary (no Node/npm toolchain), consistent with the single-binary, no-Node approach.
- Migrations via Goose, type-safe DB access via sqlc (same tooling as go-tiimit): write SQL, sqlc generates the Go; `sqlc generate` runs as part of the build.

## Deployment

Deferred (not part of Milestone 0). When it's time: initial deployment goes to the same Raspberry Pi that go-tiimit uses, via the `/home/zigelzi/projektit/go-tiimit/deploy.sh` script. Later it can move to a proper hosted service.
