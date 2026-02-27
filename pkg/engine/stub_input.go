package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// StubInput is a test input component without Ebiten dependencies.
// Implements InputProvider interface for testing.
// This stub allows tests to simulate input without requiring Ebiten runtime.
type StubInput struct {
	// Movement input
	MoveX, MoveY float64

	// Action buttons
	ActionPressed      bool
	UseItemPressed     bool
	ActionJustPressed  bool
	UseItemJustPressed bool
	AnyKeyPressed      bool // GAP-005: Track any key press

	// Spell casting
	Spell1Pressed bool
	Spell2Pressed bool
	Spell3Pressed bool
	Spell4Pressed bool
	Spell5Pressed bool

	// Mouse state
	MouseX, MouseY int
	MousePressed   bool

	// Mouse delta (Gap #8 fix)
	MouseDeltaX, MouseDeltaY int

	// Menu navigation input (for UI abstraction)
	MenuUpJustPressed      bool
	MenuDownJustPressed    bool
	MenuConfirmJustPressed bool
	MenuBackJustPressed    bool
	MenuTabJustPressed     bool

	// Touch input (for mobile/WASM testing)
	// Map of touch ID to position
	TouchPositions map[ebiten.TouchID]struct{ X, Y int }
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

// IsAnyKeyPressed implements InputProvider interface.
func (i *StubInput) IsAnyKeyPressed() bool {
	return i.AnyKeyPressed
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

// GetMouseDelta implements InputProvider interface.
func (i *StubInput) GetMouseDelta() (dx, dy int) {
	return i.MouseDeltaX, i.MouseDeltaY
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

// IsMenuUpJustPressed implements InputProvider interface.
func (i *StubInput) IsMenuUpJustPressed() bool {
	return i.MenuUpJustPressed
}

// IsMenuDownJustPressed implements InputProvider interface.
func (i *StubInput) IsMenuDownJustPressed() bool {
	return i.MenuDownJustPressed
}

// IsMenuConfirmJustPressed implements InputProvider interface.
func (i *StubInput) IsMenuConfirmJustPressed() bool {
	return i.MenuConfirmJustPressed
}

// IsMenuBackJustPressed implements InputProvider interface.
func (i *StubInput) IsMenuBackJustPressed() bool {
	return i.MenuBackJustPressed
}

// IsMenuTabJustPressed implements InputProvider interface.
func (i *StubInput) IsMenuTabJustPressed() bool {
	return i.MenuTabJustPressed
}

// GetTouchIDs implements InputProvider interface.
// Returns the list of current active touch IDs for testing.
func (i *StubInput) GetTouchIDs() []ebiten.TouchID {
	if i.TouchPositions == nil {
		return nil
	}
	ids := make([]ebiten.TouchID, 0, len(i.TouchPositions))
	for id := range i.TouchPositions {
		ids = append(ids, id)
	}
	return ids
}

// GetTouchPosition implements InputProvider interface.
// Returns the screen position of the touch with the given ID for testing.
func (i *StubInput) GetTouchPosition(id ebiten.TouchID) (x, y int) {
	if i.TouchPositions == nil {
		return 0, 0
	}
	if pos, ok := i.TouchPositions[id]; ok {
		return pos.X, pos.Y
	}
	return 0, 0
}

// NewStubInput creates a new test input component.
func NewStubInput() *StubInput {
	return &StubInput{}
}

// Compile-time interface check
var _ InputProvider = (*StubInput)(nil)
