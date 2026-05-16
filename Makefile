.PHONY: build test lint clean

VERSION ?= dev
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/agent-kay-it/tene/internal/cli.version=$(VERSION) \
           -X github.com/agent-kay-it/tene/internal/cli.commit=$(COMMIT) \
           -X github.com/agent-kay-it/tene/internal/cli.date=$(DATE)

build:
	CGO_ENABLED=0 go build -ldflags "$(LDFLAGS)" -o bin/tene ./cmd/tene

test:
	go test ./... -v -count=1

lint:
	golangci-lint run ./...

clean:
	rm -rf bin/

vet:
	go vet ./...

all: lint vet test build
