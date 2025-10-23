# CI/CD Build System Documentation

## Overview

Venture uses platform-native build jobs for all targets instead of cross-compilation. Each platform builds on its native operating system using GitHub Actions runners or specialized build environments.

## Build Matrix

### Desktop Platforms

| Platform | Runner | Architectures | Build Time |
|----------|--------|---------------|------------|
| Linux | `ubuntu-latest` | amd64, arm64 | ~2-3 min |
| Windows | `windows-latest` | amd64 | ~2-3 min |
| macOS | `macos-latest` | amd64, arm64 | ~3-4 min |

### Mobile Platforms

| Platform | Runner | Output | Build Time |
|----------|--------|--------|------------|
| Android | `ubuntu-latest` | AAR library | ~5-7 min |
| iOS | `macos-13` | XCFramework | ~5-7 min |

## Workflows

### Build Workflow (`.github/workflows/build.yml`)

**Trigger:** Push to `main` branch

**Purpose:** Continuous integration builds for all platforms

**Jobs:**
- `build-linux` - Builds Linux binaries (amd64, arm64)
- `build-windows` - Builds Windows binaries (amd64)
- `build-macos` - Builds macOS binaries (amd64, arm64)
- `build-android` - Builds Android AAR library
- `build-ios` - Builds iOS XCFramework

**Artifacts:** All build outputs are uploaded as artifacts with 7-day retention

### Release Workflow (`.github/workflows/release.yml`)

**Trigger:** 
- Nightly at 00:00 UTC (creates `nightly` tag)
- Push of semantic version tags (e.g., `v1.2.3`)

**Purpose:** Create releases with downloadable binaries

**Jobs:**
1. `prepare-release` - Determines release type, generates notes
2. `build-linux` - Builds Linux binaries (amd64, arm64)
3. `build-windows` - Builds Windows binaries (amd64)
4. `build-macos` - Builds macOS binaries (amd64, arm64)
5. `build-android` - Builds Android AAR
6. `build-ios` - Builds iOS XCFramework
7. `publish-release` - Collects all artifacts and creates GitHub release

**Release Artifacts:**
- `venture-server-linux-{amd64,arm64}.tar.gz`
- `venture-client-linux-{amd64,arm64}.tar.gz`
- `venture-server-darwin-{amd64,arm64}.tar.gz`
- `venture-client-darwin-{amd64,arm64}.tar.gz`
- `venture-server-windows-amd64.zip`
- `venture-client-windows-amd64.zip`
- `venture-android-{tag}.zip` (contains AAR)
- `venture-ios-{tag}.zip` (contains XCFramework)

### Platform-Specific Workflows

#### Android Build (`.github/workflows/android.yml`)

**Trigger:** Push/PR to `main`/`develop` affecting mobile code

**Purpose:** Specialized Android builds and testing

**Variants:**
- Debug APK (default)
- Release APK (with signing)

#### iOS Build (`.github/workflows/ios.yml`)

**Trigger:** Push/PR to `main`/`develop` affecting mobile code

**Purpose:** Specialized iOS builds and testing

**Variants:**
- Simulator build
- Device build (requires signing)
- IPA export (requires signing)

## Build Scripts

### Desktop Build Scripts

Located in `scripts/`:

- `build-linux.sh [amd64|arm64]` - Build Linux binaries
- `build-windows.sh [amd64]` - Build Windows binaries  
- `build-macos.sh [amd64|arm64]` - Build macOS binaries

**Usage:**
```bash
# Build for current platform's default architecture
./scripts/build-linux.sh

# Build for specific architecture
./scripts/build-linux.sh arm64
./scripts/build-macos.sh amd64
```

**Output:** Binaries in `build/` directory, archives in `dist/{platform}/`

### Mobile Build Scripts

- `build-android.sh [aar|apk|apk-release|aab|install]` - Android builds
- `build-ios.sh [xcframework|simulator|device|ipa|install]` - iOS builds

See `pkg/mobile/README.md` for detailed mobile build documentation.

## Runner Requirements

### Linux (Ubuntu Latest)

**Installed:**
- Go 1.24.x
- Build dependencies: libc6-dev, libgl1-mesa-dev, libxcursor-dev, libxi-dev, libxinerama-dev, libxrandr-dev, libxxf86vm-dev, libasound2-dev, pkg-config

**Native Builds:** Linux amd64, Linux arm64 (via GOARCH)

### Windows (Windows Latest)

**Installed:**
- Go 1.24.x
- PowerShell 7+

**Native Builds:** Windows amd64

### macOS (macOS Latest / macOS 13)

**Installed:**
- Go 1.24.x
- Xcode 15.0 (for iOS builds)

**Native Builds:** 
- macOS amd64, macOS arm64 (via GOARCH)
- iOS (via ebitenmobile)

### Android Build Environment

**Installed:**
- Go 1.24.x
- Java 17 (Temurin distribution)
- Android SDK
  - Build Tools 34.0.0
  - Platform API 34
  - NDK 26.1.10909125
- ebitenmobile

**Native Builds:** Android AAR/APK/AAB (all architectures: armeabi-v7a, arm64-v8a)

### iOS Build Environment

**Installed:**
- Go 1.24.x
- Xcode 15.0
- ebitenmobile

