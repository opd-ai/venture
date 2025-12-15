package engine

import (
	"encoding/json"
	"testing"
)

func TestNewVRUIComponent(t *testing.T) {
	c := NewVRUIComponent()

	if c == nil {
		t.Fatal("NewVRUIComponent returned nil")
	}

	if c.IsEnabled() {
		t.Error("Expected disabled by default")
	}

	if len(c.Panels) != 0 {
		t.Error("Expected no panels by default")
	}

	if c.GazeActivationTime != DefaultGazeActivationTime {
		t.Errorf("Expected gaze time %v, got %v", DefaultGazeActivationTime, c.GazeActivationTime)
	}

	if !c.IsComfortVignetteEnabled() {
		t.Error("Expected comfort vignette enabled by default")
	}

	if c.GetUIScale() != DefaultUIScale {
		t.Errorf("Expected UI scale %v, got %v", DefaultUIScale, c.GetUIScale())
	}
}

func TestVRUIComponent_Type(t *testing.T) {
	c := NewVRUIComponent()
	if c.Type() != "vr_ui" {
		t.Errorf("Expected type 'vr_ui', got '%s'", c.Type())
	}
}

func TestVRUIComponent_SetEnabled(t *testing.T) {
	c := NewVRUIComponent()

	c.SetEnabled(true)
	if !c.IsEnabled() {
		t.Error("Expected enabled")
	}

	c.SetEnabled(false)
	if c.IsEnabled() {
		t.Error("Expected disabled")
	}
}

func TestVRUIComponent_AddPanel(t *testing.T) {
	c := NewVRUIComponent()

	panel := c.AddPanel("test", PanelTypeHealth, 1.0, 2.0, 3.0, 0.5, 0.3)

	if panel == nil {
		t.Fatal("AddPanel returned nil")
	}

	if panel.ID != "test" {
		t.Errorf("Expected ID 'test', got '%s'", panel.ID)
	}

	if panel.PanelType != PanelTypeHealth {
		t.Errorf("Expected type %s, got %s", PanelTypeHealth, panel.PanelType)
	}

	if panel.WorldX != 1.0 || panel.WorldY != 2.0 || panel.WorldZ != 3.0 {
		t.Error("Position mismatch")
	}

	if panel.Width != 0.5 || panel.Height != 0.3 {
		t.Error("Size mismatch")
	}

	if !panel.Visible {
		t.Error("Expected visible by default")
	}

	if !panel.Interactive {
		t.Error("Expected interactive by default")
	}

	// Verify it's in the panels map
	if c.GetPanel("test") == nil {
		t.Error("Panel not found in map")
	}
}

func TestVRUIComponent_RemovePanel(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)

	c.RemovePanel("test")

	if c.GetPanel("test") != nil {
		t.Error("Panel should be removed")
	}
}

func TestVRUIComponent_GetAllPanels(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("p1", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)
	c.AddPanel("p2", PanelTypeMinimap, 0, 0, 0, 0.5, 0.3)
	c.AddPanel("p3", PanelTypeInventory, 0, 0, 0, 0.5, 0.3)

	panels := c.GetAllPanels()
	if len(panels) != 3 {
		t.Errorf("Expected 3 panels, got %d", len(panels))
	}
}

func TestVRUIComponent_SetPanelPosition(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)

	c.SetPanelPosition("test", 5.0, 6.0, 7.0)

	panel := c.GetPanel("test")
	if panel.WorldX != 5.0 || panel.WorldY != 6.0 || panel.WorldZ != 7.0 {
		t.Error("Position not updated")
	}
}

func TestVRUIComponent_SetPanelVisible(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)

	c.SetPanelVisible("test", false)
	if c.GetPanel("test").Visible {
		t.Error("Panel should be hidden")
	}

	c.SetPanelVisible("test", true)
	if !c.GetPanel("test").Visible {
		t.Error("Panel should be visible")
	}
}

func TestVRUIComponent_SetPanelFollowHead(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)

	c.SetPanelFollowHead("test", true, 3.0)

	panel := c.GetPanel("test")
	if !panel.FollowHead {
		t.Error("FollowHead should be true")
	}
	if panel.FollowDistance != 3.0 {
		t.Errorf("Expected follow distance 3.0, got %v", panel.FollowDistance)
	}
}

