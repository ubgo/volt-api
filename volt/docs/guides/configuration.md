# Configuration

How a value gets from a default, a file, or an environment variable into the struct your code reads — and how to find out which one won.

Configuration is defined in Go, in one file: `modules/config/config.go`. That file is the schema, the defaults and the validation together. Nothing is generated, so nothing can drift from what the code declares.

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

Every field has a compiled-in default, so the service configures itself with **no `.env` file, no environment variables and no tooling installed**:

```
$ cd /tmp/empty && volt config check
configuration is valid
```

This is the property everything else hangs off. A newcomer can clone and run. A test can construct a config without fixtures. CI needs no secrets to compile.

## The layers, lowest precedence first

| # | Layer | Where it comes from |
|---|---|---|
| 1 | **Compiled-in defaults** | the `default:` struct tags in `config.go` |
| 2 | **`.env`** | the process's working directory, if present |
| 3 | **Environment variables** | the real process environment |

Precedence is **positional**: the value of a field is whatever the *last* layer to claim its key said. Nothing is reordered internally, so the order in `modules/config/main.go` is the order that runs.

The environment is last on purpose — an operator must always be able to override a file without editing it, which is what makes `docker run -e PORT=9999` behave the way anyone expects.

### Seeing it happen

Starting from nothing:

```
FIELD             KEY      VALUE      SOURCE
Server.Port       PORT     9001       default
Database.Host     DB_HOST  localhost  default
```

Add a `.env` containing `PORT=7777` and `DB_HOST=db.internal`:

```
Database.Host     DB_HOST  db.internal  file:.env
Server.Port       PORT     7777         file:.env
```

Now run with `PORT=9999` in the environment:

```
Database.Host     DB_HOST  db.internal  file:.env
Server.Port       PORT     9999         environ
```

Only the key that was overridden moves. That `SOURCE` column is the point: the question in production is rarely *what is this value* but *which layer decided it*.

### Typos are named, not swallowed

A key that matches no field binds nothing — so the field keeps its default and the table above looks entirely correct. `explain` names them rather than letting the mistake stay invisible:

```
mode: dev

1 key(s) matched no field (probable typos):
  DB_HSOT (from file:.env) matched no field
```

This is **advisory** by default, because `.env` legitimately serves more than one audience: the deploy keys (`VPS_*`, `GHCR_PAT`) are unknown to the binary by design. `task config:check` therefore passes. Add `--strict` when the file is owned by this application alone and an unrecognised key really is a typo:

```sh
go run ./modules/webapp/cmd/cli config check --strict
```

Only sources that can honestly enumerate their keys contribute, so the process environment never appears here — its key set is the whole machine, and a report listing `PATH` and `HOME` would bury the one line that matters.

## The load pipeline

`config.Get()` runs one pass, in this order:

1. **Defaults** — every `default:` tag is applied.
2. **Sources** — `.env` then the environment, in that order, each overriding what came before.
3. **Decode** — strings become `int`, `bool`, `uint16`, `time.Duration`, `[]string`. A value the type cannot hold is an error naming the key, not a silent zero.
4. **`Derive()`** — fills the fields whose defaults depend on other fields. Runs *after* every source, and only fills what is still empty, so an explicit value is never overwritten.
5. **`Validate()`** — reports **every** problem at once via `errors.Join`, not the first one.

### What Derive computes

| Field | Derived as | Unless you set |
|---|---|---|
| `Core.Tmpdir` | `<VOLT_ROOT>/tmp` | `TMP_DIR` |
| `Core.Uploadsdir` | `<VOLT_ROOT>/uploads` | `UPLOADS_DIR` |
| `Logger.LocalConfig.LogFile` | `<tmp>/logs/app.log` | `LOGGER_LOG_FILE` |
| `Otel.ServiceName` | `Core.AppKey` | `OTEL_SERVICE_NAME` |

`VOLT_ROOT` anchors all of it and defaults to `.`. Set it when the working directory is not the repository root — the root `Taskfile.yml` exports it for every task.

## Validation, and why it runs in CI

`task config:check` runs the whole pipeline and returns non-zero on any problem, without constructing the application. Run it per environment you deploy:

```sh
CONFIG_ENV=prod task config:check
```

```
cfgkit: 3 problem(s):
ENCRYPTION_KEY: is a default or placeholder value, which is refused outside dev
JWT_SECRET: is a default or placeholder value, which is refused outside dev
INTERNAL_KEY: is a default or placeholder value, which is refused outside dev
```

All three at once, not one per run — a misconfigured deploy should be one fix, not one fix per restart, because each restart is another failed rollout.

### The rules

- `CONFIG_ENV` must be `dev`, `stag` or `prod`. A misspelling like `production` **fails**, rather than being read as "not prod" and quietly disabling every production rule below.
- `DB_PORT` must be a valid port; `DB_SSLMODE` must be one of Postgres's modes.
- `APP_KEY`, `HOSTNAME`, `DB_HOST`, `DB_NAME` must not be empty.
- **Outside dev**, the shipped placeholder secrets are refused. Shipping the default encryption key would make every encrypted record readable by anyone holding the source, so it fails the deploy rather than the audit.

