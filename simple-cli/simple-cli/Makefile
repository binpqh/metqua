# Makefile — simple-cli developer workflow
# Constitution Principle X: standard targets for all contributors.

BINARY     := simple-cli
DIST_DIR   := dist
MODULE     := github.com/your-org/simple-cli
VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "unknown")
LDFLAGS    := -s -w \
              -X $(MODULE)/pkg/version.Version=$(VERSION) \
              -X $(MODULE)/pkg/version.Commit=$(COMMIT) \
              -X $(MODULE)/pkg/version.BuildDate=$(BUILD_DATE)

GOTEST     := go test
COVERAGE   := coverage.out
MIN_COV    := 80

.PHONY: all build test lint install-local test-sandbox clean

all: build

## build: Compile static binary to dist/
build:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(DIST_DIR)/$(BINARY) ./cmd/...

## test: Run all unit + integration tests and enforce ≥80% coverage
test:
	$(GOTEST) ./internal/... ./pkg/... -coverprofile=$(COVERAGE) -covermode=atomic
	@go tool cover -func=$(COVERAGE) | grep 'total:' | \
		awk '{gsub(/%/,""); if ($$3 < $(MIN_COV)) \
			{print "FAIL: coverage " $$3 "% is below minimum $(MIN_COV)%"; exit 1} \
			else {print "PASS: coverage " $$3 "%"}}'

## lint: Run golangci-lint (required to pass before merge)
lint:
	golangci-lint run ./...

## install-local: Copy binary to GOPATH/bin for local dogfooding
install-local: build
	cp $(DIST_DIR)/$(BINARY) $(shell go env GOPATH)/bin/$(BINARY)
	@echo "Installed $(BINARY) to $(shell go env GOPATH)/bin/"

## test-sandbox: Run installer tests in Docker containers
## Windows sandbox requires SKIP_WINDOWS_SANDBOX != true and a Windows CI runner
test-sandbox:
	docker compose -f tests/sandbox/docker-compose.yml up \
		--build --abort-on-container-exit --exit-code-from ubuntu
ifndef SKIP_WINDOWS_SANDBOX
	@echo "Windows sandbox: set SKIP_WINDOWS_SANDBOX=true to skip on Linux CI"
endif

## clean: Remove build artifacts
clean:
	rm -rf $(DIST_DIR) $(COVERAGE)
