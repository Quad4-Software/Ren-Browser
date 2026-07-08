# Ren Browser — Makefile alternative to Taskfile (https://taskfile.dev).
# Common workflows work without installing `task`.

SHELL := /bin/bash
.PHONY: help check format test test-go test-regression frontend-check frontend-test frontend-fmt \
	build build-frontend build-server screenshots e2e coverage-go server-smoke dev clean vendor mod-tidy gosec sbom

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
		"  make e2e              Playwright E2E against renbrowser-server" \
		"  make coverage-go      Write Go coverage under coverage/" \
		"  make server-smoke     Start server binary and hit HTTP" \
		"  make build            Build desktop binary (GTK4/WebKitGTK)" \
		"  make build-server     Build headless server binary" \
		"  make screenshots      Regenerate README preview images" \
		"  make sbom             Generate SPDX and CycloneDX SBOMs (requires Trivy)" \
		"  make dev              Run Wails dev mode" \
		"  make vendor           Refresh vendor/ from go.mod" \
		"  make mod-tidy         go mod tidy + vendor refresh"

check: fmt-go test-go gosec frontend-check

format: fmt-go frontend-fmt

test: test-go frontend-test

test-regression:
	go test ./internal/regression/... ./internal/cache/... ./internal/nomadnet/... ./internal/store/... ./internal/app/... ./internal/servermw/... ./internal/limits/... ./internal/apperrors/... ./internal/db/... ./internal/plugins/... ./internal/content/... ./internal/safe/...
	go test -race -tags=stress -timeout=5m ./internal/cache/...

fmt-go:
	bash build/scripts/gofmt-project.sh write

test-go:
	go test ./...

gosec:
	bash build/scripts/gosec.sh ./...

sbom:
	@command -v trivy >/dev/null 2>&1 || { echo "Trivy not found. Install with: sh build/scripts/ci/setup-trivy.sh 0.69.3" >&2; exit 1; }
	mkdir -p sbom
	trivy fs --skip-dirs vendor --skip-dirs frontend/node_modules --skip-dirs bin --skip-dirs .cache --skip-dirs third_party --skip-dirs build/android/.gradle --skip-dirs build/android/build --format spdx-json --include-dev-deps --output sbom/sbom.spdx.json .
	trivy fs --skip-dirs vendor --skip-dirs frontend/node_modules --skip-dirs bin --skip-dirs .cache --skip-dirs third_party --skip-dirs build/android/.gradle --skip-dirs build/android/build --format cyclonedx --include-dev-deps --output sbom/sbom.cyclonedx.json .
	@echo 'SBOM files generated in sbom/ directory'

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

build: build-frontend
	bash build/scripts/patch-wails-vendor.sh
	CGO_ENABLED=1 go build -tags production -trimpath -buildvcs=false \
		-ldflags="$(LDFLAGS)" -o "$(BIN_DIR)/$(APP_NAME)"

build-server: build-frontend
	bash build/scripts/patch-wails-vendor.sh
	CGO_ENABLED=0 go build -tags server,production -trimpath -buildvcs=false \
		-ldflags="$(LDFLAGS)" -o "$(SERVER_BIN)"

screenshots: build-server
	$(PNPM) --dir frontend exec playwright install chromium
	$(PNPM) --dir frontend run screenshots

e2e: build-server
	$(PNPM) --dir frontend exec playwright install chromium
	$(PNPM) --dir frontend run test:e2e

coverage-go:
	bash build/scripts/coverage-go.sh

server-smoke: build-server
	bash build/scripts/server-smoke.sh

dev:
	wails3 dev -config ./build/config.yml

vendor:
	bash build/scripts/vendor-go.sh

mod-tidy:
	go mod tidy
	$(MAKE) vendor

clean:
	rm -rf frontend/dist frontend/node_modules bin/$(APP_NAME) bin/$(APP_NAME)-server
