# Sprite Visibility Bug Fixes

## Summary

Fixed two critical bugs in the sprite rendering system that caused sprites to fail to render on screen.

## Bugs Fixed

### Bug #1: Incorrect Culling When No Camera Exists

**Location**: `pkg/engine/camera_system.go:269` - `IsVisible()` method

**Problem**: 
When `activeCamera` is `nil`, the `WorldToScreen()` method returns world coordinates unchanged (as a fallback). However, `IsVisible()` then compared these world coordinates against screen bounds, causing incorrect culling.

For example:
- Entity at world position (1000, 500)
- Screen width = 800, height = 600
- `WorldToScreen(1000, 500)` returns (1000, 500) when no camera
- `IsVisible()` checks: `1000 <= 800` → FALSE
- Entity incorrectly culled even though it should be visible

**Impact**: 
- Sprites with world coordinates exceeding screen dimensions were incorrectly culled
- Affected games running without a camera or during camera initialization
- Could cause sprites to disappear unexpectedly

**Solution**:
Added null camera check in `IsVisible()` to return `true` when no active camera exists:

```go
func (s *CameraSystem) IsVisible(worldX, worldY, radius float64) bool {
    // BUG FIX: When there's no active camera, all entities are considered visible
    // because WorldToScreen returns world coordinates unchanged, which would
    // incorrectly be compared against screen bounds.
    if s.activeCamera == nil {
        return true
    }

    cameraComp, ok := s.activeCamera.GetComponent("camera")
    if !ok {
        return true
    }

    screenX, screenY := s.WorldToScreen(worldX, worldY)

    // Check if within screen bounds (with margin for radius)
    margin := radius * 2
    return screenX >= -margin && screenX <= float64(s.ScreenWidth)+margin &&
        screenY >= -margin && screenY <= float64(s.ScreenHeight)+margin
}
```

### Bug #2: Spatial Partition Using Wrong Camera Position

**Location**: `pkg/engine/render_system.go:523` - `getVisibleEntities()` method

**Problem**: 
The viewport culling system was using the camera entity's `PositionComponent` instead of the camera's actual position stored in `CameraComponent.X` and `CameraComponent.Y`.

The camera position includes:
- Smoothing effects (lerp between positions)
- Offset values
- Bounds clamping

These are stored in `camera.X` and `camera.Y`, NOT in the entity's position component.

**Example Scenario**:
```
Camera entity position component: (500, 500)
Camera actual position (after smoothing): (1000, 1000)
Entity to render: (1100, 1100)

BEFORE FIX:
- Viewport calculated around (500, 500)
- Entity at (1100, 1100) outside viewport
- Entity incorrectly culled

AFTER FIX:
- Viewport calculated around (1000, 1000)
- Entity at (1100, 1100) inside viewport
- Entity correctly rendered
```

**Impact**: 
- Viewport culling calculated wrong bounds
- Sprites visible on screen were incorrectly culled
- Affected games using camera smoothing or offsets
- Could cause flickering or missing sprites

**Solution**:
Changed `getVisibleEntities()` to use `camera.X` and `camera.Y` directly:

```go
func (r *EbitenRenderSystem) getVisibleEntities(entities []*Entity) []*Entity {
    // ... camera retrieval code ...

    // BUG FIX: Use camera's actual position (camera.X, camera.Y) which includes
    // smoothing and bounds clamping, NOT the entity's position component.
    // The camera position is updated by CameraSystem and represents where
    // the camera is actually looking, which may differ from the entity position
    // due to smoothing, offsets, and bounds constraints.

    // Calculate world viewport bounds
    viewportWidth := float64(r.cameraSystem.ScreenWidth) / camera.Zoom
    viewportHeight := float64(r.cameraSystem.ScreenHeight) / camera.Zoom

    viewportBounds := Bounds{
        X:      camera.X - viewportWidth/2 - margin,
        Y:      camera.Y - viewportHeight/2 - margin,
        Width:  viewportWidth + margin*2,
        Height: viewportHeight + margin*2,
    }

    // Query spatial partition for entities in viewport
    visible := r.spatialPartition.QueryBounds(viewportBounds)

    return visible
}
```

