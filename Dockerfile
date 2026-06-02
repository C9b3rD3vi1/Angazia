FROM golang:1.26-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/api ./cmd/api && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/indexer ./cmd/indexer && \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /build/syncer ./cmd/syncer

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata curl

ENV TZ=Africa/Nairobi
ENV TEMPLATE_DIR=/app/web/templates
ENV STATIC_DIR=/app/web/static

WORKDIR /app

COPY --from=builder /build/api /usr/local/bin/api
COPY --from=builder /build/indexer /usr/local/bin/indexer
COPY --from=builder /build/syncer /usr/local/bin/syncer
COPY --from=builder /src/web /app/web

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:3000/health || exit 1

ENTRYPOINT ["api"]
