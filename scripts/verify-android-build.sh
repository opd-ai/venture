#!/bin/bash
# Verification script for Android builds
# Checks that all required components are present before building

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BUILD_DIR="$PROJECT_ROOT/build/android"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}✓${NC} $1"; }
fail() { echo -e "${RED}✗${NC} $1"; }
warn() { echo -e "${YELLOW}⚠${NC} $1"; }

echo "Verifying Android build environment..."
echo ""

# Check Go installation
if command -v go &> /dev/null; then
    GO_VERSION=$(go version | awk '{print $3}')
    pass "Go installed: $GO_VERSION"
else
    fail "Go not installed"
    exit 1
fi

# Check ebitenmobile
if command -v ebitenmobile &> /dev/null; then
    pass "ebitenmobile installed"
else
    fail "ebitenmobile not installed"
    echo "  Install with: go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@v2.9.3"
    exit 1
fi

# Check Android SDK
if [ -z "$ANDROID_HOME" ]; then
    fail "ANDROID_HOME not set"
    echo "  Set with: export ANDROID_HOME=/path/to/android-sdk"
    exit 1
else
    pass "ANDROID_HOME set: $ANDROID_HOME"
fi

# Check Android NDK
if [ -z "$ANDROID_NDK_HOME" ]; then
    fail "ANDROID_NDK_HOME not set"
    echo "  Set with: export ANDROID_NDK_HOME=\$ANDROID_HOME/ndk/26.1.10909125"
    exit 1
else
    pass "ANDROID_NDK_HOME set: $ANDROID_NDK_HOME"
fi

# Check source files
if [ -f "$PROJECT_ROOT/cmd/mobile/mobile.go" ]; then
    pass "Mobile source code found"
else
    fail "Mobile source code not found: cmd/mobile/mobile.go"
    exit 1
fi

# Check Android Manifest
if [ -f "$BUILD_DIR/AndroidManifest.xml" ]; then
    pass "AndroidManifest.xml found"
    
    # Verify meta-data element exists
    if grep -q '<meta-data android:name="android.app.lib_name" android:value="mobile"' "$BUILD_DIR/AndroidManifest.xml"; then
        pass "AndroidManifest.xml contains required meta-data element"
    else
        fail "AndroidManifest.xml missing required meta-data element"
        echo "  The activity must contain: <meta-data android:name=\"android.app.lib_name\" android:value=\"mobile\" />"
        exit 1
    fi
else
    fail "AndroidManifest.xml not found"
    exit 1
fi

# Check build.gradle
if [ -f "$BUILD_DIR/build.gradle" ]; then
    pass "build.gradle found"
else
    fail "build.gradle not found"
    exit 1
fi

# Check ProGuard rules
if [ -f "$BUILD_DIR/proguard-rules.pro" ]; then
    pass "proguard-rules.pro found"
else
    warn "proguard-rules.pro not found (recommended for release builds)"
fi

echo ""
echo "Checking AAR library..."

# Check if AAR exists
if [ -f "$BUILD_DIR/libs/mobile.aar" ]; then
    AAR_SIZE=$(ls -lh "$BUILD_DIR/libs/mobile.aar" | awk '{print $5}')
    pass "mobile.aar found ($AAR_SIZE)"
    
    # Check AAR contents
    if command -v unzip &> /dev/null; then
        if unzip -l "$BUILD_DIR/libs/mobile.aar" | grep -q "classes.jar"; then
            pass "AAR contains classes.jar"
        else
            fail "AAR missing classes.jar"
        fi
        
        if unzip -l "$BUILD_DIR/libs/mobile.aar" | grep -q "libgojni.so"; then
            pass "AAR contains native libraries"
        else
            fail "AAR missing native libraries"
        fi
    fi
else
    warn "mobile.aar not found (will be built automatically)"
fi

echo ""
echo "Checking connected devices..."

# Check for connected devices
if command -v adb &> /dev/null; then
    DEVICE_COUNT=$(adb devices | grep -v "List" | grep -c "device$" || true)
    if [ "$DEVICE_COUNT" -gt 0 ]; then
        pass "$DEVICE_COUNT device(s) connected"
        adb devices | grep "device$" | while read -r line; do
            echo "  - $line"
        done
    else
        warn "No devices connected (required for 'make android-install')"
    fi
else
    warn "adb not found (install Android SDK platform-tools)"
fi

echo ""
echo -e "${GREEN}Verification complete!${NC}"
echo ""
echo "Available build commands:"
echo "  make android-aar          - Build AAR library only"
echo "  make android-apk          - Build debug APK"
echo "  make android-apk-release  - Build release APK"
echo "  make android-install      - Build and install on device"
echo ""
echo "Or use the build script:"
echo "  ./scripts/build-android.sh apk"
