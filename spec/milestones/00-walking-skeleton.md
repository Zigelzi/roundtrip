# Milestone 00: Walking skeleton

**Status:** done · **Branch:** chore/walking-skeleton (merged)
<!-- Status: draft → red-teamed → in-progress → done -->

## Why
Prove the whole stack integrates and runs locally before any feature is built. Serves no user flow directly; it de-risks the stack and tooling.

## Scope
- **In:** one HTTP route → renders one templ page → greeting read from SQLite (driver: `modernc.org/sqlite`, pure Go for easy Pi cross-compile) via a sqlc query against a throwaway seed table, styled with the standalone Tailwind CLI, schema applied by Goose. A single documented, repeatable build+run command (including `sqlc generate` and the Tailwind build) is produced.
- **Out:** any real feature or user flow; authentication; Pi deployment (deferred to its own later milestone). The seed table is throwaway skeleton scaffolding, not the start of the domain model; it is replaced when the real model arrives.

## Acceptance conditions (BDD)

### Scenario: Visitor sees a personalized greeting
- Given the app is running with a user "Parent 1" seeded by a Goose migration
- When the index route is requested
- Then the response contains a greeting to "Parent 1"
- And the name is read from the throwaway seed table via a sqlc query, not hardcoded
- And the index page links the Tailwind stylesheet

### Scenario: Tailwind stylesheet is generated and served
<!-- No real user, so named by the behaviour proven, not a "visitor" frame. -->
- Given the standalone Tailwind CLI has built the stylesheet
- When the stylesheet route linked by the index page is requested
- Then the response is non-empty
- And it contains rules for the Tailwind classes the index page uses

**Definition of done** (verified by Human-PM in step 7, not a `go test`): a single documented command builds (incl. `sqlc generate` + Tailwind) and runs the app locally.

## Open questions
- None open. SQLite driver resolved: `modernc.org/sqlite` (see Scope).
