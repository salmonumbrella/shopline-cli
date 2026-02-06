.PHONY: build test lint fmt ci setup clean docs docs-man docs-markdown coverage cli-coverage

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := -s -w \
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)

build:
	go build -ldflags="$(LDFLAGS)" -o bin/shopline ./cmd/shopline

test:
	go test -race -cover ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w .
	goimports -w .

fmt-check:
	@test -z "$$(gofumpt -l .)" || (echo "Run 'make fmt'" && exit 1)

ci: fmt-check lint test

setup:
	@command -v golangci-lint >/dev/null || go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@command -v gofumpt >/dev/null || go install mvdan.cc/gofumpt@latest
	@command -v goimports >/dev/null || go install golang.org/x/tools/cmd/goimports@latest

clean:
	rm -rf bin/ dist/

install: build
	cp bin/shopline $(GOPATH)/bin/

docs-man:
	go run ./cmd/shopline docs man ./man

docs-markdown:
	go run ./cmd/shopline docs markdown ./docs/cli

docs: docs-man docs-markdown

coverage:
	go run ./cmd/shopline-coverage

cli-coverage:
	go run ./cmd/shopline-cli-coverage
