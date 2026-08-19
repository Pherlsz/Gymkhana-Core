SHELL := /bin/sh

GO ?= go
BIN_DIR := $(CURDIR)/bin
GOEXE := $(shell $(GO) env GOEXE 2>/dev/null)
STATICCHECK := $(BIN_DIR)/staticcheck$(GOEXE)
GOVULNCHECK := $(BIN_DIR)/govulncheck$(GOEXE)
OSV_SCANNER := $(BIN_DIR)/osv-scanner$(GOEXE)

STATICCHECK_VERSION := v0.7.0
GOVULNCHECK_VERSION := v1.6.0
OSV_SCANNER_VERSION := v2.4.0

.PHONY: setup setup-quality setup-security format format-check vet lint test conformance test-race fuzz-smoke vuln osv security check clean

setup: setup-quality setup-security

setup-quality:
	@mkdir -p "$(BIN_DIR)"
	@GOBIN="$(BIN_DIR)" $(GO) install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)

setup-security:
	@mkdir -p "$(BIN_DIR)"
	@GOBIN="$(BIN_DIR)" $(GO) install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@GOBIN="$(BIN_DIR)" $(GO) install github.com/google/osv-scanner/v2/cmd/osv-scanner@$(OSV_SCANNER_VERSION)

format:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -n "$$files" ]; then gofmt -w $$files; fi

format-check:
	@files="$$(find . -type f -name '*.go' -not -path './vendor/*')"; \
	if [ -z "$$files" ]; then \
		echo "No Go files yet; skipping gofmt check."; \
		exit 0; \
	fi; \
	unformatted="$$(gofmt -l $$files)"; \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

vet:
	@$(GO) vet ./...

lint:
	@"$(STATICCHECK)" ./...

test:
	@$(GO) test ./...

conformance:
	@$(GO) test ./conformance

test-race:
	@$(GO) test -race ./...

fuzz-smoke:
	@$(GO) test -run='^$$' -fuzz='^FuzzSearchText$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzCanonicalCPF$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzCanonicalCNPJ$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzCanonicalDocument$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzIdentifyDocument$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzFormatAddressSlots$$' -fuzztime=2s -timeout=30s -parallel=1 ./normalize
	@$(GO) test -run='^$$' -fuzz='^FuzzParseCivilDate$$' -fuzztime=2s -timeout=30s -parallel=1 ./civiltime
	@$(GO) test -run='^$$' -fuzz='^FuzzParseYearMonth$$' -fuzztime=2s -timeout=30s -parallel=1 ./civiltime
	@$(GO) test -run='^$$' -fuzz='^FuzzFingerprintParse$$' -fuzztime=2s -timeout=30s -parallel=1 ./fingerprint
	@$(GO) test -run='^$$' -fuzz='^FuzzAssistantMessageJSON$$' -fuzztime=2s -timeout=30s -parallel=1 ./assistant

vuln:
	@"$(GOVULNCHECK)" ./...

osv:
	@"$(OSV_SCANNER)" scan source --recursive .

security: setup-security vuln osv

check: setup format-check vet lint test test-race fuzz-smoke vuln osv

clean:
	@rm -rf "$(BIN_DIR)" coverage
