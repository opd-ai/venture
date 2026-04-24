//go:build vr && !js

// Package engine provides OpenXR 1.x hardware integration via cgo for vr builds.
//
// This file is compiled only with -tags vr on non-WASM targets.
// The !js constraint prevents cgo-based code from WASM builds, where VR uses
// the WebXR Device API (see vr_webxr_adapters.go for the js build).
//
// # OpenXR Integration
//
// The Khronos OpenXR Loader is used directly via cgo.
// Required packages (Linux): libopenxr-loader1 libopenxr-dev
// Required packages (Fedora/RHEL): openxr-loader openxr-loader-devel
// On Windows: download the Khronos OpenXR SDK and set CGO_LDFLAGS.
//
// Initialization tries: xrCreateInstance → xrGetSystem → xrCreateSession.
// connected is set to true only after all three succeed, so the factory in
// vr_adapter_factory_openxr.go transparently falls back to the stub adapter
// on machines without an active OpenXR runtime.
//
// Build this file with:
//
//	go build -tags vr ./...
//	go test  -tags vr ./pkg/engine/...
package engine

/*
#cgo linux   LDFLAGS: -lopenxr_loader -lm
#cgo windows LDFLAGS: -lopenxr_loader
#include <openxr/openxr.h>
#include <string.h>
#include <math.h>
#include <stdlib.h>

// xrState holds all OpenXR handles shared between the headset and controller
// adapters.  Only one VR session per process is expected.
typedef struct {
    XrInstance    instance;
    XrSystemId    systemId;
    XrSession     session;
    XrSpace       refSpace;
    XrActionSet   actionSet;
    // per-hand actions
    XrAction      triggerAction;
    XrAction      gripAction;
    XrAction      thumbstickAction;
    XrAction      thumbstickClickAction;
    XrAction      buttonAAction;
    XrAction      buttonBAction;
    XrAction      buttonXAction;
    XrAction      buttonYAction;
    XrAction      menuAction;
    // dedicated haptic vibration output action (XR_ACTION_TYPE_VIBRATION_OUTPUT)
    XrAction      vibrationAction;
    XrPath        leftHandPath;
    XrPath        rightHandPath;
    int           initialized;
} xrState;

// gState is the singleton OpenXR state for this process.
static xrState gState;

// xr_euler_from_quat converts an XrQuaternionf to Euler angles (pitch, yaw, roll)
// in radians.  The conversion follows the standard right-hand ZYX convention.
static void xr_euler_from_quat(XrQuaternionf q,
                                double *pitch, double *yaw, double *roll) {
    double sinr_cosp = 2.0 * (q.w * q.x + q.y * q.z);
    double cosr_cosp = 1.0 - 2.0 * (q.x * q.x + q.y * q.y);
    *roll = atan2(sinr_cosp, cosr_cosp);

    double sinp = 2.0 * (q.w * q.y - q.z * q.x);
    if (fabs(sinp) >= 1.0)
        *pitch = copysign(M_PI / 2.0, sinp);
    else
        *pitch = asin(sinp);

    double siny_cosp = 2.0 * (q.w * q.z + q.x * q.y);
    double cosy_cosp = 1.0 - 2.0 * (q.y * q.y + q.z * q.z);
    *yaw = atan2(siny_cosp, cosy_cosp);
}

// xr_create_action is a helper that creates a single XrAction.
static XrResult xr_create_action(XrActionSet actionSet,
                                  const char *name,
                                  const char *localizedName,
                                  XrActionType type,
                                  uint32_t    countSubactionPaths,
                                  const XrPath *subactionPaths,
                                  XrAction    *out) {
    XrActionCreateInfo info;
    memset(&info, 0, sizeof(info));
    info.type                  = XR_TYPE_ACTION_CREATE_INFO;
    info.actionType            = type;
    info.countSubactionPaths   = countSubactionPaths;
    info.subactionPaths        = subactionPaths;
    strncpy(info.actionName,          name,          XR_MAX_ACTION_NAME_SIZE - 1);
    strncpy(info.localizedActionName, localizedName, XR_MAX_LOCALIZED_ACTION_NAME_SIZE - 1);
    return xrCreateAction(actionSet, &info, out);
}

// xr_suggest_bindings pushes interaction-profile binding suggestions for the
// KHR/simple_controller profile.  Failures are silently ignored so the adapter
// degrades gracefully on runtimes that only support proprietary profiles.
static void xr_suggest_simple_bindings(XrInstance instance) {
    XrPath profilePath;
    if (XR_FAILED(xrStringToPath(instance,
            "/interaction_profiles/khr/simple_controller",
            &profilePath))) return;

    XrPath paths[10];
    XrActionSuggestedBinding bindings[10];
    uint32_t n = 0;

#define BIND(action_field, path_str)                                      \
    if (XR_SUCCEEDED(xrStringToPath(instance, path_str, &paths[n]))) {   \
        bindings[n].action  = gState.action_field;                        \
        bindings[n].binding = paths[n];                                   \
        n++;                                                              \
    }

    BIND(triggerAction,        "/user/hand/left/input/select/click")
    BIND(gripAction,           "/user/hand/left/input/squeeze/click")
    BIND(menuAction,           "/user/hand/left/input/menu/click")
    BIND(triggerAction,        "/user/hand/right/input/select/click")
    BIND(gripAction,           "/user/hand/right/input/squeeze/click")
    BIND(vibrationAction,      "/user/hand/left/output/haptic")
    BIND(vibrationAction,      "/user/hand/right/output/haptic")
#undef BIND

    XrInteractionProfileSuggestedBinding sug;
    memset(&sug, 0, sizeof(sug));
    sug.type                  = XR_TYPE_INTERACTION_PROFILE_SUGGESTED_BINDING;
    sug.interactionProfile    = profilePath;
    sug.countSuggestedBindings = n;
    sug.suggestedBindings     = bindings;
    xrSuggestInteractionProfileBindings(instance, &sug);
}

// xr_init attempts the full OpenXR init sequence.
// Returns 1 on success, 0 on any failure (graceful degradation).
static int xr_init(void) {
    if (gState.initialized) return 1;

    // --- Instance --------------------------------------------------------
    XrApplicationInfo appInfo;
    memset(&appInfo, 0, sizeof(appInfo));
    strncpy(appInfo.applicationName, "Venture",        XR_MAX_APPLICATION_NAME_SIZE - 1);
    strncpy(appInfo.engineName,      "VentureEngine",  XR_MAX_ENGINE_NAME_SIZE - 1);
    appInfo.applicationVersion = 1;
    appInfo.engineVersion      = 1;
    appInfo.apiVersion         = XR_CURRENT_API_VERSION;

    XrInstanceCreateInfo instInfo;
    memset(&instInfo, 0, sizeof(instInfo));
    instInfo.type            = XR_TYPE_INSTANCE_CREATE_INFO;
    instInfo.applicationInfo = appInfo;

    if (XR_FAILED(xrCreateInstance(&instInfo, &gState.instance)))
        return 0;

    // --- System ----------------------------------------------------------
    XrSystemGetInfo sysInfo;
    memset(&sysInfo, 0, sizeof(sysInfo));
    sysInfo.type      = XR_TYPE_SYSTEM_GET_INFO;
    sysInfo.formFactor = XR_FORM_FACTOR_HEAD_MOUNTED_DISPLAY;

    if (XR_FAILED(xrGetSystem(gState.instance, &sysInfo, &gState.systemId))) {
        xrDestroyInstance(gState.instance);
        return 0;
    }

    // --- Session (no graphics binding — headless) -------------------------
    XrSessionCreateInfo sessInfo;
    memset(&sessInfo, 0, sizeof(sessInfo));
    sessInfo.type     = XR_TYPE_SESSION_CREATE_INFO;
    sessInfo.next     = NULL;   // no graphics binding (headless)
    sessInfo.systemId = gState.systemId;

    if (XR_FAILED(xrCreateSession(gState.instance, &sessInfo, &gState.session))) {
        xrDestroyInstance(gState.instance);
        return 0;
    }

    // --- Reference space -------------------------------------------------
    XrPosef identity;
    memset(&identity, 0, sizeof(identity));
    identity.orientation.w = 1.0f;

    XrReferenceSpaceCreateInfo spaceInfo;
    memset(&spaceInfo, 0, sizeof(spaceInfo));
    spaceInfo.type                 = XR_TYPE_REFERENCE_SPACE_CREATE_INFO;
    spaceInfo.referenceSpaceType   = XR_REFERENCE_SPACE_TYPE_LOCAL;
    spaceInfo.poseInReferenceSpace = identity;

    if (XR_FAILED(xrCreateReferenceSpace(gState.session, &spaceInfo, &gState.refSpace))) {
        xrDestroySession(gState.session);
        xrDestroyInstance(gState.instance);
        return 0;
    }

    // --- Begin session ---------------------------------------------------
    XrSessionBeginInfo beginInfo;
    memset(&beginInfo, 0, sizeof(beginInfo));
    beginInfo.type                         = XR_TYPE_SESSION_BEGIN_INFO;
    beginInfo.primaryViewConfigurationType = XR_VIEW_CONFIGURATION_TYPE_PRIMARY_STEREO;
    // Ignore result: session may not be in READY state yet; we retry on first use.
    xrBeginSession(gState.session, &beginInfo);

    // --- Action set ------------------------------------------------------
    XrActionSetCreateInfo setInfo;
    memset(&setInfo, 0, sizeof(setInfo));
    setInfo.type = XR_TYPE_ACTION_SET_CREATE_INFO;
    strncpy(setInfo.actionSetName,          "venture_input",  XR_MAX_ACTION_SET_NAME_SIZE - 1);
    strncpy(setInfo.localizedActionSetName, "Venture Input",  XR_MAX_LOCALIZED_ACTION_SET_NAME_SIZE - 1);
    setInfo.priority = 0;

    if (XR_FAILED(xrCreateActionSet(gState.instance, &setInfo, &gState.actionSet))) {
        xrDestroySession(gState.session);
        xrDestroyInstance(gState.instance);
        return 0;
    }

    // Subaction paths for left/right hands
    xrStringToPath(gState.instance, "/user/hand/left",  &gState.leftHandPath);
    xrStringToPath(gState.instance, "/user/hand/right", &gState.rightHandPath);
    XrPath hands[2] = {gState.leftHandPath, gState.rightHandPath};

    // --- Per-hand float / vector / boolean actions -----------------------
    xr_create_action(gState.actionSet, "trigger",         "Trigger",         XR_ACTION_TYPE_FLOAT_INPUT,     2, hands, &gState.triggerAction);
    xr_create_action(gState.actionSet, "grip",            "Grip",            XR_ACTION_TYPE_FLOAT_INPUT,     2, hands, &gState.gripAction);
    xr_create_action(gState.actionSet, "thumbstick",      "Thumbstick",      XR_ACTION_TYPE_VECTOR2F_INPUT,  2, hands, &gState.thumbstickAction);
    xr_create_action(gState.actionSet, "thumbstick_click","Thumbstick Click",XR_ACTION_TYPE_BOOLEAN_INPUT,   2, hands, &gState.thumbstickClickAction);
    xr_create_action(gState.actionSet, "button_a",        "Button A",        XR_ACTION_TYPE_BOOLEAN_INPUT,   1, &gState.rightHandPath, &gState.buttonAAction);
    xr_create_action(gState.actionSet, "button_b",        "Button B",        XR_ACTION_TYPE_BOOLEAN_INPUT,   1, &gState.rightHandPath, &gState.buttonBAction);
    xr_create_action(gState.actionSet, "button_x",        "Button X",        XR_ACTION_TYPE_BOOLEAN_INPUT,   1, &gState.leftHandPath,  &gState.buttonXAction);
    xr_create_action(gState.actionSet, "button_y",        "Button Y",        XR_ACTION_TYPE_BOOLEAN_INPUT,   1, &gState.leftHandPath,  &gState.buttonYAction);
    xr_create_action(gState.actionSet, "menu",            "Menu",            XR_ACTION_TYPE_BOOLEAN_INPUT,   1, &gState.leftHandPath,  &gState.menuAction);
    // Dedicated vibration output action (XR_ACTION_TYPE_VIBRATION_OUTPUT).
    // OpenXR requires a separate output action for xrApplyHapticFeedback; reusing
    // an input action handle returns XR_ERROR_ACTION_TYPE_MISMATCH at runtime.
    xr_create_action(gState.actionSet, "vibrate",         "Vibrate",         XR_ACTION_TYPE_VIBRATION_OUTPUT,2, hands, &gState.vibrationAction);

    // Suggest bindings for the KHR simple profile
    xr_suggest_simple_bindings(gState.instance);

    // Attach action set to session
    XrSessionActionSetsAttachInfo attachInfo;
    memset(&attachInfo, 0, sizeof(attachInfo));
    attachInfo.type         = XR_TYPE_SESSION_ACTION_SETS_ATTACH_INFO;
    attachInfo.countActionSets = 1;
    attachInfo.actionSets   = &gState.actionSet;
    xrAttachSessionActionSets(gState.session, &attachInfo);

    gState.initialized = 1;
    return 1;
}

// xr_get_head_pose locates the head view at the current time and writes pose
// data into out_pitch/yaw/roll (Euler, radians) and out_x/y/z (meters).
// Returns 1 on success, 0 if pose is unavailable.
static int xr_get_head_pose(double *out_pitch, double *out_yaw, double *out_roll,
                             double *out_x,    double *out_y,   double *out_z) {
    if (!gState.initialized) return 0;

    XrViewLocateInfo locateInfo;
    memset(&locateInfo, 0, sizeof(locateInfo));
    locateInfo.type                  = XR_TYPE_VIEW_LOCATE_INFO;
    locateInfo.viewConfigurationType = XR_VIEW_CONFIGURATION_TYPE_PRIMARY_STEREO;
    locateInfo.space                 = gState.refSpace;
    // Use a zero-time (XrTime == 0 is invalid per spec but widely accepted as "now")
    locateInfo.displayTime           = 1;

    XrViewState viewState;
    memset(&viewState, 0, sizeof(viewState));
    viewState.type = XR_TYPE_VIEW_STATE;

    XrView views[2];
    memset(views, 0, sizeof(views));
    views[0].type = XR_TYPE_VIEW;
    views[1].type = XR_TYPE_VIEW;

    uint32_t viewCount = 0;
    XrResult res = xrLocateViews(gState.session, &locateInfo, &viewState, 2, &viewCount, views);
    if (XR_FAILED(res) || viewCount == 0) return 0;

    // Check orientation/position validity flags
    if (!(viewState.viewStateFlags & XR_VIEW_STATE_ORIENTATION_VALID_BIT))
        return 0;

    xr_euler_from_quat(views[0].pose.orientation, out_pitch, out_yaw, out_roll);

    if (viewState.viewStateFlags & XR_VIEW_STATE_POSITION_VALID_BIT) {
        *out_x = (double)views[0].pose.position.x;
        *out_y = (double)views[0].pose.position.y;
        *out_z = (double)views[0].pose.position.z;
    }
    return 1;
}

// xr_sync_actions pumps the action frame so get-state calls return fresh data.
static void xr_sync_actions(void) {
    if (!gState.initialized) return;
    XrActiveActionSet active;
    memset(&active, 0, sizeof(active));
    active.actionSet = gState.actionSet;

    XrActionsSyncInfo syncInfo;
    memset(&syncInfo, 0, sizeof(syncInfo));
    syncInfo.type              = XR_TYPE_ACTIONS_SYNC_INFO;
    syncInfo.countActiveActionSets = 1;
    syncInfo.activeActionSets  = &active;
    xrSyncActions(gState.session, &syncInfo);
}

// xr_get_float reads a float action state for the given hand path.
static float xr_get_float(XrAction action, XrPath handPath) {
    if (!gState.initialized || action == XR_NULL_HANDLE) return 0.0f;
    XrActionStateGetInfo getInfo;
    memset(&getInfo, 0, sizeof(getInfo));
    getInfo.type          = XR_TYPE_ACTION_STATE_GET_INFO;
    getInfo.action        = action;
    getInfo.subactionPath = handPath;
    XrActionStateFloat state;
    memset(&state, 0, sizeof(state));
    state.type = XR_TYPE_ACTION_STATE_FLOAT;
    xrGetActionStateFloat(gState.session, &getInfo, &state);
    return state.isActive ? state.currentState : 0.0f;
}

// xr_get_vec2 reads a vector2f action state for the given hand path.
static void xr_get_vec2(XrAction action, XrPath handPath, float *x, float *y) {
    *x = 0.0f; *y = 0.0f;
    if (!gState.initialized || action == XR_NULL_HANDLE) return;
    XrActionStateGetInfo getInfo;
    memset(&getInfo, 0, sizeof(getInfo));
    getInfo.type          = XR_TYPE_ACTION_STATE_GET_INFO;
    getInfo.action        = action;
    getInfo.subactionPath = handPath;
    XrActionStateVector2f state;
    memset(&state, 0, sizeof(state));
    state.type = XR_TYPE_ACTION_STATE_VECTOR2F;
    xrGetActionStateVector2f(gState.session, &getInfo, &state);
    if (state.isActive) { *x = state.currentState.x; *y = state.currentState.y; }
}

// xr_get_bool reads a boolean action state for the given hand path.
static int xr_get_bool(XrAction action, XrPath handPath) {
    if (!gState.initialized || action == XR_NULL_HANDLE) return 0;
    XrActionStateGetInfo getInfo;
    memset(&getInfo, 0, sizeof(getInfo));
    getInfo.type          = XR_TYPE_ACTION_STATE_GET_INFO;
    getInfo.action        = action;
    getInfo.subactionPath = handPath;
    XrActionStateBoolean state;
    memset(&state, 0, sizeof(state));
    state.type = XR_TYPE_ACTION_STATE_BOOLEAN;
    xrGetActionStateBoolean(gState.session, &getInfo, &state);
    return state.isActive && state.currentState;
}

// xr_apply_haptic sends vibration to the specified hand.
static void xr_apply_haptic(XrPath handPath, float intensity, float durationSec) {
    if (!gState.initialized) return;
    XrHapticVibration vib;
    memset(&vib, 0, sizeof(vib));
    vib.type      = XR_TYPE_HAPTIC_VIBRATION;
    vib.amplitude = intensity;
    vib.duration  = (XrDuration)(durationSec * 1e9f); // seconds → nanoseconds
    vib.frequency = XR_FREQUENCY_UNSPECIFIED;

    XrHapticActionInfo info;
    memset(&info, 0, sizeof(info));
    info.type          = XR_TYPE_HAPTIC_ACTION_INFO;
    // vibrationAction is XR_ACTION_TYPE_VIBRATION_OUTPUT — the correct type for
    // xrApplyHapticFeedback (OpenXR spec §11.3).
    info.action        = gState.vibrationAction;
    info.subactionPath = handPath;
    xrApplyHapticFeedback(gState.session, &info, (XrHapticBaseHeader *)&vib);
}

// xr_hand_path returns the pre-resolved XrPath for "left" or "right".
static XrPath xr_hand_path(const char *hand) {
    if (hand && hand[0] == 'l') return gState.leftHandPath;
    return gState.rightHandPath;
}

// --- String-dispatching helpers for Go callers (avoid raw CString in Go) ---

static float xr_trigger(const char *hand) {
    xr_sync_actions();
    return xr_get_float(gState.triggerAction, xr_hand_path(hand));
}

static float xr_grip(const char *hand) {
    xr_sync_actions();
    return xr_get_float(gState.gripAction, xr_hand_path(hand));
}

static void xr_thumbstick(const char *hand, float *x, float *y) {
    xr_sync_actions();
    xr_get_vec2(gState.thumbstickAction, xr_hand_path(hand), x, y);
}

static int xr_thumbstick_click(const char *hand) {
    xr_sync_actions();
    return xr_get_bool(gState.thumbstickClickAction, xr_hand_path(hand));
}

static int xr_button(const char *hand, const char *btn) {
    xr_sync_actions();
    XrPath hp = xr_hand_path(hand);
    if (btn[0] == 'a') return xr_get_bool(gState.buttonAAction, hp);
    if (btn[0] == 'b') return xr_get_bool(gState.buttonBAction, hp);
    if (btn[0] == 'x') return xr_get_bool(gState.buttonXAction, hp);
    if (btn[0] == 'y') return xr_get_bool(gState.buttonYAction, hp);
    if (btn[0] == 'm' || btn[0] == 's') return xr_get_bool(gState.menuAction, hp);
    return 0;
}

static void xr_haptic(const char *hand, float intensity, float durationSec) {
    xr_apply_haptic(xr_hand_path(hand), intensity, durationSec);
}
*/
import "C"

