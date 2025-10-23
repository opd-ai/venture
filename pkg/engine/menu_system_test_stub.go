//go:build test
// +build test

// Package engine provides the menu system types for testing.
package engine

// MenuType represents the type of menu being displayed.
type MenuType int

const (
	MenuTypeNone MenuType = iota
	MenuTypeMain
	MenuTypeSave
	MenuTypeLoad
	MenuTypeConfirm
)

// MenuItem represents a single menu option.
type MenuItem struct {
	Label    string
	Action   func() error
	Enabled  bool
	Metadata interface{}
}

// MenuComponent stores menu state data.
type MenuComponent struct {
	Active         bool
	CurrentMenu    MenuType
	Items          []MenuItem
	SelectedIndex  int
	MenuStack      []MenuType
	ErrorMessage   string
	ErrorTimeout   float64
	ConfirmMessage string
	ConfirmAction  func() error
}

// Type returns the component type identifier.
func (m *MenuComponent) Type() string {
	return "menu"
}

// MenuSystem stub for testing (full implementation in menu_system.go).
type MenuSystem struct {
	world        *World
	screenWidth  int
	screenHeight int
	onSave       func(name string) error
	onLoad       func(name string) error
	menuEntity   *Entity
}

// NewMenuSystem creates a new menu system stub for testing.
func NewMenuSystem(world *World, screenWidth, screenHeight int, saveDir string) (*MenuSystem, error) {
	return &MenuSystem{
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
	}, nil
}

// SetSaveCallback sets the callback for save operations.
func (ms *MenuSystem) SetSaveCallback(callback func(name string) error) {
	ms.onSave = callback
}

// SetLoadCallback sets the callback for load operations.
func (ms *MenuSystem) SetLoadCallback(callback func(name string) error) {
	ms.onLoad = callback
}

// Toggle opens or closes the main menu.
func (ms *MenuSystem) Toggle() {
	if ms.menuEntity == nil {
		ms.menuEntity = ms.world.CreateEntity()
		menu := &MenuComponent{
			Active:      true,
			CurrentMenu: MenuTypeMain,
		}
		ms.menuEntity.AddComponent(menu)
		ms.world.Update(0)
	} else {
		if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
			menuComp := menu.(*MenuComponent)
			menuComp.Active = !menuComp.Active
		}
	}
}

// IsActive returns true if the menu is currently displayed.
func (ms *MenuSystem) IsActive() bool {
	if ms.menuEntity == nil {
		return false
	}
	if menu, ok := ms.menuEntity.GetComponent("menu"); ok {
		return menu.(*MenuComponent).Active
	}
	return false
}

// Update is a stub for testing.
func (ms *MenuSystem) Update(entities []*Entity, deltaTime float64) {
	// Stub for testing
}
