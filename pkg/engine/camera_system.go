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
}

// Type returns the component type identifier.
func (c *CameraComponent) Type() string {
	return "camera"
}

// NewCameraComponent creates a new camera component with default settings.
func NewCameraComponent() *CameraComponent {
	return &CameraComponent{
		OffsetX:   0,
		OffsetY:   0,
		Zoom:      1.0,
		MinX:      math.Inf(-1),
		MinY:      math.Inf(-1),
		MaxX:      math.Inf(1),
		MaxY:      math.Inf(1),
		Smoothing: 0.1,
		X:         0,
		Y:         0,
	}
}

// CameraSystem manages camera positioning and viewport.
type CameraSystem struct {
	// Screen dimensions
	ScreenWidth  int
	ScreenHeight int

	// Active camera entity (if any)
	activeCamera *Entity
}

// NewCameraSystem creates a new camera system.
func NewCameraSystem(screenWidth, screenHeight int) *CameraSystem {
	return &CameraSystem{
		ScreenWidth:  screenWidth,
		ScreenHeight: screenHeight,
	}
}

// Update updates camera positions to follow their target entities.
func (s *CameraSystem) Update(entities []*Entity, deltaTime float64) {
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

		// Apply smoothing (lerp)
		if camera.Smoothing > 0 {
			smoothFactor := 1.0 - math.Pow(camera.Smoothing, deltaTime*60) // 60 fps normalized
			camera.X += (targetX - camera.X) * smoothFactor
			camera.Y += (targetY - camera.Y) * smoothFactor
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
	}
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
