package features

import "time"

// RegisterMetaFeatures adds all meta-game and UI features to the registry
func RegisterMetaFeatures(r *FeatureRegistry) {
	// Tutorial
	r.Register(&Feature{
		ID:                   "meta.tutorial",
		Name:                 "Tutorial System",
		Category:             CategoryMeta,
		Description:          "Context-sensitive help for first-time users",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Automatically shown on first launch",
		HasTutorial:          true,
		TutorialLocation:     "Tutorial system itself",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"ui", "progression", "detection"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Settings
	r.Register(&Feature{
		ID:                   "meta.settings",
		Name:                 "Game Settings",
		Category:             CategoryMeta,
		Description:          "Graphics, audio, controls, gameplay options",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Press Esc, select Settings",
		HasTutorial:          true,
		TutorialLocation:     "Settings menu tooltips",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"config", "ui", "persistence"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Save/Load
	r.Register(&Feature{
		ID:                   "meta.save",
		Name:                 "Manual Save",
		Category:             CategoryMeta,
		Description:          "Manually save game progress",
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		AccessibilityPath:    "Press Esc, select Save Game",
		HasTutorial:          true,
		TutorialLocation:     "Save menu tooltip",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"persistence", "serialization", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "meta.autosave",
		Name:                 "Auto-Save",
		Category:             CategoryMeta,
		Description:          "Automatic periodic saves",
		Accessible:           true,
		AccessibilityTime:    10 * time.Minute,
		AccessibilityPath:    "Observe auto-save notification",
		HasTutorial:          true,
		TutorialLocation:     "Auto-save notification",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"persistence", "timer", "notifications"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "meta.load",
		Name:                 "Load Game",
		Category:             CategoryMeta,
		Description:          "Load saved games",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Main menu, select Load Game",
		HasTutorial:          true,
		TutorialLocation:     "Load menu tooltip",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"persistence", "deserialization", "world"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Hotkeys
	r.Register(&Feature{
		ID:                   "meta.hotkeys",
		Name:                 "Customizable Hotkeys",
		Category:             CategoryMeta,
		Description:          "Rebind keyboard shortcuts",
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		AccessibilityPath:    "Settings menu, select Controls",
		HasTutorial:          true,
		TutorialLocation:     "Controls settings tooltip",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"input", "config", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "meta.gamepad",
		Name:                 "Gamepad Support",
		Category:             CategoryMeta,
		Description:          "Controller input support",
		Accessible:           true,
		AccessibilityTime:    2 * time.Minute,
		AccessibilityPath:    "Connect gamepad, auto-detected",
		HasTutorial:          true,
		TutorialLocation:     "Gamepad detected notification",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"input", "detection", "mapping"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Map
	r.Register(&Feature{
		ID:                   "meta.minimap",
		Name:                 "Minimap",
		Category:             CategoryMeta,
		Description:          "Small map in corner of screen",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Visible by default in HUD",
		HasTutorial:          true,
		TutorialLocation:     "Minimap tutorial",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"rendering", "world", "ui"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "meta.worldmap",
		Name:                 "World Map",
		Category:             CategoryMeta,
		Description:          "Large full-screen map",
		Accessible:           true,
		AccessibilityTime:    3 * time.Minute,
		AccessibilityPath:    "Press M key",
		HasTutorial:          true,
		TutorialLocation:     "Map key tooltip",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"rendering", "world", "navigation"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	r.Register(&Feature{
		ID:                   "meta.fog",
		Name:                 "Fog of War",
		Category:             CategoryMeta,
		Description:          "Unexplored areas hidden on map",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Observe dark areas on minimap",
		HasTutorial:          true,
		TutorialLocation:     "Map tutorial",
		TutorialCompleteness: 0.9,
		IntegratedSystems:    []string{"exploration", "rendering", "persistence"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// HUD
	r.Register(&Feature{
		ID:                   "meta.hud",
		Name:                 "HUD Display",
		Category:             CategoryMeta,
		Description:          "Health, mana, XP bar, buffs/debuffs",
		Accessible:           true,
		AccessibilityTime:    1 * time.Minute,
		AccessibilityPath:    "Visible by default",
		HasTutorial:          true,
		TutorialLocation:     "HUD tutorial",
		TutorialCompleteness: 1.0,
		IntegratedSystems:    []string{"ui", "stats", "effects"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Adaptive Music (V4.0)
	r.Register(&Feature{
		ID:                   "meta.music",
		Name:                 "Adaptive Music",
		Category:             CategoryMeta,
		Description:          "Music responds to gameplay context",
		Accessible:           true,
		AccessibilityTime:    5 * time.Minute,
		AccessibilityPath:    "Enter combat, observe music change",
		HasTutorial:          true,
		TutorialLocation:     "Audio settings tooltip",
		TutorialCompleteness: 0.7,
		IntegratedSystems:    []string{"audio", "combat", "context"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})

	// Storytelling (V4.0)
	r.Register(&Feature{
		ID:                   "meta.lore",
		Name:                 "Environmental Storytelling",
		Category:             CategoryContent,
		Description:          "Discover story fragments in world",
		Accessible:           true,
		AccessibilityTime:    20 * time.Minute,
		AccessibilityPath:    "Explore dungeon, find story fragments",
		HasTutorial:          true,
		TutorialLocation:     "Story journal tutorial",
		TutorialCompleteness: 0.8,
		IntegratedSystems:    []string{"discovery", "journal", "narrative"},
		IntegrationCount:     3,
		Implemented:          true,
		Tested:               true,
		Functional:           true,
	})
}

// GetDefaultRegistry returns a fully populated registry with all game features registered
// across all categories (core, advanced, social, housing, guild, and meta systems).
// It is the primary entry point for feature auditing and completeness checks.
// Use this in tests and audit tools to verify feature coverage.
func GetDefaultRegistry() *FeatureRegistry {
	r := NewFeatureRegistry()
	RegisterCoreFeatures(r)
	RegisterAdvancedFeatures(r)
	RegisterSocialFeatures(r)
	RegisterHousingFeatures(r)
	RegisterGuildFeatures(r)
	RegisterMetaFeatures(r)
	return r
}
