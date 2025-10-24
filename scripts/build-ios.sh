#!/bin/bash
set -e

# iOS build script for Venture
# Requires: Go 1.24+, Xcode, ebitenmobile

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/ios"
OUTPUT_DIR="$PROJECT_ROOT/dist/ios"

# Configuration
BUNDLE_ID="com.venture.game"
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
    
    if [ "$(uname)" != "Darwin" ]; then
        echo_error "iOS builds require macOS"
        exit 1
    fi
    
    if ! command -v go &> /dev/null; then
        echo_error "Go is not installed"
        exit 1
    fi
    
    if ! command -v xcodebuild &> /dev/null; then
        echo_error "Xcode is not installed"
        exit 1
    fi
    
    if ! command -v ebitenmobile &> /dev/null; then
        echo_warn "ebitenmobile not found, installing..."
        go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
    fi
    
    echo_info "Prerequisites OK"
}

# Build XCFramework
build_xcframework() {
    echo_info "Building iOS XCFramework..."
    
    cd "$PROJECT_ROOT"
    
    # Build the XCFramework
    ebitenmobile bind \
        -target ios \
        -o "$BUILD_DIR/Mobile.xcframework" \
        ./cmd/mobile
    
    echo_info "XCFramework built successfully"
}

# Build for simulator
build_simulator() {
    echo_info "Building for iOS Simulator..."
    
    cd "$BUILD_DIR"
    
    xcodebuild \
        -scheme Venture \
        -configuration Debug \
        -sdk iphonesimulator \
        -derivedDataPath "$BUILD_DIR/DerivedData" \
        CODE_SIGN_IDENTITY="" \
        CODE_SIGNING_REQUIRED=NO
    
    echo_info "Simulator build complete"
}

# Build for device (requires signing)
build_device() {
    local build_type=${1:-Debug}
    
    echo_info "Building for iOS device ($build_type)..."
    
    cd "$BUILD_DIR"
    
    # Check for signing identity
    if [ -z "$IOS_SIGNING_IDENTITY" ]; then
        echo_warn "IOS_SIGNING_IDENTITY not set, using automatic signing"
        SIGNING_ARGS="CODE_SIGN_STYLE=Automatic"
    else
        SIGNING_ARGS="CODE_SIGN_IDENTITY=$IOS_SIGNING_IDENTITY"
    fi
    
    # Check for provisioning profile
    if [ -n "$IOS_PROVISIONING_PROFILE" ]; then
        SIGNING_ARGS="$SIGNING_ARGS PROVISIONING_PROFILE_SPECIFIER=$IOS_PROVISIONING_PROFILE"
    fi
    
    xcodebuild \
        -scheme Venture \
        -configuration $build_type \
        -sdk iphoneos \
        -derivedDataPath "$BUILD_DIR/DerivedData" \
        -archivePath "$BUILD_DIR/$APP_NAME.xcarchive" \
        $SIGNING_ARGS \
        archive
    
    echo_info "Device build archived"
}

# Export IPA
export_ipa() {
    local export_method=${1:-development}
    
    echo_info "Exporting IPA (method: $export_method)..."
    
    # Create export options plist
    cat > "$BUILD_DIR/ExportOptions.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>method</key>
    <string>$export_method</string>
    <key>teamID</key>
    <string>${IOS_TEAM_ID:-}</string>
    <key>uploadSymbols</key>
    <true/>
    <key>compileBitcode</key>
    <false/>
</dict>
</plist>
EOF
    
    xcodebuild \
        -exportArchive \
        -archivePath "$BUILD_DIR/$APP_NAME.xcarchive" \
        -exportOptionsPlist "$BUILD_DIR/ExportOptions.plist" \
        -exportPath "$OUTPUT_DIR"
    
    echo_info "IPA exported: $OUTPUT_DIR/$APP_NAME.ipa"
}

# Install on connected device
install_device() {
    echo_info "Installing on connected device..."
    
    if ! command -v ios-deploy &> /dev/null; then
        echo_warn "ios-deploy not found, installing via npm..."
        npm install -g ios-deploy
    fi
    
    IPA_FILE="$OUTPUT_DIR/$APP_NAME.ipa"
    
    if [ ! -f "$IPA_FILE" ]; then
        echo_error "IPA file not found. Build first."
        exit 1
    fi
    
    ios-deploy --bundle "$IPA_FILE"
    
    echo_info "App installed successfully"
}

# Main execution
main() {
    local command=${1:-all}
    
    check_prerequisites
    
    case $command in
        xcframework)
            build_xcframework
            ;;
        simulator)
            build_xcframework
            build_simulator
            ;;
        device)
            build_xcframework
            build_device Release
            ;;
        ipa)
            build_xcframework
            build_device Release
            export_ipa app-store
            ;;
        ipa-dev)
            build_xcframework
            build_device Debug
            export_ipa development
            ;;
        install)
            build_xcframework
            build_device Debug
            export_ipa development
            install_device
            ;;
        all)
            build_xcframework
            build_simulator
            ;;
        *)
            echo "Usage: $0 {xcframework|simulator|device|ipa|ipa-dev|install|all}"
            echo ""
            echo "Commands:"
            echo "  xcframework  - Build XCFramework only"
            echo "  simulator    - Build for iOS Simulator"
            echo "  device       - Build for iOS device (requires signing)"
            echo "  ipa          - Build and export IPA for App Store"
            echo "  ipa-dev      - Build and export IPA for development"
            echo "  install      - Build and install on connected device"
            echo "  all          - Build XCFramework and simulator app (default)"
            exit 1
            ;;
    esac
    
    echo_info "Build complete!"
}

main "$@"
