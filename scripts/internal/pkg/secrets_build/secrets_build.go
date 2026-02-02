package secretsbuild

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/joho/godotenv"
)

// Environment represents a deployment environment
type Environment string

// Common environment constants (not restrictive, just helpers)
const (
	EnvProd    Environment = "prod"
	EnvStaging Environment = "staging"
	EnvDev     Environment = "dev"
	EnvUAT     Environment = "uat"
)

// Prefix for GitHub secrets in .env files
const (
	GitHubSecretPrefix   = "GITHUB_SECRET_"
	GitHubSecretKeyFiles = "GITHUB_SECRET_KEY_FILES"
)

// projectRoot caches the discovered project root
var projectRoot string

// findProjectRoot finds the project root by looking for .git or go.mod
func findProjectRoot() (string, error) {
	if projectRoot != "" {
		return projectRoot, nil
	}

	// Start from current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up to find .git or go.mod (but not in scripts/)
	for {
		// Check for .git directory
		if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
			projectRoot = dir
			return projectRoot, nil
		}
		// Check for go.mod (but make sure it's not scripts/go.mod)
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, ".github")); err == nil {
				projectRoot = dir
				return projectRoot, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("could not find project root (no .git directory found)")
		}
		dir = parent
	}
}

// getKeyFilesDir returns the path to the _keys directory
func getKeyFilesDir() string {
	root, _ := findProjectRoot()
	return filepath.Join(root, "_keys")
}

// getSecretsBaseDir returns the path to .github directory
func getSecretsBaseDir() string {
	root, _ := findProjectRoot()
	return filepath.Join(root, ".github")
}

// getWorkflowsDir returns the path to .github/workflows directory
func getWorkflowsDir() string {
	root, _ := findProjectRoot()
	return filepath.Join(root, ".github", "workflows")
}

// getEnvFilesDir returns the path to env files directory (project root)
func getEnvFilesDir() string {
	root, _ := findProjectRoot()
	return root
}

// GetDeployYamlPath returns the path to the deploy.yml file for an environment
func GetDeployYamlPath(env Environment) string {
	return filepath.Join(getWorkflowsDir(), fmt.Sprintf("deploy.%s.yml", env))
}

// getDeployYamlTemplatePath returns the path to deploy.yml.gotmpl
func getDeployYamlTemplatePath() string {
	return filepath.Join(getWorkflowsDir(), "deploy.yml.gotmpl")
}

// GetBranchName returns the git branch name for an environment
// prod -> main, everything else -> same as env name
func GetBranchName(env Environment) string {
	if env == EnvProd {
		return "main"
	}
	return string(env)
}

// GetGeneratedSecretsFilePath returns the path to the generated secrets file for an environment
func GetGeneratedSecretsFilePath(env Environment) string {
	return filepath.Join(getSecretsBaseDir(), fmt.Sprintf(".secrets.%s", env))
}

// GetEnvFilePath returns the path to the env file for an environment
func GetEnvFilePath(env Environment) string {
	return filepath.Join(getEnvFilesDir(), fmt.Sprintf(".env.%s", env))
}

// NormalizeEnvironmentName converts a branch name to a valid environment name
// e.g., "feature/test" -> "feature-test", "release/v1.0" -> "release-v1-0"
func NormalizeEnvironmentName(name string) Environment {
	normalized := strings.ReplaceAll(name, "/", "-")
	normalized = strings.ReplaceAll(normalized, "_", "-")
	normalized = strings.ReplaceAll(normalized, ".", "-")
	normalized = strings.ToLower(normalized)
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	normalized = strings.Trim(normalized, "-")
	return Environment(normalized)
}

// IsValidEnvironmentName checks if the environment name is valid (non-empty, valid chars)
func IsValidEnvironmentName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

// IsValidEnvironment checks if the given environment name is valid
func IsValidEnvironment(env string) bool {
	return IsValidEnvironmentName(env)
}

// ListAvailableEnvironments returns a list of environments that have .env.{env} files
func ListAvailableEnvironments() []Environment {
	var available []Environment
	seen := make(map[string]bool)

	// Scan for .env.{env} files
	files, err := os.ReadDir(getEnvFilesDir())
	if err != nil {
		return available
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		// Match .env.{env} pattern (not .example)
		if strings.HasPrefix(name, ".env.") && !strings.HasSuffix(name, ".example") {
			env := strings.TrimPrefix(name, ".env.")
			if env != "" && !seen[env] {
				seen[env] = true
				available = append(available, Environment(env))
			}
		}
	}

	sort.Slice(available, func(i, j int) bool {
		return string(available[i]) < string(available[j])
	})

	return available
}

