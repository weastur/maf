.PHONY: help build clean tests unit-tests unit-tests-cov
.DEFAULT_GOAL := help

BINARY_NAME=maf
BIN_DIR=./bin
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

build: ## Build the binary
	mkdir -p $(BIN_DIR)
	go build -ldflags "-X github.com/weastur/maf/cmd.version=v0.0.0-dev" -gcflags=all="-N -l" -o $(BIN_DIR)/$(BINARY_NAME)

clean: ## Cleanup
	rm -rf $(BIN_DIR)

tests: unit-tests ## Run all tests

unit-tests: ## Run unit tests
	go test -v ./...

unit-tests-cov: ## Run unit tests with coverage
	go test -v -coverprofile=coverage.txt ./...

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
