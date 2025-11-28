package load

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

func Env(optionalWorkingDir ...string) error {
	envPath, err := findEnvFile(optionalWorkingDir...)
	if err != nil {
		return err
	}
	return loadEnvFile(envPath)
}

func findEnvFile(optionalWorkingDir ...string) (string, error) {
	var startDir string

	if len(optionalWorkingDir) > 0 {
		startDir = optionalWorkingDir[0]
	} else {
		wd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		startDir = wd
	}

	dir := startDir
	seenPaths := []string{}
	for {
		seenPaths = append(seenPaths, dir)
		envPath := filepath.Join(dir, ".env")
		slog.Debug("Checking", "path", envPath)
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}

		// Check if we've hit the project root (go.mod exists)
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("failed to find .env file in: %v", seenPaths) // No .env file found
}

func loadEnvFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open .env file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue // Skip malformed lines
		}

		key := strings.TrimSpace(parts[0])
		if key == "" {
			continue
		}
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		os.Setenv(key, value)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("failed to read .env file: %w", err)
	}

	return nil
}
