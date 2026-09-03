ifneq (,$(wildcard .env))
include .env
export
endif

MIGRATE=migrate -path db/migrations -database "$(DATABASE_URL)"
.PHONY: run worker build test test-integration lint fmt sqlc jwt migrate-up migrate-down docker-up docker-down
run:
	@if [ -n "$(RATE_LIMIT_MAX)" ]; then \
		RATE_LIMIT_MAX=$(RATE_LIMIT_MAX) go run ./cmd/server; \
	else \
		go run ./cmd/server; \
	fi
worker:
	go run ./cmd/worker
build:
	go build ./...
test:
	go test ./...
test-integration:
	TEST_INTEGRATION=1 go test ./... -count=1
lint:
	golangci-lint run
fmt:
	gofmt -w $$(find . -name '*.go')
sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@latest generate
jwt:
	@test -f .env || cp .env.example .env
	@JWT_SECRET=$$(openssl rand -hex 32); \
		python3 -c 'import pathlib, sys; p = pathlib.Path(".env"); lines = p.read_text().splitlines(); key = "JWT_SECRET="; lines = [key + sys.argv[1] if line.startswith(key) else line for line in lines]; p.write_text("\n".join(lines) + "\n")' "$$JWT_SECRET"
	@echo "Generated a random JWT_SECRET in .env"
migrate-up:
	@$(MIGRATE) up
migrate-down:
	@$(MIGRATE) down 1
docker-up:
	docker compose up -d postgres redis
docker-down:
	docker compose down
