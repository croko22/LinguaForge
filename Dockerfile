# ── Build stage ───────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o /linguaforge-server ./cmd/api/

# ── Runtime stage ─────────────────────────────────────────────────
FROM alpine:3.21

RUN apk add --no-cache ca-certificates curl

WORKDIR /app

COPY --from=builder /linguaforge-server /app/linguaforge-server

# Create data directory
RUN mkdir -p /data/uploads/tts-cache

EXPOSE 8080

ENTRYPOINT ["/app/linguaforge-server"]
