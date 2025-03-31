.PHONY: help build clean tests unit-tests unit-tests-cov version swagger go-build-deps
.DEFAULT_GOAL := help

DIST_DIR=./dist
DEVENV_DIR=./devenv
LOCAL_DEV_DIR=$(DEVENV_DIR)/local
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

clean: ## Cleanup
	@rm -rf $(DIST_DIR)
	@rm -rf $(BIN_DIR)
	@rm -rf $(LOCAL_DEV_DIR)

tests: unit-tests ## Run all tests

unit-tests: ## Run unit tests
	go test -v ./...

unit-tests-cov: ## Run unit tests with coverage
	go test -race -v -coverpkg=./internal/... -coverprofile=coverage.txt ./...
	go tool cover -html=coverage.txt -o coverage.html
