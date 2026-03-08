VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-X main.version=$(VERSION)"
BINARY := floxybot
GOFILES := $(shell find . -name '*.go' -not -path './vendor/*')

.PHONY: build test lint vet clean

build: vet
	go build $(LDFLAGS) -o $(BINARY) ./cmd/floxybot

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed, skipping"

clean:
	rm -f $(BINARY)