import (
	"unsafe"

	log "github.com/sirupsen/logrus"
)

// OpenXRHeadsetAdapter implements VRHeadsetAdapter using the OpenXR 1.x API.
//
// On machines with an active OpenXR runtime (SteamVR, Monado, Meta Link, WMR)
// the adapter initialises the runtime, creates a headless session, and reads
// live head-pose data via xrLocateViews.  On machines without a runtime the
// constructor returns an adapter with connected == false; the factory
// (vr_adapter_factory_openxr.go) falls back transparently to StubHeadsetAdapter.
//
// Implements: VRHeadsetAdapter
type OpenXRHeadsetAdapter struct {
	// connected is set to true only after xrCreateInstance + xrGetSystem +
	// xrCreateSession all succeed.
	connected bool

	// ipd is the interpupillary distance in millimetres; defaults to 63 mm.
	ipd float64
}

// NewOpenXRHeadsetAdapter creates an OpenXRHeadsetAdapter and attempts to
// initialise the OpenXR runtime (xrCreateInstance → xrGetSystem →
// xrCreateSession).  Returns the adapter regardless of whether hardware is
// detected; call IsConnected() to determine availability before use.
func NewOpenXRHeadsetAdapter() *OpenXRHeadsetAdapter {
	a := &OpenXRHeadsetAdapter{ipd: 63.0}

	if C.xr_init() != 0 {
		a.connected = true
		log.WithFields(log.Fields{
			"adapter": "openxr_headset",
			"status":  "connected",
		}).Info("OpenXR headset adapter initialised; runtime session active")
	} else {
		log.WithFields(log.Fields{
			"adapter": "openxr_headset",
			"status":  "no_runtime",
		}).Warn("OpenXR headset adapter: no runtime available — falling back to stub")
	}

	return a
}

