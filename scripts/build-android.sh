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
        go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.3
    fi

    # Warn if multiple ebitenmobile binaries are on PATH (stale copies cause version mismatches)
    EBITENMOBILE_PATHS=$(IFS=:; for dir in $PATH; do [ -x "$dir/ebitenmobile" ] && echo "$dir/ebitenmobile"; done || true)
    if [ -n "$EBITENMOBILE_PATHS" ]; then
        EBITENMOBILE_COUNT=$(echo "$EBITENMOBILE_PATHS" | wc -l | tr -d ' ')
    else
        EBITENMOBILE_COUNT=0
    fi
    if [ "$EBITENMOBILE_COUNT" -gt 1 ]; then
        echo_warn "Multiple ebitenmobile binaries found on PATH:"
        while IFS= read -r p; do [ -n "$p" ] && echo_warn "  $p"; done <<< "$EBITENMOBILE_PATHS"
        echo_warn "A stale copy may shadow the correct version. Ensure \$GOBIN or \$GOPATH/bin is first in PATH."
    fi
    
    if [ -z "$ANDROID_HOME" ]; then
        echo_error "ANDROID_HOME is not set"
        exit 1
    fi
    
    if [ -z "$ANDROID_NDK_HOME" ]; then
        echo_error "ANDROID_NDK_HOME is not set"
        exit 1
    fi

    # Verify Java compiler is available — ebitenmobile bind silently skips Java
    # compilation if javac is missing, producing an AAR with only native .so files
    # and no classes.jar. This causes ClassNotFoundException at runtime.
    if ! command -v javac &> /dev/null; then
        echo_error "javac is not installed or not on PATH"
        echo_error "ebitenmobile bind requires a JDK to compile the Java bridge sources"
        echo_error "Install a JDK and ensure JAVA_HOME is set, then retry"
        exit 1
    fi
    if [ -z "$JAVA_HOME" ]; then
        echo_warn "JAVA_HOME is not set — ebitenmobile bind may fail to locate the JDK"
        echo_warn "Set JAVA_HOME to the root of your JDK installation and retry"
    fi
    echo_info "Java compiler: $(javac -version 2>&1)"
    
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
    
    # Check if ebitenmobile is installed
    if ! command -v ebitenmobile &> /dev/null; then
        echo_error "ebitenmobile is not installed"
        echo_info "Installing ebitenmobile..."
        go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.3
        
        if ! command -v ebitenmobile &> /dev/null; then
            echo_error "Failed to install ebitenmobile. Please check your Go installation and PATH."
            exit 1
        fi
    fi
    
    # Build the AAR library.  ebitenmobile bind produces a library AAR containing
    # the Ebiten mobile bridge classes (EbitenView, EbitenSurfaceView, Mobile) and
    # native libgojni.so libraries for each target architecture.
    # The Activity class lives in the Android application project, not in this AAR.
    echo_info "Running ebitenmobile bind..."
    ebitenmobile bind \
        -v \
        -target android \
        -javapkg $PACKAGE_NAME \
        -o "$BUILD_DIR/libs/mobile.aar" \
        -androidapi 21 \
        ./cmd/mobile
    
    if [ ! -f "$BUILD_DIR/libs/mobile.aar" ]; then
        echo_error "Failed to generate mobile.aar"
        exit 1
    fi
    
    echo_info "AAR built successfully: $BUILD_DIR/libs/mobile.aar"

    # Validate the AAR contents.
    # ebitenmobile bind produces a *library* AAR (not an application), so it never contains
    # an Activity class.  The expected artifacts are:
    #   - classes.jar  containing the Ebiten mobile bridge classes (EbitenView, etc.)
    #   - libgojni.so  native library for each target architecture
    if command -v unzip &> /dev/null; then
        echo_info "AAR contents:"
        unzip -l "$BUILD_DIR/libs/mobile.aar" | grep -E "\.class|\.so" | head -20

        AAR_ERRORS=0

        # 1. Check for native libraries
        if unzip -l "$BUILD_DIR/libs/mobile.aar" | grep -q "libgojni.so"; then
            echo_info "✓ Native libraries (libgojni.so) found in AAR"
        else
            echo_error "✗ Native libraries (libgojni.so) missing from AAR"
            AAR_ERRORS=$((AAR_ERRORS + 1))
        fi

        # 2. Check for classes.jar and expected Ebiten bridge classes inside it
        if command -v jar &> /dev/null; then
            AAR_TMP=$(mktemp -d)
            unzip -q "$BUILD_DIR/libs/mobile.aar" classes.jar -d "$AAR_TMP" 2>/dev/null || true
            if [ -f "$AAR_TMP/classes.jar" ]; then
                JAR_CLASSES=$(jar tf "$AAR_TMP/classes.jar" 2>/tmp/jar_tf_stderr || true)
                if [ -s /tmp/jar_tf_stderr ]; then
                    echo_warn "jar tf reported errors: $(cat /tmp/jar_tf_stderr)"
                fi
                rm -f /tmp/jar_tf_stderr
                echo_info "classes.jar present in AAR"
                # Verify at least one of the known Ebiten mobile bridge classes is present
                BRIDGE_FOUND=0
                for BRIDGE_CLASS in "EbitenView.class" "EbitenSurfaceView.class" "Mobile.class"; do
                    if echo "$JAR_CLASSES" | grep -qF "$BRIDGE_CLASS"; then
                        echo_info "✓ Found Ebiten bridge class: $BRIDGE_CLASS"
                        BRIDGE_FOUND=1
                        break
                    fi
                done
                if [ "$BRIDGE_FOUND" -eq 0 ]; then
                    echo_error "✗ No Ebiten bridge classes found in classes.jar - Java sources may not have compiled"
                    echo_error "Common causes: missing javac, JAVA_HOME not set, or gomobile not initialised"
                    echo_error "Try re-initialising gomobile:"
                    echo_error "  go install golang.org/x/mobile/cmd/gomobile@latest"
                    echo_error "  gomobile init"
                    echo_error "Then reinstall ebitenmobile matching the project version:"
                    echo_error "  go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.3"
                    AAR_ERRORS=$((AAR_ERRORS + 1))
                fi
            else
                echo_error "✗ classes.jar missing from AAR - Java sources did not compile"
                echo_error "Common causes: missing javac, JAVA_HOME not set, or gomobile not initialised"
                echo_error "Try re-initialising gomobile:"
                echo_error "  go install golang.org/x/mobile/cmd/gomobile@latest"
                echo_error "  gomobile init"
                AAR_ERRORS=$((AAR_ERRORS + 1))
            fi
            rm -rf "$AAR_TMP"
        else
            # Fallback when jar tool is unavailable: check for classes.jar entry in the archive index
            if unzip -l "$BUILD_DIR/libs/mobile.aar" | grep -q "classes.jar"; then
                echo_info "✓ classes.jar present in AAR"
            else
                echo_error "✗ classes.jar missing from AAR - Java sources did not compile"
                AAR_ERRORS=$((AAR_ERRORS + 1))
            fi
        fi

        if [ "$AAR_ERRORS" -gt 0 ]; then
            echo_error "AAR validation failed with $AAR_ERRORS error(s)"
            exit 1
        fi
        echo_info "✓ AAR validation passed"
    fi
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
        bash gradlew assembleRelease
        APK_FILE="$BUILD_DIR/build/outputs/apk/release/*.apk"
    else
        bash gradlew assembleDebug
        APK_FILE="$BUILD_DIR/build/outputs/apk/debug/*.apk"
    fi
    
    # Copy to output directory
    mkdir -p "$OUTPUT_DIR"
    cp $APK_FILE "$OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}-${build_type}.apk"
    
    echo_info "APK built: $OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}-${build_type}.apk"
}

# Build AAB (Android App Bundle)
build_aab() {
    echo_info "Building AAB (Android App Bundle)..."
    
    cd "$BUILD_DIR"
    
    bash gradlew bundleRelease
    
    AAB_FILE="$BUILD_DIR/build/outputs/bundle/release/*.aab"
    
    # Copy to output directory
    mkdir -p "$OUTPUT_DIR"
    cp $AAB_FILE "$OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}.aab"
    
    echo_info "AAB built: $OUTPUT_DIR/${APP_NAME}-${VERSION_NAME}.aab"
}

# Install on connected device
install_debug() {
    echo_info "Installing debug APK on connected device..."
    
    cd "$BUILD_DIR"
    bash gradlew installDebug
    
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
