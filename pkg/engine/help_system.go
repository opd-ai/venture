// Package engine provides contextual help display for the game.
// This file implements EbitenHelpSystem which renders help topics and controls
// information using an in-game overlay.
package engine

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

// HelpTopic represents a single help topic with title and content
type HelpTopic struct {
	ID      string
	Title   string
	Content []string // Multiple lines of content
	Keys    []string // Related keyboard shortcuts
}

// EbitenHelpSystem provides context-sensitive help to the player (Ebiten implementation).
// Implements UISystem interface.
type EbitenHelpSystem struct {
	Enabled       bool
	Visible       bool
	CurrentTopic  string
	Topics        map[string]HelpTopic
	QuickHints    map[string]string // Context -> Hint text
	ShowQuickHint bool
	CurrentHint   string
}

// NewHelpSystem creates a new help system with default topics.
func NewHelpSystem() *EbitenHelpSystem {
	return &EbitenHelpSystem{
		Enabled:    true,
		Visible:    false,
		Topics:     createDefaultHelpTopics(),
		QuickHints: createDefaultQuickHints(),
	}
}

// createDefaultHelpTopics generates the default help topics
func createDefaultHelpTopics() map[string]HelpTopic {
	topics := make(map[string]HelpTopic)

	topics["controls"] = HelpTopic{
		ID:    "controls",
		Title: "Game Controls",
		Content: []string{
			"Movement:",
			"  W/A/S/D - Move character",
			"  Arrow Keys - Alternative movement",
			"",
			"Actions:",
			"  SPACE - Attack / Interact",
			"  E - Use item",
			"  Q/R/F - Quick spells",
			"",
			"Interface:",
			"  I - Inventory",
			"  C - Character stats",
			"  K - Skill tree",
			"  J - Quest log",
			"  M - Map",
			"  ESC - Menu",
			"",
			"Saving:",
			"  F5 - Quick save",
			"  F9 - Quick load",
		},
		Keys: []string{"WASD", "SPACE", "I", "C", "K"},
	}

	topics["combat"] = HelpTopic{
		ID:    "combat",
		Title: "Combat Guide",
		Content: []string{
			"Combat Basics:",
			"  - Approach enemies to engage",
			"  - Press SPACE to attack",
			"  - Watch health bar (top-left)",
			"  - Dodge enemy attacks by moving",
			"",
			"Combat Tips:",
			"  - Use doorways to limit enemies",
			"  - Retreat when health is low",
			"  - Collect health potions",
			"  - Learn enemy patterns",
			"  - Equipment improves damage/defense",
			"",
			"Damage Types:",
			"  - Physical: Reduced by defense",
			"  - Fire: Burning damage over time",
			"  - Ice: Slows enemy movement",
			"  - Lightning: Chains to nearby enemies",
		},
		Keys: []string{"SPACE"},
	}

	topics["inventory"] = HelpTopic{
		ID:    "inventory",
		Title: "Inventory & Equipment",
		Content: []string{
			"Managing Items:",
			"  - Press I to open inventory",
			"  - Left-click to select item",
			"  - Right-click for options",
			"  - Drag to equip/unequip",
			"",
			"Item Rarity:",
			"  Common (Gray) - Basic items",
			"  Uncommon (Green) - Better stats",
			"  Rare (Blue) - Strong items",
			"  Epic (Purple) - Very powerful",
			"  Legendary (Orange) - Best items",
			"",
			"Equipment Slots:",
			"  - Weapon: Main damage source",
			"  - Armor: Body protection",
			"  - Helmet: Head defense",
			"  - Accessory: Special bonuses",
		},
		Keys: []string{"I", "E"},
	}

	topics["progression"] = HelpTopic{
		ID:    "progression",
		Title: "Character Progression",
		Content: []string{
			"Leveling Up:",
			"  - Defeat enemies to gain XP",
			"  - Level up increases stats",
			"  - Unlock new abilities",
			"  - Access better equipment",
			"",
			"Skill Trees:",
			"  - Press K to view skills",
			"  - Spend skill points on abilities",
			"  - Passive skills always active",
			"  - Active skills must be equipped",
			"  - Ultimate skills are powerful",
			"",
			"Stats:",
			"  Attack - Physical damage",
			"  Defense - Damage reduction",
			"  Magic - Spell power",
			"  Speed - Movement rate",
		},
		Keys: []string{"C", "K"},
	}

	topics["world"] = HelpTopic{
		ID:    "world",
		Title: "World & Exploration",
		Content: []string{
			"Dungeon Layout:",
			"  - Procedurally generated levels",
			"  - Each playthrough is unique",
			"  - Find stairs to descend",
			"  - Deeper = harder enemies",
			"",
			"Points of Interest:",
			"  - Treasure rooms: Extra loot",
			"  - Shops: Buy/sell items",
			"  - Shrines: Healing & buffs",
			"  - Boss rooms: Tough fights",
			"",
			"Exploration Tips:",
			"  - Clear each room thoroughly",
			"  - Look for secret areas",
			"  - Manage inventory space",
			"  - Save before tough fights",
		},
		Keys: []string{"M"},
	}

	topics["multiplayer"] = HelpTopic{
		ID:    "multiplayer",
		Title: "Multiplayer Co-op",
		Content: []string{
			"Playing Together:",
			"  - Start a server first",
			"  - Connect clients to server",
			"  - Share loot fairly",
			"  - Coordinate strategies",
			"",
			"Team Play Tips:",
			"  - Revive downed teammates",
			"  - Share healing items",
			"  - Focus fire on tough enemies",
			"  - Communicate via voice/chat",
			"  - Stay together for safety",
			"",
			"Network Info:",
			"  - Client-side prediction",
			"  - Works on high latency",
			"  - Up to 4 players",
		},
		Keys: []string{},
	}

	return topics
}

