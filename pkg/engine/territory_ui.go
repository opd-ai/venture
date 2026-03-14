package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/world/territory"
	log "github.com/sirupsen/logrus"
)

// TerritoryUI provides a UI for viewing and managing territories.
type TerritoryUI struct {
	visible           bool
	territorySystem   *TerritorySystem
	playerEntity      *Entity
	screenWidth       int
	screenHeight      int
	selectedTerritory *territory.Territory
	scrollOffset      int
	touchHandler      *mobile.TouchInputHandler
	input             InputProvider

	// PERF: Cached images to avoid per-frame allocations (Critical Issue #2)
	cachedOverlay *ebiten.Image // Semi-transparent background overlay
}

// NewTerritoryUI creates a new territory UI.
func NewTerritoryUI(territorySystem *TerritorySystem, screenWidth, screenHeight int) *TerritoryUI {
	return &TerritoryUI{
		visible:         false,
		territorySystem: territorySystem,
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		scrollOffset:    0,
		touchHandler:    mobile.NewTouchInputHandler(),
	}
}

// SetInputProvider sets the InputProvider used for keyboard navigation.
// Call this after creation to enable testable, rebindable input handling.
func (tui *TerritoryUI) SetInputProvider(input InputProvider) {
	tui.input = input
}

// SetPlayerEntity sets the player entity for context.
func (tui *TerritoryUI) SetPlayerEntity(player *Entity) {
	tui.playerEntity = player
}

// IsVisible returns whether the UI is currently visible.
func (tui *TerritoryUI) IsVisible() bool {
	return tui.visible
}

// Toggle toggles the UI visibility.
func (tui *TerritoryUI) Toggle() {
	tui.visible = !tui.visible
	if tui.visible {
		tui.refreshCurrentTerritory()
	}
}

// Show shows the UI.
func (tui *TerritoryUI) Show() {
	tui.visible = true
	tui.refreshCurrentTerritory()
}

// Hide hides the UI.
func (tui *TerritoryUI) Hide() {
	tui.visible = false
}

// Dispose releases cached resources to prevent GPU memory leaks.
// Call this when the UI is no longer needed.
func (tui *TerritoryUI) Dispose() {
	if tui.cachedOverlay != nil {
		tui.cachedOverlay.Dispose()
		tui.cachedOverlay = nil
	}
}

// Update processes input for the territory UI.
func (tui *TerritoryUI) Update() error {
	if !tui.visible {
		return nil
	}

	tui.handleKeyboardInput()
	tui.handleTouchInput()
	return nil
}

// handleKeyboardInput processes keyboard navigation and actions.
// Uses InputProvider if set, enabling testability and input rebinding.
func (tui *TerritoryUI) handleKeyboardInput() {
	if tui.input == nil {
		return
	}

	// Keyboard navigation
	if tui.input.IsMenuUpJustPressed() {
		tui.scrollOffset--
		if tui.scrollOffset < 0 {
			tui.scrollOffset = 0
		}
	}
	if tui.input.IsMenuDownJustPressed() {
		tui.scrollOffset++
	}

	// Declare war (confirm action)
	if tui.input.IsMenuConfirmJustPressed() && tui.selectedTerritory != nil {
		tui.handleDeclareWar()
	}

	// Build structure (back/alternate action)
	if tui.input.IsMenuBackJustPressed() && tui.selectedTerritory != nil {
		tui.handleBuildStructure()
	}
}

// handleTouchInput processes touch input for UI interaction.
func (tui *TerritoryUI) handleTouchInput() {
	if tui.touchHandler != nil {
		tui.touchHandler.Update()
	}

	touches := ebiten.AppendTouchIDs(nil)
	if len(touches) > 0 {
		for _, touchID := range touches {
			x, y := ebiten.TouchPosition(touchID)

			// Close button area (top right)
			closeX := tui.screenWidth - 60
			closeY := 10
			if x >= closeX && x <= closeX+50 && y >= closeY && y <= closeY+30 {
				tui.Hide()
			}
		}
	}
}

// Draw renders the territory UI.
func (tui *TerritoryUI) Draw(screen *ebiten.Image) {
	if !tui.visible {
		return
	}

	tui.drawBackground(screen)
	yOffset := tui.drawHeader(screen)
	yOffset = tui.drawCurrentTerritoryInfo(screen, yOffset)
	yOffset = tui.drawGuildTerritories(screen, yOffset)
	yOffset = tui.drawActiveWars(screen, yOffset)
	tui.drawControls(screen)
}

func (tui *TerritoryUI) drawBackground(screen *ebiten.Image) {
	// PERF: Reuse cached overlay image, only fill on creation/resize
	if tui.cachedOverlay == nil || tui.cachedOverlay.Bounds().Dx() != tui.screenWidth || tui.cachedOverlay.Bounds().Dy() != tui.screenHeight {
		if tui.cachedOverlay != nil {
			tui.cachedOverlay.Dispose()
		}
		tui.cachedOverlay = ebiten.NewImage(tui.screenWidth, tui.screenHeight)
		tui.cachedOverlay.Fill(color.RGBA{0, 0, 0, 200})
	}
	screen.DrawImage(tui.cachedOverlay, nil)
}

