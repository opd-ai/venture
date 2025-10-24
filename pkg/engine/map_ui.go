// Package engine provides map_ui for game UI.
package engine

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/opd-ai/venture/pkg/procgen/terrain"
	"golang.org/x/image/font/basicfont"
)

// MapUI handles rendering and interaction for the world map display.
type EbitenMapUI struct {
	visible      bool
	fullScreen   bool // true = full-screen map, false = minimap
	world        *World
	playerEntity *Entity
	terrain      *terrain.Terrain
	screenWidth  int
	screenHeight int

	// Map rendering
	mapImage       *ebiten.Image // Cached map rendering
	mapNeedsUpdate bool          // Regenerate map on next frame
	fogOfWar       [][]bool      // 2D array: true = explored
	scale          float64       // Zoom level for full-screen mode
	offsetX        float64       // Pan offset X (for large maps)
	offsetY        float64       // Pan offset Y

	// Minimap settings
	minimapSize    int // Size in pixels (square)
	minimapPadding int // Distance from screen edge
}

// NewMapUI creates a new map UI system.
// Parameters:
//
//	world - ECS world instance
//	screenWidth, screenHeight - Display dimensions
//
// Returns: Initialized MapUI
// Called by: Game.NewGame()
func NewEbitenMapUI(world *World, screenWidth, screenHeight int) *EbitenMapUI {
	return &EbitenMapUI{
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
// Parameters:
//
//	entity - Player entity with PositionComponent
//
// Called by: Game.SetPlayerEntity()
func (ui *EbitenMapUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
	ui.mapNeedsUpdate = true
}

// SetTerrain sets the current level terrain to display.
// Parameters:
//
//	terrain - Terrain data from TerrainRenderSystem
//
// Called by: Game after terrain generation
func (ui *EbitenMapUI) SetTerrain(terrain *terrain.Terrain) {
	ui.terrain = terrain
	if terrain != nil {
		// Initialize fog of war to match terrain dimensions
		ui.fogOfWar = make([][]bool, terrain.Height)
		for y := range ui.fogOfWar {
			ui.fogOfWar[y] = make([]bool, terrain.Width)
		}
	}
	ui.mapNeedsUpdate = true
}

// GAP-005 REPAIR: Add fog of war getter for save/load system
// GetFogOfWar returns a copy of the fog of war exploration state.
// Returns: 2D boolean array where true = explored
// Called by: SaveManager when serializing game state
func (ui *EbitenMapUI) GetFogOfWar() [][]bool {
	if ui.fogOfWar == nil {
		return nil
	}
	// Return deep copy to prevent external modification
	fogCopy := make([][]bool, len(ui.fogOfWar))
	for y := range ui.fogOfWar {
		fogCopy[y] = make([]bool, len(ui.fogOfWar[y]))
		copy(fogCopy[y], ui.fogOfWar[y])
	}
	return fogCopy
}

// GAP-005 REPAIR: Add fog of war setter for save/load system
// SetFogOfWar restores fog of war exploration state from save file.
// Parameters:
//
//	fogOfWar - 2D boolean array where true = explored
//
// Called by: LoadManager when deserializing game state
func (ui *EbitenMapUI) SetFogOfWar(fogOfWar [][]bool) {
	if fogOfWar == nil {
		return
	}
	// Deep copy the provided fog of war data
	ui.fogOfWar = make([][]bool, len(fogOfWar))
	for y := range fogOfWar {
		ui.fogOfWar[y] = make([]bool, len(fogOfWar[y]))
		copy(ui.fogOfWar[y], fogOfWar[y])
	}
	ui.mapNeedsUpdate = true
}

// ToggleFullScreen switches between minimap and full-screen modes.
// Called by: InputSystem when M key is pressed
func (ui *EbitenMapUI) ToggleFullScreen() {
	ui.fullScreen = !ui.fullScreen
	ui.visible = true // Always make visible when toggling
	if ui.fullScreen {
		ui.centerOnPlayer()
	}
	ui.mapNeedsUpdate = true
}

// IsFullScreen returns whether full-screen map is shown.
// Returns: true if full-screen, false if minimap or hidden
// Called by: Game.Update() to block input
func (ui *EbitenMapUI) IsFullScreen() bool {
	return ui.fullScreen
}

// ShowFullScreen displays the full-screen map.
func (ui *EbitenMapUI) ShowFullScreen() {
	ui.fullScreen = true
	ui.visible = true
	ui.centerOnPlayer()
	ui.mapNeedsUpdate = true
}

// HideFullScreen returns to minimap mode.
func (ui *EbitenMapUI) HideFullScreen() {
	ui.fullScreen = false
}

// Update processes input and updates fog of war.
// Parameters:
//
//	deltaTime - Time since last frame
//
// Called by: Game.Update() every frame
func (ui *EbitenMapUI) Update(entities []*Entity, deltaTime float64) {
	// Always update fog of war (even when not visible)
	ui.updateFogOfWar()

	if !ui.visible {
		return
	}

	// Handle M key toggle for full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		ui.ToggleFullScreen()
		return
	}

	// Handle ESC key to close full-screen map
	if ui.fullScreen && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.HideFullScreen()
		return
	}

	// Handle input for full-screen mode
	if ui.fullScreen {
		// Pan with arrow keys
		panSpeed := 200.0 * deltaTime
		if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
			ui.panMap(-panSpeed, 0)
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
			ui.panMap(panSpeed, 0)
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
			ui.panMap(0, -panSpeed)
		}
		if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
			ui.panMap(0, panSpeed)
		}

		// Zoom with mouse wheel
		_, wheelY := ebiten.Wheel()
		if wheelY != 0 {
			ui.zoomMap(wheelY * 0.1)
		}

		// Center on player with Space key
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			ui.centerOnPlayer()
		}
	}

	// Regenerate map if needed
	if ui.mapNeedsUpdate {
		ui.regenerateMapImage()
	}
}

