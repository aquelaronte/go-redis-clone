test-verbose:
	go test -v ./...

test:
	go test ./...

# Run the API server in dev mode with hot-reload on code changes (requires air: go install github.com/air-verse/air@latest)
dev:
	air

run:
	go run ./cmd/redis-server/main.go

build:
	go build -o ./bin/redis-server/main ./cmd/redis-server/main.go 

start:
	./bin/redis-server/main