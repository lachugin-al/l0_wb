FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o l0_wb ./cmd/app

# Create a minimal runtime image
FROM alpine:3.18

WORKDIR /app

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/l0_wb .

# Copy web directory for static files
COPY web/ ./web/

# Copy migration files
COPY --from=builder /app/internal/db/migrations/ ./internal/db/migrations/

# Ensure migrations directory has correct permissions
RUN chmod -R 755 ./internal/db/migrations/

# Expose the HTTP port
EXPOSE 8081
# Expose the metrics port
EXPOSE 9100

# Run the application
CMD ["./l0_wb"]
