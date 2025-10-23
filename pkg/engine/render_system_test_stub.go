//go:build test
// +build test

// Package engine provides test stubs for RenderSystem.
package engine

// RenderSystem handles rendering of entities (test stub).
type RenderSystem struct {
	cameraSystem  *CameraSystem
	ShowColliders bool
	ShowGrid      bool
}

// NewRenderSystem creates a new render system (test stub).
func NewRenderSystem(cameraSystem *CameraSystem) *RenderSystem {
	return &RenderSystem{
		cameraSystem:  cameraSystem,
		ShowColliders: false,
		ShowGrid:      false,
	}
}

// Update is called every frame (test stub).
func (r *RenderSystem) Update(entities []*Entity, deltaTime float64) {
	// Stub - no op in tests
}

// Draw renders all visible entities (test stub).
func (r *RenderSystem) Draw(screen interface{}, entities []*Entity) {
	// Stub - no op in tests
}
