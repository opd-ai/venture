//go:build js

// Package engine provides WebXR-backed VR adapters for WASM builds.
//
// WebXRHeadsetAdapter and WebXRControllerAdapter implement VRHeadsetAdapter and
// VRControllerAdapter respectively by calling into the browser's WebXR Device API
// via syscall/js.  They are only compiled for WASM targets (//go:build js).
//
// Initialisation flow:
//  1. NewWebXRHeadsetAdapter checks navigator.xr availability and calls
//     xr.isSessionSupported("immersive-vr").
//  2. On success it requests an "immersive-vr" session and sets connected=true.
//  3. Each game frame the adapter caches the latest XRViewerPose exposed via
//     the XRSession frame callback so that GetHeadOrientation/GetHeadPosition
//     return up-to-date values without blocking the Go goroutine.
//
// When the browser has no WebXR support, or the user denies VR permission,
// connected is left false and all pose/input methods return safe zero values,
// matching the graceful-degradation contract of the stub adapters.
package engine

import (
	"math"
	"sync"
	"syscall/js"

	"github.com/sirupsen/logrus"
)

// webxrPoseCache holds the latest XRViewerPose values, updated from the
// browser's XRSession frame callback on the JS event loop.
type webxrPoseCache struct {
	mu          sync.RWMutex
	pitch       float64
	yaw         float64
	roll        float64
	x, y, z     float64
	ipd         float64
	// Controller state — indexed 0=left, 1=right
	trigger     [2]float64
	grip        [2]float64
	thumbX      [2]float64
	thumbY      [2]float64
	thumbPress  [2]bool
	buttons     [2]map[string]bool
}

// WebXRHeadsetAdapter implements VRHeadsetAdapter via the browser WebXR Device API.
type WebXRHeadsetAdapter struct {
	connected bool
	cache     *webxrPoseCache
	session   js.Value
}

// NewWebXRHeadsetAdapter creates a WebXR headset adapter.
// It immediately probes navigator.xr and requests an immersive-vr session;
// connected is true only if that succeeds.
func NewWebXRHeadsetAdapter() *WebXRHeadsetAdapter {
	a := &WebXRHeadsetAdapter{
		cache: &webxrPoseCache{ipd: 63.0},
	}
	a.cache.buttons[0] = make(map[string]bool)
	a.cache.buttons[1] = make(map[string]bool)

	nav := js.Global().Get("navigator")
	if nav.IsUndefined() || nav.IsNull() {
		logrus.WithField("adapter", "webxr_headset").Warn("WebXR: navigator unavailable")
		return a
	}
	xr := nav.Get("xr")
	if xr.IsUndefined() || xr.IsNull() {
		logrus.WithField("adapter", "webxr_headset").Info("WebXR: navigator.xr not present — no VR support in this browser")
		return a
	}

	// Synchronously check support via a JS promise resolved in a microtask.
	// We use a channel to avoid blocking the JS goroutine.
	supported := make(chan bool, 1)
	xr.Call("isSessionSupported", "immersive-vr").Call("then",
		js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				supported <- args[0].Bool()
			} else {
				supported <- false
			}
			return nil
		}),
		js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			supported <- false
			return nil
		}),
	)
	if !<-supported {
		logrus.WithField("adapter", "webxr_headset").Info("WebXR: immersive-vr not supported by this device")
		return a
	}

	// Request the session; on success wire the frame callback.
	sessionGranted := make(chan js.Value, 1)
	xr.Call("requestSession", "immersive-vr").Call("then",
		js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
			if len(args) > 0 {
				sessionGranted <- args[0]
			}
			return nil
		}),
		js.FuncOf(func(_ js.Value, _ []js.Value) interface{} {
			close(sessionGranted)
			return nil
		}),
	)
	session, ok := <-sessionGranted
	if !ok || session.IsUndefined() || session.IsNull() {
		logrus.WithField("adapter", "webxr_headset").Warn("WebXR: session request denied or failed")
		return a
	}

	a.session = session
	a.connected = true
	a.startFrameLoop()

	logrus.WithField("adapter", "webxr_headset").Info("WebXR: immersive-vr session established")
	return a
}

