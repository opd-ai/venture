package engine

import (
	"testing"
)

// TestModBrowserScriptingIntegration verifies mod browser integrates with scripting system.
func TestModBrowserScriptingIntegration(t *testing.T) {
	world := NewWorld()

	// Create mod browser system
	browserSystem := NewModBrowserSystem(world)
	repo := NewInMemoryModRepository()

	// Add a mod with scripting component
	repo.AddMod(ModListing{
		ID:          "script-mod",
		Name:        "Scripted Mod",
		Size:        100,
		Categories:  []string{"gameplay"},
		GameVersion: "10.0.0",
	}, []byte(`{"name":"test-script"}`))

	browserSystem.SetRepository(repo)

	// Create entity with both mod browser and scripting components
	entity := world.CreateEntity()
	browserComp := NewModBrowserComponent()
	scriptComp := NewScriptingComponent()
	entity.AddComponent(browserComp)
	entity.AddComponent(scriptComp)

	// Populate available mods
	browserComp.SetAvailableMods([]ModListing{{
		ID:          "script-mod",
		Name:        "Scripted Mod",
		Size:        100,
		Categories:  []string{"gameplay"},
		GameVersion: "10.0.0",
	}})

	// Install callback that registers scripts
	browserSystem.SetInstallCallback(func(modID string, modData []byte) error {
		// Simulate adding a script from the mod
		script := &Script{
			ID:           modID + "-script",
			ModID:        modID,
			Source:       "set(result, 42)",
			TriggerEvent: "on_load",
			Priority:     0,
			Enabled:      true,
		}
		return scriptComp.AddScript(script)
	})

	// Install the mod
	err := browserSystem.InstallMod(browserComp, "script-mod")
	if err != nil {
		t.Fatalf("failed to install mod: %v", err)
	}

	// Process download
	browserSystem.downloadMod(browserComp, "script-mod")

	// Verify mod is installed
	if !browserComp.IsInstalled("script-mod") {
		t.Error("expected mod to be installed")
	}

	// Verify script was registered
	if scriptComp.GetScriptCount() != 1 {
		t.Errorf("expected 1 script, got %d", scriptComp.GetScriptCount())
	}

	script, found := scriptComp.GetScript("script-mod-script")
	if !found {
		t.Error("expected to find script from mod")
	}
	if script.ModID != "script-mod" {
		t.Errorf("expected ModID 'script-mod', got %s", script.ModID)
	}
}

// TestModBrowserRecommendationsIntegration verifies recommendations work correctly.
func TestModBrowserRecommendationsIntegration(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()

	// Set up mods with various categories and ratings
	mods := []ModListing{
		{ID: "gameplay-1", Name: "Combat Mod", Rating: 4.5, Categories: []string{"gameplay"}, Featured: true},
		{ID: "gameplay-2", Name: "Quest Mod", Rating: 4.0, Categories: []string{"gameplay", "quests"}},
		{ID: "graphics-1", Name: "Graphics Mod", Rating: 4.8, Categories: []string{"graphics"}},
		{ID: "audio-1", Name: "Audio Mod", Rating: 3.5, Categories: []string{"audio"}},
		{ID: "ui-1", Name: "UI Mod", Rating: 4.2, Categories: []string{"ui"}},
	}
	comp.SetAvailableMods(mods)

	// Install a gameplay mod
	comp.SetInstalled("gameplay-1", true)

	// Get recommendations
	recommended := system.GetRecommendedMods(comp, 3)

	if len(recommended) != 3 {
		t.Errorf("expected 3 recommendations, got %d", len(recommended))
	}

	// gameplay-1 should not be in recommendations (already installed)
	for _, mod := range recommended {
		if mod.ID == "gameplay-1" {
			t.Error("installed mod should not be recommended")
		}
	}

	// gameplay-2 should be highly ranked (same category as installed mod)
	foundGameplay2 := false
	for _, mod := range recommended {
		if mod.ID == "gameplay-2" {
			foundGameplay2 = true
			break
		}
	}
	if !foundGameplay2 {
		t.Error("expected gameplay-2 to be recommended (same category)")
	}
}

// TestModBrowserCategoryFilterIntegration tests category filtering works.
func TestModBrowserCategoryFilterIntegration(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	mods := []ModListing{
		{ID: "mod1", Name: "Mod 1", Categories: []string{"gameplay"}},
		{ID: "mod2", Name: "Mod 2", Categories: []string{"graphics"}},
		{ID: "mod3", Name: "Mod 3", Categories: []string{"gameplay", "balance"}},
		{ID: "mod4", Name: "Mod 4", Categories: []string{"audio"}},
	}
	comp.SetAvailableMods(mods)

	// Filter by gameplay category
	gameplayMods := system.GetModsByCategory(comp, "gameplay")
	if len(gameplayMods) != 2 {
		t.Errorf("expected 2 gameplay mods, got %d", len(gameplayMods))
	}

	// Filter by graphics category
	graphicsMods := system.GetModsByCategory(comp, "graphics")
	if len(graphicsMods) != 1 {
		t.Errorf("expected 1 graphics mod, got %d", len(graphicsMods))
	}

	// Clear filter
	allMods := system.GetModsByCategory(comp, "")
	if len(allMods) != 4 {
		t.Errorf("expected 4 mods with no filter, got %d", len(allMods))
	}
}

// TestModBrowserSearchIntegration tests search functionality.
func TestModBrowserSearchIntegration(t *testing.T) {
	world := NewWorld()
	system := NewModBrowserSystem(world)

	comp := NewModBrowserComponent()
	mods := []ModListing{
		{ID: "mod1", Name: "Combat Overhaul", Description: "Improves combat"},
		{ID: "mod2", Name: "Graphics Enhancement", Description: "Better visuals"},
		{ID: "mod3", Name: "Combat AI", Description: "Smarter enemies"},
		{ID: "mod4", Name: "Sound Pack", Description: "Combat sounds"},
	}
	comp.SetAvailableMods(mods)

	// Search by name
	combatMods := system.SearchMods(comp, "combat")
	if len(combatMods) != 3 { // mod1, mod3, mod4 (combat in name or description)
		t.Errorf("expected 3 combat-related mods, got %d", len(combatMods))
	}

	// Search by specific term
	graphicsMods := system.SearchMods(comp, "graphics")
	if len(graphicsMods) != 1 {
		t.Errorf("expected 1 graphics mod, got %d", len(graphicsMods))
	}
}