**Native Builds:** iOS XCFramework (arm64, simulator)

## Advantages of Native Builds

### Performance
- **Faster compilation:** No cross-compilation overhead
- **Better optimization:** Native toolchains produce more optimized binaries
- **Parallel builds:** Matrix strategy runs builds simultaneously

### Reliability
- **Reduced errors:** No cross-compilation toolchain issues
- **Platform verification:** Binaries tested on target OS
- **Dependency handling:** Native package managers work correctly

### Maintainability
- **Simpler setup:** No complex cross-compilation toolchains
- **Clear separation:** Each platform has dedicated job
- **Easier debugging:** Platform-specific issues isolated

## Build Time Comparison

### Before (Cross-Compilation)
- Single job building all platforms sequentially: ~15-20 minutes
- All platforms built on Ubuntu with cross-compilation toolchains

### After (Native Builds)
- Parallel jobs on native platforms: ~5-7 minutes (wall time)
- Total compute time similar, but much faster end-to-end

**Improvement:** ~60-70% faster wall-clock time due to parallelization

## Environment Variables

### Common
- `GOARCH` - Target architecture (amd64, arm64)
- `GOOS` - Target OS (linux, darwin, windows)

### Android
- `ANDROID_HOME` - Android SDK location
- `ANDROID_NDK_HOME` - Android NDK location
- `VENTURE_KEYSTORE_FILE` - Keystore for release signing (base64)
- `VENTURE_KEYSTORE_PASSWORD` - Keystore password
- `VENTURE_KEY_ALIAS` - Key alias
- `VENTURE_KEY_PASSWORD` - Key password

### iOS
- `IOS_SIGNING_IDENTITY` - Code signing identity
- `IOS_PROVISIONING_PROFILE` - Provisioning profile
- `IOS_TEAM_ID` - Apple Developer Team ID

## Caching Strategy

### Go Modules
- Cached using `actions/setup-go@v5` with `cache: true`
- Cache key: Hash of `go.sum`
- Speeds up dependency download

### Platform-Specific
- **Android:** Gradle cache via `cache: 'gradle'` in `setup-java`
- **iOS:** Derived data and Go modules cached

## Troubleshooting

### Linux Build Fails
- **Issue:** Missing system libraries
- **Solution:** Update `apt-get install` step with required libraries

### Windows Build Fails
- **Issue:** PowerShell path issues
- **Solution:** Use `New-Item` for directory creation, check path separators

### macOS Build Fails
- **Issue:** Architecture mismatch
- **Solution:** Verify `GOARCH` is set correctly for target

### Android Build Fails
- **Issue:** NDK/SDK not found
- **Solution:** Check `sdkmanager` installed correct versions

### iOS Build Fails
- **Issue:** Xcode version mismatch
- **Solution:** Use `xcode-select` to set correct version

## Manual Builds

### Local Development

```bash
# Clone repository
git clone https://github.com/opd-ai/venture.git
cd venture

# Build for current platform
./scripts/build-linux.sh      # Linux
./scripts/build-macos.sh       # macOS
./scripts/build-windows.sh     # Windows (requires WSL or Git Bash)

# Build mobile
./scripts/build-android.sh aar
./scripts/build-ios.sh xcframework
```

### Using Makefile

```bash
# Build mobile targets
make mobile-deps           # Install dependencies
make android-aar          # Build Android AAR
make ios-xcframework      # Build iOS XCFramework
```

## Security Considerations

### Secrets Required for Release Builds

**Android Release:**
- `ANDROID_KEYSTORE_FILE` - Base64-encoded keystore
- `ANDROID_KEYSTORE_PASSWORD`
- `ANDROID_KEY_ALIAS`
- `ANDROID_KEY_PASSWORD`

**iOS Release:**
- `IOS_SIGNING_IDENTITY`
- `IOS_PROVISIONING_PROFILE`
- `IOS_TEAM_ID`

### Best Practices
- Store all secrets in GitHub Secrets
- Never commit signing keys or certificates
- Use separate signing keys for debug/release
- Rotate keys periodically

## Deployment Workflow

### Nightly Builds
1. Cron triggers at 00:00 UTC
2. All platforms build in parallel
3. Release created with `nightly` tag (overwrites previous)
4. Artifacts available for download

### Version Releases
1. Push tag matching `v*.*.*` pattern
2. All platforms build in parallel
3. Changelog generated from commits
4. Release created with version tag
5. Artifacts available for download

### Distribution
- GitHub Releases for desktop binaries
- Android AAR for library integration
- iOS XCFramework for library integration
- Future: App stores (Google Play, Apple App Store)

## Monitoring and Notifications

- **Build Status:** Visible in GitHub Actions tab
- **Failures:** GitHub sends email notifications
- **Artifacts:** Available in workflow run page
- **Releases:** Listed on GitHub Releases page

## Future Enhancements

- [ ] Docker container builds for reproducibility
- [ ] Code signing for all platforms
- [ ] Automated testing before release
- [ ] App store publishing automation
- [ ] Build time metrics and tracking
- [ ] Notification webhooks (Discord, Slack)
- [ ] Artifact checksum verification
- [ ] Multi-arch Docker images
