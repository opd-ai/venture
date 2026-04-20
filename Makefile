# Makefile for Venture

.PHONY: help all build test test-integration clean deps lint fmt build-all \
        build-linux build-windows build-macos \
        build-server build-client build-wasm build-vr \
        android ios mobile-deps \
        run-client run-server serve-wasm \
        validate-code-review release package checksums sign \
        docker docker-build docker-run docker-stop \
        feature-audit visual-regression parity-test balance-validate migration-validate ux-validate quality

# Default target
.DEFAULT_GOAL := help

# Detect platform
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

# Set default architecture
ARCH ?= amd64
ifeq ($(UNAME_M),arm64)
    ARCH = arm64
endif
ifeq ($(UNAME_M),aarch64)
    ARCH = arm64
endif

# Build output directories
BUILD_DIR := build
DIST_DIR := dist

help: ## Show this help message
	@echo "Venture Build Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

all: deps build test ## Install dependencies, build, and test

deps: ## Install Go dependencies
	@echo "Downloading dependencies..."
	sudo apt-get install libc6-dev libgl1-mesa-dev libxcursor-dev libxi-dev libxinerama-dev libxrandr-dev libxxf86vm-dev libasound2-dev pkg-config xvfb
	go mod download
	go mod verify

build: build-server build-client ## Build server and client for current platform

build-server: ## Build server for current platform
	@echo "Building server..."
	go build -ldflags="-s -w" -o $(BUILD_DIR)/venture-server ./cmd/server

build-client: ## Build client for current platform
	go build -ldflags="-s -w" -o $(BUILD_DIR)/venture-client ./cmd/client

build-all: build-linux build-windows build-macos ## Build for all desktop platforms

build-linux: ## Build for Linux (amd64 and arm64)
	@echo "Building for Linux..."
	./scripts/build-linux.sh amd64
	./scripts/build-linux.sh arm64

build-windows: ## Build for Windows (amd64)
	@echo "Building for Windows..."
	./scripts/build-windows.sh amd64

build-macos: ## Build for macOS (amd64 and arm64)
	@echo "Building for macOS..."
	./scripts/build-macos.sh amd64
	./scripts/build-macos.sh arm64

test: ## Run tests (uses xvfb-run if no display available)
	@echo "Running tests..."
	@if [ -n "$$DISPLAY" ] || [ -n "$$WAYLAND_DISPLAY" ]; then \
		go test -v ./...; \
	elif command -v xvfb-run >/dev/null 2>&1; then \
		xvfb-run -a go test -v ./...; \
	else \
		echo "Warning: No display and xvfb-run not found. Install xvfb: apt-get install xvfb"; \
		exit 1; \
	fi

test-coverage: ## Run tests with coverage report (uses xvfb-run if no display available)
	@echo "Running tests with coverage..."
	@if [ -n "$$DISPLAY" ] || [ -n "$$WAYLAND_DISPLAY" ]; then \
		go test -cover -coverprofile=coverage.out ./...; \
	elif command -v xvfb-run >/dev/null 2>&1; then \
		xvfb-run -a go test -cover -coverprofile=coverage.out ./...; \
	else \
		echo "Warning: No display and xvfb-run not found. Install xvfb: apt-get install xvfb"; \
		exit 1; \
	fi
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection (uses xvfb-run if no display available)
	@echo "Running tests with race detection..."
	@if [ -n "$$DISPLAY" ] || [ -n "$$WAYLAND_DISPLAY" ]; then \
		go test -race ./...; \
	elif command -v xvfb-run >/dev/null 2>&1; then \
		xvfb-run -a go test -race ./...; \
	else \
		echo "Warning: No display and xvfb-run not found. Install xvfb: apt-get install xvfb"; \
		exit 1; \
	fi

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

test-integration: ## Run integration tests (handles headless display)
	@echo "Running integration tests..."
	@bash scripts/test-integration.sh -v

audit: ## Run audit tests for procedural generators
	@echo "Running audit tests..."
	go test ./pkg/procgen/audit/... -v

