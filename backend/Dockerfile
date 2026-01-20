# Build Stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build the binary
# CGO_ENABLED=0 for static binary
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/server

# Final Stage
FROM alpine:latest

WORKDIR /app

# Install certificates for external requests if needed
RUN apk --no-cache add ca-certificates

COPY --from=builder /app/main .
COPY --from=builder /app/internal/config ./internal/config

# Expose port
EXPOSE 8080

CMD ["./main"]
