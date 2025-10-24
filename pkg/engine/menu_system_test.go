package engine

// StubMenuSystem is a test implementation of UISystem for menu testing.
type StubMenuSystem struct {
	UpdateCount int
	DrawCount   int
	active      bool
	world       *World
}

// NewStubMenuSystem creates a new stub menu system for testing.
func NewStubMenuSystem(world *World) *StubMenuSystem {
	return &StubMenuSystem{
		active: false,
		world:  world,
	}
}

// Update is called every frame but stub doesn't need to do anything.
func (s *StubMenuSystem) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

// Draw increments the draw counter. Implements UISystem interface.
func (s *StubMenuSystem) Draw(screen interface{}) {
	s.DrawCount++
}

// IsActive returns whether the menu is currently displayed.
// Implements UISystem interface.
func (s *StubMenuSystem) IsActive() bool {
	return s.active
}

// SetActive opens or closes the menu.
// Implements UISystem interface.
func (s *StubMenuSystem) SetActive(active bool) {
	s.active = active
}

// Toggle opens or closes the menu.
func (s *StubMenuSystem) Toggle() {
	s.active = !s.active
}

// Compile-time check that StubMenuSystem implements UISystem
var _ UISystem = (*StubMenuSystem)(nil)
