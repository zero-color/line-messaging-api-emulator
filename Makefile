.DEFAULT_GOAL := help

GO_TEST ?= go tool gotestsum --

.PHONY: test
test: ## Run tests
	$(GO_TEST) -race ./...

.PHONY: lint
lint: ## Run linters
	go tool golangci-lint run ./...

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
