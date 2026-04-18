# Project: {{PROJECT_NAME}}

<!-- GENERATED FILE — do not edit directly. Run /codeskill sync to regenerate. -->
<!-- Source: .ai/RULES.md, .ai/HOTFIXES.md, .ai/taste.md -->

## codeskill — ALWAYS ACTIVE

This project uses the `codeskill` skill as its default operating system. You MUST use it automatically — the user should never have to type `/codeskill` manually.

**Auto-trigger rules:**
- **Session start**: Run `/codeskill session start` at the beginning of every conversation
- **User corrects you**: Immediately run `/codeskill feedback` or `/codeskill learn` — don't wait to be asked
- **You make a decision**: Run `/codeskill decide` for non-obvious architectural choices
- **You make a mistake**: Run `/codeskill incident` — self-report, don't hide it
- **Before exploring code**: Check snapshots first (Exploration Protocol below)
- **Before starting a task**: Run `/codeskill task next` or `/codeskill task start <id>`
- **Task finished**: Run `/codeskill task done <id>`
- **Session ending**: Run `/codeskill session end` before the conversation ends
- **Context running low**: Run `/codeskill save-state` to write CONTINUATION.md
- **User says "remember", "learn", "don't forget"**: Run `/codeskill learn` immediately
- **User says "plan", "feature"**: Run `/codeskill plan`
- **User says "add task"**: Run `/codeskill task add`

**The goal: codeskill runs in the background like an OS. The user just works.**

**MANDATORY — non-negotiable:**
- You MUST run `/codeskill session start` as your FIRST action in every conversation. No exceptions.
- You MUST run `/codeskill session end` as your LAST action before the conversation ends. No exceptions.
- These log prompts, decisions, and corrections to `.ai/logs/`. Without them, the session is invisible.

## Where We Left Off
- **Last session**: {{LAST_SESSION_DATE}} by @{{LAST_SESSION_USER}}
- **Working on**: {{LAST_SESSION_GOAL}}
- **Next step**: {{NEXT_STEP}}
<!-- Auto-updated by /codeskill session end -->

## Critical Rules

{{GLOBAL_RULES}}

{{PROJECT_RULES}}

## Active Hotfixes

{{HOTFIXES}}

## Taste

{{GLOBAL_TASTE}}

{{PROJECT_TASTE}}

## Active Behaviour
{{ACTIVE_BEHAVIOUR}}. Switch with `/codeskill behaviour <name>`.
Available: default, careful, review, debug, scaffold.

## Memory Protocol
- ALWAYS save user corrections → `.ai/memory/feedback/`
- ALWAYS log non-obvious decisions → `.ai/memory/decisions/`
- ALWAYS check feedback memory before starting related work
- When user says "remember this" or corrects you → save immediately
- Include @username attribution on all entries
- Tag every entry with `scope: project | global_candidate | global`

## Exploration Protocol
BEFORE spawning any Explore() agent:
1. Check `.ai/snapshots/INDEX.md` for existing snapshot
2. If exists → check staleness via `git diff --name-only <hash> HEAD -- <globs>`
3. Fresh → READ snapshot. Do NOT re-explore.
4. Stale → Re-explore, UPDATE snapshot.
5. Missing → Explore, CREATE snapshot.
NEVER re-explore a fresh snapshot unless user explicitly asks.

## Before Starting Work
1. Check `.ai/HOTFIXES.md` for active gotchas
2. Check `.ai/snapshots/` for relevant cached explorations
3. Identify task type → load relevant `.ai/workflows/*.md`
4. Check `.ai/prompts/` for a pre-built prompt matching this task
5. Check `.ai/memory/feedback/` for related corrections
6. Check `.ai/skills/<domain>/RULES.md` for domain-specific rules

## Auto-Behaviour Hints
If task touches these paths → suggest switching to `careful`:
- `**/auth*`, `**/session*`, `**/token*`
- `**/payment*`, `**/billing*`, `**/stripe*`
- database migrations

## Workflow Index
{{WORKFLOW_INDEX}}

## After Session
ALWAYS run before ending:
1. Save any corrections received: `/codeskill feedback`
2. End session: `/codeskill session end`
If mistakes happened:
3. Create incident: `/codeskill incident`
If pattern is recurring (2+ times):
4. Promote to hotfix: `/codeskill hotfix add`

## Logging — Who Did What
All logs include user attribution (email + GitHub username from git config).
Stored in: `.ai/logs/sessions/`, `.ai/logs/activity/`, `.ai/logs/incidents/`

## Reference
- `.ai/ARCHITECTURE.md` — system overview
- `.ai/PATTERNS.md` — code conventions
- `.ai/STACK.md` — tech stack
- `.ai/snippets/` — reusable code patterns
- `.ai/prompts/` — reusable prompt templates
- `.ai/guides/` — deep reference docs
