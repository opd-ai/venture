package engine

// StubInput is a test input component without Ebiten dependencies.
// Implements InputProvider interface for testing.
type StubInput struct {
	// Movement input
	MoveX, MoveY float64

	// Action buttons
	ActionPressed      bool
	UseItemPressed     bool
	ActionJustPressed  bool
	UseItemJustPressed bool

	// Spell casting
	Spell1Pressed bool
	Spell2Pressed bool
	Spell3Pressed bool
	Spell4Pressed bool
	Spell5Pressed bool

	// Mouse state
	MouseX, MouseY int
	MousePressed   bool
}

// Type implements Component interface.
func (i *StubInput) Type() string {
	return "input"
}

// GetMovement implements InputProvider interface.
func (i *StubInput) GetMovement() (x, y float64) {
	return i.MoveX, i.MoveY
}

// IsActionPressed implements InputProvider interface.
func (i *StubInput) IsActionPressed() bool {
	return i.ActionPressed
}

// IsActionJustPressed implements InputProvider interface.
func (i *StubInput) IsActionJustPressed() bool {
	return i.ActionJustPressed
}

// IsUseItemPressed implements InputProvider interface.
func (i *StubInput) IsUseItemPressed() bool {
	return i.UseItemPressed
}

// IsUseItemJustPressed implements InputProvider interface.
func (i *StubInput) IsUseItemJustPressed() bool {
	return i.UseItemJustPressed
}

// IsSpellPressed implements InputProvider interface.
func (i *StubInput) IsSpellPressed(slot int) bool {
	switch slot {
	case 1:
		return i.Spell1Pressed
	case 2:
		return i.Spell2Pressed
	case 3:
		return i.Spell3Pressed
	case 4:
		return i.Spell4Pressed
	case 5:
		return i.Spell5Pressed
	default:
		return false
	}
}

// GetMousePosition implements InputProvider interface.
func (i *StubInput) GetMousePosition() (x, y int) {
	return i.MouseX, i.MouseY
}

// IsMousePressed implements InputProvider interface.
func (i *StubInput) IsMousePressed() bool {
	return i.MousePressed
}

// SetMovement implements InputProvider interface.
func (i *StubInput) SetMovement(x, y float64) {
	i.MoveX, i.MoveY = x, y
}

// SetActionPressed implements InputProvider interface.
func (i *StubInput) SetActionPressed(pressed bool) {
	i.ActionPressed = pressed
}

// NewStubInput creates a new test input component.
func NewStubInput() *StubInput {
	return &StubInput{}
}

// Compile-time interface check
var _ InputProvider = (*StubInput)(nil)
