# Mobile Release Integration

This document summarizes the mobile build integration added to the release workflow.

## Overview

The release workflow (`.github/workflows/release.yml`) now automatically builds mobile artifacts for iOS and Android alongside desktop builds when creating releases.

## Workflow Structure

**Total Lines**: 312 (added 160 lines for mobile builds)

**Triggers**:
- Semantic version tags (e.g., `v1.2.3`)
- Nightly cron schedule (00:00 UTC daily)

**Jobs**:
1. **release** - Builds desktop binaries (Linux, macOS, Windows) and creates GitHub release
2. **build-android** - Builds Android artifacts and uploads to release (depends on `release`)
3. **build-ios** - Builds iOS artifacts and uploads to release (depends on `release`)

## Mobile Build Outputs

### Android Artifacts

**Files**: 
- `venture.aar` - Android Archive library for integration into apps
- `app-release.apk` - Unsigned Android application package

**Build Process**:
1. Install ebitenmobile and Android SDK components
2. Build AAR using `ebitenmobile bind -target android`
3. Generate APK using Gradle with minimal project structure
4. Upload artifacts to GitHub release

**Runner**: `ubuntu-latest`

**Requirements**:
- Java 17 (Temurin distribution)
- Android SDK Build Tools 34.0.0
- Android Platform 34
- Android NDK 26.1.10909125

### iOS Artifacts

**Files**:
- `Venture-ios-{tag}.zip` - XCFramework archive for integration into Xcode projects

**Build Process**:
1. Install ebitenmobile
2. Build XCFramework using `ebitenmobile bind -target ios`
3. Archive framework as ZIP file
4. Upload artifact to GitHub release

**Runner**: `macos-13`

**Requirements**:
- Xcode 15.0+ (included in macOS runner)
- iOS SDK (included in Xcode)

## Release Creation

### Version Release

```bash
# Tag a semantic version
git tag v1.0.0
git push origin v1.0.0
```

This creates a release with:
- Desktop binaries (Linux/macOS/Windows for amd64/arm64)
- Android AAR and APK
- iOS XCFramework
- Release notes from git changelog

### Nightly Release

Automatically runs daily at 00:00 UTC, creating a "nightly" release with:
- Latest builds from main branch
- All desktop and mobile artifacts
- Recent commit history in release notes

## Artifact Usage

### Android AAR

```gradle
dependencies {
    implementation files('venture.aar')
}
```

Use in Android Studio projects to embed Venture as a game library.

### Android APK

Unsigned release APK. To sign for distribution:

```bash
# Sign with your keystore
jarsigner -verbose -sigalg SHA256withRSA -digestalg SHA-256 \
  -keystore your.keystore app-release.apk your-alias

# Align for optimization
zipalign -v 4 app-release.apk venture-signed.apk
```

### iOS XCFramework

```bash
# Extract framework
unzip Venture-ios-v1.0.0.zip

# Drag Venture.xcframework into Xcode project
# Add to "Frameworks, Libraries, and Embedded Content"
```

Use in Xcode projects to embed Venture as an iOS game library.

## CI/CD Pipeline Flow

```
┌─────────────────┐
│  Push Tag/Cron  │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────────┐
│  Job: release                       │
│  - Build desktop binaries           │
│  - Create GitHub release            │
│  - Upload desktop archives          │
└────────┬─────────────────┬──────────┘
         │                 │
         ▼                 ▼
┌────────────────┐  ┌─────────────────┐
│ Job: android   │  │  Job: ios       │
│ - Build AAR    │  │  - Build XCF    │
│ - Build APK    │  │  - Archive ZIP  │
│ - Upload       │  │  - Upload       │
└────────────────┘  └─────────────────┘
```

## Documentation

- **Build Guide**: `docs/MOBILE_BUILD.md` - Comprehensive build instructions
- **Quick Reference**: `docs/MOBILE_QUICK_REFERENCE.md` - Quick command reference
- **API Documentation**: `pkg/mobile/README.md` - Mobile package API reference

## Testing Release Workflow

To test the workflow locally before pushing a tag:

```bash
# Install act (GitHub Actions local runner)
# https://github.com/nektos/act

# Run release workflow
act -s GITHUB_TOKEN=your_token push
```

Note: Mobile builds require platform-specific tools, so local testing may not fully replicate CI environment.

## Troubleshooting

### Android Build Fails

- **Issue**: SDK components not found
- **Solution**: Check Android SDK installation in workflow logs
- **Verify**: `sdkmanager --list` shows required components

### iOS Build Fails

- **Issue**: ebitenmobile bind fails
- **Solution**: Check Go module dependencies are downloaded
- **Verify**: `go mod download` completes successfully

### Artifacts Not Attached

- **Issue**: Upload step succeeds but files missing
- **Solution**: Check file paths in workflow match actual build output
- **Verify**: `files:` patterns in `softprops/action-gh-release@v2`

## Future Enhancements

Potential improvements for mobile release workflow:

1. **Signed APK/AAB**: Add keystore secrets for signed Android releases
2. **iOS IPA**: Add provisioning profiles for signed iOS apps
3. **App Store Upload**: Automate submission to Google Play / App Store
4. **Version Bumping**: Auto-increment version codes in manifests
5. **Beta Channels**: Separate workflows for alpha/beta/production releases
6. **Size Optimization**: Enable ProGuard/R8 for Android, bitcode for iOS

## Summary

The release workflow now provides complete cross-platform build coverage:
- ✅ Desktop: Linux, macOS, Windows (x64/ARM64)
- ✅ Android: AAR library + APK app
- ✅ iOS: XCFramework library

All artifacts are automatically built and attached to GitHub releases, both for tagged versions and nightly builds.
