# Upbound Marketplace MCP Server
[![CI](https://github.com/upbound/marketplace-mcp-server/actions/workflows/ci.yaml/badge.svg)](https://github.com/upbound/marketplace-mcp-server/actions/workflows/ci.yaml)
[![Slack](https://img.shields.io/badge/slack-upbound_crossplane-purple?logo=slack)](https://crossplane.slack.com/archives/C01TRKD4623)

A Model Context Protocol (MCP) server that provides AI agents with access to the Upbound Marketplace API. This server enables agents to search, discover, and manage marketplace packages and repositories, with a focus on helping users leverage marketplace resources for Crossplane compositions and package management.

Built using the [mcp-go](https://github.com/mark3labs/mcp-go) framework for robust MCP protocol compliance and performance.

## Features

- **Package Search**: Search for packages across the Upbound Marketplace with advanced filtering
- **Package Metadata**: Get detailed information about packages including CRDs, examples, and documentation
- **Asset Access**: Retrieve package assets like CRDs, examples, docs, and package files
- **Repository Management**: Browse and manage repositories
- **Authentication**: UP CLI-based authentication for accessing private resources
- **Multi-API Support**: Supports both v1 and v2 marketplace APIs
- **Composition Focus**: Specialized tools for working with Crossplane compositions and functions

## Installation

### Using Docker (Recommended)

Build the Docker image locally:

```bash
git clone https://github.com/upbound/marketplace-mcp-server.git
cd marketplace-mcp-server
docker build --target stdio -t marketplace-mcp-server:latest .
```

**Note**: You must have the UP CLI installed and authenticated for the server to access marketplace resources. Run `up login` before using the MCP server.

### Building from Source

```bash
git clone https://github.com/upbound/marketplace-mcp-server.git
cd marketplace-mcp-server
go build ./cmd/mcp-server
```

## Usage with AI Agents

### Claude Desktop

Add the following to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "marketplace": {
      "command": "docker",
      "args": [
        "run",
        "--name", "mcp-marketplace",
        "--rm",
        "-i",
        "-v", "/Users/your-username/.up:/mcp/.up:ro",
        "marketplace-mcp-server:latest"
      ]
    }
  }
}
```

**Important**: Replace `/Users/your-username/.up` with your actual UP CLI config directory path:
- **macOS/Linux**: `~/.up` (typically `/Users/username/.up` or `/home/username/.up`)
- **Windows**: `%USERPROFILE%\.up`

### Other MCP-Compatible Agents

For agents that support MCP, configure them to connect to the server using stdio transport:

```bash
# Using the built binary
./mcp-server

# Using Docker
docker run -i --rm -v ~/.up:/mcp/.up:ro marketplace-mcp-server:latest
```

### HTTP API Interface

The server also supports HTTP transport for integration with web applications and REST clients:

```bash
# Start HTTP server locally
./mcp-http

# Or using Docker
docker run --rm -p 8765:8765 -v ~/.up:/mcp/.up:ro marketplace-mcp-server-http:latest
```

The HTTP server provides a JSON-RPC 2.0 API at `http://localhost:8765/mcp`. Example usage:

```bash
# List available tools
curl -X POST http://localhost:8765/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 1, "method": "tools/list", "params": {}}'

# Search for packages
curl -X POST http://localhost:8765/mcp \
  -H "Content-Type: application/json" \
  -d '{"jsonrpc": "2.0", "id": 2, "method": "tools/call", "params": {"name": "search_packages", "arguments": {"query": "aws", "size": 5}}}'
```

The HTTP interface operates in stateless mode, so no session initialization is required.

## Available Tools

### 1. search_packages

Search for packages in the Upbound Marketplace.

**Parameters:**
- `query` (string): Search query for packages
- `family` (string): Family repository key to filter by
- `package_type` (string): Type of package (provider, configuration, function)
- `account_name` (string): Account/organization name to filter by
- `tier` (string): Package tier (official, community, etc.)
- `public` (boolean): Filter by public/private packages
- `size` (integer): Number of results to return (max 500, default 20)
- `page` (integer): Page number (0-indexed, default 0)
- `use_v1` (boolean): Use v1 API instead of v2 (default false)

**Example:**
```json
{
  "name": "search_packages",
  "arguments": {
    "query": "aws provider",
    "tier": "official",
    "public": true,
    "size": 10
  }
}
```

### 2. get_package_metadata

Get detailed metadata for a specific package.

**Parameters:**
- `account` (string, required): Account/organization name
- `repository` (string, required): Repository name
- `version` (string): Package version (optional, gets latest if not specified)
- `use_v1` (boolean): Use v1 API instead of v2 (default false)

**Example:**
```json
{
  "name": "get_package_metadata",
  "arguments": {
    "account": "upbound",
    "repository": "provider-aws"
  }
}
```

### 3. get_package_assets

Get assets (documentation, icons, release notes, etc.) for a specific package version.

**Parameters:**
- `account` (string, required): Account/organization name
- `repository` (string, required): Repository name
- `version` (string, required): Package version or 'latest'
- `asset_type` (string, required): Type of asset (docs, icon, readme, releaseNotes, sbom)

**Example:**
```json
{
  "name": "get_package_assets",
  "arguments": {
    "account": "upbound",
    "repository": "provider-aws",
    "version": "latest",
    "asset_type": "docs"
  }
}
```

### 4. get_repositories

Get repositories for an account.

**Parameters:**
- `account` (string, required): Account/organization name
- `filter` (string): AIP-160 formatted filter (v2 only)
- `size` (integer): Number of results to return (default 20)
- `page` (integer): Page number (0-indexed, default 0)
- `use_v1` (boolean): Use v1 API instead of v2 (default false)

**Example:**
```json
{
  "name": "get_repositories",
  "arguments": {
    "account": "upbound",
    "filter": "type = 'provider' AND public = true"
  }
}
```

### 5. reload_auth

Reload authentication from UP CLI configuration. Useful when switching UP CLI profiles.

**Parameters:**
- No parameters required

**Example:**
```json
{
  "name": "reload_auth",
  "arguments": {}
}
```

### 6. get_package_version_resources

Get package version resources for a supplied repository name.

**Parameters:**
- `account` (string, required): Account/organization name. For example upbound.
- `repository_name` (string, required): The name of the repository. For example provider-aws-s3.
- `version` (string, required): The version of the package. For example v1.23.1.

**Example:**
```json
{
  "name": "get_package_version_resources",
  "arguments": {
    "account": "upbound",
    "repository_name": "provider-aws-s3",
    "version": "v1.23.1 
  }
}
```

### 7. get_package_version_composition_resources

Get package version composition resources for a supplied group, kind and version and composition.

**Parameters:**
- `account` (string, required): Account/organization name. For example upbound.
- `repository_name` (string, required): The name of the repository. For example configuration-caas.
- `version` (string, required): The version of the package. For example v0.4.0.
- `resource_group` (string, required): The group of the resource. For example caas.upbound.io.
- `resource_kind` (string, required): The kind of the resource. For example XCluster.
- `composition_name` (string, required): The kind of the resource. For example xclusters.caas.upbound.io.

**Example:**
```json
{
  "name": "get_package_version_composition_resources",
  "arguments": {
    "account": "upbound",
    "repository_name": "configuration-caas",
    "version": "v0.4.0",
    "resource_group": "caas.upbound.io",
    "resource_kind": "XCluster",
    "composition_name": "xclusters.caas.upbound.io"
  }
}
```

### 8. get_package_version_groupkind_resources

Get package version resources for a supplied group, kind and version.

**Parameters:**
- `account` (string, required): Account/organization name. For example upbound.
- `repository_name` (string, required): The name of the repository. For example provider-aws-s3.
- `version` (string, required): The version of the package. For example v1.23.1.
- `resource_group` (string, required): The group of the resource. For example s3.aws.upbound.io.
- `resource_kind` (string, required): The kind of the resource. For example Bucket.

**Example:**
```json
{
  "name": "get_package_version_groupkind_resources",
  "arguments": {
    "account": "upbound",
    "repository_name": "provider-aws-s3",
    "version": "v1.23.1",
    "resource_group": "s3.aws.upbound.io",
    "resource_kind": "Bucket"
  }
}
```

### 8. get_package_version_examples

Get package version examples for a supplied group, kind and version.

**Parameters:**
- `account` (string, required): Account/organization name. For example upbound.
- `repository_name` (string, required): The name of the repository. For example provider-aws-s3.
- `version` (string, required): The version of the package. For example v1.23.1.
- `resource_group` (string, required): The group of the resource. For example s3.aws.upbound.io.
- `resource_kind` (string, required): The kind of the resource. For example Bucket.

**Example:**
```json
{
  "name": "get_package_version_groupkind_resources",
  "arguments": {
    "account": "upbound",
    "repository_name": "provider-aws-s3",
    "version": "v1.23.1",
    "resource_group": "s3.aws.upbound.io",
    "resource_kind": "Bucket"
  }
}
```

## Authentication

The MCP server uses UP CLI authentication for accessing marketplace resources:

### Prerequisites
1. Install the UP CLI: https://docs.upbound.io/cli/
2. Authenticate with your Upbound account: `up login`
3. Ensure your UP CLI config is accessible to the Docker container

### Docker Configuration
The server automatically loads authentication from your UP CLI configuration when the container starts. Make sure to mount your UP CLI config directory:

```bash
-v ~/.up:/mcp/.up:ro
```

### Switching Profiles
If you have multiple UP CLI profiles, you can:
1. Switch profiles using `up profile use <profile-name>`
2. Use the `reload_auth` tool to reload the new authentication without restarting the server

### Private Resources
The server will automatically use your authenticated session to access private repositories and resources that your account has permission to view.

## API Filtering (v2)

The v2 API supports advanced filtering using AIP-160 format:

### Package Search Filters
- `query = 'crossplane'` - Text search
- `family = 'upbound/provider-aws'` - Family repository
- `packageType = 'provider'` - Package type
- `accountName = 'upbound'` - Account name
- `public = true` - Public packages only
- `tier = 'official'` - Official tier packages

### Repository Filters
- `type = 'provider'` - Repository type
- `name = 'my-repo'` - Repository name
- `public = true` - Public repositories
- `policy = 'publish'` - Repository policy
- `creation_date > '2023-01-01'` - Created after date

### Combining Filters
```
(accountName = 'upbound' OR accountName = 'crossplane') AND public = true AND tier = 'official'
```

## Use Cases

### 1. Package Discovery
Find packages for specific cloud providers or use cases:
```
Search for "AWS S3" packages to find providers and configurations for S3 resources.
```

### 2. Composition Development
Get examples and CRDs for building compositions:
```
1. Search for provider packages
2. Get package metadata to see available CRDs
3. Get examples to understand usage patterns
4. Use CRDs and examples to build compositions
```

### 3. Package Analysis
Analyze package dependencies and compatibility:
```
1. Get package metadata for dependency information
2. Check available versions
3. Review documentation and examples
```

### 4. Repository Management
Browse and manage organization repositories:
```
1. Ensure UP CLI is authenticated (up login)
2. List repositories with filtering
3. Get detailed repository information
```

## Configuration

The server automatically detects and loads UP CLI configuration from the 
following locations:
1. `/mcp/.up/config.json` (when running in Docker with mounted config)
2. `~/.up/config.json` (default UP CLI location)

No additional configuration is required if UP CLI is properly set up and 
authenticated.

### As an Addon
Note, the marketplace-mcp-server does still need authentication as described in
the above section. In order to fulfill that need, you should provide a secret
with the contents of the ~/.up/config.json.

For example:
```bash
  kubectl -n crossplane-system create secret generic up-config --from-file=config.json=path/to/up/config.json
```

## Development

### Prerequisites
- Go 1.23 or later (required by mcp-go framework)
- Docker (for containerization)

### Running Locally

**Stdio Transport (for MCP clients like Cursor):**
```bash
# Run the built binary directly
./mcp-server

# Or run with Docker
docker run -i --rm -v ~/.up:/mcp/.up:ro marketplace-mcp-server:latest
```

**HTTP Transport (for web applications and REST clients):**
```bash
# Run the HTTP server
./mcp-http

# Or run with Docker
docker run --rm -p 8765:8765 -v ~/.up:/mcp/.up:ro marketplace-mcp-server-http:latest
```

### Testing
```bash
go test ./...
```


## Architecture

The server is built using the [mcp-go](https://github.com/mark3labs/mcp-go) framework, which provides:
- **JSON-RPC 2.0 Compliance**: Full adherence to MCP protocol specifications
- **Multiple Transports**: Built-in support for stdio, HTTP, and SSE transports
- **Type Safety**: Strongly typed request/response handling
- **Middleware Support**: Extensible architecture for authentication and logging
- **Error Handling**: Robust error handling with proper MCP error codes

### Key Components

- **Server**: Main MCP server using mcp-go framework
- **Handlers**: Tool handlers for marketplace operations
- **Auth Manager**: UP CLI authentication integration
- **Marketplace Client**: HTTP client for Upbound Marketplace API

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

For issues and questions:
- Create an issue on GitHub
- Contact the maintainers

## Examples

### Finding AWS Providers
```json
{
  "name": "search_packages",
  "arguments": {
    "query": "aws",
    "package_type": "provider",
    "tier": "official"
  }
}
```

### Getting Package Documentation
```json
{
  "name": "get_package_assets",
  "arguments": {
    "account": "upbound",
    "repository": "provider-aws",
    "version": "latest",
    "asset_type": "docs"
  }
}
```

### Browsing Organization Repositories
```json
{
  "name": "get_repositories",
  "arguments": {
    "account": "crossplane-contrib",
    "filter": "type = 'configuration' AND public = true"
  }
}
```
