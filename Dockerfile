# Build stage
FROM golang:1.21-alpine AS builder

# Install security updates and build dependencies
RUN apk update && \
    apk add --no-cache ca-certificates tzdata && \
    update-ca-certificates

# Create non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with security flags
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /go/bin/portservice ./cmd/portservice

# Final stage
FROM scratch

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/bin/portservice /portservice

# Use non-root user
USER appuser

# Set environment variables
ENV PORT=8080
ENV MAX_MEMORY_MB=200

# Expose port
EXPOSE 8080

# Run the application
ENTRYPOINT ["/portservice"] 