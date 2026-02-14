package engine

import (
	"testing"
)

func TestNewCompanionFishingBonusSystem(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewCompanionFishingBonusSystem returned nil")
	}
	if sys.world != world {
		t.Error("System world not set correctly")
	}
	if sys.rng == nil {
		t.Error("System RNG not initialized")
	}
	if sys.appliedBonuses == nil {
		t.Error("System appliedBonuses map not initialized")
	}
	if sys.genreID != "fantasy" {
		t.Errorf("Expected default genre 'fantasy', got %s", sys.genreID)
	}
}

func TestCompanionFishingBonusSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}
	for _, genre := range genres {
		sys.SetGenre(genre)
		if sys.genreID != genre {
			t.Errorf("SetGenre(%s) failed, got %s", genre, sys.genreID)
		}
	}
}

func TestCompanionFishingBonusComponent_Type(t *testing.T) {
	comp := &CompanionFishingBonusComponent{}
	if comp.Type() != "companion_fishing_bonus" {
		t.Errorf("Expected type 'companion_fishing_bonus', got %s", comp.Type())
	}
}

func TestCompanionFishingBonusSystem_CalculateCompanionBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	tests := []struct {
		name           string
		compType       CompanionType
		loyalty        float64
		expectRareFish float64 // Minimum expected
		expectSpeed    float64 // Minimum expected
	}{
		{
			name:           "Pet with high loyalty",
			compType:       CompanionTypePet,
			loyalty:        100,
			expectRareFish: 1.0,
			expectSpeed:    1.1,
		},
		{
			name:           "Water elemental with high loyalty",
			compType:       CompanionTypeElemental,
			loyalty:        100,
			expectRareFish: 1.2,
			expectSpeed:    1.0,
		},
		{
			name:           "Spirit with medium loyalty",
			compType:       CompanionTypeSpirit,
			loyalty:        50,
			expectRareFish: 1.1,
			expectSpeed:    1.0,
		},
		{
			name:           "Robot with high loyalty",
			compType:       CompanionTypeRobot,
			loyalty:        100,
			expectRareFish: 1.25,
			expectSpeed:    1.0,
		},
		{
			name:           "Insect with high loyalty",
			compType:       CompanionTypeInsect,
			loyalty:        100,
			expectRareFish: 1.0,
			expectSpeed:    1.2,
		},
		{
			name:           "Undead with high loyalty",
			compType:       CompanionTypeUndead,
			loyalty:        100,
			expectRareFish: 1.1,
			expectSpeed:    1.0,
		},
		{
			name:           "Summon with high loyalty",
			compType:       CompanionTypeSummon,
			loyalty:        100,
			expectRareFish: 1.05,
			expectSpeed:    1.05,
		},
		{
			name:           "Hireling with high loyalty",
			compType:       CompanionTypeHireling,
			loyalty:        100,
			expectRareFish: 1.0,
			expectSpeed:    1.05,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bonus := sys.GetBonusForCompanionType(tt.compType, tt.loyalty)

			if bonus.RareFishBonus < tt.expectRareFish {
				t.Errorf("RareFishBonus = %f, want >= %f", bonus.RareFishBonus, tt.expectRareFish)
			}
			if bonus.CatchSpeedBonus < tt.expectSpeed {
				t.Errorf("CatchSpeedBonus = %f, want >= %f", bonus.CatchSpeedBonus, tt.expectSpeed)
			}
			if bonus.CompanionType != tt.compType {
				t.Errorf("CompanionType = %v, want %v", bonus.CompanionType, tt.compType)
			}
		})
	}
}

func TestCompanionFishingBonusSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name           string
		genre          string
		compType       CompanionType
		expectMinBonus float64
	}{
		{
			name:           "Fantasy spirit bonus",
			genre:          "fantasy",
			compType:       CompanionTypeSpirit,
			expectMinBonus: 1.3, // +20% base + 15% genre = 35%
		},
		{
			name:           "Scifi robot bonus",
			genre:          "scifi",
			compType:       CompanionTypeRobot,
			expectMinBonus: 1.4, // +30% base + 15% genre = 45%
		},
		{
			name:           "Horror undead bonus",
			genre:          "horror",
			compType:       CompanionTypeUndead,
			expectMinBonus: 1.25, // +15% base + 15% genre = 30%
		},
		{
			name:           "Cyberpunk robot bonus",
			genre:          "cyberpunk",
			compType:       CompanionTypeRobot,
			expectMinBonus: 1.35, // +30% base + 10% genre = 40%
		},
		{
			name:           "Postapoc insect speed bonus",
			genre:          "postapoc",
			compType:       CompanionTypeInsect,
			expectMinBonus: 1.3, // +20% base + 15% genre = 35% (catch speed)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewCompanionFishingBonusSystem(world, 12345)
			sys.SetGenre(tt.genre)

			bonus := sys.GetBonusForCompanionType(tt.compType, 100)

			// Check either rare fish or catch speed depending on type
			bonusValue := bonus.RareFishBonus
			if tt.compType == CompanionTypeInsect {
				bonusValue = bonus.CatchSpeedBonus
			}

			if bonusValue < tt.expectMinBonus {
				t.Errorf("Genre %s %v bonus = %f, want >= %f",
					tt.genre, tt.compType, bonusValue, tt.expectMinBonus)
			}
		})
	}
}

func TestCompanionFishingBonusSystem_Update(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)
	sys.SetGenre("fantasy")

	// Create a player entity with fishing component
	player := NewEntity(1)
	fishingComp := NewFishingComponent()
	fishingComp.State = FishingStateWaiting
	player.AddComponent(fishingComp)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Create a companion entity
	companion := NewEntity(2)
	compComp := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypeSpirit,
		Loyalty:       80,
	}
	companion.AddComponent(compComp)
	world.AddEntity(companion)

	entities := []*Entity{player, companion}

	// First update won't apply bonus due to interval check
	sys.Update(entities, 0.1)

	// Second update with enough time should apply bonus
	sys.Update(entities, 0.5)

	// Check that bonus component was added
	bonusComp, exists := player.GetComponent("companion_fishing_bonus")
	if !exists {
		t.Fatal("Companion fishing bonus component not added to player")
	}

	bonus := bonusComp.(*CompanionFishingBonusComponent)
	if bonus.RareFishMultiplier <= 1.0 {
		t.Errorf("Expected rare fish multiplier > 1.0, got %f", bonus.RareFishMultiplier)
	}
}

func TestCompanionFishingBonusSystem_NoCompanion(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Create a player entity with fishing component but no companion
	player := NewEntity(1)
	fishingComp := NewFishingComponent()
	fishingComp.State = FishingStateWaiting
	player.AddComponent(fishingComp)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	entities := []*Entity{player}

	// Update should not add bonus
	sys.Update(entities, 0.6)

	// Check that no bonus component was added
	_, exists := player.GetComponent("companion_fishing_bonus")
	if exists {
		t.Error("Bonus component should not be added without companion")
	}
}

func TestCompanionFishingBonusSystem_NotFishing(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Create a player entity with idle fishing state
	player := NewEntity(1)
	fishingComp := NewFishingComponent()
	fishingComp.State = FishingStateIdle
	player.AddComponent(fishingComp)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Create a companion entity
	companion := NewEntity(2)
	compComp := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypeSpirit,
		Loyalty:       80,
	}
	companion.AddComponent(compComp)
	world.AddEntity(companion)

	entities := []*Entity{player, companion}

	// Update should not add bonus since player is idle
	sys.Update(entities, 0.6)

	// Check that no bonus component was added
	_, exists := player.GetComponent("companion_fishing_bonus")
	if exists {
		t.Error("Bonus component should not be added when not fishing")
	}
}

func TestCompanionFishingBonusSystem_LoyaltyScaling(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Compare bonuses at different loyalty levels
	lowLoyalty := sys.GetBonusForCompanionType(CompanionTypeElemental, 25)
	highLoyalty := sys.GetBonusForCompanionType(CompanionTypeElemental, 100)

	if highLoyalty.RareFishBonus <= lowLoyalty.RareFishBonus {
		t.Errorf("High loyalty bonus %f should be > low loyalty bonus %f",
			highLoyalty.RareFishBonus, lowLoyalty.RareFishBonus)
	}
}

