# Support

Volt API is a commercial product. Support is available to licensed users.

## Getting help

- **Documentation** — start with the [README](README.md), the setup guide in [CONTRIBUTING.md](CONTRIBUTING.md), and the guides in [`docs/volt/`](docs/) (dev, structure, GraphQL API, deployment, workflow).
- **Health check** — run the API and hit `/api/i/bootinfo` (with the internal key) to confirm the database, Redis, NATS, and OpenTelemetry are wired up.
- **Licensing & account questions** — email **khanakia@gmail.com**.
- **Bugs** — licensed users can report reproducible issues via the repository issue tracker, or by email if the repo is private. For security-sensitive reports, follow [SECURITY.md](SECURITY.md) instead of filing a public issue.

## Before reporting a bug

- Confirm you're on the latest development branch.
- Include your OS, Go version (`go version`), and PostgreSQL / NATS versions.
- Note your `CONFIG_ENV` and whether you're running via `task dev`, `task dev:portless`, or Docker.
- Provide a minimal reproduction and the exact error output (and the relevant `trace_id` from the logs if you have one).

This project is not open to public contributions; it is maintained by the licensor.
