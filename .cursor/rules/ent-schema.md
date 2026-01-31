---
description: Rules for editing Ent database schemas in saas/schema/
globs: saas/schema/*.go
---

# Ent Schema Rules

When you modify any file in `saas/schema/`:

1.  **Strict Typing:** Follow Ent's schema-first patterns.
2.  **Naming:** Use singular PascalCase for schema names.
3.  **Mandatory Command:** After editing, you **MUST** run the generation task:
    ```bash
    task saas:entg
    ```
4.  **Migration:** If fields were added/changed, suggest running:
    ```bash
    task saas:migrate
    ```
5.  **No Direct Gen Edits:** Never suggest editing files in `dbent/gen/ent/`.
