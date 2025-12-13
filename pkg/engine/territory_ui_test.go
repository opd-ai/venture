package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/world/territory"
)

func TestNewTerritoryUI(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)

	ui := NewTerritoryUI(sys, 800, 600)

	if ui == nil {
		t.Fatal("expected UI, got nil")
	}
	if ui.visible {
		t.Error("UI should be hidden initially")
	}
	if ui.screenWidth != 800 {
		t.Errorf("expected width 800, got %d", ui.screenWidth)
	}
	if ui.screenHeight != 600 {
		t.Errorf("expected height 600, got %d", ui.screenHeight)
	}
}

func TestTerritoryUI_SetPlayerEntity(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	player := &Entity{ID: 1, components: make(map[string]Component)}

	ui.SetPlayerEntity(player)

	if ui.playerEntity != player {
		t.Error("player entity not set correctly")
	}
}

func TestTerritoryUI_IsVisible(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	if ui.IsVisible() {
		t.Error("UI should be hidden initially")
	}

	ui.visible = true

	if !ui.IsVisible() {
		t.Error("UI should be visible")
	}
}

func TestTerritoryUI_Toggle(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Create player entity with position
	player := &Entity{ID: 1, components: make(map[string]Component)}
	player.AddComponent(&PositionComponent{X: 250, Y: 250})
	ui.SetPlayerEntity(player)

	// Create territory at player position
	sys.EnsureTerritoryExists(250, 250)

	if ui.visible {
		t.Error("should start hidden")
	}

	ui.Toggle()
	if !ui.visible {
		t.Error("should be visible after toggle")
	}

	ui.Toggle()
	if ui.visible {
		t.Error("should be hidden after second toggle")
	}
}

func TestTerritoryUI_ShowHide(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	ui.Show()
	if !ui.visible {
		t.Error("Show() should make UI visible")
	}

	ui.Hide()
	if ui.visible {
		t.Error("Hide() should make UI hidden")
	}
}

func TestTerritoryUI_GetPlayerGuildID(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// No player
	guildID := ui.getPlayerGuildID()
	if guildID != "" {
		t.Error("expected empty guild ID with no player")
	}

	// Player without guild component
	player := &Entity{ID: 1, components: make(map[string]Component)}
	ui.SetPlayerEntity(player)

	guildID = ui.getPlayerGuildID()
	if guildID != "" {
		t.Error("expected empty guild ID with no guild component")
	}

	// Player with guild component
	player.AddComponent(&GuildComponent{GuildID: "test_guild"})

	guildID = ui.getPlayerGuildID()
	if guildID != "test_guild" {
		t.Errorf("expected 'test_guild', got '%s'", guildID)
	}
}

func TestTerritoryUI_RefreshCurrentTerritory(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Create territory
	terr, _ := sys.EnsureTerritoryExists(250, 250)

	// No player - should not panic
	ui.refreshCurrentTerritory()
	if ui.selectedTerritory != nil {
		t.Error("should not select territory without player")
	}

	// Player at territory location
	player := &Entity{ID: 1, components: make(map[string]Component)}
	player.AddComponent(&PositionComponent{X: 250, Y: 250})
	ui.SetPlayerEntity(player)

	ui.refreshCurrentTerritory()
	if ui.selectedTerritory == nil {
		t.Fatal("expected territory to be selected")
	}
	if ui.selectedTerritory.ID != terr.ID {
		t.Errorf("expected territory %s, got %s", terr.ID, ui.selectedTerritory.ID)
	}
}

func TestTerritoryUI_GetStatusColor(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Just verify it returns colors without panicking
	neutralColor := ui.getStatusColor(territory.StatusNeutral)
	if neutralColor == nil {
		t.Error("expected color for neutral status")
	}

	ownedColor := ui.getStatusColor(territory.StatusOwned)
	if ownedColor == nil {
		t.Error("expected color for owned status")
	}

	contestedColor := ui.getStatusColor(territory.StatusContested)
	if contestedColor == nil {
		t.Error("expected color for contested status")
	}

	// Colors should be different
	if neutralColor == ownedColor {
		t.Error("neutral and owned colors should differ")
	}
}

