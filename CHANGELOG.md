# Changelog

All notable changes to Volt API are documented here. The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Volt brand kit under `brand/` (icons, favicon, wordmark, lockups, OG covers, design tokens) driven by `brand.json` + `gen-brand.mjs`.
- `LICENSE` (Volt API Commercial License), `CONTRIBUTING.md`, `SECURITY.md`, `SUPPORT.md`, `.editorconfig`, and `.gitattributes`.
- `VOLT.md` — the preserved copy of Volt API's original README, so the "built on Volt API" reference survives after a buyer replaces `README.md` with their own app's.

### Changed
- Rebranded the project from "Go Boilerplate" to **Volt API**; restructured `README.md` to mirror the Volt (Web) landing page — centered logo header, "Why", "What's inside", features, short getting-started, and FAQ — with all setup moved into `CONTRIBUTING.md`.
- Rewrote `CONTRIBUTING.md` as the full development guide: prerequisites, install, the portless HTTPS dev workflow (`task dev:portless` → `https://api.app.localhost`), PKL configuration, code generation, migrations, scripts, project structure, API endpoints, deployment, and further docs.
- Corrected task references (`dbent:entg` → `dbent:g`, `task config` → `config:gen`) and the `modules/` monorepo layout across `README.md`, `AGENTS.md`, `CLAUDE.md`, `.claude/rules/`, docs, and templates.
- License changed from MIT (README-only, no file) to the proprietary Volt API Commercial License.
