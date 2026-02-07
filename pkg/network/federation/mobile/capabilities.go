// Package mobile capabilities detection for platform-specific features.
// This file provides detection of WebRTC and other network capabilities
// with graceful degradation when features are unavailable.
package mobile

import (
	"runtime"

	"github.com/sirupsen/logrus"
)

// NetworkCapability represents available network features
type NetworkCapability int

const (
	// CapabilityWebRTC indicates WebRTC peer connections are available
	CapabilityWebRTC NetworkCapability = 1 << iota
	// CapabilityWebSocket indicates WebSocket fallback is available
	CapabilityWebSocket
	// CapabilityHTTP indicates standard HTTP/HTTPS is available
	CapabilityHTTP
)

// PlatformCapabilities holds detected capabilities for current platform
type PlatformCapabilities struct {
	WebRTCAvailable    bool
	WebSocketAvailable bool
	HTTPAvailable      bool
	Platform           string
	Restrictions       []string
}

// DetectCapabilities determines available network features for current platform
func DetectCapabilities() *PlatformCapabilities {
	platform := runtime.GOOS
	caps := &PlatformCapabilities{
		Platform:           platform,
		HTTPAvailable:      true, // HTTP available on all platforms
		WebSocketAvailable: true, // WebSocket available on all platforms
		Restrictions:       []string{},
	}

	// WebRTC availability by platform
	switch platform {
	case "js":
		// WASM/browser - WebRTC available but requires user interaction
		caps.WebRTCAvailable = true
		caps.Restrictions = append(caps.Restrictions, "WebRTC requires HTTPS in production")
		caps.Restrictions = append(caps.Restrictions, "WebRTC requires user gesture for permissions")

	case "android":
		// Android - WebRTC available via native libraries
		caps.WebRTCAvailable = true

	case "ios":
		// iOS - WebRTC available via native frameworks
		caps.WebRTCAvailable = true

	default:
		// Desktop/unknown platforms - assume WebRTC not available without external libraries
		caps.WebRTCAvailable = false
		caps.Restrictions = append(caps.Restrictions, "WebRTC requires external library (pion/webrtc)")

		logrus.WithFields(logrus.Fields{
			"platform": platform,
			"webrtc":   false,
		}).Debug("WebRTC not available on this platform")
	}

	return caps
}

// SupportsWebRTC returns true if WebRTC is available on current platform
func SupportsWebRTC() bool {
	caps := DetectCapabilities()
	return caps.WebRTCAvailable
}

// GetFallbackTransport returns recommended fallback transport when WebRTC unavailable
func GetFallbackTransport() string {
	caps := DetectCapabilities()

	if caps.WebSocketAvailable {
		return "websocket"
	}

	if caps.HTTPAvailable {
		return "http_polling"
	}

	return "none"
}

// LogCapabilities logs detected capabilities for debugging
func LogCapabilities(logger *logrus.Entry) {
	caps := DetectCapabilities()

	logger.WithFields(logrus.Fields{
		"platform":     caps.Platform,
		"webrtc":       caps.WebRTCAvailable,
		"websocket":    caps.WebSocketAvailable,
		"http":         caps.HTTPAvailable,
		"restrictions": caps.Restrictions,
	}).Info("Detected platform capabilities")

	if !caps.WebRTCAvailable {
		logger.WithField("fallback", GetFallbackTransport()).
			Warn("WebRTC unavailable, using fallback transport")
	}
}