// Draw renders the map overlay (minimap or full-screen).
// Parameters:
//
//	screen - Ebiten image
//
// Called by: Game.Draw() every frame
func (ui *EbitenMapUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	if !ui.visible {
		return
	}

	if ui.fullScreen {
		ui.drawFullScreenMap(img)
	} else {
		ui.drawMinimap(img)
	}
}

// drawMinimap renders the compact minimap in corner.
// Parameters:
//
//	screen - Target image
//
// Called by: Draw() when fullScreen is false
func (ui *EbitenMapUI) drawMinimap(screen *ebiten.Image) {
	if ui.terrain == nil || ui.playerEntity == nil {
		return
	}

	// Calculate minimap position (top-right corner)
	mapX := ui.screenWidth - ui.minimapSize - ui.minimapPadding
	mapY := ui.minimapPadding

	// Draw minimap background
	vector.DrawFilledRect(screen, float32(mapX), float32(mapY),
		float32(ui.minimapSize), float32(ui.minimapSize),
		color.RGBA{0, 0, 0, 200}, false)

	// Draw minimap border
	vector.StrokeRect(screen, float32(mapX), float32(mapY),
		float32(ui.minimapSize), float32(ui.minimapSize), 2,
		color.RGBA{255, 255, 255, 255}, false)

	// Calculate tile scaling
	scaleX := float64(ui.minimapSize) / float64(ui.terrain.Width)
	scaleY := float64(ui.minimapSize) / float64(ui.terrain.Height)
	tileScale := math.Min(scaleX, scaleY)

	// Draw terrain tiles
	for y := 0; y < ui.terrain.Height; y++ {
		for x := 0; x < ui.terrain.Width; x++ {
			if !ui.fogOfWar[y][x] {
				continue // Skip unexplored tiles
			}

			tileType := ui.terrain.GetTile(x, y)
			tileColor := ui.getTileColor(tileType, true)

			pixelX := float32(mapX) + float32(float64(x)*tileScale)
			pixelY := float32(mapY) + float32(float64(y)*tileScale)
			pixelSize := float32(tileScale)

			if pixelSize < 1 {
				pixelSize = 1
			}

			vector.DrawFilledRect(screen, pixelX, pixelY, pixelSize, pixelSize, tileColor, false)
		}
	}

	// Draw player icon
	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
		pos := posComp.(*PositionComponent)
		// Convert world position to tile coordinates (assuming 32px tiles)
		tileX := int(pos.X / 32)
		tileY := int(pos.Y / 32)

		if tileX >= 0 && tileX < ui.terrain.Width && tileY >= 0 && tileY < ui.terrain.Height {
			pixelX := float32(mapX) + float32(float64(tileX)*tileScale)
			pixelY := float32(mapY) + float32(float64(tileY)*tileScale)

			// Draw player as blue circle
			vector.DrawFilledCircle(screen, pixelX, pixelY, 3, color.RGBA{100, 150, 255, 255}, false)
		}
	}

	// Draw compass rose (N indicator)
	compassText := "N"
	text.Draw(screen, compassText, basicfont.Face7x13,
		mapX+ui.minimapSize/2-3, mapY-5, color.RGBA{255, 255, 255, 255})
}

