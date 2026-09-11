# Configuration Rules

**Read before:** Adding or modifying application configuration.

Configuration is **Go**, in one file: `modules/config/config.go`. That file is the schema, the defaults and the validation together. Nothing is generated, so nothing can drift from what the code declares — and there is no configuration language, no codegen step, and no config file in the production image.

## Checklist

- [ ] Add the field to the right struct in `modules/config/config.go`
- [ ] Give it `env:`, a `default:` if one is sensible, `secret:"true"` if it is a credential, and a `doc:` line
- [ ] Extend `Derive()` if its default depends on another field
- [ ] Extend `Validate()` if it has a constraint
- [ ] `task config:document` to regenerate `.env.example`
- [ ] Add the key to `readsEnv` in `config_test.go`
- [ ] `task config:check` — and `CONFIG_ENV=prod task config:check` if it affects production

## File Locations

| Type | Location |
|------|----------|
| The whole schema | `modules/config/config.go` |
| Loader + source order | `modules/config/main.go` |
| Tests pinning defaults and rules | `modules/config/config_test.go` |
| The generated contract | `.env.example` (repo root, generated — don't hand-edit) |

## Do

- Give every field a compiled-in `default:` where one is sensible — **zero input must stay a valid configuration**, which is what lets a newcomer clone and run and a test construct a config without fixtures
- Tag credentials `secret:"true"` so they are structurally masked (absent from `explain`, from JSON, and from error messages)
- Write the `doc:` text for the operator who will read `.env.example`, not for the developer who wrote the field
- Put constraints in `Validate()`, so a bad value fails a CI step instead of a production boot
- Read the config from `kernel.RootConfig` inside the app

## Don't

- Hand-edit `.env.example` — it is generated
- Call `os.Getenv` in application code — bind a field instead, so the value appears in `explain`, in `.env.example`, and in validation
- Call `config.GetOrDie()` a second time in a process that already has a kernel — `kernel.RootConfig` is the loaded configuration; a second call builds a second, independent snapshot
- Commit a real secret. Placeholders are what ship, and production refuses them

## Adding a Field

```go
type App struct {
    // ExistingField …

    // NewFeatureTimeout bounds the outbound call so a hung upstream cannot
    // pin a request handler open. Why 30s: it is above the p99 of the
    // upstream and below the load balancer's own idle timeout.
    NewFeatureTimeout time.Duration `env:"NEW_FEATURE_TIMEOUT" default:"30s" doc:"upstream call timeout"`
    NewFeatureKey     string        `env:"NEW_FEATURE_KEY" secret:"true" doc:"upstream API key"`
}
```

Then:

```sh
task config:document   # regenerate .env.example
task config:check      # validate
```

## Adding a New Section

Add the struct and hang it off `Root`:

```go
// MyFeature groups the settings for … Why its own struct: the section is
// bound as a unit and passed to the subsystem that owns it, so a caller
// cannot reach a field that is not theirs.
type MyFeature struct {
    Enabled    bool          `env:"MYFEATURE_ENABLE" default:"false" doc:"master switch"`
    APIKey     string        `env:"MYFEATURE_API_KEY" secret:"true"`
    Timeout    time.Duration `env:"MYFEATURE_TIMEOUT" default:"30s"`
    MaxRetries int           `env:"MYFEATURE_MAX_RETRIES" default:"3"`
}

type Root struct {
    // … existing sections
    MyFeature MyFeature
}
```

No registration step, no regeneration of bindings — the struct is the binding.

## Accessing Config in Go

```go
// Preferred: from the kernel, which loaded it once at startup.
func doSomething(kernel *volt.Kernel) {
    cfg := kernel.RootConfig

    if cfg.MyFeature.Enabled {
        apiKey := cfg.MyFeature.APIKey
        timeout := cfg.MyFeature.Timeout
        // …
    }
}

// In a binary that builds no kernel (a migrator, a one-shot tool):
cfg, err := config.Get()
if err != nil {
    return fmt.Errorf("loading configuration: %w", err)
}

// Or, where a failure genuinely cannot be handled — the exit is explicit:
cfg := config.GetOrDie()
```

`config.Get()` returns `(*Root, error)`. It does **not** panic: the previous loader panicked on a missing directory, which turned a configuration mistake into a stack trace at container start, after the rollout and before the logger existed to record it.

## Supported Field Types

| Go type | Example value |
|---|---|
| `string` | `value` |
| `int`, `uint16`, … | `42` |
| `bool` | `true`, `false`, `1`, `0`, `t`, `f` (case-insensitive) |
| `float64` | `3.14` |
| `time.Duration` | `30s`, `5m`, `1h30m` — **a unit is required** |
| `[]string` | `a,b,c` (override the separator with `delim:"…"`) |

`yes` / `no` / `on` / `off` are **not** booleans — that is the reading which turns the country code `NO` into `false`.

## Layers and Precedence

Lowest first: compiled-in defaults → `.env` in the working directory → the process environment. The last layer to claim a key wins, and the environment is last so an operator can always override a file without editing it.

`task config:explain` prints every field with the source that set it, secrets masked — the answer to "why is this value what it is in staging".

## Gotchas

- **An empty environment variable is a value, not an absence.** `PORT=` sets the port to the empty string and fails to parse; it does not fall back to the default. In tests, *unset* rather than blank.
- **Scalars are not whitespace-trimmed.** `PORT=" 8080"` fails to parse.
- **`VOLT_ROOT` anchors `tmp/`, `uploads/` and the log file.** Nothing searches parent directories for a config file, deliberately — the previous loader walked up three levels, which meant those paths moved depending on where the process was launched from.

## Running

```sh
task config:check                    # validate
CONFIG_ENV=prod task config:check    # validate the production rules
task config:explain                  # which layer set each field
task config:document                 # regenerate .env.example

CONFIG_ENV=prod go run ./modules/webapp/cmd/api
```

Full guide: [`volt/docs/guides/configuration.md`](../docs/guides/configuration.md). Module reference: [`modules/config/README.md`](../../modules/config/README.md).