func TestTerritoryUI_HandleDeclareWar(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Create two guilds and territories
	guildA := "guild_a"
	guildB := "guild_b"

	terr, _ := sys.EnsureTerritoryExists(250, 250)
	manager.AssignOwner(terr.ID, guildB)

	// Player in guild A at guild B territory
	player := &Entity{ID: 1, components: make(map[string]Component)}
	player.AddComponent(&PositionComponent{X: 250, Y: 250})
	player.AddComponent(&GuildComponent{GuildID: guildA})
	ui.SetPlayerEntity(player)
	ui.selectedTerritory = terr

	// Declare war
	ui.handleDeclareWar()

	// Verify war exists
	if !manager.IsAtWar(guildA, guildB) {
		t.Error("war should be declared")
	}

	// Calling again should not create duplicate
	ui.handleDeclareWar()
	wars := manager.GetActiveWars()
	warCount := 0
	for _, war := range wars {
		if (war.AttackerGuild == guildA && war.DefenderGuild == guildB) ||
			(war.AttackerGuild == guildB && war.DefenderGuild == guildA) {
			warCount++
		}
	}
	if warCount != 1 {
		t.Errorf("expected 1 war, got %d", warCount)
	}
}

func TestTerritoryUI_HandleDeclareWar_NoPlayer(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	terr, _ := sys.EnsureTerritoryExists(250, 250)
	ui.selectedTerritory = terr

	// Should not panic without player
	ui.handleDeclareWar()

	if len(manager.GetActiveWars()) > 0 {
		t.Error("should not declare war without player")
	}
}

func TestTerritoryUI_HandleBuildStructure(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Create guild and territory
	guildA := "guild_a"
	terr, _ := sys.EnsureTerritoryExists(250, 250)
	manager.AssignOwner(terr.ID, guildA)

	// Player in guild A at their territory
	player := &Entity{ID: 1, components: make(map[string]Component)}
	player.AddComponent(&PositionComponent{X: 250, Y: 250})
	player.AddComponent(&GuildComponent{GuildID: guildA})
	ui.SetPlayerEntity(player)
	ui.selectedTerritory = terr

	// Build structure
	ui.handleBuildStructure()

	// Verify structure exists
	updatedTerr, _ := manager.GetTerritory(terr.ID)
	if len(updatedTerr.Structures) != 1 {
		t.Fatalf("expected 1 structure, got %d", len(updatedTerr.Structures))
	}
	if updatedTerr.Structures[0].Type != territory.StructureTypeWall {
		t.Errorf("expected wall, got %v", updatedTerr.Structures[0].Type)
	}
}

func TestTerritoryUI_HandleBuildStructure_NotOwned(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Create territory owned by different guild
	guildA := "guild_a"
	guildB := "guild_b"
	terr, _ := sys.EnsureTerritoryExists(250, 250)
	manager.AssignOwner(terr.ID, guildB)

	// Player in guild A trying to build in guild B territory
	player := &Entity{ID: 1, components: make(map[string]Component)}
	player.AddComponent(&PositionComponent{X: 250, Y: 250})
	player.AddComponent(&GuildComponent{GuildID: guildA})
	ui.SetPlayerEntity(player)
	ui.selectedTerritory = terr

	// Should not build
	ui.handleBuildStructure()

	updatedTerr, _ := manager.GetTerritory(terr.ID)
	if len(updatedTerr.Structures) > 0 {
		t.Error("should not build in enemy territory")
	}
}

func TestTerritoryUI_Update(t *testing.T) {
	manager := territory.NewManager()
	sys := NewTerritorySystem(manager, nil)
	ui := NewTerritoryUI(sys, 800, 600)

	// Should not error when hidden
	err := ui.Update()
	if err != nil {
		t.Errorf("Update() error: %v", err)
	}

	// Should not error when visible
	ui.Show()
	err = ui.Update()
	if err != nil {
		t.Errorf("Update() error when visible: %v", err)
	}
}