func TestCompanionFishingBonusSystem_MultipleCompanions(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Create a player entity with fishing component
	player := NewEntity(1)
	fishingComp := NewFishingComponent()
	fishingComp.State = FishingStateReeling
	player.AddComponent(fishingComp)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	// Create multiple companion entities
	companion1 := NewEntity(2)
	comp1 := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypePet,
		Loyalty:       50,
	}
	companion1.AddComponent(comp1)
	world.AddEntity(companion1)

	companion2 := NewEntity(3)
	comp2 := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypeRobot, // Higher rare fish bonus
		Loyalty:       100,
	}
	companion2.AddComponent(comp2)
	world.AddEntity(companion2)

	entities := []*Entity{player, companion1, companion2}

	// Update should pick the best companion bonus (robot)
	sys.Update(entities, 0.6)

	bonusComp, exists := player.GetComponent("companion_fishing_bonus")
	if !exists {
		t.Fatal("Companion fishing bonus component not added")
	}

	bonus := bonusComp.(*CompanionFishingBonusComponent)
	if bonus.CompanionTypeID != int(CompanionTypeRobot) {
		t.Errorf("Expected robot bonus to be selected, got type %d", bonus.CompanionTypeID)
	}
}

func TestCompanionFishingBonusData_CombinedValue(t *testing.T) {
	tests := []struct {
		name   string
		bonus  *CompanionFishingBonusData
		expect float64
	}{
		{
			name: "Neutral bonus",
			bonus: &CompanionFishingBonusData{
				RareFishBonus:         1.0,
				CatchSpeedBonus:       1.0,
				TensionReductionBonus: 1.0,
			},
			expect: 0.0,
		},
		{
			name: "High rare fish bonus",
			bonus: &CompanionFishingBonusData{
				RareFishBonus:         1.5,
				CatchSpeedBonus:       1.0,
				TensionReductionBonus: 1.0,
			},
			expect: 1.0, // (0.5 * 2.0) + 0 + 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bonus.combinedValue()
			if got != tt.expect {
				t.Errorf("combinedValue() = %f, want %f", got, tt.expect)
			}
		})
	}
}

func TestCompanionFishingBonusSystem_GetCompanionBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Before any updates, should return nil
	_, ok := sys.GetCompanionBonus(999)
	if ok {
		t.Error("GetCompanionBonus should return false for unknown entity")
	}
}

func TestCompanionFishingBonusSystem_RemoveBonus(t *testing.T) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Create a player entity with fishing component and companion
	player := NewEntity(1)
	fishingComp := NewFishingComponent()
	fishingComp.State = FishingStateWaiting
	player.AddComponent(fishingComp)
	player.AddComponent(NewStubInput())
	world.AddEntity(player)

	companion := NewEntity(2)
	compComp := &CompanionComponent{
		OwnerID:       1,
		CompanionType: CompanionTypeSpirit,
		Loyalty:       80,
	}
	companion.AddComponent(compComp)
	world.AddEntity(companion)

	entities := []*Entity{player, companion}

	// Apply bonus
	sys.Update(entities, 0.6)

	// Verify bonus was applied
	_, exists := player.GetComponent("companion_fishing_bonus")
	if !exists {
		t.Fatal("Bonus should be applied")
	}

	// Set fishing to idle and update with just player (no companion)
	fishingComp.State = FishingStateIdle
	sys.Update([]*Entity{player}, 0.6)

	// Verify bonus was removed
	_, exists = player.GetComponent("companion_fishing_bonus")
	if exists {
		t.Error("Bonus should be removed when not fishing")
	}
}

func BenchmarkCompanionFishingBonusSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewCompanionFishingBonusSystem(world, 12345)

	// Create multiple players with companions
	var entities []*Entity
	for i := uint64(0); i < 50; i++ {
		player := NewEntity(i * 2)
		fishingComp := NewFishingComponent()
		fishingComp.State = FishingStateWaiting
		player.AddComponent(fishingComp)
		entities = append(entities, player)

		companion := NewEntity(i*2 + 1)
		compComp := &CompanionComponent{
			OwnerID:       i * 2,
			CompanionType: CompanionType(i % 8),
			Loyalty:       50 + float64(i),
		}
		companion.AddComponent(compComp)
		entities = append(entities, companion)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.timeSinceCheck = 1.0 // Force update
		sys.Update(entities, 0.016)
	}
}
