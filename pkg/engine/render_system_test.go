package engine

// StubRenderSystem is a test implementation of RenderingSystem interface.
// It provides a simple mock for unit testing without actual rendering.
type StubRenderSystem struct {
	cameraSystem  *CameraSystem
	ShowColliders bool
	ShowGrid      bool

	// Test state tracking
	UpdateCount          int
	DrawCount            int
	LastDrawnEntityCount int
}

// NewStubRenderSystem creates a new test render system.
func NewStubRenderSystem(cameraSystem *CameraSystem) *StubRenderSystem {
	return &StubRenderSystem{
		cameraSystem:  cameraSystem,
		ShowColliders: false,
		ShowGrid:      false,
		UpdateCount:   0,
		DrawCount:     0,
	}
}

// Update implements System interface (test stub).
func (r *StubRenderSystem) Update(entities []*Entity, deltaTime float64) {
	r.UpdateCount++
}

// Draw implements RenderingSystem interface (test stub).
func (r *StubRenderSystem) Draw(screen interface{}, entities []*Entity) {
	r.DrawCount++
	r.LastDrawnEntityCount = len(entities)
}

// SetShowColliders implements RenderingSystem interface.
func (r *StubRenderSystem) SetShowColliders(show bool) {
	r.ShowColliders = show
}

// SetShowGrid implements RenderingSystem interface.
func (r *StubRenderSystem) SetShowGrid(show bool) {
	r.ShowGrid = show
}

// Compile-time interface check
var _ RenderingSystem = (*StubRenderSystem)(nil)
