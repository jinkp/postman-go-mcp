package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/postman"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// CollectionGenerateHandler handles the postman.collection.generate tool.
func CollectionGenerateHandler(parser discovery.Parser, builder postman.CollectionBuilder) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		source := req.GetString("spec_source", "")
		if source == "" {
			return mcpgo.NewToolResultError("spec_source is required"), nil
		}

		spec, err := parser.Parse(source)
		if err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("failed to parse spec: %v", err)), nil
		}

		opts := postman.BuildOptions{
			GroupByTags:    req.GetBool("group_by_tags", true),
			IncludeAuth:    req.GetBool("include_auth", true),
			CollectionName: req.GetString("collection_name", ""),
			BaseURL:        req.GetString("base_url", ""),
		}

		col, err := builder.Build(spec, opts)
		if err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("failed to build collection: %v", err)), nil
		}

		data, err := json.Marshal(map[string]interface{}{
			"collection":     col,
			"endpoint_count": col.CountEndpoints(),
		})
		if err != nil {
			return mcpgo.NewToolResultError("failed to serialize collection"), nil
		}
		return mcpgo.NewToolResultText(string(data)), nil
	}
}
