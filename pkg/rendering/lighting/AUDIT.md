# Audit: pkg/rendering/lighting
**Date**: 2026-02-12
**Status**: Needs Work

## Summary
The lighting package provides dynamic lighting effects (point, ambient, directional lights) with post-processing (bloom, ambient occlusion) for rendered scenes. The package has comprehensive test coverage (74 test functions, 2177 test lines) and clean separation from ECS (components live in pkg/engine). The code is well-structured with deterministic algorithms and proper validation, but lacks structured logging for error paths and has one swallowed shader compilation error.

## Issues Found
- [x] **high** Error handling — Shader compilation error swallowed silently without logging (`gpu_bloom.go:261-267`) - Fixed 2026-02-13: Added structured logging with logrus.WithFields for shader compilation errors
- [x] **med** Error handling — No structured logging (logrus) on any error paths; package has zero logging statements — **FIXED 2026-02-13**: Added logrus logger to System struct, NewSystemWithLogger constructor, and logrus.WithFields logging on all error paths (AddLight, RemoveLight, GetLight, UpdateLight)
- [x] **low** Documentation — Comment states system "not yet integrated into main game loop" but LightingAdapter exists in engine and system is actively used (`system.go:11-12`) — **FIXED 2026-02-13**: Updated comment to reference LightingAdapter integration
- [ ] **low** Test coverage — Tests cannot run in headless CI due to Ebiten initialization (gpu_bloom.go imports ebiten which requires display)

## Test Coverage
Unable to measure (Ebiten requires display; tests panic with "GLFW library is not initialized")

**Estimated coverage**: 85-90% based on test file analysis:
- 6 implementation files (1,564 LOC excluding tests)
- 4 test files with 74 test functions (2,177 LOC)
- Comprehensive table-driven tests for all public APIs
- Tests cover: type conversions, validation, lighting calculations, bloom extraction, AO sampling, edge cases

## Integration Status
**Well integrated:**
- Used by `pkg/engine/lighting_system.go` via `pkg/engine/lighting_adapter.go` wrapper
- Visual test integration in `pkg/visualtest/regression.go` and `pkg/visualtest/benchmark.go`
- Blank import in `cmd/client/handlers.go` (side-effects registration)
- Properly separated: no ECS components in this package (LightComponent lives in pkg/engine)

**Integration points verified:**
- ✅ LightingAdapter wraps lighting.System for ECS (engine/lighting_adapter.go:11-31)
- ✅ Component separation: LightComponent in engine, not rendering package
- ✅ GPU bloom shaders self-contained with thread-safe lazy initialization
- ⚠️ LightingSystem not registered in system_init.go (StatusEffectLightingSystem is, but not base LightingSystem)

## Recommendations
1. **HIGH PRIORITY**: Add structured logging with logrus.WithFields for shader compilation errors (gpu_bloom.go:261) ✅ **DONE**
2. **MEDIUM PRIORITY**: Add logging to error paths in System methods (AddLight, UpdateLight validation failures) ✅ **DONE**
3. **LOW PRIORITY**: Update system.go:11-12 comment to reflect current integration status via LightingAdapter ✅ **DONE**
4. **LOW PRIORITY**: Consider build tags or interface abstraction to enable headless testing (exclude gpu_bloom.go in test-only builds)