// Secret represents a key-value secret pair
type Secret struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// BuildAndUpdateSecrets builds secrets for a specific environment
// Reads from .env.{env} file, extracts GITHUB_SECRET_* vars, and processes key files
func BuildAndUpdateSecrets(env Environment) error {
	fmt.Printf("Building secrets for environment: %s\n", env)

	envFile := GetEnvFilePath(env)

	// Read the .env.{env} file
	envVars, err := godotenv.Read(envFile)
	if err != nil {
		return fmt.Errorf("reading env file %s: %w", envFile, err)
	}

	var secrets []Secret

	// 1. Extract GITHUB_SECRET_* variables (except KEY_FILES)
	var keyFilesValue string
	for key, value := range envVars {
		if !strings.HasPrefix(key, GitHubSecretPrefix) {
			continue
		}

		// Handle KEY_FILES separately
		if key == GitHubSecretKeyFiles {
			keyFilesValue = value
			continue
		}

		// Remove prefix to get the actual secret name
		secretName := strings.TrimPrefix(key, GitHubSecretPrefix)
		secrets = append(secrets, Secret{
			Key:   secretName,
			Value: value,
		})
	}

	// 2. Process key files from GITHUB_SECRET_KEY_FILES
	if keyFilesValue != "" {
		keyFileSecrets, err := processKeyFiles(keyFilesValue)
		if err != nil {
			return fmt.Errorf("processing key files: %w", err)
		}
		secrets = append(secrets, keyFileSecrets...)
	}

	// 3. Add filtered .env file as ENV_B64 (for deployment)
	// Exclude GITHUB_SECRET_* lines - they're only for GitHub Actions, not the container
	filteredEnv, err := filterEnvFileForDeployment(envFile)
	if err != nil {
		return fmt.Errorf("filtering env file: %w", err)
	}
	secrets = append(secrets, Secret{
		Key:   "ENV_B64",
		Value: base64.StdEncoding.EncodeToString([]byte(filteredEnv)),
	})

	// Write generated secrets file
	generatedFile := GetGeneratedSecretsFilePath(env)
	if err := WriteSecrets(secrets, generatedFile); err != nil {
		return fmt.Errorf("writing generated secrets: %w", err)
	}
	fmt.Printf("Generated: %s (%d secrets)\n", generatedFile, len(secrets))

	return nil
}

// filterEnvFileForDeployment reads the env file and excludes GITHUB_SECRET_* lines
// These are only for GitHub Actions secrets, not for the container's env file
func filterEnvFileForDeployment(envFile string) (string, error) {
	content, err := os.ReadFile(envFile)
	if err != nil {
		return "", err
	}

	var filtered []string
	var inMultiline bool
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments at the start
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			if !inMultiline {
				filtered = append(filtered, line)
			}
			continue
		}

		// Check if this line starts a GITHUB_SECRET_ variable
		if strings.HasPrefix(trimmed, GitHubSecretPrefix) {
			// Check if it's a multiline value (contains opening quote but no closing)
			if strings.Contains(line, `="`) && !strings.HasSuffix(trimmed, `"`) {
				inMultiline = true
			}
			continue // Skip this line
		}

		// If we're in a multiline GITHUB_SECRET_ value, skip until we find closing quote
		if inMultiline {
			if strings.HasSuffix(trimmed, `"`) || strings.HasSuffix(trimmed, `"'`) {
				inMultiline = false
			}
			continue // Skip this line
		}

		// Keep non-GITHUB_SECRET_ lines
		filtered = append(filtered, line)
	}

	return strings.Join(filtered, "\n"), nil
}

// processKeyFiles parses the KEY_FILES value and returns secrets for each file
// Format: JSON array - ["file1.pem", "file2.creds"]
func processKeyFiles(keyFilesValue string) ([]Secret, error) {
	keyFilesValue = strings.TrimSpace(keyFilesValue)
	if keyFilesValue == "" || keyFilesValue == "[]" {
		return nil, nil
	}

	var files []string
	if err := json.Unmarshal([]byte(keyFilesValue), &files); err != nil {
		return nil, fmt.Errorf("invalid KEY_FILES format (expected JSON array like [\"file1.pem\", \"file2.creds\"]): %w", err)
	}

	var secrets []Secret
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" {
			continue
		}

		// Read file from _keys directory
		filePath := filepath.Join(getKeyFilesDir(), file)
		content, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading key file %s: %w", filePath, err)
		}

		// Generate secret name from filename
		// e.g., "githubapp_thewildai_2025.pem" -> "GITHUBAPP_THEWILDAI_2025_KEY_B64"
		baseName := strings.TrimSuffix(file, filepath.Ext(file))
		secretName := strings.ToUpper(baseName)
		secretName = strings.ReplaceAll(secretName, "-", "_")
		secretName = strings.ReplaceAll(secretName, ".", "_")
		secretName += "_KEY_B64"

		secrets = append(secrets, Secret{
			Key:   secretName,
			Value: base64.StdEncoding.EncodeToString(content),
		})

		fmt.Printf("  Key file: %s -> %s\n", file, secretName)
	}

	return secrets, nil
}

