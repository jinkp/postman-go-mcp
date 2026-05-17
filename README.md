# postman-go-mcp

An MCP (Model Context Protocol) server in Go that acts as a developer intelligence layer for APIs and Postman. Connect it to OpenCode or Claude Code to automate API discovery, collection generation, documentation, testing, and quality auditing.

[![CI](https://github.com/jinkp/postman-go-mcp/actions/workflows/ci.yml/badge.svg)](https://github.com/jinkp/postman-go-mcp/actions/workflows/ci.yml)
[![Release](https://github.com/jinkp/postman-go-mcp/actions/workflows/release.yml/badge.svg)](https://github.com/jinkp/postman-go-mcp/releases)

---

## Installation

### Linux / macOS

```bash
curl -sSfL https://raw.githubusercontent.com/jinkp/postman-go-mcp/master/scripts/install.sh | sh
```

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/jinkp/postman-go-mcp/master/scripts/install.ps1 | iex
```

### Manual download

Download the latest binary from [GitHub Releases](https://github.com/jinkp/postman-go-mcp/releases) for your platform:

| Platform | File |
|----------|------|
| Linux amd64 | `mcp-postman_linux_amd64.tar.gz` |
| Linux arm64 | `mcp-postman_linux_arm64.tar.gz` |
| macOS amd64 | `mcp-postman_darwin_amd64.tar.gz` |
| macOS arm64 (M1/M2) | `mcp-postman_darwin_arm64.tar.gz` |
| Windows amd64 | `mcp-postman_windows_amd64.zip` |

### Build from source

```bash
git clone https://github.com/jinkp/postman-go-mcp
cd postman-go-mcp
go build -o mcp-postman ./cmd/server
```

---

## Setup

After installing, run the interactive setup wizard:

```bash
mcp-postman setup
```

The wizard will:
1. Ask which AI assistant to configure (OpenCode, Claude Code, or both)
2. Auto-detect the binary path
3. Show a preview of the config changes
4. Write the config atomically with automatic backup

No manual JSON editing required.

---

## MCP Tools

| Tool | Description |
|------|-------------|
| `postman.scan.openapi` | Parse an OpenAPI 3.0 or Swagger 2.0 spec from a file path or URL |
| `postman.collection.generate` | Generate a Postman Collection v2.1 from a spec |
| `postman.collection.export` | Export a collection to a JSON file |
| `postman.docs.generate` | Enrich a collection with auto-generated descriptions |
| `postman.docs.audit` | Audit documentation coverage of a collection |
| `postman.tests.generate` | Add Postman JS test scripts to every request |
| `postman.environment.generate` | Generate dev/qa/stage/prod Postman environments |
| `postman.audit.run` | Run a full API quality audit and get a score |

---

## Manual Configuration

If you prefer to configure manually, add this to your client config:

**OpenCode** (`~/.config/opencode/opencode.json`):

```json
{
  "mcpServers": {
    "postman": {
      "command": "/usr/local/bin/mcp-postman"
    }
  }
}
```

**Claude Code** (`~/.claude.json`):

```json
{
  "mcpServers": {
    "postman": {
      "command": "/usr/local/bin/mcp-postman"
    }
  }
}
```

---

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

---

## Configuration

Copy `.env.example` to `.env` and adjust:

| Variable | Default | Description |
|----------|---------|-------------|
| `MCP_TRANSPORT` | `stdio` | Transport: `stdio` or `sse` |
| `MCP_HTTP_PORT` | `8080` | HTTP port (SSE mode only) |
| `DB_PATH` | `./data/postman.db` | SQLite database path |
| `LOG_LEVEL` | `info` | Log level: debug/info/warn/error |
| `DRY_RUN` | `false` | When true, no files are written |

---

## Docker

```bash
docker compose up --build
```

---

## Architecture

Clean/Hexagonal architecture. Each domain is independent:

```
cmd/server/         — Entry point + subcommand routing
internal/
  mcp/              — MCP delivery layer (thin tool adapters)
  discovery/        — OpenAPI/Swagger parser (libopenapi)
  postman/          — Postman Collection v2.1 builder
  docs/             — Documentation generator & auditor
  tests/            — Postman test script generator
  environments/     — Environment generator
  audit/            — Quality rules engine
  storage/          — SQLite persistence (no CGO)
  config/           — Configuration
  security/         — Path validation, header redaction
  setup/            — Setup wizard logic
pkg/postmanfmt/     — Shared Postman v2.1 Go types
scripts/            — Install scripts (install.sh, install.ps1)
```

---

## Creating a Release

Tag a version to trigger the automated release pipeline:

```bash
git tag v1.0.0
git push origin v1.0.0
```

GitHub Actions will build binaries for all platforms and publish them to GitHub Releases automatically.

---

## License

MIT
