GO ?= go
BINARY ?= cerne
GOFILES := $(shell git ls-files '*.go')

GOBIN_VALUE := $(shell $(GO) env GOBIN)
GOPATH_VALUE := $(shell $(GO) env GOPATH)
INSTALL_DIR := $(if $(GOBIN_VALUE),$(GOBIN_VALUE),$(GOPATH_VALUE)/bin)

.PHONY: build install-local test test-fresh vet fmt check install-path

build:
	$(GO) build -o $(BINARY) ./cmd/cerne

install-local:
	$(GO) install ./cmd/cerne
	@printf 'installed: %s/%s\n' '$(INSTALL_DIR)' '$(BINARY)'

test:
	$(GO) test ./...

test-fresh:
	$(GO) test -count=1 ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w $(GOFILES)

check: test vet

install-path:
	@printf '%s/%s\n' '$(INSTALL_DIR)' '$(BINARY)'
