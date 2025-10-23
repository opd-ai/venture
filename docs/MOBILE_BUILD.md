# Mobile Build Guide

This guide covers building and deploying Venture for iOS and Android platforms.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Start](#quick-start)
- [Android Build](#android-build)
- [iOS Build](#ios-build)
- [Touch Controls](#touch-controls)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Common Requirements

- **Go 1.24+**: Required for all builds
- **ebitenmobile**: Install with `go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest`
- **Git**: For cloning and version control

### Android Requirements

- **Android SDK**: API Level 21+ (Android 5.0+)
- **Android NDK**: Version 26.1.10909125 or later
- **JDK 11**: For Gradle builds
- **Gradle**: Will be downloaded automatically via wrapper

#### Installing Android SDK on Linux/macOS

```bash
# Download Android command-line tools
wget https://dl.google.com/android/repository/commandlinetools-linux-9477386_latest.zip
unzip commandlinetools-linux-9477386_latest.zip -d ~/android-sdk
mv ~/android-sdk/cmdline-tools ~/android-sdk/latest
mkdir -p ~/android-sdk/cmdline-tools
mv ~/android-sdk/latest ~/android-sdk/cmdline-tools/

# Set environment variables
export ANDROID_HOME=~/android-sdk
export ANDROID_NDK_HOME=~/android-sdk/ndk/26.1.10909125
export PATH=$PATH:$ANDROID_HOME/cmdline-tools/latest/bin:$ANDROID_HOME/platform-tools

# Install required SDK components
sdkmanager "platform-tools" "platforms;android-34" "build-tools;34.0.0" "ndk;26.1.10909125"
```

### iOS Requirements (macOS Only)

- **macOS 12+**: Required for iOS development
- **Xcode 14+**: Install from Mac App Store
- **Command Line Tools**: `xcode-select --install`
- **ios-deploy** (optional): `npm install -g ios-deploy` for device deployment

## Quick Start

### Install Mobile Dependencies

```bash
# Install ebitenmobile
make mobile-deps

# Or manually:
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
```

### Build for Android

```bash
# Build debug APK
make android-apk

# Install on connected device
make android-install

# Build release APK (requires signing configuration)
make android-apk-release
```

### Build for iOS

```bash
# Build for iOS Simulator
make ios-simulator

# Build for device (requires signing)
make ios-device

# Build and export IPA
make ios-ipa
```

## Android Build

### Build Types

#### 1. Debug APK (Development)

```bash
# Using Makefile
make android-apk

# Using build script directly
./scripts/build-android.sh apk

# Output: dist/android/Venture-1.0.0-debug.apk
```

#### 2. Release APK (Production)

Requires signing configuration. Set up environment variables:

```bash
export VENTURE_KEYSTORE_FILE=/path/to/keystore.jks
export VENTURE_KEYSTORE_PASSWORD=your_keystore_password
export VENTURE_KEY_ALIAS=your_key_alias
export VENTURE_KEY_PASSWORD=your_key_password

# Build signed release APK
make android-apk-release

# Output: dist/android/Venture-1.0.0-release.apk
```

#### 3. Android App Bundle (Play Store)

```bash
# Build AAB for Google Play Store
make android-aab

# Output: dist/android/Venture-1.0.0.aab
```

### Generating a Signing Key

```bash
# Generate keystore (one-time setup)
keytool -genkey -v -keystore venture.keystore \
  -alias venture -keyalg RSA -keysize 2048 -validity 10000

# Store credentials securely (e.g., in ~/.gradle/gradle.properties):
VENTURE_KEYSTORE_FILE=/path/to/venture.keystore
VENTURE_KEYSTORE_PASSWORD=your_password
VENTURE_KEY_ALIAS=venture
VENTURE_KEY_PASSWORD=your_key_password
```

### Installing on Device

```bash
# Connect Android device via USB with USB debugging enabled
# Build and install debug APK
make android-install

# Or use adb directly
adb install -r dist/android/Venture-1.0.0-debug.apk
```

### Testing on Emulator

```bash
# List available emulators
emulator -list-avds

# Start emulator
emulator -avd Pixel_6_API_34 &

# Wait for boot, then install
adb wait-for-device
adb install -r dist/android/Venture-1.0.0-debug.apk
```

## iOS Build

### Build Types

#### 1. iOS Simulator (Development)

```bash
# Build for simulator
make ios-simulator

# The build will be in build/ios/DerivedData/
# You can run it from Xcode or with:
xcrun simctl boot "iPhone 15"
xcrun simctl install booted build/ios/DerivedData/Build/Products/Debug-iphonesimulator/Venture.app
xcrun simctl launch booted com.venture.game
```

#### 2. Device Build (Testing)

Requires Apple Developer account and provisioning profile:

```bash
# Set environment variables
export IOS_SIGNING_IDENTITY="Apple Development: Your Name (TEAM123)"
export IOS_PROVISIONING_PROFILE="Venture Development Profile"
export IOS_TEAM_ID="TEAM123"

# Build for device
make ios-device
```

#### 3. IPA Export (Distribution)

```bash
# Export IPA for distribution
make ios-ipa

# Output: dist/ios/Venture.ipa
```

### Code Signing Setup

1. **Create App ID**:
   - Open [Apple Developer Portal](https://developer.apple.com/)
   - Go to Certificates, Identifiers & Profiles
   - Create new App ID: `com.venture.game`

2. **Generate Certificate**:
   - Request certificate from Certificate Authority
   - Upload CSR to Apple Developer Portal
   - Download and install certificate

3. **Create Provisioning Profile**:
   - Development profile for testing
   - Distribution profile for App Store
   - Link to your App ID and certificate

4. **Configure Xcode**:
   - Open Xcode
   - Preferences → Accounts → Add Apple ID
   - Download Manual Profiles

### Installing on Device

```bash
# Using ios-deploy
make ios-install

# Or manually via Xcode:
# 1. Open build/ios/Venture.xcodeproj in Xcode
# 2. Connect device
# 3. Select device in target dropdown
# 4. Click Run (Cmd+R)
```

## Touch Controls

### Virtual Controls Layout

The game automatically displays virtual controls on mobile:

- **Virtual D-Pad** (Bottom Left): Movement in 8 directions
- **Action Buttons** (Bottom Right): 
  - A Button: Primary action (attack, interact)
  - B Button: Secondary action (dodge, cancel)
  - X Button: Open inventory
  - Y Button: Open menu
- **Menu Button** (Top Right): Pause/Settings

### Gesture Support

- **Tap**: Select UI elements, perform actions
- **Double-Tap**: Quick action (sprint, special move)
- **Swipe**: Movement direction when not using D-pad
- **Pinch**: Zoom in/out (when implemented)
- **Long Press**: Context menu, detailed info

### Customization

Touch controls are defined in `pkg/mobile/controls.go`:

```go
// Adjust control sizes for different devices
control := &VirtualControl{
    Position: Vec2{X: 100, Y: screenHeight - 150},
    Radius:   80, // Adjust for device size
    Type:     ControlTypeDPad,
}
```

### Orientation Support

- **Landscape Mode**: Default, optimal for gameplay
- **Portrait Mode**: Supported, UI adapts automatically
- **Rotation**: Handled dynamically, no app restart needed

## CI/CD Integration

### GitHub Actions Workflows

#### Release Workflow

Workflow file: `.github/workflows/release.yml`

The release workflow automatically builds mobile artifacts alongside desktop builds:

```yaml
# Triggered on:
# - Semantic version tags (v1.2.3)
# - Nightly builds (cron schedule)

# Produces:
# - Android AAR library (venture.aar)
# - Android APK (unsigned release build)
# - iOS XCFramework (Venture.xcframework.zip)

# Jobs:
# 1. release: Builds desktop binaries, creates GitHub release
# 2. build-android: Builds Android AAR and APK, uploads to release
# 3. build-ios: Builds iOS XCFramework, uploads to release
```

To create a new release:
```bash
# Tag a version release
git tag v1.0.0
git push origin v1.0.0

# Or let nightly build run automatically (00:00 UTC daily)
```

Mobile artifacts are attached to the GitHub release:
- `venture.aar` - Android library for integration into apps
- `*.apk` - Android application package (unsigned)
- `Venture-ios-*.zip` - iOS framework for integration into apps

#### Android Build

Workflow file: `.github/workflows/android.yml`

```yaml
# Triggered on:
# - Push to main/develop with mobile code changes
# - Pull requests
# - Manual dispatch

# Secrets required for release builds:
# - ANDROID_KEYSTORE_FILE (base64 encoded)
# - ANDROID_KEYSTORE_PASSWORD
# - ANDROID_KEY_ALIAS
# - ANDROID_KEY_PASSWORD
```

#### iOS Build

Workflow file: `.github/workflows/ios.yml`

```yaml
# Triggered on:
# - Push to main/develop with mobile code changes
# - Pull requests
# - Manual dispatch

# Secrets required for device/IPA builds:
# - IOS_SIGNING_IDENTITY
# - IOS_PROVISIONING_PROFILE
# - IOS_TEAM_ID
```

### Setting Up Secrets

1. **Android Keystore**:
```bash
# Encode keystore to base64
base64 -i venture.keystore -o keystore.txt

# Add to GitHub Secrets:
# Repository → Settings → Secrets → New secret
# Name: ANDROID_KEYSTORE_FILE
# Value: <paste contents of keystore.txt>
```

2. **iOS Certificates**:
```bash
# Export certificate and provisioning profile from Keychain
# Add to GitHub Secrets similar to Android
```

### Manual Workflow Dispatch

```bash
# Trigger Android build via gh CLI
gh workflow run android.yml -f build_type=release

# Trigger iOS build
gh workflow run ios.yml -f build_type=ipa
```

## Troubleshooting

### Common Issues

#### Android

**Issue**: `ANDROID_HOME not set`
```bash
# Solution: Set environment variable
export ANDROID_HOME=~/Android/Sdk
export PATH=$PATH:$ANDROID_HOME/tools:$ANDROID_HOME/platform-tools
```

**Issue**: `NDK not found`
```bash
# Solution: Install NDK via SDK Manager
sdkmanager "ndk;26.1.10909125"
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.1.10909125
```

**Issue**: `Gradle build fails`
```bash
# Solution: Clean and rebuild
cd build/android
./gradlew clean
cd ../..
make android-apk
```

**Issue**: `Device unauthorized`
```bash
# Solution: Accept USB debugging prompt on device
adb kill-server
adb start-server
adb devices  # Should show device as "device" not "unauthorized"
```

#### iOS

**Issue**: `xcodebuild: command not found`
```bash
# Solution: Install Xcode Command Line Tools
xcode-select --install
```

**Issue**: `Code signing error`
```bash
# Solution: Verify signing identity
security find-identity -v -p codesigning

# Open Keychain Access and verify certificate is valid
```

**Issue**: `Provisioning profile not found`
```bash
# Solution: Download profiles manually
# 1. Open Xcode
# 2. Preferences → Accounts → Download Manual Profiles
# Or use: 
open ~/Library/MobileDevice/Provisioning\ Profiles/
```

**Issue**: `ebitenmobile bind fails on M1 Mac`
```bash
# Solution: Ensure Go and tools are native ARM64
which go
file $(which go)  # Should show "arm64"

# Reinstall if needed:
brew reinstall go
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
```

### Performance Issues

**Low FPS on device**:
- Ensure device is not in low-power mode
- Check for background processes
- Reduce particle effects in settings
- Lower resolution if needed

**High battery drain**:
- Implement battery-friendly mode in settings
- Reduce frame rate when idle
- Disable haptic feedback
- Lower screen brightness

**Large APK/IPA size**:
- Use ProGuard/R8 for Android (release builds)
- Enable bitcode for iOS
- Compress assets
- Remove unused code

### Testing Recommendations

#### Unit Tests

```bash
# Run all mobile tests
go test -tags test ./pkg/mobile/...

# Test touch input system
go test -tags test ./pkg/mobile -run TestTouchInput

# Benchmark touch processing
go test -tags test ./pkg/mobile -bench=BenchmarkTouchProcessing
```

#### Integration Tests

```bash
# Test on Android emulator
./scripts/test-android-emulator.sh

# Test on iOS simulator
./scripts/test-ios-simulator.sh
```

#### Device Testing

**Android**:
- Test on multiple screen sizes (phone, tablet)
- Test different Android versions (5.0 - 14.0)
- Test with different DPI settings
- Test in landscape and portrait modes
- Test with poor network connectivity

**iOS**:
- Test on iPhone (various models)
- Test on iPad
- Test with different iOS versions (14.0+)
- Test with VoiceOver enabled
- Test with reduced motion settings

### Getting Help

- **Documentation**: See `docs/` directory
- **Issues**: https://github.com/opd-ai/venture/issues
- **Discussions**: https://github.com/opd-ai/venture/discussions
- **Discord**: [Join our community]

## Version Information

- **Current Version**: 1.0.0
- **Minimum Android**: API 21 (Android 5.0)
- **Minimum iOS**: iOS 14.0
- **Supported Architectures**:
  - Android: armeabi-v7a, arm64-v8a
  - iOS: arm64

## Build Matrix

| Platform | Architecture | Min Version | Build Time | Binary Size |
|----------|--------------|-------------|------------|-------------|
| Android  | armeabi-v7a  | API 21      | ~3 min     | ~25 MB      |
| Android  | arm64-v8a    | API 21      | ~3 min     | ~30 MB      |
| iOS      | arm64        | iOS 14.0    | ~5 min     | ~35 MB      |

## Next Steps

1. Review mobile-specific code in `pkg/mobile/`
2. Customize touch controls for your game
3. Set up signing certificates
4. Configure CI/CD workflows
5. Test on physical devices
6. Prepare for app store submission

## Additional Resources

- [Ebiten Mobile Documentation](https://ebitengine.org/en/documents/mobile.html)
- [Android Developer Guide](https://developer.android.com/)
- [iOS Developer Guide](https://developer.apple.com/ios/)
- [ebitenmobile GitHub](https://github.com/hajimehoshi/ebiten/tree/main/mobile)
