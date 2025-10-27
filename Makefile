# Makefile for Venture

.PHONY: help all build test clean deps lint fmt build-all \
        build-linux build-windows build-macos \
        build-server build-client build-wasm \
        android ios mobile-deps \
        run-client run-server serve-wasm

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

test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

test-race: ## Run tests with race detection
	@echo "Running tests with race detection..."
	go test -race ./...

bench: ## Run benchmarks
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

lint: ## Run linters
	@echo "Running linters..."
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping"; \
	fi

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

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
	@echo "Building WebAssembly..."
	@mkdir -p build/wasm
	GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o build/wasm/venture.wasm ./cmd/client
	@echo "Copying wasm_exec.js..."
	cp $$(go env GOROOT)/lib/wasm/wasm_exec.js build/wasm/
	@echo "WebAssembly build complete: build/wasm/venture.wasm"
	@echo "Run 'make serve-wasm' to test locally"

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

.PHONY: help
