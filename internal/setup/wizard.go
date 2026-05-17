package setup

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("205")).
			MarginBottom(1)

	successStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("82"))

	errorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("196"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240"))

	codeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("214")).
			Background(lipgloss.Color("235")).
			Padding(0, 1)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Padding(1, 2).
			MarginTop(1)
)

// WizardInput holds all collected values from the wizard form.
type WizardInput struct {
	Clients    []Client // "opencode", "claudecode", or both
	ServerName string
	BinaryPath string
	Confirmed  bool
}

// RunDirect performs a non-interactive setup for the specified client.
// Used by CLI shortcuts like: mcp-postman setup opencode
func RunDirect(client Client) error {
	binaryPath, err := DetectBinaryPath()
	if err != nil {
		return fmt.Errorf("detect binary path: %w", err)
	}

	info := ClientInfoFor(client)
	serverName := "postman"

	fmt.Println()
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Configuring %s...", info.Name)))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Config: %s", info.ConfigPath)))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Binary: %s", binaryPath)))
	fmt.Println()

	result, err := MergeConfig(info.ConfigPath, serverName, binaryPath, ForClient(client))
	if err != nil {
		fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ %s — %v", info.Name, err)))
		return fmt.Errorf("setup %s: %w", info.Name, err)
	}

	line := fmt.Sprintf("  ✔ %s configured", info.Name)
	if !result.IsNew && result.BackupPath != "" {
		line += dimStyle.Render(fmt.Sprintf(" (backup: %s)", result.BackupPath))
	}
	fmt.Println(successStyle.Render(line))
	fmt.Println()
	fmt.Println(successStyle.Render("Restart your AI assistant to activate the MCP server."))
	return nil
}

// Run launches the interactive setup wizard and applies the configuration.
func Run() error {
	printBanner()

	// Step 1-3: collect inputs
	input, err := collectInputs()
	if err != nil {
		if err == huh.ErrUserAborted {
			fmt.Println(dimStyle.Render("\nSetup cancelled."))
			return nil
		}
		return fmt.Errorf("wizard: %w", err)
	}

	if !input.Confirmed {
		fmt.Println(dimStyle.Render("\nSetup cancelled."))
		return nil
	}

	// Step 5: apply config for each selected client
	fmt.Println()
	var hasError bool
	for _, client := range input.Clients {
		info := ClientInfoFor(client)
		result, err := MergeConfig(info.ConfigPath, input.ServerName, input.BinaryPath, ForClient(client))
		if err != nil {
			fmt.Println(errorStyle.Render(fmt.Sprintf("  ✗ %s — %v", info.Name, err)))
			hasError = true
			continue
		}

		line := fmt.Sprintf("  ✔ %s configured", info.Name)
		if !result.IsNew && result.BackupPath != "" {
			line += dimStyle.Render(fmt.Sprintf(" (backup: %s)", result.BackupPath))
		}
		fmt.Println(successStyle.Render(line))
		fmt.Println(dimStyle.Render(fmt.Sprintf("    Config: %s", result.ConfigPath)))
	}

	fmt.Println()
	if hasError {
		fmt.Println(errorStyle.Render("Some clients could not be configured. Check the errors above."))
		return fmt.Errorf("setup completed with errors")
	}

	fmt.Println(successStyle.Render("Setup complete! Restart your AI assistant to activate the MCP server."))
	fmt.Println(dimStyle.Render(fmt.Sprintf("\nServer name: %q  •  Binary: %s", input.ServerName, input.BinaryPath)))
	return nil
}

