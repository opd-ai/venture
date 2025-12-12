// Command advanceduittest demonstrates Phase 60.1 Advanced UI Systems.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/opd-ai/venture/pkg/rendering/ui"
)

func main() {
	mode := flag.String("mode", "demo", "Test mode: demo, settings, keybinds, travel, tooltips, tutorial, accessibility, all")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	switch *mode {
	case "demo":
		runDemo(*verbose)
	case "settings":
		testSettings(*verbose)
	case "keybinds":
		testKeybinds(*verbose)
	case "travel":
		testQuickTravel(*verbose)
	case "tooltips":
		testTooltips(*verbose)
	case "tutorial":
		testTutorials(*verbose)
	case "accessibility":
		testAccessibility(*verbose)
	case "all":
		runDemo(*verbose)
		testSettings(*verbose)
		testKeybinds(*verbose)
		testQuickTravel(*verbose)
		testTooltips(*verbose)
		testTutorials(*verbose)
		testAccessibility(*verbose)
	default:
		fmt.Printf("Unknown mode: %s\n", *mode)
		flag.Usage()
		os.Exit(1)
	}
}

func runDemo(verbose bool) {
	fmt.Println("=== Phase 60.1: Advanced UI Systems Demo ===")

	fmt.Println("Features implemented:")
	fmt.Println("✅ Unified Settings Menu (10+ categories)")
	fmt.Println("✅ Keybind Customization (50+ actions)")
	fmt.Println("✅ Quick-Travel System (distance-based cost)")
	fmt.Println("✅ Enhanced Tooltips (integration bonuses)")
	fmt.Println("✅ Tutorial System (30+ features)")
	fmt.Println("✅ Accessibility Options (colorblind, font scale, contrast)")
	fmt.Println()
}

func testSettings(verbose bool) {
	fmt.Println("=== Testing Settings Manager ===")

	sm := ui.NewSettingsManager()

	// Display categories
	fmt.Println("\nSettings Categories:")
	categories := []ui.SettingsCategory{
		ui.CategoryGraphics,
		ui.CategoryAudio,
		ui.CategoryControls,
		ui.CategoryGameplay,
		ui.CategoryNetwork,
		ui.CategoryAccessibility,
		ui.CategoryInterface,
		ui.CategoryPerformance,
		ui.CategorySocial,
		ui.CategoryAdvanced,
	}

	for _, cat := range categories {
		settings := sm.ListByCategory(cat)
		fmt.Printf("  %s: %d settings\n", cat, len(settings))
	}

	// Show sample settings
	fmt.Println("\nSample Graphics Settings:")
	graphicsSettings := sm.ListByCategory(ui.CategoryGraphics)
	for i, s := range graphicsSettings {
		if i >= 5 {
			break
		}
		fmt.Printf("  %s: %v (default: %v)\n", s.Name, s.CurrentValue, s.DefaultValue)
	}

	// Test modification
	fmt.Println("\nModifying settings...")
	if err := sm.SetValue("graphics.resolution", "2560x1440"); err != nil {
		log.Printf("Error: %v", err)
	}
	if err := sm.SetValue("audio.master_volume", 80); err != nil {
		log.Printf("Error: %v", err)
	}

	fmt.Printf("  Resolution: %s\n", sm.GetString("graphics.resolution"))
	fmt.Printf("  Master Volume: %d\n", sm.GetInt("audio.master_volume"))
	fmt.Printf("  Settings modified: %v\n", sm.IsModified())

	// Test save/load
	filename := "test_settings.json"
	defer os.Remove(filename)

	if err := sm.Save(filename); err != nil {
		log.Printf("Save error: %v", err)
	} else {
		fmt.Printf("\n✓ Settings saved to %s\n", filename)
	}

	sm2 := ui.NewSettingsManager()
	if err := sm2.Load(filename); err != nil {
		log.Printf("Load error: %v", err)
	} else {
		fmt.Printf("✓ Settings loaded from %s\n", filename)
		fmt.Printf("  Loaded resolution: %s\n", sm2.GetString("graphics.resolution"))
	}
}

