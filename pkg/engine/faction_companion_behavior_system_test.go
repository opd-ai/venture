package engine

import (
	"testing"
)

func TestNewFactionCompanionBehaviorSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)

	if sys == nil {
		t.Fatal("NewFactionCompanionBehaviorSystem returned nil")
	}
	if sys.world != world {
		t.Error("world reference not set")
	}
	if sys.rng == nil {
		t.Error("RNG not initialized")
	}
}

func TestFactionCompanionBehaviorSystem_SetFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)

	sys.SetFactionSystem(factionSys)

	if sys.factionSystem != factionSys {
		t.Error("faction system not set")
	}
}

func TestFactionCompanionBehaviorSystem_SetGenre(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)

	sys.SetGenre("horror")

	if sys.genreID != "horror" {
		t.Errorf("genre = %s, want horror", sys.genreID)
	}
}

func TestFactionCompanionBehaviorSystem_UpdateWithoutFactionSystem(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)

	// Should not panic without faction system
	sys.Update(nil, 0.016)
}

func TestFactionCompanionBehaviorSystem_UpdateWithoutPlayer(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Should not panic without player entity
	sys.Update(nil, 0.016)
}

func TestFactionCompanionBehaviorSystem_CompanionSkipsFriendlyTarget(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player with faction tracking
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{
		FactionID:       "merchants",
		Reputation:      75, // Friendly
		IsPlayerFaction: true,
	})

	// Create companion owned by player
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50,
		Behavior:      BehaviorAggressive,
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{})

	// Create friendly faction NPC (target)
	friendlyNPC := world.CreateEntity()
	friendlyNPC.AddComponent(&PositionComponent{X: 120, Y: 120})
	friendlyNPC.AddComponent(&AIComponent{})
	friendlyNPC.AddComponent(&FactionComponent{
		FactionID:       "merchants",
		IsPlayerFaction: false,
	})

	// Set companion to target friendly NPC
	aiComp, _ := companion.GetComponent("ai")
	aiComp.(*AIComponent).Target = friendlyNPC

	// Update system - should clear target
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Verify target was cleared
	aiComp2, _ := companion.GetComponent("ai")
	if aiComp2.(*AIComponent).Target != nil {
		t.Error("companion should not target friendly faction member")
	}
}

func TestFactionCompanionBehaviorSystem_CompanionAttacksHostileTarget(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player with hostile faction reputation
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		Reputation:      -75, // Hostile
		IsPlayerFaction: true,
	})

	// Create companion owned by player
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50,
		Behavior:      BehaviorAggressive,
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{})

	// Create hostile faction NPC (target)
	hostileNPC := world.CreateEntity()
	hostileNPC.AddComponent(&PositionComponent{X: 120, Y: 120})
	hostileNPC.AddComponent(&AIComponent{})
	hostileNPC.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	// Set companion to target hostile NPC
	aiComp, _ := companion.GetComponent("ai")
	aiComp.(*AIComponent).Target = hostileNPC

	// Update system - should keep target
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Verify target remains
	aiComp2, _ := companion.GetComponent("ai")
	if aiComp2.(*AIComponent).Target == nil {
		t.Error("companion should continue targeting hostile faction member")
	}
}

func TestFactionCompanionBehaviorSystem_LoyalCompanionProtectsNeutral(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player with neutral faction reputation
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{
		FactionID:       "villagers",
		Reputation:      25, // Neutral
		IsPlayerFaction: true,
	})

	// Create high-loyalty companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       90, // Very loyal
		Behavior:      BehaviorAggressive,
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{})

	// Create neutral faction NPC
	neutralNPC := world.CreateEntity()
	neutralNPC.AddComponent(&PositionComponent{X: 120, Y: 120})
	neutralNPC.AddComponent(&AIComponent{})
	neutralNPC.AddComponent(&FactionComponent{
		FactionID:       "villagers",
		IsPlayerFaction: false,
	})

	// Set companion to target neutral NPC
	aiComp, _ := companion.GetComponent("ai")
	aiComp.(*AIComponent).Target = neutralNPC

	// Update system - loyal companion should clear target
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Verify target was cleared
	aiComp2, _ := companion.GetComponent("ai")
	if aiComp2.(*AIComponent).Target != nil {
		t.Error("loyal companion should protect neutral faction member")
	}
}

func TestFactionCompanionBehaviorSystem_FindsPriorityTarget(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player with hostile faction reputation for bandits
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		Reputation:      -80, // Very hostile
		IsPlayerFaction: true,
	})

	// Create aggressive companion without target
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50,
		Behavior:      BehaviorAggressive, // Will seek targets
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{Target: nil})

	// Create nearby hostile bandit
	bandit := world.CreateEntity()
	bandit.AddComponent(&PositionComponent{X: 150, Y: 150})
	bandit.AddComponent(&AIComponent{})
	bandit.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	// Update system - should acquire target
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Verify target was acquired
	aiComp, _ := companion.GetComponent("ai")
	if aiComp.(*AIComponent).Target != bandit {
		t.Error("aggressive companion should acquire hostile faction target")
	}
}

