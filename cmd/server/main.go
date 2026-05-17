package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/isai-salazar-enc/postman-go-mcp/internal/audit"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/config"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/discovery"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/docs"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/environments"
	mcpserver "github.com/isai-salazar-enc/postman-go-mcp/internal/mcp"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/postman"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/setup"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/storage"
	"github.com/isai-salazar-enc/postman-go-mcp/internal/tests"
	"github.com/mark3labs/mcp-go/server"
)

const version = "1.0.0"

func main() {
	// Subcommand routing
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "setup":
			runSetup()
			return
		case "--version", "-v", "version":
			fmt.Printf("mcp-postman %s\n", version)
			return
		case "--help", "-h", "help":
			printHelp()
			return
		}
	}

	runMCPServer()
}

func runSetup() {
	// Check for direct client shortcuts: mcp-postman setup opencode / mcp-postman setup claude
	if len(os.Args) > 2 {
		var client setup.Client
		switch os.Args[2] {
		case "opencode":
			client = setup.ClientOpenCode
		case "claude":
			client = setup.ClientClaudeCode
		default:
			fmt.Fprintf(os.Stderr, "unknown setup target: %q (use 'opencode' or 'claude')\n", os.Args[2])
			os.Exit(1)
		}
		if err := setup.RunDirect(client); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	// No argument — launch interactive wizard
	if err := setup.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMCPServer() {
	// MCP transport owns stdout entirely — redirect all log output to stderr.
	log.SetOutput(os.Stderr)

	cfg := config.Load()

	// Ensure data directory exists
	dbDir := filepath.Dir(cfg.DBPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}

	// Initialize SQLite store
	store, err := storage.New(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Wire dependencies
	deps := mcpserver.Dependencies{
		Config:  cfg,
		Parser:  discovery.NewLibOpenAPIParser(),
		Builder: postman.NewBuilder(),
		DocGen:  docs.NewGenerator(),
		TestGen: tests.NewGenerator(),
		EnvGen:  environments.NewGenerator(),
		Auditor: audit.NewAuditor(),
		Store:   store,
	}

	// Create and start MCP server
	s := mcpserver.NewServer(deps)

	log.Printf("Starting postman-go-mcp v%s (transport: %s)", version, cfg.MCPTransport)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}

func printHelp() {
	fmt.Printf(`mcp-postman %s — Postman Developer Assistant MCP Server

USAGE:
  mcp-postman                  Start the MCP server (stdio transport)
  mcp-postman setup            Run the interactive setup wizard
  mcp-postman setup opencode   Configure OpenCode directly (non-interactive)
  mcp-postman setup claude     Configure Claude Code directly (non-interactive)
  mcp-postman version          Show version
  mcp-postman help             Show this help

ENVIRONMENT:
  MCP_TRANSPORT    Transport mode: stdio | sse (default: stdio)
  MCP_HTTP_PORT    HTTP port for SSE mode (default: 8080)
  DB_PATH          SQLite database path (default: ./data/postman.db)
  LOG_LEVEL        Log level: debug|info|warn|error (default: info)
  DRY_RUN          When true, skip file writes (default: false)

MCP TOOLS:
  postman.scan.openapi         Parse an OpenAPI/Swagger spec
  postman.collection.generate  Generate a Postman Collection v2.1
  postman.collection.export    Export collection to JSON file
  postman.docs.generate        Enrich collection with descriptions
  postman.docs.audit           Audit documentation coverage
  postman.tests.generate       Add Postman JS test scripts
  postman.environment.generate Generate dev/qa/stage/prod environments
  postman.audit.run            Run a full API quality audit

SETUP:
  'mcp-postman setup'            Interactive TUI wizard (select client, preview, confirm)
  'mcp-postman setup opencode'   Direct: writes to OpenCode config (mcp.postman)
  'mcp-postman setup claude'     Direct: writes to Claude Code config (mcpServers.postman)
`, version)
}
