.PHONY: run build test lint migrate-up migrate-down docker-build

run:
	go run ./cmd/goauction

build:
	go build -o bin/goauction ./cmd/goauction

test:
	go test -v -race ./...

lint:
	golangci-lint run

migrate-up:
	migrate -database ${DATABASE_URL} -path migrations up

migrate-down:
	migrate -database ${DATABASE_URL} -path migrations down

docker-build:
	docker build -t goauction:latest .