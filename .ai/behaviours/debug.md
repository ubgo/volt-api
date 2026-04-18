---
name: debug
description: Root-cause investigation — no fixes until cause known
extends: default
---

# Debug Mode

## Overrides
- No speculative fixes. Find the root cause first.
- Read logs, traces, and code paths before proposing anything
- State hypotheses explicitly; verify each before acting

## Emphasis
- Check `/api/i/bootinfo` for service health
- Inspect OTEL traces via `trace_id` from logs
- Database driver logs include query origin (`file:line`)
- Reproduce locally before fixing — `task docker:up` then attach CLI

## Verbosity
- Hypothesis → evidence → conclusion format
- Show the smoking-gun log line or trace span
