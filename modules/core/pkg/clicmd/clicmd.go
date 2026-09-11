package clicmd

import (
	"context"
	"core/pkg/applog"
	"core/pkg/volt"
	"dbent/pkg/dbent_migrate"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	buildinfootel "github.com/ubgo/buildinfo/contrib/buildinfo-otel"
	ubgocrypt "github.com/ubgo/crypt"
	ulog "github.com/ubgo/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

// KernelFunc builds the kernel on demand.
//
// Why a function and not a *Kernel: constructing the kernel loads the
// configuration, pings Redis and refuses a placeholder encryption key outside
// dev — all of which are correct for a running service and wrong for `ping`,
// `--help`, or a command that only inspects configuration. Taking a func lets
// each command decide, and callers should wrap theirs in sync.OnceValue so a
// process still builds at most one kernel.
//
// Invariant: implementations exit or panic on failure rather than returning
// nil, matching volt.New — a command body may use the result without a nil check.
type KernelFunc func() *volt.Kernel

// AddCommands registers the platform commands on rootCmd.
//
// rootCmd is passed in rather than read off the kernel because registration
// must happen before any kernel exists; see the Kernel doc comment.
func AddCommands(rootCmd *cobra.Command, kernel KernelFunc) {
	// ping takes no KernelFunc at all: it is the smoke test for "is this binary
	// runnable", which must stay true on a machine with no database, no Redis
	// and no configuration.
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
			tmpFile := filepath.Join(kernel().RootConfig.Core.Tmpdir, "datetime.txt")

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

	rootCmd.AddCommand(pingCommand)
	rootCmd.AddCommand(dbEntMigrateCommand(kernel))
	rootCmd.AddCommand(writeDateTime)
	rootCmd.AddCommand(cryptCommand(kernel))
	rootCmd.AddCommand(cryptRekeyCommand())
}

func dbEntMigrateCommand(kernel KernelFunc) *cobra.Command {
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

			err := dbent_migrate.Migrate(ctx, kernel().DB)
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
func cryptCommand(kernel KernelFunc) *cobra.Command {
	var encrypt bool
	var decrypt bool

	cmd := &cobra.Command{
		Use:   "crypt",
		Short: "Encrypt or decrypt text",
		Long:  "Encrypt or decrypt text using AES encryption. Use --encrypt or --decrypt flag to specify operation.",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			text := args[0]

			// Read inside Run, not at construction: the key comes from the
			// configuration, and reading it while building the command tree
			// would load config for every invocation of every command.
			key := []byte(kernel().RootConfig.Core.EncryptionKey)

			// AES-CBC (hex(IV||PKCS7) wire format) preserves compatibility with
			// values previously written by the old lace/crypt package.
			if encrypt {
				encrypted, err := ubgocrypt.EncryptCBC(key, []byte(text))
				if err != nil {
					fmt.Printf("Error encrypting text: %v\n", err)
					return
				}
				fmt.Printf("Encrypted: %s\n", encrypted)
			} else if decrypt {
				decrypted, err := ubgocrypt.DecryptCBC(key, text)
				if err != nil {
					fmt.Printf("Error decrypting text: %v\n", err)
					return
				}
				fmt.Printf("Decrypted: %s\n", string(decrypted))
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
