# Общая стадия builder
FROM golang:1.24.2-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x && go mod verify
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o qiptflow .

# Prod стадия
FROM alpine:3.21 AS prod
WORKDIR /app
COPY --from=builder /app/qiptflow /app/qiptflow
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app/qiptflow"]

# Dev стадия (только для hot-reload)
FROM golang:1.24.2-alpine AS dev
RUN apk add --no-cache bash inotify-tools && go install github.com/air-verse/air@latest
WORKDIR /app
COPY go.mod go.sum .air.toml ./
RUN go mod download -x && go mod verify
CMD ["air", "-c", ".air.toml"]