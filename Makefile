.PHONY: run test fmt build verify docker-build compose-config kics prepare-data deploy docker-ps docker-logs docker-stop

GOCMS_DATA_DIR ?= ./data

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

build:
	go tool templ generate ./...
	bun run vendor:assets
	bun run build:css
	bun run build:versioned
	go build ./cmd/server

verify:
	bun install --frozen-lockfile
	bun run verify

docker-build:
	docker build -t fastygo/cms:ci .

compose-config:
	docker compose config

kics:
	mkdir -p kics-results
	# On Docker Desktop for Windows, KICS may exit with "download not supported for scheme 'c'"; use WSL or rely on GitHub Actions.
	docker run --rm \
		-v "$(CURDIR):/scan:ro" \
		-v "$(CURDIR)/kics-results:/out" \
		checkmarx/kics:v2.1.20 scan \
		-p /scan/Dockerfile \
		-p /scan/docker-compose.yml \
		-o /out \
		--report-formats sarif,json \
		--silent \
		--disable-full-descriptions \
		--fail-on critical,high,medium,low,info

prepare-data:
	mkdir -p "$(GOCMS_DATA_DIR)"

deploy: prepare-data
	docker compose up --force-recreate --no-deps data-permissions
	docker compose up -d --build cms

docker-ps:
	docker compose ps

docker-logs:
	docker compose logs --tail=100 -f cms

docker-stop:
	docker compose down
