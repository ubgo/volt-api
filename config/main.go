package config

import (
	"config/gen/rootconfig"
	"context"
	"fmt"
	"os"
	"path/filepath"
)

func Get() *rootconfig.RootConfig {
	configDir, err := findDirectoryUp("config")
	if err != nil {
		panic(err)
	}

	// Allow overriding config environment via CONFIG_ENV (default, prod, etc.)
	configEnv := os.Getenv("CONFIG_ENV")
	if configEnv == "" {
		configEnv = "default"
	}

	configPath := filepath.Join(configDir, "pkl", "env", configEnv, "app.pkl")
	cfg, err := rootconfig.LoadFromPath(context.Background(), configPath)
	if err != nil {
		panic(err)
	}
	return &cfg
}

// findDirectoryUp searches for a directory up to 3 levels deep from the current working directory
// Returns the full path to the directory if found, or an error if not found
func findDirectoryUp(dirName string) (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	currentPath := wd
	for level := 0; level < 3; level++ {
		checkPath := filepath.Join(currentPath, dirName)
		info, err := os.Stat(checkPath)
		if err == nil && info.IsDir() {
			return checkPath, nil
		}

		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			break
		}
		currentPath = parentPath
	}

	return "", fmt.Errorf("directory '%s' not found within 3 levels from %s", dirName, wd)
}
