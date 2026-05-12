.PHONY: run test fmt build verify deploy docker-build compose-config kics docker-ps docker-logs docker-stop

run:
	go run ./cmd/server

test:
	go test ./...

fmt:
	gofmt -w ./cmd ./internal

build:
	go tool templ generate ./...
	npm run vendor:assets
	npm run build:css
	npm run build:versioned
	go build ./cmd/server

verify:
	npm ci
	npm run verify

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

deploy:
	mkdir -p data
	docker compose up -d --build cms

docker-ps:
	docker compose ps

docker-logs:
	docker compose logs --tail=100 -f cms

docker-stop:
	docker compose down