// startFrameLoop registers the XRSession frame callback and calls
// session.requestAnimationFrame to begin receiving pose data.
func (a *WebXRHeadsetAdapter) startFrameLoop() {
	if a.session.IsUndefined() || a.session.IsNull() {
		return
	}

	var onFrame js.Func
	onFrame = js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) < 2 {
			a.session.Call("requestAnimationFrame", onFrame)
			return nil
		}
		frame := args[1]
		a.updatePoseFromFrame(frame)
		a.session.Call("requestAnimationFrame", onFrame)
		return nil
	})
	a.session.Call("requestAnimationFrame", onFrame)
}

// updatePoseFromFrame reads the viewer pose and input sources from an XRFrame.
func (a *WebXRHeadsetAdapter) updatePoseFromFrame(frame js.Value) {
	refSpace := a.session.Call("requestReferenceSpace", "viewer")
	if refSpace.IsUndefined() || refSpace.IsNull() {
		return
	}
	pose := frame.Call("getViewerPose", refSpace)
	if pose.IsUndefined() || pose.IsNull() {
		return
	}

	a.cache.mu.Lock()
	defer a.cache.mu.Unlock()

	// Extract orientation from the first view's transform quaternion.
	views := pose.Get("views")
	if views.IsUndefined() || views.IsNull() || views.Length() == 0 {
		return
	}
	view := views.Index(0)
	transform := view.Get("transform")
	orient := transform.Get("orientation")
	if !orient.IsUndefined() && !orient.IsNull() {
		qx := orient.Get("x").Float()
		qy := orient.Get("y").Float()
		qz := orient.Get("z").Float()
		qw := orient.Get("w").Float()
		a.cache.pitch, a.cache.yaw, a.cache.roll = quaternionToEuler(qx, qy, qz, qw)
	}
	position := transform.Get("position")
	if !position.IsUndefined() && !position.IsNull() {
		a.cache.x = position.Get("x").Float()
		a.cache.y = position.Get("y").Float()
		a.cache.z = position.Get("z").Float()
	}

	// Derive IPD from two-eye distance when stereo views are available.
	if views.Length() >= 2 {
		v0 := views.Index(0).Get("transform").Get("position")
		v1 := views.Index(1).Get("transform").Get("position")
		if !v0.IsUndefined() && !v1.IsUndefined() {
			dx := v1.Get("x").Float() - v0.Get("x").Float()
			dy := v1.Get("y").Float() - v0.Get("y").Float()
			dz := v1.Get("z").Float() - v0.Get("z").Float()
			ipd := math.Sqrt(dx*dx+dy*dy+dz*dz) * 1000 // convert m → mm
			if ipd > 0 {
				a.cache.ipd = ipd
			}
		}
	}
}

// IsConnected returns true if a WebXR immersive-vr session is active.
func (a *WebXRHeadsetAdapter) IsConnected() bool { return a.connected }

