// health bars, stats, and other UI elements.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// TerritoryBonusProvider provides territory bonus information for display.
// Implemented by TerritorySystem to allow HUD to show bonuses.
type TerritoryBonusProvider interface {
	// GetBonusesForGuild returns resource and XP bonuses for a guild.
	GetBonusesForGuild(guildID string) (resourceBonus, xpBonus float64)
}

// HUDSystem renders the heads-up display (health bars, stats, etc).
type EbitenHUDSystem struct {
	screen       *ebiten.Image
	screenWidth  int
	screenHeight int

	// HUD visibility
	Visible bool

	// Player entity to display stats for
	playerEntity *Entity

	// Network client for displaying latency (optional, only in multiplayer)
	networkClient NetworkClient

	// Territory bonus provider for displaying guild territory bonuses
	territoryBonusProvider TerritoryBonusProvider
}

// NewEbitenHUDSystem creates a new HUD system.
func NewEbitenHUDSystem(screenWidth, screenHeight int) *EbitenHUDSystem {
	return &EbitenHUDSystem{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		Visible:      true,
	}
}

// SetPlayerEntity sets the player entity whose stats will be displayed.
func (h *EbitenHUDSystem) SetPlayerEntity(entity *Entity) {
	h.playerEntity = entity
}

// SetNetworkClient sets the network client for displaying latency.
// Pass nil to disable network status display.
func (h *EbitenHUDSystem) SetNetworkClient(client NetworkClient) {
	h.networkClient = client
}

// SetTerritoryBonusProvider sets the territory bonus provider for displaying guild bonuses.
// Pass nil to disable territory bonus display.
func (h *EbitenHUDSystem) SetTerritoryBonusProvider(provider TerritoryBonusProvider) {
	h.territoryBonusProvider = provider
}

// Update is called every frame but HUD doesn't need to update entities.
func (h *EbitenHUDSystem) Update(entities []*Entity, deltaTime float64) {
	// HUD doesn't modify entities, just reads their state
}

// Draw renders the HUD overlay on the screen.
// Implements UISystem interface.
func (h *EbitenHUDSystem) Draw(screen interface{}) {
	if !h.Visible || h.playerEntity == nil {
		return
	}

	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	h.screen = img

	// Draw health bar
	h.drawHealthBar()

	// Draw mana bar below health bar (G24 fix)
	h.drawManaBar()

	// Draw stats panel
	h.drawStatsPanel()

	// Draw experience bar
	h.drawExperienceBar()

	// Note: Aim indicator is now drawn by EbitenRenderSystem (below player sprite layer)
	// Draw network status if in multiplayer mode
	h.drawNetworkStatus()

	// Draw territory bonuses if available
	h.drawTerritoryBonuses()
}

// drawHealthBar draws the player's health bar at the top left.
func (h *EbitenHUDSystem) drawHealthBar() {
	healthComp, ok := h.playerEntity.GetComponent("health")
	if !ok {
		return
	}
	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return
	}

	// Health bar dimensions
	barX := float32(20)
	barY := float32(20)
	barWidth := float32(200)
	barHeight := float32(20)

	// Background (dark gray)
	vector.DrawFilledRect(h.screen, barX, barY, barWidth, barHeight,
		color.RGBA{40, 40, 40, 255}, false)

	// Health fill (red to green based on health %)
	if health.Max == 0 {
		return
	}
	// G27 fix: use explicit float32 division and clamp to [0,1] to prevent
	// overheal from drawing the fill outside the background rectangle.
	healthPct := float32(health.Current) / float32(health.Max)
	if healthPct > 1.0 {
		healthPct = 1.0
	}
	if healthPct < 0.0 {
		healthPct = 0.0
	}
	fillWidth := barWidth * healthPct

	healthColor := h.getHealthColor(healthPct)
	vector.DrawFilledRect(h.screen, barX, barY, fillWidth, barHeight,
		healthColor, false)

	// Border
	vector.StrokeRect(h.screen, barX, barY, barWidth, barHeight, 2,
		color.RGBA{255, 255, 255, 255}, false)

	// Health text
	healthText := fmt.Sprintf("%.0f / %.0f", health.Current, health.Max)
	h.drawText(healthText, int(barX+barWidth/2-30), int(barY+5), color.White)
}

