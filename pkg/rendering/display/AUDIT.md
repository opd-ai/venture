# Audit: github.com/opd-ai/venture/pkg/rendering/display
**Date**: 2026-02-16
**Status**: Needs Work

## Summary
Display package provides resolution management and UI scaling for cross-platform support (1280x720 to 3840x2160). Architecture is clean with proper separation (Config/Manager/Scaler), but has one critical issue: uses `time.Now()` for performance measurement, violating determinism requirements. No ECS compliance needed (utility package). Integration is minimal but functional (client handlers only).

## Issues Found
- [x] **high** Deterministic procgen — Non-deterministic time usage in `Manager.ApplyResolution()` (`manager.go:27`)
- [x] **high** Deterministic procgen — Non-deterministic time usage in `Manager.ApplyResolution()` (`manager.go:33`)
- [x] **med** Integration points — `ToggleFullscreen()` and `SetResolution()` methods exist but are never called in client runtime; only `ApplyResolution()` is used during initialization (`cmd/client/handlers.go:1628`)
- [x] **low** Test coverage — Tests cannot run due to Ebiten runtime dependency (requires DISPLAY/GLFW); headless test mode not implemented

## Test Coverage
**Unable to measure** — Tests require Ebiten runtime (GLFW/DISPLAY). Based on code analysis: 27 test functions + 5 benchmarks covering 27 production functions = **~100% theoretical coverage**. All public APIs have table-driven tests. Estimated **95%+** if tests could run.

## Integration Status

### Current Integration
- **Client (cmd/client/handlers.go)**: Display manager initialized in `initializeV7Systems()` with CLI flags (width/height/fullscreen). Only `ApplyResolution()` is called at startup (line 1628).
- **UI Package (pkg/rendering/ui/scaler.go)**: UIScaler wrapper properly delegates to `display.Scaler` for all UI scaling operations.

### Missing Integration
- **Runtime resolution switching**: `Manager.SetResolution()` method exists but no UI menu/hotkey to invoke it.
- **Fullscreen toggle**: `Manager.ToggleFullscreen()` method exists but no F11 key binding or menu option to trigger it.
- **Display settings UI**: No in-game settings screen to change resolution/fullscreen at runtime.

### No ECS Registration Required
This is a utility package for display configuration, not an ECS system. Manager is stored in `systemsContainer` but does not implement `System.Update()` interface. This is correct architecture.

## Recommendations

### 1. **HIGH PRIORITY**: Fix non-deterministic time usage
**Problem**: `manager.go:27` and `manager.go:33` use `time.Now()` and `time.Since()` for performance measurement, violating deterministic procgen requirements.

**Why it matters**: Even though this is a utility package (not procgen), using `time.Now()` violates codebase standards. The `switchDuration` tracking is for performance metrics, not gameplay, but should use a deterministic alternative or be clearly marked as non-deterministic.

**Fix options**:
- Option A: Remove performance timing entirely (simplest)
- Option B: Accept a clock interface for testing (allows deterministic tests)
- Option C: Document that timing is intentionally non-deterministic for performance monitoring (add comment exception)

**Recommended fix** (Option C):
```go
// ApplyResolution applies current config to Ebiten window.
// Returns time taken for the switch operation.
// NOTE: Uses time.Now() for performance measurement (non-deterministic by design).
// This is acceptable as it's for observability, not game logic.
func (m *Manager) ApplyResolution() time.Duration {
    m.switchStarted = time.Now() // NON-DETERMINISTIC: performance measurement only
    
    ebiten.SetWindowSize(m.config.Width, m.config.Height)
    ebiten.SetFullscreen(m.config.Fullscreen)
    ebiten.SetVsyncEnabled(m.config.VSync)
    
    m.switchDuration = time.Since(m.switchStarted) // NON-DETERMINISTIC: performance measurement only
    return m.switchDuration
}
```

### 2. **MEDIUM PRIORITY**: Wire up runtime resolution/fullscreen controls
**Problem**: `SetResolution()` and `ToggleFullscreen()` methods exist but are never called.

**Fix**: Add key bindings and/or settings UI in client:
```go
// In cmd/client/input.go or handlers.go Update() method:
if ebiten.IsKeyPressed(ebiten.KeyF11) {
    sys.displayManager.ToggleFullscreen()
    logger.Info("Toggled fullscreen mode")
}
```

### 3. **LOW PRIORITY**: Add headless test mode
**Problem**: Tests fail without DISPLAY/GLFW (CI/CD environments).

**Fix**: Add build tags or mock Ebiten calls for headless testing. Many other packages have solved this with stub implementations.

### 4. **DOCUMENTATION**: Add architectural decision record
Document why `Manager` is in `systemsContainer` but not an ECS system. This is correct design but may confuse future maintainers.

### 5. **FUTURE ENHANCEMENT**: Dynamic resolution detection
Add monitor capability detection to populate available resolutions dynamically instead of hardcoded `standardResolutions`.