func (tui *TerritoryUI) drawHeader(screen *ebiten.Image) int {
	titleY := 20.0
	drawTextCentered(screen, "Territory Control", tui.screenWidth/2, int(titleY), color.White)
	return 60
}

func (tui *TerritoryUI) drawCurrentTerritoryInfo(screen *ebiten.Image, yOffset int) int {
	if tui.selectedTerritory != nil {
		tui.drawTerritoryDetails(screen, &yOffset)
		return yOffset
	}

	if tui.playerEntity == nil {
		return yOffset
	}

	pos, ok := tui.playerEntity.GetComponent("position")
	if !ok {
		return yOffset
	}

	p := pos.(*PositionComponent)
	terr, err := tui.territorySystem.GetTerritoryAtPosition(p.X, p.Y)
	if err == nil && terr != nil {
		tui.selectedTerritory = terr
		tui.drawTerritoryDetails(screen, &yOffset)
	} else {
		drawText(screen, "Not in any territory", 50, yOffset, color.White)
		yOffset += 25
	}

	return yOffset
}

func (tui *TerritoryUI) drawGuildTerritories(screen *ebiten.Image, yOffset int) int {
	yOffset += 20
	drawText(screen, "Guild Territories:", 50, yOffset, color.RGBA{255, 255, 0, 255})
	yOffset += 30

	playerGuildID := tui.getPlayerGuildID()
	if playerGuildID == "" {
		drawText(screen, "Not in a guild", 70, yOffset, color.Gray{128})
		return yOffset + 25
	}

	territories := tui.territorySystem.GetManager().GetGuildTerritories(playerGuildID)
	if len(territories) == 0 {
		drawText(screen, "No territories controlled", 70, yOffset, color.Gray{128})
		return yOffset + 25
	}

	for _, terr := range territories {
		statusColor := tui.getStatusColor(terr.Status)
		text := fmt.Sprintf("%s - %s", terr.ID, terr.Status)
		drawText(screen, text, 70, yOffset, statusColor)
		yOffset += 25
	}

	return yOffset
}

func (tui *TerritoryUI) drawActiveWars(screen *ebiten.Image, yOffset int) int {
	yOffset += 20
	drawText(screen, "Active Wars:", 50, yOffset, color.RGBA{255, 100, 100, 255})
	yOffset += 30

	playerGuildID := tui.getPlayerGuildID()
	if playerGuildID == "" {
		return yOffset
	}

	wars := tui.territorySystem.GetManager().GetGuildWars(playerGuildID)
	activeWars := 0
	for _, war := range wars {
		if war.Active {
			opponent := war.DefenderGuild
			if war.DefenderGuild == playerGuildID {
				opponent = war.AttackerGuild
			}
			warText := fmt.Sprintf("War with %s (ends in %.0f days)", opponent, war.EndsAt.Sub(war.DeclaredAt).Hours()/24)
			drawText(screen, warText, 70, yOffset, color.RGBA{255, 150, 150, 255})
			yOffset += 25
			activeWars++
		}
	}

	if activeWars == 0 {
		drawText(screen, "No active wars", 70, yOffset, color.Gray{128})
		yOffset += 25
	}

	return yOffset
}

func (tui *TerritoryUI) drawControls(screen *ebiten.Image) {
	yOffset := tui.screenHeight - 80
	drawText(screen, "Controls:", 50, yOffset, color.RGBA{200, 200, 200, 255})
	yOffset += 25
	drawText(screen, "W - Declare War  |  B - Build Structure  |  ESC - Close", 50, yOffset, color.RGBA{150, 150, 150, 255})

	closeX := tui.screenWidth - 60
	closeY := 10
	drawTextCentered(screen, "[X]", closeX+25, closeY+15, color.RGBA{255, 100, 100, 255})
}

