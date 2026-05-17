package tools

import (
	"context"
	"encoding/json"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/audit"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// AuditRunHandler handles the postman.audit.run tool.
func AuditRunHandler(auditor audit.Auditor) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
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

		report, err := auditor.Run(&col)
		if err != nil {
			return mcpgo.NewToolResultError("failed to run audit: " + err.Error()), nil
		}

		data, _ := json.Marshal(report)
		return mcpgo.NewToolResultText(string(data)), nil
	}
}
