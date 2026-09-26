# Spec structure & how to write specs

Read this first before adding or editing anything under `spec/`. It defines where specs live, what each file is responsible for, and how a spec gets written. The workflow that consumes these specs (red-team → refine → implement → review) lives in [`constitution.md`](constitution.md).

## Guiding rule

One fact lives in exactly one file. When two files would state the same thing, one links to the other instead of repeating it. Duplicated docs drift apart, and that is the failure mode this structure exists to prevent.

## File responsibilities

| File | Responsibility |
|---|---|
| `constitution.md` | How we work: agentic workflow, tech stack, git conventions, Milestone 0. Governance. |
| `user-flows.md` | Who / why at product level: the family, the high-level flows. |
| `domain-model.md` | Shared entities (Trip, Item, …). Created lazily, only when the first feature introduces real entities. Until then it does not exist. |
| `activity-catalogue.md` | The activities and their default items, plus the design decisions behind them. Feeds milestone 04 and the activity milestones after it. |
| `allocation-log.md` | Running record of the main-session vs sub-agent experiment, one row per slice. |
| `milestones/NN-name.md` | One unit of work. See below. |
| `milestones/_template.md` | The template new milestone specs are copied from. |

## Milestones

```
spec/milestones/
  _template.md
  00-walking-skeleton.md
  01-<name>.md
  ...
```

- One milestone = one file = one branch = one squashed commit. This 1:1:1:1 chain means the filename mirrors the branch description (`01-packing-list.md` ↔ `feature/packing-list`), and reviewing a milestone means reading one spec against one diff.
- Number by build order (`00`, `01`, …). Copy `_template.md` to start a new one.
- Each spec has four sections (kept deliberately minimal):
  - Status + Branch: the spec is the status tracker; there is no separate tool. `draft → red-teamed → in-progress → done`.
  - Why: the problem and user value, and which `user-flows.md` step it serves.
  - Scope (In / Out): the Out list is the main lever against scope creep: state what this milestone deliberately does not do.
  - Acceptance conditions (BDD): Given/When/Then, grouped into named scenarios. Write them so someone with no code knowledge (Parent 2) can read and challenge them: no table, query or attribute names. Every scenario has all three of Given, When and Then. How a scenario is tested goes in a separate Test notes section below the scenarios. Name each from the user's perspective (a capability or something they see); name a scenario with no real user (an infrastructure/build check) by the behaviour proven instead. Each scenario becomes a `go test` case, written before implementation. This is the section that matters most.
  - Number scenarios `S1`, `S2`, … in the heading (`### S3: Parent creates a trip`) so they can be referenced in conversation, open items and code review. Numbers are stable: a new scenario takes the next free number, a removed one leaves a gap; never renumber. The matching test carries the number in its comment (`// S3: Parent creates a trip`). Outside the spec, prefix the milestone: `01/S3`.
  - Verify each scenario the most feasible way. A `go test` is the default; behaviour a Go test can't observe (e.g. keyboard focus on a phone) is marked `Verification: manual` under the scenario and checked by Human-PM at step 7. Still test the part that is observable (e.g. the server's reply) automatically.
  - Open questions: where a draft hands off to the red-team pass and back to Human-PM. Number them `Q1`, `Q2`, … so they can be referenced in conversation and in the decisions log. Same stability rule as scenarios: never renumber, a resolved one keeps its number and moves to the decisions log.

## Referencing things by number

One letter per kind of thing, so a number is never ambiguous across documents:

| Prefix | What | Lives in |
|---|---|---|
| `S` | A BDD scenario | A milestone spec. Prefix the milestone outside it: `03/S8`. |
| `Q` | An open question about that milestone | The same milestone spec. |
| `A` | An open item about the activity catalogue | [`activity-catalogue.md`](activity-catalogue.md), because those questions outlive any one milestone. |

All three follow the scenario rule: numbers are stable, a removed one leaves a gap, never renumber. Note that `constitution.md` separately uses "Option A" and "Option B" for the allocation experiment; that is unrelated to the `A` prefix here.

## Starting a milestone

Open a fresh session and describe what you want to build; the session creates the `feature/<name>` branch and scaffolds the spec file from the template, then you write the spec into it. (Mechanics live in the workflow in `constitution.md`; this section is about the describing.)

A good kickoff description covers:

- The user goal: what the person is trying to do and why, from their point of view. "Let me set up a trip so I can plan what to pack," not "add a trips table."
- The thin slice: the smallest end-to-end path that delivers that goal. If it doesn't fit in a sentence, it is probably two milestones; say so and it gets split.
- One user goal per milestone. A kickoff joining two goals with "and" ("create a trip and add items to it") is two milestones, even if it fits in one sentence. Rough ceiling: about 8 scenarios. Milestone 01 bundled both goals and grew to 18 scenarios and 29 commits.
- In / out: what this milestone deliberately leaves for later. This is the strongest guard against scope creep.
- Which `user-flows.md` step it serves.
- What you are unsure about: these seed the red-team pass and the open questions.

Leave out the how: no database schema, no page layout, no test wording. You are describing intent; the refine step turns it into precise acceptance conditions and implementation decides the design. Rough is fine; red-team and refine sharpen it.

Example kickoff:

> Milestone 1: let me create a trip (name + dates) and add the activities we plan to do on it (swimming, hiking, …), so later I can derive a packing list from them. In scope: creating one trip and listing/adding its activities, on mobile. Out: the packing list itself, editing/deleting, multiple trips at once, auth. Serves the "plan activities" step. Unsure: do activities come from a fixed list or free text?

That is enough for the session to scaffold and for you to expand into Why + Scope + rough scenarios.

## How a spec gets written

This maps onto the workflow steps in `constitution.md`:

1. Human-PM drafts Why + Scope + rough acceptance conditions + open questions. Light and minimal: intent, not perfect BDD.
2. Red-team pass (a spawned sub-agent) pokes it for weak points and adds to Open questions.
3. Human-PM resolves the open questions.
4. Main session refines the acceptance conditions into precise Given/When/Then that become the tests.

You are not expected to write final BDD up front: you write intent, and it gets sharpened in step 4.
