# Cross-Platform Build Fix Documentation

## Overview
This document describes the fix for WASM build failures caused by the zenity dependency's use of platform-specific syscalls.

## Problem
The `zenity` library (used for native file dialogs) uses `syscall.Exec` which is not available in WebAssembly (WASM) or mobile platforms. This caused compilation failures when building for these targets.

### Error Message
```
../../../go/pkg/mod/github.com/ncruces/zenity@v0.10.14/internal/zenutil/run_unix.go:43:11: undefined: syscall.Exec
../../../go/pkg/mod/github.com/ncruces/zenity@v0.10.14/internal/zenutil/run_unix.go:63:11: undefined: syscall.Exec
```

## Solution
Implemented platform-specific code using Go build tags to provide different implementations based on the target platform.

### File Structure
1. **pkg/engine/character_creation.go** - Common code (all platforms)
   - Contains all shared types, structs, and functions
   - Declares `OpenPortraitDialog()` signature without implementation
   - Works across desktop, mobile, and WASM

2. **pkg/engine/character_creation_desktop.go** - Desktop implementation
   - Build tag: `//go:build !js && !android && !ios`
   - Imports and uses zenity for native file dialogs
   - Provides full file picker functionality on Linux, macOS, Windows

3. **pkg/engine/character_creation_mobile.go** - Mobile/WASM stub
   - Build tag: `//go:build js || android || ios`
   - Returns "not supported" error for file dialog requests
   - Allows compilation on WASM, Android, and iOS platforms

### Build Tag Strategy
```go
// Desktop version (Linux, macOS, Windows)
//go:build !js && !android && !ios
// +build !js,!android,!ios

// Mobile/WASM version
//go:build js || android || ios
// +build js android ios
```

## Verification

### WASM Build (Primary Fix)
```bash
GOOS=js GOARCH=wasm go build ./cmd/client
# ✅ SUCCESS
```

### Desktop Builds
```bash
# Windows (cross-compile from Linux)
GOOS=windows GOARCH=amd64 go build ./cmd/client
# ✅ SUCCESS

# Linux (requires X11 libraries installed)
go build ./cmd/client
# ✅ SUCCESS (when graphics libraries present)
```

### Mobile Builds
Mobile builds require the `ebitenmobile` tool and proper SDK/NDK:
```bash
# Android
ebitenmobile bind -target android -o mobile.aar ./cmd/mobile

# iOS
ebitenmobile bind -target ios -o Mobile.xcframework ./cmd/mobile
```

## API Compatibility
The public API remains unchanged:
- `OpenPortraitDialog()` has the same signature on all platforms
- Desktop users get full functionality
- Mobile/WASM users get a clear error message

### Example Usage
```go
// Works on all platforms
filename, err := engine.OpenPortraitDialog()
if err != nil {
    // On desktop: user cancelled or error occurred
    // On mobile/WASM: "file dialogs are not supported on mobile/WASM platforms"
    log.Printf("File dialog error: %v", err)
}
```

## Platform-Specific Behavior

### Desktop (Linux, macOS, Windows)
- ✅ Full native file dialog support via zenity
- ✅ PNG file filter
- ✅ Default to Pictures directory
- ✅ User cancellation handling

### WASM (Browser)
- ❌ File dialogs not supported (returns error)
- ✅ Manual path entry still works
- ℹ️ Future: Could use HTML5 file input instead

### Mobile (Android, iOS)
- ❌ File dialogs not supported (returns error)
- ℹ️ Future: Should use native mobile file pickers
- ℹ️ iOS: `UIDocumentPickerViewController`
- ℹ️ Android: `Intent.ACTION_OPEN_DOCUMENT`

## Build Requirements

### Desktop
- Go 1.24.5+
- Graphics libraries:
  - Linux: X11 development libraries (`libx11-dev`, `libgl1-mesa-dev`, etc.)
  - macOS: Xcode Command Line Tools
  - Windows: No additional requirements

### WASM
- Go 1.24.5+
- No additional dependencies

### Mobile
- Go 1.24.5+
- `ebitenmobile` tool: `go install github.com/hajimehoshi/ebiten/v2/cmd/ebitenmobile@latest`
- Android: Android SDK, Android NDK
- iOS: Xcode (macOS only)

## Testing

### Unit Tests
```bash
# Test packages without Ebiten dependencies
go test ./pkg/procgen/...
# ✅ All tests pass

# Test packages with Ebiten (requires graphics environment)
go test ./pkg/engine/...
# ⚠️ Requires X11/graphics libraries
```

### Build Tests
```bash
# Verify WASM builds
GOOS=js GOARCH=wasm go build ./cmd/client

# Verify desktop builds
GOOS=windows GOARCH=amd64 go build ./cmd/client
GOOS=linux GOARCH=amd64 go build ./cmd/client
GOOS=darwin GOARCH=arm64 go build ./cmd/client
```

## Future Improvements

### WASM
- Implement HTML5 file input as alternative
- Use JavaScript file APIs via syscall/js

### Mobile
- Implement native file pickers:
  - Android: `android.content.Intent.ACTION_OPEN_DOCUMENT`
  - iOS: `UIDocumentPickerViewController`
- Add build-tag-specific implementations
- Integrate with ebitenmobile framework

### Cross-Platform
- Consider abstracting file picker behind interface
- Support multiple file selection
- Add file type filtering for mobile

## Related Files
- `pkg/engine/character_creation.go` - Common implementation
- `pkg/engine/character_creation_desktop.go` - Desktop-specific (zenity)
- `pkg/engine/character_creation_mobile.go` - Mobile/WASM stub
- `scripts/build-android.sh` - Android build script
- `scripts/build-ios.sh` - iOS build script
- `Makefile` - Build targets

## References
- [Go Build Tags Documentation](https://pkg.go.dev/go/build#hdr-Build_Constraints)
- [Ebiten Mobile Documentation](https://ebitengine.org/en/documents/mobile.html)
- [WebAssembly Support](https://github.com/golang/go/wiki/WebAssembly)
- [zenity Library](https://github.com/ncruces/zenity)
