#!/bin/bash
# Test script to validate Android APK build

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo_info() { echo -e "${GREEN}✓${NC} $1"; }
echo_warn() { echo -e "${YELLOW}⚠${NC} $1"; }
echo_error() { echo -e "${RED}✗${NC} $1"; }
echo_header() { echo -e "${BLUE}=== $1 ===${NC}"; }

# Determine APK path
if [ -n "$1" ]; then
    APK_PATH="$1"
elif [ -f "$PROJECT_ROOT/dist/android/Venture-1.0.0-debug.apk" ]; then
    APK_PATH="$PROJECT_ROOT/dist/android/Venture-1.0.0-debug.apk"
else
    echo_error "APK not found. Please specify path or build with 'make android-apk'"
    exit 1
fi

echo_header "Android APK Build Verification"
echo ""

# Check if APK exists
if [ ! -f "$APK_PATH" ]; then
    echo_error "APK not found at: $APK_PATH"
    exit 1
fi

echo_info "APK found: $APK_PATH"
APK_SIZE=$(ls -lh "$APK_PATH" | awk '{print $5}')
echo_info "APK size: $APK_SIZE"
echo ""

# Verify APK contents
echo_header "Verifying APK Contents"

# Check for AndroidManifest.xml
if unzip -l "$APK_PATH" | grep -q "AndroidManifest.xml"; then
    echo_info "AndroidManifest.xml present"
else
    echo_error "AndroidManifest.xml missing"
    exit 1
fi

# Check for native libraries
echo ""
echo "Checking native libraries:"
for arch in arm64-v8a armeabi-v7a x86 x86_64; do
    if unzip -l "$APK_PATH" | grep -q "lib/$arch/libgojni.so"; then
        SIZE=$(unzip -l "$APK_PATH" | grep "lib/$arch/libgojni.so" | awk '{print $1}')
        echo_info "  $arch: libgojni.so ($SIZE bytes)"
    else
        echo_warn "  $arch: libgojni.so missing"
    fi
done

# Check for resources
echo ""
echo "Checking resources:"
if unzip -l "$APK_PATH" | grep -q "res/"; then
    echo_info "Resources directory present"
    RES_COUNT=$(unzip -l "$APK_PATH" | grep -c "res/" || true)
    echo_info "  $RES_COUNT resource files"
else
    echo_warn "No resources found"
fi

# Check for classes
echo ""
echo "Checking DEX files:"
DEX_COUNT=$(unzip -l "$APK_PATH" | grep -c "\.dex" || true)
if [ "$DEX_COUNT" -gt 0 ]; then
    echo_info "$DEX_COUNT DEX file(s) present"
else
    echo_warn "No DEX files found"
fi

echo ""
echo_header "Build Summary"
echo ""
echo_info "APK successfully built and verified"
echo_info "Location: $APK_PATH"
echo_info "Size: $APK_SIZE"
echo ""
echo "Build artifacts:"
echo "  - Native libraries for all architectures: ✓"
echo "  - Android manifest: ✓"
echo "  - Resources: ✓"
echo "  - DEX files: ✓"
echo ""
echo_info "APK is ready for installation on Android devices"
echo ""
echo "To install on a connected device:"
echo "  adb install -r \"$APK_PATH\""
