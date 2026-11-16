GO ?= go
BIN := bin/server
VERSION ?= 1.0.0
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: dev build run test cover bench fmt vet tidy docker-build docker-run clean

dev:           ## run with demo data on :8080
	SEED_DEMO=true $(GO) run ./cmd/server

build:
	mkdir -p bin
	CGO_ENABLED=1 $(GO) build -trimpath -ldflags "$(LDFLAGS)" -o $(BIN) ./cmd/server

run: build
	SEED_DEMO=true DATABASE_PATH=./app.db ./$(BIN)

test:
	$(GO) test ./...

cover:
	$(GO) test -coverpkg=./internal/...,./pkg/... -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -1

bench:
	$(GO) run ./cmd/bench -n 500 -c 8 -out docs/bench-results.json

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './vendor/*')

vet:
	$(GO) vet ./...

tidy:
	$(GO) mod tidy

docker-build:
	docker build -t go-service-template .

docker-run:
	docker run --rm -p 8080:8080 -e SEED_DEMO=true go-service-template

clean:
	rm -rf bin app.db app.db-shm app.db-wal coverage.out
