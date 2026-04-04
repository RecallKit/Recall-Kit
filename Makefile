# RecallKit Makefile
# Run `make help` to see all targets.

BINARY     := recallkit
GO_FILES   := $(shell find . -name '*.go' -not -path './vendor/*')
UI_DIR     := ui
DIST_DIR   := $(UI_DIR)/dist

.PHONY: help build test lint clean build-ui dev

# Default target
all: build

## help: Print this help message
help:
	@echo ""
	@echo "RecallKit — available make targets:"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/## /  /'
	@echo ""

# ── Go targets (no Node required) ─────────────────────────────────────────────

## build: Compile the binary (uses whatever is already in ui/dist/)
build:
	go build -o $(BINARY) .

## test: Run the full test suite
test:
	go test ./...

## test-race: Run tests with the race detector (requires CGO / native GCC)
test-race:
	go test -race ./...

## lint: Run go vet and staticcheck
lint:
	go vet ./...
	@which staticcheck > /dev/null 2>&1 && staticcheck ./... || echo "staticcheck not installed — skipping (go install honnef.co/go/tools/cmd/staticcheck@latest)"

## clean: Remove compiled binary and test cache
clean:
	rm -f $(BINARY) $(BINARY).exe
	go clean -testcache

# ── Svelte / UI targets (requires Node 20+ and npm) ──────────────────────────

## build-ui: Build the Svelte frontend and deposit assets into ui/dist/
build-ui:
	@echo "→ Building Svelte frontend..."
	cd $(UI_DIR) && npm ci && npm run build
	@echo "✓ Frontend assets written to $(DIST_DIR)/"
	@echo "  Commit ui/dist/ to update the embedded assets in the binary."

## dev-ui: Start the Svelte dev server (hot-reload, no Go binary needed)
dev-ui:
	cd $(UI_DIR) && npm run dev

## dev: Build UI assets, then build and run the Go binary
dev: build-ui build
	./$(BINARY) start
