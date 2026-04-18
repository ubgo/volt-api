---
name: careful
description: Maximum verification — auth, payments, migrations, prod infra
extends: default
---

# Careful Mode

## Overrides
- Confirm before every destructive or irreversible action
- Keep diffs small — one concern per edit
- No speculative refactors, no adjacent cleanup
- Read referenced files fully before editing — no assumptions

## Emphasis
- `.ai/skills/backend/RULES.md` — especially sensitive fields, auth context extraction
- Double-check generated-code headers
- For migrations: plan two-step (nullable → backfill → required)
- Verify secrets aren't being logged or echoed

## Verbosity
- Explain the risk before acting
- State rollback path when making schema/infra changes