func TestVRUIComponent_SetPanelLockedToHand(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeInventory, 0, 0, 0, 0.5, 0.3)

	c.SetPanelLockedToHand("test", ControllerLeft)

	panel := c.GetPanel("test")
	if panel.LockedToHand != ControllerLeft {
		t.Error("Panel should be locked to left hand")
	}
}

func TestVRUIComponent_GazeTarget(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("p1", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)
	c.AddPanel("p2", PanelTypeMinimap, 0, 0, 0, 0.5, 0.3)

	c.SetGazeTarget("p1")
	if c.GetGazeTarget() != "p1" {
		t.Error("Gaze target should be p1")
	}

	// Panel should be highlighted
	if !c.GetPanel("p1").GazeHighlighted {
		t.Error("p1 should be highlighted")
	}

	// Switch target
	c.SetGazeTarget("p2")
	if c.GetGazeTarget() != "p2" {
		t.Error("Gaze target should be p2")
	}

	// Previous should be unhighlighted
	if c.GetPanel("p1").GazeHighlighted {
		t.Error("p1 should not be highlighted")
	}

	// New should be highlighted
	if !c.GetPanel("p2").GazeHighlighted {
		t.Error("p2 should be highlighted")
	}
}

func TestVRUIComponent_UpdateGazeHover(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)
	c.SetGazeActivationTime(1.0) // 1 second activation

	// No target - should return false
	if c.UpdateGazeHover(0.5) {
		t.Error("Should return false with no target")
	}

	// Set target
	c.SetGazeTarget("test")

	// Partial hover
	if c.UpdateGazeHover(0.5) {
		t.Error("Should return false before activation time")
	}

	// Check progress
	progress := c.GetGazeProgress()
	if progress < 0.49 || progress > 0.51 {
		t.Errorf("Expected progress ~0.5, got %v", progress)
	}

	// Complete hover
	if !c.UpdateGazeHover(0.6) {
		t.Error("Should return true after activation time reached")
	}

	// Progress should reset
	progress = c.GetGazeProgress()
	if progress > 0.1 {
		t.Errorf("Expected progress reset, got %v", progress)
	}
}

func TestVRUIComponent_ComfortVignette(t *testing.T) {
	c := NewVRUIComponent()

	// Test disable
	c.SetComfortVignetteEnabled(false)
	c.UpdateComfortVignette(true, 0.1)
	if c.GetComfortVignetteIntensity() != 0 {
		t.Error("Vignette should be 0 when disabled")
	}

	// Test enable with movement
	c.SetComfortVignetteEnabled(true)
	c.SetComfortVignetteIntensity(0)

	// Simulate movement over time
	for i := 0; i < 10; i++ {
		c.UpdateComfortVignette(true, 0.016)
	}

	if c.GetComfortVignetteIntensity() <= 0 {
		t.Error("Vignette should increase when moving")
	}

	// Stop moving
	for i := 0; i < 20; i++ {
		c.UpdateComfortVignette(false, 0.016)
	}

	if c.GetComfortVignetteIntensity() >= 0.3 {
		t.Error("Vignette should decrease when not moving")
	}
}

func TestVRUIComponent_SetUIScale(t *testing.T) {
	c := NewVRUIComponent()

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"normal", 1.5, 1.5},
		{"too low clamp", 0.2, 0.5},
		{"too high clamp", 3.0, 2.0},
		{"at min", 0.5, 0.5},
		{"at max", 2.0, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.SetUIScale(tt.input)
			if c.GetUIScale() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, c.GetUIScale())
			}
		})
	}
}

func TestVRUIComponent_CalculateGazeRayIntersection(t *testing.T) {
	c := NewVRUIComponent()
	c.SetUIScale(1.0)

	// Add a panel at z=2, centered at origin, size 1x1
	c.AddPanel("test", PanelTypeMenu, 0, 0, 2.0, 1.0, 1.0)

	// Gaze directly at center - should hit
	hit := c.CalculateGazeRayIntersection(0, 0, 0, 0, 0, 1)
	if hit != "test" {
		t.Errorf("Expected hit 'test', got '%s'", hit)
	}

	// Gaze to the side - should miss
	hit = c.CalculateGazeRayIntersection(0, 0, 0, 1, 0, 1) // 45 degrees right
	if hit == "test" {
		t.Error("Should miss when gazing to the side")
	}

	// Gaze backward - should miss (negative Z)
	hit = c.CalculateGazeRayIntersection(0, 0, 0, 0, 0, -1)
	if hit == "test" {
		t.Error("Should miss when gazing backward")
	}
}