// drawManaBar draws the player's mana bar below the health bar (G24 fix).
// Mirrors the health bar layout but uses a blue fill.
func (h *EbitenHUDSystem) drawManaBar() {
	manaComp, ok := h.playerEntity.GetComponent("mana")
	if !ok {
		return
	}
	mana, ok := manaComp.(*ManaComponent)
	if !ok {
		return
	}

	barX := float32(20)
	barY := float32(46) // Directly below the health bar (20 + 20 height + 6 gap)
	barWidth := float32(200)
	barHeight := float32(14)

	// Background (dark blue-gray)
	vector.DrawFilledRect(h.screen, barX, barY, barWidth, barHeight,
		color.RGBA{20, 20, 60, 255}, false)

	if mana.Max == 0 {
		return
	}
	manaPct := float32(mana.Current) / float32(mana.Max)
	if manaPct > 1.0 {
		manaPct = 1.0
	}
	if manaPct < 0.0 {
		manaPct = 0.0
	}

	// Blue fill
	vector.DrawFilledRect(h.screen, barX, barY, barWidth*manaPct, barHeight,
		color.RGBA{60, 120, 255, 255}, false)

	// Border
	vector.StrokeRect(h.screen, barX, barY, barWidth, barHeight, 1,
		color.RGBA{180, 180, 255, 200}, false)

	// Mana text
	manaText := fmt.Sprintf("%d / %d", mana.Current, mana.Max)
	h.drawText(manaText, int(barX+barWidth/2-20), int(barY+2), color.RGBA{200, 220, 255, 255})
}

// drawStatsPanel draws the player's stats in the top right.
func (h *EbitenHUDSystem) drawStatsPanel() {
	statsComp, hasStats := h.playerEntity.GetComponent("stats")
	expComp, hasExp := h.playerEntity.GetComponent("experience")

	if !hasStats && !hasExp {
		return
	}

	x := h.screenWidth - 200
	y := 20
	lineHeight := 20

	// Draw background panel
	panelWidth := float32(180)
	panelHeight := float32(100)
	vector.DrawFilledRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, color.RGBA{20, 20, 30, 200}, false)
	vector.StrokeRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, 2, color.RGBA{255, 255, 255, 128}, false)

	// Draw level if available
	if hasExp {
		exp, ok := expComp.(*ExperienceComponent)
		if !ok {
			hasExp = false
		} else {
			levelText := fmt.Sprintf("Level: %d", exp.Level)
			h.drawText(levelText, x, y, color.White)
			y += lineHeight
		}
	}

	// Draw stats if available
	if hasStats {
		stats, ok := statsComp.(*StatsComponent)
		if !ok {
			hasStats = false
		} else {
			h.drawText(fmt.Sprintf("ATK: %.0f", stats.Attack), x, y, color.RGBA{255, 200, 200, 255})
			y += lineHeight
			h.drawText(fmt.Sprintf("DEF: %.0f", stats.Defense), x, y, color.RGBA{200, 200, 255, 255})
			y += lineHeight
			h.drawText(fmt.Sprintf("MAG: %.0f", stats.MagicPower), x, y, color.RGBA{200, 255, 200, 255})
		}
	}
}

// drawExperienceBar draws the experience progress bar at the bottom.
func (h *EbitenHUDSystem) drawExperienceBar() {
	expComp, ok := h.playerEntity.GetComponent("experience")
	if !ok {
		return
	}
	exp, ok := expComp.(*ExperienceComponent)
	if !ok {
		return
	}

	// Experience bar dimensions
	barX := float32(20)
	barY := float32(h.screenHeight - 40)
	barWidth := float32(300)
	barHeight := float32(15)

	// Background
	vector.DrawFilledRect(h.screen, barX, barY, barWidth, barHeight,
		color.RGBA{40, 40, 40, 255}, false)

	// Experience fill
	expPct := float32(exp.ProgressToNextLevel())
	fillWidth := barWidth * expPct
	vector.DrawFilledRect(h.screen, barX, barY, fillWidth, barHeight,
		color.RGBA{100, 200, 255, 255}, false)

	// Border
	vector.StrokeRect(h.screen, barX, barY, barWidth, barHeight, 2,
		color.RGBA{255, 255, 255, 255}, false)

	// XP text
	xpText := fmt.Sprintf("XP: %d / %d", exp.CurrentXP, exp.RequiredXP)
	h.drawText(xpText, int(barX+barWidth/2-40), int(barY+2), color.White)
}

