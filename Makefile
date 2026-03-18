SHELL := /bin/bash

TOOLS_BIN := $(CURDIR)/.bin
GOLANGCI_LINT := $(TOOLS_BIN)/golangci-lint
GOIMPORTS := $(TOOLS_BIN)/goimports
GOFUMPT := $(TOOLS_BIN)/gofumpt

export PATH := $(TOOLS_BIN):$(PATH)

.PHONY: bootstrap build check fmt fmt-check install-hooks lint test tools

bootstrap: tools install-hooks

build:
	mkdir -p dist
	go build -o dist/capsule ./cmd/cli

check: fmt-check lint test

fmt: tools
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*')"; \
	if [[ -z "$$files" ]]; then \
		exit 0; \
	fi; \
	$(GOIMPORTS) -w $$files; \
	$(GOFUMPT) -w $$files

fmt-check: tools
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*' -not -path './.git/*')"; \
	if [[ -z "$$files" ]]; then \
		exit 0; \
	fi; \
	unformatted="$$( { $(GOIMPORTS) -l $$files; $(GOFUMPT) -l $$files; } | sort -u )"; \
	if [[ -n "$$unformatted" ]]; then \
		echo "Unformatted files:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

install-hooks:
	git config core.hooksPath .githooks

lint: tools
	$(GOLANGCI_LINT) run ./...

test:
	go test ./...

tools: $(GOLANGCI_LINT) $(GOIMPORTS) $(GOFUMPT)

$(TOOLS_BIN):
	mkdir -p $(TOOLS_BIN)

$(GOLANGCI_LINT): | $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8

$(GOIMPORTS): | $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install golang.org/x/tools/cmd/goimports@v0.31.0

$(GOFUMPT): | $(TOOLS_BIN)
	GOBIN=$(TOOLS_BIN) go install mvdan.cc/gofumpt@v0.8.0
