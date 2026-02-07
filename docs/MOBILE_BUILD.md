# Mobile Build Guide

**Version:** 1.0.0  
**Last Updated:** February 2026

Guide for building and deploying Venture on iOS and Android platforms.

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Android Build](#android-build)
4. [iOS Build](#ios-build)
5. [Touch Input](#touch-input)
6. [Troubleshooting](#troubleshooting)

---

## Overview

Venture supports native mobile builds for iOS and Android using Ebiten's ebitenmobile tool.

**Features:**
- Native performance (no emulation)
- Touch input with virtual controls
- Auto-scaling for different screen sizes
- Cross-platform codebase (shared with desktop)

---

## Prerequisites

### General

- Go 1.24.5+
- ebitenmobile: `go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest`

### Android

- Android SDK (API 21+)
- Android NDK
- Java Development Kit (JDK) 8+

**Install via Android Studio** or:

```bash
# Linux/Mac
export ANDROID_HOME=$HOME/Android/Sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/<version>
export PATH=$PATH:$ANDROID_HOME/platform-tools
```

### iOS

- macOS with Xcode 12+
- iOS SDK
- Apple Developer account (for distribution)

---

## Android Build

### APK (Debug)

```bash
make android-apk
# Output: venture.apk
```

### AAB (Release)

```bash
make android-aab
# Output: venture.aab (for Google Play)
```

### Manual Build

```bash
# AAR library
ebitenmobile bind -target android -o venture.aar ./cmd/mobile

# APK
cd build/android
./gradlew assembleDebug

# Install to device
adb install -r app/build/outputs/apk/debug/app-debug.apk
```

### Signing (Release)

```bash
# Generate keystore
keytool -genkey -v -keystore release.keystore -alias venture -keyalg RSA -keysize 2048 -validity 10000

# Sign AAB
jarsigner -verbose -sigalg SHA256withRSA -digestalg SHA-256 -keystore release.keystore venture.aab venture
```

---

## iOS Build

### XCFramework (Library)

```bash
make ios-xcframework
# Output: Venture.xcframework
```

### IPA (Distribution)

```bash
make ios-ipa
# Output: venture.ipa (for App Store)
```

### Manual Build

```bash
# Bind framework
ebitenmobile bind -target ios -o Venture.xcframework ./cmd/mobile

# Build with Xcode
open build/ios/Venture.xcodeproj
# Select target, configure signing, build
```

### App Store Submission

1. Archive build in Xcode
2. Upload to App Store Connect
3. Submit for review

**Requirements:** App icons, screenshots, privacy policy, bundle ID

---

## Touch Input

### Virtual Controls

**D-Pad (bottom-left):** Movement control  
**Action Buttons (bottom-right):** A (attack), B (use item)  
**Menu Button (top-right):** Pause/settings

### Platform Detection

```go
// pkg/mobile/platform.go
IsTouchCapable() bool  // iOS, Android, WASM
IsMobilePlatform() bool  // iOS, Android only
```

### Input Integration

Touch input automatically enabled on mobile platforms. See [TOUCH_INPUT_WASM.md](TOUCH_INPUT_WASM.md) for architecture details.

---

## Troubleshooting

### Android

**Build Errors:**
- Verify ANDROID_HOME and ANDROID_NDK_HOME
- Check API level compatibility (21+)
- Update NDK/SDK versions

**Install Failures:**
- Enable USB debugging on device
- Check ADB connectivity: `adb devices`
- Uninstall old version first

### iOS

**Signing Errors:**
- Configure provisioning profile in Xcode
- Verify Apple Developer account status
- Check bundle ID matches certificate

**Build Errors:**
- Update Xcode to latest version
- Clean build folder (Product → Clean Build Folder)
- Verify Go/ebitenmobile versions

### Performance

**Low FPS:**
- Reduce particle effects
- Lower screen resolution
- Optimize sprite generation

**Touch Responsiveness:**
- Verify virtual controls are properly positioned
- Check touch event handling in logs
- Test on physical device (not just emulator)

---

## Platform-Specific Notes

### Android

**Minimum SDK:** API 21 (Android 5.0 Lollipop)  
**Target SDK:** API 33 (Android 13)  
**Permissions:** None required (offline game)

### iOS

**Minimum Version:** iOS 12.0  
**Target Version:** iOS 16.0  
**Capabilities:** No special capabilities required

---

## Build Scripts

See `scripts/` directory:
- `build-android.sh` - Android APK/AAB build
- `build-ios.sh` - iOS XCFramework/IPA build

See `Makefile.mobile` for build targets.

---

## Additional Resources

- [Ebiten Mobile Docs](https://ebitengine.org/en/documents/mobile.html)
- [Touch Input Guide](TOUCH_INPUT_WASM.md)
- [Development Guide](DEVELOPMENT.md)

---

**Version:** 1.0.0  
**Last Updated:** February 2026  
**Maintained By:** Venture Development Team
