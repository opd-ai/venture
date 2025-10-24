// Package engine provides quest_ui for game UI.
package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// QuestUI handles rendering and interaction for the quest log.
type EbitenQuestUI struct {
	visible      bool
	world        *World
	playerEntity *Entity

	// Layout
	screenWidth  int
	screenHeight int

	// Tab selection
	currentTab int // 0 = Active, 1 = Completed
}

// NewQuestUI creates a new quest UI.
func NewEbitenQuestUI(world *World, screenWidth, screenHeight int) *EbitenQuestUI {
	return &EbitenQuestUI{
		visible:      false,
		world:        world,
		screenWidth:  screenWidth,
		screenHeight: screenHeight,
		currentTab:   0,
	}
}

// SetPlayerEntity sets the player entity whose quests to display.
func (ui *EbitenQuestUI) SetPlayerEntity(entity *Entity) {
	ui.playerEntity = entity
}

// Toggle shows or hides the quest UI.
func (ui *EbitenQuestUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the quest log is currently shown.
func (ui *EbitenQuestUI) IsVisible() bool {
	return ui.visible
}

// Show displays the quest UI.
func (ui *EbitenQuestUI) Show() {
	ui.visible = true
}

// Hide hides the quest UI.
func (ui *EbitenQuestUI) Hide() {
	ui.visible = false
}

// Update processes input for the quest UI.
func (ui *EbitenQuestUI) Update(entities []*Entity, deltaTime float64) {
	// Always check for toggle key, even when not visible
	if inpututil.IsKeyJustPressed(ebiten.KeyJ) {
		ui.Toggle()
		return // Don't process other input on the same frame as toggle
	}

	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Handle tab switching
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		ui.currentTab = 0 // Active
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		ui.currentTab = 1 // Completed
	}
}

// Draw renders the quest UI.
func (ui *EbitenQuestUI) Draw(screen interface{}) {
	img, ok := screen.(*ebiten.Image)
	if !ok {
		return
	}
	if !ui.visible || ui.playerEntity == nil {
		return
	}

	// Get quest tracker component
	trackerComp, ok := ui.playerEntity.GetComponent("questtracker")
	if !ok {
		return
	}
	tracker := trackerComp.(*QuestTrackerComponent)

	// Draw semi-transparent overlay
	overlay := ebiten.NewImage(ui.screenWidth, ui.screenHeight)
	overlay.Fill(color.RGBA{0, 0, 0, 180})
	img.DrawImage(overlay, nil)

	// Calculate window position
	windowWidth := 600
	windowHeight := 500
	windowX := (ui.screenWidth - windowWidth) / 2
	windowY := (ui.screenHeight - windowHeight) / 2

	// Draw window background
	windowBg := ebiten.NewImage(windowWidth, windowHeight)
	windowBg.Fill(color.RGBA{40, 40, 50, 255})
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(windowX), float64(windowY))
	img.DrawImage(windowBg, opts)

	// Draw title
	ebitenutil.DebugPrintAt(img, "QUEST LOG", windowX+10, windowY+10)

	// Draw tabs
	tabY := windowY + 40
	tabs := []string{"Active", "Completed"}
	for i, tabName := range tabs {
		tabX := windowX + 10 + i*100
		tabColor := color.RGBA{60, 60, 70, 255}
		if i == ui.currentTab {
			tabColor = color.RGBA{80, 80, 100, 255}
		}

		tabBg := ebiten.NewImage(90, 30)
		tabBg.Fill(tabColor)
		tabOpts := &ebiten.DrawImageOptions{}
		tabOpts.GeoM.Translate(float64(tabX), float64(tabY))
		img.DrawImage(tabBg, tabOpts)

		ebitenutil.DebugPrintAt(img, tabName, tabX+10, tabY+10)
	}

	// Draw quest list based on current tab
	listY := tabY + 40
	var quests []*TrackedQuest
	if ui.currentTab == 0 {
		quests = tracker.ActiveQuests
	} else {
		quests = tracker.CompletedQuests
	}

	if len(quests) == 0 {
		ebitenutil.DebugPrintAt(img, "No quests", windowX+20, listY+20)
	} else {
		y := listY + 10
		for _, tracked := range quests {
			// Draw quest name
			ebitenutil.DebugPrintAt(img, tracked.Quest.Name, windowX+20, y)
			y += 20

			// Draw quest type and difficulty
			info := fmt.Sprintf("%s | %s", tracked.Quest.Type.String(), tracked.Quest.Difficulty.String())
			ebitenutil.DebugPrintAt(img, info, windowX+30, y)
			y += 20

			// Draw objectives
			for i, obj := range tracked.Quest.Objectives {
				progress := fmt.Sprintf("  [%d/%d] %s", obj.Current, obj.Required, obj.Description)
				ebitenutil.DebugPrintAt(img, progress, windowX+30, y)
				y += 15

				// Draw progress bar
				barWidth := 200
				barHeight := 8
				barX := windowX + 240
				barY := y - 10

				// Background
				barBg := ebiten.NewImage(barWidth, barHeight)
				barBg.Fill(color.RGBA{60, 60, 70, 255})
				barOpts := &ebiten.DrawImageOptions{}
				barOpts.GeoM.Translate(float64(barX), float64(barY))
				img.DrawImage(barBg, barOpts)

				// Progress
				progressPct := obj.Progress()
				if progressPct > 0 {
					progressWidth := int(float64(barWidth) * progressPct)
					barFill := ebiten.NewImage(progressWidth, barHeight)
					fillColor := color.RGBA{80, 180, 80, 255}
					if obj.IsComplete() {
						fillColor = color.RGBA{100, 220, 100, 255}
					}
					barFill.Fill(fillColor)
					img.DrawImage(barFill, barOpts)
				}

				if i < len(tracked.Quest.Objectives)-1 {
					y += 5
				}
			}

			// Draw rewards
			y += 20
			rewards := fmt.Sprintf("  Rewards: %d XP, %d Gold", tracked.Quest.Reward.XP, tracked.Quest.Reward.Gold)
			ebitenutil.DebugPrintAt(img, rewards, windowX+30, y)
			y += 30

			if y > windowY+windowHeight-40 {
				break // Don't overflow window
			}
		}
	}

	// Draw controls hint
	controlsY := windowY + windowHeight - 20
	ebitenutil.DebugPrintAt(img, "J: Close | 1: Active | 2: Completed", windowX+10, controlsY)
}

// IsActive returns whether the quest UI is currently visible.
// Implements UISystem interface.
func (q *EbitenQuestUI) IsActive() bool {
return q.visible
}

// SetActive sets whether the quest UI is visible.
// Implements UISystem interface.
func (q *EbitenQuestUI) SetActive(active bool) {
q.visible = active
}

// Compile-time check that EbitenQuestUI implements UISystem
var _ UISystem = (*EbitenQuestUI)(nil)
