// Package engine tests the integration between CraftingSystem and housing_crafting package.
package engine

import (
	"testing"
)

// MockStationManager implements StationBonusProvider for testing.
type MockStationManager struct {
	bonuses map[string]map[string]float64 // playerID -> recipeID -> bonus
}

// NewMockStationManager creates a new mock station manager.
func NewMockStationManager() *MockStationManager {
	return &MockStationManager{
		bonuses: make(map[string]map[string]float64),
	}
}

// GetCraftingBonus returns the crafting bonus for a player/recipe pair.
func (m *MockStationManager) GetCraftingBonus(playerID, recipeID string) float64 {
	if recipes, ok := m.bonuses[playerID]; ok {
		if bonus, ok := recipes[recipeID]; ok {
			return bonus
		}
	}
	return 1.0 // No bonus
}

// SetBonus configures a bonus for testing.
func (m *MockStationManager) SetBonus(playerID, recipeID string, bonus float64) {
	if m.bonuses[playerID] == nil {
		m.bonuses[playerID] = make(map[string]float64)
	}
	m.bonuses[playerID][recipeID] = bonus
}

// TestCraftingSystem_SetStationManager verifies station manager injection.
func TestCraftingSystem_SetStationManager(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	mockManager := NewMockStationManager()
	craftingSystem.SetStationManager(mockManager)

	if craftingSystem.stationManager == nil {
		t.Error("SetStationManager failed to set station manager")
	}
}

// TestCraftingSystem_ExtractStationBonus_AutoDiscovery tests auto-discovery via StationManager.
func TestCraftingSystem_ExtractStationBonus_AutoDiscovery(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	// Create a player entity with network component
	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})
	world.Update(0) // Flush entity additions

	// Configure mock station manager with bonus
	mockManager := NewMockStationManager()
	mockManager.SetBonus("12345", "iron_sword", 1.5) // 1.5x multiplier
	craftingSystem.SetStationManager(mockManager)

	// Extract bonus should auto-discover from housing stations
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", 0)

	// Multiplier 1.5 converts to success bonus 0.5
	expectedBonus := 0.5
	if bonus != expectedBonus {
		t.Errorf("Expected auto-discovered bonus %.2f, got %.2f", expectedBonus, bonus)
	}
}

// TestCraftingSystem_ExtractStationBonus_NoBonus tests no bonus case.
func TestCraftingSystem_ExtractStationBonus_NoBonus(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	// Create a player entity
	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})
	world.Update(0) // Flush entity additions

	// Configure mock with no bonuses
	mockManager := NewMockStationManager()
	craftingSystem.SetStationManager(mockManager)

	// Should return 0.0 (no bonus)
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", 0)
	if bonus != 0.0 {
		t.Errorf("Expected no bonus (0.0), got %.2f", bonus)
	}
}

// TestCraftingSystem_ExtractStationBonus_FallbackToEntity tests fallback to entity station.
func TestCraftingSystem_ExtractStationBonus_FallbackToEntity(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	// Create player entity
	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})

	// Create station entity with bonus
	stationEntity := world.CreateEntity()
	stationEntity.AddComponent(&CraftingStationComponent{
		StationType:         RecipePotion,
		BonusSuccessChance:  0.25,
		CraftTimeMultiplier: 0.9,
	})
	world.Update(0) // Flush entity additions

	// No station manager configured - should fallback to entity lookup
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", stationEntity.ID)
	if bonus != 0.25 {
		t.Errorf("Expected entity station bonus 0.25, got %.2f", bonus)
	}
}

// TestCraftingSystem_ExtractStationBonus_NoManager tests behavior without manager.
func TestCraftingSystem_ExtractStationBonus_NoManager(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})
	world.Update(0) // Flush entity additions

	// No station manager, no station ID
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", 0)
	if bonus != 0.0 {
		t.Errorf("Expected no bonus without manager or station, got %.2f", bonus)
	}
}

// TestCraftingSystem_ExtractStationBonus_NoNetworkComponent tests entity without network component.
func TestCraftingSystem_ExtractStationBonus_NoNetworkComponent(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	// Entity without network component
	playerEntity := world.CreateEntity()

	mockManager := NewMockStationManager()
	craftingSystem.SetStationManager(mockManager)

	// Should return 0.0 since we can't get player ID
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", 0)
	if bonus != 0.0 {
		t.Errorf("Expected no bonus without network component, got %.2f", bonus)
	}
}

