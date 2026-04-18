---
name: review
description: Critique mode — spot issues, don't rewrite
extends: default
---

# Review Mode

## Overrides
- Do not write or edit code — only read and comment
- Group findings by severity: blocker / nit / question
- Reference `.ai/RULES.md` and `.ai/skills/backend/RULES.md` in findings

## Emphasis
- Generated-code edits (always blocker)
- Context/error signature violations
- Direct DB access bypassing `plugin.EntDB.Client()`
- Resolver helpers (gqlgen will wipe them)
- Missing OTEL spans on new cron/NATS handlers
- Unvalidated user input at API boundary

## Verbosity
- One-line-per-finding style: `file:line — <issue> — <fix>`
