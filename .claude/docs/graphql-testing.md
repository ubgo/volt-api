# GraphQL Testing Quick Reference

## Endpoints

| API | URL | Purpose |
|-----|-----|---------|
| apidash | `http://localhost:2315/sa/query` | Admin/Service API |
| gql | `http://localhost:2315/query` | Public API |

Playground: `http://localhost:2315/sa?key=1234`

## Prerequisites

See [CLI Reference](cli/README.md) for:

* `seedusers` - Create test users
* `logintoken` - Generate JWT token for testing

## Quick Test

```bash
# 1. Check server is running
curl -s http://localhost:2315/sa/query \
  -H "Content-Type: application/json" \
  -d '{"query": "{ ping }"}' | jq

# 2. Test authenticated query (get token from CLI first)
curl -s http://localhost:2315/sa/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"query": "{ me { id firstName lastName email status } }"}' | jq
```

## Database Access

```bash
# Connect
PGPASSWORD=root psql -h localhost -U postgres -d wildai_githubapp_db

# Quick queries
PGPASSWORD=root psql -h localhost -U postgres -d wildai_githubapp_db -c "SELECT id, email FROM users;"
```

**Connection:** `localhost:5432`, user `postgres`, password `root`, db `wildai_githubapp_db`

## Authentication Methods

The API accepts tokens via (in order of precedence):

1. Query param: `?token=xxx`
2. Header: `token: xxx`
3. Header: `Authorization: Bearer xxx`

***

## Available Queries

### Public (no auth)

```graphql
{ ping }
```

### Authenticated

```graphql
{
  me {
    id
    firstName
    lastName
    email
    status
  }
}
```

***

## Troubleshooting

### "unauthorized" error

* Token missing or invalid
* User doesn't exist or `status=false`
* Token expired (36 hour expiry)

### Server not responding

```bash
# Check if running
lsof -i :2315

# Start server
task dev  # or your usual start command
```

### No users in database

Run `seedusers` - see [CLI Reference](cli/README.md)

***

## Key Files

| File | Purpose |
|------|---------|
| `apidash/internal/graph/schemas/schema.graphql` | Schema definitions |
| `apidash/internal/graph/resolver/schema.resolvers.go` | Query implementations |
| `saas/pkg/middleware/middleware.go` | Auth middleware |
| `saas/pkg/auth/common.go` | JWT create/validate |
| `saas/cmd/cli/main.go` | CLI commands (seedusers, logintoken) |
| `config/pkl/default/app.pkl` | Database config |
