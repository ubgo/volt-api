# New CLI Command Rules

**Read before:** Adding a new CLI command.

## Checklist

- [ ] Add command in `webapp/cmd/cli/main.go`
- [ ] Register with `rootCmd.AddCommand()`
- [ ] Add tracing if long-running
- [ ] Test with `go run ./webapp/cmd/cli <command>`

## File Location

| Type | Location |
|------|----------|
| Commands | `webapp/cmd/cli/main.go` |
| Complex commands | `webapp/cmd/cli/<command>.go` (separate file) |

## Do

- Use Cobra command structure
- Access services via `plugin` parameter
- Add tracing for operations that touch DB/external services
- Use `fmt.Println` for user output
- Handle errors gracefully (don't panic)

## Don't

- Use `log.Fatal` (prevents graceful shutdown)
- Create separate binaries for simple commands
- Forget to register command with `AddCommand()`

## Simple Command

Commands receive a `clicmd.KernelFunc` — an accessor that builds the kernel on first call — not a built `*volt.Kernel`. A command that needs no dependencies simply never calls it, and stays runnable on a machine with no database, no Redis and no valid configuration.

```go
func AddCommands(rootCmd *cobra.Command, kernel clicmd.KernelFunc) {
    var myCommand = &cobra.Command{
        Use:   "mycommand",
        Short: "Brief description",
        Long:  "Longer description of what this command does",
        Run: func(cmd *cobra.Command, args []string) {
            // Needs nothing? Never call kernel(). This command then works
            // even under CONFIG_ENV=prod with placeholder secrets.
            fmt.Println("Command executed")
        },
    }

    rootCmd.AddCommand(myCommand)
}
```

**Call `kernel()` inside `Run`, never at construction.** Reading `kernel().RootConfig.X` while building the command tree loads the configuration for *every* invocation of *every* command — which is the bug this design exists to prevent:

```go
Run: func(cmd *cobra.Command, args []string) {
    cfg := kernel().RootConfig            // built here, on demand
    db := kernel().DB                     // same instance; sync.OnceValue
    ...
}
```

## Command with Flags

```go
func myCommand() *cobra.Command {
    var flagName string
    var verbose bool

    cmd := &cobra.Command{
        Use:   "mycommand [args]",
        Short: "Brief description",
        Args:  cobra.ExactArgs(1),  // or MinimumNArgs, MaximumNArgs
        Run: func(cmd *cobra.Command, args []string) {
            input := args[0]
            if verbose {
                fmt.Printf("Processing: %s with flag: %s\n", input, flagName)
            }
            // Logic here
        },
    }

    cmd.Flags().StringVarP(&flagName, "name", "n", "default", "Description")
    cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
    cmd.MarkFlagRequired("name")  // if required

    return cmd
}

// Register in AddCommands:
rootCmd.AddCommand(myCommand())
```

## Command with Tracing

```go
func tracedCommand(kernel clicmd.KernelFunc) *cobra.Command {
    return &cobra.Command{
        Use:   "traced",
        Short: "Command with OpenTelemetry tracing",
        Run: func(cmd *cobra.Command, args []string) {
            // Telemetry is bootstrapped by otelkit at startup; spans use the
            // standard OTel API (no lace/gotel helper in v2).
            ctx, span := otel.Tracer("cli").Start(context.Background(), "cli.traced")
            span.SetAttributes(
                attribute.String("command", "traced"),
            )
            defer span.End()

            // Do work with ctx
            err := doWork(ctx, plugin)
            if err != nil {
                applog.FromCtx(ctx).Error("command failed", ulog.Err(err))
                span.SetStatus(codes.Error, err.Error())
                return
            }

            span.SetStatus(codes.Ok, "success")
            span.AddEvent("command_completed")

            // Allow telemetry to flush
            time.Sleep(2 * time.Second)
        },
    }
}
```

## Subcommands

```go
func parentCommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "parent",
        Short: "Parent command with subcommands",
    }

    cmd.AddCommand(subCommand1())
    cmd.AddCommand(subCommand2())

    return cmd
}
```

## Running

```bash
# Development
go run ./webapp/cmd/cli mycommand --flag=value

# With args
go run ./webapp/cmd/cli mycommand arg1 arg2

# Help
go run ./webapp/cmd/cli mycommand --help
```
