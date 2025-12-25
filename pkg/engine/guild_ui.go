//go:build !headless
// +build !headless

package engine

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/opd-ai/venture/pkg/mobile"
	"github.com/opd-ai/venture/pkg/network/federation/guild"
)

// GuildUI provides UI for guild management
type GuildUI struct {
	world        *World
	guildSystem  *GuildSystem
	visible      bool
	width        int
	height       int
	touchHandler *mobile.TouchInputHandler

	// UI state
	selectedTab  int // 0=info, 1=members, 2=treasury
	scrollOffset int
}

// NewGuildUI creates a new guild UI
func NewGuildUI(world *World, guildSystem *GuildSystem, width, height int) *GuildUI {
	return &GuildUI{
		world:        world,
		guildSystem:  guildSystem,
		visible:      false,
		width:        width,
		height:       height,
		touchHandler: mobile.NewTouchInputHandler(),
		selectedTab:  0,
		scrollOffset: 0,
	}
}

// Toggle toggles the guild UI visibility
func (ui *GuildUI) Toggle() {
	ui.visible = !ui.visible
}

// IsVisible returns whether the UI is currently visible
func (ui *GuildUI) IsVisible() bool {
	return ui.visible
}

// Update updates the guild UI state
func (ui *GuildUI) Update() error {
	if !ui.visible {
		return nil
	}

	if ui.handleKeyboardInput() {
		return nil
	}

	ui.handleScrolling()
	ui.handleTouchInput()

	return nil
}

// handleKeyboardInput processes keyboard input and returns true if UI should close.
func (ui *GuildUI) handleKeyboardInput() bool {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		ui.visible = false
		return true
	}

	ui.handleTabSwitching()
	return false
}

// handleTabSwitching processes tab switching key presses.
func (ui *GuildUI) handleTabSwitching() {
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		ui.selectedTab = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		ui.selectedTab = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.Key3) {
		ui.selectedTab = 2
	}
}

// handleScrolling processes mouse wheel scrolling.
func (ui *GuildUI) handleScrolling() {
	_, dy := ebiten.Wheel()
	if dy != 0 {
		ui.scrollOffset -= int(dy * 20)
		if ui.scrollOffset < 0 {
			ui.scrollOffset = 0
		}
	}
}

// handleTouchInput processes touch input for tab and close buttons.
func (ui *GuildUI) handleTouchInput() {
	ui.touchHandler.Update()
	touches := ui.touchHandler.GetActiveTouches()

	for _, touch := range touches {
		if touch.State == mobile.TouchStateStarted || touch.State == mobile.TouchStateMoved {
			ui.processTouchOnButtons(touch)
		}
	}
}

// processTouchOnButtons checks if touch hits tab or close buttons.
func (ui *GuildUI) processTouchOnButtons(touch *mobile.Touch) {
	if ui.checkTabButtons(touch) {
		return
	}
	ui.checkCloseButton(touch)
}

// checkTabButtons checks if touch is on any tab button.
func (ui *GuildUI) checkTabButtons(touch *mobile.Touch) bool {
	tabY := 100
	tabWidth := 150
	for i := 0; i < 3; i++ {
		tabX := 50 + i*160
		if touch.X >= tabX && touch.X <= tabX+tabWidth &&
			touch.Y >= tabY && touch.Y <= tabY+40 {
			ui.selectedTab = i
			return true
		}
	}
	return false
}

// checkCloseButton checks if touch is on the close button.
func (ui *GuildUI) checkCloseButton(touch *mobile.Touch) {
	closeX := ui.width - 80
	closeY := 50
	if touch.X >= closeX && touch.X <= closeX+60 &&
		touch.Y >= closeY && touch.Y <= closeY+40 {
		ui.visible = false
	}
}

