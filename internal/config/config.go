// Package config loads application configuration from environment variables.
package config

import (
	"os"
	"strconv"
)

// Config holds all runtime configuration for the MCP server.
type Config struct {
	// MCPTransport is the MCP server transport: "stdio" or "sse".
	MCPTransport string
	// MCPHTTPPort is the HTTP port for SSE transport.
	MCPHTTPPort string
	// DBPath is the SQLite database file path.
	DBPath string
	// LogLevel is the log level: debug | info | warn | error.
	LogLevel string
	// DryRun when true prevents any file writes to disk.
	DryRun bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		MCPTransport: getEnv("MCP_TRANSPORT", "stdio"),
		MCPHTTPPort:  getEnv("MCP_HTTP_PORT", "8080"),
		DBPath:       getEnv("DB_PATH", "./data/postman.db"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		DryRun:       getBoolEnv("DRY_RUN", false),
	}
}

func getEnv(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return defaultValue
	}
	return b
}
