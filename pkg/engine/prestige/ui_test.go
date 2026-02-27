package prestige

import (
	"testing"
)

func TestPrestigeMenuOption_String(t *testing.T) {
	tests := []struct {
		option   PrestigeMenuOption
		expected string
	}{
		{PrestigeOptionHealth, "Health"},
		{PrestigeOptionDamage, "Damage"},
		{PrestigeOptionDefense, "Defense"},
		{PrestigeOptionSpeed, "Speed"},
		{PrestigeOptionCritical, "Critical"},
		{PrestigeOptionRespec, "Respec All"},
		{PrestigeOptionBack, "Back"},
		{PrestigeMenuOption(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.option.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPrestigeMenuOption_toParagonStat(t *testing.T) {
	tests := []struct {
		option   PrestigeMenuOption
		expected ParagonStat
	}{
		{PrestigeOptionHealth, StatHealth},
		{PrestigeOptionDamage, StatDamage},
		{PrestigeOptionDefense, StatDefense},
		{PrestigeOptionSpeed, StatSpeed},
		{PrestigeOptionCritical, StatCritical},
		{PrestigeOptionRespec, ParagonStat(-1)},
		{PrestigeOptionBack, ParagonStat(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.option.String(), func(t *testing.T) {
			if got := tt.option.toParagonStat(); got != tt.expected {
				t.Errorf("toParagonStat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewPrestigeUI(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)

	if ui == nil {
		t.Fatal("NewPrestigeUI returned nil")
	}
	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}
	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}
	if ui.manager != manager {
		t.Error("manager not set correctly")
	}
	if ui.visible {
		t.Error("UI should not be visible by default")
	}
	if len(ui.options) != 7 {
		t.Errorf("options count = %d, want 7", len(ui.options))
	}
	if len(ui.allocButtons) != 5 {
		t.Errorf("allocButtons count = %d, want 5", len(ui.allocButtons))
	}
}

func TestPrestigeUI_ShowHide(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)

	if ui.IsVisible() {
		t.Error("UI should not be visible initially")
	}

	ui.Show("player1", "Warrior")
	if !ui.IsVisible() {
		t.Error("UI should be visible after Show()")
	}
	if ui.playerID != "player1" {
		t.Errorf("playerID = %s, want player1", ui.playerID)
	}
	if ui.className != "Warrior" {
		t.Errorf("className = %s, want Warrior", ui.className)
	}
	if ui.selectedIdx != 0 {
		t.Errorf("selectedIdx = %d, want 0", ui.selectedIdx)
	}

	ui.Hide()
	if ui.IsVisible() {
		t.Error("UI should not be visible after Hide()")
	}
}

func TestPrestigeUI_SetCallbacks(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)

	backCalled := false
	ui.SetBackCallback(func() { backCalled = true })

	respecCalled := false
	ui.SetRespecCallback(func(cost int) bool {
		respecCalled = true
		return true
	})

	// Test that callbacks are set
	if ui.onBack == nil {
		t.Error("onBack callback not set")
	}
	if ui.onRespecConfirm == nil {
		t.Error("onRespecConfirm callback not set")
	}

	// Verify callbacks fire
	ui.onBack()
	if !backCalled {
		t.Error("onBack callback not called")
	}
	ui.onRespecConfirm(1000)
	if !respecCalled {
		t.Error("onRespecConfirm callback not called")
	}
}

func TestPrestigeUI_Navigation(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)
	ui.Show("player1", "Warrior")

	// Initial position
	if ui.selectedIdx != 0 {
		t.Errorf("initial selectedIdx = %d, want 0", ui.selectedIdx)
	}

	// Navigate down
	mock.setMenuDown()
	ui.Update()
	if ui.selectedIdx != 1 {
		t.Errorf("after down, selectedIdx = %d, want 1", ui.selectedIdx)
	}
	mock.clearKeys()

	// Navigate down with S key
	mock.setMenuDown()
	ui.Update()
	if ui.selectedIdx != 2 {
		t.Errorf("after S, selectedIdx = %d, want 2", ui.selectedIdx)
	}
	mock.clearKeys()

	// Navigate up
	mock.setMenuUp()
	ui.Update()
	if ui.selectedIdx != 1 {
		t.Errorf("after up, selectedIdx = %d, want 1", ui.selectedIdx)
	}
	mock.clearKeys()

	// Navigate up with W key
	mock.setMenuUp()
	ui.Update()
	if ui.selectedIdx != 0 {
		t.Errorf("after W, selectedIdx = %d, want 0", ui.selectedIdx)
	}
	mock.clearKeys()

	// Wrap around at top
	mock.setMenuUp()
	ui.Update()
	if ui.selectedIdx != len(ui.options)-1 {
		t.Errorf("after wrap up, selectedIdx = %d, want %d", ui.selectedIdx, len(ui.options)-1)
	}
	mock.clearKeys()

	// Wrap around at bottom
	mock.setMenuDown()
	ui.Update()
	if ui.selectedIdx != 0 {
		t.Errorf("after wrap down, selectedIdx = %d, want 0", ui.selectedIdx)
	}
}

func TestPrestigeUI_AllocatePoint(t *testing.T) {
	manager := NewManager()
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 5)

	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)
	ui.Show("player1", "Warrior")

	// Select Health (index 0) and allocate
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	prestige := manager.GetPlayerPrestige("player1")
	if prestige.ParagonAllocations[StatHealth] != 1 {
		t.Errorf("Health allocation = %d, want 1", prestige.ParagonAllocations[StatHealth])
	}
	if prestige.ParagonPoints != 4 {
		t.Errorf("ParagonPoints = %d, want 4", prestige.ParagonPoints)
	}

	// Navigate to damage and allocate with Space
	mock.setMenuDown()
	ui.Update()
	mock.clearKeys()

	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	prestige = manager.GetPlayerPrestige("player1")
	if prestige.ParagonAllocations[StatDamage] != 1 {
		t.Errorf("Damage allocation = %d, want 1", prestige.ParagonAllocations[StatDamage])
	}
}

func TestPrestigeUI_Respec(t *testing.T) {
	manager := NewManager()
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 3)

	// Allocate some points
	manager.AllocateParagonPoint("player1", StatHealth)
	manager.AllocateParagonPoint("player1", StatDamage)
	manager.AllocateParagonPoint("player1", StatDefense)

	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)

	respecConfirmed := false
	ui.SetRespecCallback(func(cost int) bool {
		respecConfirmed = true
		expectedCost := 3 * RespecCostPerPoint
		if cost != expectedCost {
			t.Errorf("respec cost = %d, want %d", cost, expectedCost)
		}
		return true
	})

	ui.Show("player1", "Warrior")

	// Navigate to respec option (index 5)
	for i := 0; i < 5; i++ {
		mock.setMenuDown()
		ui.Update()
		mock.clearKeys()
	}

	if ui.selectedIdx != 5 {
		t.Errorf("selectedIdx = %d, want 5 (respec)", ui.selectedIdx)
	}

	// Activate respec
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	if !respecConfirmed {
		t.Error("respec callback not called")
	}

	// Verify points returned
	prestige := manager.GetPlayerPrestige("player1")
	if prestige.ParagonPoints != 3 {
		t.Errorf("ParagonPoints after respec = %d, want 3", prestige.ParagonPoints)
	}
	totalAllocated := 0
	for _, pts := range prestige.ParagonAllocations {
		totalAllocated += pts
	}
	if totalAllocated != 0 {
		t.Errorf("total allocated after respec = %d, want 0", totalAllocated)
	}
}

func TestPrestigeUI_RespecDenied(t *testing.T) {
	manager := NewManager()
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 2)
	manager.AllocateParagonPoint("player1", StatHealth)
	manager.AllocateParagonPoint("player1", StatDamage)

	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)

	// Set callback that denies respec
	ui.SetRespecCallback(func(cost int) bool {
		return false // Cannot afford
	})

	ui.Show("player1", "Warrior")

	// Navigate to respec
	for i := 0; i < 5; i++ {
		mock.setMenuDown()
		ui.Update()
		mock.clearKeys()
	}

	// Try to respec
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	// Verify points NOT returned
	prestige := manager.GetPlayerPrestige("player1")
	if prestige.ParagonPoints != 0 {
		t.Errorf("ParagonPoints = %d, want 0 (respec should be denied)", prestige.ParagonPoints)
	}
	totalAllocated := prestige.ParagonAllocations[StatHealth] + prestige.ParagonAllocations[StatDamage]
	if totalAllocated != 2 {
		t.Errorf("total allocated = %d, want 2 (unchanged)", totalAllocated)
	}
}