### Performance Optimization Re-enabled

**Location**: `pkg/engine/render_system.go:177` - `NewRenderSystem()`

**Problem**: 
Viewport culling was disabled with comment: `// TEMPORARY: Disabled culling due to spatial partition issue`

**Solution**:
Re-enabled culling now that the spatial partition bug is fixed:

```go
func NewRenderSystem(cameraSystem *CameraSystem) *EbitenRenderSystem {
    return &EbitenRenderSystem{
        cameraSystem:     cameraSystem,
        spatialPartition: nil,  // Will be set when world bounds are known
        enableCulling:    true, // Culling enabled by default (spatial partition bug fixed)
        enableBatching:   true, // Batching enabled by default
        // ...
    }
}
```

**Impact**: 
- Viewport culling optimization now works correctly
- Improved rendering performance (1,635x speedup as documented)
- Reduces number of entities processed per frame

## Testing

### Regression Tests Added

1. **`TestCameraSystem_IsVisible_NoCamera`** - Tests visibility when no camera is active
2. **`TestCameraSystem_IsVisible_WithCamera`** - Tests visibility with active camera at various positions
3. **`TestCameraSystem_IsVisible_NoComponent`** - Tests visibility when camera entity has no component
4. **`TestRenderSystem_SpatialPartition_CameraPosition`** - Tests spatial partition uses correct camera position
5. **`TestRenderSystem_EnableCulling`** - Tests culling enable/disable functionality

### Test Files Modified

- `pkg/engine/camera_component_test.go` - Added 3 new test functions (106 lines)
- `pkg/engine/render_system_test.go` - Added 2 new test functions (80 lines)

## Files Modified

1. **`pkg/engine/camera_system.go`**
   - Modified `IsVisible()` method to handle null camera case
   - Added explanatory comments
   - +12 lines

2. **`pkg/engine/render_system.go`**
   - Fixed `getVisibleEntities()` to use `camera.X/Y` instead of entity position
   - Re-enabled culling by default
   - Added explanatory comments
   - +10 lines, -11 lines

3. **`pkg/engine/camera_component_test.go`**
   - Added regression tests for IsVisible fixes
   - +106 lines

4. **`pkg/engine/render_system_test.go`**
   - Added test for spatial partition fix
   - +80 lines

**Total Changes**: 4 files, +208 lines, -11 lines

## Optimizations Preserved

All existing performance optimizations remain intact and functional:

1. **Viewport Culling** - Re-enabled, now working correctly (1,635x speedup)
2. **Batch Rendering** - Unchanged, groups entities by sprite (1,667x speedup)
3. **Sprite Caching** - Unchanged, avoids regeneration (37x speedup, 95.9% hit rate)
4. **Object Pooling** - Unchanged, reduces allocations (2x speedup)

**Combined Performance**: 1,625x overall rendering improvement maintained

## Verification

The fixes ensure:

✅ Sprites render visibly when expected  
✅ Camera initialization edge cases handled  
✅ Viewport culling uses correct camera position  
✅ Performance optimizations working as designed  
✅ Code properly formatted with `go fmt`  
✅ Comprehensive regression tests added  

## Related Issues

- Closes issue mentioned in line 177 comment: "TEMPORARY: Disabled culling due to spatial partition issue"
- Fixes potential flickering/missing sprites during camera transitions
- Resolves incorrect culling during game initialization

## Future Considerations

1. Consider adding debug visualization for camera bounds and viewport
2. Add telemetry for culling statistics (already exists via `GetStats()`)
3. Document camera smoothing behavior more thoroughly
4. Consider adding validation for camera position vs entity position mismatch
