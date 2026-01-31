# Repository Guidelines for AI Agents & Developers

**CRITICAL CONTEXT:**
This project relies heavily on **Code Generation**.

1. **NEVER** modify files inside `gen/`, `generated/`, or `ent/` directories directly.
2. **ALWAYS** edit the **Source** (Schema, GraphQL definitions, PKL) and run the appropriate generation task.

***

## Documentation Index

For detailed instructions, rules, and navigation, please refer to the specialized guides below.

### Core Intelligence & Rules (.ai/)

These documents define the "Physics" of the project. AI Agents must strictly adhere to these.

* **[Agent Protocol (AGENTS.md)](.ai/AGENTS.md)**: **START HERE.** Rules for coding, testing, and navigation.
* **[Architecture (ARCHITECTURE.md)](.ai/ARCHITECTURE.md)**: High-level system design and patterns.

### Claude Code Project Docs (.claude/docs/)

Project-specific documentation maintained by Claude Code sessions.

* **[Overview](.claude/docs/README.md)**: Index and quick start.
* **[cli/](.claude/docs/cli/)**: CLI commands reference.
* **[auth/](.claude/docs/auth/)**: Authentication & JWT patterns.
* **[database/](.claude/docs/database/)**: Ent ORM patterns & entity docs.
* **[api/](.claude/docs/api/)**: GraphQL API patterns.

### Developer Guides (docs/)

Practical "How-to" guides for humans and agents executing specific tasks.

* **[Project Structure (STRUCTURE.md)](docs/STRUCTURE.md)**: Detailed breakdown of folders and modules.
* **[Dev Workflow (WORKFLOW\_GUIDE.md)](docs/WORKFLOW_GUIDE.md)**: How to develop features, branch strategies, and PRs.
* **[API Guide (API\_GQL\_GUIDE.md)](docs/API_GQL_GUIDE.md)**: Working with GraphQL, Resolvers, and Authentication.
* **[Deployment (DEPLOY\_GUIDE.md)](docs/DEPLOY_GUIDE.md)**: Deploying to production.
* **[Docker & Infra (DOCKER\_GUIDE.md)](docs/DOCKER_GUIDE.md)**: Managing local infrastructure (Postgres, NATS).
* **[General Dev Guide (DEV\_GUIDE.md)](docs/DEV_GUIDE.md)**: Setup, troubleshooting, and tips.

***

## Key Configuration Files

Analyzing these files provides immediate context on dependencies and workflows:

* **`go.work`**: Defines the active modules in the workspace. **Source of Truth for module resolution.**
* **`Taskfile.yml`**: The registry of all runnable commands. Check this before proposing shell scripts.
* **`dbent/schema/`**: The Database Schema definitions (Ent schemas).
* **`saas/cmd/cli/`**: The location for CLI tools and scrapers.
* **`apidash/internal/graph/*.graphqls`**: The API Contract definitions.
* **`config/RootConfig.pkl`**: The Application Configuration schema.

***

## Quick Task Reference

When instructing the user, prefer these `task` commands over raw `go` commands:

* `task workvendor` -> Sync workspace dependencies.
* `task config` -> Regenerate Pkl config bindings.
* `task dbent:entg` -> Regenerate Ent ORM code.
* `task dbent:migrate` -> Run database migrations after schema changes.
* `task gql` -> Regenerate GraphQL resolvers.
* `task gitgen` -> Stage all generated files for commit.
* `task dev` -> Start local Docker infrastructure.
* `task lint` -> Run linters.
* `task test` -> Run tests.

See [CLI Reference](.claude/docs/cli/) for full documentation.

## Git Commit Rules

* **NEVER** include "Generated with Claude Code" or similar AI attribution in commit messages
* **NEVER** include "Co-Authored-By: Claude" or any AI co-author lines
* Keep commit messages clean and focused on the changes only

## Summary

The structure separates concerns based on the 'reader' (Humans vs. AI vs. Editor):

* `**docs/**` These are explanatory guides for **human developers** (setup, workflows, API usage).

* `**AGENT.md**` **&** `**.ai/**` This is specifically for the AI Agent. `AGENT.md` acts as the entry point (system prompt) standardizing the context initialization. The `.ai` folder contains imperative rules, architectural constraints, and code examples optimized for LLM token efficiency, not human readability.

* `**.cursor/rules**` These aren't documentation but **active rules** for the Cursor editor. They inject specific context automatically before edits. If we aren't using Cursor, this folder is safe to remove.
