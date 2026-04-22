//go:build vr && !js

package engine

// NewRuntimeHeadsetAdapter returns an OpenXR-backed adapter for vr builds,
// falling back to the stub adapter if hardware/runtime is unavailable.
func NewRuntimeHeadsetAdapter() VRHeadsetAdapter {
	openxr := NewOpenXRHeadsetAdapter()
	if openxr.IsConnected() {
		return openxr
	}
	return NewStubHeadsetAdapter()
}

// NewRuntimeControllerAdapter returns an OpenXR-backed controller adapter for
// vr builds, falling back to the stub adapter if hardware/runtime is unavailable.
func NewRuntimeControllerAdapter() VRControllerAdapter {
	openxr := NewOpenXRControllerAdapter()
	if openxr.IsConnected("left") || openxr.IsConnected("right") {
		return openxr
	}
	return NewStubControllerAdapter()
}
