.PHONY: help build test lint grpc

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the netdebug binary
	go build ./cmd/netdebug

test: ## Run the test suite
	go test ./...

lint: ## Run golangci-lint over the module
	./bin/golangci-lint run ./...

grpc: ## Regenerate gRPC code from .proto files
	PATH="$(CURDIR)/bin:$$PATH" protoc \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		./pkg/echo/v1/echo.proto