// IsConnected returns true when the OpenXR runtime confirmed a headset session.
func (a *OpenXRHeadsetAdapter) IsConnected() bool {
	return a.connected
}

// GetHeadOrientation returns head pose orientation (pitch, yaw, roll) in radians
// by calling xrLocateViews and converting the returned XrQuaternionf to Euler
// angles.  Returns (0,0,0) when pose is unavailable.
func (a *OpenXRHeadsetAdapter) GetHeadOrientation() (pitch, yaw, roll float64) {
	if !a.connected {
		return 0, 0, 0
	}
	var cPitch, cYaw, cRoll, cX, cY, cZ C.double
	if C.xr_get_head_pose(&cPitch, &cYaw, &cRoll, &cX, &cY, &cZ) == 0 {
		return 0, 0, 0
	}
	return float64(cPitch), float64(cYaw), float64(cRoll)
}

// GetHeadPosition returns head position (x, y, z) in metres relative to the
// play-area origin via xrLocateViews.  Returns (0,0,0) when pose is unavailable.
func (a *OpenXRHeadsetAdapter) GetHeadPosition() (x, y, z float64) {
	if !a.connected {
		return 0, 0, 0
	}
	var cPitch, cYaw, cRoll, cX, cY, cZ C.double
	if C.xr_get_head_pose(&cPitch, &cYaw, &cRoll, &cX, &cY, &cZ) == 0 {
		return 0, 0, 0
	}
	return float64(cX), float64(cY), float64(cZ)
}

