.PHONY: build run test test-all lint migrate-up docker-up docker-down

build:
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

test:
	go test -v ./tests/unit/...

test-all:
	docker-compose -f docker-compose-e2e.yml up -d
	sleep 5
	go test -v ./tests/...
	docker-compose -f docker-compose-e2e.yml down

lint:
	golangci-lint run

migrate-up:
	goose -dir migrations postgres "postgres://user:pass@localhost:5432/pr_service?sslmode=disable" up

migrate-down:
	goose -dir migrations postgres "postgres://user:pass@localhost:5432/pr_service?sslmode=disable" down

docker-up:
	docker-compose up -d --build

docker-down:
	docker-compose down

