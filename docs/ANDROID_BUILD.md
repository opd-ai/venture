# Android APK Build Documentation

## Overview

This document describes the Android APK build process for the Venture game using Ebiten and ebitenmobile.

## Build Status

✅ **APK Build: SUCCESSFUL**

- **APK Location**: `dist/android/Venture-1.0.0-debug.apk`
- **APK Size**: 85 MB
- **Build Time**: ~2 minutes (first build), ~10 seconds (incremental)
- **Target Architectures**: arm64-v8a, armeabi-v7a, x86, x86_64

## Prerequisites

### Required Software

1. **Go 1.24.5+**
   ```bash
   go version  # Should show go1.24.5 or higher
   ```

2. **Android SDK**
   ```bash
   export ANDROID_HOME=/usr/local/lib/android/sdk
   ```

3. **Android NDK**
   ```bash
   export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/27.3.13750724
   ```

4. **Java 11+**
   ```bash
   java -version  # Should show version 11 or higher
   ```

5. **ebitenmobile**
   ```bash
   go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

### Environment Setup

Add to your `~/.bashrc` or `~/.zshrc`:

```bash
export ANDROID_HOME=/usr/local/lib/android/sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/27.3.13750724
export PATH=$PATH:$(go env GOPATH)/bin:$ANDROID_HOME/platform-tools
```

## Build Process

### Quick Build (Using Makefile)

```bash
# Verify environment
make android-verify

# Build debug APK
make android-apk

# Build and install on connected device
make android-install
```

### Manual Build

#### Step 1: Build AAR Library

```bash
cd /path/to/venture
export PATH=$PATH:$(go env GOPATH)/bin

ebitenmobile bind \
  -target android \
  -javapkg com.venture.game \
  -o build/android/libs/mobile.aar \
  -androidapi 21 \
  ./cmd/mobile
```

This creates a 43MB AAR file containing:
- Native libraries (libgojni.so) for all architectures
- Java bindings for the Go mobile code
- GoNativeActivity class for the Android activity

#### Step 2: Build APK with Gradle

```bash
cd build/android

# Using bash to avoid shell compatibility issues
bash gradlew assembleDebug
```

The APK will be created at:
```
build/android/build/outputs/apk/debug/android-debug.apk
```

#### Step 3: Copy to Distribution Directory

```bash
mkdir -p dist/android
cp build/android/build/outputs/apk/debug/android-debug.apk \
   dist/android/Venture-1.0.0-debug.apk
```

## Build Configuration

### AndroidManifest.xml

Key configuration:
- **Package**: `com.venture.game`
- **Min SDK**: 21 (Android 5.0)
- **Target SDK**: 34 (Android 14)
- **Activity**: `org.ebitengine.gomobile.GoNativeActivity`
- **Screen Orientation**: Portrait
- **Permissions**: Internet, Network State, Vibrate

Critical meta-data:
```xml
<meta-data android:name="android.app.lib_name" android:value="mobile" />
```
This tells Android to load the `libgojni.so` library (built from `cmd/mobile`).

### build.gradle

Key settings:
- **Android Gradle Plugin**: 7.4.2
- **Gradle Version**: 7.6
- **Compile SDK**: 34
- **Java Version**: 11
- **Supported ABIs**: armeabi-v7a, arm64-v8a, x86, x86_64

### Mobile Package Structure

The `cmd/mobile` package:
- Uses `package mobile` (required by ebitenmobile)
- Calls `mobile.SetGame()` in `init()` function
- Implements Ebiten game interface (Update, Draw, Layout)
- Does NOT have a `main()` function (handled by gomobile)

## Verification

### Automated Verification

```bash
./scripts/test-android-apk.sh
```

This script checks:
- APK existence and size
- AndroidManifest.xml presence
- Native libraries for all architectures
- Resources and DEX files

### Manual Verification

```bash
# List APK contents
unzip -l dist/android/Venture-1.0.0-debug.apk

# Check native libraries
unzip -l dist/android/Venture-1.0.0-debug.apk | grep libgojni.so

# Extract and view manifest
unzip -p dist/android/Venture-1.0.0-debug.apk AndroidManifest.xml | \
  adb shell "cat > /tmp/manifest.xml && aapt2 dump xmltree /tmp/manifest.xml"