// collectInputs runs the huh form wizard and returns the collected inputs.
func collectInputs() (*WizardInput, error) {
	binaryPath, _ := DetectBinaryPath()

	// Form state
	var clientChoice string = "both"
	var serverName string = "postman"
	var binaryInput string = binaryPath
	var confirmed bool

	// Build preview for the confirm step (dynamic)
	previewFor := func() string {
		clients := resolveClients(clientChoice)
		if len(clients) == 0 {
			return ""
		}
		var sb strings.Builder
		for _, c := range clients {
			info := ClientInfoFor(c)
			preview, _ := PreviewJSON(info.ConfigPath, serverName, binaryInput, c)
			sb.WriteString(dimStyle.Render(fmt.Sprintf("%s  →  %s\n\n", info.Name, info.ConfigPath)))
			sb.WriteString(codeStyle.Render(truncate(preview, 400)))
			sb.WriteString("\n\n")
		}
		return boxStyle.Render(sb.String())
	}

	form := huh.NewForm(
		// Group 1: client + server name + binary path
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Which AI assistant do you want to configure?").
				Description("The MCP server will be added to the selected client's config.").
				Options(
					huh.NewOption("OpenCode", "opencode"),
					huh.NewOption("Claude Code", "claudecode"),
					huh.NewOption("Both", "both"),
				).
				Value(&clientChoice),

			huh.NewInput().
				Title("MCP server name").
				Description("The name used to identify this server in your config (e.g. \"postman\").").
				Placeholder("postman").
				Value(&serverName).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("server name cannot be empty")
					}
					for _, c := range s {
						if !isValidNameChar(c) {
							return fmt.Errorf("only letters, numbers, hyphens and underscores allowed")
						}
					}
					return nil
				}),

			huh.NewInput().
				Title("Path to mcp-postman binary").
				Description("Auto-detected from the running executable. Edit if needed.").
				Value(&binaryInput).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("binary path cannot be empty")
					}
					if !BinaryExists(s) {
						return fmt.Errorf("file not found or not executable: %s", s)
					}
					return nil
				}),
		),

		// Group 2: confirm
		huh.NewGroup(
			huh.NewNote().
				Title("Preview — what will be written").
				DescriptionFunc(previewFor, &struct{ a, b, c *string }{&clientChoice, &serverName, &binaryInput}),

			huh.NewConfirm().
				Title("Apply this configuration?").
				Description("Your existing config will be backed up before any changes.").
				Affirmative("Yes, configure it").
				Negative("No, cancel").
				Value(&confirmed),
		),
	).WithTheme(huh.ThemeCatppuccin())

	if err := form.Run(); err != nil {
		return nil, err
	}

	return &WizardInput{
		Clients:    resolveClients(clientChoice),
		ServerName: strings.TrimSpace(serverName),
		BinaryPath: strings.TrimSpace(binaryInput),
		Confirmed:  confirmed,
	}, nil
}

func resolveClients(choice string) []Client {
	switch choice {
	case "opencode":
		return []Client{ClientOpenCode}
	case "claudecode":
		return []Client{ClientClaudeCode}
	case "both":
		return []Client{ClientOpenCode, ClientClaudeCode}
	}
	return nil
}

func isValidNameChar(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_'
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}

func printBanner() {
	banner := titleStyle.Render("  postman-go-mcp  ·  Setup Wizard")
	subtitle := dimStyle.Render("  Configure the MCP server for OpenCode and/or Claude Code\n")
	fmt.Println()
	fmt.Println(banner)
	fmt.Println(subtitle)

	// Show detected config paths
	opencodePath := ConfigPathForClient(ClientOpenCode)
	claudePath := ConfigPathForClient(ClientClaudeCode)
	fmt.Println(dimStyle.Render(fmt.Sprintf("  OpenCode config:    %s", opencodePath)))
	fmt.Println(dimStyle.Render(fmt.Sprintf("  Claude Code config: %s", claudePath)))
	fmt.Println()

	// Show warning if a config has potential JSON issues
	for _, path := range []string{opencodePath, claudePath} {
		if data, err := os.ReadFile(path); err == nil {
			if looksLikeJSONC(string(data)) {
				fmt.Println(errorStyle.Render(fmt.Sprintf(
					"  ⚠ %s appears to contain comments (JSONC). The wizard may fail to merge it.\n  Consider removing comments before running setup.",
					path,
				)))
			}
		}
	}
}

// looksLikeJSONC returns true if the content appears to contain // or /* comments.
func looksLikeJSONC(content string) bool {
	return strings.Contains(content, "//") || strings.Contains(content, "/*")
}
