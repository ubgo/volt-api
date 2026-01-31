---
description: Enforce usage of 'lace' shared libraries over standard lib
globs: **/*.go
---

# Lace Library Usage

This project uses the `lace` module for standard utilities. Prefer these over standard library equivalents:

1.  **Logging:**
    *   ❌ `log.Println`, `fmt.Println`
    *   ✅ `import "lace/nlog"` -> `nlog.Info(...)`, `nlog.Error(...)`

2.  **Error Handling:**
    *   ❌ `errors.New(...)`
    *   ✅ `import "lace/cerr"` -> `cerr.New(...)`, `cerr.NewFromCtx(ctx)`

3.  **JSON Types:**
    *   ✅ `import "lace/jsontype"` -> `jsontype.JSON` (for DB fields)

4.  **Context:**
    *   Always pass `context.Context` as the first argument to functions performing I/O.
