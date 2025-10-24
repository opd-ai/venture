package engine

// StubHUDSystem is a test implementation of UISystem for HUD testing.
type StubHUDSystem struct {
	UpdateCount  int
	DrawCount    int
	active       bool
	playerEntity *Entity
}

// NewStubHUDSystem creates a new stub HUD system for testing.
func NewStubHUDSystem() *StubHUDSystem {
	return &StubHUDSystem{
		active: true,
	}
}

// SetPlayerEntity sets the player entity (no-op for testing).
func (s *StubHUDSystem) SetPlayerEntity(entity *Entity) {
	s.playerEntity = entity
}

// Update is called every frame but stub doesn't need to do anything.
func (s *StubHUDSystem) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

// Draw increments the draw counter. Implements UISystem interface.
func (s *StubHUDSystem) Draw(screen interface{}) {
	s.DrawCount++
}

// IsActive returns whether the HUD is currently visible.
// Implements UISystem interface.
func (s *StubHUDSystem) IsActive() bool {
	return s.active
}

// SetActive sets whether the HUD is visible.
// Implements UISystem interface.
func (s *StubHUDSystem) SetActive(active bool) {
	s.active = active
}

// Compile-time check that StubHUDSystem implements UISystem
var _ UISystem = (*StubHUDSystem)(nil)
