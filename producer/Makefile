MAKEFILE_DIR := $(dir $(lastword $(MAKEFILE_LIST)))

.PHONY: sqlc
sqlc: 
	sqlc generate --file=$(MAKEFILE_DIR)pkg/sqlc/sqlc.yaml