// drawTerritoryDetails draws detailed information about the selected territory.
func (tui *TerritoryUI) drawTerritoryDetails(screen *ebiten.Image, yOffset *int) {
	terr := tui.selectedTerritory

	drawText(screen, "Current Territory:", 50, *yOffset, color.RGBA{255, 255, 0, 255})
	*yOffset += 30

	// Territory ID
	drawText(screen, fmt.Sprintf("ID: %s", terr.ID), 70, *yOffset, color.White)
	*yOffset += 25

	// Status
	statusColor := tui.getStatusColor(terr.Status)
	drawText(screen, fmt.Sprintf("Status: %s", terr.Status), 70, *yOffset, statusColor)
	*yOffset += 25

	// Owner
	ownerText := "None (Neutral)"
	if terr.OwnerGuildID != "" {
		ownerText = terr.OwnerGuildID
	}
	drawText(screen, fmt.Sprintf("Owner: %s", ownerText), 70, *yOffset, color.White)
	*yOffset += 25

	// Capture progress
	if terr.Status == territory.StatusContested {
		progressText := fmt.Sprintf("Capture Progress: %.1f%% (%s)", terr.CaptureProgress*100, terr.CapturingGuild)
		drawText(screen, progressText, 70, *yOffset, color.RGBA{255, 200, 100, 255})
		*yOffset += 25
	}

	// Bonuses
	drawText(screen, fmt.Sprintf("Resource Bonus: +%.0f%%", terr.ResourceBonus*100), 70, *yOffset, color.RGBA{100, 255, 100, 255})
	*yOffset += 25
	drawText(screen, fmt.Sprintf("XP Bonus: +%.0f%%", terr.XPBonus*100), 70, *yOffset, color.RGBA{100, 200, 255, 255})
	*yOffset += 25

	// Defensive structures
	if len(terr.Structures) > 0 {
		*yOffset += 10
		drawText(screen, "Defensive Structures:", 70, *yOffset, color.RGBA{200, 200, 200, 255})
		*yOffset += 25

		for _, structure := range terr.Structures {
			hpPercent := (structure.HP / structure.MaxHP) * 100
			structText := fmt.Sprintf("  %s (%.0f%% HP)", structure.Type, hpPercent)
			drawText(screen, structText, 90, *yOffset, color.RGBA{180, 180, 255, 255})
			*yOffset += 20
		}
	}
}

// getStatusColor returns the color for a territory status.
func (tui *TerritoryUI) getStatusColor(status territory.TerritoryStatus) color.Color {
	switch status {
	case territory.StatusNeutral:
		return color.Gray{200}
	case territory.StatusOwned:
		return color.RGBA{100, 255, 100, 255}
	case territory.StatusContested:
		return color.RGBA{255, 200, 100, 255}
	default:
		return color.White
	}
}

// getPlayerGuildID returns the player's guild ID if they are in a guild.
func (tui *TerritoryUI) getPlayerGuildID() string {
	if tui.playerEntity == nil {
		return ""
	}
	guildComp, ok := tui.playerEntity.GetComponent("guild")
	if !ok {
		return ""
	}
	return guildComp.(*GuildComponent).GuildID
}

// refreshCurrentTerritory updates the selected territory based on player position.
func (tui *TerritoryUI) refreshCurrentTerritory() {
	if tui.playerEntity == nil {
		return
	}
	pos, ok := tui.playerEntity.GetComponent("position")
	if !ok {
		return
	}
	p := pos.(*PositionComponent)
	terr, err := tui.territorySystem.GetTerritoryAtPosition(p.X, p.Y)
	if err == nil && terr != nil {
		tui.selectedTerritory = terr
	}
}

// handleDeclareWar handles the war declaration action.
func (tui *TerritoryUI) handleDeclareWar() {
	if tui.selectedTerritory == nil || tui.playerEntity == nil {
		return
	}

	playerGuildID := tui.getPlayerGuildID()
	if playerGuildID == "" {
		return
	}

	targetGuildID := tui.selectedTerritory.OwnerGuildID
	if targetGuildID == "" || targetGuildID == playerGuildID {
		return
	}

	// Check if already at war
	if tui.territorySystem.GetManager().IsAtWar(playerGuildID, targetGuildID) {
		return
	}

	// Declare war
	_, err := tui.territorySystem.GetManager().DeclareWar(playerGuildID, targetGuildID)
	if err != nil {
		log.WithFields(log.Fields{
			"attacker_guild": playerGuildID,
			"defender_guild": targetGuildID,
			"error":          err,
		}).Warn("failed to declare war from territory UI")
		return
	}
}

// handleBuildStructure handles building a defensive structure.
func (tui *TerritoryUI) handleBuildStructure() {
	if tui.selectedTerritory == nil || tui.playerEntity == nil {
		return
	}

	playerGuildID := tui.getPlayerGuildID()
	if playerGuildID == "" {
		return
	}

	// Can only build in owned territory
	if tui.selectedTerritory.OwnerGuildID != playerGuildID {
		return
	}

	// Get player position for structure placement
	pos, ok := tui.playerEntity.GetComponent("position")
	if !ok {
		return
	}
	p := pos.(*PositionComponent)

	// Build a wall (default structure)
	_, err := tui.territorySystem.GetManager().BuildDefensiveStructure(
		tui.selectedTerritory.ID,
		territory.StructureTypeWall,
		p.X,
		p.Y,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"territory_id": tui.selectedTerritory.ID,
			"guild_id":     playerGuildID,
			"x":            p.X,
			"y":            p.Y,
			"error":        err,
		}).Warn("failed to build structure from territory UI")
		return
	}

	// Refresh territory data
	tui.refreshCurrentTerritory()
}

// Helper functions for text rendering
func drawText(screen *ebiten.Image, txt string, x, y int, clr color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(clr)
	text.Draw(screen, txt, &text.GoTextFace{Size: 16}, op)
}

func drawTextCentered(screen *ebiten.Image, txt string, x, y int, clr color.Color) {
	// Approximate centering (exact would require measuring text width)
	approxWidth := len(txt) * 8
	drawText(screen, txt, x-approxWidth/2, y, clr)
}
