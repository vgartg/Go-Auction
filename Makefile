.PHONY: run build test test-integration lint coverage docker-build

run:
	go run ./cmd/goauction

build:
	go build -o bin/goauction ./cmd/goauction

test:
	go test -v -race ./...

test-integration:
	go test -v -race -tags=integration ./internal/repository

lint:
	golangci-lint run

coverage:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker-build:
	docker build -t goauction:latest .