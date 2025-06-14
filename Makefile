NAME = srsc
COMMIT = $(shell git rev-parse --short HEAD)
TAGS ?= with_acme

GOHOSTOS = $(shell go env GOHOSTOS)
GOHOSTARCH = $(shell go env GOHOSTARCH)
VERSION=$(shell CGO_ENABLED=0 GOOS=$(GOHOSTOS) GOARCH=$(GOHOSTARCH) go run github.com/sagernet/sing-box/cmd/internal/read_tag@latest)

PARAMS = -v -trimpath -ldflags "-X 'github.com/sagernet/srsc/constant.Version=$(VERSION)' -s -w -buildid="
MAIN_PARAMS = $(PARAMS) -tags "$(TAGS)"
MAIN = ./cmd/srsc
PREFIX ?= $(shell go env GOPATH)

.PHONY: test release docs build

build:
	export GOTOOLCHAIN=local && \
	go build $(MAIN_PARAMS) $(MAIN)

ci_build:
	export GOTOOLCHAIN=local && \
	go build $(PARAMS) $(MAIN) && \
	go build $(MAIN_PARAMS) $(MAIN)

generate_completions:
	go run -v --tags "$(TAGS),generate,generate_completions" $(MAIN)

install:
	go build -o $(PREFIX)/bin/$(NAME) $(MAIN_PARAMS) $(MAIN)

fmt:
	@gofumpt -l -w .
	@gofmt -s -w .
	@gci write --custom-order -s standard -s "prefix(github.com/sagernet/)" -s "default" .

fmt_install:
	go install -v mvdan.cc/gofumpt@latest
	go install -v github.com/daixiang0/gci@latest

lint:
	GOOS=linux golangci-lint run ./...
	GOOS=android golangci-lint run ./...
	GOOS=windows golangci-lint run ./...
	GOOS=darwin golangci-lint run ./...
	GOOS=freebsd golangci-lint run ./...

lint_install:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@latest

docs:
	venv/bin/mkdocs serve

publish_docs:
	venv/bin/mkdocs gh-deploy -m "Update" --force --ignore-version --no-history

docs_install:
	python -m venv venv
	source ./venv/bin/activate && pip install --force-reinstall mkdocs-material=="9.*" mkdocs-static-i18n=="1.2.*"
