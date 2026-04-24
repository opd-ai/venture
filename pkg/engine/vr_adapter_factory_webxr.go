//go:build vr && js

package engine

import "sync"

// webxrHeadsetSingleton ensures both factory functions share the same
// WebXRHeadsetAdapter and therefore the same XRSession and pose cache.
// A second call to xr.requestSession after a session is already active would
// either return the same session (some runtimes) or raise an error (others).
var (
	webxrHeadsetOnce     sync.Once
	webxrHeadsetInstance *WebXRHeadsetAdapter
)

func sharedWebXRHeadset() *WebXRHeadsetAdapter {
	webxrHeadsetOnce.Do(func() {
		webxrHeadsetInstance = NewWebXRHeadsetAdapter()
	})
	return webxrHeadsetInstance
}

// NewRuntimeHeadsetAdapter returns a WebXR-backed adapter for WASM+vr builds,
// falling back to the stub adapter if the browser has no WebXR support or the
// user denies the immersive-vr session request.
func NewRuntimeHeadsetAdapter() VRHeadsetAdapter {
	webxr := sharedWebXRHeadset()
	if webxr.IsConnected() {
		return webxr
	}
	return NewStubHeadsetAdapter()
}

// NewRuntimeControllerAdapter returns a WebXR-backed controller adapter for
// WASM+vr builds, sharing the session from the headset adapter.
// Falls back to the stub adapter when no WebXR session is active.
func NewRuntimeControllerAdapter() VRControllerAdapter {
	webxr := sharedWebXRHeadset()
	ctrl := NewWebXRControllerAdapter(webxr)
	if ctrl.IsConnected("left") || ctrl.IsConnected("right") {
		return ctrl
	}
	return NewStubControllerAdapter()
}
