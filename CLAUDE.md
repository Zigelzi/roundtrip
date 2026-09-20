# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project status

Milestone 01 (packing list, `spec/milestones/01-packing-list.md`) is merged: create trips (departure + return date or duration), list them upcoming first, and add/remove items per family member on a trip page, with items added in place via HTMX, earlier names suggested, one item per name per member. Milestone 0 (walking skeleton) proved the stack before it. Structure mirrors go-tiimit: entrypoint in `cmd/web` (`main.go`, `app.go`, `view/` templ templates), SQLite access in `internal/db` (sqlc-generated), Goose migrations + sqlc queries under `sql/`.

Commands (`make`, all wrap codegen via the `generate` target: sqlc, templ and the standalone Tailwind CLI):
- `make test`: run `go test ./...` (BDD scenarios live in `cmd/web/*_test.go`)
- `make run`: start the server at http://127.0.0.1:8080
- `make run-lan`: same, but reachable from a phone on the home network (sets `ADDR=0.0.0.0:8080`, prints the URL). No auth, trusted network only; WSL needs mirrored networking, a Hyper-V firewall rule for port 8080, and the Windows network profile set to Private. From the desktop itself use `localhost:8080`, since the LAN IP only works from other devices (unless `hostAddressLoopback` is enabled)
- `make dev`: live reload for development: templ watch + proxy on :8080 (reachable from a phone, same as run-lan), Air rebuilding the app on :8081, Tailwind watch. Edits show up in the browser without restarting. sqlc isn't watched: after editing SQL run `make generate`. htmx replies send `templ-skip-modify` so the proxy doesn't inject its reload script into fragments
- `make build`: single static binary to `./build/roundtrip`
- Run a single test: `go test ./cmd/web/ -run TestName -v`

Generated Go code (sqlc `internal/db`, `*_templ.go`) is committed so it can be read and reviewed; `make generate` refreshes it. The built Tailwind bundle (`cmd/web/static/tailwind.css`) is gitignored (only the input `cmd/web/tailwind.css` is tracked), so on a fresh checkout run `make generate` once before a plain `go build`/`go test` (the `//go:embed static` fails without it). The runtime SQLite file and `build/` are also gitignored.

Stack (decided; see `spec/constitution.md` for the per-choice rationale): Go + [templ](https://templ.guide/llms.md) + HTMX + TailwindCSS, backed by SQLite (single file, no server process), with Goose migrations and sqlc for type-safe DB access (same tooling as go-tiimit). Alpine.js is the chosen tool for client-side reactive interactions but is added only when a milestone first needs one; prefer native HTML (`<details>`, `<dialog>`) for simple toggles/modals. The stack optimizes for interpretability, few dependencies, and single-binary deploy to the Pi; live sync is not a v1 requirement.

## What Roundtrip is

A packing app for a single family that travels in Finland. It is a two-player collaboration app (the two parents pack together), intentionally not multi-tenant, with minimal configuration and ideally login via Google account only. See `README.md` and `spec/user-flows.md`.

Core user flow (`spec/user-flows.md`): plan trip activities → derive a packing list from them → gather items in one place → review what's missing and needs buying → pack once confirmed. A distinguishing goal is the post-trip review: mark which packed items were actually useful so the future packing list improves over time, something the user finds too time-consuming with current tools.

## How work is meant to be done here

`spec/constitution.md` is the source of truth for the workflow, and `spec/README.md` for how specs are structured and written. Read them, don't rely on this summary. It defines a deliberately lean spec-driven pipeline. The operative rules for a session working here:

- Two actors: Human-PM (the user, who makes the product decisions) and the main session (does most phases inline).
- Spawn sub-agents at only two points, where independent context earns its token cost: red-teaming the spec, and code-reviewing the diff. Everything else runs inline, because each spawn starts cold and re-derives context, which is expensive.
- When goals conflict, speed and token minimization win over practicing elaborate multi-agent orchestration. Prefer targeted work over broad codebase sweeps; flag approaches that cost notably more tokens.
- BDD acceptance conditions become tests before implementation, written independently of the code that satisfies them.
- Implement as vertical slices, sequentially (parallel dev agents only if slices provably share no files). Right-sized = one agent can implement and self-verify it in one context without re-loading the whole app.
- Code-review loop is bounded (max 2 rounds). Disputes resolve by type: factual → settle with a test; subjective → default to the simpler option; product/scope → escalate to Human-PM as a plain-language trade-off, never a technical vote.
- Starting a milestone: Human-PM opens a fresh session and describes what to build; the main session creates the `feature/<name>` branch (branch-first, at step 1) and scaffolds the spec from `_template.md` on that branch. The spec is authored on the branch, not on `main`.
- Commits are the agent's job, with standing authorization to commit without asking. Human-PM makes only the initial baseline commit. Milestone = branch off `main`; each vertical slice = one focused commit once its test is green; merge to `main` only after Human-PM accepts (step 7).
- Git conventions: Conventional Commits (`type(scope): summary`); branch names per [Conventional Branch](https://conventionalbranch.org/) (`<type>/<description>`, lowercase + hyphens, e.g. `feature/packing-list`); squash the branch's slice commits into one Conventional Commit before merging to `main` (keep a clean linear log). No remote yet, so merges are local (`git merge --squash`); once a private GitHub remote is added, use squash-merge PRs.
- Milestone 0 is a local walking skeleton (one route → templ page → SQLite query, Tailwind-styled) proving the stack end-to-end before any feature. Pi deployment is deferred to its own later milestone; don't wire up deploy early. Stack constraints: BDD conditions → standard `go test` (not `godog`); Tailwind via the standalone CLI (no Node); migrations via Goose + type-safe DB access via sqlc (same as go-tiimit; `sqlc generate` is part of the build).

The three goals shaping every decision: (1) right-size tasks for a good feedback loop, (2) minimize token usage for cost and environmental reasons, (3) maximize code quality and maintainability.

## Scope constraints to respect

- Single family, ~2 active users. Do not design for multi-tenancy or scale unless explicitly asked; treat scaling as a thought experiment and call out the specific point where it would actually matter.
- Keep configuration and auth minimal (Google login is the preferred ceiling).

## Writing style

No em dashes anywhere in this repo: specs, docs, code comments, commit messages, UI text. Use a colon, comma, semicolon, parentheses, or two sentences instead. Keep one only where it is genuinely the right punctuation and no other mark fits (for example quoting a source verbatim).

## Deployment

Initial target (per `spec/constitution.md`) is the same Raspberry Pi used for the `go-tiimit` project, deployed via `/home/zigelzi/projektit/go-tiimit/deploy.sh`. A proper hosted service may come later.
