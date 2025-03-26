.PHONY: help build clean tests unit-tests unit-tests-cov version swagger go-build-deps
.DEFAULT_GOAL := help

BINARY_NAME=maf
BIN_DIR=./bin
DIST_DIR=./dist
DEVENV_DIR=./devenv
LOCAL_DEV_DIR=$(DEVENV_DIR)/local
ROOT_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))

build: swagger ## Build the binary
	@mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 go build -tags netgo,static_build,osusergo,feature -ldflags "-extldflags "-static" -X github.com/weastur/maf/internal/utils.version=v0.0.0-dev" -gcflags=all="-N -l" -o $(BIN_DIR)/$(BINARY_NAME)

generate: ## Generate code
	@PATH=$(ROOT_DIR)/gen:$(PATH) go generate ./...

swagger: go-build-deps generate ## Generate swagger docs
	@swag init --quiet --generalInfo v1alpha.go --dir internal/agent/worker/fiber/http/api/v1alpha,internal/utils/http --output internal/agent/worker/fiber/http/api/v1alpha --outputTypes json
	@swag init --quiet --generalInfo v1alpha.go --dir internal/server/worker/fiber/http/api/v1alpha,internal/utils/http --output internal/server/worker/fiber/http/api/v1alpha --outputTypes json

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

version: ## Create new version. Bump, tag, commit, create tag
	@bump-my-version bump --verbose $(filter-out $@,$(MAKECMDGOALS))

go-build-deps: ## Install go deps to build the project
	@go install github.com/swaggo/swag/cmd/swag@latest
	@go install github.com/weastur/replacer/cmd/replacer@v0.0.0-dev2

%:
	@:

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