// GetIPD returns the interpupillary distance in millimetres (default 63 mm).
func (a *OpenXRHeadsetAdapter) GetIPD() float64 {
	return a.ipd
}

// OpenXRControllerAdapter implements VRControllerAdapter using the OpenXR 1.x
// action-input system (XrActionSet / XrAction + interaction profiles).
//
// Per-frame xr_sync_actions is called before each state read to ensure fresh
// data.  Interaction-profile bindings are registered for the KHR simple
// controller profile; runtimes translate to their hardware-specific equivalents.
//
// Implements: VRControllerAdapter
type OpenXRControllerAdapter struct {
	// connected is true when the OpenXR session is active.
	connected bool
}

// NewOpenXRControllerAdapter creates an OpenXRControllerAdapter.  The shared
// OpenXR session is re-used (or initialised if not yet done by the headset
// adapter).  Returns regardless of hardware availability; call IsConnected to check.
func NewOpenXRControllerAdapter() *OpenXRControllerAdapter {
	a := &OpenXRControllerAdapter{}

	if C.xr_init() != 0 {
		a.connected = true
		log.WithFields(log.Fields{
			"adapter": "openxr_controller",
			"status":  "connected",
		}).Info("OpenXR controller adapter initialised; action set active")
	} else {
		log.WithFields(log.Fields{
			"adapter": "openxr_controller",
			"status":  "no_runtime",
		}).Warn("OpenXR controller adapter: no runtime available — falling back to stub")
	}

	return a
}

