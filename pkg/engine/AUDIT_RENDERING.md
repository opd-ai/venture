# Engine Rendering Systems Sub-Audit

**Date**: 2026-02-16
**Scope**: Core rendering systems in `pkg/engine/` — render, camera, lighting, shadow, particle, post-processing
**Files Audited**: 13 source files (~5,300 lines)

## Summary

| Severity | Found | Fixed |
|----------|-------|-------|
| High     | 2     | 2     |
| Medium   | 1     | 1     |
| Low      | 2     | 0     |

## Issues Found & Fixed

### HIGH-001: GPUBloom headless stub missing ApplyToBuffer (FIXED)
- **File**: `pkg/rendering/lighting/gpu_bloom_headless.go`
- **Impact**: Build failure with `-tags headless` — prevented all headless testing of `pkg/engine/`
- **Root Cause**: `ApplyToBuffer(*ebiten.Image)` method added to `gpu_bloom.go` but not mirrored in headless stub
- **Fix**: Added `ApplyToBuffer` no-op stub with `ebiten.Image` import

### HIGH-002: GPUProcessor headless stub missing ApplyAll (FIXED)
- **File**: `pkg/rendering/postprocess/gpu_processor_headless.go`
- **Impact**: Build failure with `-tags headless` — prevented all headless testing of `pkg/engine/`
- **Root Cause**: `ApplyAll(*ebiten.Image) *ebiten.Image` method exists on non-headless `GPUProcessor` but not in headless stub
- **Fix**: Added `ApplyAll` stub that returns input unchanged, with `ebiten.Image` import

### MED-001: Silent particle generation failure (FIXED)
- **File**: `pkg/engine/particle_system.go`, line 80
- **Impact**: Particle generation errors silently dropped, making debugging difficult
- **Fix**: Added `logrus.Warn` logging with error details on generation failure

### LOW-001: Shadow image cache unbounded (NOT FIXED)
- **File**: `pkg/engine/shadow_system.go`, `imageCache` field
- **Impact**: Cache grows with unique (width, height) combinations; no eviction policy
- **Risk**: Very low in practice — shadow sizes are limited by entity types
- **Recommendation**: Monitor; add LRU eviction if memory profiling shows growth

### LOW-002: Lighting logger level check pattern
- **File**: `pkg/engine/lighting_system.go`, multiple locations
- **Impact**: `s.logger.Logger.GetLevel()` accesses inner Logger field directly
- **Risk**: Very low — logrus always populates Logger field when creating Entry via `WithField()`
- **Recommendation**: No action needed; pattern is consistent across codebase

## Files Audited

| File | Lines | Test File | Status |
|------|-------|-----------|--------|
| `render_system.go` | 1,300 | ✅ | Clean |
| `camera_system.go` | 652 | ✅ | Clean |
| `camera_component.go` | 234 | ✅ | Clean |
| `lighting_system.go` | 1,253 | ✅ | LOW-002 (acceptable) |
| `lighting_components.go` | 584 | ✅ | Clean |
| `lighting_adapter.go` | 175 | ✅ | Clean |
| `shadow_system.go` | 517 | ✅ | LOW-001 (deferred) |
| `shadow_components.go` | 164 | ✅ | Clean |
| `post_processor.go` | 151 | ✅ | Clean |
| `particle_system.go` | 299 | ✅ | MED-001 fixed |
| `particle_components.go` | 107 | ✅ | Clean |
| `render_drop_shadow.go` | 185 | ✅ | Clean |
| `rendering_optimization_adapters.go` | 52 | ✅ | Clean |

## Verification

- `go build ./...` — PASS
- `go build -tags headless ./...` — PASS (was FAIL before fixes)
- Rendering tests via xvfb — all PASS
