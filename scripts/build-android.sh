#!/bin/bash
set -e

# Android build script for Venture
# Requires: Go 1.24+, Android SDK, NDK, ebitenmobile

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/android"
OUTPUT_DIR="$PROJECT_ROOT/dist/android"

# Configuration
PACKAGE_NAME="com.venture.game"
APP_NAME="Venture"
VERSION_NAME="1.0.0"
VERSION_CODE="1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    echo_info "Checking prerequisites..."
    
    if ! command -v go &> /dev/null; then
        echo_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v ebitenmobile &> /dev/null; then
        echo_warn "ebitenmobile not found, installing..."
        go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
    fi
    
    if [ -z "$ANDROID_HOME" ]; then
        echo_error "ANDROID_HOME is not set"
        exit 1
    fi
    
    if [ -z "$ANDROID_NDK_HOME" ]; then
        echo_error "ANDROID_NDK_HOME is not set"
        exit 1
    fi
    
    echo_info "Prerequisites OK"
}

# Generate Android resources (icons, etc.)
generate_resources() {
    echo_info "Generating Android resources..."
    
    # Generate launcher icons
    "$SCRIPT_DIR/generate-android-icons.sh"
    
    echo_info "Resources generated successfully"
}

# Build AAR library
build_aar() {
    echo_info "Building Android AAR library..."
    
    cd "$PROJECT_ROOT"
    
    # Ensure output directory exists
    mkdir -p "$BUILD_DIR/libs"
    
    # Build the AAR
    ebitenmobile bind \
        -target android \
        -javapkg $PACKAGE_NAME \
        -o "$BUILD_DIR/libs/mobile.aar" \
        ./cmd/mobile
    
    echo_info "AAR built successfully"
}

# Build APK
build_apk() {
    local build_type=${1:-debug}
    
    echo_info "Building APK ($build_type)..."
    
    cd "$BUILD_DIR"
    
    # Ensure gradle wrapper exists
    if [ ! -f "gradlew" ]; then
        echo_info "Initializing Gradle wrapper..."
        gradle wrapper
    fi
    
    # Build APK
    if [ "$build_type" == "release" ]; then
        ./gradlew assembleRelease
        APK_FILE="$BUILD_DIR/app/build/outputs/apk/release/app-release.apk"
    else
        ./gradlew assembleDebug
        APK_FILE="$BUILD_DIR/app/build/outputs/apk/debug/app-debug.apk"
    fi
    
    # Copy to output directory
    mkdir -p "$OUTPUT_DIR"
    cp "$APK_FILE" "$OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}-${build_type}.apk"
    
    echo_info "APK built: $OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}-${build_type}.apk"
}

# Build AAB (Android App Bundle)
build_aab() {
    echo_info "Building AAB (Android App Bundle)..."
    
    cd "$BUILD_DIR"
    
    ./gradlew bundleRelease
    
    AAB_FILE="$BUILD_DIR/app/build/outputs/bundle/release/app-release.aab"
    
    # Copy to output directory
    mkdir -p "$OUTPUT_DIR"
    cp "$AAB_FILE" "$OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}.aab"
    
    echo_info "AAB built: $OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}.aab"
}

# Install on connected device
install_debug() {
    echo_info "Installing debug APK on connected device..."
    
    cd "$BUILD_DIR"
    ./gradlew installDebug
    
    echo_info "App installed successfully"
}

# Main execution
main() {
    local command=${1:-all}
    
    check_prerequisites
    generate_resources
    
    case $command in
        aar)
            build_aar
            ;;
        apk)
            build_aar
            build_apk debug
            ;;
        apk-release)
            build_aar
            build_apk release
            ;;
        aab)
            build_aar
            build_aab
            ;;
        install)
            build_aar
            build_apk debug
            install_debug
            ;;
        all)
            build_aar
            build_apk debug
            ;;
        *)
            echo "Usage: $0 {aar|apk|apk-release|aab|install|all}"
            echo ""
            echo "Commands:"
            echo "  aar          - Build AAR library only"
            echo "  apk          - Build debug APK"
            echo "  apk-release  - Build release APK (requires signing config)"
            echo "  aab          - Build Android App Bundle for Play Store"
            echo "  install      - Build and install debug APK on connected device"
            echo "  all          - Build AAR and debug APK (default)"
            exit 1
            ;;
    esac
    
    echo_info "Build complete!"
}

main "$@"
