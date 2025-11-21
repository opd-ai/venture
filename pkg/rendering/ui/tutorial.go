// Package ui provides tutorial and accessibility systems for Phase 60.1.
package ui

import (
	"fmt"
	"sync"
	"time"
)

// TutorialTopic represents a feature with tutorial content
type TutorialTopic string

const (
	TutorialMovement    TutorialTopic = "movement"
	TutorialCombat      TutorialTopic = "combat"
	TutorialInventory   TutorialTopic = "inventory"
	TutorialCrafting    TutorialTopic = "crafting"
	TutorialHousing     TutorialTopic = "housing"
	TutorialGuilds      TutorialTopic = "guilds"
	TutorialVehicles    TutorialTopic = "vehicles"
	TutorialCompanions  TutorialTopic = "companions"
	TutorialQuests      TutorialTopic = "quests"
	TutorialSkills      TutorialTopic = "skills"
	TutorialTrading     TutorialTopic = "trading"
	TutorialMultiplayer TutorialTopic = "multiplayer"
	TutorialMap         TutorialTopic = "map"
	TutorialSettings    TutorialTopic = "settings"
	TutorialKeybinds    TutorialTopic = "keybinds"
	TutorialQuickTravel TutorialTopic = "quicktravel"
	TutorialFederation  TutorialTopic = "federation"
	TutorialRaids       TutorialTopic = "raids"
	TutorialPrestige    TutorialTopic = "prestige"
	TutorialPolitics    TutorialTopic = "politics"
	TutorialEconomy     TutorialTopic = "economy"
	TutorialNarratives  TutorialTopic = "narratives"
	TutorialSieges      TutorialTopic = "sieges"
	TutorialLegendary   TutorialTopic = "legendary"
	TutorialBranching   TutorialTopic = "branching"
	TutorialPhysics     TutorialTopic = "physics"
	TutorialFluids      TutorialTopic = "fluids"
	TutorialDestruction TutorialTopic = "destruction"
	TutorialMods        TutorialTopic = "mods"
	TutorialBlueprints  TutorialTopic = "blueprints"
)

// Tutorial represents a single tutorial entry
type Tutorial struct {
	Topic     TutorialTopic
	Title     string
	Content   []string // multi-line tutorial text
	Keybind   string   // associated keybind hint
	ImagePath string   // optional screenshot/diagram path
	VideoURL  string   // optional video URL
	Viewed    bool
	ViewedAt  time.Time
}

// TutorialManager manages context-sensitive help system
type TutorialManager struct {
	mu        sync.RWMutex
	tutorials map[TutorialTopic]*Tutorial
	enabled   bool
}

// NewTutorialManager creates a new tutorial manager
func NewTutorialManager() *TutorialManager {
	tm := &TutorialManager{
		tutorials: make(map[TutorialTopic]*Tutorial),
		enabled:   true,
	}
	tm.registerDefaultTutorials()
	return tm
}

