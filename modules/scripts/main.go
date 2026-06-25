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
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("scripts version %s\n", version)
	},
}

// Environment flag
var envFlag string

// gh:secrets:list - List available environments
var listEnvsCmd = &cobra.Command{
	Use:   "gh:secrets:list",
	Short: "List available environments",
	Run: func(cmd *cobra.Command, args []string) {
		envs := secretsbuild.ListAvailableEnvironments()
		if len(envs) == 0 {
			fmt.Println("No environments found. Create .github/.secrets.{env} files to configure environments.")
			return
		}
		fmt.Println("Available environments:")
		for _, env := range envs {
			fmt.Printf("  - %s\n", env)
		}
	},
}

// gh:secrets:update - Build and update secrets
var updateSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:update",
	Short: "Build and update secrets (generates .secrets.{env}.generated and deploy.yml)",
	Long: `Build and update secrets for deployment.

If --env is specified, builds secrets for that environment only.
Otherwise, builds secrets for all available environments (detected from .secrets.* files).

Environment names can be anything: prod, staging, dev, feature-xxx, etc.
Branch names like "feature/test" become "feature-test" as environment names.

Examples:
  go run ./scripts gh:secrets:update                     # All environments
  go run ./scripts gh:secrets:update --env=prod          # Production only
  go run ./scripts gh:secrets:update --env=feature-test  # Feature branch`,
	Run: func(cmd *cobra.Command, args []string) {
		if envFlag != "" {
			if !secretsbuild.IsValidEnvironment(envFlag) {
				fmt.Fprintf(os.Stderr, "Invalid environment: %s\nValid options: prod, staging, dev, uat\n", envFlag)
				os.Exit(1)
			}
			env := secretsbuild.Environment(envFlag)
			if err := secretsbuild.BuildAndUpdateSecrets(env); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Also generate deploy.yml
			if err := secretsbuild.CreateGithubDeployYaml([]secretsbuild.Environment{env}); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating deploy.yml: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := secretsbuild.BuildAndUpdateAllSecrets(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		fmt.Println("\nDone! Next steps:")
		fmt.Println("  1. Review generated files in .github/")
		fmt.Println("  2. Run: go run ./scripts gh:secrets:push [--env=ENV]")
	},
}

// gh:secrets:push - Push secrets to GitHub
var pushSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:push",
	Short: "Push secrets to GitHub Environments",
	Long: `Push secrets to GitHub Environments.

If --env is specified, pushes to that environment only.
Otherwise, pushes to all available environments.

Prerequisites:
  - GitHub CLI (gh) must be installed and authenticated
  - Run gh:secrets:update first to generate secrets

Examples:
  go run ./scripts gh:secrets:push              # All environments
  go run ./scripts gh:secrets:push --env=prod   # Production only`,
	Run: func(cmd *cobra.Command, args []string) {
		if envFlag != "" {
			if !secretsbuild.IsValidEnvironment(envFlag) {
				fmt.Fprintf(os.Stderr, "Invalid environment: %s\n", envFlag)
				os.Exit(1)
			}
			env := secretsbuild.Environment(envFlag)
			if err := secretsbuild.PushSecretsToGithubEnvironment(env); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := secretsbuild.PushSecretsToAllEnvironments(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// gh:secrets:clean - Clean orphan secrets
var cleanSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:clean",
	Short: "Remove orphan secrets from GitHub Environments",
	Long: `Remove secrets from GitHub Environments that are no longer in the expected set.

If --env is specified, cleans that environment only.
Otherwise, cleans all available environments.`,
	Run: func(cmd *cobra.Command, args []string) {
		if envFlag != "" {
			if !secretsbuild.IsValidEnvironment(envFlag) {
				fmt.Fprintf(os.Stderr, "Invalid environment: %s\n", envFlag)
				os.Exit(1)
			}
			env := secretsbuild.Environment(envFlag)
			if err := secretsbuild.CleanNonExistingSecrets(env); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := secretsbuild.CleanAllEnvironments(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// gh:env:setup - Full environment setup
var setupEnvCmd = &cobra.Command{
	Use:   "gh:env:setup",
	Short: "Full environment setup (create env, build & push secrets, push variables)",
	Long: `Performs complete environment setup:
  1. Creates GitHub Environment (if not exists)
  2. Builds secrets from local files
  3. Pushes secrets to GitHub Environment
  4. Pushes variables (CONTAINER_NAME, DEPLOY_PATH) to GitHub Environment
  5. Generates deploy.yml

Environment names can be anything: prod, staging, feature-test, etc.
Requires: .github/.secrets.{env} and optionally .env.{env} files.

If --env is specified, sets up that environment only.
Otherwise, sets up all available environments (detected from config files).

Examples:
  go run ./scripts gh:env:setup                      # All environments
  go run ./scripts gh:env:setup --env=dev            # Development only
  go run ./scripts gh:env:setup --env=feature-test   # Feature branch env`,
	Run: func(cmd *cobra.Command, args []string) {
		if envFlag != "" {
			if !secretsbuild.IsValidEnvironment(envFlag) {
				fmt.Fprintf(os.Stderr, "Invalid environment: %s\n", envFlag)
				os.Exit(1)
			}
			env := secretsbuild.Environment(envFlag)
			if err := secretsbuild.SetupEnvironment(env); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			// Also generate deploy.yml
			envs := secretsbuild.ListAvailableEnvironments()
			if err := secretsbuild.CreateGithubDeployYaml(envs); err != nil {
				fmt.Fprintf(os.Stderr, "Error generating deploy.yml: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := secretsbuild.SetupAllEnvironments(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
	},
}

// gh:env:create - Create GitHub Environment
var createEnvCmd = &cobra.Command{
	Use:   "gh:env:create",
	Short: "Create a GitHub Environment",
	Long: `Create a GitHub Environment.

Requires --env flag to specify which environment to create.

Example:
  go run ./scripts gh:env:create --env=staging`,
	Run: func(cmd *cobra.Command, args []string) {
		if envFlag == "" {
			fmt.Fprintln(os.Stderr, "Error: --env flag is required")
			os.Exit(1)
		}
		if !secretsbuild.IsValidEnvironment(envFlag) {
			fmt.Fprintf(os.Stderr, "Invalid environment: %s\n", envFlag)
			os.Exit(1)
		}
		env := secretsbuild.Environment(envFlag)
		if err := secretsbuild.CreateGithubEnvironment(env); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

// gh:secrets:read - Read and display secrets (for debugging)
var readSecretsCmd = &cobra.Command{
	Use:   "gh:secrets:read",
	Short: "Read and display secret keys (values hidden)",
	Run: func(cmd *cobra.Command, args []string) {
		env := secretsbuild.EnvProd
		if envFlag != "" {
			if !secretsbuild.IsValidEnvironment(envFlag) {
				fmt.Fprintf(os.Stderr, "Invalid environment: %s\n", envFlag)
				os.Exit(1)
			}
			env = secretsbuild.Environment(envFlag)
		}
		secretsFile := secretsbuild.GetSecretsFilePath(env)
		fmt.Printf("Reading secrets from: %s\n", secretsFile)
		secrets := secretsbuild.ReadSecretsFile(secretsFile)
		fmt.Printf("Found %d secrets:\n", len(secrets))
		for _, s := range secrets {
			// Show key names only, not values
			fmt.Printf("  - %s\n", s.Key)
		}
	},
}

// gh:env:normalize - Show normalized environment name for a branch
var normalizeEnvCmd = &cobra.Command{
	Use:   "gh:env:normalize [branch-name]",
	Short: "Show normalized environment name for a branch",
	Long: `Convert a branch name to its normalized environment name.

Examples:
  go run ./scripts gh:env:normalize feature/test    # → feature-test
  go run ./scripts gh:env:normalize release/v1.0    # → release-v1-0
  go run ./scripts gh:env:normalize my_branch       # → my-branch`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		normalized := secretsbuild.NormalizeEnvironmentName(args[0])
		fmt.Printf("Branch:      %s\n", args[0])
		fmt.Printf("Environment: %s\n", normalized)
		fmt.Printf("\nTo set up this environment:\n")
		fmt.Printf("  1. Create .github/.secrets.%s\n", normalized)
		fmt.Printf("  2. Create .env.%s\n", normalized)
		fmt.Printf("  3. Run: go run ./scripts gh:env:setup --env=%s\n", normalized)
	},
}

func init() {
	// Add --env flag to relevant commands (accepts any valid environment name)
	updateSecretsCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")
	pushSecretsCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")
	cleanSecretsCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")
	setupEnvCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")
	createEnvCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")
	readSecretsCmd.Flags().StringVar(&envFlag, "env", "", "Environment name (e.g., prod, dev, feature-test)")

	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(listEnvsCmd)
	rootCmd.AddCommand(updateSecretsCmd)
	rootCmd.AddCommand(pushSecretsCmd)
	rootCmd.AddCommand(cleanSecretsCmd)
	rootCmd.AddCommand(setupEnvCmd)
	rootCmd.AddCommand(createEnvCmd)
	rootCmd.AddCommand(readSecretsCmd)
	rootCmd.AddCommand(normalizeEnvCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
