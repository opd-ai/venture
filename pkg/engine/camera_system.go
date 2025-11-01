// Package engine provides camera control for view management.
// This file implements CameraSystem which handles camera positioning, zoom,
// and viewport calculations for rendering.
package engine

import "math"

// CameraComponent represents a camera that follows an entity.
type CameraComponent struct {
	// Target offset from entity position
	OffsetX, OffsetY float64

	// Zoom level (1.0 = normal, 2.0 = 2x zoom, etc.)
	Zoom float64

	// Camera bounds (for limiting camera movement)
	MinX, MinY float64
	MaxX, MaxY float64

	// Smoothing factor for camera movement (0.0 = instant, 1.0 = very smooth)
	Smoothing float64

	// Current camera position (world coordinates)
	X, Y float64

	// GAP-012 REPAIR: Screen shake for visual feedback
	ShakeIntensity float64 // Current shake intensity (pixels)
	ShakeDecay     float64 // Shake decay rate per second
	ShakeOffsetX   float64 // Current shake offset X
	ShakeOffsetY   float64 // Current shake offset Y
}

// Type returns the component type identifier.
func (c *CameraComponent) Type() string {
	return "camera"
}

// NewCameraComponent creates a new camera component with default settings.
func NewCameraComponent() *CameraComponent {
	return &CameraComponent{
		OffsetX:        0,
		OffsetY:        0,
		Zoom:           1.0,
		MinX:           math.Inf(-1),
		MinY:           math.Inf(-1),
		MaxX:           math.Inf(1),
		MaxY:           math.Inf(1),
		Smoothing:      0.1,
		X:              0,
		Y:              0,
		ShakeIntensity: 0,
		ShakeDecay:     5.0, // Shake decays in ~0.2 seconds
		ShakeOffsetX:   0,
		ShakeOffsetY:   0,
	}
}

// CameraSystem manages camera positioning and viewport.
type CameraSystem struct {
	// Screen dimensions
	ScreenWidth  int
	ScreenHeight int

	// Active camera entity (if any)
	activeCamera *Entity

	// Phase 10.3: Accessibility settings for screen shake and effects
	Accessibility *AccessibilitySettings
}

// NewCameraSystem creates a new camera system.
func NewCameraSystem(screenWidth, screenHeight int) *CameraSystem {
	return &CameraSystem{
		ScreenWidth:   screenWidth,
		ScreenHeight:  screenHeight,
		Accessibility: NewAccessibilitySettings(), // Phase 10.3: Default accessibility
	}
}

// Update updates camera positions to follow their target entities.
func (s *CameraSystem) Update(entities []*Entity, deltaTime float64) {
	// Phase 10.3: Apply hit-stop time dilation
	effectiveDeltaTime := s.calculateEffectiveDeltaTime(entities, deltaTime)

	for _, entity := range entities {
		cameraComp, ok := entity.GetComponent("camera")
		if !ok {
			continue
		}

		camera := cameraComp.(*CameraComponent)

		// Get entity position
		posComp, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		pos := posComp.(*PositionComponent)

		// Calculate target camera position (entity position + offset)
		targetX := pos.X + camera.OffsetX
		targetY := pos.Y + camera.OffsetY

		// Apply smoothing (lerp with frame-rate independent exponential decay)
		if camera.Smoothing > 0 {
			// Use exponential decay formula for frame-rate independence
			// Higher smoothing value = slower camera tracking
			// alpha approaches 1 as deltaTime increases, ensuring smooth convergence
			alpha := 1.0 - math.Exp(-effectiveDeltaTime/camera.Smoothing)
			camera.X += (targetX - camera.X) * alpha
			camera.Y += (targetY - camera.Y) * alpha
		} else {
			camera.X = targetX
			camera.Y = targetY
		}

		// Apply bounds
		if camera.X < camera.MinX {
			camera.X = camera.MinX
		}
		if camera.X > camera.MaxX {
			camera.X = camera.MaxX
		}
		if camera.Y < camera.MinY {
			camera.Y = camera.MinY
		}
		if camera.Y > camera.MaxY {
			camera.Y = camera.MaxY
		}

		// GAP-012 REPAIR: Update screen shake (basic)
		if camera.ShakeIntensity > 0 {
			// Decay shake intensity over time
			camera.ShakeIntensity -= camera.ShakeDecay * effectiveDeltaTime
			if camera.ShakeIntensity < 0 {
				camera.ShakeIntensity = 0
				camera.ShakeOffsetX = 0
				camera.ShakeOffsetY = 0
			} else {
				// Generate random shake offset within intensity radius
				// Use simple pseudo-random based on time for shake variation
				angle := float64(int(camera.X*1000+camera.Y*1000)%360) * (math.Pi / 180.0)
				camera.ShakeOffsetX = math.Cos(angle) * camera.ShakeIntensity
				camera.ShakeOffsetY = math.Sin(angle) * camera.ShakeIntensity
			}
		}

		// Phase 10.3: Update advanced screen shake
		s.updateAdvancedShake(entity, effectiveDeltaTime)
	}
}

