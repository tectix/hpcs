.PHONY: build test clean run-server run-client lint fmt vet deps benchmark docker

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
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
	$(GOTEST) -bench=. -benchmem ./...

clean:
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

run-server:
	$(GOBUILD) -o $(SERVER_BINARY) ./cmd/hpcs-server
	./$(SERVER_BINARY) --config configs/config.yaml

deps:
	$(GOMOD) download
	$(GOMOD) tidy

fmt:
	$(GOFMT) -s -w .

vet:
	$(GOCMD) vet ./...

lint:
	$(GOLINT) run

# Docker targets
docker-build:
	docker build -t hpcs:$(VERSION) .

docker-run:
	docker run -p 6379:6379 -p 8080:8080 hpcs:$(VERSION)

# Performance testing
load-test:
	go run test/loadtest/main.go

# Development helpers
dev-setup:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest

watch:
	air -c .air.toml