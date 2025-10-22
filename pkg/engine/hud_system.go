//go:build !test
// +build !test

// Package engine provides HUD rendering for game UI.
// This file implements HUDSystem which renders the heads-up display including
// health bars, stats, and other UI elements.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// HUDSystem renders the heads-up display (health bars, stats, etc).
type HUDSystem struct {
	screen       *ebiten.Image
	screenWidth  int
	screenHeight int

	// Font for text rendering (if available)
	fontFace text.Face

	// HUD visibility
	Visible bool

	// Player entity to display stats for
	playerEntity *Entity
}

// NewHUDSystem creates a new HUD system.
func NewHUDSystem(screenWidth, screenHeight int) *HUDSystem {
	return &HUDSystem{
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		Visible:      true,
	}
}

// SetPlayerEntity sets the player entity whose stats will be displayed.
func (h *HUDSystem) SetPlayerEntity(entity *Entity) {
	h.playerEntity = entity
}

// Update is called every frame but HUD doesn't need to update entities.
func (h *HUDSystem) Update(entities []*Entity, deltaTime float64) {
	// HUD doesn't modify entities, just reads their state
}

// Draw renders the HUD overlay on the screen.
func (h *HUDSystem) Draw(screen *ebiten.Image) {
	if !h.Visible || h.playerEntity == nil {
		return
	}

	h.screen = screen

	// Draw health bar
	h.drawHealthBar()

	// Draw stats panel
	h.drawStatsPanel()

	// Draw experience bar
	h.drawExperienceBar()
}

// drawHealthBar draws the player's health bar at the top left.
func (h *HUDSystem) drawHealthBar() {
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
func (h *HUDSystem) drawStatsPanel() {
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
func (h *HUDSystem) drawExperienceBar() {
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
func (h *HUDSystem) getHealthColor(healthPct float32) color.Color {
	if healthPct > 0.6 {
		// Green to yellow
		return color.RGBA{
			R: uint8((1.0 - healthPct) * 255 * 2.5),
			G: 200,
			B: 0,
			A: 255,
		}
	} else if healthPct > 0.3 {
		// Yellow to orange
		return color.RGBA{R: 255, G: 180, B: 0, A: 255}
	} else {
		// Red
		return color.RGBA{R: 220, G: 50, B: 50, A: 255}
	}
}

// drawText draws text at the specified position.
// This is a simple fallback implementation without proper font rendering.
func (h *HUDSystem) drawText(str string, x, y int, col color.Color) {
	// Note: This uses ebiten's debug text which is very basic
	// In a real implementation, you'd use ebitengine/text with proper fonts
	// For now, we'll skip text rendering to keep it simple
	// The bars and visual elements are the main HUD features
}
