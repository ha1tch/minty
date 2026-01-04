# Build stage
ARG GO_VERSION=1.22
FROM golang:${GO_VERSION}-alpine AS builder

ARG VERSION=dev

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

# Copy root go.mod for minty dependency
COPY go.mod go.sum ./

# Copy insurance-quote go.mod
COPY examples/insurance-quote/go.mod examples/insurance-quote/go.sum ./examples/insurance-quote/

# Download dependencies
RUN go mod download
RUN cd examples/insurance-quote && go mod download

# Copy source code
COPY . .

# Build the binary
RUN cd examples/insurance-quote && CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o /app/insurequote \
    ./cmd/insurequote

# Final stage
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 app && \
    adduser -u 1000 -G app -s /bin/sh -D app

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/insurequote .

# Set ownership
RUN chown -R app:app /app

USER app

# Expose default port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the application
ENTRYPOINT ["./insurequote"]
