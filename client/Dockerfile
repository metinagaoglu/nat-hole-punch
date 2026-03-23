# Multi-stage build for optimized image size
# Stage 1: Build
FROM golang:1.23-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /build

# Copy source code
COPY go.mod ./
COPY *.go ./

# Download dependencies (if any)
RUN go mod tidy 2>/dev/null || true

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a \
    -o udp-client \
    .

# Stage 2: Runtime
FROM scratch

# Copy CA certificates
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy binary from builder
COPY --from=builder /build/udp-client /udp-client

# Run as non-root user
USER 65534:65534

# Set entrypoint
ENTRYPOINT ["/udp-client"]

# Default flags (can be overridden)
CMD ["-signal-address=udp-server:3986", "-local-address=0.0.0.0:0", "-room-key=default"]
