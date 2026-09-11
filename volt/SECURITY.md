# Security Policy

Volt API is proprietary software. We take security seriously and appreciate responsible disclosure.

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Report privately to **khanakia@gmail.com**. Include:

- A description of the vulnerability and its impact.
- Steps to reproduce (proof-of-concept if possible).
- Affected version, branch, or commit.

You can expect an acknowledgment within a few business days. Please give us a reasonable window to investigate and release a fix before any public discussion.

## Scope

This policy covers the source code in this repository. It does not cover third-party dependencies (report those upstream) or deployments you operate yourself.

## Handling Secrets

Never commit secrets. This repo excludes `.env`, `*.creds`, `*.pem`, `*.key`, and API keys from version control. Configuration secrets are declared as fields in `modules/config/config.go` tagged `secret:"true"`, which masks them structurally — they never reach `config explain`, the JSON view, or an error message — and, outside dev, the shipped placeholder values are refused by `config check`; sensitive Ent fields use `.Sensitive()`. If a secret is ever leaked, rotate it immediately and scrub it from history.
