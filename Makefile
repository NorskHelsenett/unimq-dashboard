GO := go
GOFMT := gofmt
MAIN_PATH := ./cmd/unimq
GENERATOR_PATH := ./cmd/generator
GO_VERSION := $(shell go version | cut -d' ' -f3)

# Docker variables
DOCKER_IMAGE := ror-api
DOCKER_TAG := latest
DOCKER_REGISTRY := 

# Build info
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

# LDFLAGS for build
LDFLAGS := -X main.Version=$(VERSION) \
		   -X main.GitCommit=$(GIT_COMMIT) \
		   -X main.BuildTime=$(BUILD_TIME) \
		   -X main.GitBranch=$(GIT_BRANCH)

# Colors for output
RED := \033[0;31m
GREEN := \033[0;32m
YELLOW := \033[0;33m
BLUE := \033[0;34m
RESET := \033[0m

GO_VERSION := $(shell awk '/^go /{print $$2}' go.mod)
GOLANGCI_LINT = golangci-lint
GOLANGCI_LINT_VERSION ?= latest

##@ Prerequisites
# Prerequisite check targets
.PHONY: check-go check-docker

check-go: ## Check if Go is installed
	@echo "${YELLOW}Checking if Go is installed...${RESET}"
	@which ${GO} > /dev/null || (echo "${RED}Error: Go is not installed or not in PATH${RESET}" && exit 1)
	@echo "${GREEN}Go is installed: $$(${GO} version)${RESET}"

check-docker: ## Check if Docker is installed
	@echo "${YELLOW}Checking if Docker is installed...${RESET}"
	@which docker > /dev/null || (echo "${RED}Error: Docker is not installed or not in PATH${RESET}" && exit 1)
	@echo "${GREEN}Docker is installed: $$(docker --version)${RESET}"

# Check if swag is installed
check-swag: ## Check if swag is installed, install if not
	@if ! command -v swag &> /dev/null; then \
		echo "${YELLOW}swag not found, installing...${RESET}"; \
		${GO} install github.com/swaggo/swag/cmd/swag@latest; \
		echo "${GREEN}swag installed!${RESET}"; \
	else \
		echo "${GREEN}swag is installed${RESET}"; \
	fi

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN) ## Download golangci-lint locally if necessary.
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/v2/cmd/golangci-lint,$(GOLANGCI_LINT_VERSION))

##@ Build and test stuff

# Run go vet
vet: check-go ## Run go vet to catch potential issues
	@echo "${YELLOW}Running go vet...${RESET}"
	${GO} vet ./...
	@echo "${GREEN}go vet completed!${RESET}"

# Run linting
lint: golangci-lint ## Run linting with golangci-lint
	@echo "${YELLOW}Running linter...${RESET}"
	$(GOLANGCI_LINT) run --timeout 5m ./... --config .golangci.yaml
	@echo "${GREEN}Linting completed!${RESET}"

# Generate Swagger documentation
docs-run: check-go check-swag ## Generate Swagger documentation
	@echo "${YELLOW}Generating Swagger docs...${RESET}"
	swag init -g ${MAIN_PATH}/main.go -o ./internal/docs --parseInternal
	@echo "${GREEN}Swagger documentation generated!${RESET}"

# Format swagger documentation
docs-fmt: check-go check-swag ## Format generated Swagger documentation
	@echo "${YELLOW}Formatting Swagger docs...${RESET}"
	swag fmt -d ./
	@echo "${GREEN}Swagger documentation formatted!${RESET}"

# Start 
run: check-go ## Start the application
	@echo "${YELLOW}starting ${MAIN_PATH}/main.go...${RESET}"
	go run ${MAIN_PATH}/main.go

generate: check-go ## Run the generator
	@echo "${YELLOW}starting ${GENERATOR_PATH}/main.go...${RESET}"
	go run ${GENERATOR_PATH}/main.go
