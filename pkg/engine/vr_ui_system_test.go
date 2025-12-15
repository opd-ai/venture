package engine

import (
	"math"
	"testing"
)

func TestNewVRUISystem(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	if sys == nil {
		t.Fatal("NewVRUISystem returned nil")
	}

	if sys.IsEnabled() {
		t.Error("Expected disabled by default")
	}

	if sys.IsMenuOpen() {
		t.Error("Menu should be closed by default")
	}
}

func TestVRUISystem_SetEnabled(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	sys.SetEnabled(true)
	if !sys.IsEnabled() {
		t.Error("Expected enabled")
	}

	sys.SetEnabled(false)
	if sys.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestVRUISystem_SetHeadTracking(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	sys.SetHeadTracking(1, 2, 3, 0.5, 1.0)

	// Values are internal but we can verify no crash
	// In a real implementation, we would verify panel positions update correctly
}

func TestVRUISystem_SetHandTracking(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	sys.SetHandTracking(1, 2, 3, 4, 5, 6)

	// Values are internal but we can verify no crash
}

func TestVRUISystem_Update_Disabled(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)
	// System disabled

	entity := NewEntity(1)
	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	entity.AddComponent(ui)

	// Should not crash when disabled
	sys.Update([]*Entity{entity}, 0.016)
}

func TestVRUISystem_Update_Enabled(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)
	sys.SetEnabled(true)

	entity := NewEntity(1)
	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	ui.CreateDefaultPanels()
	entity.AddComponent(ui)

	// Should process without error
	sys.Update([]*Entity{entity}, 0.016)
}

func TestVRUISystem_OpenCloseMenu(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)
	sys.SetEnabled(true)

	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	ui.CreateDefaultPanels()

	openCalled := false
	closeCalled := false

	sys.SetMenuOpenCallback(func() {
		openCalled = true
	})

	sys.SetMenuCloseCallback(func() {
		closeCalled = true
	})

	// Open menu
	sys.OpenMenu(ui)
	if !sys.IsMenuOpen() {
		t.Error("Menu should be open")
	}
	if !openCalled {
		t.Error("Open callback not called")
	}
	if !ui.GetPanel("menu").Visible {
		t.Error("Menu panel should be visible")
	}

	// Close menu
	sys.CloseMenu(ui)
	if sys.IsMenuOpen() {
		t.Error("Menu should be closed")
	}
	if !closeCalled {
		t.Error("Close callback not called")
	}
	if ui.GetPanel("menu").Visible {
		t.Error("Menu panel should be hidden")
	}
}

func TestVRUISystem_ToggleMenu(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)
	sys.SetEnabled(true)

	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	ui.CreateDefaultPanels()

	// Toggle on
	sys.ToggleMenu(ui)
	if !sys.IsMenuOpen() {
		t.Error("Menu should be open after toggle")
	}

	// Toggle off
	sys.ToggleMenu(ui)
	if sys.IsMenuOpen() {
		t.Error("Menu should be closed after toggle")
	}
}

func TestVRUISystem_ToggleInventory(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	ui := NewVRUIComponent()
	ui.CreateDefaultPanels()

	// Initially hidden
	if ui.GetPanel("inventory").Visible {
		t.Error("Inventory should be hidden by default")
	}

	// Toggle on
	sys.ToggleInventory(ui)
	if !ui.GetPanel("inventory").Visible {
		t.Error("Inventory should be visible after toggle")
	}

	// Toggle off
	sys.ToggleInventory(ui)
	if ui.GetPanel("inventory").Visible {
		t.Error("Inventory should be hidden after toggle")
	}
}

func TestVRUISystem_ShowNotification(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	ui := NewVRUIComponent()
	ui.CreateDefaultPanels()

	sys.ShowNotification(ui, true)
	if !ui.GetPanel("notifications").Visible {
		t.Error("Notifications should be visible")
	}

	sys.ShowNotification(ui, false)
	if ui.GetPanel("notifications").Visible {
		t.Error("Notifications should be hidden")
	}
}

func TestVRUISystem_GazeActivation(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)
	sys.SetEnabled(true)

	activatedPanel := ""
	sys.SetGazeActivateCallback(func(panelID string) {
		activatedPanel = panelID
	})

	entity := NewEntity(1)
	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	ui.SetGazeActivationTime(0.1) // Fast activation for testing

	// Add a panel directly in front
	ui.AddPanel("test", PanelTypeMenu, 0, 0, 2.0, 1.0, 1.0)
	entity.AddComponent(ui)

	// Set head to look at panel
	sys.SetHeadTracking(0, 0, 0, 0, 0)

	// Multiple updates to accumulate gaze time
	for i := 0; i < 10; i++ {
		sys.Update([]*Entity{entity}, 0.02)
	}

	if activatedPanel != "test" {
		t.Errorf("Expected panel 'test' activated, got '%s'", activatedPanel)
	}
}

func TestSinApprox(t *testing.T) {
	// Test at key angles
	tests := []struct {
		angle    float64
		expected float64
		epsilon  float64
	}{
		{0, 0, 0.01},
		{math.Pi / 2, 1.0, 0.01},
		{math.Pi, 0, 0.02},
		{-math.Pi / 2, -1.0, 0.01},
	}

	for _, tt := range tests {
		result := sinApprox(tt.angle)
		if math.Abs(result-tt.expected) > tt.epsilon {
			t.Errorf("sinApprox(%v) = %v, want ~%v", tt.angle, result, tt.expected)
		}
	}
}

func TestCosApprox(t *testing.T) {
	tests := []struct {
		angle    float64
		expected float64
		epsilon  float64
	}{
		{0, 1.0, 0.01},
		{math.Pi / 2, 0, 0.02},
		{math.Pi, -1.0, 0.02},
	}

	for _, tt := range tests {
		result := cosApprox(tt.angle)
		if math.Abs(result-tt.expected) > tt.epsilon {
			t.Errorf("cosApprox(%v) = %v, want ~%v", tt.angle, result, tt.expected)
		}
	}
}

func TestVRUISystem_ThreadSafety(t *testing.T) {
	world := &World{}
	sys := NewVRUISystem(world)

	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetEnabled(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetHeadTracking(float64(i), 0, 0, 0, 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			sys.SetHandTracking(0, 0, 0, float64(i), 0, 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			_ = sys.IsMenuOpen()
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkVRUISystem_Update(b *testing.B) {
	world := &World{}
	sys := NewVRUISystem(world)
	sys.SetEnabled(true)

	entity := NewEntity(1)
	ui := NewVRUIComponent()
	ui.SetEnabled(true)
	ui.CreateDefaultPanels()
	entity.AddComponent(ui)

	entities := []*Entity{entity}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
