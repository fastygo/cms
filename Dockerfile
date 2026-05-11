FROM golang:1.25-bookworm AS go-base

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

FROM go-base AS generated

COPY . .
RUN go tool templ generate ./... \
    && go run github.com/fastygo/ui8kit/scripts/cmd/sync-assets web/static

FROM node:22-bookworm AS assets

WORKDIR /src

COPY package.json package-lock.json ./
RUN npm ci

COPY --from=generated /src ./
RUN npm run build:css \
    && npm run build:versioned \
    && rm -rf node_modules

FROM go-base AS build

WORKDIR /src

COPY --from=assets /src ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/gocms ./cmd/server \
    && mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=build /out/gocms /gocms
COPY --from=build --chown=nonroot:nonroot /out/data /data
COPY --from=assets /src/web/static /app/web/static

ENV APP_BIND=0.0.0.0:8080
ENV APP_DATA_SOURCE=file:/data/gocms.db
ENV GOCMS_PRESET=full
ENV GOCMS_STORAGE_PROFILE=sqlite
ENV GOCMS_DEPLOYMENT_PROFILE=container

EXPOSE 8080

ENTRYPOINT ["/gocms"]
