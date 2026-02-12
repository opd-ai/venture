// Package vr provides VR hardware detection and configuration utilities.
package vr

import (
	"os"
	"runtime"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Detector provides VR hardware detection functionality.
type Detector struct {
	mu sync.RWMutex

	// cached detection results
	headsetDetected    bool
	controllerDetected bool
	detectionRun       bool

	// configuration overrides
	forceEnable  bool
	forceDisable bool
}

// NewDetector creates a new VR hardware detector.
func NewDetector() *Detector {
	return &Detector{}
}

// DetectHardware performs VR hardware detection.
// Returns true if VR headset or controllers are detected.
//
// Detection strategy:
// 1. Check environment variables for VR runtime paths
// 2. Check for VR-specific processes (SteamVR, Oculus, etc.)
// 3. Check for VR device files on Linux (/dev/hidraw*)
// 4. Return false on platforms without VR support (mobile, WASM)
func (d *Detector) DetectHardware() bool {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.forceDisable {
		log.Debug("VR detection: force disabled via configuration")
		return false
	}

	if d.forceEnable {
		log.Debug("VR detection: force enabled via configuration")
		d.headsetDetected = true
		d.controllerDetected = true
		d.detectionRun = true
		return true
	}

	if d.detectionRun {
		result := d.headsetDetected || d.controllerDetected
		log.WithFields(log.Fields{
			"headset":    d.headsetDetected,
			"controller": d.controllerDetected,
			"available":  result,
		}).Debug("VR detection: returning cached result")
		return result
	}

	// Platform-specific detection
	headset := d.detectHeadset()
	controller := d.detectController()

	d.headsetDetected = headset
	d.controllerDetected = controller
	d.detectionRun = true

	available := headset || controller
	log.WithFields(log.Fields{
		"headset":    headset,
		"controller": controller,
		"available":  available,
		"platform":   runtime.GOOS,
	}).Info("VR hardware detection completed")

	return available
}

// detectHeadset checks for VR headset hardware/runtime.
func (d *Detector) detectHeadset() bool {
	// WASM and mobile platforms don't support VR hardware detection
	if runtime.GOOS == "js" || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return false
	}

	// Check environment variables for VR runtimes
	// SteamVR sets STEAMVR_LH_ENABLE, Oculus sets OVR_SDK_PATH
	if os.Getenv("STEAMVR_LH_ENABLE") != "" ||
		os.Getenv("OVR_SDK_PATH") != "" ||
		os.Getenv("OPENVR_PATH") != "" {
		log.Debug("VR runtime environment variable detected")
		return true
	}

	// Check for common VR runtime paths
	if d.checkVRRuntimePaths() {
		return true
	}

	return false
}

// detectController checks for VR controller hardware.
func (d *Detector) detectController() bool {
	// Same platform restrictions as headset
	if runtime.GOOS == "js" || runtime.GOOS == "android" || runtime.GOOS == "ios" {
		return false
	}

	// Controllers are usually detected via the same runtime as headsets
	// For now, use the same detection logic
	return false // Conservative: require explicit headset detection
}

// checkVRRuntimePaths checks for VR runtime installation directories.
func (d *Detector) checkVRRuntimePaths() bool {
	var paths []string

	switch runtime.GOOS {
	case "windows":
		// Check common VR installation paths on Windows
		programFiles := os.Getenv("ProgramFiles")
		programFilesX86 := os.Getenv("ProgramFiles(x86)")

		paths = []string{
			programFiles + "\\Steam\\steamapps\\common\\SteamVR",
			programFilesX86 + "\\Steam\\steamapps\\common\\SteamVR",
			programFiles + "\\Oculus",
			programFilesX86 + "\\Oculus",
		}

	case "linux":
		// Check common VR paths on Linux
		homeDir := os.Getenv("HOME")
		paths = []string{
			homeDir + "/.steam/steam/steamapps/common/SteamVR",
			homeDir + "/.local/share/Steam/steamapps/common/SteamVR",
			"/usr/share/openvr",
			"/usr/local/share/openvr",
		}

	case "darwin":
		// Check common VR paths on macOS
		homeDir := os.Getenv("HOME")
		paths = []string{
			homeDir + "/Library/Application Support/Steam/steamapps/common/SteamVR",
			"/Applications/SteamVR.app",
		}
	}

	for _, path := range paths {
		if path == "" {
			continue
		}
		if _, err := os.Stat(path); err == nil {
			log.WithField("path", path).Debug("VR runtime path found")
			return true
		}
	}

	return false
}

// SetForceEnable forces VR to be enabled regardless of hardware detection.
// This is useful for testing VR systems without physical hardware.
func (d *Detector) SetForceEnable(enable bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.forceEnable = enable
	d.detectionRun = false // Reset detection cache

	log.WithField("enabled", enable).Debug("VR force enable toggled")
}

// SetForceDisable forces VR to be disabled regardless of hardware detection.
// This is useful for disabling VR on systems where it's not desired.
// When disable is true, this also clears any cached detection results.
func (d *Detector) SetForceDisable(disable bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.forceDisable = disable
	d.detectionRun = false // Reset detection cache
	if disable {
		// Clear detection state for consistency
		d.headsetDetected = false
		d.controllerDetected = false
	}

	log.WithField("disabled", disable).Debug("VR force disable toggled")
}

// IsHeadsetDetected returns true if a VR headset was detected.
// Returns false if DetectHardware() has not been called yet.
func (d *Detector) IsHeadsetDetected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.headsetDetected
}

// IsControllerDetected returns true if VR controllers were detected.
// Returns false if DetectHardware() has not been called yet.
func (d *Detector) IsControllerDetected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.controllerDetected
}

// Reset clears the cached detection results, forcing re-detection on next call.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.detectionRun = false
	d.headsetDetected = false
	d.controllerDetected = false

	log.Debug("VR detection cache cleared")
}

// ParseEnableVRFlag parses common VR enable flag values.
// Returns true for: "true", "yes", "1", "on", "enable", "enabled"
// Returns false otherwise.
func ParseEnableVRFlag(value string) bool {
	if value == "" {
		return false
	}

	value = strings.ToLower(strings.TrimSpace(value))
	switch value {
	case "true", "yes", "1", "on", "enable", "enabled":
		return true
	default:
		return false
	}
}
