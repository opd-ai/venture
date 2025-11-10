# WASM Build Crash Fix Summary

## Problem
The WebAssembly (WASM) build crashed after character creation with the error:
```
panic: runtime error: invalid memory address or nil pointer dereference
```

This occurred because `ebiten.NewImage()` was being called during initialization, which fails in WASM environments where the graphics context is not available until the first Update/Draw cycle.

## Root Cause
In WebAssembly builds, the Ebiten graphics context is initialized asynchronously. Any calls to `ebiten.NewImage()` during initialization (before the first Update/Draw cycle) will fail with a nil pointer dereference or similar error.

Two locations were calling `ebiten.NewImage()` during initialization:

1. **`pkg/engine/game.go` line 152**: Scene buffer for lighting system
2. **`cmd/client/main.go` line 1661**: Player sprite image

## Solution
Implemented lazy initialization for both image allocations:

### 1. Scene Buffer (pkg/engine/game.go)
**Before:**
```go
// Create reusable scene buffer for lighting post-processing
// Allocated once to avoid per-frame allocations (60+ FPS)
sceneBuffer := ebiten.NewImage(screenWidth, screenHeight)
```

**After:**
```go
// WASM FIX: Scene buffer will be lazily initialized on first Draw() call
// Cannot call ebiten.NewImage() during initialization in WASM builds
// The graphics context is not available until the first Update/Draw cycle
var sceneBuffer *ebiten.Image = nil
```

Then in the Draw() method, added lazy initialization:
```go
// WASM FIX: Lazy initialization of scene buffer on first use
// In WASM, ebiten.NewImage() can only be called after graphics context is ready
if g.sceneBuffer == nil {
    g.sceneBuffer = ebiten.NewImage(g.ScreenWidth, g.ScreenHeight)
}
```

### 2. Player Sprite Image (cmd/client/main.go)
**Before:**
```go
playerSprite := &engine.EbitenSprite{
    Image:   ebiten.NewImage(28, 28), // Initial image (will be replaced by animation)
    Width:   28,
    Height:  28,
    Visible: true,
    Layer:   10,
}
```

**After:**
```go
// WASM FIX: Set Image to nil initially; animation system will create it on first update
// In WASM, ebiten.NewImage() can only be called after graphics context is ready
playerSprite := &engine.EbitenSprite{
    Image:   nil, // Will be created by animation system on first update
    Width:   28,
    Height:  28,
    Visible: true,
    Layer:   10,
}
```

The render system already has a nil-check that draws a colored rectangle as a fallback when the image is nil, so this change is safe and seamless.

## Testing

### Build Verification
- ✅ Desktop build compiles successfully: `make build-client`
- ✅ WASM build compiles successfully: `make build-wasm`
- ✅ Non-graphical tests pass: `go test ./pkg/procgen/... ./pkg/combat/... ./pkg/world/...`

### Manual Testing Required
To fully verify the fix, manual testing is needed:

1. Build and serve WASM version:
   ```bash
   make serve-wasm
   ```

2. Open http://localhost:8080 in a web browser

3. Verify the following flow:
   - Main menu loads without crash
   - Enter character creation
   - Complete all steps (name, class, portrait, confirmation)
   - Game transitions to gameplay screen without crash
   - Player sprite renders correctly
   - Lighting effects work if enabled

### Verification Checklist
- [ ] WASM loads without crash
- [ ] Character creation UI renders properly
- [ ] Character creation completes successfully
- [ ] Game transitions to gameplay after character creation
- [ ] Player sprite renders (or shows colored rectangle until animated)
- [ ] Lighting system works (if enabled)
- [ ] No console errors in browser developer tools

## Related Files Changed
- `pkg/engine/game.go`: Scene buffer lazy initialization
- `cmd/client/main.go`: Player sprite lazy initialization, removed unused import

## Additional Notes
- All other `ebiten.NewImage()` calls in the codebase are in Draw() methods or similar functions that execute after the graphics context is initialized, so they are safe for WASM.
- The fix follows the lazy initialization pattern which is a standard approach for WASM compatibility with Ebiten.
- No functionality changes for desktop builds - they behave identically to before.

## Security Review
✅ CodeQL security scan passed with 0 alerts

## References
- Ebiten WASM documentation: https://ebitengine.org/en/documents/webassembly.html
- Repository custom instructions specify WASM compatibility requirements
- Similar patterns used throughout the codebase for UI components that also defer image creation to Draw() methods
