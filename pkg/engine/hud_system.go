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

// HUDSystem renders the heads-up display (health bars, stats, etc).
type EbitenHUDSystem struct {
	screen       *ebiten.Image
	screenWidth  int
	screenHeight int

	// HUD visibility
	Visible bool

	// Player entity to display stats for
	playerEntity *Entity
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

	// Draw stats panel
	h.drawStatsPanel()

	// Draw experience bar
	h.drawExperienceBar()
}

// drawHealthBar draws the player's health bar at the top left.
func (h *EbitenHUDSystem) drawHealthBar() {
	healthComp, ok := h.playerEntity.GetComponent("health")
	if !ok {
		return
	}
	health := healthComp.(*HealthComponent)

	// Health bar dimensions
	barX := float32(20)
	barY := float32(20)
	barWidth := float32(200)
	barHeight := float32(20)

	// Background (dark gray)
	vector.DrawFilledRect(h.screen, barX, barY, barWidth, barHeight,
		color.RGBA{40, 40, 40, 255}, false)

	// Health fill (red to green based on health %)
	healthPct := float32(health.Current / health.Max)
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
		exp := expComp.(*ExperienceComponent)
		levelText := fmt.Sprintf("Level: %d", exp.Level)
		h.drawText(levelText, x, y, color.White)
		y += lineHeight
	}

	// Draw stats if available
	if hasStats {
		stats := statsComp.(*StatsComponent)
		h.drawText(fmt.Sprintf("ATK: %.0f", stats.Attack), x, y, color.RGBA{255, 200, 200, 255})
		y += lineHeight
		h.drawText(fmt.Sprintf("DEF: %.0f", stats.Defense), x, y, color.RGBA{200, 200, 255, 255})
		y += lineHeight
		h.drawText(fmt.Sprintf("MAG: %.0f", stats.MagicPower), x, y, color.RGBA{200, 255, 200, 255})
	}
}

// drawExperienceBar draws the experience progress bar at the bottom.
func (h *EbitenHUDSystem) drawExperienceBar() {
	expComp, ok := h.playerEntity.GetComponent("experience")
	if !ok {
		return
	}
	exp := expComp.(*ExperienceComponent)

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

// Compile-time check that EbitenHUDSystem implements UISystem
var _ UISystem = (*EbitenHUDSystem)(nil)