// getHealthColor returns a color based on health percentage.
func (h *EbitenHUDSystem) getHealthColor(healthPct float32) color.Color {
	if healthPct > 0.75 {
		// 100%-75%: Pure green to slight yellow tint
		// R increases from 0 to 100
		redAmount := uint8((1.0 - healthPct) * 4.0 * 100) // 0 at 100%, 100 at 75%
		return color.RGBA{R: redAmount, G: 200, B: 0, A: 255}
	} else if healthPct > 0.5 {
		// 75%-50%: Yellow-green to yellow
		// R increases from 100 to 255
		redAmount := uint8(100 + ((0.75 - healthPct) * 4.0 * 155)) // 100 at 75%, 255 at 50%
		return color.RGBA{R: redAmount, G: 200, B: 0, A: 255}
	} else if healthPct > 0.25 {
		// 50%-25%: Yellow to orange
		// G decreases from 200 to 180
		greenAmount := uint8(200 - ((0.5 - healthPct) * 4.0 * 20)) // 200 at 50%, 180 at 25%
		return color.RGBA{R: 255, G: greenAmount, B: 0, A: 255}
	} else {
		// <25%: Orange to red
		// G decreases from 180 to 50 (minimum)
		greenAmount := uint8(180 * (healthPct * 4.0)) // 180 at 25%, 0 at 0%
		if greenAmount < 50 {
			greenAmount = 50 // Minimum green for visibility
		}
		return color.RGBA{R: 220, G: greenAmount, B: 50, A: 255}
	}
}

// drawText draws text at the specified position using basicfont.
// This provides readable text for HUD elements (health values, stats, XP).
func (h *EbitenHUDSystem) drawText(str string, x, y int, col color.Color) {
	// Use basicfont.Face7x13 for consistent text rendering across all UI systems
	// Note: y coordinate is the baseline, not top-left, so text appears below y
	text.Draw(h.screen, str, basicfont.Face7x13, x, y+13, col)
}

// drawAimIndicator draws a crosshair showing the player's aim direction.
// Phase 10.1: Visual feedback for 360° mouse aim system.
func (h *EbitenHUDSystem) drawAimIndicator() {
	// Get player aim component
	aimComp, ok := h.playerEntity.GetComponent("aim")
	if !ok {
		return // No aim component, skip indicator
	}
	aim, ok := aimComp.(*AimComponent)
	if !ok {
		return
	}

	// Draw direction arrow from player center (screen center since camera follows player)
	// Calculate endpoint 60 pixels away in aim direction
	dirX, dirY := aim.GetAimDirection()
	arrowLength := float32(60.0)

	// Center of screen (player is always centered)
	centerX := float32(h.screenWidth / 2)
	centerY := float32(h.screenHeight / 2)

	endX := centerX + float32(dirX)*arrowLength
	endY := centerY + float32(dirY)*arrowLength

	// Draw aim line (semi-transparent white)
	vector.StrokeLine(h.screen, centerX, centerY, endX, endY, 2,
		color.RGBA{255, 255, 255, 128}, false)

	// Draw arrowhead
	arrowSize := float32(8.0)
	perpX := -float32(dirY) // Perpendicular vector
	perpY := float32(dirX)

	// Three points of the arrowhead triangle
	tipX := endX
	tipY := endY
	left1X := tipX - float32(dirX)*arrowSize + perpX*arrowSize*0.5
	left1Y := tipY - float32(dirY)*arrowSize + perpY*arrowSize*0.5
	left2X := tipX - float32(dirX)*arrowSize - perpX*arrowSize*0.5
	left2Y := tipY - float32(dirY)*arrowSize - perpY*arrowSize*0.5

	// Draw filled triangle for arrowhead
	vector.DrawFilledCircle(h.screen, tipX, tipY, 3, color.RGBA{255, 255, 255, 180}, false)
	vector.StrokeLine(h.screen, left1X, left1Y, tipX, tipY, 2,
		color.RGBA{255, 255, 255, 180}, false)
	vector.StrokeLine(h.screen, left2X, left2Y, tipX, tipY, 2,
		color.RGBA{255, 255, 255, 180}, false)
}

// IsActive returns whether the HUD is currently visible.
// Implements UISystem interface.
func (h *EbitenHUDSystem) IsActive() bool {
	return h.Visible
}

// SetActive sets whether the HUD is visible.
// Implements UISystem interface.
func (h *EbitenHUDSystem) SetActive(active bool) {
	h.Visible = active
}

