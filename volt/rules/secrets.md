# Secrets Rules

**Read before:** Handling secrets, API keys, or sensitive configuration.

## Checklist

- [ ] Never hardcode secrets in code
- [ ] Use PKL config with environment variables
- [ ] Add to `.gitignore` if file-based
- [ ] Use `Sensitive()` for Ent fields

## Do

- Store secrets in environment variables
- Use PKL `read("env:VAR_NAME")` for config
- Mark Ent fields as `Sensitive()` to hide from logs
- Encrypt sensitive data at rest using `lace/crypt`

## Don't

- Commit secrets to git (even in "test" files)
- Log sensitive values
- Pass secrets in URLs/query strings
- Store plain-text passwords

## Secret Storage Hierarchy

| Environment | Method |
|-------------|--------|
| Local dev | `.env` file (gitignored) |
| Production | Environment variables / Secret manager |
| CI/CD | GitHub Secrets / CI variables |

## PKL Config with Secrets

```pkl
// config/pkl/env/default/app.pkl
core {
    // Read from environment variable
    encryptionKey = read("env:ENCRYPTION_KEY")
    internalApiKey = read("env:INTERNAL_API_KEY")
}

database {
    password = read("env:DB_PASSWORD")
}
```

## .env File (Local Development)

```bash
# .env (gitignored)
ENCRYPTION_KEY=your-32-byte-encryption-key-here
INTERNAL_API_KEY=your-internal-api-key
DB_PASSWORD=localpassword
NATS_CREDS_FILE=/path/to/nats.creds
```

## .gitignore Entries

```gitignore
# Already in .gitignore - verify these exist
.env
.env.local
.env.*.local
*.creds
*.pem
*.key
```

## Ent Schema Sensitive Fields

```go
// dbent/schema/user.go
func (User) Fields() []ent.Field {
    return []ent.Field{
        field.String("password").Sensitive(),  // Won't appear in logs/debug
        field.String("api_token").Sensitive(),
        field.String("ssn").Sensitive(),
    }
}
```

## Encrypting Data at Rest

```go
import "lace/crypt"

// Encrypt before storing
encrypted, err := crypt.Encrypt(sensitiveData)
if err != nil {
    return fmt.Errorf("encrypt: %w", err)
}
// Store encrypted value

// Decrypt when reading
decrypted, err := crypt.Decrypt(encrypted)
if err != nil {
    return fmt.Errorf("decrypt: %w", err)
}
```

## CLI for Encryption

```bash
# Encrypt a value
go run ./saas/cmd/cli crypt -e "sensitive value"

# Decrypt a value
go run ./saas/cmd/cli crypt -d "encrypted-string"
```

## Accessing Secrets in Code

```go
// Via config (preferred)
apiKey := plugin.RootConfig.Core.InternalApiKey

// Never do this
apiKey := "hardcoded-key"  // BAD!
```

## Logging Safety

```go
// Bad - logs sensitive data
applog.GetLogger().Info("user created", ulog.String("password", user.Password))

// Good - omit sensitive fields
applog.GetLogger().Info("user created", ulog.String("id", user.ID))

// If must log, mask it
applog.GetLogger().Info("api call", ulog.String("api_key", maskSecret(apiKey)))

func maskSecret(s string) string {
    if len(s) <= 4 {
        return "****"
    }
    return s[:2] + "****" + s[len(s)-2:]
}
```

## API Key Validation

```go
// Constant-time comparison to prevent timing attacks
import "crypto/subtle"

func validateAPIKey(provided, expected string) bool {
    return subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) == 1
}
```

## Production Checklist

- [ ] All secrets in environment variables
- [ ] No secrets in git history
- [ ] Database passwords rotated
- [ ] API keys have minimal required permissions
- [ ] Encryption keys backed up securely
- [ ] Audit log for secret access (if required)

## Rotating Secrets

1. Generate new secret
2. Update environment variables
3. Deploy with new secret
4. Revoke old secret
5. Verify application still works

## If Secret is Leaked

1. **Immediately** revoke/rotate the secret
2. Check git history: `git log -p --all -S 'secret-value'`
3. Use `git filter-branch` or BFG to remove from history
4. Force push (coordinate with team)
5. Audit access logs for misuse
