# VR Adapter Architecture

## Overview

The Venture VR subsystem uses an adapter pattern to abstract VR hardware interfaces, enabling the game to support VR features with graceful degradation when hardware is unavailable.

## Current Implementations

### Production Stub Adapters

**Location:** `pkg/engine/vr_stub_adapters.go`

Production-ready stub implementations that report "no hardware detected" status:

- **StubHeadsetAdapter**: Implements `VRHeadsetAdapter`
  - Returns `IsConnected() = false`
  - Returns zero orientation/position values
  - Returns standard IPD of 63mm for stereoscopic rendering
  - Enables HeadTrackingSystem to fall back to mouse input

- **StubControllerAdapter**: Implements `VRControllerAdapter`
  - Returns `IsConnected(hand) = false` for all hands
  - Returns zero input values (trigger, grip, thumbstick)
  - Returns false for all button states
  - No-op haptic feedback (no hardware to vibrate)
  - Enables VRControllerSystem to degrade gracefully to keyboard/mouse

### Test Mock Adapters

**Location:** `pkg/engine/head_tracking_system.go`, `pkg/engine/vr_controller_system.go`

Configurable mock implementations for testing VR system behavior:

- **MockHeadset**: Test implementation with setter methods for orientation/position
- **MockController**: Test implementation with setter methods for all inputs

## Usage

### Client Initialization

The client uses stub adapters in production when VR runtime is detected but no hardware SDK is available:

```go
// From cmd/client/handlers.go

// Head tracking with stub adapter
headSystem := engine.NewHeadTrackingSystem(world)
stubHeadset := engine.NewStubHeadsetAdapter()
headSystem.SetHeadsetAdapter(stubHeadset)

// VR controllers with stub adapter
ctrlSystem := engine.NewVRControllerSystem(world)
stubController := engine.NewStubControllerAdapter()
ctrlSystem.SetControllerAdapter(stubController)
```

### Testing

Tests use mock adapters to simulate VR hardware behavior:

```go
// Create mock with configurable values
mock := engine.NewMockHeadset()
mock.SetHeadOrientation(0.5, 1.0, 0.2)
mock.SetHeadPosition(1.0, 2.0, 3.0)

system := engine.NewHeadTrackingSystem(world)
system.SetHeadsetAdapter(mock)
```

## Graceful Degradation

The VR systems gracefully degrade when stub adapters report no hardware:

1. **HeadTrackingSystem**: Falls back to mouse input for camera control
2. **VRControllerSystem**: System remains inactive; keyboard/mouse handles input
3. **StereoscopicSystem**: Continues rendering in dual-eye mode using standard IPD
4. **VRUISystem**: Adapts UI layout but uses standard 2D rendering

## Future Hardware Integration

### Planned: OpenVR/OpenXR SDK Integration

To integrate real VR hardware, create SDK-backed adapters:

```go
// Example future implementation
type OpenVRHeadsetAdapter struct {
    session *openvr.Session
}

func NewOpenVRHeadsetAdapter() (*OpenVRHeadsetAdapter, error) {
    session, err := openvr.Init()
    if err != nil {
        return nil, err
    }
    return &OpenVRHeadsetAdapter{session: session}, nil
}

func (a *OpenVRHeadsetAdapter) IsConnected() bool {
    return a.session.IsHMDPresent()
}

func (a *OpenVRHeadsetAdapter) GetHeadOrientation() (pitch, yaw, roll float64) {
    pose := a.session.GetDevicePose(openvr.HMD)
    return pose.Pitch(), pose.Yaw(), pose.Roll()
}
```

Then update client initialization to prefer hardware adapters:

```go
// Try hardware SDK first
vrAdapter, err := engine.NewOpenVRHeadsetAdapter()
if err != nil {
    // Fall back to stub adapter
    vrAdapter = engine.NewStubHeadsetAdapter()
}
headSystem.SetHeadsetAdapter(vrAdapter)
```

## Design Benefits

1. **Zero Hardware Dependencies**: Game builds and runs without VR SDK installations
2. **Testability**: Mock adapters enable comprehensive VR system testing
3. **Extensibility**: New hardware SDKs can be added without changing VR systems
4. **Progressive Enhancement**: VR features activate only when hardware is available
5. **Clean Separation**: VR system logic is independent of hardware implementation

## Status

**Current:** Experimental (stub adapters only)  
**Planned:** OpenVR/OpenXR integration in future releases  
**Maintained:** Mock adapters for testing

See `README.md` section "VR Mode (Experimental)" for user-facing documentation.
