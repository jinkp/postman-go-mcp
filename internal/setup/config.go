package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServerEntry is the JSON structure for a single MCP server entry.
type MCPServerEntry struct {
	Command string   `json:"command"`
	Args    []string `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// MergeResult holds the outcome of a config write operation.
type MergeResult struct {
	ConfigPath string
	BackupPath string
	IsNew      bool
	JSON       string // the JSON block that was written
}

// MergeConfig reads the existing config at configPath (or starts fresh),
// merges the new MCP server entry, and writes atomically.
// It creates the directory if it doesn't exist.
// It backs up the original file to configPath + ".bak" before writing.
func MergeConfig(configPath, serverName, binaryPath string) (*MergeResult, error) {
	entry := MCPServerEntry{Command: binaryPath}

	// Read existing config or start fresh
	raw, err := os.ReadFile(configPath)
	isNew := false
	if err != nil {
		if os.IsNotExist(err) {
			raw = []byte("{}")
			isNew = true
		} else {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	// Parse into generic map to preserve unknown fields
	var cfg map[string]interface{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("parse config (may contain comments or be invalid JSON): %w", err)
	}

	// Ensure mcpServers key exists
	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = map[string]interface{}{}
	}

	// Set our server entry
	mcpServers[serverName] = entry
	cfg["mcpServers"] = mcpServers

	// Marshal result
	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("serialize config: %w", err)
	}

	result := &MergeResult{
		ConfigPath: configPath,
		IsNew:      isNew,
		JSON:       string(pretty),
	}

	// Create directory if needed
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create config directory: %w", err)
	}

	// Backup existing file
	if !isNew {
		backupPath := configPath + ".bak"
		if err := copyFile(configPath, backupPath); err != nil {
			return nil, fmt.Errorf("backup config: %w", err)
		}
		result.BackupPath = backupPath
	}

	// Write atomically: write to .tmp then rename
	tmpPath := configPath + ".tmp"
	if err := os.WriteFile(tmpPath, pretty, 0644); err != nil {
		return nil, fmt.Errorf("write temp config: %w", err)
	}
	if err := os.Rename(tmpPath, configPath); err != nil {
		os.Remove(tmpPath) // cleanup on failure
		return nil, fmt.Errorf("atomic rename: %w", err)
	}

	return result, nil
}

// PreviewJSON returns the JSON that would be written without modifying any files.
func PreviewJSON(configPath, serverName, binaryPath string) (string, error) {
	entry := MCPServerEntry{Command: binaryPath}

	raw, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			raw = []byte("{}")
		} else {
			return "", fmt.Errorf("read config: %w", err)
		}
	}

	var cfg map[string]interface{}
	if err := json.Unmarshal(raw, &cfg); err != nil {
		// Config has syntax issues — show a fresh minimal config
		cfg = map[string]interface{}{}
	}

	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	if !ok {
		mcpServers = map[string]interface{}{}
	}
	mcpServers[serverName] = entry
	cfg["mcpServers"] = mcpServers

	pretty, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(pretty), nil
}

func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}
