package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/opd-ai/venture/pkg/world/territory"
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
	touchHandler      *TouchHandler
}

// NewTerritoryUI creates a new territory UI.
func NewTerritoryUI(territorySystem *TerritorySystem, screenWidth, screenHeight int) *TerritoryUI {
	return &TerritoryUI{
		visible:         false,
		territorySystem: territorySystem,
		screenWidth:     screenWidth,
		screenHeight:    screenHeight,
		scrollOffset:    0,
		touchHandler:    NewTouchHandler(),
	}
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

// Update processes input for the territory UI.
func (tui *TerritoryUI) Update() error {
	if !tui.visible {
		return nil
	}

	// Keyboard navigation
	if ebiten.IsKeyPressed(ebiten.KeyUp) {
		tui.scrollOffset--
		if tui.scrollOffset < 0 {
			tui.scrollOffset = 0
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) {
		tui.scrollOffset++
	}

	// Declare war (W key)
	if ebiten.IsKeyPressed(ebiten.KeyW) && tui.selectedTerritory != nil {
		tui.handleDeclareWar()
	}

	// Build structure (B key)
	if ebiten.IsKeyPressed(ebiten.KeyB) && tui.selectedTerritory != nil {
		tui.handleBuildStructure()
	}

	// Touch input
	touches := ebiten.AppendTouchIDs(nil)
	if len(touches) > 0 {
		for _, touchID := range touches {
			x, y := ebiten.TouchPosition(touchID)
			tui.touchHandler.HandleTouch(touchID, float64(x), float64(y))

			// Close button area (top right)
			closeX := tui.screenWidth - 60
			closeY := 10
			if x >= closeX && x <= closeX+50 && y >= closeY && y <= closeY+30 {
				tui.Hide()
			}
		}
	}

	return nil
}

// Draw renders the territory UI.
func (tui *TerritoryUI) Draw(screen *ebiten.Image) {
	if !tui.visible {
		return
	}

	// Semi-transparent background
	bg := ebiten.NewImage(tui.screenWidth, tui.screenHeight)
	bg.Fill(color.RGBA{0, 0, 0, 200})
	screen.DrawImage(bg, nil)

	// Title
	titleY := 20.0
	drawTextCentered(screen, "Territory Control", tui.screenWidth/2, int(titleY), color.White)

	// Current territory info
	yOffset := 60
	if tui.selectedTerritory != nil {
		tui.drawTerritoryDetails(screen, &yOffset)
	} else if tui.playerEntity != nil {
		// Show player's current location territory
		pos := tui.playerEntity.GetComponent("position")
		if pos != nil {
			p := pos.(*PositionComponent)
			terr, err := tui.territorySystem.GetTerritoryAtPosition(p.X, p.Y)
			if err == nil && terr != nil {
				tui.selectedTerritory = terr
				tui.drawTerritoryDetails(screen, &yOffset)
			} else {
				drawText(screen, "Not in any territory", 50, yOffset, color.White)
				yOffset += 25
			}
		}
	}

	// Guild territories section
	yOffset += 20
	drawText(screen, "Guild Territories:", 50, yOffset, color.RGBA{255, 255, 0, 255})
	yOffset += 30

	playerGuildID := tui.getPlayerGuildID()
	if playerGuildID != "" {
		territories := tui.territorySystem.GetManager().GetGuildTerritories(playerGuildID)
		if len(territories) > 0 {
			for _, terr := range territories {
				statusColor := tui.getStatusColor(terr.Status)
				text := fmt.Sprintf("%s - %s", terr.ID, terr.Status)
				drawText(screen, text, 70, yOffset, statusColor)
				yOffset += 25
			}
		} else {
			drawText(screen, "No territories controlled", 70, yOffset, color.Gray{128})
			yOffset += 25
		}
	} else {
		drawText(screen, "Not in a guild", 70, yOffset, color.Gray{128})
		yOffset += 25
	}

	// Active wars section
	yOffset += 20
	drawText(screen, "Active Wars:", 50, yOffset, color.RGBA{255, 100, 100, 255})
	yOffset += 30

	if playerGuildID != "" {
		wars := tui.territorySystem.GetManager().GetGuildWars(playerGuildID)
		activeWars := 0
		for _, war := range wars {
			if war.Active {
				opponent := war.DefenderGuild
				if war.DefenderGuild == playerGuildID {
					opponent = war.AttackerGuild
				}
				warText := fmt.Sprintf("War with %s (ends in %d days)", opponent, war.EndsAt.Sub(war.DeclaredAt).Hours()/24)
				drawText(screen, warText, 70, yOffset, color.RGBA{255, 150, 150, 255})
				yOffset += 25
				activeWars++
			}
		}
		if activeWars == 0 {
			drawText(screen, "No active wars", 70, yOffset, color.Gray{128})
			yOffset += 25
		}
	}

	// Controls help
	yOffset = tui.screenHeight - 80
	drawText(screen, "Controls:", 50, yOffset, color.RGBA{200, 200, 200, 255})
	yOffset += 25
	drawText(screen, "W - Declare War  |  B - Build Structure  |  ESC - Close", 50, yOffset, color.RGBA{150, 150, 150, 255})

	// Close button for touch
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
	guildComp := tui.playerEntity.GetComponent("guild")
	if guildComp == nil {
		return ""
	}
	return guildComp.(*GuildComponent).GuildID
}

// refreshCurrentTerritory updates the selected territory based on player position.
func (tui *TerritoryUI) refreshCurrentTerritory() {
	if tui.playerEntity == nil {
		return
	}
	pos := tui.playerEntity.GetComponent("position")
	if pos == nil {
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
		// Could log error, but UI doesn't have logger
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
	pos := tui.playerEntity.GetComponent("position")
	if pos == nil {
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
