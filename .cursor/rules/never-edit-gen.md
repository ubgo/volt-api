***

description: Critical rule to prevent manual editing of generated files
globs: **/gen/**, **/generated/**, \*\*/models\_gen.go
------------------------------------------------------

# NEVER EDIT GENERATED CODE

You are prohibited from manually editing files in these directories:

* `dbent/gen/ent/`
* `apidash/internal/graph/generated/`
* `config/gen/`

**Why?** Any changes will be overwritten during the next `task` execution.

**Action:** If a change is needed in these files, identify the **Source File** (Schema or Pkl) and edit that instead, then run the corresponding `task`.
