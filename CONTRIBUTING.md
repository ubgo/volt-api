# Contributing to Volt API

Volt API is proprietary, private software (see [LICENSE](LICENSE)). Contributions are limited to authorized collaborators. This guide covers the internal workflow.

## Prerequisites

See the [README](README.md#prerequisites): Go 1.25+, PostgreSQL 15+, NATS with JetStream, [Task](https://taskfile.dev/), and the [PKL CLI](https://pkl-lang.org/). Redis is optional.

## Setup

```bash
git clone git@github.com:khanakia/boilerplate-golang.git
cd boilerplate-golang
task workvendor
cp env.sample .env
task dev
```

## Workflow

1. Branch from the active development branch (e.g. `v2`), not `main`.
2. Make focused changes — one concern per branch. Don't refactor adjacent code that isn't broken.
3. **Never edit generated code** (`gen/`, `generated/`, `ent/`). Edit the source schema and regenerate:
   - Ent schema (`modules/dbent/schema/`) → `task dbent:g`
   - GraphQL schema (`*.graphqls`) → `task gql`
   - PKL config (`modules/config/pkl/`) → `task config:gen`
4. Run `task lint` (gofmt + go vet) and `task test` before opening a PR.
5. Open a PR against the development branch and request review.

## Conventions

- **Commits:** Conventional Commits — `type(scope): description` (e.g. `feat(apidash): add user query`). No AI attribution or co-author lines.
- **Context first, error last:** `func Do(ctx context.Context, ...) (*Result, error)`.
- **Errors are values** — wrap with `fmt.Errorf("op: %w", err)`; never `panic` in libraries.
- **Structured logging** via `gozap.FromCtx(ctx)`; never `fmt.Println` in production code.
- **Config** via PKL bindings, never raw `os.Getenv` in app code.
- **Secrets** never committed — no `.env`, `*.creds`, `*.pem`, `*.key`, or API keys in git.

More detail lives in [`AGENTS.md`](AGENTS.md) and `.claude/rules/`.

## Reporting issues

Use the private issue tracker. For security-sensitive reports, follow [SECURITY.md](SECURITY.md) instead of filing a public issue.