func TestPrestigeUI_Back(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)

	backCalled := false
	ui.SetBackCallback(func() { backCalled = true })

	ui.Show("player1", "Warrior")

	// Navigate to back option (last option)
	for i := 0; i < len(ui.options)-1; i++ {
		mock.setMenuDown()
		ui.Update()
		mock.clearKeys()
	}

	// Activate back
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	if ui.IsVisible() {
		t.Error("UI should be hidden after back")
	}
	if !backCalled {
		t.Error("back callback not called")
	}
}

func TestPrestigeUI_EscapeKey(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)

	backCalled := false
	ui.SetBackCallback(func() { backCalled = true })

	ui.Show("player1", "Warrior")

	// Press escape
	mock.setMenuBack()
	result := ui.Update()
	mock.clearKeys()

	if !result {
		t.Error("Update should return true on escape")
	}
	if ui.IsVisible() {
		t.Error("UI should be hidden after escape")
	}
	if !backCalled {
		t.Error("back callback not called on escape")
	}
}

func TestPrestigeUI_UpdateWhenHidden(t *testing.T) {
	manager := NewManager()
	ui := NewPrestigeUI(800, 600, manager)

	// Update while hidden should return false
	if ui.Update() {
		t.Error("Update while hidden should return false")
	}
}

func TestPrestigeUI_AllocateWithNoPoints(t *testing.T) {
	manager := NewManager()
	manager.CreatePlayer("player1", "Warrior", "account1")
	// No paragon points added

	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)
	ui.Show("player1", "Warrior")

	// Try to allocate (should fail silently)
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	prestige := manager.GetPlayerPrestige("player1")
	if prestige.ParagonAllocations[StatHealth] != 0 {
		t.Errorf("Health allocation = %d, want 0 (no points to allocate)", prestige.ParagonAllocations[StatHealth])
	}
}

