package tools

import (
	"context"
	"encoding/json"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/environments"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// EnvironmentGenerateHandler handles the postman.environment.generate tool.
func EnvironmentGenerateHandler(gen environments.EnvironmentGenerator) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		baseURL := req.GetString("base_url", "")
		if baseURL == "" {
			return mcpgo.NewToolResultError("base_url is required"), nil
		}

		rawArgs := req.GetArguments()

		var envNames []string
		if raw, ok := rawArgs["environments"].([]interface{}); ok {
			for _, e := range raw {
				if s, ok := e.(string); ok {
					envNames = append(envNames, s)
				}
			}
		}

		extras := map[string]string{}
		if raw, ok := rawArgs["extra_variables"].(map[string]interface{}); ok {
			for k, v := range raw {
				if s, ok := v.(string); ok {
					extras[k] = s
				}
			}
		}

		envs, err := gen.Generate(baseURL, envNames, extras)
		if err != nil {
			return mcpgo.NewToolResultError("failed to generate environments: " + err.Error()), nil
		}

		data, _ := json.Marshal(map[string]interface{}{"environments": envs})
		return mcpgo.NewToolResultText(string(data)), nil
	}
}