// registerDefaultTutorials sets up all tutorial content
func (tm *TutorialManager) registerDefaultTutorials() {
	tutorials := []*Tutorial{
		{
			Topic: TutorialMovement,
			Title: "Movement Controls",
			Content: []string{
				"Use WASD or Arrow Keys to move your character.",
				"Hold Shift to sprint (consumes stamina).",
				"Press Space to dodge roll (invincibility frames).",
			},
			Keybind: "WASD / Arrows",
		},
		{
			Topic: TutorialCombat,
			Title: "Combat Basics",
			Content: []string{
				"Left-Click to attack enemies.",
				"Right-Click to block incoming damage.",
				"Use number keys 1-4 for abilities.",
				"Press R to unleash your Ultimate ability.",
			},
			Keybind: "Mouse / 1-4 / R",
		},
		{
			Topic: TutorialInventory,
			Title: "Inventory Management",
			Content: []string{
				"Press I to open your inventory.",
				"Press E near items to pick them up.",
				"Press G to drop selected items.",
				"Use quick slots 5-8 for consumables.",
			},
			Keybind: "I / E / G / 5-8",
		},
		{
			Topic: TutorialCrafting,
			Title: "Crafting System",
			Content: []string{
				"Press N to open the crafting menu.",
				"Build crafting stations in your house for bonuses.",
				"Higher quality stations unlock more recipes.",
				"Master tier stations provide +100% stat bonuses!",
			},
			Keybind: "N",
		},
		{
			Topic: TutorialHousing,
			Title: "Player Housing",
			Content: []string{
				"Press H to access housing management.",
				"Place crafting stations to unlock recipes.",
				"Add companion beds to boost loyalty.",
				"Guild houses provide communal crafting spaces.",
			},
			Keybind: "H",
		},
		{
			Topic: TutorialGuilds,
			Title: "Guild System",
			Content: []string{
				"Press O to open the guild menu.",
				"Create or join guilds for social play.",
				"Guild banks share resources across members.",
				"Participate in territory sieges for rewards.",
			},
			Keybind: "O",
		},
		{
			Topic: TutorialVehicles,
			Title: "Vehicle System",
			Content: []string{
				"Press V to mount/dismount vehicles.",
				"Vehicles provide fast travel and combat options.",
				"Guilds can own vehicle fleets with formations.",
				"Siege engines deal massive damage to structures.",
			},
			Keybind: "V",
		},
		{
			Topic: TutorialCompanions,
			Title: "Companion System",
			Content: []string{
				"Companions fight alongside you in combat.",
				"F1: Follow, F2: Stay, F3: Attack commands.",
				"Build companion beds in houses for loyalty bonuses.",
				"High loyalty unlocks personal companion quests.",
			},
			Keybind: "F1-F3",
		},
		{
			Topic: TutorialQuests,
			Title: "Quest System",
			Content: []string{
				"Press J to view your quest log.",
				"Complete quests for XP, gold, and items.",
				"Choices in quests affect future story branches.",
				"Legendary quests require 10-20 hours to complete!",
			},
			Keybind: "J",
		},
		{
			Topic: TutorialSkills,
			Title: "Skills & Progression",
			Content: []string{
				"Press K to view your skill tree.",
				"Allocate skill points from leveling up.",
				"Training areas in houses grant +50% XP.",
				"Companions also have skill trees that evolve.",
			},
			Keybind: "K",
		},
		{
			Topic: TutorialQuickTravel,
			Title: "Quick Travel",
			Content: []string{
				"Press F to quick travel to unlocked locations.",
				"Cost increases with distance (100-1000 gold).",
				"Own houses and guild halls are travel destinations.",
				"Unlock destinations by visiting them first.",
			},
			Keybind: "F",
		},
		{
			Topic: TutorialRaids,
			Title: "Raid Dungeons",
			Content: []string{
				"Raids require 5-10 players to complete.",
				"5 difficulty tiers: Normal to Nightmare.",
				"Each raid has 3-5 bosses with unique mechanics.",
				"Weekly lockouts prevent repeated farming.",
			},
		},
		{
			Topic: TutorialPrestige,
			Title: "Prestige System",
			Content: []string{
				"After level 50, gain Prestige levels infinitely.",
				"Each level grants 1 Paragon Point for stat bonuses.",
				"Unlock prestige abilities at levels 10, 25, 50, 100.",
				"Account-wide bonuses: +5% XP per prestige 100 character.",
			},
		},
		{
			Topic: TutorialPolitics,
			Title: "Political Warfare",
			Content: []string{
				"Guilds can declare war (24-hour preparation).",
				"Trade embargoes increase prices by 50-90%.",
				"Diplomatic victories avoid combat via negotiation.",
				"Aggressive actions reduce NPC faction reputation.",
			},
		},
		{
			Topic: TutorialEconomy,
			Title: "Federated Economy",
			Content: []string{
				"Marketplace spans all federated servers.",
				"Search for items across the entire network.",
				"Dynamic pricing based on supply/demand.",
				"Transaction fees: 5% + 2% per server hop (max 15%).",
			},
		},
	}

	for _, t := range tutorials {
		tm.tutorials[t.Topic] = t
	}
}