// createDefaultQuickHints generates context-sensitive hints
func createDefaultQuickHints() map[string]string {
	hints := make(map[string]string)

	hints["low_health"] = "Health low! Find healing or retreat to safety"
	hints["level_up"] = "Level up! Press C to view new stats, K to spend skill points"
	hints["inventory_full"] = "Inventory full! Press I to manage items"
	hints["no_mana"] = "Out of mana! Wait for regeneration or use mana potion"
	hints["enemy_nearby"] = "Enemy nearby! Prepare for combat"
	hints["item_dropped"] = "Item dropped! Walk over it to pick up, press E"
	hints["boss_ahead"] = "Boss ahead! Save your game and prepare for tough fight"
	hints["quest_complete"] = "Quest complete! Press J to view rewards"
	hints["first_death"] = "You died! Load your save with F9 to continue"

	return hints
}

// ShowTopic displays a specific help topic
func (hs *EbitenHelpSystem) ShowTopic(topicID string) {
	if !hs.Enabled {
		return
	}

	if _, exists := hs.Topics[topicID]; exists {
		hs.CurrentTopic = topicID
		hs.Visible = true
	}
}

// Hide hides the help display
func (hs *EbitenHelpSystem) Hide() {
	hs.Visible = false
}

// Toggle toggles the help display visibility
func (hs *EbitenHelpSystem) Toggle() {
	hs.Visible = !hs.Visible
	if hs.Visible && hs.CurrentTopic == "" {
		// Default to controls topic
		hs.CurrentTopic = "controls"
	}
}

// ShowQuickHintFor displays a context-sensitive hint
func (hs *EbitenHelpSystem) ShowQuickHintFor(context string) {
	if !hs.Enabled {
		return
	}

	if hint, exists := hs.QuickHints[context]; exists {
		hs.CurrentHint = hint
		hs.ShowQuickHint = true
	}
}

// HideQuickHint hides the current hint
func (hs *EbitenHelpSystem) HideQuickHint() {
	hs.ShowQuickHint = false
	hs.CurrentHint = ""
}

// Update processes the help system (can be used for auto-hiding hints)
func (hs *EbitenHelpSystem) Update(entities []*Entity, deltaTime float64) {
	if !hs.Enabled {
		return
	}

	// Auto-detect contexts and show hints
	for _, entity := range entities {
		// Check for player entity
		if !entity.HasComponent("input") {
			continue
		}

		// Check health
		if entity.HasComponent("health") {
			comp, ok := entity.GetComponent("health")
			if !ok {
				continue
			}
			health := comp.(*HealthComponent)
			if health.Current < health.Max*0.25 && !hs.ShowQuickHint {
				hs.ShowQuickHintFor("low_health")
			}
		}

		// Check inventory
		if entity.HasComponent("inventory") {
			comp, ok := entity.GetComponent("inventory")
			if !ok {
				continue
			}
			inv := comp.(*InventoryComponent)
			if len(inv.Items) >= inv.MaxItems && !hs.ShowQuickHint {
				hs.ShowQuickHintFor("inventory_full")
			}
		}
	}
}

