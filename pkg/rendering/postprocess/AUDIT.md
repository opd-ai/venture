# Audit: pkg/rendering/postprocess

**Date**: 2026-02-16
**Auditor**: Copilot
**Coverage**: 85.4% (headless, excluding GPU processor)

## Summary

Post-processing effects package providing CPU-based and GPU-accelerated screen-space effects (motion blur, depth blur, color grading, vignette, chromatic aberration) with genre presets.

## Issues Found: 3 (0 high, 1 med, 2 low)

### Fixed

- [x] <severity:med> **Division by zero in motion blur** — `motion_blur.go` line 48: when `Samples == 1`, `float64(samples-1) == 0` caused division by zero. Guard changed from `Samples < 1` to `Samples < 2` to match chromatic aberration pattern.

- [x] <severity:low> **GPU tests crash headless environments** — `gpu_processor.go` imports `ebiten/v2` which triggers GLFW init, crashing all tests in headless/CI. Added `//go:build !headless` to `gpu_processor.go` and `gpu_processor_test.go`, with `gpu_processor_headless.go` stub for headless builds. Run `go test -tags headless` for CI without display.

- [x] <severity:low> **Missing config validation** — `ValidationError` type existed but was unused. Added `Validate()` methods on `Config`, `MotionBlurConfig`, `DepthBlurConfig`, and `ChromaticAberrationConfig` to check value ranges. Added table-driven tests.

### Noted (not fixed)

- <severity:low> **doc.go example used private field** — Example code referenced `processor.config.ColorGrading` (private field). Fixed to use `GetConfig()`/`SetConfig()` public API.

## Test Results

```
go test -tags headless ./pkg/rendering/postprocess/... -cover
PASS coverage: 85.4% of statements
```

## Files Modified

- `gpu_processor.go` — Added `//go:build !headless`
- `gpu_processor_test.go` — Added `//go:build !headless`
- `gpu_processor_headless.go` — New stub for headless builds
- `motion_blur.go` — Fixed Samples guard from `< 1` to `< 2`
- `types.go` — Added `Validate()` methods
- `types_test.go` — Added `TestConfig_Validate` table-driven tests
- `doc.go` — Fixed example to use public API