lint: ## Run linters
	@echo "Running linters..."
	go vet ./...
	@echo "Validating network type interfaces..."
	@bash scripts/validate-network-types.sh
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	go install mvdan.cc/gofumpt@latest
	find . -name '*.go' -exec gofumpt -w -s -extra {} +

validate-code-review: ## Validate code against quality gates (CODE_REVIEW_PLAN.md)
	@echo "Validating code review quality gates..."
	@./scripts/validate-code-review.sh

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -rf coverage.out coverage.html
	rm -f cpu.prof mem.prof
	rm -f build/wasm/venture.wasm build/wasm/wasm_exec.js

run-client: build-client ## Build and run client
	./$(BUILD_DIR)/venture-client

run-server: build-server ## Build and run server
	./$(BUILD_DIR)/venture-server

# Mobile targets (includes from Makefile.mobile)
mobile-deps: ## Install mobile build dependencies
	@echo "Installing ebitenmobile..."
	go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
	@echo "Mobile dependencies installed"

android-aar: mobile-deps ## Build Android AAR library
	./scripts/build-android.sh aar

android-apk: mobile-deps ## Build debug APK
	./scripts/build-android.sh apk

android-apk-release: mobile-deps ## Build release APK (requires signing)
	./scripts/build-android.sh apk-release

android-aab: mobile-deps ## Build Android App Bundle
	./scripts/build-android.sh aab

android-install: mobile-deps ## Build and install debug APK on device
	./scripts/build-android.sh install

ios-xcframework: mobile-deps ## Build iOS XCFramework
	./scripts/build-ios.sh xcframework

ios-simulator: mobile-deps ## Build for iOS Simulator
	./scripts/build-ios.sh simulator

ios-device: mobile-deps ## Build for iOS device
	./scripts/build-ios.sh device

ios-ipa: mobile-deps ## Build and export IPA
	./scripts/build-ios.sh ipa

ios-install: mobile-deps ## Build and install on connected device
	./scripts/build-ios.sh install

