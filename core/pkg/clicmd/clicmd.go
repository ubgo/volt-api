package clicmd

import (
	"context"
	"core/pkg/applog"
	"core/pkg/plugin"
	"dbent/pkg/dbent_migrate"
	"fmt"
	"lace/crypt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	buildinfootel "github.com/ubgo/buildinfo/contrib/buildinfo-otel"
	ulog "github.com/ubgo/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

func AddCommands(plugin *plugin.Plugin) {
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
}

func dbEntMigrateCommand(plugin *plugin.Plugin) *cobra.Command {
	// This command allows running migrations from within the main CLI binary.
	// Useful for Docker deployments where we don't want to build a separate
	// migration binary. In development, if the CLI fails to compile, you can
	// still run migrations directly via: go run ./dbent/cmd/dbmigrate

	return &cobra.Command{
		Use:   "migrate",
		Short: "run migration command after re-deploy the app",
		Run: func(cmd *cobra.Command, args []string) {
			ctx, span := otel.Tracer("cli").Start(context.Background(), "migrate-cli-command")
			span.SetAttributes(
				buildinfootel.Attributes(attribute.String("job.name", "migrate"))...,
			)

			err := dbent_migrate.Migrate(ctx, plugin.DB)
			if err != nil {
				applog.FromCtx(ctx).Error("failed to migrate the database", ulog.Err(err))
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
