# Secrets Rules

**Read before:** Handling secrets, API keys, or sensitive configuration.

## Checklist

- [ ] Never hardcode secrets in code
- [ ] Bind it as a config field tagged `secret:"true"`
- [ ] Add to `.gitignore` if file-based
- [ ] Use `Sensitive()` for Ent fields

## Do

- Store secrets in environment variables
- Bind every secret as a field in `modules/config/config.go` with `env:` and `secret:"true"` — never `os.Getenv` in app code
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

## Secrets in Config

Declare the field once, in `modules/config/config.go`:

```go
type Core struct {
    // EncryptionKey encrypts data at rest. secret:"true" is not cosmetic:
    // the value is STRUCTURALLY masked — it never reaches `config explain`,
    // the JSON view, or an error message, so that output stays safe to paste
    // into an issue.
    EncryptionKey string `env:"ENCRYPTION_KEY" secret:"true" doc:"data encryption key"`
}
```

Two consequences worth knowing:

- **The shipped placeholder is refused outside dev.** `Validate()` fails the load when `CONFIG_ENV` is `stag` or `prod` and the value is still one of `knownPlaceholderSecrets` — a deploy carrying the default encryption key fails the gate rather than the audit. Dev tolerates it, which is what keeps a laptop with nothing set able to boot.
- **`task config:check` catches it before the rollout.** Run `CONFIG_ENV=prod task config:check` in CI; it reports every missing or placeholder secret at once, without constructing the application.

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
go run ./webapp/cmd/cli crypt -e "sensitive value"

# Decrypt a value
go run ./webapp/cmd/cli crypt -d "encrypted-string"
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

## Pre-commit secret scanner (gitleaks)

Human attention loses to fatigue; a hook doesn't. The repo ships a `.githooks/pre-commit` scanner backed by **[gitleaks](https://github.com/gitleaks/gitleaks)** — activate it once per clone:

```bash
task hooks:install        # sets core.hooksPath=.githooks + installs gitleaks (via brew) if missing
```

`gitleaks` is the primary scanner (with a regex fallback for known token shapes). Two on-demand tasks:

```bash
task secrets:scan         # gitleaks against the FULL repo history (slow, thorough)
task secrets:scan:staged  # gitleaks against the staged diff only (fast)
```

Install gitleaks manually if you don't use Homebrew: [github.com/gitleaks/gitleaks#installing](https://github.com/gitleaks/gitleaks#installing). The hook is defense-in-depth, not the primary defense — keep secrets out of tracked files in the first place. `git commit --no-verify` bypasses it for emergencies only.
