MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

.PHONY: build
build:
	go build -o bin/server cmd/main.go

.PHONY: sqlc
sqlc: 
	sqlc generate --file=$(MAKEFILE_DIR)pkg/sqlc/sqlc.yaml
