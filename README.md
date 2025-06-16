# Upbound Marketplace MCP Server

A Model Context Protocol (MCP) server that provides AI agents with access to the Upbound Marketplace API. This server enables agents to search, discover, and manage marketplace packages and repositories, with a focus on helping users leverage marketplace resources for Crossplane compositions and package management.

## Features

- **Package Search**: Search for packages across the Upbound Marketplace with advanced filtering
- **Package Metadata**: Get detailed information about packages including CRDs, examples, and documentation
- **Asset Access**: Retrieve package assets like CRDs, examples, docs, and package files
- **Repository Management**: Browse and manage repositories
- **Authentication**: OAuth-based authentication for accessing private resources
- **Multi-API Support**: Supports both v1 and v2 marketplace APIs
- **Composition Focus**: Specialized tools for working with Crossplane compositions and functions

## Installation

### Using Docker (Recommended)

Pull the pre-built image from the Upbound registry:

```bash
docker pull xpkg.upbound.io/upbound/marketplace-mcp-server:latest
```

### Building from Source

```bash
git clone https://github.com/upbound/marketplace-mcp-server.git
cd marketplace-mcp-server
go build -o marketplace-mcp-server .
```

### Building Docker Image

```bash
docker build -t marketplace-mcp-server .
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
        "-i",
        "--rm",
        "-p", "8765:8765",
        "xpkg.upbound.io/upbound/marketplace-mcp-server:latest"
      ]
    }
  }
}
```

### Other MCP-Compatible Agents

For agents that support MCP, configure them to connect to the server using stdio transport:

```bash
./marketplace-mcp-server
```

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

Get assets (CRDs, examples, docs) for a specific package version.

**Parameters:**
- `account` (string, required): Account/organization name
- `repository` (string, required): Repository name
- `version` (string, required): Package version or 'latest'
- `asset_type` (string, required): Type of asset (crds, examples, docs, package)

**Example:**
```json
{
  "name": "get_package_assets",
  "arguments": {
    "account": "upbound",
    "repository": "provider-aws",
    "version": "latest",
    "asset_type": "examples"
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

### 5. authenticate

Authenticate with Upbound to access private resources.

**Parameters:**
- `client_id` (string): OAuth client ID (optional, uses default if not provided)
- `scopes` (array): OAuth scopes to request (default: ["read:packages", "read:repositories"])

**Example:**
```json
{
  "name": "authenticate",
  "arguments": {
    "scopes": ["read:packages", "read:repositories", "write:packages"]
  }
}
```

## Authentication

For accessing private repositories or performing write operations, you'll need to authenticate:

1. Use the `authenticate` tool in your agent
2. A browser window will open for OAuth login
3. Complete the login process
4. The server will cache your token for subsequent requests

The authentication uses OAuth 2.0 with PKCE for security.

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
1. Authenticate with your account
2. List repositories with filtering
3. Get detailed repository information
```

## Environment Variables

- `MARKETPLACE_BASE_URL`: Override the default marketplace URL (default: https://registry.upbound.io)
- `OAUTH_CLIENT_ID`: Default OAuth client ID
- `OAUTH_CLIENT_SECRET`: OAuth client secret (if required)

## Development

### Prerequisites
- Go 1.21 or later
- Docker (for containerization)

### Running Locally
```bash
go run main.go
```

### Testing
```bash
go test ./...
```

### Building
```bash
go build -o marketplace-mcp-server .
```

## Docker Registry

The Docker image is available at:
```
xpkg.upbound.io/upbound/marketplace-mcp-server:latest
```

### Pushing to Registry

```bash
# Build the image
docker build -t marketplace-mcp-server .

# Tag for registry
docker tag marketplace-mcp-server xpkg.upbound.io/upbound/marketplace-mcp-server:latest

# Push to registry
docker push xpkg.upbound.io/upbound/marketplace-mcp-server:latest
```

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

### Getting Composition Examples
```json
{
  "name": "get_package_assets",
  "arguments": {
    "account": "upbound",
    "repository": "configuration-aws-eks",
    "version": "latest",
    "asset_type": "examples"
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
