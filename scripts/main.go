package main

import (
	"fmt"
	"os"
	secretsbuild "scripts/internal/pkg/secrets_build"

	"github.com/spf13/cobra"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "scripts",
	Short: "Scripts CLI tool",
	Long:  "A CLI tool for running various scripts and utilities",
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  "Print the version number of the scripts CLI tool",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("scripts version %s\n", version)
	},
}

var readSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:read",
	Short: "Read secrets from the .env file",
	Long:  "Read secrets from the .env file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Reading secrets from the .github/.secrets file")
		secretsbuild.ReadSecretsFile(secretsbuild.GetSecretFilePath())
	},
}

var updateSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:update",
	Short: "Build and update secrets in the .github/.secrets file",
	Long:  "Build and update secrets in the .github/.secrets file",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Updating secrets in the .github/.secrets file and create .github/workflows/deploy.yml")
		secretsbuild.BuldAndUpdateSecrets()
	},
}

var pushSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:push",
	Short: "Push secrets to GitHub",
	Long:  "Push secrets to GitHub",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Pushing secrets to GitHub")
		secretsbuild.PushSecretsToGithub()
	},
}

var cleanSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:clean",
	Short: "Remove GitHub Actions secrets that are no longer in the expected set",
	Long:  "Lists repo secrets on GitHub and deletes any that are not in the expected set (from .secrets.prod, prod key files, and ENV_PROD_B64).",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Cleaning orphan secrets from GitHub...")
		secretsbuild.CleanNonExistingSecrets()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(readSecretsCmd)
	rootCmd.AddCommand(updateSecretsCmd)
	rootCmd.AddCommand(pushSecretsCmd)
	rootCmd.AddCommand(cleanSecretsCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
