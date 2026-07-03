package main

import (
	"context"
	"core/pkg/clicmd"
	"fmt"
	"webapp"

	"github.com/spf13/cobra"
)

func main() {
	circuit := webapp.New()
	clicmd.AddCommands(&circuit.Kernel)
	AddCommands(circuit)
	circuit.Cli.Execute()
}

func AddCommands(circuit *webapp.Circuit) {
	rootCmd := circuit.Cli.RootCmd
	rootCmd.AddCommand(seedUsersCommand(circuit))
}

func seedUsersCommand(circuit *webapp.Circuit) *cobra.Command {
	return &cobra.Command{
		Use:   "seed-users",
		Short: "Seed dummy users into the database",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := circuit.EntDB.Client()

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
