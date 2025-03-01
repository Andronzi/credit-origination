FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git
RUN go install github.com/pressly/goose/v3/cmd/goose@latest

FROM alpine:3.19

RUN apk add --no-cache bash libc6-compat && \
    apk add --no-cache --virtual .deps dos2unix

COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY entrypoint.sh /entrypoint.sh
COPY db/migrations /db/migrations

RUN dos2unix /entrypoint.sh && \
    chmod +x /entrypoint.sh && \
    apk del .deps  # Удаляем временные зависимости

ENTRYPOINT ["/entrypoint.sh"]