# Get the currently used golang install path
GOPATH ?= $(shell go env GOPATH)
GOBIN ?= $(GOPATH)/bin

# Tool versions
CONTROLLER_GEN_VERSION ?= v0.19.0
GOLANGCI_LINT_VERSION ?= v2.8.0

CONTROLLER_GEN := $(GOBIN)/controller-gen
GOLANGCI_LINT := $(GOBIN)/golangci-lint

.PHONY: all
all: generate manifests fmt vet lint

##@ Development

.PHONY: fmt
fmt: ## Run go fmt
	go fmt ./...

.PHONY: vet
vet: ## Run go vet
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golangci-lint
	$(GOLANGCI_LINT) run --timeout=5m

.PHONY: lint-fix
lint-fix: golangci-lint ## Run golangci-lint with auto-fix
	$(GOLANGCI_LINT) run --fix --timeout=5m

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -v -race -coverprofile=coverage.out

.PHONY: test-short
test-short: ## Run tests without race detector
	go test ./... -v -short

.PHONY: coverage
coverage: test ## Generate coverage report
	go tool cover -html=coverage.out -o coverage.html

##@ Code Generation

.PHONY: generate
generate: controller-gen ## Generate DeepCopy methods
	$(CONTROLLER_GEN) object paths="./api/..."

.PHONY: manifests
manifests: controller-gen ## Generate CRD manifests
	$(CONTROLLER_GEN) crd paths="./api/..." output:crd:artifacts:config=crds

.PHONY: verify-generate
verify-generate: generate manifests ## Verify generated code is up to date
	@if [ -n "$$(git status --porcelain api/ crds/)" ]; then \
		echo "Generated files are out of date. Run 'make generate manifests' and commit the changes."; \
		git diff api/ crds/; \
		exit 1; \
	fi

##@ Build

.PHONY: build
build: generate fmt vet ## Build the module
	go build ./...

.PHONY: tidy
tidy: ## Run go mod tidy
	go mod tidy

.PHONY: verify-tidy
verify-tidy: tidy ## Verify go.mod is tidy
	@if [ -n "$$(git status --porcelain go.mod go.sum)" ]; then \
		echo "go.mod or go.sum is not tidy. Run 'go mod tidy' and commit the changes."; \
		git diff go.mod go.sum; \
		exit 1; \
	fi

##@ Tools

.PHONY: controller-gen
controller-gen: ## Install controller-gen
	@echo "Installing controller-gen $(CONTROLLER_GEN_VERSION)..."
	@go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_GEN_VERSION)

.PHONY: golangci-lint
golangci-lint: ## Install golangci-lint
	@echo "Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."
	@go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: tools
tools: controller-gen golangci-lint ## Install all tools

##@ CI

.PHONY: ci
ci: tidy generate manifests fmt vet lint test ## Run all CI checks

.PHONY: verify
verify: verify-tidy verify-generate ## Verify all generated content is up to date

##@ Help

.PHONY: help
help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
