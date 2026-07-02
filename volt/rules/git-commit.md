# Git Commit Rules

**Read this before making any git commit.**

## Required

- Keep commit messages clean and focused on the changes only
- Use conventional commit format: `type(scope): description`

## Forbidden

- NEVER include "Generated with Claude Code" or similar AI attribution
- NEVER include "Co-Authored-By: Claude" or any AI co-author lines
- NEVER include any AI-related metadata or comments

## Pre-commit secret scanner

Commits run through a `.githooks/pre-commit` hook backed by **[gitleaks](https://github.com/gitleaks/gitleaks)** that refuses any commit introducing a credential-shaped value. Activate it once per clone with `task hooks:install` (also installs gitleaks via brew if missing). On-demand scans: `task secrets:scan` (full history) / `task secrets:scan:staged` (staged diff). Don't rely on the hook alone — keep secrets in `.env` only (see [secrets.md](secrets.md)). `--no-verify` bypasses it for emergencies only.

## Conventional Commit Types

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `refactor` | Code refactoring (no feature change) |
| `docs` | Documentation only |
| `chore` | Maintenance tasks, dependencies |
| `test` | Adding or updating tests |
| `style` | Formatting, whitespace (no code change) |

## Examples

```bash
# Good
feat(cronfn): add CleanDeletedFiles job function
fix(auth): handle expired JWT tokens correctly
refactor(wildaisync): extract file processing logic
docs(agents): add doc.go convention

# Bad - DO NOT DO THIS
feat(cronfn): add CleanDeletedFiles job function

Co-Authored-By: Claude <noreply@anthropic.com>

# Bad - DO NOT DO THIS
fix(auth): handle expired JWT tokens correctly
# Generated with Claude Code
```
