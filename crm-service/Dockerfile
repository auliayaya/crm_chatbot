FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o crm-service ./cmd/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/crm-service .

EXPOSE 8092
CMD ["./crm-service"]