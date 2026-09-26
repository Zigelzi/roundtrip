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
| 04 / slice 1: basics generated with the trip (S1 to S7) | B | Opus tests + Opus impl | Opus test authoring ~15k context growth (estimate). Opus sub-agent 66k, 9 tool calls, 1m35s. Main session re-read ~5k | 11 files, +235/-9 (55 generated sqlc) | 0 | First run with Opus as implementer (Human-PM's call: best cost/quality now). Agent changed no test and no spec requirement, made 3 small choices the spec left open (500 on failure, no same-name rule in basic_item so S7's setup works, transaction in cmd/web). Fewer tool calls than either 03 run despite a bigger change. |
| 04 / slice 2: 14-day cap and double-tap guard (S11, S12) | B | Opus tests + Opus impl | Opus test authoring ~6k (estimate). Opus sub-agent 43k, 7 tool calls, 51s. Main session re-read ~3k | 3 files, +18/-3 | 1 | Defect: the double-tap guard never fired, because its script ran before the button below it was parsed, so the lookup found nothing. Passed the diff reads by both the agent and the main session; only the phone check caught it, which is what S12 being manual was for. Fixed inline (a7e17c0). A 15-line change cost 43k: almost all of it is the spawn's fixed start-up cost. Inline in the main session this would have been a few thousand tokens. For slices this small the spawn overhead dominates, whichever model. |
| 04 / slice 3: quantity editing (S8 to S10) | B | Opus tests + Opus impl | Opus test authoring ~20k (estimate). Opus sub-agent 71k, 11 tool calls, 3m11s. Main session re-read and a 2-line fix ~6k | 8 files, +682/-139 (about 560 generated templ) | 4 | Defects, all UI clarity that every test passed: the open row did not stand out, Cancel did not read as a button, locked rows still looked tappable, and autofocus raised the keyboard and shifted the view. Plus one change Human-PM asked for, not counted: keep the view in place (htmx swap of all lists, e166137), built inline in the main session because it was four attributes. | Largest slice of the milestone and the spawn cost is closest to earning its keep here. Agent went beyond the brief sensibly: add forms carry ?edit= so adding while a row is open keeps the lock (fixed an edge case the main session had accepted). One test defect was the main session's: the S10 check pinned Cancel to the bare trip URL, which forced Cancel to land at the page top; loosened to the row anchor and fixed inline (e9d14e4). |

Reference points that are not A-vs-B data but sit in the same budget:

| Spawn | Model | Tokens | Value returned |
|---|---|---|---|
| 03: red-team on the spec | Opus | 107k, 19 tool calls, 3m31s | 11 findings, 3 blocking. One (bucket id 5 silently disarming the `FAMILY_NAMES` guard) was a real bug the spec asserted the opposite of. |
| 04: red-team on the spec | Opus | 58.5k, 9 tool calls, 1m36s | 10 findings: 6 product decisions (quantity editing UI, packed items, trip length cap, double tap, phone length, basics corrections) and 4 technical. |
| 04: second red-team, scenario wording only (Human-PM asked) | Opus | 44.5k, 4 tool calls, 59s | 8 findings, 7 applied. Read two files and no code, yet cost about three quarters of the full pass: most of a spawn's cost is fixed start-up, not reading. |
| 04: code review of the branch diff | Opus | 77.5k, 13 tool calls, 1m50s | 5 findings, none blocking. 2 fixed (a misleading comment, 4 missing tests), 1 declined, 2 accepted as known edge cases. Also verified the 75 seed rows against the catalogue, though it miscounted them as 76. |
| 04: front-end debugging of the edit transition flicker (Human-PM asked) | Opus | 102.5k, 48 tool calls, 12m40s | Measured the root cause instead of guessing: the old view-transition snapshot rasterised blurrier than the live page at phone pixel density, so unchanged text changed sharpness mid-fade. Fixed by removing cross-fades and painting section cards in the transition group (9cc6dfb). Most expensive spawn of the milestone; the only one that ran a real browser (headless Edge on the Windows host, via WSL interop). |

## Running read

Too early to conclude, and the one implementation run so far does not count: slice 1 had the agent write its own tests, which is neither option. The definitions were sharpened afterwards (2026-09-20) so that both options share Opus-written tests and differ only in where implementation happens. That is the comparison worth making, because it isolates one variable.

First real signal, from milestone 03's two runs. The proper Option B slice cost 97k and 39 tool calls against the outlier's 122k and 48, but that comparison flatters it: the outlier also wrote its own tests. Counting the main session's test authoring (~18k) the two are close to level on tokens. So the case for pre-written tests is not that it is cheaper; it is that the implementer is judged by a contract it did not write, and this run showed the agent honouring that (it hit a test it could not pass by its first design and changed the design, not the test).

A separate reading from the same session, worth keeping: after a full spec cycle, a red-team pass and a slice, the main session sat at 172k of 1m tokens with 141.6k of it messages. The "main session context grows too much" worry that motivates Option B is not biting at this repo's size. Which means the case for splitting rests on test independence, not on tokens.

**The comparison only works if a slice of milestone 03 runs on Option A.** Comparing across milestones would confound the allocation with the work being different.

That chance is now a single slice. A probe after slice 1 showed the packing page's Family section, progress and filter already worked from the seeded row alone, so the planned slice 2 collapsed to one line (Family second in the filter row) and was merged into slice 3. Milestone 03 therefore ends with one more slice, and it either runs on Option A or this milestone produces no Option A data at all.

A lesson for the slice plan itself, worth more than the token counts: slices drawn before the first one lands can be wrong about how much is left, and storage option 1 delivered more for free than the spec's own options table predicted. Re-check the remaining slices after each one, with a probe rather than by reading the code.

What to watch, from the constitution: total tokens, how soon the first slice was testable on a phone, defects found at the phone check that an earlier stop would have caught, and whether Human-PM felt in the loop.