// WriteSecrets writes secrets to a file
func WriteSecrets(secrets []Secret, path string) error {
	var buf bytes.Buffer
	buf.WriteString("# Code generated by secrets_build.go, DO NOT EDIT\n\n")

	// Sort for consistent output
	sort.Slice(secrets, func(i, j int) bool {
		return secrets[i].Key < secrets[j].Key
	})

	for _, secret := range secrets {
		fmt.Fprintf(&buf, "%s=\"%s\"\n", secret.Key, secret.Value)
	}
	return os.WriteFile(path, buf.Bytes(), 0644)
}

// ReadSecretsFile reads secrets from a generated file
func ReadSecretsFile(path string) []Secret {
	envVars, err := godotenv.Read(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	var result []Secret
	for key, value := range envVars {
		result = append(result, Secret{Key: key, Value: value})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})
	return result
}

// EnvData holds template data for a single environment
type EnvData struct {
	Name       string
	NameUpper  string
	Branch     string
	EnvVarName string
}

// CreateGithubDeployYaml generates separate deploy.{env}.yml files for each environment
func CreateGithubDeployYaml(envs []Environment) error {
	templateContent, err := os.ReadFile(getDeployYamlTemplatePath())
	if err != nil {
		return fmt.Errorf("reading template file: %w", err)
	}

	tmpl, err := template.New("deploy").Delims("[[", "]]").Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	for _, env := range envs {
		data := EnvData{
			Name:       string(env),
			NameUpper:  strings.ToUpper(string(env)),
			Branch:     GetBranchName(env),
			EnvVarName: "ENV_B64",
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("executing template for %s: %w", env, err)
		}

		outputPath := GetDeployYamlPath(env)
		if err := os.WriteFile(outputPath, buf.Bytes(), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", outputPath, err)
		}

		fmt.Printf("Generated: %s\n", outputPath)
	}

	return nil
}

// CreateGithubDeployYamlForEnv generates deploy.{env}.yml for a single environment
func CreateGithubDeployYamlForEnv(env Environment) error {
	return CreateGithubDeployYaml([]Environment{env})
}

// PushSecretsToGithubEnvironment pushes secrets to a GitHub Environment
func PushSecretsToGithubEnvironment(env Environment) error {
	generatedFile := GetGeneratedSecretsFilePath(env)
	secrets := ReadSecretsFile(generatedFile)

	// Filter out secrets that GitHub doesn't allow (GITHUB_* prefix)
	var pushable []Secret
	var skipped []string
	for _, secret := range secrets {
		if strings.HasPrefix(secret.Key, "GITHUB_") {
			skipped = append(skipped, secret.Key)
			continue
		}
		pushable = append(pushable, secret)
	}

	fmt.Printf("Pushing %d secrets to GitHub Environment: %s\n", len(pushable), env)
	if len(skipped) > 0 {
		fmt.Printf("  (skipping %d reserved: %v)\n", len(skipped), skipped)
	}

	for _, secret := range pushable {
		if err := pushSecretToEnvironment(string(env), secret.Key, secret.Value); err != nil {
			return fmt.Errorf("pushing secret %s: %w", secret.Key, err)
		}
		fmt.Printf("  ✓ %s\n", secret.Key)
	}

	return nil
}

func pushSecretToEnvironment(env, key, value string) error {
	cmd := exec.Command("gh", "secret", "set", key, "--env", env)
	cmd.Stdin = strings.NewReader(value)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}

// ghSecretListItem is used to parse "gh secret list --json name" output.
type ghSecretListItem struct {
	Name string `json:"name"`
}

// ListGithubEnvironmentSecrets returns the names of secrets in a GitHub Environment
func ListGithubEnvironmentSecrets(env Environment) ([]string, error) {
	cmd := exec.Command("gh", "secret", "list", "--env", string(env), "--json", "name")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("gh secret list: %w", err)
	}
	var items []ghSecretListItem
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parse gh secret list output: %w", err)
	}
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names, nil
}

// DeleteGithubEnvironmentSecret removes a secret from a GitHub Environment
func DeleteGithubEnvironmentSecret(env Environment, name string) error {
	cmd := exec.Command("gh", "secret", "delete", name, "--env", string(env))
	cmd.Stdin = strings.NewReader("y\n")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("gh secret delete %s: %w: %s", name, err, string(out))
	}
	return nil
}