Dev tolerates the placeholders. That is what keeps zero-input working: a laptop with nothing set still boots, while a production deploy carrying a placeholder does not.

## Secrets

A field tagged `secret:"true"` is **structurally masked** — absent from `config:explain`, from the JSON view, and from error messages alike. Not hidden by output styling; the value never reaches the renderer.

```
Core.Jwtsecret    JWT_SECRET    ••••••    default
```

The field is still listed, so you can see *whether* it was set and *which layer* set it, without seeing what it is. That makes `task config:explain` output safe to paste into an issue.

## The contract file

`.env.example` is **generated** from the struct:

```sh
task config:document
```

Every key, its documentation comment, and its default — with secrets emitted empty, so the file is safe to commit, and it is: `.gitignore` blanket-ignores `.env*` and then negates this one path. It cannot drift from the code, because there is no hand-maintained copy to drift.

It is the **only** place the application's settings are enumerated. `env.sample`, the file `getting-started` tells you to copy, carries the deploy and Docker Compose variables (`VPS_*`, `GHCR_PAT`, `CONTAINER_NAME`, `PORT`) and **none** of the 44 config keys — the two files are complementary, not duplicates.

## Where the config goes at runtime

```
webapp/cmd/api/main.go
  └── webapp.New()                        modules/webapp/app.go
        └── volt.New()                    modules/core/pkg/volt/app.go
              ├── config.GetOrDie()       ← THE configuration is loaded here, once
              ├── initDB(cfg.Database)
              ├── initRedis(&cfg.Redis)
              ├── encryption-key check    (fails closed outside dev)
              └── initGlobalLogger(cfg.Otel, cfg.Logger, cfg.Core.Environment)

  Kernel.RootConfig  ──►  GraphQL resolver   (Graphql.Key)
                     ──►  HTTP server        (Server.Port, Core.Environment)
                     ──►  CLI commands       (Core.Tmpdir, Core.EncryptionKey)
```

`Circuit` embeds `volt.Kernel`, so `circuit.RootConfig` is the already-loaded configuration. **Read it from there rather than calling `config.GetOrDie()` again** — a second call builds a second, independent snapshot: wasted work, and two values that could in principle disagree about what the environment said.

Each binary loads once:

| Binary | Loads at |
|---|---|
| `api` | `volt.New()`, reached through `webapp.New()` |
| `cron` | same kernel path |
| `migrate` | its own `main`, since it does not build a kernel |
| `config` (the CLI) | per invocation |

## Adding a field

1. Add it to the right struct in `modules/config/config.go` with an `env:` key, a `default:` if one is sensible, `secret:"true"` if it is a credential, and a `doc:` line — the `doc:` text becomes the comment in `.env.example`.
2. If it is a credential with a shipped placeholder, add that placeholder to `knownPlaceholderSecrets` so production refuses it.
3. If it needs a computed default, extend `Derive()`.
4. If it has a constraint, extend `Validate()`.
5. `task config:document` to regenerate the contract.
6. Add the key to `readsEnv` in `config_test.go` — a test fails if you forget, because otherwise the suite would inherit that variable from the developer's environment.

## Gotchas

**An empty environment variable is a value, not an absence.** `PORT=` sets the port to the empty string and fails to parse — it does not fall back to the default. That is deliberate: an operator who clears a variable meant to clear it. In tests, *unset* rather than blank.

**Scalars are not whitespace-trimmed.** `PORT=" 8080"` fails to parse. List and map elements *are* trimmed, because a delimiter already implies structure.

**Durations need a unit.** `30s`, `1h30m` — not `500`. A bare number is ambiguous between milliseconds and nanoseconds, and being wrong about that is a timeout nobody notices.

**Booleans are Go's set**: `true/false/1/0/t/f` and case variants. `yes`, `no`, `on`, `off` are **rejected** — the reading that turns a country code `NO` into `false`.

**`VOLT_ROOT` matters more than it looks.** It anchors tmp, uploads and the log file. Nothing searches parent directories for a config file, deliberately: the previous loader walked up three levels, which meant those paths moved depending on where the process was launched from.

## Background

Configuration used to be Pkl schemas plus generated Go bindings. That required the `pkl` CLI and `pkl-gen-go` installed to develop, a codegen step whose output was committed, and a Pkl binary downloaded into the **production image** — plus a glibc compatibility layer to run it on Alpine.

It is now [`github.com/ubgo/cfgkit`](https://github.com/ubgo/cfgkit): defaults compile into the binary, the production image ships no config evaluator and no config files, and the loader returns an error instead of panicking on a missing directory.

See also: [Observability](observability.md) for how the otel config is consumed, and [OTEL service naming](otel-service-naming.md) for the `OTEL_SERVICE_NAME` convention.
