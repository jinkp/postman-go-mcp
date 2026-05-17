# postman-go-mcp

An MCP (Model Context Protocol) server in Go that acts as a developer intelligence layer for APIs and Postman. Connect it to OpenCode, Claude, or any MCP-compatible AI assistant to automate API discovery, collection generation, documentation, testing, and quality auditing.

## Features

| MCP Tool | Description |
|----------|-------------|
| `postman.scan.openapi` | Parse an OpenAPI 3.0 or Swagger 2.0 spec from a file path or URL |
| `postman.collection.generate` | Generate a Postman Collection v2.1 from a spec |
| `postman.collection.export` | Export a collection to a JSON file |
| `postman.docs.generate` | Enrich a collection with auto-generated descriptions |
| `postman.docs.audit` | Audit documentation coverage of a collection |
| `postman.tests.generate` | Add Postman JS test scripts to every request |
| `postman.environment.generate` | Generate dev/qa/stage/prod Postman environments |
| `postman.audit.run` | Run a full API quality audit and get a score |

## Installation

```bash
git clone https://github.com/isai-salazar-enc/postman-go-mcp
cd postman-go-mcp
cp .env.example .env
go build -o mcp-postman ./cmd/server
```

## Usage

### stdio (for AI assistants)

```bash
./mcp-postman
```

Add to your OpenCode / Claude Desktop config:

```json
{
  "mcpServers": {
    "postman": {
      "command": "/path/to/mcp-postman"
    }
  }
}
```

### Docker

```bash
docker compose up --build
```

## Configuration

Copy `.env.example` to `.env` and adjust:

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | Transport: `stdio` or `sse` |
| `MCP_HTTP_PORT` | `8080` | HTTP port (SSE mode only) |
| `DB_PATH` | `./data/postman.db` | SQLite database path |
| `LOG_LEVEL` | `info` | Log level: debug/info/warn/error |
| `DRY_RUN` | `false` | When true, no files are written |

## Example Workflow

```
1. Scan your API spec
   → postman.scan.openapi { source: "./api/openapi.yaml" }

2. Generate a collection
   → postman.collection.generate { spec_source: "./api/openapi.yaml" }

3. Add documentation
   → postman.docs.generate { collection: <output from step 2> }

4. Add tests
   → postman.tests.generate { collection: <output from step 3> }

5. Run quality audit
   → postman.audit.run { collection: <output from step 4> }

6. Export to file
   → postman.collection.export { collection: <output>, output_path: "./postman/collection.json" }
```

## Architecture

Clean/Hexagonal architecture. Each domain is independent:

```
cmd/server/         — Entry point
internal/
  mcp/              — MCP delivery layer (thin tool adapters)
  discovery/        — OpenAPI/Swagger parser
  postman/          — Postman Collection v2.1 builder
  docs/             — Documentation generator & auditor
  tests/            — Postman test script generator
  environments/     — Environment generator
  audit/            — Quality rules engine
  storage/          — SQLite persistence
  config/           — Configuration
  security/         — Path validation, header redaction
pkg/postmanfmt/     — Shared Postman v2.1 Go types
```

## License

MIT
