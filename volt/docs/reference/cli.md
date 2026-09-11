# CLI Commands Reference

Quick reference for all CLI commands. Copy and run these directly.

---

## Running CLI Commands

```bash
# Build and run
go run ./webapp/cmd/cli <command> [args]

# Or build first, then run
go build -o cli ./webapp/cmd/cli
./cli <command> [args]
```

---

## Available Commands

### ping
Test that CLI is working.

```bash
go run ./webapp/cmd/cli ping
```

**Output:** `pong`

---

### seedusers
Create demo users, workspaces, and user-workspace assignments for testing.

```bash
go run ./webapp/cmd/cli seedusers
```

**Output:**
```
=== Seeding Demo Data ===

--- Users ---
  [created] John Doe <john.doe@example.com> - udemo_john_doe_00001
  [created] Jane Smith <jane.smith@example.com> - udemo_jane_smith_01
  [created] Bob Johnson <bob.johnson@example.com> - udemo_bob_johnson_1
  [created] Alice Williams <alice.williams@example.com> - udemo_alice_willia

--- Workspaces ---
  [created] Personal Workspace - ws_demo_personal_01
  [created] Acme Corp - ws_demo_acme_corp_1
  [created] Startup Inc - ws_demo_startup_001

--- Assignments ---
  [created] udemo_john_doe_00001 -> ws_demo_personal_01 (owner)
  ...

=== Done ===
```

**Demo Users (Hardcoded IDs):**
| ID | Name | Email |
|----|------|-------|
| `udemo_john_doe_00001` | John Doe | john.doe@example.com |
| `udemo_jane_smith_01` | Jane Smith | jane.smith@example.com |
| `udemo_bob_johnson_1` | Bob Johnson | bob.johnson@example.com |
| `udemo_alice_willia` | Alice Williams | alice.williams@example.com |

**Demo Workspaces (Hardcoded IDs):**
| ID | Name |
|----|------|
| `ws_demo_personal_01` | Personal Workspace |
| `ws_demo_acme_corp_1` | Acme Corp |
| `ws_demo_startup_001` | Startup Inc |

**User-Workspace Assignments:**
| User | Workspace | Role |
|------|-----------|------|
| John | Personal | owner |
| John | Acme | member |
| Jane | Acme | owner |
| Jane | Personal | member |
| Bob | Acme | admin |
| Alice | Startup | owner |

---

### logintoken
Generate a JWT login token for a user within a specific workspace.

```bash
go run ./webapp/cmd/cli logintoken <user_id> <workspace_id>
```

---

### testgithub
Test GitHub App configuration and API connectivity.

```bash
# Test config only
go run ./webapp/cmd/cli testgithub

# Test config + fetch installation details
go run ./webapp/cmd/cli testgithub <installation_id>
```

**Example:**
```bash
go run ./webapp/cmd/cli testgithub 101705340
```

**Output:**
```
=== GitHub Service Test ===

--- Config ---
  AppId:          2553636
  AppName:        thewildai
  ClientId:       Iv23lizyD1iX1vN1q6LE
  PrivateKeyPath: /path/to/_keys/githubapp-thewildai.2025-12-28.private-key.pem
  CallbackUrl:    http://localhost:5173/github/callback

--- Private Key Check ---
  [OK] Private key file exists: /path/to/_keys/...

--- Service Init ---
  [OK] GitHub service initialized successfully

--- Fetch Installation ---
  Fetching installation 101705340...
  [OK] Installation found:
       ID:      101705340
       Account: username (User)
       Avatar:  https://avatars.githubusercontent.com/u/...

=== Test Complete ===
```

**Use cases:**
- Verify private key path is correct
- Test GitHub API connectivity
- Debug installation issues

---

### syncwildai
Sync pending files to WildAI service.

```bash
# Sync all pending files
go run ./webapp/cmd/cli syncwildai

# Sync specific item group
go run ./webapp/cmd/cli syncwildai <item_group_id>
```

**Output:**
```
=== WildAI Sync ===

Syncing all pending files...

--- Results ---
  Item Groups Processed: 2
  Files Processed:       5
  Files Success:         4
  Files Error:           1

--- Errors ---
  - File igf_xxx: failed to get file content...

=== Done ===
```

---

### cleandeleted
Permanently delete soft-deleted item group files.

```bash
go run ./webapp/cmd/cli cleandeleted
```

