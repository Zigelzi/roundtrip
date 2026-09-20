# LLM Constitution

This projects purpose is to learn and try agentic coding workflows ("vibe coding") and how to build a small scale prototype of a product that can grow into scalable commercial product.

## Agentic workflow

Two actors: Human-PM (me, making the product decisions) and the main session (the agent driving the work and, where noted, spawning sub-agents).

Goals:
1. Optimize development speed and the feedback loop by right-sizing tasks.
2. Minimize token usage to keep costs in check and be environmentally conscious.
3. Maximize code quality and maintainability.

**When goals conflict, speed and token minimization win over practicing elaborate multi-agent orchestration.** This workflow is deliberately lighter than "full" multi-agent ceremony: at this scale (one family, ~2 users) the app doesn't justify the cost, and the main session does most phases inline. Sub-agents are spawned only at the two points where independent context genuinely pays for itself: red-teaming the spec, and reviewing the diff. (A spawned agent starts cold and re-derives context, so each spawn has a real token cost.)

Version control is part of the workflow. Human-PM makes the initial baseline commit (spec + CLAUDE.md); after that all commits are made by the main session, not the Human-PM; this is a standing authorization to commit without asking each time. Each milestone runs on its own branch off `main`; each vertical slice is one focused commit on that branch once its acceptance test is green. The branch diff is what the code-review loop reviews, and the branch merges to `main` only after Human-PM accepts it. The branch is created at the start of the milestone (step 1) and the spec is authored on it (branch-first), so a discarded milestone leaves nothing on `main`.

