# Build Tag Migration Guide

## Overview

This document describes the migration from build tag-based test exclusion to a cleaner architecture where build tags are only used for platform-specific code.

## What Changed

### Before

The codebase used build tags in two ways:
1. `//go:build test` on example programs (only compile in test mode)
2. `//go:build !test` on rendering code (exclude from test builds to avoid X11 dependency)

This created a bifurcation where:
- Running tests without tags failed due to missing X11
- Running tests with `-tags test` failed due to missing implementations

### After

Build tags are now **only used for platform-specific code**:
- `//go:build ios` for iOS-specific implementations
- `//go:build android` for Android-specific implementations

All other code compiles without build tags.

## Rationale

### Why Remove Build Tags?

1. **Simplified Build Process**: No need to remember `-tags test` flag
2. **Cleaner Architecture**: Data types always available
3. **Better IDE Support**: No conditional compilation confusion
4. **Standard Go Practices**: Build tags reserved for platform/architecture differences

### Graphics Dependency Limitation

Some tests require graphics context (X11/Wayland on Linux, display on other platforms):
- Shape generation tests
- Sprite generation tests  
- Animation tests

**This is expected and acceptable.** Per project guidelines:
> "Target minimum 65% code coverage per package, excluding functions that require Ebiten runtime initialization"

The tiles package demonstrates the ideal pattern: it uses `*image.RGBA` instead of `*ebiten.Image`, avoiding the graphics dependency.

## Migration Impact

### For Contributors

**Running Tests:**
```bash
# Before
go test -tags test ./...

# After  
go test ./...
```

**Building Examples:**
```bash
# Before
go build -tags test ./examples/audio_demo

# After
go build ./examples/audio_demo
```

### For CI/CD

Tests that require graphics will fail in headless environments. This is expected:

```bash
# These will fail without display:
go test ./pkg/rendering/shapes/...
go test ./pkg/rendering/sprites/...

# These will pass:
go test ./pkg/rendering/tiles/...
go test ./pkg/procgen/...
go test ./pkg/audio/...
```

**Solution**: Run graphics tests only on platforms with display available, or accept partial test coverage in CI.

## Files Modified

### Data Type Files (No Build Tags)
- `pkg/rendering/sprites/anatomy_template.go` - Body part types, templates
- `pkg/rendering/sprites/item_template.go` - Item types, templates
- `pkg/rendering/patterns/types.go` - Pattern types
- `pkg/rendering/shapes/types.go` - Shape types (already no tags)
- `pkg/rendering/sprites/types.go` - Sprite types (already no tags)

### Implementation Files (Build Tags Removed)
- All files in `pkg/rendering/shapes/`
- All files in `pkg/rendering/sprites/`
- `pkg/rendering/interfaces.go`
- `pkg/engine/equipment_visual_system.go`
- `pkg/engine/animation_system_test.go`

### Example Programs (Build Tags Removed)
- `examples/audio_demo/main.go`
- `examples/combat_demo/main.go`
- `examples/genre_blending_demo/main.go`
- `examples/lag_compensation_demo/main.go`
- `examples/movement_collision_demo/main.go`
- `examples/multiplayer_demo/main.go`
- `examples/network_demo/main.go`
- `examples/prediction_demo/main.go`

### Platform-Specific Files (Build Tags Retained)
- `pkg/mobile/platform_ios.go` - `//go:build ios`
- `pkg/mobile/platform_android.go` - `//go:build android`

## Testing Guidelines

### Tests That Should Pass Everywhere

Tests of pure data structures and logic:
```go
func TestBodyPart_String(t *testing.T) { /* ... */ }
func TestDefaultConfig(t *testing.T) { /* ... */ }
func TestItemRarity_String(t *testing.T) { /* ... */ }
```

These tests have no graphics dependency and run in CI.

### Tests That Require Graphics

Tests calling generator functions:
```go
func TestGenerateShape(t *testing.T) {
    gen := NewGenerator()
    img, err := gen.Generate(config) // Requires graphics context
    // ...
}
```

These tests will fail in headless CI. This is documented and acceptable.

## Best Practices Going Forward

### When to Use Build Tags

**DO use build tags for:**
- Platform-specific code (iOS, Android, Windows, Linux, macOS)
- Architecture-specific code (amd64, arm64)
- Feature flags for optional dependencies

**DON'T use build tags for:**
- Test vs. production builds
- Excluding code that's "hard to test"
- Working around CI limitations

### How to Write Testable Code

**Prefer data structures without graphics dependencies:**
```go
// Good - Returns standard library type
func (g *Generator) GenerateData(config Config) (*image.RGBA, error)

// Requires graphics - Use sparingly
func (g *Generator) GenerateSprite(config Config) (*ebiten.Image, error)
```

**Separate data types from implementation:**
```go
// types.go - No build tags, always available
type BodyPart int
const (
    PartHead BodyPart = iota
    PartBody
)

// generator.go - May use graphics, tests may fail in CI
func (g *Generator) Generate(config Config) (*ebiten.Image, error)
```

## Troubleshooting

### Test Failures in CI

**Error:**
```
fatal error: X11/Xlib.h: No such file or directory
```

**Solution:** This is expected for graphics tests. Either:
1. Install X11 dev libraries in CI
2. Run tests with virtual display (xvfb)
3. Accept that these tests don't run in CI

### Missing Types in Tests

**Error:**
```
undefined: BodyPart
```

**Solution:** The file defining the type likely has a build tag. Remove it (data types should always be available).

### Build Failures

**Error:**
```
undefined: NewGenerator
```

**Solution:** Build tags were removed. The function should now always be available. If still failing, check:
1. Import paths are correct
2. No cached build artifacts (`go clean -cache`)

## Future Improvements

### Ideal Architecture

The tiles package demonstrates the ideal pattern:

```go
// Returns standard library image type
func (g *TileGenerator) Generate(config Config) (*image.RGBA, error) {
    // Generate using standard library
    img := image.NewRGBA(...)
    // ...
    return img, nil
}

// Caller converts when needed
tileImg, err := tileGen.Generate(config)
ebitenImg := ebiten.NewImageFromImage(tileImg)
```

This approach:
- ✅ Tests run in CI without graphics
- ✅ No build tags needed
- ✅ Clean separation of concerns
- ✅ Standard library types in API

### Migration Path

To fully remove graphics dependencies:

1. Refactor shape generator to return `*image.RGBA`
2. Refactor sprite generator to return `*image.RGBA`
3. Update callers to convert to `*ebiten.Image` when needed
4. All tests can then run without graphics context

This is a breaking API change and should be done in a major version bump.

## Questions?

See the [Development Guide](DEVELOPMENT.md) for more information about the project architecture and testing practices.
