# Implementation Gaps — 2026-04-23

## OpenXR Adapter Logic Is Stubbed (Desktop VR)
- **Intended Behavior**: In `-tags vr` builds, desktop VR adapters should bind to OpenXR runtime/session and provide real headset/controller state.
- **Current State**: `pkg/engine/vr_openxr_adapters.go` contains TODO-backed placeholders and default-return methods (`0/false/no-op`) at lines 94, 117, 126, 147, 160, 176, 183, 191, 199, 207, 215, 224.
- **Blocked Goal**: ROADMAP Priority 4 items requiring real head tracking/controller input and real-hardware validation (`ROADMAP.md:157-162`).
- **Implementation Path**:
  1. Implement OpenXR runtime initialization in `NewOpenXRHeadsetAdapter` (`xrCreateInstance`, `xrGetSystem`, `xrCreateSession`).
  2. Implement pose reads (`xrLocateViews`) in `GetHeadOrientation`/`GetHeadPosition`.
  3. Implement controller action setup and state polling (`xrSyncActions`, `xrGetActionState*`) in controller adapter methods.
  4. Implement haptics via `xrApplyHapticFeedback`.
  5. Add build-tagged integration tests for successful runtime session and non-zero state reads when hardware/runtime is available.
- **Dependencies**: OpenXR SDK availability in CI/dev environments and cgo linker configuration.
- **Effort**: large

## OpenXR Runtime Path Never Selected at Runtime
- **Intended Behavior**: `NewRuntimeHeadsetAdapter` / `NewRuntimeControllerAdapter` should choose OpenXR adapters when a valid VR runtime/hardware session is available.
- **Current State**: Factory code (`pkg/engine/vr_adapter_factory_openxr.go:8-12`, `18-22`) gates on `IsConnected*`; OpenXR adapters never set `connected` true due missing integration (`pkg/engine/vr_openxr_adapters.go:82,112,177`).
- **Blocked Goal**: Real VR runtime route cannot activate, so VR remains stub-only even in vr-tagged builds.
- **Implementation Path**:
  1. Set `connected` true only after successful runtime/session/action initialization.
  2. Add explicit adapter-state tests proving both branches: fallback-to-stub and connected-OpenXR.
  3. Ensure error logging distinguishes “SDK unavailable” vs “runtime unavailable” vs “hardware unavailable”.
- **Dependencies**: Completion of OpenXR adapter integration gap above.
- **Effort**: medium

## LightingConfig.EnableShadows Is a No-Op API Field
- **Intended Behavior**: Public config fields should either affect behavior or be removed/deprecated with migration path.
- **Current State**: `pkg/rendering/lighting/types.go:116-122` explicitly states `EnableShadows` has no implementation; no operational usage in lighting runtime path.
- **Blocked Goal**: Misleading API and configuration contract for lighting behavior.
- **Implementation Path**:
  1. Decide ownership: rendering-lighting shadow toggle vs engine-level shadow system toggle.
  2. If retained, wire field into active render/update logic and add tests for enabled/disabled behavior.
  3. If removed, deprecate and migrate callers to canonical shadow control path with clear release note.
- **Dependencies**: Alignment between `pkg/rendering/lighting` and `pkg/engine` shadow systems.
- **Effort**: medium

## WebXR Adapter File Is Documented but Missing
- **Intended Behavior**: WASM VR path should have a js-build adapter implementation per package documentation.
- **Current State**: `pkg/vr/doc.go:79,84` references future `pkg/engine/vr_webxr_adapters.go`, but no such file exists.
- **Blocked Goal**: Documented WASM VR integration path is not yet implementable.
- **Implementation Path**:
  1. Add `pkg/engine/vr_webxr_adapters.go` with `//go:build js` constraints.
  2. Implement `VRHeadsetAdapter` and `VRControllerAdapter` via `syscall/js` WebXR APIs.
  3. Add compile-time interface assertions and js-target smoke tests.
- **Dependencies**: Browser WebXR support and js-target test strategy.
- **Effort**: medium