// GetExpectedSecretNames returns the list of expected secret names for an environment
func GetExpectedSecretNames(env Environment) []string {
	generatedFile := GetGeneratedSecretsFilePath(env)
	if _, err := os.Stat(generatedFile); os.IsNotExist(err) {
		return nil
	}
	secrets := ReadSecretsFile(generatedFile)
	names := make([]string, 0, len(secrets))
	for _, s := range secrets {
		names = append(names, s.Key)
	}
	return names
}

// CleanNonExistingSecrets removes orphan secrets from a GitHub Environment
func CleanNonExistingSecrets(env Environment) error {
	expected := GetExpectedSecretNames(env)
	allowed := make(map[string]struct{}, len(expected))
	for _, name := range expected {
		allowed[name] = struct{}{}
	}

	current, err := ListGithubEnvironmentSecrets(env)
	if err != nil {
		return fmt.Errorf("listing secrets: %w", err)
	}

	var removed int
	for _, name := range current {
		if _, ok := allowed[name]; ok {
			continue
		}
		if err := DeleteGithubEnvironmentSecret(env, name); err != nil {
			return fmt.Errorf("deleting secret %q: %w", name, err)
		}
		fmt.Printf("Removed secret: %s\n", name)
		removed++
	}

	if removed == 0 {
		fmt.Println("No orphan secrets to remove.")
	} else {
		fmt.Printf("Removed %d orphan secret(s).\n", removed)
	}
	return nil
}

// CreateGithubEnvironment creates a GitHub Environment if it doesn't exist
func CreateGithubEnvironment(env Environment) error {
	cmd := exec.Command("gh", "api", fmt.Sprintf("repos/{owner}/{repo}/environments/%s", env), "-X", "PUT")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("creating environment: %w: %s", err, string(output))
	}
	fmt.Printf("Created/verified GitHub Environment: %s\n", env)
	return nil
}

// SetupEnvironment creates the environment and pushes all secrets
func SetupEnvironment(env Environment) error {
	// 1. Create the GitHub Environment
	if err := CreateGithubEnvironment(env); err != nil {
		return err
	}

	// 2. Build secrets
	if err := BuildAndUpdateSecrets(env); err != nil {
		return err
	}

	// 3. Push secrets
	if err := PushSecretsToGithubEnvironment(env); err != nil {
		return err
	}

	// 4. Generate deploy.yml
	if err := CreateGithubDeployYamlForEnv(env); err != nil {
		return err
	}

	fmt.Printf("\n✓ Environment %s setup complete!\n", env)
	return nil
}

// GetSecretsFilePath returns the path to the secrets file for an environment
// This is an alias for GetGeneratedSecretsFilePath for backward compatibility
func GetSecretsFilePath(env Environment) string {
	return GetGeneratedSecretsFilePath(env)
}

// BuildAndUpdateAllSecrets builds secrets for all available environments
func BuildAndUpdateAllSecrets() error {
	envs := ListAvailableEnvironments()
	if len(envs) == 0 {
		return fmt.Errorf("no environment configurations found")
	}

	fmt.Printf("Found %d environment(s): %v\n", len(envs), envs)

	for _, env := range envs {
		if err := BuildAndUpdateSecrets(env); err != nil {
			return fmt.Errorf("building secrets for %s: %w", env, err)
		}
	}

	// Generate the deploy.yml files
	if err := CreateGithubDeployYaml(envs); err != nil {
		return fmt.Errorf("generating deploy.yml: %w", err)
	}

	return nil
}

// PushSecretsToAllEnvironments pushes secrets to all GitHub Environments
func PushSecretsToAllEnvironments() error {
	envs := ListAvailableEnvironments()
	for _, env := range envs {
		if err := PushSecretsToGithubEnvironment(env); err != nil {
			return err
		}
	}
	return nil
}

// CleanAllEnvironments cleans orphan secrets from all environments
func CleanAllEnvironments() error {
	envs := ListAvailableEnvironments()
	for _, env := range envs {
		fmt.Printf("\nCleaning environment: %s\n", env)
		if err := CleanNonExistingSecrets(env); err != nil {
			return err
		}
	}
	return nil
}

// SetupAllEnvironments sets up all available environments
func SetupAllEnvironments() error {
	envs := ListAvailableEnvironments()
	if len(envs) == 0 {
		return fmt.Errorf("no environment configurations found")
	}

	for _, env := range envs {
		fmt.Printf("\n=== Setting up environment: %s ===\n", env)
		if err := SetupEnvironment(env); err != nil {
			return err
		}
	}

	// Generate deploy.yml
	return CreateGithubDeployYaml(envs)
}