// drawFullScreenMap renders the large detailed map.
// Parameters:
//
//	screen - Target image
//
// Called by: Draw() when fullScreen is true
func (ui *EbitenMapUI) drawFullScreenMap(screen *ebiten.Image) {
	if ui.terrain == nil {
		return
	}

	// Draw semi-transparent overlay
	vector.DrawFilledRect(screen, 0, 0, float32(ui.screenWidth), float32(ui.screenHeight),
		color.RGBA{0, 0, 0, 200}, false)

	// Draw map panel
	panelWidth := ui.screenWidth - 100
	panelHeight := ui.screenHeight - 100
	panelX := 50
	panelY := 50

	vector.DrawFilledRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 30, 255}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight), 2,
		color.RGBA{100, 150, 200, 255}, false)

	// Title
	titleText := "WORLD MAP"
	titleX := panelX + panelWidth/2 - len(titleText)*3
	text.Draw(screen, titleText, basicfont.Face7x13, titleX, panelY+20,
		color.RGBA{255, 255, 100, 255})

	// Draw terrain
	mapAreaX := panelX + 10
	mapAreaY := panelY + 40
	mapAreaWidth := panelWidth - 20
	mapAreaHeight := panelHeight - 80

	// Calculate visible area based on scale and offset
	tileSize := 8.0 * ui.scale // Base 8px per tile, scaled
	startTileX := int(ui.offsetX / tileSize)
	startTileY := int(ui.offsetY / tileSize)

	if startTileX < 0 {
		startTileX = 0
	}
	if startTileY < 0 {
		startTileY = 0
	}

	endTileX := startTileX + int(float64(mapAreaWidth)/tileSize) + 1
	endTileY := startTileY + int(float64(mapAreaHeight)/tileSize) + 1

	if endTileX > ui.terrain.Width {
		endTileX = ui.terrain.Width
	}
	if endTileY > ui.terrain.Height {
		endTileY = ui.terrain.Height
	}

	// Draw tiles
	for y := startTileY; y < endTileY; y++ {
		for x := startTileX; x < endTileX; x++ {
			if y < 0 || y >= ui.terrain.Height || x < 0 || x >= ui.terrain.Width {
				continue
			}

			explored := ui.fogOfWar[y][x]
			tileType := ui.terrain.GetTile(x, y)

			screenX := float32(mapAreaX) + float32((float64(x)*tileSize)-(ui.offsetX))
			screenY := float32(mapAreaY) + float32((float64(y)*tileSize)-(ui.offsetY))

			tileColor := ui.getTileColor(tileType, explored)

			vector.DrawFilledRect(screen, screenX, screenY,
				float32(tileSize), float32(tileSize), tileColor, false)
		}
	}

	// Draw map icons (player, enemies, items)
	ui.drawMapIcons(screen, mapAreaX, mapAreaY, tileSize, startTileX, startTileY)

	// Draw legend
	ui.drawLegend(screen, panelX+10, panelY+panelHeight-60)

	// Draw controls
	controlsText := "[Arrow Keys/WASD] Pan | [Mouse Wheel] Zoom | [Space] Center | [M] or [ESC] Close"
	text.Draw(screen, controlsText, basicfont.Face7x13,
		panelX+10, panelY+panelHeight-10, color.RGBA{180, 180, 180, 255})
}

// updateFogOfWar marks tiles as explored based on player visibility.
// Called by: Update() every frame
func (ui *EbitenMapUI) updateFogOfWar() {
	if ui.terrain == nil || ui.playerEntity == nil || ui.fogOfWar == nil {
		return
	}

	// Get player position
	posComp, ok := ui.playerEntity.GetComponent("position")
	if !ok {
		return
	}

	pos := posComp.(*PositionComponent)
	// Convert world position to tile coordinates (assuming 32px tiles)
	centerX := int(pos.X / 32)
	centerY := int(pos.Y / 32)

	// Reveal tiles within vision radius
	radius := ui.getVisibleRadius()

	for dy := -radius; dy <= radius; dy++ {
		for dx := -radius; dx <= radius; dx++ {
			// Check if within circular radius
			dist := math.Sqrt(float64(dx*dx + dy*dy))
			if dist > float64(radius) {
				continue
			}

			tileX := centerX + dx
			tileY := centerY + dy

			// Bounds check
			if tileX < 0 || tileX >= ui.terrain.Width || tileY < 0 || tileY >= ui.terrain.Height {
				continue
			}

			// Mark as explored
			if !ui.fogOfWar[tileY][tileX] {
				ui.fogOfWar[tileY][tileX] = true
				ui.mapNeedsUpdate = true
			}
		}
	}
}

