FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/feed-tgbot

FROM alpine:3.22

RUN apk add --no-cache ca-certificates su-exec \
    && addgroup -S bot \
    && adduser -S -G bot bot

WORKDIR /app

COPY --from=builder /out/feed-tgbot /usr/local/bin/feed-tgbot
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh

RUN chmod +x /usr/local/bin/docker-entrypoint.sh \
    && mkdir -p /app \
    && chown -R bot:bot /app

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
CMD ["/usr/local/bin/feed-tgbot"]
