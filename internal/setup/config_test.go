package setup_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMergeConfig_NewFile_ClaudeCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman", setup.ForClient(setup.ClientClaudeCode))
	require.NoError(t, err)

	assert.True(t, result.IsNew)
	assert.Equal(t, configPath, result.ConfigPath)
	assert.Empty(t, result.BackupPath)

	// Verify written JSON
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	mcpServers, ok := cfg["mcpServers"].(map[string]interface{})
	require.True(t, ok)

	postman, ok := mcpServers["postman"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "/usr/local/bin/mcp-postman", postman["command"])
}

func TestMergeConfig_NewFile_OpenCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman", setup.ForClient(setup.ClientOpenCode))
	require.NoError(t, err)

	assert.True(t, result.IsNew)
	assert.Equal(t, configPath, result.ConfigPath)

	// Verify written JSON uses OpenCode format: mcp.{name} with type: "local"
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	mcpSection, ok := cfg["mcp"].(map[string]interface{})
	require.True(t, ok, "OpenCode config should have 'mcp' key, not 'mcpServers'")

	postman, ok := mcpSection["postman"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "local", postman["type"])
	assert.Equal(t, "/usr/local/bin/mcp-postman", postman["command"])
}

func TestMergeConfig_ExistingFile_Merges_ClaudeCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write existing config with unrelated field
	existing := `{"theme": "dark", "mcpServers": {"other-tool": {"command": "/bin/other"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0644))

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman", setup.ForClient(setup.ClientClaudeCode))
	require.NoError(t, err)

	assert.False(t, result.IsNew)
	assert.NotEmpty(t, result.BackupPath)

	// Verify merge preserved existing fields
	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	// Original field preserved
	assert.Equal(t, "dark", cfg["theme"])

	mcpServers := cfg["mcpServers"].(map[string]interface{})

	// Existing server preserved
	_, hasOther := mcpServers["other-tool"]
	assert.True(t, hasOther, "existing server should be preserved")

	// New server added
	postman, ok := mcpServers["postman"].(map[string]interface{})
	require.True(t, ok, "new server should be added")
	assert.Equal(t, "/usr/local/bin/mcp-postman", postman["command"])
}

func TestMergeConfig_ExistingFile_Merges_OpenCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write existing config with unrelated fields and an existing mcp entry
	existing := `{"theme": "dark", "mcp": {"other-tool": {"type": "local", "command": "/bin/other"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0644))

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman", setup.ForClient(setup.ClientOpenCode))
	require.NoError(t, err)

	assert.False(t, result.IsNew)
	assert.NotEmpty(t, result.BackupPath)

	data, err := os.ReadFile(configPath)
	require.NoError(t, err)

	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	// Original field preserved
	assert.Equal(t, "dark", cfg["theme"])

	mcpSection := cfg["mcp"].(map[string]interface{})

	// Existing server preserved
	_, hasOther := mcpSection["other-tool"]
	assert.True(t, hasOther, "existing server should be preserved")

	// New server added with correct format
	postman, ok := mcpSection["postman"].(map[string]interface{})
	require.True(t, ok, "new server should be added")
	assert.Equal(t, "local", postman["type"])
	assert.Equal(t, "/usr/local/bin/mcp-postman", postman["command"])
}

func TestMergeConfig_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nested", "deep", "config.json")

	_, err := setup.MergeConfig(configPath, "postman", "/bin/mcp-postman", setup.ForClient(setup.ClientClaudeCode))
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config file should exist after mkdir")
}

func TestMergeConfig_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"existing": true}`), 0644))

	result, err := setup.MergeConfig(configPath, "postman", "/bin/mcp-postman", setup.ForClient(setup.ClientClaudeCode))
	require.NoError(t, err)

	_, err = os.Stat(result.BackupPath)
	assert.NoError(t, err, "backup file should exist")

	data, _ := os.ReadFile(result.BackupPath)
	assert.Contains(t, string(data), "existing")
}

func TestPreviewJSON_NoSideEffects_ClaudeCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	preview, err := setup.PreviewJSON(configPath, "postman", "/bin/mcp-postman", setup.ClientClaudeCode)
	require.NoError(t, err)
	assert.Contains(t, preview, `"postman"`)
	assert.Contains(t, preview, `"command"`)
	assert.Contains(t, preview, `"mcpServers"`)

	// File should NOT have been created
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err), "PreviewJSON should not create files")
}

func TestPreviewJSON_NoSideEffects_OpenCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	preview, err := setup.PreviewJSON(configPath, "postman", "/bin/mcp-postman", setup.ClientOpenCode)
	require.NoError(t, err)
	assert.Contains(t, preview, `"postman"`)
	assert.Contains(t, preview, `"command"`)
	assert.Contains(t, preview, `"mcp"`)
	assert.Contains(t, preview, `"type"`)
	assert.Contains(t, preview, `"local"`)
	assert.NotContains(t, preview, `"mcpServers"`)

	// File should NOT have been created
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err), "PreviewJSON should not create files")
}

func TestMergeConfig_OverwritesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	existing := `{"mcpServers": {"postman": {"command": "/old/path"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0644))

	_, err := setup.MergeConfig(configPath, "postman", "/new/path", setup.ForClient(setup.ClientClaudeCode))
	require.NoError(t, err)

	data, _ := os.ReadFile(configPath)
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	mcpServers := cfg["mcpServers"].(map[string]interface{})
	postman := mcpServers["postman"].(map[string]interface{})
	assert.Equal(t, "/new/path", postman["command"])
}

func TestMergeConfig_DefaultClient_IsClaudeCode(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// No ForClient option — should default to Claude Code format
	_, err := setup.MergeConfig(configPath, "postman", "/bin/mcp-postman")
	require.NoError(t, err)

	data, _ := os.ReadFile(configPath)
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	_, hasMcpServers := cfg["mcpServers"]
	assert.True(t, hasMcpServers, "default should use mcpServers (Claude Code format)")
}
