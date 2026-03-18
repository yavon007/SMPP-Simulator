# Build stage
FROM golang:1.21-alpine AS builder

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

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy binary from builder
COPY --from=builder /app/smpp-simulator .

# Expose ports
EXPOSE 2775 8080

# Run
CMD ["./smpp-simulator"]