func TestPrestigeUI_RespecWithNoAllocations(t *testing.T) {
	manager := NewManager()
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 5)

	ui := NewPrestigeUI(800, 600, manager)
	mock := newMockInputProvider()
	ui.SetInputProvider(mock)

	respecCalled := false
	ui.SetRespecCallback(func(cost int) bool {
		respecCalled = true
		return true
	})

	ui.Show("player1", "Warrior")

	// Navigate to respec
	for i := 0; i < 5; i++ {
		mock.setMenuDown()
		ui.Update()
		mock.clearKeys()
	}

	// Try to respec with no allocations
	mock.setMenuConfirm()
	ui.Update()
	mock.clearKeys()

	// Callback should not be called when nothing to respec
	if respecCalled {
		t.Error("respec callback should not be called when no allocations")
	}
}

// Test getter methods added to manager
func TestManager_GetPlayerPrestige(t *testing.T) {
	manager := NewManager()

	// Test non-existent player
	if prestige := manager.GetPlayerPrestige("nonexistent"); prestige != nil {
		t.Error("GetPlayerPrestige should return nil for non-existent player")
	}

	// Test existing player
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 5)
	manager.AllocateParagonPoint("player1", StatHealth)
	manager.AllocateParagonPoint("player1", StatDamage)

	prestige := manager.GetPlayerPrestige("player1")
	if prestige == nil {
		t.Fatal("GetPlayerPrestige returned nil for existing player")
	}
	if prestige.PlayerID != "player1" {
		t.Errorf("PlayerID = %s, want player1", prestige.PlayerID)
	}
	if prestige.ParagonPoints != 3 {
		t.Errorf("ParagonPoints = %d, want 3", prestige.ParagonPoints)
	}
	if prestige.ParagonAllocations[StatHealth] != 1 {
		t.Errorf("Health allocation = %d, want 1", prestige.ParagonAllocations[StatHealth])
	}
}

func TestManager_GetXPProgress(t *testing.T) {
	manager := NewManager()

	// Test non-existent player
	current, required := manager.GetXPProgress("nonexistent")
	if current != 0 {
		t.Errorf("current XP for nonexistent = %d, want 0", current)
	}
	if required != BasePrestigeXP {
		t.Errorf("required XP for nonexistent = %d, want %d", required, BasePrestigeXP)
	}

	// Test existing player
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddPrestigeXP("player1", "Warrior", 50000)

	current, required = manager.GetXPProgress("player1")
	if current != 50000 {
		t.Errorf("current XP = %d, want 50000", current)
	}
	if required != BasePrestigeXP {
		t.Errorf("required XP = %d, want %d", required, BasePrestigeXP)
	}
}

func TestManager_GetTotalAllocatedPoints(t *testing.T) {
	manager := NewManager()

	// Test non-existent player
	if total := manager.GetTotalAllocatedPoints("nonexistent"); total != 0 {
		t.Errorf("total for nonexistent = %d, want 0", total)
	}

	// Test existing player
	manager.CreatePlayer("player1", "Warrior", "account1")
	manager.AddParagonPoints("player1", 10)
	manager.AllocateParagonPoint("player1", StatHealth)
	manager.AllocateParagonPoint("player1", StatHealth)
	manager.AllocateParagonPoint("player1", StatDamage)
	manager.AllocateParagonPoint("player1", StatDefense)
	manager.AllocateParagonPoint("player1", StatSpeed)

	total := manager.GetTotalAllocatedPoints("player1")
	if total != 5 {
		t.Errorf("total allocated = %d, want 5", total)
	}
}
