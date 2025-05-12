FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy Go module files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o api-gateway ./cmd/main.go

# Create a minimal production image
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/api-gateway .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./api-gateway"]