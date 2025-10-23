// +build test

// Package engine provides map UI testing.
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// TestMapUI_NewMapUI tests initialization of MapUI.
func TestMapUI_NewMapUI(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	if ui == nil {
		t.Fatal("NewMapUI returned nil")
	}

	if ui.screenWidth != 800 {
		t.Errorf("Expected screenWidth 800, got %d", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("Expected screenHeight 600, got %d", ui.screenHeight)
	}

	if ui.visible {
		t.Error("MapUI should not be visible on initialization")
	}

	if ui.fullScreen {
		t.Error("MapUI should not be full-screen on initialization")
	}
}

// TestMapUI_ToggleFullScreen tests full-screen toggling.
func TestMapUI_ToggleFullScreen(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	// Initially not full-screen
	if ui.IsFullScreen() {
		t.Error("MapUI should not be full-screen initially")
	}

	// Toggle to full-screen
	ui.ToggleFullScreen()
	if !ui.IsFullScreen() {
		t.Error("MapUI should be full-screen after toggle")
	}
	if !ui.visible {
		t.Error("MapUI should be visible after toggling to full-screen")
	}

	// Toggle back
	ui.ToggleFullScreen()
	if ui.IsFullScreen() {
		t.Error("MapUI should not be full-screen after second toggle")
	}
}

// TestMapUI_SetPlayerEntity tests player entity assignment.
func TestMapUI_SetPlayerEntity(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	entity := world.CreateEntity()
	entity.AddComponent(&PositionComponent{X: 100, Y: 200})

	ui.SetPlayerEntity(entity)

	if ui.playerEntity != entity {
		t.Error("Player entity was not set correctly")
	}
}

// TestMapUI_SetTerrain tests terrain initialization.
func TestMapUI_SetTerrain(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	// Create test terrain
	terrain := terrain.NewTerrain(50, 50, 12345)

	ui.SetTerrain(terrain)

	if ui.terrain != terrain {
		t.Error("Terrain was not set correctly")
	}

	// Check fog of war initialization
	fogOfWar := ui.GetFogOfWar()
	if fogOfWar == nil {
		t.Fatal("Fog of war not initialized")
	}

	if len(fogOfWar) != terrain.Height {
		t.Errorf("Fog of war height mismatch: expected %d, got %d", terrain.Height, len(fogOfWar))
	}

	if len(fogOfWar[0]) != terrain.Width {
		t.Errorf("Fog of war width mismatch: expected %d, got %d", terrain.Width, len(fogOfWar[0]))
	}

	// All tiles should start unexplored
	for y := 0; y < terrain.Height; y++ {
		for x := 0; x < terrain.Width; x++ {
			if fogOfWar[y][x] {
				t.Errorf("Tile (%d,%d) should be unexplored initially", x, y)
			}
		}
	}
}

// TestMapUI_ShowFullScreen tests explicit full-screen display.
func TestMapUI_ShowFullScreen(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	ui.ShowFullScreen()

	if !ui.IsFullScreen() {
		t.Error("MapUI should be full-screen after ShowFullScreen")
	}
	if !ui.visible {
		t.Error("MapUI should be visible after ShowFullScreen")
	}
}

// TestMapUI_HideFullScreen tests hiding full-screen mode.
func TestMapUI_HideFullScreen(t *testing.T) {
	world := NewWorld()
	ui := NewMapUI(world, 800, 600)

	ui.ShowFullScreen()
	ui.HideFullScreen()

	if ui.IsFullScreen() {
		t.Error("MapUI should not be full-screen after HideFullScreen")
	}
}
