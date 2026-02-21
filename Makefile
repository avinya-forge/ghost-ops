.PHONY: build test lint run clean help

# Binary name
BINARY_NAME=ghost-ops

build: ## Build the binary
	go build -ldflags "-X main.Version=$(shell cat VERSION)" -o $(BINARY_NAME) cmd/ghost-ops/main.go

test: ## Run tests
	go test -v ./...

lint: ## Run linter (go vet)
	go vet ./...

run: build ## Run the application (Mock mode)
	./$(BINARY_NAME) -engine mock -llm mock

clean: ## Clean up binary
	rm -f $(BINARY_NAME)
	rm -f *.wasm

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
