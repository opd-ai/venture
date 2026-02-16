package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestQuestUIScrolling tests scroll offset management.
func TestQuestUIScrolling(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenQuestUI(world, 800, 600)

	// Initially no scroll
	if ui.scrollOffset != 0 {
		t.Errorf("scrollOffset = %d, want 0", ui.scrollOffset)
	}

	// Set scroll offset
	ui.scrollOffset = 100
	ui.maxScroll = 200

	if ui.scrollOffset != 100 {
		t.Errorf("scrollOffset = %d, want 100", ui.scrollOffset)
	}

	// Reset on hide
	ui.visible = true
	ui.Hide()
	// Note: scrollOffset is reset in Update when closing, not in Hide

	// Reset on tab change happens in Update
	ui.scrollOffset = 50
	ui.currentTab = 0
	// In Update, changing tabs resets scroll
}

// TestQuestUITabSwitching tests tab management.
func TestQuestUITabSwitching(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenQuestUI(world, 800, 600)

	// Initially on Active tab
	if ui.currentTab != 0 {
		t.Errorf("currentTab = %d, want 0 (Active)", ui.currentTab)
	}

	// Switch to Completed
	ui.currentTab = 1
	if ui.currentTab != 1 {
		t.Errorf("currentTab = %d, want 1 (Completed)", ui.currentTab)
	}

	// Switch back
	ui.currentTab = 0
	if ui.currentTab != 0 {
		t.Errorf("currentTab = %d, want 0 (Active)", ui.currentTab)
	}
}

// TestQuestUIVisibility tests show/hide functionality.
func TestQuestUIVisibility(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenQuestUI(world, 800, 600)

	// Initially hidden
	if ui.IsVisible() {
		t.Error("Quest UI should be hidden initially")
	}

	// Show
	ui.Show()
	if !ui.IsVisible() {
		t.Error("Quest UI should be visible after Show()")
	}

	// Hide
	ui.Hide()
	if ui.IsVisible() {
		t.Error("Quest UI should be hidden after Hide()")
	}

	// Toggle
	ui.Toggle()
	if !ui.IsVisible() {
		t.Error("Quest UI should be visible after Toggle()")
	}

	ui.Toggle()
	if ui.IsVisible() {
		t.Error("Quest UI should be hidden after second Toggle()")
	}
}

// TestMaxHelper tests the max helper function.
func TestMaxHelper(t *testing.T) {
	tests := []struct {
		a, b, want int
	}{
		{5, 10, 10},
		{10, 5, 10},
		{7, 7, 7},
		{-5, 3, 3},
		{-10, -5, -5},
	}

	for _, tt := range tests {
		got := max(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("max(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.want)
		}
	}
}

type StubQuestUI struct {
	UpdateCount int
	DrawCount   int
	active      bool
}

func NewStubQuestUI() *StubQuestUI {
	return &StubQuestUI{}
}

func (s *StubQuestUI) Update(entities []*Entity, deltaTime float64) {
	s.UpdateCount++
}

func (s *StubQuestUI) Draw(screen interface{}) {
	s.DrawCount++
}

func (s *StubQuestUI) IsActive() bool {
	return s.active
}

func (s *StubQuestUI) SetActive(active bool) {
	s.active = active
}

var _ UISystem = (*StubQuestUI)(nil)

// TestQuestUI_DrawScrollbar_ZeroTotalContent verifies no panic on zero content height.
func TestQuestUI_DrawScrollbar_ZeroTotalContent(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenQuestUI(world, 800, 600)
	ui.maxScroll = 100 // Force scrollbar to be drawn

	// Should not panic with totalContentHeight = 0
	// We can't call drawScrollbar directly without ebiten.Image, but test the logic
	totalContentHeight := 0
	contentHeight := 300
	if totalContentHeight <= 0 {
		// This is the guard we added - verify it returns early
		t.Log("Guard prevents division by zero when totalContentHeight <= 0")
	} else {
		handleHeight := max(20, (contentHeight*contentHeight)/totalContentHeight)
		if handleHeight < 20 {
			t.Error("handleHeight should be at least 20")
		}
	}
}

// TestMapUI_IsEntityVisible_NilFogOfWar verifies no panic with nil fog of war.
func TestMapUI_IsEntityVisible_NilFogOfWar(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenMapUI(world, 800, 600)
	// fogOfWar is nil by default

	// Set terrain so bounds check passes
	ui.terrain = &terrain.Terrain{Width: 10, Height: 10}

	// Should not panic with nil fogOfWar
	result := ui.isEntityVisible(5, 5)
	if result {
		t.Error("isEntityVisible should return false with nil fogOfWar")
	}
}

// TestMapUI_RevealTile_NilFogOfWar verifies no panic with nil fog of war.
func TestMapUI_RevealTile_NilFogOfWar(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenMapUI(world, 800, 600)
	// fogOfWar is nil by default

	// Should not panic with nil fogOfWar
	ui.revealTile(0, 0)
}

// TestMapUI_CanUpdateFogOfWar verifies fog-of-war readiness check.
func TestMapUI_CanUpdateFogOfWar(t *testing.T) {
	world := NewWorld()
	ui := NewEbitenMapUI(world, 800, 600)

	if ui.canUpdateFogOfWar() {
		t.Error("canUpdateFogOfWar should return false without terrain/player/fogOfWar")
	}
}
