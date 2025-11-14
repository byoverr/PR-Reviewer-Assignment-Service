migrate-up:
	goose -dir migrations postgres "postgres://user:pass@localhost:5432/pr_service?sslmode=disable" up

build:
	go build -o bin/main cmd/main.go

run:
	go run cmd/main.go

test:
	go test ./...

docker-up:
	docker-compose up --build