.PHONY: run test fmt build deploy docker-ps docker-logs docker-stop

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

deploy:
	docker compose up -d --build cms

docker-ps:
	docker compose ps

docker-logs:
	docker compose logs --tail=100 -f cms

docker-stop:
	docker compose down
