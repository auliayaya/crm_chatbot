FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-service ./cmd/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/chat-service .

EXPOSE 8081
CMD ["./chat-service"]