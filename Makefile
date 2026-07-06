# Ren Browser — Makefile alternative to Taskfile (https://taskfile.dev).
# Common workflows work without installing `task`.

SHELL := /bin/bash
.PHONY: help check format test test-go frontend-check frontend-test frontend-fmt \
	build-server screenshots dev clean vendor mod-tidy

export GOFLAGS ?= -mod=vendor
export PNPM ?= pnpm
APP_NAME ?= renbrowser
BIN_DIR ?= bin
VERSION := $(shell grep '^version:' build/brand.yml | sed 's/version: *//' | tr -d '"')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)
SERVER_BIN := $(BIN_DIR)/$(APP_NAME)-server
LDFLAGS := -w -s -X renbrowser/internal/buildinfo.Version=$(VERSION) -X renbrowser/internal/buildinfo.Commit=$(GIT_COMMIT)

help:
	@printf '%s\n' \
		"Targets:" \
		"  make check            Run Go and frontend quality gates" \
		"  make format           Format Go and frontend sources" \
		"  make test             Run Go and frontend unit tests" \
		"  make build-server     Build headless server binary" \
		"  make screenshots      Regenerate README preview images" \
		"  make dev              Run Wails dev mode" \
		"  make vendor           Refresh vendor/ from go.mod" \
		"  make mod-tidy         go mod tidy + vendor refresh"

check: fmt-go test-go gosec frontend-check

format: fmt-go frontend-fmt

test: test-go frontend-test

fmt-go:
	bash build/scripts/gofmt-project.sh write

test-go:
	go test ./...

gosec:
	bash build/scripts/gosec.sh ./...

frontend-install:
	$(PNPM) install --dir frontend

frontend-check: frontend-install
	$(PNPM) --dir frontend run check
	$(PNPM) --dir frontend run lint
	$(PNPM) --dir frontend run format:check
	$(PNPM) --dir frontend run knip
	$(PNPM) --dir frontend run audit
	$(PNPM) --dir frontend run test

frontend-test: frontend-install
	$(PNPM) --dir frontend run test

frontend-fmt: frontend-install
	$(PNPM) --dir frontend run format

build-frontend: frontend-install
	$(PNPM) --dir frontend run build

build-server: build-frontend
	bash build/scripts/patch-wails-vendor.sh
	CGO_ENABLED=0 go build -tags server,production -trimpath -buildvcs=false \
		-ldflags="$(LDFLAGS)" -o "$(SERVER_BIN)"

screenshots: build-server
	$(PNPM) --dir frontend exec playwright install chromium
	$(PNPM) --dir frontend run screenshots

dev:
	wails3 dev -config ./build/config.yml

vendor:
	bash build/scripts/vendor-go.sh

mod-tidy:
	go mod tidy
	$(MAKE) vendor

clean:
	rm -rf frontend/dist frontend/node_modules bin/$(APP_NAME) bin/$(APP_NAME)-server
