# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build

WORKDIR /src
RUN apk add --no-cache build-base

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o /out/server ./cmd/server

FROM alpine:3.20

RUN adduser -D -u 10001 app && apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/server /app/server
RUN mkdir -p /app/data && chown -R app:app /app

USER app
ENV HTTP_ADDR=:8080 \
    DATABASE_PATH=/app/data/app.db \
    LOG_LEVEL=INFO
EXPOSE 8080
VOLUME ["/app/data"]

HEALTHCHECK --interval=30s --timeout=3s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/healthz >/dev/null || exit 1

ENTRYPOINT ["/app/server"]
