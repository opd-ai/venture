//go:build vr && js

package engine

// NewRuntimeHeadsetAdapter returns a WebXR-backed adapter for WASM+vr builds,
// falling back to the stub adapter if the browser has no WebXR support or the
// user denies the immersive-vr session request.
func NewRuntimeHeadsetAdapter() VRHeadsetAdapter {
	webxr := NewWebXRHeadsetAdapter()
	if webxr.IsConnected() {
		return webxr
	}
	return NewStubHeadsetAdapter()
}

// NewRuntimeControllerAdapter returns a WebXR-backed controller adapter for
// WASM+vr builds, sharing the session from the headset adapter.
// Falls back to the stub adapter when no WebXR session is active.
func NewRuntimeControllerAdapter() VRControllerAdapter {
	webxr := NewWebXRHeadsetAdapter()
	ctrl := NewWebXRControllerAdapter(webxr)
	if ctrl.IsConnected("left") || ctrl.IsConnected("right") {
		return ctrl
	}
	return NewStubControllerAdapter()
}
