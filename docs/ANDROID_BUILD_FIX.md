# Android Build Fix - Missing Launcher Icons

## Problem

The Android build was failing with the error:
```
AAPT: error: resource mipmap/ic_launcher (aka com.venture.game.debug:mipmap/ic_launcher) not found.
```

This occurred because the `AndroidManifest.xml` referenced launcher icons (`@mipmap/ic_launcher`) that didn't exist in the project resources.

## Solution

### 1. Created Icon Generation Script
- **File**: `scripts/generate-android-icons.sh`
- Automatically generates launcher icons for all Android densities:
  - mdpi (48x48), hdpi (72x72), xhdpi (96x96), xxhdpi (144x144), xxxhdpi (192x192)
- Uses ImageMagick if available to create PNG icons with blue background and white "V"
- Falls back to XML drawable icons if ImageMagick is not installed
- Creates adaptive icon resources for Android 8.0+ (API 26+)

### 2. Updated Build Script
- **File**: `scripts/build-android.sh`
- Added `generate_resources()` function
- Icon generation now runs automatically before AAR/APK build
- Ensures icons exist before Gradle attempts to link resources

### 3. Updated AndroidManifest
- **File**: `build/android/AndroidManifest.xml`
- Added `android:roundIcon="@mipmap/ic_launcher"` for round icon support
- Maintains compatibility with both legacy and adaptive icon systems

### 4. Updated CI/CD Workflow
- **File**: `.github/workflows/android.yml`
- Added ImageMagick installation step to ensure PNG icons are generated
- Icons will be automatically created in CI environment

### 5. Created Documentation
- **File**: `build/android/README.md`
- Comprehensive guide to Android build configuration
- Instructions for customizing icons, package name, version, etc.
- Troubleshooting section for common build issues

## Files Created/Modified

### Created:
- `scripts/generate-android-icons.sh` - Icon generation script
- `build/android/README.md` - Android build documentation
- `build/android/res/mipmap-*/ic_launcher.png` - Launcher icons (all densities)
- `build/android/res/mipmap-anydpi-v26/ic_launcher.xml` - Adaptive icon
- `build/android/res/drawable/ic_launcher_background.xml` - Adaptive icon background
- `build/android/res/drawable/ic_launcher_foreground.xml` - Adaptive icon foreground

### Modified:
- `scripts/build-android.sh` - Added icon generation step
- `build/android/AndroidManifest.xml` - Added round icon support
- `.github/workflows/android.yml` - Added ImageMagick installation

## Testing

Icons were successfully generated locally:
```bash
./scripts/generate-android-icons.sh
```

Output:
```
[INFO] Generating Android launcher icons...
[INFO] Creating mipmap directories...
[INFO] Generating PNG icons with ImageMagick...
[INFO] PNG icons generated successfully
[INFO] Creating adaptive icon resources...
[INFO] Adaptive icon created
[INFO] Icon generation complete!
```

All required resource files are now present in `build/android/res/`.

## Next Steps

The Android build should now succeed in GitHub Actions. The workflow will:
1. Install ImageMagick
2. Run the build script which generates icons
3. Build the AAR library
4. Build the APK without resource linking errors

## Customization

To customize the launcher icons:
1. Replace the generated PNG files in `build/android/res/mipmap-*/`
2. Modify the adaptive icon drawables in `build/android/res/drawable/`
3. Or edit `scripts/generate-android-icons.sh` to generate different icons

The current placeholder icons show a blue square with a white "V" letter, which can be replaced with proper app branding.
