.PHONY: build run test test-all test-cover bench bench-cover lint migrate-up docker-up docker-down

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

COVERPKG=./internal/handlers/...,./internal/services/...,./internal/repository/...,./internal/models/...,./internal/config/...,./internal/logger/...,./internal/apperrors/...

test-cover:
	docker-compose -f docker-compose-e2e.yml up -d || true
	sleep 5
	go test -v -coverprofile=coverage-all.out -coverpkg=$(COVERPKG) ./tests/... || true
	docker-compose -f docker-compose-e2e.yml down || true
	@go tool cover -html=coverage-all.out -o coverage-all.html 2>/dev/null || true
	@go tool cover -func=coverage-all.out | tail -1

bench:
	go test -bench=. -benchmem ./tests/benchmark/...

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

