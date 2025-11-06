# Android Build Troubleshooting

## ClassNotFoundException: org.golang.app.GoNativeActivity

### Symptom

```
java.lang.ClassNotFoundException: Didn't find class "org.golang.app.GoNativeActivity"
```

The app crashes immediately on launch with this error.

### Root Cause

The `GoNativeActivity` class is provided by the gomobile/ebitenmobile library and is included in the AAR (Android Archive) file generated during the build process. This error occurs when:

1. The AAR file was not built before building the APK
2. The AAR file was not properly included in the APK
3. The AAR file is corrupted or incomplete
4. ProGuard/R8 stripped the class during obfuscation (release builds)

### Solution

#### Step 1: Verify ebitenmobile Installation

```bash
# Check if ebitenmobile is installed
which ebitenmobile

# If not installed, install it
go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest

# Verify installation
ebitenmobile version
```

#### Step 2: Build the AAR File

```bash
# Navigate to project root
cd /path/to/venture

# Build the AAR library
./scripts/build-android.sh aar

# Verify the AAR was created
ls -lh build/android/libs/mobile.aar
```

The AAR file should be several megabytes in size. If it's missing or very small, the build failed.

#### Step 3: Verify AAR Contents

```bash
# List contents of the AAR
unzip -l build/android/libs/mobile.aar

# Check for GoNativeActivity class
unzip -l build/android/libs/mobile.aar | grep GoNativeActivity
```

You should see entries like:
```
classes.jar
AndroidManifest.xml
jni/arm64-v8a/libgojni.so
jni/armeabi-v7a/libgojni.so
```

The `classes.jar` inside the AAR contains the `GoNativeActivity` class.

#### Step 4: Rebuild the APK

```bash
# Clean previous builds
cd build/android
./gradlew clean

# Return to project root
cd ../..

# Build fresh APK
./scripts/build-android.sh apk
```

#### Step 5: Verify APK Contents

```bash
# Check APK contents
unzip -l dist/android/Venture-*.apk | grep -E "\.so|classes"

# Look for native libraries
unzip -l dist/android/Venture-*.apk | grep libgojni.so
```

You should see:
```
lib/arm64-v8a/libgojni.so
lib/armeabi-v7a/libgojni.so
classes.dex (or classes2.dex, classes3.dex)
```

### Alternative: Use Makefile Targets

The Makefile handles the build process correctly:

```bash
# Build debug APK (includes AAR build)
make android-apk

# Or install directly on device
make android-install
```

### For Release Builds

If the error occurs in release builds but not debug builds, ProGuard may be stripping the class.

The project includes ProGuard rules (`build/android/proguard-rules.pro`) to prevent this:

```proguard
# Keep GoNativeActivity and all Ebiten/gomobile classes
-keep class org.golang.app.** { *; }
-keep class go.** { *; }
-keep class mobile.** { *; }

# Keep all native methods
-keepclasseswithmembernames class * {
    native <methods>;
}
```

Verify these rules are being applied in `build.gradle`:

```gradle
buildTypes {
    release {
        minifyEnabled true
        proguardFiles getDefaultProguardFile('proguard-android-optimize.txt'), 'proguard-rules.pro'
    }
}
```

### Common Mistakes

1. **Building APK without building AAR first**
   ```bash
   # WRONG - Don't do this:
   cd build/android
   ./gradlew assembleDebug
   
   # RIGHT - Use the build script:
   ./scripts/build-android.sh apk
   ```

2. **Missing ebitenmobile installation**
   ```bash
   # Install ebitenmobile
   go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest
   
   # Add to PATH if needed
   export PATH=$PATH:$(go env GOPATH)/bin
   ```

3. **Stale AAR file**
   ```bash
   # Remove old AAR
   rm build/android/libs/mobile.aar
   
   # Rebuild
   ./scripts/build-android.sh aar
   ```

4. **Android SDK/NDK not configured**
   ```bash
   # Set environment variables
   export ANDROID_HOME=/path/to/android-sdk
   export ANDROID_NDK_HOME=$ANDROID_HOME/ndk/26.1.10909125
   ```

### Debugging Steps

#### 1. Check Build Logs

```bash
# Build with verbose logging
./scripts/build-android.sh apk 2>&1 | tee build.log

# Check for errors
grep -i error build.log
```

#### 2. Verify Package Name

Ensure the package name in `AndroidManifest.xml` matches the one used in `ebitenmobile bind`:

```xml
<!-- AndroidManifest.xml -->
<manifest package="com.venture.game">
```

```bash
# build-android.sh
ebitenmobile bind -javapkg com.venture.game ...
```

#### 3. Check Device Architecture

```bash
# List connected devices
adb devices

# Check device architecture
adb shell getprop ro.product.cpu.abi

# Build supports: armeabi-v7a, arm64-v8a
```

#### 4. Install Manually

```bash
# Uninstall old version
adb uninstall com.venture.game.debug

# Install fresh
adb install dist/android/Venture-1.0.0-debug.apk

# Check logcat for errors
adb logcat | grep venture
```

### Still Having Issues?

1. Clean everything:
   ```bash
   # Clean Go cache
   go clean -cache -modcache
   
   # Clean build directories
   rm -rf build/android/build build/android/libs dist/android
   
   # Rebuild from scratch
   ./scripts/build-android.sh apk
   ```

2. Verify Go version:
   ```bash
   go version
   # Should be 1.24.5 or later
   ```

3. Update dependencies:
   ```bash
   go get -u github.com/hajimehoshi/ebiten/v2
   go mod tidy
   ```

4. Check for conflicting libraries:
   ```bash
   # Remove any conflicting AARs
   rm build/android/libs/*.aar
   
   # Rebuild AAR
   ./scripts/build-android.sh aar
   ```

### Reference

- **Ebiten Mobile Documentation**: https://ebiten.org/documents/mobile.html
- **ebitenmobile GitHub**: https://github.com/hajimehoshi/ebiten/tree/main/cmd/ebitenmobile
- **Android NDK**: https://developer.android.com/ndk
- **ProGuard Rules**: https://www.guardsquare.com/manual/configuration/usage

### Quick Fix Summary

```bash
# The fastest way to fix this issue:
cd /path/to/venture
rm -rf build/android/libs build/android/build
./scripts/build-android.sh install
```

This will:
1. Remove old build artifacts
2. Build fresh AAR with GoNativeActivity
3. Build and install APK on connected device
