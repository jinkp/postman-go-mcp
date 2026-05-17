package setup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// MCPServerEntry is the JSON structure for a Claude Code MCP server entry.
type MCPServerEntry struct {
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// OpenCodeMCPEntry is the JSON structure for an OpenCode MCP server entry.
// OpenCode uses "mcp" (not "mcpServers") with a "type" field.
type OpenCodeMCPEntry struct {
	Type    string            `json:"type"`
	Command string            `json:"command"`
	Args    []string          `json:"args,omitempty"`
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
// merges the new MCP server entry using the correct format for the given client,
// and writes atomically.
// It creates the directory if it doesn't exist.
// It backs up the original file to configPath + ".bak" before writing.
func MergeConfig(configPath, serverName, binaryPath string, opts ...MergeOption) (*MergeResult, error) {
	o := defaultMergeOptions()
	for _, opt := range opts {
		opt(&o)
	}

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

	// Merge entry using the correct format per client
	mergeEntry(cfg, o.client, serverName, binaryPath)

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

// MergeOption configures MergeConfig behavior.
type MergeOption func(*mergeOptions)

type mergeOptions struct {
	client Client
}

func defaultMergeOptions() mergeOptions {
	return mergeOptions{client: ClientClaudeCode}
}

// ForClient sets which client format to use when merging the config entry.
func ForClient(c Client) MergeOption {
	return func(o *mergeOptions) {
		o.client = c
	}
}

// mergeEntry inserts the MCP server entry using the correct JSON structure per client.
//
// OpenCode format:   { "mcp": { "<name>": { "type": "local", "command": "...", "args": [...] } } }
// Claude Code format: { "mcpServers": { "<name>": { "command": "..." } } }
func mergeEntry(cfg map[string]interface{}, client Client, serverName, binaryPath string) {
	switch client {
	case ClientOpenCode:
		section, ok := cfg["mcp"].(map[string]interface{})
		if !ok {
			section = map[string]interface{}{}
		}
		section[serverName] = OpenCodeMCPEntry{
			Type:    "local",
			Command: binaryPath,
		}
		cfg["mcp"] = section

	default: // ClientClaudeCode
		section, ok := cfg["mcpServers"].(map[string]interface{})
		if !ok {
			section = map[string]interface{}{}
		}
		section[serverName] = MCPServerEntry{
			Command: binaryPath,
		}
		cfg["mcpServers"] = section
	}
}

// PreviewJSON returns the JSON that would be written without modifying any files.
func PreviewJSON(configPath, serverName, binaryPath string, client Client) (string, error) {
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

	mergeEntry(cfg, client, serverName, binaryPath)

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
