# Design canvas

The mobile design canvas (a Design artifact on claude.ai): https://claude.ai/artifact/MhRgkSTt11nyBjsGZL8pwD

The app is the source of truth. The canvas follows it after a milestone (step 7 in [`../spec/constitution.md`](../spec/constitution.md)), and is also the place to prototype an idea before building it. The decisions made on it and the original brief are in [`claude-design-prompt.md`](claude-design-prompt.md) and [`../spec/mobile-redesign.md`](../spec/mobile-redesign.md).

## Frames

Each frame has an ID in its title on the canvas (`F1 Trips`), so "update F8" is enough to name one. IDs are stable: a new frame takes the next free number, a removed one leaves a gap, never renumber.

| ID | Frame | Board file (`project/`) | App template and the part that is drawn |
|---|---|---|---|
| F1 | Trips | `Main.dc.html` | `cmd/web/view/index.templ`: `Index`, `tripRow` |
| F2 | New trip | `NewTrip.dc.html` | `trip_new.templ`: `NewTrip`, the nights slider |
| F3 | Trip page | `TripPlan.dc.html` | `trip_page.templ`: `TripPage`, `memberSection`, `packingBar` |
| F4 | Packing, everyone | `Packing.dc.html` | `pack_page.templ`: `PackPage`, `FilterNav`, `PackedPart` |
| F5 | Basics, edit open | `Basics.dc.html` | `basics_page.templ`: `BasicsPage`, `editBasicRow` |
| F6 | Empty: no trips | `TripsEmpty.dc.html` | `index.templ`: the "No trips yet" branch of `Index` |
| F7 | Validation error | `NewTripError.dc.html` | `trip_new.templ` with `shared.templ`: `errorList` |
| F8 | Trip page: edit open | `TripEdit.dc.html` | `trip_page.templ`: `editItemRow` |
| F9 | Packing: one person | `PackingPerson.dc.html` | `pack_page.templ`: `filterLink` (active), `toPackRow` |
| F10 | Packing: all packed | `PackingDone.dc.html` | `pack_page.templ`: `doneNote`, `PackProgress` |
| F11 | Design tokens | `Tokens.dc.html` | `cmd/web/tailwind.css`: `@theme`; sizes from `shared.templ` |

Shared pieces (buttons, stepper, jump links, input) live in `shared.templ` and show up in several frames. When one changes, update every frame that draws it.

## Syncing a frame

Edit only the frames the milestone touched, so the sync stays cheap.

1. Read the frame's template in the table (just the named part) and its board file: Artifact `read` with `path: project/<board file>`. Read `project/canvas.json` only to add, remove, move or retitle a frame.
2. Edit a copy of the board file in the scratchpad, under a folder that mirrors `project/`.
3. Publish with `url`, `root` (that folder), `file_path` (one changed file) and the other changed files in `files`. Send `canvas.json` only if the layout or a title changed.
4. A new frame gets the next free ID, a title starting with it, and a row in the table above.

Notes that save a wrong turn:
- The `DesignSync` tool is for design-system projects. It does not apply to this canvas.
- The shared styles are an uploaded stylesheet referenced by every board. To differ from it, override the rule in the board's `<helmet>` instead of re-uploading it; re-uploading means repointing every board.
- Frames are drawn by hand, not generated from the templates, so a sync is a manual read-and-compare. Nothing tests it; a skipped sync fails silently.
