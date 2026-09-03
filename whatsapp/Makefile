.PHONY: run build fmt vet tidy

run:
	go run ./cmd/server

build:
	go build -o bin/whatsapp ./cmd/server

fmt:
	gofmt -w cmd internal

vet:
	go vet ./...

tidy:
	go mod tidy

test:
	go test ./... -count=1
