// Package engine provides the stereoscopic rendering system for VR support.

package engine

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// StereoscopicSystem manages VR stereoscopic rendering by coordinating
// dual-eye camera offsets and render target management.
type StereoscopicSystem struct {
	mu sync.RWMutex

	world *World

	// leftEyeCallback is called before rendering the left eye
	leftEyeCallback func(offsetX float64)

	// rightEyeCallback is called before rendering the right eye
	rightEyeCallback func(offsetX float64)

	// postRenderCallback is called after both eyes are rendered
	postRenderCallback func()

	// renderPhase tracks the current rendering phase
	renderPhase string

	// frameCount tracks frames for performance monitoring
	frameCount uint64

	// enabled tracks if the system is active
	enabled bool
}

const (
	// RenderPhaseIdle indicates no VR rendering in progress
	RenderPhaseIdle = "idle"

	// RenderPhaseLeftEye indicates left eye rendering in progress
	RenderPhaseLeftEye = "left_eye"

	// RenderPhaseRightEye indicates right eye rendering in progress
	RenderPhaseRightEye = "right_eye"

	// RenderPhaseComposite indicates final composition in progress
	RenderPhaseComposite = "composite"
)

// NewStereoscopicSystem creates a new stereoscopic rendering system.
func NewStereoscopicSystem(world *World) *StereoscopicSystem {
	log.WithFields(log.Fields{
		"system_name": "stereoscopic",
	}).Debug("Creating stereoscopic system")

	return &StereoscopicSystem{
		world:       world,
		renderPhase: RenderPhaseIdle,
		enabled:     false,
	}
}

// Update processes stereoscopic rendering for entities with StereoscopicComponent.
// This system coordinates the dual-eye render passes.
func (s *StereoscopicSystem) Update(entities []*Entity, deltaTime float64) {
	s.mu.Lock()
	if !s.enabled {
		s.mu.Unlock()
		return
	}

	s.frameCount++
	s.mu.Unlock()

	for _, entity := range entities {
		comp, ok := entity.GetComponent("stereoscopic")
		if !ok || comp == nil {
			continue
		}

		stereo, ok := comp.(*StereoscopicComponent)
		if !ok {
			continue
		}

		if !stereo.IsEnabled() {
			continue
		}

		// Process VR rendering for this entity
		s.processStereoscopicEntity(stereo)
	}
}

// processStereoscopicEntity handles rendering for a single VR-enabled entity.
func (s *StereoscopicSystem) processStereoscopicEntity(stereo *StereoscopicComponent) {
	// Render left eye
	s.mu.Lock()
	s.renderPhase = RenderPhaseLeftEye
	leftCallback := s.leftEyeCallback
	s.mu.Unlock()

	stereo.SetCurrentEye(EyeLeft)
	leftOffset := stereo.GetEyeOffset(EyeLeft)

	if leftCallback != nil {
		leftCallback(leftOffset)
	}

	// Render right eye
	s.mu.Lock()
	s.renderPhase = RenderPhaseRightEye
	rightCallback := s.rightEyeCallback
	s.mu.Unlock()

	stereo.SetCurrentEye(EyeRight)
	rightOffset := stereo.GetEyeOffset(EyeRight)

	if rightCallback != nil {
		rightCallback(rightOffset)
	}

	// Composite both eyes
	s.mu.Lock()
	s.renderPhase = RenderPhaseComposite
	postCallback := s.postRenderCallback
	s.mu.Unlock()

	if postCallback != nil {
		postCallback()
	}

	s.mu.Lock()
	s.renderPhase = RenderPhaseIdle
	s.mu.Unlock()
}

// SetLeftEyeCallback sets the callback invoked before left eye rendering.
// The callback receives the camera X offset for the left eye.
func (s *StereoscopicSystem) SetLeftEyeCallback(cb func(offsetX float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.leftEyeCallback = cb
}

// SetRightEyeCallback sets the callback invoked before right eye rendering.
// The callback receives the camera X offset for the right eye.
func (s *StereoscopicSystem) SetRightEyeCallback(cb func(offsetX float64)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rightEyeCallback = cb
}

// SetPostRenderCallback sets the callback invoked after both eyes are rendered.
// This is typically used to composite the final side-by-side image.
func (s *StereoscopicSystem) SetPostRenderCallback(cb func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.postRenderCallback = cb
}

// GetRenderPhase returns the current rendering phase.
func (s *StereoscopicSystem) GetRenderPhase() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.renderPhase
}

// SetEnabled enables or disables the stereoscopic system.
func (s *StereoscopicSystem) SetEnabled(enabled bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.enabled = enabled

	log.WithFields(log.Fields{
		"system_name": "stereoscopic",
		"enabled":     enabled,
	}).Info("Stereoscopic system toggled")
}

// IsEnabled returns whether the system is enabled.
func (s *StereoscopicSystem) IsEnabled() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.enabled
}

// GetFrameCount returns the number of frames processed.
func (s *StereoscopicSystem) GetFrameCount() uint64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.frameCount
}

// CalculateStereoProjection calculates the asymmetric projection offset
// for stereo rendering based on eye offset and convergence distance.
// Returns the horizontal projection shift.
func CalculateStereoProjection(eyeOffset, convergence, fov float64) float64 {
	if convergence <= 0 {
		convergence = 1.0
	}

	// Calculate the horizontal shift based on eye offset and convergence
	// This creates toe-in effect where both eyes converge at the focal plane
	shift := eyeOffset / convergence

	// Scale by field of view factor (assuming radians)
	// Wider FOV requires smaller shift
	if fov > 0 {
		shift *= 1.0 / fov
	}

	return shift
}

// CalculateViewportForEye returns the viewport coordinates for side-by-side rendering.
// Returns (x, y, width, height) for the specified eye.
func CalculateViewportForEye(eye string, totalWidth, totalHeight int) (x, y, width, height int) {
	halfWidth := totalWidth / 2

	if eye == EyeLeft {
		return 0, 0, halfWidth, totalHeight
	}
	return halfWidth, 0, halfWidth, totalHeight
}

// ApplyAsymmetricFrustum calculates asymmetric frustum parameters for stereo cameras.
// This prevents distortion when rendering off-axis views.
// Returns (left, right, top, bottom) frustum plane distances at the near plane.
func ApplyAsymmetricFrustum(eyeOffset, near, fov, aspect float64) (left, right, top, bottom float64) {
	// Calculate frustum half-dimensions at near plane
	halfHeight := near * fov / 2.0 // Simplified, assumes fov is half-angle tangent
	halfWidth := halfHeight * aspect

	// Shift frustum based on eye offset (scaled to near plane)
	shift := eyeOffset * near

	left = -halfWidth + shift
	right = halfWidth + shift
	top = halfHeight
	bottom = -halfHeight

	return left, right, top, bottom
}

// VRRenderStats holds statistics about VR rendering performance.
type VRRenderStats struct {
	FrameCount       uint64
	LeftEyeRenderMs  float64
	RightEyeRenderMs float64
	CompositeMs      float64
	TotalFrameMs     float64
}

// GetStats returns current VR rendering statistics.
func (s *StereoscopicSystem) GetStats() VRRenderStats {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return VRRenderStats{
		FrameCount: s.frameCount,
		// Actual timing would be populated by render callbacks
	}
}