// TestCraftingSystem_ExtractStationBonus_MultipleRecipes tests different recipes.
func TestCraftingSystem_ExtractStationBonus_MultipleRecipes(t *testing.T) {
	tests := []struct {
		name            string
		recipeID        string
		configuredBonus float64
		expectedBonus   float64
	}{
		{"basic_quality", "iron_sword", 1.2, 0.2},
		{"standard_quality", "steel_sword", 1.5, 0.5},
		{"advanced_quality", "mithril_sword", 1.8, 0.8},
		{"master_quality", "adamantine_sword", 2.0, 1.0},
		{"no_station", "wooden_club", 1.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			invSystem := NewInventorySystem(world)
			craftingSystem := NewCraftingSystem(world, invSystem, nil)

			playerEntity := world.CreateEntity()
			playerEntity.AddComponent(&NetworkComponent{
				PlayerID: 99999,
				Synced:   true,
			})
			world.Update(0) // Flush entity additions

			mockManager := NewMockStationManager()
			if tt.configuredBonus > 1.0 {
				mockManager.SetBonus("99999", tt.recipeID, tt.configuredBonus)
			}
			craftingSystem.SetStationManager(mockManager)

			bonus := craftingSystem.extractStationBonus(playerEntity.ID, tt.recipeID, 0)
			const epsilon = 0.0001
			if diff := bonus - tt.expectedBonus; diff < -epsilon || diff > epsilon {
				t.Errorf("Recipe %s: expected bonus %.2f, got %.2f", tt.recipeID, tt.expectedBonus, bonus)
			}
		})
	}
}

// TestCraftingSystem_ExtractStationBonus_PriorityOrder tests that housing bonus takes priority.
func TestCraftingSystem_ExtractStationBonus_PriorityOrder(t *testing.T) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})

	// Create station entity with lower bonus
	stationEntity := world.CreateEntity()
	stationEntity.AddComponent(&CraftingStationComponent{
		StationType:         RecipePotion,
		BonusSuccessChance:  0.15,
		CraftTimeMultiplier: 0.9,
	})
	world.Update(0) // Flush entity additions

	// Configure housing manager with higher bonus
	mockManager := NewMockStationManager()
	mockManager.SetBonus("12345", "iron_sword", 1.5) // 0.5 bonus
	craftingSystem.SetStationManager(mockManager)

	// Should use housing bonus (0.5), not entity bonus (0.15)
	bonus := craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", stationEntity.ID)
	if bonus != 0.5 {
		t.Errorf("Expected housing bonus to take priority (0.5), got %.2f", bonus)
	}
}

// BenchmarkCraftingSystem_ExtractStationBonus_AutoDiscovery benchmarks auto-discovery.
func BenchmarkCraftingSystem_ExtractStationBonus_AutoDiscovery(b *testing.B) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	playerEntity := world.CreateEntity()
	playerEntity.AddComponent(&NetworkComponent{
		PlayerID: 12345,
		Synced:   true,
	})
	world.Update(0) // Flush entity additions

	mockManager := NewMockStationManager()
	mockManager.SetBonus("12345", "iron_sword", 1.5)
	craftingSystem.SetStationManager(mockManager)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", 0)
	}
}

// BenchmarkCraftingSystem_ExtractStationBonus_EntityFallback benchmarks entity lookup.
func BenchmarkCraftingSystem_ExtractStationBonus_EntityFallback(b *testing.B) {
	world := NewWorld()
	invSystem := NewInventorySystem(world)
	craftingSystem := NewCraftingSystem(world, invSystem, nil)

	playerEntity := world.CreateEntity()
	stationEntity := world.CreateEntity()
	stationEntity.AddComponent(&CraftingStationComponent{
		StationType:         RecipePotion,
		BonusSuccessChance:  0.25,
		CraftTimeMultiplier: 0.9,
	})
	world.Update(0) // Flush entity additions

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		craftingSystem.extractStationBonus(playerEntity.ID, "iron_sword", stationEntity.ID)
	}
}