clean-mobile: ## Clean mobile build artifacts
	@echo "Cleaning mobile build artifacts..."
	rm -rf build/android/libs/*.aar
	rm -rf build/android/app/build
	rm -rf build/ios/Mobile.xcframework
	rm -rf build/ios/DerivedData
	rm -rf build/ios/*.xcarchive
	rm -rf dist/android
	rm -rf dist/ios
	@echo "Mobile artifacts cleaned"

# Development helpers
dev-setup: deps mobile-deps ## Setup development environment
	@echo "Development environment setup complete"
	@echo ""
	@echo "Next steps:"
	@echo "  1. Run 'make build' to build the project"
	@echo "  2. Run 'make test' to run tests"
	@echo "  3. Run 'make run-client' to start the game client"

install-tools: ## Install development tools
	@echo "Installing development tools..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Development tools installed"

profile-cpu: ## Run CPU profiling
	@echo "Running CPU profiling..."
	go test -cpuprofile=cpu.prof -bench=. ./...
	go tool pprof cpu.prof

profile-mem: ## Run memory profiling
	@echo "Running memory profiling..."
	go test -memprofile=mem.prof -bench=. ./...
	go tool pprof mem.prof

# Documentation
docs: ## Generate documentation
	@echo "Generating documentation..."
	@echo "Godoc server starting at http://localhost:6060"
	@echo "Press Ctrl+C to stop"
	godoc -http=:6060

# WebAssembly build
build-wasm: ## Build WebAssembly version for web browsers
	@echo "Building WebAssembly with optimizations..."
	@mkdir -p build/wasm
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/wasm/venture.wasm ./cmd/client
	@echo "Copying wasm_exec.js..."
	cp $$(go env GOROOT)/lib/wasm/wasm_exec.js build/wasm/
	@echo "WebAssembly build complete: build/wasm/venture.wasm"
	@echo "Optimizations enabled: viewport culling, batch rendering, sprite caching (300 sequences)"
	@echo "Run 'make serve-wasm' to test locally"

build-vr: ## Build client with experimental VR support (-tags vr) — requires OpenXR SDK
	@echo "Building with experimental VR support (OpenXR)..."
	@echo "Requirements: OpenXR loader installed (libopenxr-loader1 libopenxr-dev on Ubuntu)"
	@mkdir -p build/vr
	go build -tags vr -ldflags="-s -w" -o build/vr/venture-vr ./cmd/client
	@echo "VR build complete: build/vr/venture-vr"
	@echo "NOTE: VR SDK integration is experimental. Use --force-vr to test without hardware."

serve-wasm: build-wasm ## Build and serve WebAssembly version locally
	@echo "Starting local server at http://localhost:8080"
	@echo "Press Ctrl+C to stop"
	@cd build/wasm && python3 -m http.server 8080 || \
		(echo "Python3 not found, trying Go..." && \
		go run -tags http.Server -ldflags="-s -w" \
		-modfile <(echo "module main"; echo "go 1.24") \
		-exec "cd build/wasm &&" . :8080)

# Git helpers
git-clean: ## Remove all untracked files (use with caution!)
	@echo "This will remove all untracked files. Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]
	git clean -fdx

# Release automation (Phase 66.1)
release: clean ## Build all platforms, package, generate checksums (one-command release)
	@echo "========================================="
	@echo "  Venture Release Build"
	@echo "========================================="
	@./scripts/build-all-platforms.sh $(VERSION)
	@./scripts/package-release.sh $(VERSION)
	@./scripts/generate-checksums.sh dist
	@echo ""
	@echo "✓ Release build complete!"
	@echo "  Packages: dist/"
	@echo "  Checksums: dist/SHA256SUMS.txt"
	@echo ""
	@echo "Optional: Sign binaries with GPG"
	@echo "  make sign VERSION=$(VERSION) GPG_KEY=<your-key-id>"

package: ## Package existing build artifacts
	@./scripts/package-release.sh $(VERSION)

checksums: ## Generate SHA256 checksums for dist/ directory
	@./scripts/generate-checksums.sh dist

sign: ## Sign release binaries with GPG
	@if [ -z "$(GPG_KEY)" ]; then \
		echo "Error: GPG_KEY not set. Usage: make sign GPG_KEY=<your-key-id>"; \
		exit 1; \
	fi
	@./scripts/sign-binaries.sh dist $(GPG_KEY)

# Docker targets (Phase 66.1)
docker-build: ## Build Docker image for server
	@echo "Building Docker image..."
	docker build -t venture-server:latest -t venture-server:$(VERSION) .

docker-run: docker-build ## Build and run Docker container
	@echo "Starting Venture server in Docker..."
	docker-compose up -d

docker-stop: ## Stop Docker containers
	@echo "Stopping Venture server..."
	docker-compose down

docker: docker-run ## Alias for docker-run

# Development Tools (Phase 5: PLAN.md)
feature-audit: ## Run feature completeness audit (Phase 65.1)
	@echo "Running feature completeness audit..."
	go run ./examples/featureaudit

visual-regression: ## Run visual regression tests (Phase 63)
	@echo "Running visual regression tests..."
	xvfb-run -s "-screen 0 1920x1080x24" go run ./examples/visualregressiontest

parity-test: ## Run cross-platform parity tests (Phase 63.3)
	@echo "Running parity tests..."
	go run ./examples/paritytest

balance-validate: build-server ## Run combat/economic balance validation (Phase 6.1)
	@echo "Running balance validation..."
	./$(BUILD_DIR)/venture-server --balance-validate

migration-validate: build-server ## Run save file migration validation (Phase 6.2)
	@echo "Running migration validation..."
	./$(BUILD_DIR)/venture-server --migration-validate

ux-validate: build-server ## Run user experience journey validation (Phase 6.4)
	@echo "Running UX journey validation..."
	./$(BUILD_DIR)/venture-server --ux-validate

quality: feature-audit visual-regression parity-test balance-validate migration-validate ux-validate ## Run all quality validation tools
	@echo ""
	@echo "✓ All quality checks passed!"

.PHONY: help
