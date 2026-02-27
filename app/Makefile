.PHONY: build run test test-cover lint clean

build:
	go build -o ./bin/server ./cmd/api

run:
	go run ./cmd/api

test:
	go test ./...

test-cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run

clean:
	rm -rf ./bin coverage.out coverage.html
