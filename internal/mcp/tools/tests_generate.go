package tools

import (
	"context"
	"encoding/json"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/tests"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// TestsGenerateHandler handles the postman.tests.generate tool.
func TestsGenerateHandler(gen tests.TestGenerator) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		rawArgs := req.GetArguments()
		rawCol, ok := rawArgs["collection"]
		if !ok || rawCol == nil {
			return mcpgo.NewToolResultError("collection is required"), nil
		}

		colBytes, _ := json.Marshal(rawCol)
		var col postmanfmt.Collection
		if err := json.Unmarshal(colBytes, &col); err != nil {
			return mcpgo.NewToolResultError("failed to decode collection"), nil
		}

		var testTypes []tests.TestType
		if raw, ok := rawArgs["test_types"].([]interface{}); ok {
			for _, t := range raw {
				if s, ok := t.(string); ok {
					testTypes = append(testTypes, tests.TestType(s))
				}
			}
		}

		enriched, err := gen.Generate(&col, testTypes)
		if err != nil {
			return mcpgo.NewToolResultError("failed to generate tests: " + err.Error()), nil
		}

		data, _ := json.Marshal(map[string]interface{}{"collection": enriched})
		return mcpgo.NewToolResultText(string(data)), nil
	}
}