// regenerateMapImage rebuilds the cached map rendering.
// Called by: Update() when mapNeedsUpdate is true
func (ui *EbitenMapUI) regenerateMapImage() {
	// For now, we regenerate on each draw
	// Future optimization: pre-render to an image
	ui.mapNeedsUpdate = false
}

// tileToScreen converts tile coordinates to screen coordinates.
// Parameters:
//
//	tileX, tileY - Tile coordinates
//
// Returns: Screen pixel coordinates
func (ui *EbitenMapUI) tileToScreen(tileX, tileY int) (int, int) {
	tileSize := 8.0 * ui.scale
	screenX := (float64(tileX) * tileSize) - ui.offsetX
	screenY := (float64(tileY) * tileSize) - ui.offsetY
	return int(screenX), int(screenY)
}

// screenToTile converts screen coordinates to tile coordinates.
// Parameters:
//
//	screenX, screenY - Screen pixel coordinates
//
// Returns: Tile coordinates
func (ui *EbitenMapUI) screenToTile(screenX, screenY int) (int, int) {
	tileSize := 8.0 * ui.scale
	tileX := int((float64(screenX) + ui.offsetX) / tileSize)
	tileY := int((float64(screenY) + ui.offsetY) / tileSize)
	return tileX, tileY
}

// drawMapTile renders a single tile on the map (unused - kept for API completeness).
func (ui *EbitenMapUI) drawMapTile(screen *ebiten.Image, tileX, tileY int, tileType terrain.TileType, explored bool) {
	// Implementation merged into drawFullScreenMap for performance
}

// drawMapIcons renders player, enemy, item markers.
// Parameters:
//
//	screen - Target image
//
// Called by: drawFullScreenMap() and drawMinimap()
func (ui *EbitenMapUI) drawMapIcons(screen *ebiten.Image, mapAreaX, mapAreaY int, tileSize float64, startTileX, startTileY int) {
	if ui.playerEntity == nil {
		return
	}

	// Draw player icon
	if posComp, ok := ui.playerEntity.GetComponent("position"); ok {
		pos := posComp.(*PositionComponent)
		tileX := int(pos.X / 32)
		tileY := int(pos.Y / 32)

		if tileX >= startTileX && tileY >= startTileY {
			screenX := float32(mapAreaX) + float32((float64(tileX)*tileSize)-(ui.offsetX))
			screenY := float32(mapAreaY) + float32((float64(tileY)*tileSize)-(ui.offsetY))

			// Draw player as blue circle
			vector.DrawFilledCircle(screen, screenX+float32(tileSize)/2, screenY+float32(tileSize)/2,
				float32(tileSize)/2, color.RGBA{100, 150, 255, 255}, false)
		}
	}

	// Draw other entities (enemies, items)
	for _, entity := range ui.world.GetEntities() {
		if entity == ui.playerEntity {
			continue
		}

		posComp, hasPos := entity.GetComponent("position")
		if !hasPos {
			continue
		}

		pos := posComp.(*PositionComponent)
		tileX := int(pos.X / 32)
		tileY := int(pos.Y / 32)

		// Only draw if explored
		if tileX < 0 || tileX >= ui.terrain.Width || tileY < 0 || tileY >= ui.terrain.Height {
			continue
		}
		if !ui.fogOfWar[tileY][tileX] {
			continue
		}

		screenX := float32(mapAreaX) + float32((float64(tileX)*tileSize)-(ui.offsetX))
		screenY := float32(mapAreaY) + float32((float64(tileY)*tileSize)-(ui.offsetY))

		// Determine icon color based on entity type
		iconColor := color.RGBA{200, 200, 200, 255} // Default gray

		// Check if enemy
		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
			team := teamComp.(*TeamComponent)
			if team.TeamID == 2 { // Enemy team
				iconColor = color.RGBA{255, 100, 100, 255} // Red
			}
		}

		// Draw entity marker (small square)
		markerSize := float32(tileSize) / 3
		vector.DrawFilledRect(screen, screenX+float32(tileSize)/2-markerSize/2,
			screenY+float32(tileSize)/2-markerSize/2,
			markerSize, markerSize, iconColor, false)
	}
}

// getVisibleRadius returns the player's vision radius in tiles.
// Returns: Number of tiles visible around player
func (ui *EbitenMapUI) getVisibleRadius() int {
	// Default vision radius
	return 10
}

