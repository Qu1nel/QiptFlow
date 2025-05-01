ARG GO_VERSION=1.24.2
ARG ALPINE_VERSION=3.21

# Общий этап для кэширования зависимостей
FROM golang:${GO_VERSION}-alpine AS deps
RUN apk add --no-cache git ca-certificates
WORKDIR /app

COPY backend/go.mod backend/go.sum ./
RUN go mod download -x && go mod verify


# Общая стадия builder
FROM golang:${GO_VERSION}-alpine AS builder
WORKDIR /app

COPY --from=deps /app/go.mod /app/go.sum ./
COPY --from=deps /go/pkg /go/pkg
COPY configs/backend/config.yaml /app/configs/backend/config.yaml
COPY ./backend .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags='-w -s' -o ./bin/main .


# Prod стадия
FROM alpine:${ALPINE_VERSION} AS prod
WORKDIR /app

COPY --from=builder /app/bin/main /app/bin/main
COPY --from=builder /app/configs/backend/config.yaml /app/configs/backend/config.yaml
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

CMD ["/app/bin/main"]


# Dev стадия (только для hot-reload)
FROM golang:${GO_VERSION}-alpine AS dev
RUN apk add --no-cache bash inotify-tools && go install github.com/air-verse/air@latest
WORKDIR /app

COPY . .
COPY configs/backend/config.yaml /app/configs/backend/config.yaml

CMD ["air", "-c", ".air.toml"]