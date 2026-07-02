# Your project rules

This folder is **your** space for context-specific rules — read on-demand by AI agents before a matching action (keeps them out of every prompt). Nothing here ships from Volt, so Volt updates never touch it.

## How to use

1. Add `[action-name].md` here (e.g. `pr-creation.md`, `deploy-checklist.md`).
2. Reference it from the root `CLAUDE.md` — add a line after the `@volt/CLAUDE.md` import so agents know when to read it.
3. Keep each rule focused: a clear title, a "read this before…" line, and required/forbidden sections.

## Volt's own rules

Volt ships its baseline rules — **code standards**, **git-commit conventions**, and the per-action Go guides — in [`../../volt/rules/`](../../volt/rules/), referenced by `volt/CLAUDE.md`. Edit those to change Volt's defaults, or override them with your own rule here (yours wins on conflict if you say so in `CLAUDE.md`).
