# Build stage
FROM golang:1.21-alpine AS builder

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

# Build the server
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-mcp-server ./cmd/mcp-server

# Build the proxy
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o marketplace-mcp-proxy ./cmd/mcp-proxy

# Separate targets for stdio vs http transport implementations
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

# Expose port (needed for Auth callback)
EXPOSE 8765

# Run the application
ENTRYPOINT ["./marketplace-mcp-server"]

FROM alpine:latest AS http

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN adduser -D -s /bin/sh mcp

# Set working directory
WORKDIR /home/mcp

# Copy the binary from builder stage
COPY --from=builder /app/marketplace-mcp-server ./bin/
COPY --from=builder /app/marketplace-mcp-proxy .

# Change ownership to mcp user
RUN chown mcp:mcp ./bin/marketplace-mcp-server
RUN chown mcp:mcp marketplace-mcp-proxy

# Switch to non-root user
USER mcp

# Expose port (if needed for OAuth callback)
EXPOSE 8765

# Run the application
ENTRYPOINT ["./marketplace-mcp-proxy"] 