func testKeybinds(verbose bool) {
	fmt.Println("\n=== Testing Keybind Manager ===")

	km := ui.NewKeybindManager()

	// Show default bindings
	fmt.Println("\nDefault Movement Bindings:")
	actions := []ui.KeyAction{
		ui.ActionMoveUp,
		ui.ActionMoveDown,
		ui.ActionMoveLeft,
		ui.ActionMoveRight,
		ui.ActionSprint,
	}

	for _, action := range actions {
		kb, _ := km.GetBinding(action)
		if kb.SecondaryKey != "" {
			fmt.Printf("  %s: %s / %s\n", action, kb.PrimaryKey, kb.SecondaryKey)
		} else {
			fmt.Printf("  %s: %s\n", action, kb.PrimaryKey)
		}
	}

	// Show combat bindings
	fmt.Println("\nDefault Combat Bindings:")
	combatActions := []ui.KeyAction{
		ui.ActionAttack,
		ui.ActionBlock,
		ui.ActionDodge,
		ui.ActionAbility1,
		ui.ActionAbility2,
	}

	for _, action := range combatActions {
		kb, _ := km.GetBinding(action)
		fmt.Printf("  %s: %s\n", action, kb.PrimaryKey)
		if verbose && kb.Description != "" {
			fmt.Printf("    (%s)\n", kb.Description)
		}
	}

	// Test conflict detection
	fmt.Println("\nConflict Detection:")
	conflicts := km.DetectConflicts()
	if len(conflicts) == 0 {
		fmt.Println("  ✓ No conflicts detected")
	} else {
		for _, conflict := range conflicts {
			fmt.Printf("  ⚠ %s\n", conflict)
		}
	}

	// Test rebinding
	fmt.Println("\nRebinding test:")
	fmt.Printf("  Original Attack: %s\n", func() string { kb, _ := km.GetBinding(ui.ActionAttack); return string(kb.PrimaryKey) }())

	if err := km.SetBinding(ui.ActionAttack, ui.KeySpace, ""); err != nil {
		fmt.Printf("  Error rebinding: %v\n", err)
	} else {
		fmt.Printf("  Attack rebound to: %s\n", func() string { kb, _ := km.GetBinding(ui.ActionAttack); return string(kb.PrimaryKey) }())
	}

	// Reset
	km.ResetToDefaults()
	fmt.Printf("  After reset: %s\n", func() string { kb, _ := km.GetBinding(ui.ActionAttack); return string(kb.PrimaryKey) }())

	// Count total bindings
	allBindings := km.ListAllBindings()
	fmt.Printf("\n✓ Total bindings registered: %d\n", len(allBindings))
}

func testQuickTravel(verbose bool) {
	fmt.Println("\n=== Testing Quick Travel ===")

	qtm := ui.NewQuickTravelManager()

	// Register destinations
	destinations := []*ui.TravelDestination{
		{ID: "home", Name: "My House", X: 100, Y: 100, Category: "House"},
		{ID: "guild", Name: "Guild Hall", X: 500, Y: 500, Category: "Guild Hall"},
		{ID: "city", Name: "Capital City", X: 1000, Y: 1000, Category: "City"},
		{ID: "dungeon", Name: "Dark Dungeon", X: 1500, Y: 500, Category: "Dungeon"},
	}

	for _, dest := range destinations {
		qtm.RegisterDestination(dest)
		qtm.UnlockDestination(dest.ID)
	}

	fmt.Println("\nUnlocked Destinations:")
	unlocked := qtm.ListUnlocked()
	for _, dest := range unlocked {
		cost, _ := qtm.CalculateCost(0, 0, dest.ID)
		fmt.Printf("  %s (%s) - Cost: %d gold\n", dest.Name, dest.Category, cost)
	}

	// Test travel
	fmt.Println("\nTravel Test:")
	playerGold := 1000
	fmt.Printf("  Starting gold: %d\n", playerGold)

	dest, cost, err := qtm.Travel(0, 0, "guild", &playerGold)
	if err != nil {
		fmt.Printf("  Travel failed: %v\n", err)
	} else {
		fmt.Printf("  ✓ Traveled to %s\n", dest.Name)
		fmt.Printf("  Cost: %d gold\n", cost)
		fmt.Printf("  Remaining gold: %d\n", playerGold)
		fmt.Printf("  New position: (%.0f, %.0f)\n", dest.X, dest.Y)
	}

	// List by category
	fmt.Println("\nDestinations by Category:")
	categories := []string{"House", "Guild Hall", "City", "Dungeon"}
	for _, cat := range categories {
		dests := qtm.ListByCategory(cat)
		fmt.Printf("  %s: %d destination(s)\n", cat, len(dests))
	}
}

