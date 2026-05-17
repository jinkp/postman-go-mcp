package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/security"
	"github.com/isai-salazar-enc/postman-go-mcp/pkg/postmanfmt"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
)

// CollectionExportHandler handles the postman.collection.export tool.
func CollectionExportHandler(dryRun bool) func(context.Context, mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
	return func(ctx context.Context, req mcpgo.CallToolRequest) (*mcpgo.CallToolResult, error) {
		rawArgs := req.GetArguments()

		rawCol, ok := rawArgs["collection"]
		if !ok || rawCol == nil {
			return mcpgo.NewToolResultError("collection is required"), nil
		}

		outputPath := req.GetString("output_path", "")
		if outputPath == "" {
			return mcpgo.NewToolResultError("output_path is required"), nil
		}

		if err := security.ValidatePath(outputPath); err != nil {
			return mcpgo.NewToolResultError(err.Error()), nil
		}

		dir := filepath.Dir(outputPath)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			return mcpgo.NewToolResultError(fmt.Sprintf("directory %q does not exist", dir)), nil
		}

		colBytes, err := json.Marshal(rawCol)
		if err != nil {
			return mcpgo.NewToolResultError("invalid collection object"), nil
		}

		var col postmanfmt.Collection
		if err := json.Unmarshal(colBytes, &col); err != nil {
			return mcpgo.NewToolResultError("failed to decode collection"), nil
		}

		endpointCount := col.CountEndpoints()

		if security.IsDryRun(dryRun) {
			result, _ := json.Marshal(map[string]interface{}{
				"success":        true,
				"path":           outputPath,
				"endpoint_count": endpointCount,
				"dry_run":        true,
				"message":        "DRY_RUN=true: file not written",
			})
			return mcpgo.NewToolResultText(string(result)), nil
		}

		pretty, err := json.MarshalIndent(col, "", "  ")
		if err != nil {
			return mcpgo.NewToolResultError("failed to serialize collection"), nil
		}

		if err := os.WriteFile(outputPath, pretty, 0644); err != nil {
			return mcpgo.NewToolResultError(fmt.Sprintf("failed to write file: %v", err)), nil
		}

		result, _ := json.Marshal(map[string]interface{}{
			"success":        true,
			"path":           outputPath,
			"endpoint_count": endpointCount,
		})
		return mcpgo.NewToolResultText(string(result)), nil
	}
}
