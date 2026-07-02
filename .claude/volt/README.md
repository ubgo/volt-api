# Volt API — Claude / AI agent resources

**These are [Volt API](../../README.md)'s agent resources** — the context-specific rules, deep reference docs, and code templates that ship with the Go backend boilerplate. They live under `.claude/volt/` on purpose: it keeps Volt's AI guidance namespaced and separable from any rules you add for your own app. Your own project-level Claude config (`.claude/settings*.json`, your own commands/skills/agents) stays at the `.claude/` root; only Volt's bundled `rules/`, `docs/`, and `templates/` are grouped here.

Referenced from [AGENTS.md](../../AGENTS.md) and [CLAUDE.md](../../CLAUDE.md).

## Contents

| Folder | What |
|--------|------|
| [rules/](rules/) | On-demand, context-specific rules — read the matching file before an action (new endpoint, schema, migration, NATS handler, logging, secrets, deployment, …). Start at [rules/README.md](rules/README.md). |
| [docs/](docs/) | Deeper reference guides (backend packages, CLI, deployment, cross-service OTel tracing). |
| [templates/](templates/) | Code templates (service, handler, schema) used when scaffolding. |

> These are a reference set for building on Volt API. Trim or extend them for your own app — but keep `rules/logging.md` and friends accurate if you change the underlying stack.
