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

func TestMergeConfig_NewFile(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman")
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

func TestMergeConfig_ExistingFile_Merges(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	// Write existing config with unrelated field
	existing := `{"theme": "dark", "mcpServers": {"other-tool": {"command": "/bin/other"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0644))

	result, err := setup.MergeConfig(configPath, "postman", "/usr/local/bin/mcp-postman")
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

func TestMergeConfig_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "nested", "deep", "config.json")

	_, err := setup.MergeConfig(configPath, "postman", "/bin/mcp-postman")
	require.NoError(t, err)

	_, err = os.Stat(configPath)
	assert.NoError(t, err, "config file should exist after mkdir")
}

func TestMergeConfig_BackupCreated(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	require.NoError(t, os.WriteFile(configPath, []byte(`{"existing": true}`), 0644))

	result, err := setup.MergeConfig(configPath, "postman", "/bin/mcp-postman")
	require.NoError(t, err)

	_, err = os.Stat(result.BackupPath)
	assert.NoError(t, err, "backup file should exist")

	data, _ := os.ReadFile(result.BackupPath)
	assert.Contains(t, string(data), "existing")
}

func TestPreviewJSON_NoSideEffects(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	preview, err := setup.PreviewJSON(configPath, "postman", "/bin/mcp-postman")
	require.NoError(t, err)
	assert.Contains(t, preview, `"postman"`)
	assert.Contains(t, preview, `"command"`)

	// File should NOT have been created
	_, err = os.Stat(configPath)
	assert.True(t, os.IsNotExist(err), "PreviewJSON should not create files")
}

func TestMergeConfig_OverwritesExistingEntry(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")

	existing := `{"mcpServers": {"postman": {"command": "/old/path"}}}`
	require.NoError(t, os.WriteFile(configPath, []byte(existing), 0644))

	_, err := setup.MergeConfig(configPath, "postman", "/new/path")
	require.NoError(t, err)

	data, _ := os.ReadFile(configPath)
	var cfg map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &cfg))

	mcpServers := cfg["mcpServers"].(map[string]interface{})
	postman := mcpServers["postman"].(map[string]interface{})
	assert.Equal(t, "/new/path", postman["command"])
}
