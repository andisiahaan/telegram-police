# Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /build

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bot ./cmd/bot/

# Stage 2: Runtime
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /build/bot .

# Config bisa di-mount atau diisi via env variable
# COPY config.yaml . (opsional, gunakan volume atau env)

EXPOSE 8080

ENTRYPOINT ["/app/bot"]
