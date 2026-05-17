package mcp

import (
	"github.com/isai-salazar-enc/postman-go-mcp/internal/mcp/tools"
	mcpgo "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerScanOpenAPI(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.scan.openapi",
		mcpgo.WithDescription("Parse an OpenAPI 3.0 or Swagger 2.0 spec from a file path or HTTP URL. Returns a list of all endpoints with method, path, tags, parameters, and response codes."),
		mcpgo.WithString("source",
			mcpgo.Required(),
			mcpgo.Description("File path or HTTP/HTTPS URL to the OpenAPI/Swagger spec"),
		),
		mcpgo.WithString("format",
			mcpgo.Description("Format hint: 'yaml' or 'json'. Auto-detected if omitted."),
		),
	)
	s.AddTool(tool, tools.ScanOpenAPIHandler(deps.Parser, deps.Store))
}

func registerCollectionGenerate(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.collection.generate",
		mcpgo.WithDescription("Generate a Postman Collection v2.1 from an OpenAPI/Swagger spec. Returns the full collection object."),
		mcpgo.WithString("spec_source",
			mcpgo.Required(),
			mcpgo.Description("File path or HTTP/HTTPS URL to the OpenAPI/Swagger spec"),
		),
		mcpgo.WithString("collection_name",
			mcpgo.Description("Override the collection name (defaults to spec info.title)"),
		),
		mcpgo.WithString("base_url",
			mcpgo.Description("Override the base URL (defaults to servers[0].url from spec)"),
		),
		mcpgo.WithBoolean("include_auth",
			mcpgo.Description("Include auth configuration (default: true)"),
		),
		mcpgo.WithBoolean("group_by_tags",
			mcpgo.Description("Group endpoints into folders by tag (default: true)"),
		),
	)
	s.AddTool(tool, tools.CollectionGenerateHandler(deps.Parser, deps.Builder))
}

func registerCollectionExport(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.collection.export",
		mcpgo.WithDescription("Export a Postman collection object to a JSON file on disk."),
		mcpgo.WithObject("collection",
			mcpgo.Required(),
			mcpgo.Description("Postman Collection v2.1 object (output of postman.collection.generate)"),
		),
		mcpgo.WithString("output_path",
			mcpgo.Required(),
			mcpgo.Description("File path to write the collection JSON to (directory must exist)"),
		),
	)
	s.AddTool(tool, tools.CollectionExportHandler(deps.Config.DryRun))
}

func registerDocsGenerate(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.docs.generate",
		mcpgo.WithDescription("Enrich a Postman collection with auto-generated descriptions for every request."),
		mcpgo.WithObject("collection",
			mcpgo.Required(),
			mcpgo.Description("Postman Collection v2.1 object"),
		),
		mcpgo.WithString("style",
			mcpgo.Description("Documentation style: 'concise' (default) or 'detailed'"),
		),
	)
	s.AddTool(tool, tools.DocsGenerateHandler(deps.DocGen))
}

func registerDocsAudit(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.docs.audit",
		mcpgo.WithDescription("Audit documentation coverage of a Postman collection. Returns coverage percentage and list of undocumented endpoints."),
		mcpgo.WithObject("collection",
			mcpgo.Required(),
			mcpgo.Description("Postman Collection v2.1 object"),
		),
	)
	s.AddTool(tool, tools.DocsAuditHandler(deps.DocGen))
}

func registerTestsGenerate(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.tests.generate",
		mcpgo.WithDescription("Add Postman JS test scripts to every request in a collection."),
		mcpgo.WithObject("collection",
			mcpgo.Required(),
			mcpgo.Description("Postman Collection v2.1 object"),
		),
		mcpgo.WithArray("test_types",
			mcpgo.Description("Test types to generate: status_code, response_time, schema, auth, required_fields. Defaults to all."),
		),
	)
	s.AddTool(tool, tools.TestsGenerateHandler(deps.TestGen))
}

func registerEnvironmentGenerate(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.environment.generate",
		mcpgo.WithDescription("Generate Postman environment objects for dev, qa, stage, and prod (or custom list)."),
		mcpgo.WithString("base_url",
			mcpgo.Required(),
			mcpgo.Description("Base URL of the API (e.g. https://api.example.com)"),
		),
		mcpgo.WithArray("environments",
			mcpgo.Description("List of environment names (default: [dev, qa, stage, prod])"),
		),
		mcpgo.WithObject("extra_variables",
			mcpgo.Description("Additional key-value pairs to include in all environments"),
		),
	)
	s.AddTool(tool, tools.EnvironmentGenerateHandler(deps.EnvGen))
}

func registerAuditRun(s *server.MCPServer, deps Dependencies) {
	tool := mcpgo.NewTool("postman.audit.run",
		mcpgo.WithDescription("Run a full API quality audit on a Postman collection. Returns a score (0-100) and a list of issues with severity, rule, location, and message."),
		mcpgo.WithObject("collection",
			mcpgo.Required(),
			mcpgo.Description("Postman Collection v2.1 object"),
		),
	)
	s.AddTool(tool, tools.AuditRunHandler(deps.Auditor))
}