// Draw renders the guild UI
func (ui *GuildUI) Draw(screen *ebiten.Image) {
	if !ui.visible {
		return
	}

	// Get player entity
	player := ui.getPlayerEntity()
	if player == nil {
		return
	}

	// Get guild component
	guildComp, ok := player.GetComponent("guild")
	if !ok || guildComp == nil {
		ui.drawNoGuildScreen(screen)
		return
	}

	gc := guildComp.(*GuildComponent)
	if gc.GuildID == "" {
		ui.drawNoGuildScreen(screen)
		return
	}

	// Check if guild system is available
	if ui.guildSystem == nil {
		ui.drawErrorScreen(screen, fmt.Errorf("guild system not initialized"))
		return
	}

	// Get guild info
	guildInfo, err := ui.guildSystem.GetGuildInfo(gc.GuildID)
	if err != nil {
		ui.drawErrorScreen(screen, err)
		return
	}

	// Draw background
	bgColor := color.RGBA{20, 20, 30, 240}
	ebitenutil.DrawRect(screen, 20, 20, float64(ui.width-40), float64(ui.height-40), bgColor)

	// Draw title
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Guild: %s", guildInfo.Name), 40, 40)

	// Draw close button
	closeX := ui.width - 80
	ebitenutil.DrawRect(screen, float64(closeX), 50, 60, 40, color.RGBA{100, 50, 50, 255})
	ebitenutil.DebugPrintAt(screen, "X", closeX+20, 60)

	// Draw tabs
	tabs := []string{"Info", "Members", "Treasury"}
	for i, tab := range tabs {
		tabX := 50 + i*160
		tabY := 100
		tabColor := color.RGBA{60, 60, 80, 255}
		if i == ui.selectedTab {
			tabColor = color.RGBA{100, 100, 140, 255}
		}
		ebitenutil.DrawRect(screen, float64(tabX), float64(tabY), 150, 40, tabColor)
		ebitenutil.DebugPrintAt(screen, tab, tabX+40, tabY+15)
	}

	// Draw tab content
	contentY := 160
	switch ui.selectedTab {
	case 0:
		ui.drawInfoTab(screen, guildInfo, gc, contentY)
	case 1:
		ui.drawMembersTab(screen, guildInfo, contentY)
	case 2:
		ui.drawTreasuryTab(screen, guildInfo, gc, contentY)
	}

	// Draw instructions
	instructionsY := ui.height - 60
	ebitenutil.DebugPrintAt(screen, "Keys: 1-3=Tabs, ESC=Close", 40, instructionsY)
}

// drawInfoTab draws the guild info tab
func (ui *GuildUI) drawInfoTab(screen *ebiten.Image, guildInfo *guild.Guild, gc *GuildComponent, y int) {
	lines := []string{
		fmt.Sprintf("Guild ID: %s", guildInfo.ID),
		fmt.Sprintf("Leader: %s", guildInfo.LeaderID),
		fmt.Sprintf("Your Rank: %s", gc.Rank),
		fmt.Sprintf("Members: %d", len(guildInfo.Members)),
		fmt.Sprintf("Treasury: %d gold", guildInfo.Treasury),
		"",
		"Message of the Day:",
		guildInfo.MOTD,
	}

	for i, line := range lines {
		ebitenutil.DebugPrintAt(screen, line, 50, y+i*25)
	}

	// Draw emblem info if available
	if guildInfo.Emblem != nil {
		emblemY := y + len(lines)*25 + 20
		ebitenutil.DebugPrintAt(screen, "Emblem:", 50, emblemY)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Shape: %s", guildInfo.Emblem.Shape), 50, emblemY+20)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("  Symbol: %s", guildInfo.Emblem.Symbol), 50, emblemY+40)

		// Draw emblem color preview
		emblemColor := color.RGBA{
			guildInfo.Emblem.PrimaryR,
			guildInfo.Emblem.PrimaryG,
			guildInfo.Emblem.PrimaryB,
			255,
		}
		ebitenutil.DrawRect(screen, 200, float64(emblemY+60), 50, 30, emblemColor)
	}
}

