test-verbose:
	go test -v ./...

test:
	go test ./...

run:
	go run ./cmd/redis-server/main.go

build:
	go build -o ./bin/redis-server/main ./cmd/redis-server/main.go 