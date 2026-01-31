---
description: Rules for creating and modifying CLI commands
globs: saas/cmd/cli/**/*.go
---

# CLI Command Rules

When the user asks to create or modify a CLI command:

1.  **Location:** All CLI registration code **MUST** reside in `saas/cmd/cli/main.go` (or additional files in the `main` package within `saas/cmd/cli/`).
2.  **Framework:** Use [Cobra](https://github.com/spf13/cobra).
3.  **Pattern:**
    *   Logic belongs in `saas/pkg/...` (Business Logic Layer).
    *   CLI commands are just a **Delivery Layer** wrapper. They should parse flags and call services.
4.  **Registration:**
    *   Define a function like `newMyCommand(plugin app.Plugin) *cobra.Command`.
    *   Register it in the `main()` function using `root.AddCommand(newMyCommand(plugin))`.
5.  **Boilerplate Context:** Use the `plugin` object (of type `app.Plugin`) to access the database (`plugin.Ent.Client()`), configuration, and other shared resources.