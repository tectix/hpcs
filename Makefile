.PHONY: build test clean run-server test-client benchmark load-test performance-test deps fmt vet lint

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary names
SERVER_BINARY=bin/hpcs-server

# Build flags
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"
VERSION ?= dev
BUILD_TIME ?= $(shell date -u '+%Y-%m-%d_%H:%M:%S')

all: test build

build: build-server

build-server:
	$(GOBUILD) $(LDFLAGS) -o $(SERVER_BINARY) ./cmd/hpcs-server

test:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

benchmark:
	$(GOTEST) -bench=. -benchmem ./internal/cache/

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

run-server:
	$(GOBUILD) -o $(SERVER_BINARY) ./cmd/hpcs-server
	./$(SERVER_BINARY) --config configs/config.yaml

test-client:
	go run test/simple_client.go

test-integration: build
	./$(SERVER_BINARY) --config configs/config.yaml &
	sleep 2
	go run test/simple_client.go
	pkill -f hpcs-server || true

load-test: build
	./$(SERVER_BINARY) --config configs/config.yaml &
	sleep 2
	go run test/loadtest/main.go -connections=50 -duration=30s
	pkill -f hpcs-server || true

performance-test: benchmark load-test

deps:
	$(GOMOD) download
	$(GOMOD) tidy

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOCMD) vet ./...

lint:
	$(GOLINT) run