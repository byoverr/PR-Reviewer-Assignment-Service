# Stage 1: Builder
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download


RUN go install github.com/pressly/goose/v3/cmd/goose@v3.26.0

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /build/main ./cmd/main.go

# Stage 2: Runtime
FROM alpine:latest
RUN apk --no-cache add ca-certificates bash

WORKDIR /app

COPY --from=builder /build/main .
COPY --from=builder /go/bin/goose /usr/local/bin/goose
COPY migrations ./migrations

EXPOSE 8080

CMD ["sh", "-c", "goose -dir ./migrations postgres \"$DB_URL\" up && ./main"]