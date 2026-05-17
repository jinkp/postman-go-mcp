// Package setup provides logic for the setup wizard: path detection and config writing.
package setup

import (
	"os"
	"path/filepath"
	"runtime"
)

// Client identifies the MCP client to configure.
type Client string

const (
	ClientOpenCode   Client = "opencode"
	ClientClaudeCode Client = "claudecode"
)

// ClientInfo holds display name and config path for a client.
type ClientInfo struct {
	Name       string
	ConfigPath string
}

// DetectBinaryPath returns the absolute path of the running executable.
func DetectBinaryPath() (string, error) {
	return os.Executable()
}

// ConfigPathForClient returns the expected config file path for each client on the current OS.
func ConfigPathForClient(client Client) string {
	home, _ := os.UserHomeDir()

	switch client {
	case ClientOpenCode:
		switch runtime.GOOS {
		case "windows":
			appData := os.Getenv("APPDATA")
			if appData == "" {
				appData = filepath.Join(home, "AppData", "Roaming")
			}
			return filepath.Join(appData, "opencode", "opencode.json")
		default: // linux, darwin
			configDir := os.Getenv("XDG_CONFIG_HOME")
			if configDir == "" {
				configDir = filepath.Join(home, ".config")
			}
			return filepath.Join(configDir, "opencode", "opencode.json")
		}

	case ClientClaudeCode:
		// Claude Code uses ~/.claude.json on all platforms
		return filepath.Join(home, ".claude.json")
	}

	return ""
}

// ClientInfoFor returns display metadata for a client.
func ClientInfoFor(client Client) ClientInfo {
	switch client {
	case ClientOpenCode:
		return ClientInfo{
			Name:       "OpenCode",
			ConfigPath: ConfigPathForClient(ClientOpenCode),
		}
	case ClientClaudeCode:
		return ClientInfo{
			Name:       "Claude Code",
			ConfigPath: ConfigPathForClient(ClientClaudeCode),
		}
	}
	return ClientInfo{}
}

// BinaryExists reports whether the file at path exists and is executable.
func BinaryExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	if info.IsDir() {
		return false
	}
	// On Windows all files are "executable" if they exist
	if runtime.GOOS == "windows" {
		return true
	}
	return info.Mode()&0111 != 0
}
