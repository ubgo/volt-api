# The CLI and the kernel

How a command gets registered, when the application's dependencies get built, and why those are two separate moments.

## The rule in one line

**Registering a command is free. Building the kernel is not. So the kernel is built only if the command you actually typed asks for it.**

## What the kernel is

`volt.Kernel` is the application's live dependencies — the database handle, the Ent client, the cache, Redis, the loaded configuration, the health registry. `volt.New()` builds it, and building it is expensive and *fallible*:

1. loads and validates the configuration (`config.GetOrDie()` — exits on any problem)
2. opens the database pool
3. **pings Redis** and panics if unreachable (`REDIS_STATUS` defaults to `true`)
4. **refuses a placeholder encryption key** outside dev
5. initializes global logging and telemetry

Every one of those is correct for a running service. Not one of them is appropriate for `cli ping`.

## What the kernel is not

It is **not** the CLI. `Kernel` deliberately holds no cobra root command.

It used to. The root command lived at `Kernel.Cli.RootCmd`, which meant the only way to reach the root command was to build a kernel first — so `main` had to construct the entire application before cobra could look at `os.Args` and decide which command you wanted:

```go
// the old shape
func main() {
	circuit := webapp.New()            // ← config, Redis, key check — ALWAYS
	clicmd.AddCommands(&circuit.Kernel)
	circuit.Cli.Execute()              // ← cobra picks your command only HERE
}
```

Line 1 runs before line 3. Typing `ping` still paid for all of it.

## The shape now

```go
func main() {
	root := cli.New()                                        // a plain cobra root

	circuit := webapp.Lazy()                                 // a FUNCTION; builds nothing
	kernel := func() *volt.Kernel { return &circuit().Kernel }

	clicmd.AddCommands(root.RootCmd, kernel)                 // hand out the function
	AddCommands(root.RootCmd, circuit)

	root.Execute()                                           // cobra picks the command...
}                                                            // ...which calls the function only if it needs to
```

`webapp.Lazy()` returns an accessor rather than a Circuit:

```go
func Lazy(options ...Option) func() *Circuit {
	return sync.OnceValue(func() *Circuit { return New(options...) })
}
```

`sync.OnceValue` is stdlib: call the accessor ten times, `New()` runs once and the other nine calls get the same value. So a lazy CLI keeps the same guarantee the api binary has — **at most one kernel, and therefore one configuration snapshot, per process**.

## Writing a command

The only rule: **call the accessor inside `Run`, never at construction time.**

```go
// Needs nothing — never calls kernel(). Runs on a laptop with no database,
// no Redis, and no valid configuration.
&cobra.Command{
	Use: "ping",
	Run: func(cmd *cobra.Command, args []string) { fmt.Println("pong") },
}

// Needs the database — calls kernel() when it actually runs.
&cobra.Command{
	Use: "migrate",
	Run: func(cmd *cobra.Command, args []string) {
		err := dbent_migrate.Migrate(ctx, kernel().DB)
	},
}
```

Reading `kernel().RootConfig.X` while *building* the command tree defeats the whole design — that loads configuration for every invocation of every command, which is the bug this exists to prevent. `crypt` is the worked example: it used to receive the encryption key as a `string` argument at construction, and now reads it from `kernel()` inside `Run`.

## Plugging in a command from anywhere

This is what falling out of the kernel buys. The registration surface is:

```go
func AddCommands(rootCmd *cobra.Command, kernel clicmd.KernelFunc)
```

`rootCmd` is a plain `*cobra.Command`. So **any package that can produce a `*cobra.Command` can be plugged in, with no knowledge of volt, core, or the kernel** — its only import is `github.com/spf13/cobra`:

```go
root.RootCmd.AddCommand(somethirdparty.Command())
```

Two boundaries worth stating precisely, because they are easy to blur:

- **A command you plug in owes nothing to volt.** It does not need a kernel, a `KernelFunc`, or any volt type. If it needs *nothing*, it will run under any conditions the binary itself runs under.
- **`clicmd.AddCommands` still takes a `KernelFunc`** — but that is because *volt's own* commands (`migrate`, `crypt`, `writedatetime`) need a kernel. That is a property of those commands, not a tax on yours.

This is not hypothetical — it is how the `config` command works today:

```go
// modules/webapp/cmd/cli/main.go
root.RootCmd.AddCommand(cobracfg.Command[config.Root](config.Options()))
```

`cobracfg` is [`github.com/ubgo/cfgkit/contrib/cli-cobra`](https://github.com/ubgo/cfgkit/tree/main/contrib/cli-cobra). It knows nothing about volt, imports no volt package, and opens no database — which is exactly why it works when the configuration is the thing that is broken. `config.Options()` is exported so the command inspects **precisely** what the application loads; a command checking a different source list would be checking something the application never runs.

Under the old shape this was impossible: the kernel booted first, and a `CONFIG_ENV=prod` invocation with a placeholder secret died before cobra ever saw the command — so the tool for diagnosing a broken configuration was unreachable precisely when the configuration was broken.

## Verified behaviour

All rows below are `CONFIG_ENV=prod` with the shipped placeholder secrets — the state a misconfigured deployment is actually in.

| Command | Before | After |
|---|---|---|
| `cli ping` | exit 1, three config errors | `pong` |
| `cli --help` | exit 1, three config errors | help text |
| `cli ping`, Redis at an unroutable address | panic on the Redis ping | `pong` |
| `cli migrate` | exit 1, three config errors | exit 1, three config errors |
| `cli config check` | did not exist; would have exited 1 before running | runs, reports the three problems, exits 1 |
| `cli config explain`, DB and Redis unreachable | did not exist | prints the full provenance table |

The `migrate` row is the one that matters most. The safety property did not loosen: a command that touches the database still refuses to run against placeholder secrets in a non-dev environment. What changed is that commands which touch nothing no longer pay for infrastructure they never use — and that the tool for diagnosing the failure now runs *during* the failure.

A note on the database specifically: `initDB` does **not** ping. It opens a `*sql.DB`, which is lazy, and registers a live readiness checker instead of a boot-time ping. So an unreachable database was never what blocked the CLI — Redis and the configuration were.

## Which binary builds what

| Binary | Builds a kernel? |
|---|---|
| `api` | Yes, once, in `webapp.New()` → `volt.New()`. `main` reads `circuit.RootConfig` rather than loading config a second time. |
| `cron` | Yes, same path. |
| `cli` | **Only if the invoked command calls the accessor.** |
| `migrate` (standalone) | No kernel — it calls `config.GetOrDie()` directly, since it needs a database host and nothing else. |

## See also

- [Configuration](configuration.md) — the layers, validation, and what `check` / `explain` / `document` do.
- [`../rules/new-cli-command.md`](../../rules/new-cli-command.md) — the checklist for adding one.
- [CLI reference](../reference/cli.md) — every command the binary ships.
