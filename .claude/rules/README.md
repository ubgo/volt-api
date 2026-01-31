# Context-Specific Rules

On-demand rules loaded only when performing specific actions. Keeps token usage minimal.

## Rule Index

### Development Workflows

| File | Read Before |
|------|-------------|
| `new-endpoint.md` | Adding GraphQL query/mutation |
| `new-schema.md` | Adding Ent entity or field |
| `new-service.md` | Adding business logic service |
| `new-cli-command.md` | Adding CLI command |
| `new-cron-job.md` | Adding scheduled job |
| `new-lace-package.md` | Adding shared utility to lace/ |

### Infrastructure

| File | Read Before |
|------|-------------|
| `configuration.md` | Adding/modifying PKL config |
| `nats-handler.md` | Adding NATS message handlers |
| `migrations.md` | Complex database schema changes |
| `docker-dev.md` | Local development setup |

### Code Quality

| File | Read Before |
|------|-------------|
| `error-handling.md` | Implementing error handling |
| `logging.md` | Adding logging |
| `tracing.md` | Adding OpenTelemetry tracing |
| `validation.md` | Input validation |
| `testing.md` | Writing tests |
| `libraries.md` | Adding dependencies |

### Security & Operations

| File | Read Before |
|------|-------------|
| `authentication.md` | Auth-related code |
| `secrets.md` | Handling secrets/API keys |
| `debugging.md` | Investigating issues |
| `git-commit.md` | Committing code |

### Deployment

| File | Read Before |
|------|-------------|
| `deployment.md` | Deploying to VPS |
| `github-secrets.md` | Updating GitHub secrets |

## Deep Documentation

For comprehensive guides (500+ lines), see `.claude/docs/`:

| Topic | Location |
|-------|----------|
| GitHub Actions CI/CD | `.claude/docs/backend/github-actions-deployment.md` |
| Docker Production | `.claude/docs/backend/docker-production.md` |
| Secrets CLI | `.claude/docs/backend/secrets-cli.md` |

Rules provide quick reference; deep docs provide full context when needed.

## Rule Format

Each rule file follows this structure:

```markdown
# [Action] Rules

**Read before:** [trigger description]

## Checklist
- [ ] Step 1
- [ ] Step 2

## Do
- Required practices

## Don't
- Anti-patterns to avoid

## Example
[Real code example]
```

## Adding Rules

1. Create `[action-name].md` in this directory
2. Add entry to table above
3. Add entry to `AGENTS.md` "Context-Specific Rules" table
4. Keep under 100 lines
