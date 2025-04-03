# Project metadata
MODULE   := $(shell go list -m)
DATE     := $(shell date +%FT%T%z)
VERSION  := $(shell git describe --tags --always --dirty --match=v* 2>/dev/null || echo v0)

# Paths and tools
BIN      := bin
TARGET   := cosmosign
SRC      := ./cmd/$(TARGET)/main.go
GO       := go
GOBIN    := $(shell $(GO) env GOPATH)/bin
DOCKER   ?= docker
PROTO_BUILDER_IMAGE := ghcr.io/cosmos/proto-builder:0.14.0

.PHONY: all
all: lint test

.PHONY: lint
lint:
	@echo "Running golangci-lint..."
	@$(GO) tool golangci-lint run ./...

.PHONY: fix
fix: 
	@echo "Running golangci-lint fix..."
	@$(GO) tool golangci-lint run --fix ./...

## Testing
.PHONY: test
test:
	@echo "Running tests..."
	@$(GO) test ./...

.PHONY: test-cover
test-cover:
	@echo "Running tests with coverage..."
	@$(GO) test -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html

## Mocks
.PHONY: generate-mocks
generate-mocks:
	@echo "Generating mocks..."
	@mockery

## Protobuf generation
.PHONY: proto-gen
proto-gen:
	@echo "Generating Protobuf files..."
	@$(DOCKER) run --rm -u 0 -v $(CURDIR):/workspace --workdir /workspace \
		$(PROTO_BUILDER_IMAGE) sh ./protocgen.sh

## Cleanup
.PHONY: clean
clean:
	@echo "Cleaning up..."
	@rm -rf $(BIN) coverage.out coverage.html

## Version
.PHONY: version
version:
	@echo $(VERSION)

## Help
.PHONY: help
help:
	@echo "Available commands:"
	@echo "  all             Run lint and tests"
	@echo "  lint            Run golangci-lint"
	@echo "  fix             Auto-fix lint issues"
	@echo "  test            Run unit tests"
	@echo "  test-cover      Run tests with coverage report"
	@echo "  generate-mocks  Generate mocks using mockery"
	@echo "  proto-gen       Generate protobuf files"
	@echo "  clean           Remove build artifacts"
	@echo "  version         Show project version"
	@echo "  help            Show this help message"
