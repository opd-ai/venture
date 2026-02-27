package prestige

import "github.com/opd-ai/venture/pkg/engine"

// mockInputProvider implements engine.InputProvider for testing prestige UI.
// It maps conceptual actions (menu up/down/confirm/back) to the interface methods.
type mockInputProvider struct {
	menuUp      bool
	menuDown    bool
	menuConfirm bool
	menuBack    bool
	moveX       float64
	moveY       float64
}

// newMockInputProvider creates a new mock input provider for testing.
func newMockInputProvider() *mockInputProvider {
	return &mockInputProvider{}
}

// setMenuUp simulates pressing the menu up key (Up arrow / W).
func (m *mockInputProvider) setMenuUp() {
	m.menuUp = true
}

// setMenuDown simulates pressing the menu down key (Down arrow / S).
func (m *mockInputProvider) setMenuDown() {
	m.menuDown = true
}

// setMenuConfirm simulates pressing the confirm key (Enter / Space).
func (m *mockInputProvider) setMenuConfirm() {
	m.menuConfirm = true
}

// setMenuBack simulates pressing the back key (Escape).
func (m *mockInputProvider) setMenuBack() {
	m.menuBack = true
}

// clearKeys resets all key states.
func (m *mockInputProvider) clearKeys() {
	m.menuUp = false
	m.menuDown = false
	m.menuConfirm = false
	m.menuBack = false
}

// --- engine.InputProvider interface implementation ---

func (m *mockInputProvider) Type() string                    { return "input" }
func (m *mockInputProvider) GetMovement() (float64, float64) { return m.moveX, m.moveY }
func (m *mockInputProvider) IsActionPressed() bool           { return false }
func (m *mockInputProvider) IsActionJustPressed() bool       { return m.menuConfirm }
func (m *mockInputProvider) IsAnyKeyPressed() bool {
	return m.menuUp || m.menuDown || m.menuConfirm || m.menuBack
}
func (m *mockInputProvider) IsUseItemPressed() bool         { return false }
func (m *mockInputProvider) IsUseItemJustPressed() bool     { return false }
func (m *mockInputProvider) IsSpellPressed(slot int) bool   { return false }
func (m *mockInputProvider) GetMousePosition() (int, int)   { return 0, 0 }
func (m *mockInputProvider) GetMouseDelta() (int, int)      { return 0, 0 }
func (m *mockInputProvider) IsMousePressed() bool           { return false }
func (m *mockInputProvider) SetMovement(x, y float64)       { m.moveX, m.moveY = x, y }
func (m *mockInputProvider) SetActionPressed(pressed bool)  {}
func (m *mockInputProvider) IsMenuUpJustPressed() bool      { return m.menuUp }
func (m *mockInputProvider) IsMenuDownJustPressed() bool    { return m.menuDown }
func (m *mockInputProvider) IsMenuConfirmJustPressed() bool { return m.menuConfirm }
func (m *mockInputProvider) IsMenuBackJustPressed() bool    { return m.menuBack }
func (m *mockInputProvider) IsMenuTabJustPressed() bool     { return false }

// Compile-time interface check
var _ engine.InputProvider = (*mockInputProvider)(nil)
