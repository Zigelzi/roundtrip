# Spec structure & how to write specs

Read this first before adding or editing anything under `spec/`. It defines where specs live, what each file is responsible for, and how a spec gets written. The workflow that consumes these specs (red-team → refine → implement → review) lives in [`constitution.md`](constitution.md).

## Guiding rule

One fact lives in exactly one file. When two files would state the same thing, one links to the other instead of repeating it. Duplicated docs drift apart — that is the failure mode this structure exists to prevent.

## File responsibilities

| File | Responsibility |
|---|---|
| `constitution.md` | How we work: agentic workflow, tech stack, git conventions, Milestone 0. Governance. |
| `user-flows.md` | Who / why at product level: the family, the high-level flows. |
| `domain-model.md` | Shared entities (Trip, Item, …). Created lazily — only when the first feature introduces real entities. Until then it does not exist. |
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
  - Status + Branch — the spec is the status tracker; there is no separate tool. `draft → red-teamed → in-progress → done`.
  - Why — the problem and user value, and which `user-flows.md` step it serves.
  - Scope (In / Out) — the Out list is the main lever against scope creep: state what this milestone deliberately does not do.
  - Acceptance conditions (BDD) — Given/When/Then, grouped into named scenarios. Name each from the user's perspective (a capability or something they see); name a scenario with no real user (an infrastructure/build check) by the behaviour proven instead. Each scenario becomes a `go test` case, written before implementation. This is the section that matters most.
  - Open questions — where a draft hands off to the red-team pass and back to Human-PM.

## How a spec gets written

This maps onto the workflow steps in `constitution.md`:

1. Human-PM drafts Why + Scope + rough acceptance conditions + open questions. Light and minimal — intent, not perfect BDD.
2. Red-team pass (a spawned sub-agent) pokes it for weak points and adds to Open questions.
3. Human-PM resolves the open questions.
4. Main session refines the acceptance conditions into precise Given/When/Then that become the tests.

You are not expected to write final BDD up front — you write intent, and it gets sharpened in step 4.
