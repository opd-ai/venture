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
    
    # Ensure output directory exists
    mkdir -p "$OUTPUT_DIR"
    
    # Build the XCFramework
    ebitenmobile bind \
        -target ios \
        -o "$BUILD_DIR/Mobile.xcframework" \
        ./cmd/mobile
    
    echo_info "XCFramework built successfully"
    echo_info "Output: $BUILD_DIR/Mobile.xcframework"
}

# Package XCFramework for distribution
package_xcframework() {
    echo_info "Packaging XCFramework..."
    
    cd "$BUILD_DIR"
    
    if [ ! -d "Mobile.xcframework" ]; then
        echo_error "XCFramework not found. Build it first with: $0 xcframework"
        exit 1
    fi
    
    # Create zip archive
    zip -r "$OUTPUT_DIR/Venture.xcframework.zip" Mobile.xcframework
    
    echo_info "XCFramework packaged: $OUTPUT_DIR/Venture.xcframework.zip"
}

# Build for simulator (requires Xcode project)
build_simulator() {
    echo_warn "Building iOS simulator app requires an Xcode project."
    echo_warn "The XCFramework is available for integration into your Xcode project."
    echo_warn "See docs/MOBILE_BUILD.md for integration instructions."
    
    if [ ! -d "$BUILD_DIR/Venture.xcodeproj" ]; then
        echo_error "Xcode project not found at $BUILD_DIR/Venture.xcodeproj"
        echo_error "To create a simulator build:"
        echo_error "1. Create an Xcode project in $BUILD_DIR"
        echo_error "2. Link the Mobile.xcframework"
        echo_error "3. Build with xcodebuild"
        exit 1
    fi
    
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

# Build for device (requires Xcode project and signing)
build_device() {
    local build_type=${1:-Debug}
    
    echo_warn "Building iOS device app requires an Xcode project and code signing."
    echo_warn "The XCFramework is available for integration into your Xcode project."
    
    if [ ! -d "$BUILD_DIR/Venture.xcodeproj" ]; then
        echo_error "Xcode project not found at $BUILD_DIR/Venture.xcodeproj"
        exit 1
    fi
    
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

# Export IPA (requires device build)
export_ipa() {
    local export_method=${1:-development}
    
    if [ ! -d "$BUILD_DIR/$APP_NAME.xcarchive" ]; then
        echo_error "Archive not found. Build for device first with: $0 device"
        exit 1
    fi
    
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

# Install on connected device (requires IPA)
install_device() {
    echo_info "Installing on connected device..."
    
    if ! command -v ios-deploy &> /dev/null; then
        echo_warn "ios-deploy not found, installing via npm..."
        npm install -g ios-deploy
    fi
    
    IPA_FILE="$OUTPUT_DIR/$APP_NAME.ipa"
    
    if [ ! -f "$IPA_FILE" ]; then
        echo_error "IPA file not found. Build first with: $0 ipa"
        exit 1
    fi
    
    ios-deploy --bundle "$IPA_FILE"
    
    echo_info "App installed successfully"
}

# Main execution
main() {
    local command=${1:-xcframework}
    
    check_prerequisites
    
    case $command in
        xcframework)
            build_xcframework
            ;;
        package)
            package_xcframework
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
            package_xcframework
            ;;
        *)
            echo "Usage: $0 {xcframework|package|simulator|device|ipa|ipa-dev|install|all}"
            echo ""
            echo "Commands:"
            echo "  xcframework  - Build XCFramework only (default, CI-recommended)"
            echo "  package      - Package XCFramework as zip for distribution"
            echo "  simulator    - Build for iOS Simulator (requires Xcode project)"
            echo "  device       - Build for iOS device (requires Xcode project + signing)"
            echo "  ipa          - Build and export IPA for App Store (requires project + signing)"
            echo "  ipa-dev      - Build and export IPA for development (requires project + signing)"
            echo "  install      - Build and install on connected device (requires project + signing)"
            echo "  all          - Build and package XCFramework"
            echo ""
            echo "Note: simulator, device, ipa, and install commands require an Xcode project."
            echo "See docs/MOBILE_BUILD.md for instructions on creating an Xcode project wrapper."
            exit 1
            ;;
    esac
    
    echo_info "Build complete!"
}

main "$@"