func TestFactionCompanionBehaviorSystem_GenreModifiers(t *testing.T) {
	tests := []struct {
		name       string
		genreID    string
		reputation int
		wantMin    int
		wantMax    int
	}{
		{"horror reduces priority", "horror", -80, 50, 80},
		{"cyberpunk adds bonus", "cyberpunk", -80, 90, 120},
		{"fantasy balanced", "fantasy", -80, 100, 110},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			sys := NewFactionCompanionBehaviorSystem(world, 12345)
			sys.SetGenre(tt.genreID)

			priority := sys.calculateTargetPriority(tt.reputation, 100)

			if priority < tt.wantMin || priority > tt.wantMax {
				t.Errorf("priority = %d, want %d-%d for %s", priority, tt.wantMin, tt.wantMax, tt.genreID)
			}
		})
	}
}

func TestFactionCompanionBehaviorSystem_SkipsOwnedByOthers(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})

	// Create companion owned by different entity
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       99999, // Not player
		CompanionType: CompanionTypePet,
		Loyalty:       50,
		Behavior:      BehaviorAggressive,
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{Target: &Entity{ID: 123}})

	// Update system - should not modify non-player companion
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Verify target unchanged
	aiComp, _ := companion.GetComponent("ai")
	if aiComp.(*AIComponent).Target == nil {
		t.Error("should not modify companions owned by others")
	}
}

func TestFactionCompanionBehaviorSystem_DefensiveSkipsTargeting(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)

	// Create player with hostile faction
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		Reputation:      -80,
		IsPlayerFaction: true,
	})

	// Create defensive companion (should not seek targets)
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       player.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50,
		Behavior:      BehaviorDefensive, // Won't seek targets
	})
	companion.AddComponent(&PositionComponent{X: 105, Y: 105})
	companion.AddComponent(&AIComponent{Target: nil})

	// Create nearby hostile
	bandit := world.CreateEntity()
	bandit.AddComponent(&PositionComponent{X: 150, Y: 150})
	bandit.AddComponent(&AIComponent{})
	bandit.AddComponent(&FactionComponent{
		FactionID:       "bandits",
		IsPlayerFaction: false,
	})

	// Update system
	entities := []*Entity{companion}
	sys.Update(entities, 0.016)

	// Defensive companion should not acquire target
	aiComp, _ := companion.GetComponent("ai")
	if aiComp.(*AIComponent).Target != nil {
		t.Error("defensive companion should not seek targets")
	}
}

func TestFactionCompanionBehaviorSystem_TargetPriorityLevels(t *testing.T) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)

	tests := []struct {
		name       string
		reputation int
		wantSign   int // -1 for negative, 0 for zero, 1 for positive
	}{
		{"hostile high priority", -80, 1},
		{"suspicious medium priority", -25, 1},
		{"neutral low priority", 25, 1},
		{"friendly no attack", 75, -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			priority := sys.calculateTargetPriority(tt.reputation, 100)

			gotSign := 0
			if priority > 0 {
				gotSign = 1
			} else if priority < 0 {
				gotSign = -1
			}

			if gotSign != tt.wantSign {
				t.Errorf("priority sign = %d, want %d for reputation %d", gotSign, tt.wantSign, tt.reputation)
			}
		})
	}
}

func BenchmarkFactionCompanionBehaviorSystem_Update(b *testing.B) {
	world := NewWorld()
	sys := NewFactionCompanionBehaviorSystem(world, 12345)
	factionSys := NewFactionSystem(world, nil)
	sys.SetFactionSystem(factionSys)
	sys.SetGenre("fantasy")

	// Create player
	player := world.CreateEntity()
	player.AddComponent(NewStubInput())
	player.AddComponent(&PositionComponent{X: 100, Y: 100})
	player.AddComponent(&FactionComponent{FactionID: "guild", Reputation: -60, IsPlayerFaction: true})

	// Create companions
	entities := make([]*Entity, 10)
	for i := 0; i < 10; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:  player.ID,
			Loyalty:  50,
			Behavior: BehaviorAggressive,
		})
		companion.AddComponent(&PositionComponent{X: float64(100 + i*10), Y: 100})
		companion.AddComponent(&AIComponent{})
		entities[i] = companion
	}

	// Create targets
	for i := 0; i < 20; i++ {
		enemy := world.CreateEntity()
		enemy.AddComponent(&PositionComponent{X: float64(200 + i*10), Y: 200})
		enemy.AddComponent(&AIComponent{})
		enemy.AddComponent(&FactionComponent{FactionID: "guild", IsPlayerFaction: false})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.Update(entities, 0.016)
	}
}
