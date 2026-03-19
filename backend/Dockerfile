# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Set Go proxy for China network
ENV GOPROXY=https://goproxy.cn,direct

# Install dependencies
RUN apk add --no-cache git gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o smpp-simulator ./cmd/server

# Runtime stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1000 -S smpp && \
    adduser -u 1000 -S smpp -G smpp

# Copy binary from builder
COPY --from=builder --chown=smpp:smpp /app/smpp-simulator .

# Create directory for database and set permissions
RUN mkdir -p /app/data && chown -R smpp:smpp /app/data

# Switch to non-root user
USER smpp

# Expose ports
EXPOSE 2775 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run
CMD ["./smpp-simulator"]
