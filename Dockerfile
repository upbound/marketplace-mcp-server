# Build stage
FROM golang:1.23-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the stdio server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-mcp-server ./cmd/mcp-server

# Build the HTTP server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-mcp-http ./cmd/mcp-http

# stdio target for MCP clients
FROM alpine:latest AS stdio

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh mcp

# Set working directory
WORKDIR /home/mcp

# Copy the binary from builder stage
COPY --from=builder /app/marketplace-mcp-server .

# Change ownership to mcp user
RUN chown mcp:mcp marketplace-mcp-server

# Switch to non-root user
USER mcp

# Run the application
ENTRYPOINT ["./marketplace-mcp-server"]

# http target for HTTP/SSE transport
FROM alpine:latest AS http

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh mcp

# Set working directory
WORKDIR /home/mcp

# Copy the binary from builder stage
COPY --from=builder /app/marketplace-mcp-http .

# Change ownership to mcp user
RUN chown mcp:mcp marketplace-mcp-http

# Switch to non-root user
USER mcp

# Expose port for HTTP server
EXPOSE 8765

# Run the application
ENTRYPOINT ["./marketplace-mcp-http"] 