# Roundtrip

A packing app for one family that travels around Finland.

Roundtrip is where the two parents in our family plan a trip together, build the packing list for it, and tick items off while packing. It is deliberately built for us and nobody else: one family, two active users, no multi-tenancy, no configuration to speak of.

The part that makes it more than a shared checklist is what happens *after* a trip: marking which packed items actually got used, so the next trip's list gets shorter and better. That review is the thing current tools make too tedious to keep up, and it is the reason this app exists.

## ⚠️ Read this before judging the code

**This repository is an experiment in agentic coding. Nearly all of the code here was written by Claude Code, not by me. It is not a sample of my programming ability and should not be read as one.**

I'm a product manager by profession; software is a hobby I'm still fairly junior at. In my other projects I write the code myself, because I enjoy the problem solving. This project is the opposite on purpose: an honest test of how far a spec-driven, agent-written workflow gets when a non-expert drives it — and what that feels like from a PM's seat.

What I actually do here:

- decide what gets built and why, and write the intent for each milestone
- resolve the open questions a red-team pass raises against the spec
- accept or reject the result after testing it on my phone
- design the workflow itself (see below) and adjust it when it wastes time or tokens

What the agent does: writes the specs into testable acceptance conditions, writes the tests, writes the implementation, reviews its own diff with an independent reviewer agent, and commits.

So: if the code is good, credit the model and the workflow. If it is bad, that's a finding about the workflow — which is the point of running the experiment in the open.

## How the work is organised

The process lives in [`spec/`](spec/) and is the actual subject of the experiment:

- [`spec/constitution.md`](spec/constitution.md) — how the work is done: the agentic workflow, the two points where sub-agents are worth their token cost, git conventions, the tech stack and why each piece was chosen.
- [`spec/README.md`](spec/README.md) — how specs are written and structured.
- [`spec/user-flows.md`](spec/user-flows.md) — who uses this and what they are trying to do.
- [`spec/milestones/`](spec/milestones/) — one file per unit of work, each with Given/When/Then acceptance conditions that become Go tests before any implementation exists.

One milestone = one spec file = one branch = one squashed commit, so a milestone can be reviewed by reading one spec against one diff.

Three goals shape every decision: right-size tasks for a fast feedback loop, keep token usage (cost and footprint) low, and keep the code readable enough that I can follow what it does.

## Stack

Go + [templ](https://templ.guide) + [HTMX](https://htmx.org) + TailwindCSS, on SQLite with [Goose](https://github.com/pressly/goose) migrations and [sqlc](https://sqlc.dev) for type-safe queries. It builds to a single static binary with no Node toolchain (Tailwind runs via its standalone CLI). Rationale for each choice is in the constitution.

Mobile is the primary target — the app is used on a phone roughly all of the time.

## Running it

Requires Go 1.26+ and `make`.

```bash
make run   # http://127.0.0.1:8080
```

`run`, `run-lan`, `test` and `build` run codegen (sqlc + templ + Tailwind) first, so there is no separate setup step. A plain `go build` or `go test` does need `make generate` once on a fresh checkout — the Tailwind bundle is gitignored and the `//go:embed` fails without it.

Other targets:

| Command | What it does |
|---|---|
| `make test` | `go test ./...` — the BDD scenarios live in `cmd/web/*_test.go` |
| `make dev` | live reload (templ watch + Air + Tailwind watch), reachable from a phone on the LAN; after editing SQL, re-run `make generate` |
| `make run-lan` | plain run, bound to the LAN so a phone can reach it |
| `make build` | single static binary to `./build/roundtrip` |

There is no auth — run it on a trusted network only.

### Configuration

Copy `.env.example` to `.env` and fill it in; `make` loads it automatically and `.env` is gitignored. It is also where the family members' names come from — this repository is public, so the migration seeds placeholders (`Parent 1`, `Child 1`, …) and `FAMILY_NAMES` supplies the real ones at startup. The app runs fine without it, with the placeholder names.

## Layout

```
cmd/web/        entrypoint, HTTP handlers, templ templates in view/
internal/db/    sqlc-generated SQLite access
sql/            Goose migrations + the hand-written queries sqlc generates from
spec/           the specs and the workflow described above
```

Generated Go (`internal/db`, `*_templ.go`) is committed on purpose so it can be read and reviewed; `make generate` refreshes it.

## License

[MIT](LICENSE) — use it however you like, notice included.
