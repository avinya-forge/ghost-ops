.PHONY: build test lint run clean help hooks dev fmt vet staticcheck gosec govulncheck coverage audit demo

BINARY_NAME=ghost-ops
COVERAGE_THRESHOLD?=60
GOBIN=$(shell go env GOPATH)/bin

build: ## Build the binary
	go build -ldflags "-X main.Version=$(shell cat VERSION)" -o $(BINARY_NAME) cmd/ghost-ops/main.go

test: ## Run tests with race detector
	go test -race -count=1 ./...

lint: ## Run vet + staticcheck + gofmt check
	@$(MAKE) --no-print-directory fmt
	@$(MAKE) --no-print-directory vet
	@$(MAKE) --no-print-directory staticcheck

fmt: ## Verify gofmt produces no diffs
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then echo "gofmt issues:"; echo "$$out"; exit 1; fi

vet: ## go vet
	go vet ./...

staticcheck: ## Run staticcheck (auto-installs if missing)
	@test -x $(GOBIN)/staticcheck || go install honnef.co/go/tools/cmd/staticcheck@latest
	$(GOBIN)/staticcheck ./...

gosec: ## Run gosec (high severity, high confidence)
	@test -x $(GOBIN)/gosec || go install github.com/securego/gosec/v2/cmd/gosec@latest
	$(GOBIN)/gosec -severity high -confidence high -no-fail ./...

govulncheck: ## Run govulncheck (advisory until BUG-024 lands)
	@test -x $(GOBIN)/govulncheck || go install golang.org/x/vuln/cmd/govulncheck@latest
	-$(GOBIN)/govulncheck ./...

coverage: ## Run tests with coverage; enforce $(COVERAGE_THRESHOLD)%
	go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...
	@pct=$$(go tool cover -func=coverage.out | tail -1 | awk '{print $$3}' | tr -d '%'); \
	echo "Total coverage: $$pct% (threshold $(COVERAGE_THRESHOLD)%)"; \
	awk -v p="$$pct" -v t=$(COVERAGE_THRESHOLD) 'BEGIN { if (p+0 < t+0) { exit 1 } }'

audit: ## Adversarial Triad: lint + gosec + govulncheck + coverage
	@$(MAKE) --no-print-directory lint
	@$(MAKE) --no-print-directory gosec
	@$(MAKE) --no-print-directory govulncheck
	@$(MAKE) --no-print-directory coverage

run: build ## Run the application (Mock mode)
	./$(BINARY_NAME) -engine mock -llm mock

demo: build ## Run the stakeholder demo (D20+) — see docs/BACKLOG.md §5
	@echo "Demo target stub — wired in D20 (TASK-1100)."
	@exit 1

clean: ## Clean up binary + coverage
	rm -f $(BINARY_NAME) ghost-dev coverage.out
	rm -f *.wasm

hooks: ## Install git pre-commit hooks
	git config core.hooksPath .githooks

dev: ## Run with hot reload
	go run cmd/ghost-dev/main.go

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
