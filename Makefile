.PHONY: run build migrate-up migrate-down

run:
	go run ./cmd/goauction

build:
	go build -o bin/goauction ./cmd/goauction

migrate-up:
	migrate -database ${DATABASE_URL} -path migrations up

migrate-down:
	migrate -database ${DATABASE_URL} -path migrations down