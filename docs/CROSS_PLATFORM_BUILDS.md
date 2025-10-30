# Cross-Platform Build Support

This document describes the cross-platform build support for Android, iOS, and WebAssembly (WASM) targets.

## Overview

The Venture game engine supports building for multiple platforms:
- **Desktop**: Linux, macOS, Windows (primary platforms)
- **Mobile**: Android, iOS (via stubs for regular builds, ebitenmobile for full builds)
- **Web**: WebAssembly/JavaScript

## Build Status

### Android

| Target | Architecture | Status | Notes |
|--------|-------------|--------|-------|
| cmd/client | arm64 | ✅ PASS | Stub implementation |
| cmd/client | amd64 | ⚠️ SKIP | Requires CGO (x86 emulator) |
| cmd/server | arm64 | ✅ PASS | Stub implementation |
| cmd/server | amd64 | ⚠️ SKIP | Requires CGO (x86 emulator) |

**Build Command:**
```bash
# ARM64 (most Android devices)
GOOS=android GOARCH=arm64 go build ./cmd/client
GOOS=android GOARCH=arm64 go build ./cmd/server

# Note: These produce stub binaries that direct users to use ebitenmobile
```

### iOS

| Target | Architecture | Status | Notes |
|--------|-------------|--------|-------|
| cmd/client | amd64 | ✅ PASS | Simulator builds |
| cmd/client | arm64 | ⚠️ SKIP | Requires CGO (device builds) |
| cmd/server | amd64 | ✅ PASS | Simulator builds |
| cmd/server | arm64 | ⚠️ SKIP | Requires CGO (device builds) |

**Build Commands:**
```bash
# iOS Simulator (amd64 - Intel Macs, x86_64 simulator)
CGO_ENABLED=0 GOOS=ios GOARCH=amd64 go build -buildmode=exe ./cmd/client
CGO_ENABLED=0 GOOS=ios GOARCH=amd64 go build -buildmode=exe ./cmd/server

# iOS Simulator (arm64 - Apple Silicon Macs, ARM64 simulator)
CGO_ENABLED=0 GOOS=ios GOARCH=arm64 go build -buildmode=exe ./cmd/client
# Note: This may still require CGO depending on Go version

# iOS Device (arm64 - actual iPhones/iPads)
# Cannot build without CGO and iOS SDK
# Use ebitenmobile instead
```

**iOS arm64 Device Limitation:**

iOS device builds (arm64) require:
1. CGO enabled
2. iOS SDK headers and toolchain
3. External linking support

This is a fundamental Go toolchain requirement, not a limitation of this project. The standard approach is to use `ebitenmobile bind` which handles all the complexity:

```bash
ebitenmobile bind -target ios -o Mobile.xcframework ./cmd/mobile
```

### WebAssembly (WASM)

| Target | Status | Notes |
|--------|--------|-------|
| cmd/client | ✅ PASS | Full browser support |
| cmd/server | ✅ PASS | Can run in Node.js |

**Build Commands:**
```bash
GOOS=js GOARCH=wasm go build -o client.wasm ./cmd/client
GOOS=js GOARCH=wasm go build -o server.wasm ./cmd/server
```

## Package-Level Build Support

All core packages build successfully for Android/iOS/WASM:

✅ **Fully Supported (no Ebiten dependencies):**
- `pkg/procgen/*` - All procedural generation packages
- `pkg/combat` - Combat mechanics
- `pkg/world` - World state management
- `pkg/saveload` - Save/load system
- `pkg/logging` - Structured logging
- `pkg/audio/*` - Audio synthesis (no playback, generation only)
- `pkg/rendering/lighting` - Lighting calculations
- `pkg/rendering/palette` - Color palette generation
- `pkg/rendering/patterns` - Pattern generation
- `pkg/rendering/tiles` - Tile data structures
- `pkg/rendering/ui` - UI data structures
- `pkg/visualtest` - Visual testing utilities

⚠️ **Requires ebitenmobile (Ebiten dependencies):**
- `pkg/engine` - Game engine (uses ebiten.Image, ebiten.Game)
- `pkg/rendering/sprites` - Sprite rendering (uses ebiten.Image)
- `pkg/rendering/cache` - Sprite caching (uses ebiten.Image)
- `pkg/rendering/pool` - Object pooling (uses ebiten.Image)
- `pkg/rendering/shapes` - Shape rendering (uses ebiten.Image)
- `pkg/procgen/recipe` - Recipe generation (imports engine)
- `pkg/network` - Networking (imports engine)
- `pkg/hostplay` - Host-and-play (imports engine)
- `pkg/mobile` - Mobile controls (uses ebiten directly)

## Mobile Build Approaches

### Approach 1: Stub Binaries (Current Implementation)

