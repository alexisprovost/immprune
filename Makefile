APP := immprune
CMD_PATH := ./cmd/immprune
DIST_DIR := dist

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make tidy      - go mod tidy"
	@echo "  make fmt       - go fmt ./..."
	@echo "  make vet       - go vet ./..."
	@echo "  make test      - go test ./..."
	@echo "  make build     - build local binary in ./bin"
	@echo "  make install   - install binary in GOBIN/GOPATH/bin"
	@echo "  make dist      - cross-build binaries in ./dist"
	@echo "  make clean     - remove build artifacts"
	@echo "  make ci        - tidy + fmt + vet + test + build"

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: test
test:
	go test ./...

.PHONY: build
build:
	mkdir -p bin
	go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(APP) $(CMD_PATH)

.PHONY: install
install:
	go install -trimpath -ldflags "$(LDFLAGS)" $(CMD_PATH)

.PHONY: dist
dist: clean
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-linux-amd64 $(CMD_PATH)
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-linux-arm64 $(CMD_PATH)
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-darwin-amd64 $(CMD_PATH)
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-darwin-arm64 $(CMD_PATH)
	GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-windows-amd64.exe $(CMD_PATH)
	GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "$(LDFLAGS)" -o $(DIST_DIR)/$(APP)-windows-arm64.exe $(CMD_PATH)

.PHONY: clean
clean:
	rm -rf bin $(DIST_DIR)

.PHONY: ci
ci: tidy fmt vet test build