```

## Installation

### Install on Connected Device

```bash
# Check connected devices
adb devices

# Install APK
adb install -r dist/android/Venture-1.0.0-debug.apk

# Monitor logs during runtime
adb logcat | grep -E "(Venture|GoLog|ebitengine)"
```

### Install via APK File

1. Copy APK to device
2. Enable "Install from Unknown Sources" in Android settings
3. Tap the APK file to install

## Troubleshooting

### Issue: "ebitenmobile: command not found"

**Solution**:
```bash
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

### Issue: "ANDROID_HOME is not set"

**Solution**:
```bash
export ANDROID_HOME=/usr/local/lib/android/sdk
export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/27.3.13750724
```

### Issue: Gradle wrapper fails with "cd: Illegal option -r"

**Solution**: Use `bash` explicitly:
```bash
bash gradlew assembleDebug
```

This is due to dash shell incompatibility on some systems.

### Issue: "resource mipmap/ic_launcher not found"

**Solution**: Ensure `build.gradle` includes the `res` directory:
```groovy
sourceSets {
    main {
        manifest.srcFile 'AndroidManifest.xml'
        res.srcDirs = ['res', 'src/main/res']
    }
}
```

### Issue: AAR doesn't contain GoNativeActivity

**Solution**: This is expected for newer ebitenmobile versions. The GoNativeActivity class comes from the ebitengine/gomobile dependency, not the AAR itself.

## Build Artifacts

### Directory Structure

```
build/android/
├── AndroidManifest.xml          # App manifest
├── build.gradle                 # Gradle build script
├── settings.gradle              # Gradle settings
├── gradle.properties            # Gradle properties
├── proguard-rules.pro          # ProGuard rules
├── gradle/
│   └── wrapper/                # Gradle wrapper files
├── libs/
│   └── mobile.aar              # Go mobile library (43MB)
├── res/                        # App resources
│   ├── drawable/               # Launcher icons
│   ├── mipmap-*/              # Density-specific icons
│   └── values/                # Strings and values
└── build/
    └── outputs/
        └── apk/
            └── debug/
                └── android-debug.apk  # Final APK (85MB)
```

### Artifact Sizes

- **AAR Library**: 43 MB (native code for all architectures)
- **Final APK**: 85 MB (AAR + Android framework + resources)
- **Native Library Breakdown**:
  - arm64-v8a: 22.3 MB
  - armeabi-v7a: 21.0 MB
  - x86: 21.6 MB
  - x86_64: 23.5 MB

## Release Build

To create a release APK (requires signing configuration):

```bash
# Create keystore (first time only)
keytool -genkey -v -keystore venture.keystore \
  -alias venture -keyalg RSA -keysize 2048 -validity 10000

# Create signing configuration
cat > build/android/keystore.properties << EOF
storePassword=YOUR_STORE_PASSWORD
keyPassword=YOUR_KEY_PASSWORD
keyAlias=venture
storeFile=venture.keystore
EOF

# Build release APK
make android-apk-release

# Or manually:
cd build/android
bash gradlew assembleRelease
```

## Performance Notes

- **Cold Build**: ~2 minutes (includes AAR compilation)
- **Incremental Build**: ~10 seconds (APK assembly only)
- **APK Installation**: ~30 seconds on device
- **First Launch**: ~3-5 seconds (JIT compilation)
- **Subsequent Launches**: ~1-2 seconds

## Platform-Specific Code

The mobile build uses build tags to include platform-specific code:

```go
// +build android ios

package mobile

// Mobile-specific implementation
```

For desktop-only code:
```go
// +build !android,!ios

package main

// Desktop-specific implementation
```

## Future Enhancements

- [ ] Add release signing configuration
- [ ] Implement Android App Bundle (.aab) builds for Play Store
- [ ] Add automated APK upload to GitHub Releases
- [ ] Implement crash reporting (Firebase Crashlytics)
- [ ] Add in-app update mechanism
- [ ] Optimize APK size (strip symbols, compress resources)

## References

- [Ebiten Mobile Documentation](https://ebitengine.org/en/documents/mobile.html)
- [ebitenmobile Command](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile)
- [Android Gradle Plugin](https://developer.android.com/build)
- [Go Mobile](https://github.com/golang/mobile)
