# Run `make` or `make help` to list targets.
.DEFAULT_GOAL := help
.PHONY: help build test cover vet fmt lint lint-go lint-web open clean node_modules check e2e mutation golden

OUT  := dist/recipe-math.html
WASM := build/recipemath.wasm

help: ## List available targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*## "} {printf "  \033[1mmake %-7s\033[0m %s\n", $$1, $$2}'

$(WASM): $(shell find frac kitchen cmd/wasm -name '*.go' -not -name '*_test.go') go.mod
	GOOS=js GOARCH=wasm go build -trimpath -ldflags="-s -w" -o $@ ./cmd/wasm

$(OUT): $(WASM) web/index.html web/style.css web/app.js cmd/bundle/main.go
	go run ./cmd/bundle -wasm $(WASM) -out $@

build: $(OUT) ## Build the single-file app → dist/recipe-math.html

test: ## Run the Go tests
	go test ./...

cover: ## Go coverage → reports/coverage.html (fails under 90%)
	./scripts/coverage.sh

golden: ## Rewrite the golden files after an intended content change (review the diff!)
	go test ./kitchen -run TestGoldenShifts -update

mutation: ## Mutation testing with Gremlins → reports/mutation.md
	@command -v $$(go env GOPATH)/bin/gremlins >/dev/null || go install github.com/go-gremlins/gremlins/cmd/gremlins@v0.6.0
	./scripts/mutation.sh

e2e: build node_modules ## Playwright UI tests (desktop + phone) → reports/playwright
	pnpm exec playwright install chromium
	rm -rf reports/e2e-coverage-raw
	pnpm e2e

vet: ## go vet (including the WebAssembly entry point)
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./cmd/wasm

fmt: node_modules ## Format Go (gofmt/goimports) and web code (Prettier)
	golangci-lint fmt ./...
	pnpm format

lint: lint-go lint-web ## All static analysis

lint-go: ## golangci-lint with the Google Go style rules (see .golangci.yml)
	golangci-lint run ./...
	GOOS=js GOARCH=wasm golangci-lint run ./cmd/wasm

lint-web: node_modules ## ESLint + Prettier check for the browser code and tests
	pnpm lint
	pnpm format:check

node_modules: package.json pnpm-lock.yaml
	pnpm install
	@touch node_modules

open: build ## Build, then open the app in your browser
	open $(OUT) 2>/dev/null || xdg-open $(OUT)

check: lint cover build e2e ## Everything CI runs (except mutation testing)

clean: ## Remove build output and reports
	rm -rf build dist reports coverage playwright-report test-results