// drawMembersTab draws the guild members tab
func (ui *GuildUI) drawMembersTab(screen *ebiten.Image, guildInfo *guild.Guild, y int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Members (%d):", len(guildInfo.Members)), 50, y)

	startIdx := ui.scrollOffset / 25
	maxVisible := (ui.height - y - 100) / 25

	for i := startIdx; i < len(guildInfo.Members) && i < startIdx+maxVisible; i++ {
		member := guildInfo.Members[i]
		lineY := y + 25 + (i-startIdx)*25

		memberText := fmt.Sprintf("%s - %s", member.PlayerID, member.Rank)
		if member.PlayerID == guildInfo.LeaderID {
			memberText += " [Leader]"
		}

		ebitenutil.DebugPrintAt(screen, memberText, 50, lineY)
	}

	if len(guildInfo.Members) > maxVisible {
		ebitenutil.DebugPrintAt(screen, "Scroll to see more...", 50, ui.height-100)
	}
}

// drawTreasuryTab draws the guild treasury tab
func (ui *GuildUI) drawTreasuryTab(screen *ebiten.Image, guildInfo *guild.Guild, gc *GuildComponent, y int) {
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Treasury: %d gold", guildInfo.Treasury), 50, y)

	y += 40
	ebitenutil.DebugPrintAt(screen, "Recent Transactions:", 50, y)

	y += 25
	startIdx := ui.scrollOffset / 20
	maxVisible := (ui.height - y - 100) / 20

	transactionCount := len(guildInfo.Transactions)
	if transactionCount == 0 {
		ebitenutil.DebugPrintAt(screen, "No transactions yet", 50, y)
		return
	}

	for i := startIdx; i < transactionCount && i < startIdx+maxVisible; i++ {
		tx := guildInfo.Transactions[transactionCount-1-i] // Show newest first
		lineY := y + i*20

		action := "Deposit"
		if tx.Amount < 0 {
			action = "Withdraw"
		}

		txText := fmt.Sprintf("%s: %s %d gold by %s",
			tx.Timestamp.Format("15:04:05"),
			action,
			absInt(tx.Amount),
			tx.PlayerID)

		ebitenutil.DebugPrintAt(screen, txText, 50, lineY)
	}

	if transactionCount > maxVisible {
		ebitenutil.DebugPrintAt(screen, "Scroll to see more...", 50, ui.height-100)
	}
}

// drawNoGuildScreen draws the screen when player is not in a guild
func (ui *GuildUI) drawNoGuildScreen(screen *ebiten.Image) {
	// Draw background
	bgColor := color.RGBA{20, 20, 30, 240}
	ebitenutil.DrawRect(screen, 20, 20, float64(ui.width-40), float64(ui.height-40), bgColor)

	// Draw message
	centerY := ui.height / 2
	ebitenutil.DebugPrintAt(screen, "You are not in a guild", ui.width/2-100, centerY-20)
	ebitenutil.DebugPrintAt(screen, "Create or join a guild to access this feature", ui.width/2-180, centerY+10)
	ebitenutil.DebugPrintAt(screen, "Press ESC to close", ui.width/2-80, centerY+40)
}

// drawErrorScreen draws the screen when there's an error
func (ui *GuildUI) drawErrorScreen(screen *ebiten.Image, err error) {
	// Draw background
	bgColor := color.RGBA{30, 20, 20, 240}
	ebitenutil.DrawRect(screen, 20, 20, float64(ui.width-40), float64(ui.height-40), bgColor)

	// Draw error message
	centerY := ui.height / 2
	ebitenutil.DebugPrintAt(screen, "Error loading guild", ui.width/2-80, centerY-20)
	ebitenutil.DebugPrintAt(screen, err.Error(), ui.width/2-150, centerY+10)
	ebitenutil.DebugPrintAt(screen, "Press ESC to close", ui.width/2-80, centerY+40)
}

// getPlayerEntity finds the player entity in the world
func (ui *GuildUI) getPlayerEntity() *Entity {
	entities := ui.world.GetEntitiesWith("player")
	if len(entities) > 0 {
		return entities[0]
	}
	return nil
}

// absInt returns absolute value of an integer
func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