// calculateEffectiveDeltaTime applies hit-stop time dilation.
// Phase 10.3: Checks for active hit-stop and adjusts delta time.
func (s *CameraSystem) calculateEffectiveDeltaTime(entities []*Entity, deltaTime float64) float64 {
	// Find any active hit-stop component
	for _, entity := range entities {
		hitStopComp, ok := entity.GetComponent("hitStop")
		if !ok {
			continue
		}

		hitStop := hitStopComp.(*HitStopComponent)
		if hitStop.IsActive() {
			// Update hit-stop elapsed time with REAL delta time (not scaled)
			hitStop.Elapsed += deltaTime

			// Check if hit-stop finished
			if hitStop.Elapsed >= hitStop.Duration {
				hitStop.Reset()
				return deltaTime
			}

			// Return scaled delta time
			return deltaTime * hitStop.GetTimeScale()
		}
	}

	return deltaTime
}

// updateAdvancedShake updates the advanced screen shake component.
// Phase 10.3: Handles ScreenShakeComponent with frequency control.
func (s *CameraSystem) updateAdvancedShake(entity *Entity, deltaTime float64) {
	shakeComp, ok := entity.GetComponent("screenShake")
	if !ok {
		return
	}

	shake := shakeComp.(*ScreenShakeComponent)
	if !shake.IsShaking() {
		return
	}

	// Update elapsed time
	shake.Elapsed += deltaTime

	// Check if shake finished
	if shake.Elapsed >= shake.Duration {
		shake.Reset()
		return
	}

	// Calculate offset
	shake.CalculateOffset()
}

// SetActiveCamera sets the active camera for rendering.
func (s *CameraSystem) SetActiveCamera(entity *Entity) {
	s.activeCamera = entity
}

// GetActiveCamera returns the currently active camera entity.
func (s *CameraSystem) GetActiveCamera() *Entity {
	return s.activeCamera
}

// WorldToScreen converts world coordinates to screen coordinates using the active camera.
func (s *CameraSystem) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	if s.activeCamera == nil {
		return worldX, worldY
	}

	cameraComp, ok := s.activeCamera.GetComponent("camera")
	if !ok {
		return worldX, worldY
	}
	camera := cameraComp.(*CameraComponent)

	// Apply camera transform
	screenX = (worldX - camera.X) * camera.Zoom
	screenY = (worldY - camera.Y) * camera.Zoom

	// Center on screen
	screenX += float64(s.ScreenWidth) / 2
	screenY += float64(s.ScreenHeight) / 2

	// GAP-012 REPAIR: Apply screen shake offset
	screenX += camera.ShakeOffsetX
	screenY += camera.ShakeOffsetY

	return screenX, screenY
}

// ScreenToWorld converts screen coordinates to world coordinates using the active camera.
func (s *CameraSystem) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	if s.activeCamera == nil {
		return screenX, screenY
	}

	cameraComp, ok := s.activeCamera.GetComponent("camera")
	if !ok {
		return screenX, screenY
	}
	camera := cameraComp.(*CameraComponent)

	// Remove screen centering
	worldX = screenX - float64(s.ScreenWidth)/2
	worldY = screenY - float64(s.ScreenHeight)/2

	// Apply inverse camera transform
	worldX = worldX/camera.Zoom + camera.X
	worldY = worldY/camera.Zoom + camera.Y

	return worldX, worldY
}

// IsVisible checks if a world position is visible on screen.
func (s *CameraSystem) IsVisible(worldX, worldY, radius float64) bool {
	screenX, screenY := s.WorldToScreen(worldX, worldY)

	// Check if within screen bounds (with margin for radius)
	margin := radius * 2
	return screenX >= -margin && screenX <= float64(s.ScreenWidth)+margin &&
		screenY >= -margin && screenY <= float64(s.ScreenHeight)+margin
}

