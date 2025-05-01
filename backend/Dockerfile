# Общая стадия builder
FROM golang:1.24.2-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY configs/backend/config.yaml /app/configs/backend/config.yaml
COPY backend/go.mod backend/go.sum ./
RUN go mod download -x && go mod verify
COPY ./backend .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o ./bin/main .


# Prod стадия
FROM alpine:3.21 AS prod
RUN apk add --no-cache bash
WORKDIR /app
COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/configs/backend/config.yaml /app/configs/backend/config.yaml
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
CMD ["/app/bin/main"]


# Dev стадия (только для hot-reload)
FROM golang:1.24.2-alpine AS dev
RUN apk add --no-cache bash inotify-tools && go install github.com/air-verse/air@latest
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
COPY . .
COPY configs/backend/config.yaml /app/config/backend/config.yaml
RUN go mod download -x && go mod verify
CMD ["air", "-c", ".air.toml"]