Regular `go build` produces stub binaries that:
- Compile successfully for all supported GOOS/GOARCH combinations
- Print helpful error messages at runtime
- Direct users to the correct build approach

**Pros:**
- ✅ Satisfies `go build` requirements
- ✅ Clear error messages for users
- ✅ No CGO complications
- ✅ Fast compilation

**Cons:**
- ❌ Not functional binaries
- ❌ Requires ebitenmobile for actual mobile apps

### Approach 2: ebitenmobile (Production Mobile Builds)

Use `ebitenmobile bind` for actual mobile applications:

```bash
# Android
ebitenmobile bind -target android -o mobile.aar ./cmd/mobile

# iOS
ebitenmobile bind -target ios -o Mobile.xcframework ./cmd/mobile
```

**Pros:**
- ✅ Full Ebiten functionality
- ✅ Native mobile integration
- ✅ Haptic feedback, sensors, etc.
- ✅ App Store/Play Store compatible

**Cons:**
- ⚠️ Requires Android SDK/NDK or Xcode
- ⚠️ More complex build process
- ⚠️ Platform-specific tooling

## Testing

### Test Suite Results

```bash
# Run all testable packages (no X11/graphics required)
go test ./pkg/procgen/... ./pkg/combat ./pkg/world ./pkg/saveload \
        ./pkg/logging ./pkg/audio/... ./pkg/rendering/lighting \
        ./pkg/rendering/palette ./pkg/rendering/particles \
        ./pkg/rendering/patterns ./pkg/rendering/tiles \
        ./pkg/rendering/ui ./pkg/visualtest

# Result: 25/26 packages pass (recipe fails due to engine dependency)
```

### Cross-Platform Build Test

Run the test script to verify all platform builds:

```bash
./scripts/test-cross-platform-builds.sh
```

**Expected Results:**
- ✅ 24 builds pass
- ⚠️ 4 builds skipped (require CGO)
- ❌ 0 builds fail

## Build Tags

### cmd/client and cmd/server

**Desktop/WASM Build** (`main.go`):
```go
//go:build !android && !ios
// +build !android,!ios
```

**Mobile Stub** (`main_mobile.go`):
```go
//go:build android || ios
// +build android ios
```

### pkg/mobile Platform Files

**Android Platform** (`platform_android.go`):
```go
//go:build android && cgo && ebitenmobilebind
// +build android,cgo,ebitenmobilebind
```

**iOS Platform** (`platform_ios.go`):
```go
//go:build ios && cgo && ebitenmobilebind
// +build ios,cgo,ebitenmobilebind
```

These files are only included when building with `ebitenmobile bind`, not with regular `go build`.

## Common Issues

### 1. "requires external (cgo) linking, but cgo is not enabled"

**Issue:** iOS arm64 and Android amd64 require CGO.

**Solution:** 
- For iOS simulator (amd64): Use `CGO_ENABLED=0` with `-buildmode=exe`
- For iOS device (arm64): Use ebitenmobile or accept the limitation
- For Android amd64: Use ebitenmobile or accept the limitation

### 2. "build constraints exclude all Go files in internal/vibrate"

**Issue:** Ebiten's internal vibrate package has no matching files for regular mobile builds.

**Solution:** This is by design. Ebiten mobile apps must use ebitenmobile, not regular `go build`. Our stub implementations work around this for basic compilation.

### 3. "X11/Xlib.h: No such file or directory"

**Issue:** Ebiten requires graphics libraries for desktop builds in CI environments.

**Solution:** 
- Install required libraries (libx11-dev, libgl1-mesa-dev, etc.)
- Or test only packages without Ebiten dependencies
- Or use ebitenmobile for mobile platforms

## References

- [Go Build Constraints](https://pkg.go.dev/go/build#hdr-Build_Constraints)
- [Ebiten Mobile Documentation](https://ebitengine.org/en/documents/mobile.html)
- [Go WebAssembly](https://github.com/golang/go/wiki/WebAssembly)
- [iOS Build Modes](https://github.com/golang/go/blob/master/misc/ios/README)

## Summary

| Platform | Regular `go build` | ebitenmobile | Notes |
|----------|-------------------|--------------|-------|
| Android arm64 | ✅ Stub | ✅ Full | Stub for quick compilation |
| Android amd64 | ⚠️ CGO | ✅ Full | Emulator needs CGO |
| iOS amd64 | ✅ Stub | ✅ Full | Simulator builds |
| iOS arm64 | ⚠️ CGO | ✅ Full | Device needs CGO/SDK |
| WASM | ✅ Full | N/A | Browser builds work fully |

**Recommendation:** Use regular `go build` for development/testing of non-Ebiten packages. Use ebitenmobile for actual mobile application builds.