// drawNetworkStatus displays network latency and connection quality in multiplayer mode.
// Shown in the top-right corner below the stats panel.
// Color-coded: Green (<100ms), Yellow (100-300ms), Orange (300-1000ms), Red (>1000ms).
func (h *EbitenHUDSystem) drawNetworkStatus() {
	// Only draw if network client is configured (multiplayer mode)
	if h.networkClient == nil {
		return
	}

	// Only draw if screen is initialized (not in test environment)
	if h.screen == nil {
		return
	}

	// Check if connected
	if !h.networkClient.IsConnected() {
		return
	}

	// Get current latency
	latency := h.networkClient.GetLatency()
	latencyMs := latency.Milliseconds()

	// Position below stats panel in top-right
	x := h.screenWidth - 200
	y := 130 // Below stats panel (which ends around y=100)
	lineHeight := 20

	// Draw background panel
	panelWidth := float32(180)
	panelHeight := float32(40)
	vector.DrawFilledRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, color.RGBA{20, 20, 30, 200}, false)
	vector.StrokeRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, 2, color.RGBA{255, 255, 255, 128}, false)

	// Draw "Network" label
	h.drawText("Network:", x, y, color.White)
	y += lineHeight

	// Determine connection quality color based on latency
	// Green: <100ms (excellent), Yellow: 100-300ms (good),
	// Orange: 300-1000ms (fair), Red: >1000ms (poor)
	var qualityColor color.Color
	var qualityText string

	if latencyMs < 100 {
		qualityColor = color.RGBA{100, 255, 100, 255} // Green
		qualityText = "Excellent"
	} else if latencyMs < 300 {
		qualityColor = color.RGBA{255, 255, 100, 255} // Yellow
		qualityText = "Good"
	} else if latencyMs < 1000 {
		qualityColor = color.RGBA{255, 180, 100, 255} // Orange
		qualityText = "Fair"
	} else {
		qualityColor = color.RGBA{255, 100, 100, 255} // Red
		qualityText = "Poor"
	}

	// Draw latency with quality indicator
	latencyText := fmt.Sprintf("%dms (%s)", latencyMs, qualityText)
	h.drawText(latencyText, x, y, qualityColor)
}

// drawTerritoryBonuses displays the player's guild territory bonuses.
// Shown below network status in top-right corner (or below stats if no network status).
// Displays resource and XP multiplier bonuses from controlled territories.
func (h *EbitenHUDSystem) drawTerritoryBonuses() {
	// Only draw if territory bonus provider is configured
	if h.territoryBonusProvider == nil {
		return
	}

	// Only draw if screen is initialized (not in test environment)
	if h.screen == nil {
		return
	}

	// Get player's guild ID from guild component
	guildComp, ok := h.playerEntity.GetComponent("guild")
	if !ok {
		return // Player not in a guild
	}
	guild, ok := guildComp.(*GuildComponent)
	if !ok || guild.GuildID == "" {
		return // No valid guild
	}

	// Get bonuses from the provider
	resourceBonus, xpBonus := h.territoryBonusProvider.GetBonusesForGuild(guild.GuildID)

	// Don't show if no bonuses
	if resourceBonus == 0 && xpBonus == 0 {
		return
	}

	// Position below network status (or below stats if no network)
	x := h.screenWidth - 200
	y := 180 // Below network status panel
	if h.networkClient == nil || !h.networkClient.IsConnected() {
		y = 130 // Position at network status location if no network
	}
	lineHeight := 18

	// Draw background panel
	panelWidth := float32(180)
	panelHeight := float32(55)
	vector.DrawFilledRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, color.RGBA{20, 40, 20, 200}, false)
	vector.StrokeRect(h.screen, float32(x-10), float32(y-5),
		panelWidth, panelHeight, 2, color.RGBA{100, 200, 100, 128}, false)

	// Draw "Territory Bonuses" label
	h.drawText("Territory Bonuses:", x, y, color.RGBA{100, 255, 100, 255})
	y += lineHeight

	// Draw resource bonus (gold/green tint)
	resourceText := fmt.Sprintf("Resources: +%.0f%%", resourceBonus*100)
	h.drawText(resourceText, x, y, color.RGBA{255, 215, 0, 255})
	y += lineHeight

	// Draw XP bonus (blue/cyan tint)
	xpText := fmt.Sprintf("XP: +%.0f%%", xpBonus*100)
	h.drawText(xpText, x, y, color.RGBA{100, 200, 255, 255})
}

// Compile-time check that EbitenHUDSystem implements UISystem
var _ UISystem = (*EbitenHUDSystem)(nil)