// panMap adjusts offsetX/offsetY for map panning (full-screen mode).
// Parameters:
//
//	dx, dy - Delta movement
//
// Called by: Update() when arrow keys pressed in full-screen mode
func (ui *EbitenMapUI) panMap(dx, dy float64) {
	ui.offsetX += dx
	ui.offsetY += dy

	// Clamp to terrain bounds
	maxOffsetX := float64(ui.terrain.Width) * 8.0 * ui.scale
	maxOffsetY := float64(ui.terrain.Height) * 8.0 * ui.scale

	if ui.offsetX < 0 {
		ui.offsetX = 0
	}
	if ui.offsetY < 0 {
		ui.offsetY = 0
	}
	if ui.offsetX > maxOffsetX {
		ui.offsetX = maxOffsetX
	}
	if ui.offsetY > maxOffsetY {
		ui.offsetY = maxOffsetY
	}
}

// zoomMap adjusts scale for zooming (full-screen mode).
// Parameters:
//
//	delta - Zoom delta (positive = zoom in, negative = zoom out)
//
// Called by: Update() on mouse wheel input
func (ui *EbitenMapUI) zoomMap(delta float64) {
	ui.scale += delta
	if ui.scale < 0.5 {
		ui.scale = 0.5
	}
	if ui.scale > 4.0 {
		ui.scale = 4.0
	}
	ui.mapNeedsUpdate = true
}

// centerOnPlayer resets pan/zoom to center on player.
// Called by: ShowFullScreen() when opening map
func (ui *EbitenMapUI) centerOnPlayer() {
	if ui.playerEntity == nil || ui.terrain == nil {
		return
	}

	// Get player position
	posComp, ok := ui.playerEntity.GetComponent("position")
	if !ok {
		return
	}

	pos := posComp.(*PositionComponent)
	tileX := int(pos.X / 32)
	tileY := int(pos.Y / 32)

	// Center offset on player
	tileSize := 8.0 * ui.scale
	mapAreaWidth := ui.screenWidth - 120 // Account for panel margins
	mapAreaHeight := ui.screenHeight - 180

	ui.offsetX = (float64(tileX) * tileSize) - float64(mapAreaWidth)/2
	ui.offsetY = (float64(tileY) * tileSize) - float64(mapAreaHeight)/2

	// Clamp
	if ui.offsetX < 0 {
		ui.offsetX = 0
	}
	if ui.offsetY < 0 {
		ui.offsetY = 0
	}

	ui.mapNeedsUpdate = true
}

// getTileColor returns a color for a terrain tile type.
func (ui *EbitenMapUI) getTileColor(tileType terrain.TileType, explored bool) color.Color {
	var baseColor color.Color

	switch tileType {
	case terrain.TileWall:
		baseColor = color.RGBA{60, 60, 60, 255}
	case terrain.TileFloor:
		baseColor = color.RGBA{180, 180, 180, 255}
	case terrain.TileDoor:
		baseColor = color.RGBA{139, 69, 19, 255}
	case terrain.TileCorridor:
		baseColor = color.RGBA{150, 150, 150, 255}
	default:
		baseColor = color.RGBA{100, 100, 100, 255}
	}

	if !explored {
		// Unexplored tiles are black
		return color.RGBA{0, 0, 0, 255}
	}

	return baseColor
}

// drawLegend renders the map legend explaining tile colors.
func (ui *EbitenMapUI) drawLegend(screen *ebiten.Image, x, y int) {
	text.Draw(screen, "Legend:", basicfont.Face7x13, x, y, color.RGBA{200, 200, 200, 255})

	legendItems := []struct {
		color color.Color
		label string
	}{
		{color.RGBA{180, 180, 180, 255}, "Floor"},
		{color.RGBA{60, 60, 60, 255}, "Wall"},
		{color.RGBA{139, 69, 19, 255}, "Door"},
		{color.RGBA{100, 150, 255, 255}, "Player"},
		{color.RGBA{255, 100, 100, 255}, "Enemy"},
	}

	y += 15
	for _, item := range legendItems {
		// Draw color swatch
		vector.DrawFilledRect(screen, float32(x), float32(y-8), 10, 10, item.color, false)

		// Draw label
		text.Draw(screen, item.label, basicfont.Face7x13, x+15, y, color.RGBA{180, 180, 180, 255})
		y += 15
	}
}

// IsActive returns whether the map UI is currently visible.
// Implements UISystem interface.
func (m *EbitenMapUI) IsActive() bool {
	return m.visible
}

// SetActive sets whether the map UI is visible.
// Implements UISystem interface.
func (m *EbitenMapUI) SetActive(active bool) {
	m.visible = active
}

// Compile-time check that EbitenMapUI implements UISystem
var _ UISystem = (*EbitenMapUI)(nil)
