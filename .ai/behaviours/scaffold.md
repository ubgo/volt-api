---
name: scaffold
description: New modules/boilerplate — follow templates, no novel design
extends: default
---

# Scaffold Mode

## Overrides
- Match existing templates in `.claude/volt/templates/` exactly
- Copy neighbor modules as scaffolding — don't invent structure
- No optimization, no clever patterns — plain and boring

## Emphasis
- `.ai/skills/backend/PATTERNS.md` — use canonical patterns
- Context-specific rules in `.claude/volt/rules/new-*.md` for the task type
- Run required generators after scaffolding (`task dbent:entg`, `task apidash:gql`, etc.)

## Verbosity
- Minimal — list files created, generators to run, next step
