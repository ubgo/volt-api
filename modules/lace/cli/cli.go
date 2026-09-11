package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

type Cli struct {
	RootCmd *cobra.Command
}

func (cli Cli) Execute() {
	if err := cli.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func New() Cli {
	rootCmd := &cobra.Command{
		Use: "saasfly",
		// Execute below is the single place an error reaches the operator.
		// Without this, cobra prints the error too and every failure appears
		// twice — which on a multi-line config report means the same list of
		// bad keys scrolling past a second time.
		SilenceErrors: true,
		// 		Long: `
		// Framework is a Fast and Flexible Go Boilerplate built with love @ Khanakia Inc.`,
	}

	cli := Cli{
		RootCmd: rootCmd,
	}
	return cli
}
