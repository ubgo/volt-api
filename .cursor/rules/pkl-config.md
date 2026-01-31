***

description: Rules for editing PKL configuration files
globs: config/\*\*/\*.pkl
-------------------------

# PKL Configuration Rules

When you modify `.pkl` files:

1. **Type Safety:** Respect the types defined in `RootConfig.pkl`.
2. **Mandatory Command:** After editing, you **MUST** regenerate the Go bindings:
   ```bash
   task config
   ```
3. **Environment Sync:** If you add a new key in `RootConfig.pkl`, make sure to provide a default value in `default/app.pkl`.