// Draw renders the help system UI (implements UISystem interface).
// The screen parameter should be *ebiten.Image in production.
func (hs *EbitenHelpSystem) Draw(screen interface{}) {
	// Type assert to *ebiten.Image
	ebitenScreen, ok := screen.(*ebiten.Image)
	if !ok {
		return // Invalid screen type
	}

	if !hs.Enabled {
		return
	}

	// Draw quick hint if active
	if hs.ShowQuickHint && hs.CurrentHint != "" {
		hs.drawQuickHint(ebitenScreen)
	}

	// Draw full help panel if visible
	if hs.Visible && hs.CurrentTopic != "" {
		hs.drawHelpPanel(ebitenScreen)
	}
}

// drawQuickHint renders a small hint at the top of screen
func (hs *EbitenHelpSystem) drawQuickHint(screen *ebiten.Image) {
	screenWidth := screen.Bounds().Dx()

	hintWidth := 600
	hintHeight := 40
	hintX := (screenWidth - hintWidth) / 2
	hintY := 50

	// Semi-transparent background
	vector.DrawFilledRect(screen,
		float32(hintX), float32(hintY),
		float32(hintWidth), float32(hintHeight),
		color.RGBA{50, 50, 100, 200}, false)

	// Border
	vector.StrokeRect(screen,
		float32(hintX), float32(hintY),
		float32(hintWidth), float32(hintHeight),
		2, color.RGBA{100, 150, 255, 255}, false)

	// Hint text
	textColor := color.RGBA{255, 255, 150, 255}
	text.Draw(screen, "💡 "+hs.CurrentHint, basicfont.Face7x13, hintX+10, hintY+25, textColor)
}

// drawHelpPanel renders the full help panel
func (hs *EbitenHelpSystem) drawHelpPanel(screen *ebiten.Image) {
	topic, exists := hs.Topics[hs.CurrentTopic]
	if !exists {
		return
	}

	screenWidth := screen.Bounds().Dx()
	screenHeight := screen.Bounds().Dy()

	panelWidth := 600
	panelHeight := 500
	panelX := (screenWidth - panelWidth) / 2
	panelY := (screenHeight - panelHeight) / 2

	// Background
	vector.DrawFilledRect(screen,
		float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		color.RGBA{20, 20, 40, 240}, false)

	// Border
	vector.StrokeRect(screen,
		float32(panelX), float32(panelY),
		float32(panelWidth), float32(panelHeight),
		3, color.RGBA{100, 150, 255, 255}, false)

	// Title bar
	vector.DrawFilledRect(screen,
		float32(panelX), float32(panelY),
		float32(panelWidth), 40,
		color.RGBA{50, 80, 150, 255}, false)

	titleColor := color.RGBA{255, 255, 255, 255}
	text.Draw(screen, "Help: "+topic.Title, basicfont.Face7x13, panelX+20, panelY+25, titleColor)

	// Close hint
	closeColor := color.RGBA{200, 200, 200, 255}
	text.Draw(screen, "[ESC to close]", basicfont.Face7x13, panelX+panelWidth-150, panelY+25, closeColor)

	// Content
	contentY := panelY + 60
	lineHeight := 15
	contentColor := color.RGBA{220, 220, 220, 255}

	for i, line := range topic.Content {
		if contentY+i*lineHeight > panelY+panelHeight-40 {
			break // Don't overflow panel
		}
		text.Draw(screen, line, basicfont.Face7x13, panelX+20, contentY+i*lineHeight, contentColor)
	}

	// Topic selector at bottom
	selectorY := panelY + panelHeight - 30
	selectorColor := color.RGBA{150, 180, 255, 255}

	topicList := "Topics: [1]Controls [2]Combat [3]Inventory [4]Progression [5]World [6]Multiplayer"
	text.Draw(screen, topicList, basicfont.Face7x13, panelX+20, selectorY, selectorColor)
}

// GetTopicList returns all available topic IDs
func (hs *EbitenHelpSystem) GetTopicList() []string {
	topics := make([]string, 0, len(hs.Topics))
	for id := range hs.Topics {
		topics = append(topics, id)
	}
	return topics
}

// GetTopic returns a specific help topic
func (hs *EbitenHelpSystem) GetTopic(id string) (*HelpTopic, bool) {
	topic, exists := hs.Topics[id]
	return &topic, exists
}

// IsActive implements UISystem interface.
func (hs *EbitenHelpSystem) IsActive() bool {
	return hs.Visible
}

// SetActive implements UISystem interface.
func (hs *EbitenHelpSystem) SetActive(active bool) {
	hs.Visible = active
}

// Compile-time interface check
var _ UISystem = (*EbitenHelpSystem)(nil)