// ShowTutorial marks a tutorial as viewed
func (tm *TutorialManager) ShowTutorial(topic TutorialTopic) (*Tutorial, error) {
	if !tm.enabled {
		return nil, fmt.Errorf("tutorials disabled")
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()

	t, exists := tm.tutorials[topic]
	if !exists {
		return nil, fmt.Errorf("tutorial not found: %s", topic)
	}

	if !t.Viewed {
		t.Viewed = true
		t.ViewedAt = time.Now()
	}

	return t, nil
}

// GetTutorial retrieves a tutorial without marking it viewed
func (tm *TutorialManager) GetTutorial(topic TutorialTopic) (*Tutorial, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	t, exists := tm.tutorials[topic]
	if !exists {
		return nil, fmt.Errorf("tutorial not found: %s", topic)
	}
	return t, nil
}

// ListUnviewed returns all unviewed tutorials
func (tm *TutorialManager) ListUnviewed() []*Tutorial {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	result := make([]*Tutorial, 0)
	for _, t := range tm.tutorials {
		if !t.Viewed {
			result = append(result, t)
		}
	}
	return result
}

// Enable enables tutorial popups
func (tm *TutorialManager) Enable() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabled = true
}

// Disable disables tutorial popups
func (tm *TutorialManager) Disable() {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.enabled = false
}

// IsEnabled returns tutorial enabled state
func (tm *TutorialManager) IsEnabled() bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.enabled
}

// ResetProgress resets all tutorials to unviewed
func (tm *TutorialManager) ResetProgress() {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	for _, t := range tm.tutorials {
		t.Viewed = false
		t.ViewedAt = time.Time{}
	}
}

// ColorblindMode represents colorblind accessibility modes
type ColorblindMode int

const (
	ColorblindNone         ColorblindMode = iota
	ColorblindProtanopia                  // red-blind
	ColorblindDeuteranopia                // green-blind
	ColorblindTritanopia                  // blue-blind
)

func (c ColorblindMode) String() string {
	switch c {
	case ColorblindNone:
		return "None"
	case ColorblindProtanopia:
		return "Protanopia"
	case ColorblindDeuteranopia:
		return "Deuteranopia"
	case ColorblindTritanopia:
		return "Tritanopia"
	default:
		return "Unknown"
	}
}

// AccessibilityConfig holds accessibility settings
type AccessibilityConfig struct {
	ColorblindMode ColorblindMode
	FontScale      float64 // 0.5-2.0
	HighContrast   bool
	ScreenReader   bool
	ReduceMotion   bool
	ClosedCaptions bool
	SimplifiedUI   bool
}

// NewAccessibilityConfig creates default accessibility config
func NewAccessibilityConfig() *AccessibilityConfig {
	return &AccessibilityConfig{
		ColorblindMode: ColorblindNone,
		FontScale:      1.0,
		HighContrast:   false,
		ScreenReader:   false,
		ReduceMotion:   false,
		ClosedCaptions: false,
		SimplifiedUI:   false,
	}
}

// Validate ensures accessibility settings are within valid ranges
func (ac *AccessibilityConfig) Validate() error {
	if ac.FontScale < 0.5 || ac.FontScale > 2.0 {
		return fmt.Errorf("font scale must be between 0.5 and 2.0")
	}
	return nil
}

// ApplyColorblindFilter applies color adjustments for colorblind modes
func (ac *AccessibilityConfig) ApplyColorblindFilter(r, g, b uint8) (uint8, uint8, uint8) {
	switch ac.ColorblindMode {
	case ColorblindProtanopia:
		// Red-blind: shift red channel to green
		return uint8(float64(g)*0.5 + float64(r)*0.5), g, b
	case ColorblindDeuteranopia:
		// Green-blind: shift green to red/blue
		return uint8(float64(r)*0.8 + float64(g)*0.2), uint8(float64(g)*0.5 + float64(b)*0.5), b
	case ColorblindTritanopia:
		// Blue-blind: shift blue to green
		return r, uint8(float64(g)*0.7 + float64(b)*0.3), uint8(float64(b)*0.5 + float64(g)*0.5)
	default:
		return r, g, b
	}
}

// GetContrastMultiplier returns UI contrast multiplier
func (ac *AccessibilityConfig) GetContrastMultiplier() float64 {
	if ac.HighContrast {
		return 1.5
	}
	return 1.0
}

// ShouldShowCaptions returns whether to display captions
func (ac *AccessibilityConfig) ShouldShowCaptions() bool {
	return ac.ClosedCaptions
}
