.PHONY: build run test

build:
	@go build -o bin/importer cmd/app/main.go

run: build
	@./bin/importer

test:
	@go test -v ./...
