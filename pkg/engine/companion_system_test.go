package engine

import (
	"testing"
)

func TestCompanionSystem(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&VelocityComponent{})

	// Create companion
	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       10,
		Level:         1,
		Behavior:      BehaviorDefensive,
		Commands:      []CommandType{CommandFollow},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 200, Y: 100})
	companion.AddComponent(&VelocityComponent{})
	companion.AddComponent(&HealthComponent{Health: 100, MaxHealth: 100})

	// Test initial state
	if companionComp.Loyalty != 10 {
		t.Errorf("Initial loyalty = %f, want 10", companionComp.Loyalty)
	}

	// Test follow command
	system.Update(0.1)

	// Companion should start moving toward owner
	velocityComp, _ := companion.GetComponent("velocity").(*VelocityComponent)
	if velocityComp.Vx == 0 && velocityComp.Vy == 0 {
		t.Error("Companion should be moving toward owner")
	}
}

func TestCompanionSystemBonding(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       0,
		BondingPerks:  []BondingPerk{},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100}) // Same position (bonding range)
	companion.AddComponent(&HealthComponent{Health: 100, MaxHealth: 100})

	initialLoyalty := companionComp.Loyalty

	// Update for bonding time
	for i := 0; i < 100; i++ {
		system.Update(0.1) // 10 seconds total
	}

	// Loyalty should increase
	if companionComp.Loyalty <= initialLoyalty {
		t.Errorf("Loyalty should increase, got %f, initial %f", companionComp.Loyalty, initialLoyalty)
	}

	// Time with owner should accumulate
	if companionComp.TimeWithOwner == 0 {
		t.Error("TimeWithOwner should accumulate when near owner")
	}
}

func TestCompanionSystemPerkUnlock(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:       owner.ID,
		Loyalty:       20,
		TimeWithOwner: 300, // 5 minutes - enough for first perk
		BondingPerks:  []BondingPerk{},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&HealthComponent{Health: 100, MaxHealth: 100})

	// Update to trigger perk check
	system.Update(0.1)

	// Should have unlocked PerkExtraHealth (300s, 20% loyalty)
	if !system.HasPerk(companion, PerkExtraHealth) {
		t.Error("Should have unlocked PerkExtraHealth with 300s bonding time and 20% loyalty")
	}
}

func TestCompanionSystemCommands(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandFollow},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&VelocityComponent{})

	// Test issuing new command
	err := system.IssueCommand(companion, CommandStay)
	if err != nil {
		t.Fatalf("IssueCommand() failed: %v", err)
	}

	if len(companionComp.Commands) != 1 || companionComp.Commands[0] != CommandStay {
		t.Error("Command should be updated to CommandStay")
	}

	// Test Stay command stops movement
	system.Update(0.1)
	velocityComp, _ := companion.GetComponent("velocity").(*VelocityComponent)
	if velocityComp.Vx != 0 || velocityComp.Vy != 0 {
		t.Error("Companion should stop when commanded to Stay")
	}
}

func TestCompanionSystemGetLoyalty(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		Loyalty: 75.5,
	}
	companion.AddComponent(companionComp)

	loyalty := system.GetLoyalty(companion)
	if loyalty != 75.5 {
		t.Errorf("GetLoyalty() = %f, want 75.5", loyalty)
	}
}

func TestCompanionSystemDefendCommand(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandDefend},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 300, Y: 100}) // Far from owner
	companion.AddComponent(&VelocityComponent{})

	// Update - companion should move to defensive position
	system.Update(0.1)

	velocityComp, _ := companion.GetComponent("velocity").(*VelocityComponent)
	if velocityComp.Vx == 0 && velocityComp.Vy == 0 {
		t.Error("Companion should move to defensive position near owner")
	}
}

func BenchmarkCompanionSystemUpdate(b *testing.B) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	// Create 50 owner-companion pairs
	for i := 0; i < 50; i++ {
		owner := world.CreateEntity()
		owner.AddComponent(&PositionComponent{X: float64(i * 10), Y: 0})

		companion := world.CreateEntity()
		companionComp := &CompanionComponent{
			OwnerID:  owner.ID,
			Loyalty:  50,
			Commands: []CommandType{CommandFollow},
		}
		companion.AddComponent(companionComp)
		companion.AddComponent(&PositionComponent{X: float64(i*10 + 100), Y: 0})
		companion.AddComponent(&VelocityComponent{})
		companion.AddComponent(&HealthComponent{Health: 100, MaxHealth: 100})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}