func TestVRUIComponent_CalculateGazeRayIntersection_MultipleHits(t *testing.T) {
	c := NewVRUIComponent()
	c.SetUIScale(1.0)

	// Two panels at different distances
	c.AddPanel("near", PanelTypeMenu, 0, 0, 1.0, 1.0, 1.0)
	c.AddPanel("far", PanelTypeMenu, 0, 0, 3.0, 2.0, 2.0)

	// Gaze at center - should hit nearest
	hit := c.CalculateGazeRayIntersection(0, 0, 0, 0, 0, 1)
	if hit != "near" {
		t.Errorf("Expected hit 'near', got '%s'", hit)
	}
}

func TestVRUIComponent_CalculateGazeRayIntersection_Hidden(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeMenu, 0, 0, 2.0, 1.0, 1.0)
	c.SetPanelVisible("test", false)

	// Should not hit hidden panels
	hit := c.CalculateGazeRayIntersection(0, 0, 0, 0, 0, 1)
	if hit != "" {
		t.Error("Should not hit hidden panels")
	}
}

func TestVRUIComponent_CreateDefaultPanels(t *testing.T) {
	c := NewVRUIComponent()
	c.CreateDefaultPanels()

	// Should have standard panels
	if c.GetPanel("health") == nil {
		t.Error("Missing health panel")
	}
	if c.GetPanel("minimap") == nil {
		t.Error("Missing minimap panel")
	}
	if c.GetPanel("inventory") == nil {
		t.Error("Missing inventory panel")
	}
	if c.GetPanel("menu") == nil {
		t.Error("Missing menu panel")
	}
	if c.GetPanel("quest") == nil {
		t.Error("Missing quest panel")
	}
	if c.GetPanel("notifications") == nil {
		t.Error("Missing notifications panel")
	}

	// Check some settings
	health := c.GetPanel("health")
	if !health.FollowHead {
		t.Error("Health should follow head")
	}

	inv := c.GetPanel("inventory")
	if inv.LockedToHand != ControllerLeft {
		t.Error("Inventory should be locked to left hand")
	}
	if inv.Visible {
		t.Error("Inventory should be hidden by default")
	}
}

func TestVRUIComponent_Serialize(t *testing.T) {
	c := NewVRUIComponent()
	c.SetEnabled(true)
	c.AddPanel("test", PanelTypeHealth, 1, 2, 3, 0.5, 0.3)
	c.SetUIScale(1.5)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	if m["enabled"] != true {
		t.Error("Expected enabled true")
	}
}

func TestVRUIComponent_Deserialize(t *testing.T) {
	original := NewVRUIComponent()
	original.SetEnabled(true)
	original.AddPanel("test", PanelTypeHealth, 1, 2, 3, 0.5, 0.3)
	original.SetUIScale(1.5)
	original.SetGazeActivationTime(2.0)

	data, _ := original.Serialize()

	restored := NewVRUIComponent()
	if err := restored.Deserialize(data); err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.IsEnabled() != original.IsEnabled() {
		t.Error("Enabled mismatch")
	}
	if restored.GetUIScale() != original.GetUIScale() {
		t.Error("UIScale mismatch")
	}
	if restored.GetPanel("test") == nil {
		t.Error("Panel not restored")
	}
}

func TestVRUIComponent_ThreadSafety(t *testing.T) {
	c := NewVRUIComponent()
	c.AddPanel("test", PanelTypeHealth, 0, 0, 0, 0.5, 0.3)

	done := make(chan bool, 4)

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetEnabled(i%2 == 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetPanelPosition("test", float64(i), 0, 0)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.SetGazeTarget("test")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			c.UpdateComfortVignette(i%2 == 0, 0.016)
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}

func BenchmarkVRUIComponent_CalculateGazeRayIntersection(b *testing.B) {
	c := NewVRUIComponent()
	for i := 0; i < 10; i++ {
		c.AddPanel("panel_"+string(rune('a'+i)), PanelTypeMenu, float64(i), 0, 2.0, 1.0, 1.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.CalculateGazeRayIntersection(0, 0, 0, 0, 0, 1)
	}
}
