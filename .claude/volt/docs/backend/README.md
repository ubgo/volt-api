# Backend Package Documentation

Documentation for Go packages in `saas/pkg/`.

## Purpose

Each package documentation provides:
- Package purpose and overview
- Key functions/types and their usage
- Code patterns and examples
- Dependencies and related files

## Documented Packages

| Package | Doc File | Path |
|---------|----------|------|
| cronfn | [cronfn.md](./cronfn.md) | `saas/pkg/cronfn/` |

## Adding Documentation

When creating/modifying a package:

1. Create `[package-name].md` in this directory
2. Include:
   - Package path and purpose
   - Public functions/types
   - Usage examples
   - Dependencies
   - Related files
3. Update this README with the new package

## Package Categories

### Job/Task Packages
- `cronfn` - Reusable cron job functions

### Service Packages
- `github` - GitHub API service (not yet documented)
- `wildaisync` - WildAI sync logic (not yet documented)

### Core Packages
- `app` - Application plugin/config (not yet documented)
- `auth` - Authentication/JWT (not yet documented)
