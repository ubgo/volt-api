package main

import (
	"config"
	"context"
	"core/pkg/clicmd"
	"core/pkg/volt"
	"fmt"
	"lace/cli"
	"webapp"

	"github.com/spf13/cobra"
	cobracfg "github.com/ubgo/cfgkit/contrib/cli-cobra"
)

// The CLI owns the root command and builds the Circuit LAZILY.
//
// Building it here unconditionally — which is what this file used to do — meant
// every invocation loaded the configuration, pinged Redis and ran the
// non-dev encryption-key check before cobra had chosen a subcommand. So `ping`
// and `--help` required infrastructure, and under CONFIG_ENV=prod the process
// exited on a placeholder secret before any command could run, which made the
// configuration commands unreachable exactly when they were needed to diagnose
// it. Commands now receive an accessor and pull the Circuit only if they use it.
func main() {
	root := cli.New()

	circuit := webapp.Lazy()
	kernel := func() *volt.Kernel { return &circuit().Kernel }

	clicmd.AddCommands(root.RootCmd, kernel)
	AddCommands(root.RootCmd, circuit)

	// The config verbs come from cfgkit and need NO kernel — no database, no
	// Redis, no valid configuration. That is what makes them usable when the
	// configuration is the thing that is broken. config.Options() is passed so
	// the command inspects exactly what the application loads.
	root.RootCmd.AddCommand(cobracfg.Command[config.Root](config.Options()))

	root.Execute()
}

// AddCommands registers this application's own commands.
//
// It takes the same lazy accessor as the platform commands so an app command
// that needs no dependencies stays runnable without them.
func AddCommands(rootCmd *cobra.Command, circuit func() *webapp.Circuit) {
	rootCmd.AddCommand(seedUsersCommand(circuit))
}

func seedUsersCommand(circuit func() *webapp.Circuit) *cobra.Command {
	return &cobra.Command{
		Use:   "seed-users",
		Short: "Seed dummy users into the database",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := circuit().EntDB.Client()

			users := []struct {
				FirstName string
				LastName  string
				Email     string
			}{
				{"John", "Doe", "john.doe@example.com"},
				{"Jane", "Smith", "jane.smith@example.com"},
				{"Bob", "Johnson", "bob.johnson@example.com"},
				{"Alice", "Williams", "alice.williams@example.com"},
				{"Charlie", "Brown", "charlie.brown@example.com"},
				{"Diana", "Davis", "diana.davis@example.com"},
				{"Edward", "Miller", "edward.miller@example.com"},
				{"Fiona", "Wilson", "fiona.wilson@example.com"},
				{"George", "Taylor", "george.taylor@example.com"},
				{"Hannah", "Anderson", "hannah.anderson@example.com"},
			}

			for _, u := range users {
				_, err := client.User.Create().
					SetFirstName(u.FirstName).
					SetLastName(u.LastName).
					SetEmail(u.Email).
					SetIsActive(true).
					Save(ctx)
				if err != nil {
					fmt.Printf("Failed to create user %s: %v\n", u.Email, err)
					continue
				}
				fmt.Printf("Created user: %s %s (%s)\n", u.FirstName, u.LastName, u.Email)
			}

			fmt.Println("Seeding completed!")
		},
	}
}