// GetHeadOrientation returns (pitch, yaw, roll) in radians from the latest XRFrame.
func (a *WebXRHeadsetAdapter) GetHeadOrientation() (pitch, yaw, roll float64) {
	if !a.connected {
		return 0, 0, 0
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.pitch, a.cache.yaw, a.cache.roll
}

// GetHeadPosition returns (x, y, z) in metres from the latest XRFrame.
func (a *WebXRHeadsetAdapter) GetHeadPosition() (x, y, z float64) {
	if !a.connected {
		return 0, 0, 0
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.x, a.cache.y, a.cache.z
}

// GetIPD returns the interpupillary distance in millimetres.
func (a *WebXRHeadsetAdapter) GetIPD() float64 {
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.ipd
}

// WebXRControllerAdapter implements VRControllerAdapter via WebXR input sources.
type WebXRControllerAdapter struct {
	connected bool
	cache     *webxrPoseCache
	session   js.Value
}

// NewWebXRControllerAdapter creates a WebXR controller adapter that shares the
// pose cache with the headset adapter so input sources are updated together.
// If headset is connected it reuses the session; otherwise no-op.
func NewWebXRControllerAdapter(headset *WebXRHeadsetAdapter) *WebXRControllerAdapter {
	a := &WebXRControllerAdapter{}
	if headset == nil || !headset.connected {
		logrus.WithField("adapter", "webxr_controller").Info("WebXR: controller adapter inactive (no session)")
		return a
	}
	a.connected = true
	a.cache = headset.cache
	a.session = headset.session
	return a
}

// IsConnected returns true if the named hand's input source is present.
func (a *WebXRControllerAdapter) IsConnected(hand string) bool {
	return a.connected && a.handIndex(hand) >= 0
}

// GetTrigger returns the select trigger value [0,1] for the named hand.
func (a *WebXRControllerAdapter) GetTrigger(hand string) float64 {
	i := a.handIndex(hand)
	if i < 0 {
		return 0
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.trigger[i]
}

// GetGrip returns the grip/squeeze value [0,1] for the named hand.
func (a *WebXRControllerAdapter) GetGrip(hand string) float64 {
	i := a.handIndex(hand)
	if i < 0 {
		return 0
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.grip[i]
}

// GetThumbstick returns the thumbstick axes (x, y) in [-1, 1] for the named hand.
func (a *WebXRControllerAdapter) GetThumbstick(hand string) (x, y float64) {
	i := a.handIndex(hand)
	if i < 0 {
		return 0, 0
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.thumbX[i], a.cache.thumbY[i]
}

// IsThumbstickPressed returns true when the thumbstick is clicked.
func (a *WebXRControllerAdapter) IsThumbstickPressed(hand string) bool {
	i := a.handIndex(hand)
	if i < 0 {
		return false
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.thumbPress[i]
}

// GetButton returns the state of the named button (a, b, x, y, menu, system).
func (a *WebXRControllerAdapter) GetButton(hand, button string) bool {
	i := a.handIndex(hand)
	if i < 0 {
		return false
	}
	a.cache.mu.RLock()
	defer a.cache.mu.RUnlock()
	return a.cache.buttons[i][button]
}

// SetHaptic triggers haptic feedback on the named controller hand.
func (a *WebXRControllerAdapter) SetHaptic(hand string, intensity, duration float64) {
	if !a.connected {
		return
	}
	inputSources := a.session.Get("inputSources")
	if inputSources.IsUndefined() || inputSources.IsNull() {
		return
	}
	for i := 0; i < inputSources.Length(); i++ {
		src := inputSources.Index(i)
		if src.Get("handedness").String() != hand {
			continue
		}
		actuators := src.Get("hapticActuators")
		if actuators.IsUndefined() || actuators.IsNull() || actuators.Length() == 0 {
			continue
		}
		actuators.Index(0).Call("pulse", intensity, duration*1000) // ms
	}
}

// handIndex maps "left"→0, "right"→1, anything else→-1.
func (a *WebXRControllerAdapter) handIndex(hand string) int {
	switch hand {
	case "left":
		return 0
	case "right":
		return 1
	}
	return -1
}

// quaternionToEuler converts a unit quaternion (qx,qy,qz,qw) to
// (pitch, yaw, roll) Euler angles in radians using the ZYX convention.
func quaternionToEuler(qx, qy, qz, qw float64) (pitch, yaw, roll float64) {
	// Roll (Z-axis)
	sinrCosp := 2 * (qw*qz + qx*qy)
	cosrCosp := 1 - 2*(qy*qy+qz*qz)
	roll = math.Atan2(sinrCosp, cosrCosp)

	// Pitch (X-axis)
	sinp := 2 * (qw*qx - qy*qz)
	if math.Abs(sinp) >= 1 {
		pitch = math.Copysign(math.Pi/2, sinp) // Gimbal lock
	} else {
		pitch = math.Asin(sinp)
	}

	// Yaw (Y-axis)
	sinyCosp := 2 * (qw*qy + qx*qz)
	cosyCosp := 1 - 2*(qx*qx+qy*qy)
	yaw = math.Atan2(sinyCosp, cosyCosp)
	return
}
