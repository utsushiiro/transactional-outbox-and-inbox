MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

.PHONY: build
build:
	go build -o bin/server cmd/main.go

.PHONY: lint
lint:
	golangci-lint run -c $(MAKEFILE_DIR).golangci.yaml --fix

.PHONY: sqlc
sqlc: 
	sqlc generate --file=$(MAKEFILE_DIR)sqlc.yaml
