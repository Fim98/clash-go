# clash-go — development tasks. `make` alone lists them.
SHELL := /bin/sh
GO ?= go
MYGO := $(GO) tool mygo

# The targets `make build` packages, comma separated `GOOS/GOARCH`.
PLATFORMS ?= darwin/arm64,darwin/amd64,windows/amd64,linux/amd64

.DEFAULT_GOAL := help
.PHONY: help setup fmt vet test check dev build clean

help: ## List the targets
	@grep -hE '^[a-zA-Z_-]+:.*?## ' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-8s\033[0m %s\n", $$1, $$2}'

setup: ## Configure this clone: git hooks and Go modules
	git config core.hooksPath .githooks
	@chmod +x .githooks/* scripts/*.sh 2>/dev/null || true
	$(GO) mod download
	@echo "git hooks: $$(git config --get core.hooksPath)"

fmt: ## Format the Go sources
	gofmt -w .

vet: ## go vet, then mygo's own checks
	$(GO) vet ./...
	$(MYGO) vet .

test: ## Run the tests
	CGO_ENABLED=0 $(GO) test ./...

check: fmt vet test ## What CI runs

dev: ## Run the app, rebuilding as sources change
	$(MYGO) dev

build: ## Package the app for $(PLATFORMS)
	$(MYGO) build -platform "$(PLATFORMS)"

clean: ## Remove build output
	rm -rf build