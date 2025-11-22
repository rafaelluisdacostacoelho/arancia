package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	Port           int
	StorageBackend string
	BoltDBPath     string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		Port:           8080,
		StorageBackend: "memory",
		BoltDBPath:     "/data/todos.db",
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err != nil {
			return nil, fmt.Errorf("invalid PORT: %w", err)
		}
		cfg.Port = port
	}

	if backend := os.Getenv("STORAGE_BACKEND"); backend != "" {
		if backend != "memory" && backend != "boltdb" {
			return nil, fmt.Errorf("invalid STORAGE_BACKEND: must be 'memory' or 'boltdb'")
		}
		cfg.StorageBackend = backend
	}

	if path := os.Getenv("BOLTDB_PATH"); path != "" {
		cfg.BoltDBPath = path
	}

	return cfg, nil
}

