SHELL := /usr/bin/env bash -o errexit -o pipefail -o nounset
MAKEFLAGS := --warn-undefined-variables --no-builtin-rules

OUT_DIR = out
BIN_DIR = $(OUT_DIR)/bin
REPORT_DIR = $(OUT_DIR)/report
VERSION = $(shell git tag --list --sort=-version:refname "v*" | head -n 1 | grep "." || echo "v0.0.0")

COLORS ?= true
RED    := $(if $(findstring $(COLORS),true),$(shell tput -Txterm setaf 1))
GREEN  := $(if $(findstring $(COLORS),true),$(shell tput -Txterm setaf 2))
YELLOW := $(if $(findstring $(COLORS),true),$(shell tput -Txterm setaf 3))
WHITE  := $(if $(findstring $(COLORS),true),$(shell tput -Txterm setaf 7))
CYAN   := $(if $(findstring $(COLORS),true),$(shell tput -Txterm setaf 6))
RESET  := $(if $(findstring $(COLORS),true),$(shell tput -Txterm sgr0))

# Tools
GOFUMPT_CMD = go run mvdan.cc/gofumpt@v0.4.0
GOJUNITREP_CMD = go run github.com/jstemmer/go-junit-report/v2@v2.0.0
GOWEIGHT_CMD = go run github.com/jondot/goweight@latest
GOMODOUTDATED_CMD = go run github.com/psampaz/go-mod-outdated@latest
# Dockerized tools
GOLANGCI_LINT_VERSION = v1.51.2

define task
 @echo "${CYAN}>>> $(1)${RESET}"
endef

.PHONY: all
all: clean lint test build

## Build:
.PHONY: build
build: ## Build project
	$(call task,build)
	@go build ./...

.PHONY: ci
ci: clean lint coverage build ## Build project for CI (clean, lint, coverage, build)

.PHONY: clean
clean: ## Remove build related files (default ./out)
	$(call task,clean)
	@rm -fr $(OUT_DIR)

## Test:
.PHONY: test
test: ## Run tests
	$(call task,test)
	@rm -rf $(REPORT_DIR)/test
	@mkdir -p $(REPORT_DIR)/test
	@go test -v -race ./... \
		| tee >($(GOJUNITREP_CMD) -set-exit-code > $(REPORT_DIR)/test/junit-report.xml)

.PHONY: coverage
coverage: ## Run tests and create coverage report
	$(call task,coverage)
	@rm -rf $(REPORT_DIR)/test
	@mkdir -p $(REPORT_DIR)/test
	@go test -cover -covermode=atomic -coverpkg=./... -coverprofile=$(REPORT_DIR)/test/coverage.out -v ./... \
		| tee >($(GOJUNITREP_CMD) -set-exit-code > $(REPORT_DIR)/test/junit-report.xml)
	@go tool cover -html=$(REPORT_DIR)/test/coverage.out -o $(REPORT_DIR)/test/coverage.html

.PHONY: bench
bench: ## Run benchmark tests
	go test -benchmem -count 3 -bench ./...

## Lint:
.PHONY: lint
lint: ## Lint go source files
	$(call task,lint)
	@rm -f $(REPORT_DIR)/checktyle/format-go-*
	@rm -f $(REPORT_DIR)/checktyle/checkstyle-go.*
	@mkdir -p $(REPORT_DIR)/checkstyle
	@echo "Checking gofumpt"
ifneq ($(shell $(GOFUMPT_CMD) -l . | wc -l),0)
	@echo "${YELLOW}Detected unformatted code${RESET} (fix: make format)"
	@$(GOFUMPT_CMD) -l . | tee $(REPORT_DIR)/checkstyle/format-go-files.txt
	@$(GOFUMPT_CMD) -d . | tee $(REPORT_DIR)/checkstyle/format-go-diff.txt
	@exit 1
endif
	@echo "Checking golangci_lint"
	@docker run --rm -t -v $(shell pwd):/app -v ~/.cache/golangci-lint/$(GOLANGCI_LINT_VERSION):/root/.cache -w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint run \
		--deadline=65s \
		--out-format checkstyle:$(REPORT_DIR)/checkstyle/checkstyle-go.xml,colored-line-number \
		./...

.PHONY: format
format: ## Format source files
	$(call task,format)
	@$(GOFUMPT_CMD) -l -w .
	@go mod tidy -e
	@docker run --rm -t -v $(shell pwd):/app -v ~/.cache/golangci-lint/$(GOLANGCI_LINT_VERSION):/root/.cache -w /app \
		golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
		golangci-lint --fix \
		--deadline=65s \
		./...

## Meta:
.PHONY: outdated
outdated: ## Print outdated dependencies
	@go mod tidy
	@go list -u -m -json all | $(GOMODOUTDATED_CMD) -update -direct

.PHONY: weight
weight: ## Print info about package weight
	$(call task,weight)
	$(GOWEIGHT_CMD)

## Help:
.PHONY: help
help: ## Show this help
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} - run whole build process (clean, lint, test, build)'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET} - run single target'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} { \
		if (/^[a-zA-Z_-]+:.*?##.*$$/) {printf "    ${YELLOW}%-20s${GREEN}%s${RESET}\n", $$1, $$2} \
		else if (/^## .*$$/) {printf "  ${CYAN}%s${RESET}\n", substr($$1,4)} \
		}' $(MAKEFILE_LIST)

.PHONY: version
version: ## Print project version
	@echo '${VERSION}'
