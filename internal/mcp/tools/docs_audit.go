package tools

import (
	"context"
	"encoding/json"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/docs"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// DocsAuditHandler handles the postman.docs.audit tool.
func DocsAuditHandler(gen docs.DocGenerator) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
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

		report, err := gen.Audit(&col)
		if err != nil {
			return mcpgo.NewToolResultError("failed to audit docs: " + err.Error()), nil
		}

		data, _ := json.Marshal(report)
		return mcpgo.NewToolResultText(string(data)), nil
	}
}