// IsConnected returns true when the OpenXR session is active for the given hand.
func (a *OpenXRControllerAdapter) IsConnected(hand string) bool {
	return a.connected
}

// GetTrigger returns the trigger axis value [0,1] for the given hand by reading
// the XrAction bound to /user/hand/{hand}/input/trigger/value.
func (a *OpenXRControllerAdapter) GetTrigger(hand string) float64 {
	if !a.connected {
		return 0
	}
	cs := C.CString(hand)
	defer C.free(unsafe.Pointer(cs))
	return float64(C.xr_trigger(cs))
}

// GetGrip returns the grip axis value [0,1] for the given hand.
func (a *OpenXRControllerAdapter) GetGrip(hand string) float64 {
	if !a.connected {
		return 0
	}
	cs := C.CString(hand)
	defer C.free(unsafe.Pointer(cs))
	return float64(C.xr_grip(cs))
}

// GetThumbstick returns thumbstick position [-1,1]×[-1,1] for the given hand.
func (a *OpenXRControllerAdapter) GetThumbstick(hand string) (x, y float64) {
	if !a.connected {
		return 0, 0
	}
	cs := C.CString(hand)
	defer C.free(unsafe.Pointer(cs))
	var cx, cy C.float
	C.xr_thumbstick(cs, &cx, &cy)
	return float64(cx), float64(cy)
}

// IsThumbstickPressed returns whether the thumbstick click is active.
func (a *OpenXRControllerAdapter) IsThumbstickPressed(hand string) bool {
	if !a.connected {
		return false
	}
	cs := C.CString(hand)
	defer C.free(unsafe.Pointer(cs))
	return C.xr_thumbstick_click(cs) != 0
}

// GetButton returns whether the named face button is pressed.
// Supported button names: "a", "b", "x", "y", "menu", "system".
func (a *OpenXRControllerAdapter) GetButton(hand, button string) bool {
	if !a.connected {
		return false
	}
	chand := C.CString(hand)
	defer C.free(unsafe.Pointer(chand))
	cbtn := C.CString(button)
	defer C.free(unsafe.Pointer(cbtn))
	return C.xr_button(chand, cbtn) != 0
}

// SetHaptic triggers haptic feedback on the given controller.
// intensity is [0,1] and duration is in seconds.
func (a *OpenXRControllerAdapter) SetHaptic(hand string, intensity, duration float64) {
	if !a.connected {
		return
	}
	cs := C.CString(hand)
	defer C.free(unsafe.Pointer(cs))
	C.xr_haptic(cs, C.float(intensity), C.float(duration))
}
