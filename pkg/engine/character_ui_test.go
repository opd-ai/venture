package engine

// StubCharacterUI is a test implementation of UISystem for character UI testing.
type StubCharacterUI struct {
UpdateCount int
DrawCount   int
active      bool
world       *World
}

// NewStubCharacterUI creates a new stub character UI for testing.
func NewStubCharacterUI(world *World) *StubCharacterUI {
return &StubCharacterUI{
active: false,
world:  world,
}
}

// SetPlayerEntity sets the player entity (no-op for testing).
func (s *StubCharacterUI) SetPlayerEntity(entity *Entity) {
// No-op for testing
}

// Update increments the update counter. Implements System interface.
func (s *StubCharacterUI) Update(entities []*Entity, deltaTime float64) {
s.UpdateCount++
}

// Draw increments the draw counter. Implements UISystem interface.
func (s *StubCharacterUI) Draw(screen interface{}) {
s.DrawCount++
}

// IsActive returns whether the character UI is currently visible.
// Implements UISystem interface.
func (s *StubCharacterUI) IsActive() bool {
return s.active
}

// SetActive sets whether the character UI is visible.
// Implements UISystem interface.
func (s *StubCharacterUI) SetActive(active bool) {
s.active = active
}

// Toggle toggles character UI visibility.
func (s *StubCharacterUI) Toggle() {
s.active = !s.active
}

// Compile-time check that StubCharacterUI implements UISystem
var _ UISystem = (*StubCharacterUI)(nil)
