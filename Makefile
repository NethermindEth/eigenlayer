include .env

.DEFAULT_GOAL 	:= help
.PHONY: generate build build-linux build-docker install run gomod_tidy format staticcheck test codecov-test pre-commit install-deps install-gofumpt install-mockgen install-staticcheck help

generate:
	@go generate ./...

build: generate ## Compile the binary
	@mkdir -p bin
	@go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

build-linux-amd64: generate ## Compile the binary for amd64
	@env GOOS=linux GOARCH=amd64 go build -o bin/$(APP_NAME)-linux-amd64 cmd/$(APP_NAME)/main.go

build-linux-arm64: generate ## Compile the binary for arm64
	@env GOOS=linux GOARCH=arm64 go build -o bin/$(APP_NAME)-linux-arm64 cmd/$(APP_NAME)/main.go

build-linux: generate ## Compile the binary for linux
	@env GOOS=linux go build -o bin/$(APP_NAME) cmd/$(APP_NAME)/main.go

build-docker: build-linux ## Build docker image
	@docker build -t $(APP_NAME) .

install: build ## compile the binary and copy it to PATH
	@sudo cp bin/$(APP_NAME) /usr/local/bin

run: build ## Compile and run the binary
	@export DOCKER_API_VERSION=1.41
	@./bin/$(APP_NAME)

gomod_tidy: ## Run go mod tidy to clean up & install dependencies
	@go mod tidy

format: ## Run gofumpt against code to format it
	@gofumpt -l -w .

staticcheck: ## Run staticcheck against code
	@staticcheck ./...

test: unit-test e2e-test ## Run tests

unit-test: generate ## Run unit tests
	@go test -v -count=1 ./cli/... ./internal/... ./pkg/...

e2e-test: generate ## Run e2e tests
	@go test -timeout 15m  -v -count=1 ./e2e/...

codecov-test: generate ## Run tests with coverage
	@mkdir -p coverage
	@go test -coverprofile=coverage/coverage.out -covermode=count ./...
	@go tool cover -html=coverage/coverage.out -o coverage/coverage.html

pre-commit: generate format staticcheck build test

install-deps: install-gofumpt install-mockgen install-staticcheck ## Install dependencies

install-gofumpt: ## Install gofumpt for formatting
	go install mvdan.cc/gofumpt@$(GOFUMPT_VERSION)

install-mockgen: ## Install mockgen for generating mocks
	go install github.com/golang/mock/mockgen@$(MOCKGEN_VERSION)
	go get github.com/golang/mock/mockgen/model

install-staticcheck:
	go install honnef.co/go/tools/cmd/staticcheck@$(STATICCHECK_VERSION)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'