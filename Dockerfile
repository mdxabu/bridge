# Build stage
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o bridge .

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache ca-certificates iptables iproute2

WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/bridge /app/bridge

# Copy configuration
COPY bridgeconfig.yaml /app/

# Expose API port
EXPOSE 8080

# Add capabilities for network operations
# Note: Container must run with --cap-add=NET_ADMIN

ENTRYPOINT ["/app/bridge"]
CMD ["start"]
