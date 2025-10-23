//go:build test
// +build test

package engine

import (
	"testing"
)

func TestMenuComponent(t *testing.T) {
	menu := &MenuComponent{
		Active:      true,
		CurrentMenu: MenuTypeMain,
		Items: []MenuItem{
			{Label: "Save", Enabled: true},
			{Label: "Load", Enabled: true},
			{Label: "Exit", Enabled: false},
		},
		SelectedIndex: 0,
	}

	if menu.Type() != "menu" {
		t.Errorf("MenuComponent.Type() = %s, want menu", menu.Type())
	}

	if len(menu.Items) != 3 {
		t.Errorf("len(menu.Items) = %d, want 3", len(menu.Items))
	}
}

func TestMenuSystemCreation(t *testing.T) {
	world := NewWorld()
	ms, err := NewMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewMenuSystem() error = %v, want nil", err)
	}

	if ms.screenWidth != 800 {
		t.Errorf("ms.screenWidth = %d, want 800", ms.screenWidth)
	}

	if ms.screenHeight != 600 {
		t.Errorf("ms.screenHeight = %d, want 600", ms.screenHeight)
	}
}

func TestMenuSystemToggle(t *testing.T) {
	world := NewWorld()
	ms, err := NewMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewMenuSystem() error = %v", err)
	}

	// Initially inactive
	if ms.IsActive() {
		t.Error("Menu should be inactive initially")
	}

	// Toggle to open
	ms.Toggle()
	if !ms.IsActive() {
		t.Error("Menu should be active after Toggle()")
	}

	// Toggle to close
	ms.Toggle()
	if ms.IsActive() {
		t.Error("Menu should be inactive after second Toggle()")
	}
}

func TestMenuSystemCallbacks(t *testing.T) {
	world := NewWorld()
	ms, err := NewMenuSystem(world, 800, 600, "./test_saves")
	if err != nil {
		t.Fatalf("NewMenuSystem() error = %v", err)
	}

	ms.SetSaveCallback(func(name string) error {
		return nil
	})

	ms.SetLoadCallback(func(name string) error {
		return nil
	})

	// Verify callbacks are set (can't easily test execution without full input simulation)
	if ms.onSave == nil {
		t.Error("SaveCallback not set")
	}

	if ms.onLoad == nil {
		t.Error("LoadCallback not set")
	}
}

func TestMenuItemAction(t *testing.T) {
	actionCalled := false
	item := MenuItem{
		Label:   "Test",
		Enabled: true,
		Action: func() error {
			actionCalled = true
			return nil
		},
	}

	if err := item.Action(); err != nil {
		t.Errorf("Action() error = %v, want nil", err)
	}

	if !actionCalled {
		t.Error("Action callback not called")
	}
}

func TestMenuTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		menuType MenuType
		want     int
	}{
		{"MenuTypeNone", MenuTypeNone, 0},
		{"MenuTypeMain", MenuTypeMain, 1},
		{"MenuTypeSave", MenuTypeSave, 2},
		{"MenuTypeLoad", MenuTypeLoad, 3},
		{"MenuTypeConfirm", MenuTypeConfirm, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if int(tt.menuType) != tt.want {
				t.Errorf("MenuType %s = %d, want %d", tt.name, tt.menuType, tt.want)
			}
		})
	}
}
