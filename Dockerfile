# Build stage
FROM golang:1.25.6-alpine AS builder

# Install build dependencies for CGO (required by webp library)
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with CGO enabled
RUN CGO_ENABLED=1 GOOS=linux go build -o bin/familyshare ./cmd/app

# Final stage
FROM alpine:latest

# Install CA certificates for HTTPS and timezone data
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/familyshare /app/familyshare

# Create necessary directories
RUN mkdir -p /app/data /app/tmp_uploads

# Set environment variables
ENV SERVER_ADDR=":8080" \
    DATA_DIR="/app/data" \
    DATABASE_PATH="/app/data/familyshare.db" \
    TEMP_UPLOAD_DIR="/app/tmp_uploads" \
    GIN_MODE="release"

# Expose port
EXPOSE 8080

# Run the binary
CMD ["/app/familyshare"]