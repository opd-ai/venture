# Android Build Fix Summary

## Issue
The Android app was crashing on launch with:
```
java.lang.ClassNotFoundException: Didn't find class "org.golang.app.GoNativeActivity"
```

## Root Cause
The `GoNativeActivity` class is provided by the ebitenmobile library and is included in the AAR (Android Archive) file generated during the build process. The error occurred because:

1. The AAR file (`build/android/libs/mobile.aar`) was not being built before the APK
2. The Gradle build configuration didn't have proper error handling when the AAR was missing
3. No verification step existed to ensure all prerequisites were met
4. ProGuard rules were missing to prevent class stripping in release builds

## Changes Made

### 1. Enhanced Build Configuration (`build/android/build.gradle`)
- ✅ Added `jniLibs.srcDirs = ['libs']` to source sets to properly include native libraries
- ✅ Added validation to check if `mobile.aar` exists before building
- ✅ Added clear error messages when AAR is missing

### 2. Improved Build Script (`scripts/build-android.sh`)
- ✅ Added automatic installation of ebitenmobile if missing
- ✅ Added validation to ensure AAR was successfully created
- ✅ Added AAR contents listing for debugging
- ✅ Enhanced error messages with actionable steps

### 3. Added ProGuard Rules (`build/android/proguard-rules.pro`)
- ✅ Prevents stripping of `org.golang.app.**` classes
- ✅ Keeps all native methods intact
- ✅ Preserves gomobile and Ebiten classes
- ✅ Protects game package classes

### 4. Added Build Verification Script (`scripts/verify-android-build.sh`)
- ✅ Checks Go installation and version
- ✅ Verifies ebitenmobile is installed
- ✅ Validates Android SDK and NDK configuration
- ✅ Checks for AAR file and its contents
- ✅ Lists connected devices
- ✅ Provides clear next steps

### 5. Enhanced Documentation
- ✅ Created `docs/ANDROID_TROUBLESHOOTING.md` - Comprehensive troubleshooting guide
- ✅ Updated `build/android/README.md` - Added visual build process diagram
- ✅ Created `build/android/QUICKREF.md` - Quick reference for common tasks
- ✅ Added `build/android/gradle.properties` - Gradle optimization settings

### 6. Updated Makefile (`Makefile.mobile`)
- ✅ Added `android-verify` target to run verification script
- ✅ Updated help text to include verification command

## Build Process Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Android Build Process                        │
└─────────────────────────────────────────────────────────────────┘

Step 1: Build AAR (ebitenmobile bind)
┌──────────────┐     ┌──────────────┐     ┌──────────────────┐
│ cmd/mobile/  │ --> │ ebitenmobile │ --> │ libs/mobile.aar  │
│  mobile.go   │     │     bind     │     │ (GoNativeActivity)│
└──────────────┘     └──────────────┘     └──────────────────┘
                                                    ↓
                                           Contains:
                                           • GoNativeActivity class
                                           • libgojni.so (native code)
                                           • Android resources

Step 2: Build APK (Gradle)
┌──────────────────┐     ┌─────────────┐     ┌──────────────┐
│ libs/mobile.aar  │     │  Gradle     │     │ Venture.apk  │
│ AndroidManifest  │ --> │ assembleDebug│ --> │ (Installable)│
│ Resources (res/) │     │             │     │              │
└──────────────────┘     └─────────────┘     └──────────────┘
```

## How to Use

### Quick Fix (Recommended)
```bash
# Verify environment
make android-verify

# Build and install
make android-install
```

### Manual Fix
```bash
# Clean old builds
rm -rf build/android/libs build/android/build

# Build AAR
./scripts/build-android.sh aar

# Verify AAR exists and contains required classes
ls -lh build/android/libs/mobile.aar
unzip -l build/android/libs/mobile.aar | grep GoNativeActivity

# Build and install APK
./scripts/build-android.sh install
```

### Verification Steps
```bash
# 1. Check prerequisites
make android-verify

# 2. Verify AAR contents
unzip -l build/android/libs/mobile.aar

# 3. Check APK contents
unzip -l dist/android/Venture-*.apk | grep -E "\.so|classes"

# 4. Install on device
adb install -r dist/android/Venture-1.0.0-debug.apk

# 5. Monitor for crashes
adb logcat | grep venture
```

## Prevention

The build script now automatically:
1. ✅ Checks for ebitenmobile and installs if missing
2. ✅ Validates AAR was successfully created
3. ✅ Shows AAR contents for verification
4. ✅ Fails with clear error messages if prerequisites are missing

The Gradle configuration now:
1. ✅ Validates AAR exists before attempting to build
2. ✅ Includes native libraries directory in source sets
3. ✅ Provides clear error messages when AAR is missing
4. ✅ Protects required classes from ProGuard stripping

## Testing

To verify the fix works:

```bash
# 1. Clean everything
rm -rf build/android/libs build/android/build dist/android

# 2. Run verification
make android-verify

# 3. Build and install
make android-install

# 4. Launch app on device and verify it doesn't crash
adb shell am start -n com.venture.game.debug/org.golang.app.GoNativeActivity

# 5. Check logs for successful startup
adb logcat | grep -E "venture|GoNativeActivity"
```

## Related Documentation

- **Comprehensive Guide**: `docs/ANDROID_TROUBLESHOOTING.md`
- **Build Instructions**: `build/android/README.md`
- **Quick Reference**: `build/android/QUICKREF.md`
- **Mobile Build Guide**: `docs/MOBILE_BUILD.md`

## Summary

The fix ensures that:
1. The AAR file containing `GoNativeActivity` is always built before the APK
2. Clear error messages guide users when prerequisites are missing
3. ProGuard rules protect required classes in release builds
4. Verification tools help diagnose issues before they occur
5. Documentation provides clear troubleshooting steps

The app should now build and launch successfully on Android devices without the `ClassNotFoundException` error.
