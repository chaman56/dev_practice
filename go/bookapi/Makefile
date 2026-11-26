.PHONY: run test lint build docker

run:
	go run ./cmd/api

test:
	go test -race ./...

lint:
	go vet ./...
	gofmt -l .

build:
	go build -o bin/api ./cmd/api

docker:
	docker build -t bookapi .
