FROM golang:1.25-bookworm AS go-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

FROM oven/bun:1-debian AS bun-deps

WORKDIR /src

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

FROM go-base AS generated

COPY . .
COPY --from=bun-deps /src/node_modules ./node_modules
COPY --from=bun-deps /usr/local/bin/bun /usr/local/bin/bun
RUN go tool templ generate ./... \
    && go run github.com/fastygo/ui8kit/scripts/cmd/sync-assets web/static \
    && bun ./scripts/append-gocms-locale-sync.mjs

FROM oven/bun:1-debian AS assets

WORKDIR /src

COPY --from=generated /src ./
RUN bun run build:css \
    && bun run build:versioned \
    && rm -rf node_modules

FROM go-base AS build

WORKDIR /src

COPY --from=assets /src ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/gocms ./cmd/server \
    && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/healthcheck ./cmd/healthcheck \
    && mkdir -p /out/data \
    && chown -R 65532:65532 /out/data

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/gocms /gocms
COPY --from=build /out/healthcheck /healthcheck
COPY --from=build /out/data /data
COPY --from=assets /src/web/static /app/web/static

ENV APP_BIND=0.0.0.0:8080
ENV APP_DATA_SOURCE=file:/data/gocms.db
ENV GOCMS_PRESET=full
ENV GOCMS_STORAGE_PROFILE=sqlite
ENV GOCMS_DEPLOYMENT_PROFILE=container
ENV HEALTHCHECK_URL=http://127.0.0.1:8080/readyz

EXPOSE 8080

USER nonroot:nonroot

HEALTHCHECK --interval=30s --timeout=5s --start-period=15s --retries=3 \
    CMD ["/healthcheck"]

ENTRYPOINT ["/gocms"]
