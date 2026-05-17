package setup_test

import (
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/setup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigPathForClient_ClaudeCode(t *testing.T) {
	path := setup.ConfigPathForClient(setup.ClientClaudeCode)
	home, _ := os.UserHomeDir()

	assert.True(t, strings.HasSuffix(path, ".claude.json"), "should end in .claude.json, got: %s", path)
	assert.True(t, strings.HasPrefix(path, home), "should be in home dir")
}

func TestConfigPathForClient_OpenCode(t *testing.T) {
	path := setup.ConfigPathForClient(setup.ClientOpenCode)

	assert.NotEmpty(t, path)
	assert.True(t, strings.HasSuffix(path, "opencode.json"), "should end in opencode.json, got: %s", path)

	if runtime.GOOS == "windows" {
		assert.True(t, strings.Contains(path, "opencode"), "windows path should contain opencode")
	} else {
		assert.True(t, strings.Contains(path, ".config") || strings.Contains(path, "opencode"),
			"unix path should contain .config or opencode dir")
	}
}

func TestClientInfoFor(t *testing.T) {
	tests := []struct {
		client       setup.Client
		wantName     string
		wantPathSuffix string
	}{
		{setup.ClientOpenCode, "OpenCode", "opencode.json"},
		{setup.ClientClaudeCode, "Claude Code", ".claude.json"},
	}

	for _, tt := range tests {
		t.Run(string(tt.client), func(t *testing.T) {
			info := setup.ClientInfoFor(tt.client)
			assert.Equal(t, tt.wantName, info.Name)
			assert.True(t, strings.HasSuffix(info.ConfigPath, tt.wantPathSuffix),
				"expected suffix %q, got %q", tt.wantPathSuffix, info.ConfigPath)
		})
	}
}

func TestDetectBinaryPath(t *testing.T) {
	path, err := setup.DetectBinaryPath()
	require.NoError(t, err)
	assert.NotEmpty(t, path)
}

func TestBinaryExists_NonExistent(t *testing.T) {
	assert.False(t, setup.BinaryExists("/nonexistent/path/binary"))
}

func TestBinaryExists_Directory(t *testing.T) {
	assert.False(t, setup.BinaryExists(t.TempDir()))
}

func TestBinaryExists_RealFile(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "testbin")
	require.NoError(t, err)
	f.Close()

	if runtime.GOOS != "windows" {
		os.Chmod(f.Name(), 0755)
	}

	assert.True(t, setup.BinaryExists(f.Name()))
}
