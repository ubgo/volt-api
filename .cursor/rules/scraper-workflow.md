---
description: Guidance for building scrapers and CLI commands
globs: saas/cmd/cli/**/*.go, saas/pkg/scraper/**/*.go
---

# Scraper & CLI Rules

When building a new scraper or CLI command:

1.  **Location:** Do NOT put business logic inside `cmd/cli`.
    *   Create a dedicated package: `saas/pkg/scraper` (or similar).
    *   `cmd/cli` should only parse flags and call the package.

2.  **Dependencies:** Inject dependencies (like `ent.Client` or `lace.Logger`) into your scraper struct. Do not use global variables.

3.  **Data Persistence:** If saving data, ensure an Ent schema exists in `saas/schema/`.

4.  **Example Structure:**
    ```go
    // saas/pkg/scraper/myscraper.go
    type Scraper struct {
        Client *ent.Client
        Log    nlog.Logger
    }
    ```