**What it does:**
1. Finds all files with `status="deleted"`
2. Files **without** `wildai_source_id`: Deleted directly from database
3. Files **with** `wildai_source_id`: Deleted from WildAI API first, then from database

**Output:**
```
=== Clean Deleted Files ===

Cleaning up deleted files...

--- Results ---
  Files Processed:        10
  Files Deleted:          8
  - Deleted Direct:       5 (no wildai_source_id)
  - Deleted from API:     3 (deleted from WildAI first)
  Files Error:            2

--- Errors ---
  - File igf_xxx: failed to delete from WildAI API: ...

=== Done ===
```

**Notes:**
- Files that fail API deletion are skipped and will retry on next run
- This runs automatically every 10 minutes via cron

---

### logintoken (continued)

**Quick Test (copy-paste ready):**
```bash
# John as owner of Personal workspace
go run ./webapp/cmd/cli logintoken udemo_john_doe_00001 ws_demo_personal_01

# Jane as owner of Acme Corp
go run ./webapp/cmd/cli logintoken udemo_jane_smith_01 ws_demo_acme_corp_1

# Bob as admin of Acme Corp
go run ./webapp/cmd/cli logintoken udemo_bob_johnson_1 ws_demo_acme_corp_1

# Alice as owner of Startup Inc
go run ./webapp/cmd/cli logintoken udemo_alice_willia ws_demo_startup_001
```

**Output:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Token Claims:**
```json
{
  "sub": "udemo_john_doe_00001",
  "email": "john.doe@example.com",
  "name": "John Doe",
  "workspace_id": "ws_demo_personal_01",
  "exp": 1234567890
}
```

**Notes:**
- Token expires in 30 days
- Token is scoped to a single workspace
- Use in API requests: `Authorization: Bearer <token>`

**Web App Login:**
```bash
# Generate token and open in browser
TOKEN=$(go run ./webapp/cmd/cli logintoken udemo_john_doe_00001 ws_demo_personal_01)
open "http://localhost:5173/login?token=$TOKEN"
```

Or construct URL manually: `http://localhost:5173/login?token=<jwt>`

---

## Taskfile Commands

These are run via `task` (Taskfile.yml):

```bash
# Development
task dev              # Start Docker (Postgres, NATS)
task lint             # Run linters
task test             # Run tests

# Code Generation
task dbent:g        # Generate Ent ORM code
task config:check     # validate the configuration (non-zero on any problem)
task workvendor       # Sync go.work dependencies
task gitgen           # Stage generated files for commit

# Database
task migrate          # Run database migrations

# GraphQL (run from gql/ directory)
cd gql && task gql    # Generate GraphQL code
```

---

## Adding New CLI Commands

1. Edit `modules/webapp/cmd/cli/main.go`
2. Add the command in `AddCommands()`. It receives the root command and a lazy accessor — call the accessor inside `Run` only if the command actually needs the kernel, so commands that don't stay runnable without a database:

```go
var myCommand = &cobra.Command{
    Use:   "mycommand [args]",
    Short: "Brief description",
    Long:  "Detailed description for help text",
    Run: func(cmd *cobra.Command, args []string) {
        entClient := plugin.EntDB.Client()
        ctx := context.Background()

        // Your logic here

        fmt.Println("Done")
    },
}

rootCmd.AddCommand(myCommand)
```

3. Update this file with the new command documentation

---

## Common Patterns

### Query a user
```go
user, err := entClient.User.Query().Where(user.ID(args[0])).First(ctx)
if err != nil {
    fmt.Println(err)
    return
}
```

### Create a user
```go
user, err := entClient.User.Create().
    SetFirstName("Name").
    SetEmail("email@example.com").
    SetStatus(true).
    Save(ctx)
```

### Check if user exists
```go
exists, _ := entClient.User.Query().Where(user.Email(email)).Exist(ctx)
if exists {
    fmt.Println("User already exists")
    return
}
```

---

## Troubleshooting

### "NATS CONNECTION" error on startup
This is normal if NATS credentials aren't configured. The CLI will still work for database operations.

### User not found
Make sure:
1. Database is running: `task dev`
2. Migrations ran: `task migrate`
3. User exists: run `seedusers` first

### Build errors after schema changes
```bash
task dbent:g   # Regenerate Ent code
task migrate     # Update database
```
