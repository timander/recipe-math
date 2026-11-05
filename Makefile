# Run `make` or `make help` to list targets.
.DEFAULT_GOAL := help
.PHONY: help build test cover vet fmt open serve clean

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

cover: ## Tests with a coverage summary
	go test -cover ./...

vet: ## go vet (including the WebAssembly entry point)
	go vet ./...
	GOOS=js GOARCH=wasm go vet ./cmd/wasm

fmt: ## Format all Go code
	gofmt -w .

open: build ## Build, then open the app in your browser
	open $(OUT) 2>/dev/null || xdg-open $(OUT)

clean: ## Remove build output
	rm -rf build dist