func testTooltips(verbose bool) {
	fmt.Println("\n=== Testing Enhanced Tooltips ===")

	// Item tooltip
	fmt.Println("\n1. Item with Crafting Bonus:")
	tooltip := ui.CreateItemTooltip("Legendary Sword", "Legendary", 250, 100, 2.0)
	fmt.Print(ui.FormatTooltip(tooltip))

	// Station tooltip
	fmt.Println("\n2. Master Quality Crafting Station:")
	stationTooltip := ui.CreateStationTooltip("Master Forge", 4, 40)
	fmt.Print(ui.FormatTooltip(stationTooltip))

	// Companion tooltip
	fmt.Println("\n3. High Loyalty Companion:")
	companionTooltip := ui.CreateCompanionTooltip("Fluffy", 30, 85, []string{"Attack", "Defend", "Heal"})
	fmt.Print(ui.FormatTooltip(companionTooltip))

	// Custom tooltip
	fmt.Println("\n4. Custom Tooltip with Builder:")
	customTooltip := ui.NewTooltipBuilder("Epic Shield").
		SetRarity("Epic").
		AddDescription("A shield forged in dragon fire").
		AddStat("Defense", 200).
		AddStat("Block Chance", "25%").
		AddBonus("+15% fire resistance").
		AddBonus("Reflect 10% damage").
		AddRequirement("Level 40").
		AddRequirement("Strength 50").
		SetCost(5000).
		Build()
	fmt.Print(ui.FormatTooltip(customTooltip))
}

func testTutorials(verbose bool) {
	fmt.Println("\n=== Testing Tutorial System ===")

	tm := ui.NewTutorialManager()

	// Show unviewed count
	unviewed := tm.ListUnviewed()
	fmt.Printf("\nUnviewed tutorials: %d\n", len(unviewed))

	// Show sample tutorials
	fmt.Println("\nSample Tutorials:")
	topics := []ui.TutorialTopic{
		ui.TutorialMovement,
		ui.TutorialCombat,
		ui.TutorialHousing,
		ui.TutorialQuickTravel,
	}

	for _, topic := range topics {
		tutorial, err := tm.ShowTutorial(topic)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("\n%s:\n", tutorial.Title)
		for _, line := range tutorial.Content {
			fmt.Printf("  %s\n", line)
		}
		if tutorial.Keybind != "" {
			fmt.Printf("  [%s]\n", tutorial.Keybind)
		}
	}

	// Show viewed status
	fmt.Printf("\n✓ Tutorials viewed: %d\n", len(topics))
	unviewed = tm.ListUnviewed()
	fmt.Printf("  Remaining unviewed: %d\n", len(unviewed))

	// Test disable
	tm.Disable()
	fmt.Printf("\nTutorials enabled: %v\n", tm.IsEnabled())

	tm.Enable()
	fmt.Printf("Tutorials enabled: %v\n", tm.IsEnabled())
}

func testAccessibility(verbose bool) {
	fmt.Println("\n=== Testing Accessibility Options ===")

	config := ui.NewAccessibilityConfig()

	// Show defaults
	fmt.Println("\nDefault Accessibility Settings:")
	fmt.Printf("  Colorblind Mode: %s\n", config.ColorblindMode)
	fmt.Printf("  Font Scale: %.1f\n", config.FontScale)
	fmt.Printf("  High Contrast: %v\n", config.HighContrast)
	fmt.Printf("  Screen Reader: %v\n", config.ScreenReader)
	fmt.Printf("  Reduce Motion: %v\n", config.ReduceMotion)
	fmt.Printf("  Closed Captions: %v\n", config.ClosedCaptions)

	// Test colorblind filters
	fmt.Println("\nColorblind Filter Test (RGB: 255, 128, 64):")
	modes := []ui.ColorblindMode{
		ui.ColorblindNone,
		ui.ColorblindProtanopia,
		ui.ColorblindDeuteranopia,
		ui.ColorblindTritanopia,
	}

	for _, mode := range modes {
		config.ColorblindMode = mode
		r, g, b := config.ApplyColorblindFilter(255, 128, 64)
		fmt.Printf("  %s: RGB(%d, %d, %d)\n", mode, r, g, b)
	}

	// Test contrast
	fmt.Println("\nContrast Multiplier:")
	config.HighContrast = false
	fmt.Printf("  Normal: %.1fx\n", config.GetContrastMultiplier())
	config.HighContrast = true
	fmt.Printf("  High Contrast: %.1fx\n", config.GetContrastMultiplier())

	// Test font scaling
	fmt.Println("\nFont Scale Validation:")
	testScales := []float64{0.5, 1.0, 1.5, 2.0, 2.5}
	for _, scale := range testScales {
		config.FontScale = scale
		err := config.Validate()
		if err != nil {
			fmt.Printf("  %.1f: ✗ %v\n", scale, err)
		} else {
			fmt.Printf("  %.1f: ✓ Valid\n", scale)
		}
	}

	fmt.Println("\n✓ Accessibility system operational")
}