// GetPosition returns the current camera position in world coordinates.
//
// Returns:
//
//	x, y - the world coordinates of the camera center.
//
// If there is no active camera, or if the active camera entity does not have a camera component,
// returns (0, 0).
func (s *CameraSystem) GetPosition() (float64, float64) {
	if s.activeCamera == nil {
		return 0, 0
	}

	cameraComp, ok := s.activeCamera.GetComponent("camera")
	if !ok {
		return 0, 0
	}
	camera := cameraComp.(*CameraComponent)

	return camera.X, camera.Y
}

// Shake triggers a screen shake effect on the active camera.
// GAP-012 REPAIR: Provides visual feedback for impacts and heavy actions.
// Phase 10.3: Respects accessibility settings.
// intensity: shake magnitude in pixels (typical values: 2-10)
func (s *CameraSystem) Shake(intensity float64) {
	if s.activeCamera == nil {
		return
	}

	// Phase 10.3: Apply accessibility multiplier
	intensity = s.Accessibility.ApplyShakeIntensity(intensity)
	if intensity == 0.0 {
		return // Shake disabled via accessibility
	}

	cameraComp, ok := s.activeCamera.GetComponent("camera")
	if !ok {
		return
	}
	camera := cameraComp.(*CameraComponent)

	// Add to existing shake (allows stacking)
	camera.ShakeIntensity += intensity

	// Cap maximum shake intensity to prevent extreme values
	if camera.ShakeIntensity > 30.0 {
		camera.ShakeIntensity = 30.0
	}
}

// ShakeAdvanced triggers an advanced screen shake with duration control.
// Phase 10.3: Enhanced shake system with frequency and duration.
// Uses ScreenShakeComponent if available, falls back to basic shake.
// Respects accessibility settings.
func (s *CameraSystem) ShakeAdvanced(intensity, duration float64) {
	if s.activeCamera == nil {
		return
	}

	// Phase 10.3: Apply accessibility multiplier
	intensity = s.Accessibility.ApplyShakeIntensity(intensity)
	if intensity == 0.0 {
		return // Shake disabled via accessibility
	}

	// Try advanced shake component first
	shakeComp, ok := s.activeCamera.GetComponent("screenShake")
	if ok {
		advanced := shakeComp.(*ScreenShakeComponent)
		advanced.TriggerShake(intensity, duration)
		return
	}

	// Fall back to basic shake
	s.Shake(intensity)
}

// TriggerHitStop triggers a hit-stop effect on the active camera.
// Phase 10.3: Time dilation for impactful moments.
// Respects accessibility settings.
// duration: seconds to pause/slow (typical: 0.05-0.2s)
// timeScale: 0 = full stop, 0.1 = slow motion, 1.0 = normal
func (s *CameraSystem) TriggerHitStop(duration, timeScale float64) {
	if s.activeCamera == nil {
		return
	}

	// Phase 10.3: Check accessibility settings
	if !s.Accessibility.ShouldApplyHitStop() {
		return // Hit-stop disabled via accessibility
	}

	hitStopComp, ok := s.activeCamera.GetComponent("hitStop")
	if !ok {
		return // No hit-stop component, silently ignore
	}

	hitStop := hitStopComp.(*HitStopComponent)
	hitStop.TriggerHitStop(duration, timeScale)
}

// IsHitStopActive returns true if hit-stop is currently active.
// Phase 10.3: For systems that need to know if time is dilated.
func (s *CameraSystem) IsHitStopActive() bool {
	if s.activeCamera == nil {
		return false
	}

	hitStopComp, ok := s.activeCamera.GetComponent("hitStop")
	if !ok {
		return false
	}

	hitStop := hitStopComp.(*HitStopComponent)
	return hitStop.IsActive()
}

// GetTimeScale returns the current time scale (for hit-stop).
// Phase 10.3: Systems can query this to apply time dilation.
// Returns 1.0 when no hit-stop is active (normal time).
func (s *CameraSystem) GetTimeScale() float64 {
	if s.activeCamera == nil {
		return 1.0
	}

	hitStopComp, ok := s.activeCamera.GetComponent("hitStop")
	if !ok {
		return 1.0
	}

	hitStop := hitStopComp.(*HitStopComponent)
	return hitStop.GetTimeScale()
}
