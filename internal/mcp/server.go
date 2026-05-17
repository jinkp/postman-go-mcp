// Package mcp is the MCP delivery layer. It registers all tools and wires domain dependencies.
package mcp

import (
	"github.com/isai-salazar-enc/postman-go-mcp/internal/audit"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/config"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/docs"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/environments"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/postman"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/storage"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/tests"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Dependencies holds all domain interfaces needed by the MCP layer.
type Dependencies struct {
	Config      *config.Config
	Parser      discovery.Parser
	Builder     postman.CollectionBuilder
	DocGen      docs.DocGenerator
	TestGen     tests.TestGenerator
	EnvGen      environments.EnvironmentGenerator
	Auditor     audit.Auditor
	Store       storage.Store
}

// NewServer creates and configures the MCP server with all tools registered.
func NewServer(deps Dependencies) *server.MCPServer {
	s := server.NewMCPServer(
		"postman-go-mcp",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	registerScanOpenAPI(s, deps)
	registerCollectionGenerate(s, deps)
	registerCollectionExport(s, deps)
	registerDocsGenerate(s, deps)
	registerDocsAudit(s, deps)
	registerTestsGenerate(s, deps)
	registerEnvironmentGenerate(s, deps)
	registerAuditRun(s, deps)

	return s
}

// toolError creates a CallToolResult with isError=true and a plain-text message.
func toolError(msg string) (*mcpgo.CallToolResult, error) {
	return mcpgo.NewToolResultError(msg), nil
}

// toolText creates a successful CallToolResult with a text content.
func toolText(text string) (*mcpgo.CallToolResult, error) {
	return mcpgo.NewToolResultText(text), nil
}
