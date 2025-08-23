.DEFAULT_GOAL := help

GO_TEST ?= go tool gotestsum --

.PHONY: run
run: ## Run the application
	go run ./cmd/server

.PHONY: test
test: ## Run tests
	$(GO_TEST) -race ./...

.PHONY: lint
lint: ## Run linters
	go tool golangci-lint run ./...

.PHONY: generate
generate: ## Run go generate
	go generate ./...

.PHONY: db-migrate
db-migrate: ## Run database migrations
	cd db && atlas schema apply --env local && sqlc generate

.PHONY: help
help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
