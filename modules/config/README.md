# Config Module

The application's configuration: schema, defaults and validation, defined in Go.

There is no configuration language, no code generation, and no config files in the production image. `config.go` **is** the schema — which means it cannot drift from what the code declares, because there is nothing else to drift from.

Built on [`github.com/ubgo/cfgkit`](https://github.com/ubgo/cfgkit).

## The one-minute version

```go
appcfg := config.GetOrDie()      // or config.Get() for an error you can handle
fmt.Println(appcfg.Server.Port)  // typed, validated, with a compiled-in default
```

Three commands answer almost every question:

```sh
task config:check      # is this configuration valid? non-zero exit if not
task config:explain    # which source set each field? (secrets masked)
task config:document   # regenerate .env.example from the struct
```

## Zero input is a valid configuration

Every field has a compiled-in default, so the service configures itself with no `.env` file, no environment variables and no tooling installed. That is the property everything else hangs off: a newcomer can clone and run, a test can construct a config without fixtures, and CI needs no secrets to compile.

## Layers, lowest precedence first

| # | Layer | Where it comes from |
|---|---|---|
| 1 | Compiled-in defaults | the `default:` struct tags in `config.go` |
| 2 | `.env` | the process's working directory, if present |
| 3 | Environment variables | the real process environment |

Precedence is positional — the last layer to claim a key wins — and the order lives in `sources()` in `main.go`. The environment is last so an operator can always override a file without editing it.

The file is **not** searched for up the directory tree. The previous loader walked up three levels looking for a config directory, which meant `tmp/`, `uploads/` and the log file moved depending on where the process was launched from. Anchor paths with `VOLT_ROOT` instead: one variable, explicit.

## Files

| File | What lives there |
|---|---|
| `config.go` | Every struct, tag, `Derive()` and `Validate()` — the whole schema |
| `main.go` | `Get` / `GetOrDie` / `Options` and the source order — four exports, nothing else |
| `config_test.go` | Pins the defaults, the derivations, the validation rules and the env-key list |
| `Taskfile.yml` | Those three tasks |

The `check` / `explain` / `document` verbs are **not implemented here**, and there are deliberately no wrapper functions for them either — one implementation, not two that can disagree about what "check" means. They come from [`github.com/ubgo/cfgkit/contrib/cli-cobra`](https://github.com/ubgo/cfgkit/tree/main/contrib/cli-cobra), mounted on the application CLI in `modules/webapp/cmd/cli/main.go`:

```go
root.RootCmd.AddCommand(cobracfg.Command[config.Root](config.Options()))
```

`config.Options()` is exported precisely so the command inspects **exactly** what the application loads — a command checking a different source list would be checking something the application never runs. The command needs no kernel, so it works with the database and cache unreachable, which is when it matters.

## The structs

`Root` composes them: `Server`, `Core`, `Database`, `Graphql`, `Otel`, `Logger` (which embeds `LocalConfig`), `Redis`, `ApiKeys`, `Gocron`, `App`.

Tags carried by fields:

| Tag | Meaning |
|---|---|
| `env:"KEY"` | which environment variable / `.env` key binds this field |
| `default:"…"` | the compiled-in value; its absence means the zero value |
| `secret:"true"` | structurally masked — never reaches `explain`, JSON, or an error message |
| `doc:"…"` | the comment written above the key in `.env.example` |

## Derive and Validate

`Derive()` runs **after** every source and fills only what is still empty, so an explicit value is never overwritten:

| Field | Derived as | Unless you set |
|---|---|---|
| `Core.Tmpdir` | `<VOLT_ROOT>/tmp` | `TMP_DIR` |
| `Core.Uploadsdir` | `<VOLT_ROOT>/uploads` | `UPLOADS_DIR` |
| `Logger.LocalConfig.LogFile` | `<tmp>/logs/app.log` | `LOGGER_LOG_FILE` |
| `Otel.ServiceName` | `Core.AppKey` | `OTEL_SERVICE_NAME` |

`Validate()` reports **every** problem at once via `errors.Join`, not the first one — a misconfigured deploy should be one fix, not one fix per restart. The rules:

- `CONFIG_ENV` must be `dev`, `stag` or `prod`. A misspelling like `production` fails rather than being read as "not prod" and quietly disabling every production rule below.
- `DB_PORT` must be a valid port; `DB_SSLMODE` must be one of Postgres's modes.
- `APP_KEY`, `HOSTNAME`, `DB_HOST`, `DB_NAME` must not be empty.
- Outside dev, the shipped placeholder secrets (`knownPlaceholderSecrets`) are refused. Shipping the default encryption key would make every encrypted record readable by anyone holding the source, so it fails the deploy rather than the audit.

Dev tolerates the placeholders. That is what keeps zero-input working.

## Adding a field

1. Add it to the right struct with an `env:` key, a `default:` if one is sensible, `secret:"true"` if it is a credential, and a `doc:` line.
2. If it is a credential with a shipped placeholder, add that placeholder to `knownPlaceholderSecrets` so production refuses it.
3. If it needs a computed default, extend `Derive()`.
4. If it has a constraint, extend `Validate()`.
5. `task config:document` to regenerate `.env.example`.
6. Add the key to `readsEnv` in `config_test.go` — a test fails if you forget, because otherwise the suite would inherit that variable from the developer's environment.

## Gotchas

- **An empty environment variable is a value, not an absence.** `PORT=` sets the port to the empty string and fails to parse; it does not fall back to the default. An operator who clears a variable meant to clear it. In tests, *unset* rather than blank.
- **Scalars are not whitespace-trimmed.** `PORT=" 8080"` fails to parse. List and map elements are trimmed, because a delimiter already implies structure.
- **Durations need a unit** — `30s`, `1h30m`, not `500`.
- **Booleans are Go's set**: `true/false/1/0/t/f` and case variants. `yes`/`no`/`on`/`off` are rejected — that is the reading which turns the country code `NO` into `false`.

## Full guide

[`volt/docs/guides/configuration.md`](../../volt/docs/guides/configuration.md) — the runtime flow, per-binary load points, live `explain` output, and the background on what this replaced.
