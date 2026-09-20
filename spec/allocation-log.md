# Allocation log: main session vs sub-agents

Running record for the open experiment in [`constitution.md`](constitution.md#two-allocations-an-open-experiment): In both options the Opus main session writes the BDD tests first. They differ only in where the implementation happens: **Option A** in the main session on Sonnet, **Option B** in Sonnet sub-agents working against those red tests.

Append one row per slice. Keep it short: this is here to make a reasoned choice, not to produce a study.

## What the numbers mean

The token figures are not directly comparable and should not be read as "cost of the feature":

- **Option B** reports the sub-agent's total token use across its whole run. It is a real number, reported by the harness on completion.
- **Option A** has no equivalent single figure. Its cost is the main session's context growth multiplied by the turns after it, which is an estimate, not a reading.
- Either way the main session pays a **re-read** to review the diff and report. For B that is the whole diff cold; for A it is usually already in context.

So compare them on the other three columns first, and treat tokens as a rough signal.

## Log

| Milestone / slice | Option | Model | Tokens | Change | Defects at phone check | Notes |
|---|---|---|---|---|---|---|
| 03 / slice 1: Family bucket + trip page | **B (outlier)** | Sonnet | 122k sub-agent, 48 tool calls, 5m14s. Main session re-read ~8k (estimate) | 7 files, +190/-6 | 0 | **Not a valid Option B run:** the agent wrote its own tests as well as the implementation, so it is neither option as defined above. Kept for the token shape only, not for comparison. Spec was self-contained: agent reported nothing ambiguous and had to invent only 2 decisions (constant placement, migration timestamp). |

| 03 / final slice: filter order, S8 wording, S9 layout | B | Opus tests + Sonnet impl | Opus test authoring ~18k context growth (estimate). Sonnet 97k, 39 tool calls, 5m17s. Main session re-read ~6k | 4 files, +281/-184 (most of it generated templ) | 0 | First valid Option B run. Agent changed no test and no spec file, disputed nothing, and reported one real snag it solved itself (templ escapes strings through `{ }`, so the pinned apostrophe in "family's" only survives as static template text). |

Reference points that are not A-vs-B data but sit in the same budget:

| Spawn | Model | Tokens | Value returned |
|---|---|---|---|
| 03: red-team on the spec | Opus | 107k, 19 tool calls, 3m31s | 11 findings, 3 blocking. One (bucket id 5 silently disarming the `FAMILY_NAMES` guard) was a real bug the spec asserted the opposite of. |

## Running read

Too early to conclude, and the one implementation run so far does not count: slice 1 had the agent write its own tests, which is neither option. The definitions were sharpened afterwards (2026-09-20) so that both options share Opus-written tests and differ only in where implementation happens. That is the comparison worth making, because it isolates one variable.

First real signal, from milestone 03's two runs. The proper Option B slice cost 97k and 39 tool calls against the outlier's 122k and 48, but that comparison flatters it: the outlier also wrote its own tests. Counting the main session's test authoring (~18k) the two are close to level on tokens. So the case for pre-written tests is not that it is cheaper; it is that the implementer is judged by a contract it did not write, and this run showed the agent honouring that (it hit a test it could not pass by its first design and changed the design, not the test).

A separate reading from the same session, worth keeping: after a full spec cycle, a red-team pass and a slice, the main session sat at 172k of 1m tokens with 141.6k of it messages. The "main session context grows too much" worry that motivates Option B is not biting at this repo's size. Which means the case for splitting rests on test independence, not on tokens.

**The comparison only works if a slice of milestone 03 runs on Option A.** Comparing across milestones would confound the allocation with the work being different.

That chance is now a single slice. A probe after slice 1 showed the packing page's Family section, progress and filter already worked from the seeded row alone, so the planned slice 2 collapsed to one line (Family second in the filter row) and was merged into slice 3. Milestone 03 therefore ends with one more slice, and it either runs on Option A or this milestone produces no Option A data at all.

A lesson for the slice plan itself, worth more than the token counts: slices drawn before the first one lands can be wrong about how much is left, and storage option 1 delivered more for free than the spec's own options table predicted. Re-check the remaining slices after each one, with a probe rather than by reading the code.

What to watch, from the constitution: total tokens, how soon the first slice was testable on a phone, defects found at the phone check that an earlier stop would have caught, and whether Human-PM felt in the loop.
