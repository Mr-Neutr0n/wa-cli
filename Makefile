BINARY     := wa
CMD_DIR    := .
BIN_DIR    := ./bin
PKG        := github.com/codebyoketch/wa-cli/internal/version

VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT     ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -X $(PKG).Version=$(VERSION) -X $(PKG).Commit=$(COMMIT) -X $(PKG).BuildDate=$(BUILD_DATE)

# -race needs cgo and a C compiler; run `make test TESTFLAGS=` without one.
TESTFLAGS ?= -race -cover

.PHONY: all build test lint fmt fmt-check vet verify clean run install

all: build

build:
	mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/$(BINARY) $(CMD_DIR)

run: build
	$(BIN_DIR)/$(BINARY)

install:
	go install -ldflags "$(LDFLAGS)" $(CMD_DIR)

test:
	go test $(TESTFLAGS) ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@files="$$(gofmt -l .)"; \
	if [ -n "$$files" ]; then echo "Not gofmt-formatted (run 'make fmt'):"; echo "$$files"; exit 1; fi

# verify is the check to pass before every commit: every package builds,
# vet and gofmt are clean, and the tests pass.
verify:
	go build ./...
	go vet ./...
	@$(MAKE) --no-print-directory fmt-check
	go test $(TESTFLAGS) ./...

lint:
	@command -v golangci-lint >/dev/null || { echo "golangci-lint is not installed: https://golangci-lint.run/welcome/install/"; exit 1; }
	golangci-lint run ./...

clean:
	rm -rf $(BIN_DIR)
