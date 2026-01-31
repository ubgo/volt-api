package main

import (
	"context"
	"dbent/pkg/dbent_migrate"
	"fmt"
	"lace/buildinfo"
	"lace/crypt"
	"lace/gotel"
	"lace/gozap"
	"os"
	"path/filepath"
	"saas/pkg/app"
	"time"

	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
)

func main() {
	app := app.New()
	AddCommands(app)
	app.Cli.Execute()
}

func AddCommands(plugin *app.Plugin) {
	var pingCommand = &cobra.Command{
		Use:   "ping",
		Short: "Just a ping command",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("pong")
		},
	}

	var writeDateTime = &cobra.Command{
		Use:   "writedatetime",
		Short: "Write current date time to a tmp file",
		Long:  "Creates a file in the tmp directory with current date and time",
		Run: func(cmd *cobra.Command, args []string) {
			currentTime := time.Now().Format("2006-01-02 15:04:05") + "\n"
			tmpFile := filepath.Join(plugin.RootConfig.Core.Tmpdir, "datetime.txt")

			// Open file in append mode, create if not exists
			f, err := os.OpenFile(tmpFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Error opening file: %v\n", err)
				return
			}
			defer f.Close()

			if _, err := f.Write([]byte(currentTime)); err != nil {
				fmt.Printf("Error writing to file: %v\n", err)
				return
			}

			fmt.Printf("DateTime appended to: %s\n", tmpFile)
		},
	}

	rootCmd := plugin.Cli.RootCmd
	rootCmd.AddCommand(pingCommand)
	rootCmd.AddCommand(dbEntMigrateCommand(plugin))
	rootCmd.AddCommand(writeDateTime)
	rootCmd.AddCommand(cryptCommand())
	rootCmd.AddCommand(seedUsersCommand(plugin))
}

func dbEntMigrateCommand(plugin *app.Plugin) *cobra.Command {
	// This command allows running migrations from within the main CLI binary.
	// Useful for Docker deployments where we don't want to build a separate
	// migration binary. In development, if the CLI fails to compile, you can
	// still run migrations directly via: go run ./dbent/cmd/dbmigrate

	return &cobra.Command{
		Use:   "migrate",
		Short: "run migration command after re-deploy the app",
		Run: func(cmd *cobra.Command, args []string) {
			telem := gotel.MustGetTelemetry()
			ctx, span := telem.TraceStart(context.Background(), "migrate-cli-command")
			span.SetAttributes(
				buildinfo.WrapOtelAttributes(attribute.String("job.name", "migrate"))...,
			)

			err := dbent_migrate.Migrate(ctx, plugin.DB)
			if err != nil {
				gozap.FromCtx(ctx).Error("failed to migrate the database", zap.Error(err))
				return
			}
			span.SetStatus(codes.Ok, "success")
			span.AddEvent("job_completed")

			span.End()
			time.Sleep(10 * time.Second)
		},
	}
}

/*
- go run ./cmd/cli crypt --encrypt "hello world"
- go run ./cmd/cli crypt -e "hello world"
- go run ./cmd/cli crypt -d "encrypted text here"
*/
func cryptCommand() *cobra.Command {
	var encrypt bool
	var decrypt bool

	cmd := &cobra.Command{
		Use:   "crypt",
		Short: "Encrypt or decrypt text",
		Long:  "Encrypt or decrypt text using AES encryption. Use --encrypt or --decrypt flag to specify operation.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			text := args[0]

			if encrypt {
				encrypted, err := crypt.Encrypt(text)
				if err != nil {
					fmt.Printf("Error encrypting text: %v\n", err)
					return
				}
				fmt.Printf("Encrypted: %s\n", encrypted)
			} else if decrypt {
				decrypted, err := crypt.Decrypt(text)
				if err != nil {
					fmt.Printf("Error decrypting text: %v\n", err)
					return
				}
				fmt.Printf("Decrypted: %s\n", decrypted)
			} else {
				fmt.Println("Please specify either --encrypt or --decrypt flag")
				cmd.Help()
			}
		},
	}

	cmd.Flags().BoolVarP(&encrypt, "encrypt", "e", false, "Encrypt the provided text")
	cmd.Flags().BoolVarP(&decrypt, "decrypt", "d", false, "Decrypt the provided text")
	cmd.MarkFlagsMutuallyExclusive("encrypt", "decrypt")

	return cmd
}

func seedUsersCommand(plugin *app.Plugin) *cobra.Command {
	return &cobra.Command{
		Use:   "seed-users",
		Short: "Seed dummy users into the database",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := context.Background()
			client := plugin.EntDB.Client()

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
