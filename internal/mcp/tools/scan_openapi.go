package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/storage"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// ScanOpenAPIHandler handles the postman.scan.openapi tool.
func ScanOpenAPIHandler(parser discovery.Parser, store storage.Store) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		source := req.GetString("source", "")
		if source == "" {
			return mcpgo.NewToolResultError("source is required"), nil
		}

		spec, err := parser.Parse(source)
		if err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("failed to parse spec: %v", err)), nil
		}

		if store != nil {
			_ = store.SaveScan(source, len(spec.Endpoints))
		}

		result := map[string]interface{}{
			"title":       spec.Title,
			"version":     spec.Version,
			"description": spec.Description,
			"base_url":    spec.BaseURL,
			"count":       len(spec.Endpoints),
			"endpoints":   endpointSummaries(spec.Endpoints),
		}

		data, err := json.Marshal(result)
		if err != nil {
			return mcpgo.NewToolResultError("failed to serialize result"), nil
		}
		return mcpgo.NewToolResultText(string(data)), nil
	}
}

type endpointSummary struct {
	Method      string   `json:"method"`
	Path        string   `json:"path"`
	OperationID string   `json:"operationId,omitempty"`
	Summary     string   `json:"summary,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

func endpointSummaries(endpoints []discovery.Endpoint) []endpointSummary {
	summaries := make([]endpointSummary, len(endpoints))
	for i, ep := range endpoints {
		summaries[i] = endpointSummary{
			Method:      ep.Method,
			Path:        ep.Path,
			OperationID: ep.OperationID,
			Summary:     ep.Summary,
			Tags:        ep.Tags,
		}
	}
	return summaries
}
