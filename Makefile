.PHONY: run build templ templ-watch test test-integration lint coverage docker-build tools

# Generate templ Go code from .templ files (runs before build/run)
templ:
	templ generate

# Hot-reload templ codegen during dev: re-runs on .templ changes
templ-watch:
	templ generate --watch

run: templ
	go run ./cmd/goauction

build: templ
	go build -o bin/goauction ./cmd/goauction

test: templ
	go test -v -race ./...

test-integration:
	go test -v -race -tags=integration ./internal/repository

lint:
	golangci-lint run

coverage: templ
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

docker-build:
	docker build -t goauction:latest .

# Install required dev tools (templ codegen)
tools:
	go install github.com/a-h/templ/cmd/templ@latest
