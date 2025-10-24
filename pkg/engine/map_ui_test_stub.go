//go:build test
// +build test

// Package engine provides map UI testing stubs.
// This file implements test-safe versions of MapUI for testing without Ebiten.
package engine

import (
	"github.com/opd-ai/venture/pkg/procgen/terrain"
)

// MapUI stub for testing
type MapUI struct {
	visible        bool
	fullScreen     bool
	world          *World
	playerEntity   *Entity
	terrain        *terrain.Terrain
	screenWidth    int
	screenHeight   int
	fogOfWar       [][]bool
	scale          float64
	offsetX        float64
	offsetY        float64
	minimapSize    int
	minimapPadding int
	mapNeedsUpdate bool
}

// NewMapUI creates a new map UI stub for testing.
func NewMapUI(world *World, screenWidth, screenHeight int) *MapUI {
	return &MapUI{
		visible:        false,
		fullScreen:     false,
		world:          world,
		screenWidth:    screenWidth,
		screenHeight:   screenHeight,
		scale:          1.0,
		minimapSize:    150,
		minimapPadding: 10,
		mapNeedsUpdate: true,
	}
}

// SetPlayerEntity sets the player entity whose position to track.
func (ui *MapUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
	ui.mapNeedsUpdate = true
}

// SetTerrain sets the current level terrain to display.
func (ui *MapUI) SetTerrain(terrain *terrain.Terrain) {
	ui.terrain = terrain
	if terrain != nil {
		ui.fogOfWar = make([][]bool, terrain.Height)
		for y := range ui.fogOfWar {
			ui.fogOfWar[y] = make([]bool, terrain.Width)
		}
	}
	ui.mapNeedsUpdate = true
}

// ToggleFullScreen switches between minimap and full-screen modes.
func (ui *MapUI) ToggleFullScreen() {
	ui.fullScreen = !ui.fullScreen
	ui.visible = true
	ui.mapNeedsUpdate = true
}

// IsFullScreen returns whether full-screen map is shown.
func (ui *MapUI) IsFullScreen() bool {
	return ui.fullScreen
}

// ShowFullScreen displays the full-screen map.
func (ui *MapUI) ShowFullScreen() {
	ui.fullScreen = true
	ui.visible = true
	ui.mapNeedsUpdate = true
}

// HideFullScreen returns to minimap mode.
func (ui *MapUI) HideFullScreen() {
	ui.fullScreen = false
}

// Update processes input and updates fog of war (stub).
func (ui *MapUI) Update(deltaTime float64) {
	// Stub: No Ebiten input in tests
}

// Draw renders the map overlay (stub).
func (ui *MapUI) Draw(screen interface{}) {
	// Stub: No Ebiten rendering in tests
}

// GetFogOfWar returns the fog of war state for testing.
func (ui *MapUI) GetFogOfWar() [][]bool {
	return ui.fogOfWar
}

// SetFogOfWar sets the fog of war state (for save/load).
func (ui *MapUI) SetFogOfWar(fog [][]bool) {
	ui.fogOfWar = fog
	ui.mapNeedsUpdate = true
}
