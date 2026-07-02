# New CLI Command Rules

**Read before:** Adding a new CLI command.

## Checklist

- [ ] Add command in `saas/cmd/cli/main.go`
- [ ] Register with `rootCmd.AddCommand()`
- [ ] Add tracing if long-running
- [ ] Test with `go run ./saas/cmd/cli <command>`

## File Location

| Type | Location |
|------|----------|
| Commands | `saas/cmd/cli/main.go` |
| Complex commands | `saas/cmd/cli/<command>.go` (separate file) |

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

```go
func AddCommands(plugin *app.Plugin) {
    var myCommand = &cobra.Command{
        Use:   "mycommand",
        Short: "Brief description",
        Long:  "Longer description of what this command does",
        Run: func(cmd *cobra.Command, args []string) {
            // Command logic
            fmt.Println("Command executed")
        },
    }

    rootCmd := plugin.Cli.RootCmd
    rootCmd.AddCommand(myCommand)
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
func tracedCommand(plugin *app.Plugin) *cobra.Command {
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
go run ./saas/cmd/cli mycommand --flag=value

# With args
go run ./saas/cmd/cli mycommand arg1 arg2

# Help
go run ./saas/cmd/cli mycommand --help
```