Conventions:
- Commit messages follow Conventional Commits: `type(scope): summary` (types: `feat`, `fix`, `chore`, `docs`, `refactor`, `test`). Per-slice commits on the branch are working checkpoints.
- Branch names follow [Conventional Branch](https://conventionalbranch.org/): `<type>/<description>`, lowercase `a-z0-9` + hyphens, no leading/trailing/consecutive hyphens (e.g. `feature/packing-list`). Use the purpose prefix matching the milestone (`feature`/`fix`/`chore`/`hotfix`/`release`).
- Squash before merge. On acceptance, squash the branch's slice commits into a single Conventional Commit so `main` stays a clean, linear log of one commit per milestone.
- Remote: a private GitHub repo will be added as remote later. Until then merges are local (`git merge --squash`); once the remote exists, use squash-merge PRs.

### Workflow:
1. In a new session, Human-PM describes what to build. The main session scaffolds the milestone: it creates the `feature/<name>` branch off `main`, copies `_template.md` to `NN-<name>.md` with Status and Branch filled, and commits it. Then Human-PM writes the minimal spec (what and why) into the file and the main session commits the draft.
2. Spawn a red-team agent on the spec: weak points, clarifying questions, usability/feasibility risks.
3. Human-PM decides which findings to accept. *(checkpoint)*
4. Main session refines the spec into milestones that deliver a feature/flow E2E, each with Behaviour-Driven Development (BDD) acceptance conditions. Those conditions become acceptance tests before implementation, written independently of the code that will satisfy them.
5. Main session implements milestone by milestone as vertical slices, sequentially, each slice satisfying its BDD acceptance test. Parallel dev sub-agents only if slices provably share no files. A task is right-sized if one agent can implement and self-verify it in one context without re-loading the whole app. Each slice is committed to the milestone branch once its test is green.
   - **Stop after every slice.** Commit it, hand Human-PM a 2–4 step phone checklist for that slice alone, and wait. This is the default, not a question to ask; chaining slices together and delivering one long checklist at the end is what this rule exists to prevent. Milestone 02 proved the cost: the packed section gave no sign it could be opened, a defect present from slice 1, but it surfaced only after all three slices and a review round were done, so the later slices were built on top of it.
   - **Before the first slice, raise the model choice** (see [Model & effort](#model--effort)). The main session says in one line what kind of work the slices actually are (mechanical, or a genuine design problem), and Human-PM decides. Rolling straight from a finished spec into code takes that decision away from him.
6. Code-review loop (bounded).
   a. Spawn a code-review agent (`/code-review`) on the finished feature's diff. It tags each finding blocking (correctness / bug / security / a failing BDD condition) or non-blocking (style / preference).
   b. Implementer addresses each: fix it, or decline a non-blocking one with a one-line reason.
   c. Disagreements resolve by type: factual (is there a bug?): settle with a test, the test decides; subjective (too complex?): default to the simpler option; product/scope in disguise: escalate to Human-PM as a plain-language trade-off, never as a technical vote.
   d. Max 2 rounds. Blocking findings must be fixed or proven-not-a-bug before the feature is done; anything unresolved escalates to Human-PM.
   e. Changes after review. Small fixes found while testing (wording, styling, a tweak to an existing scenario) go straight in with their tests. A new scenario either gets a short second review round on its own diff or moves to the next milestone; Human-PM picks.
7. Report to Human-PM: what's done, what to test/verify, and request feedback → continue / complete / discard. The report includes a phone checklist: a few concrete steps to try on a phone with `make dev`, each with what should happen, drawn from the milestone's scenarios, always covering the ones marked `Verification: manual`. On accept, the main session merges the milestone branch to `main`. *(checkpoint)*

Sessions. Start each milestone in a fresh session to keep context and token cost small; the spec, `CLAUDE.md`, and memory carry what a new session needs to pick up. Optionally start another fresh session at implementation (step 5); the finalized spec is on the branch for it to read. The spawned sub-agents (steps 2 and 6) run in isolated context, so they never require restarting the main session.

Skills vs sub-agents. A skill encodes how to do a repeatable task; a sub-agent provides isolated fresh context. The two are orthogonal, and they compose (a sub-agent can run a skill). Reuse the built-in skills that already fit: `/code-review` (step 6), `/verify` and `/run` (step 7). Do not build custom skills upfront; extract one (e.g. a red-team-the-spec skill) only after running the procedure manually a couple of times, so it's shaped by real use. Refine / implement / orchestrate stay as prose here, not skills.

Spec organization and how specs are written: see [`README.md`](README.md).

## Model & effort

Two dials trade cost for quality: model tier (Haiku fast/cheap → Sonnet balanced → Opus most capable) and reasoning effort (how much it deliberates before answering). The pattern worth learning: match both to two questions: how hard is the reasoning, and how costly is a mistake here? Spend capability where a mistake is expensive or the thinking is genuinely hard; save it where the work is mechanical.

Light default: don't over-tune it. At ~2 users, context size and number of turns cost more than model tier, so fresh sessions and targeted reads matter more than a per-step model matrix.

- Main session: see the two allocations below; which one we use is an open experiment.
- The two spawned critics (red-team in step 2, code review in step 6): strong model at high effort. They read small inputs (a short spec, a diff), so capability here is cheap in absolute tokens and high-leverage: a flaw caught pays for itself in avoided rework.
- Mechanical stretches (scaffolding, commits, the step-7 report): a fast, cheap model at low effort is fine.
- Code-review effort is a skill argument: `/code-review medium` by default, `high` for changes touching data integrity or auth, and avoid `ultra` (heavy, billed) unless a milestone is large or risky.

Decide per task with the two questions above, not by a fixed table.

### Where the main session's tokens actually go

Not duration. A long session of small turns is cheap. The cost is **context size × number of turns**: everything the session reads stays in its context and is re-sent every turn after (prompt caching softens this, it does not remove it), and the model tier multiplies against that growing prompt. So the thing worth isolating is not "coding" as an activity; it is the *file reading that coding drags in*. That is what the two options below trade differently.

### Two allocations: an open experiment

Decided 2026-09-20: draft both, run one per milestone, compare before concluding. What to compare afterwards: total tokens for the milestone, how soon the first slice was testable on a phone, how many defects Human-PM found at his check that an earlier stop would have caught, and whether he felt in the loop.

**Option A: one session, switch model at the boundary.**
- Opus for spec drafting, red-team triage, code-review triage and the step-7 report.
- At the step 4 → 5 boundary Human-PM switches to Sonnet (`/model`); the implementation slices run on Sonnet in the same thread. Switch back to Opus to interpret review findings.
- No cold start: every decision made in conversation is still there. Costs two cache re-warms (caches are per-model, so each switch re-pays the input once, so switch twice, not a dozen times).
- The main session's context keeps growing, because every file the implementation reads lands in it.

**Option B: Opus main session, Sonnet implementation sub-agents.**
- The main session stays Opus and never writes feature code: it specs, orchestrates, triages findings and reports.
- Each vertical slice is delegated to a Sonnet sub-agent that gets the milestone spec, that slice's scope and a pointer to the repo conventions. It writes the test first, implements, runs `make test`, commits, and returns a summary.
- The main session's context stays small, because the file reading happens in the agent and does not come back.
- Each spawn pays cold start (re-deriving spec + conventions + the patterns it must match; roughly 10–15k in this repo), and the main session partly re-pays it by reading the diff to report.
- **Only viable if decisions live in the spec, not in the conversation.** Milestone 02's `<details>` collapse hazard is the example: it came out of the red-team pass and was written into the spec's decisions log, which is the only reason an implementer with no memory of the discussion would have handled it.

Expected shape of the answer, to be confirmed by the experiment rather than assumed: at this repo's size (~15 files, a slice touching 2–4 of them) cold start is a large fraction of a slice, which favours A. On a codebase where a slice means searching hundreds of files, that flips and B wins clearly.

Either way the two critic spawns (red-team, code review) stay on a strong model. They read small inputs, so capability is cheap there in absolute tokens, and both earned their cost in milestone 02.

## Tech stack

Chosen to maximize interpretability (I'm a junior dev and won't hand-edit much, but want to be able to read what's happening), minimize dependencies and token cost, and deploy simply to the Raspberry Pi. Live collaboration is not a v1 requirement, and a future commercial version would be a separate native mobile build, so neither argues for a JS/reactive stack now.

- Go: backend language; the language I'm most familiar with, compiles to a single static binary for trivial Pi deployment.
- templ: type-safe HTML templating in Go (https://templ.guide/llms.md).
- HTMX: server-driven interactions; responses are literal HTML, the most transparent thing to debug. Live sync, if ever wanted, is a later incremental milestone via SSE.
- TailwindCSS: styling.
- Alpine.js: chosen tool for client-side reactive interactions (live filtering, "select all", counters), added only when a milestone first needs one. Native HTML (`<details>`, `<dialog>`) is preferred for simple toggles/modals; not a baseline dependency.
- SQLite: database; a single file with no separate server process, backs up by copying one file. Its only weakness (many concurrent writers) is irrelevant at ~2 users.
- Goose: versioned database migrations (up/down SQL), same as go-tiimit.
- sqlc: generates type-safe Go from hand-written SQL queries; keeps queries readable (interpretability) while giving compile-time checking. Adds a `sqlc generate` codegen step to the build.

## UI principles

The app is used on our phones ~99% of the time, so mobile is the primary target, not an afterthought. Prioritize every screen from the mobile viewpoint.

- Mobile first: build and test each feature at phone width first; a desktop/tablet layout is a bonus, never the driver.
- Function before polish: ship basic, unpolished styling and keep moving; a dedicated visual pass comes later, once the functional parts are far enough along.
- Usability over style on mobile: prefer large tap targets and simple interactions over visual refinement: when the two conflict, the easier-to-tap option wins.
- The line between the two: mobile usability essentials (tap-target size, readable text, reachable controls, obvious state) count as function, not polish; don't defer them under "basic styling."

## Milestone 0: walking skeleton

Before any feature, build the thinnest end-to-end slice that exercises the whole stack locally: one HTTP route → renders one templ page → from one SQLite query, styled with Tailwind. It does nothing useful; its only job is to prove the stack integrates and runs on the dev machine before features are stacked on top. Pi deployment is deferred: do not wire up deploy in Milestone 0; it becomes its own later milestone.

Stack constraints this milestone locks in (so agents don't pick heavier defaults):
- BDD conditions map to standard `go test` (table-driven), not a Gherkin framework like `godog`.
- Tailwind via the standalone CLI binary (no Node/npm toolchain), consistent with the single-binary, no-Node approach.
- Migrations via Goose, type-safe DB access via sqlc (same tooling as go-tiimit): write SQL, sqlc generates the Go; `sqlc generate` runs as part of the build.

## Deployment

Deferred (not part of Milestone 0). When it's time: initial deployment goes to the same Raspberry Pi that go-tiimit uses, via the `/home/zigelzi/projektit/go-tiimit/deploy.sh` script. Later it can move to a proper hosted service.
