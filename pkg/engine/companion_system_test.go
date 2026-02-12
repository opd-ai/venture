package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
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
	companion.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Test initial state

	// Process pending entities
	world.Update(0)
	if companionComp.Loyalty != 10 {
		t.Errorf("Initial loyalty = %f, want 10", companionComp.Loyalty)
	}

	// Test follow command
	system.Update(0.1)

	// Companion should start moving toward owner
	velComp, _ := companion.GetComponent("velocity")
	velocityComp, _ := velComp.(*VelocityComponent)
	if velocityComp.VX == 0 && velocityComp.VY == 0 {
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
	companion.AddComponent(&HealthComponent{Current: 100, Max: 100})

	initialLoyalty := companionComp.Loyalty

	// Process pending entities
	world.Update(0)

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
	companion.AddComponent(&HealthComponent{Current: 100, Max: 100})

	// Process pending entities
	world.Update(0)

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
	velComp, _ := companion.GetComponent("velocity")
	velocityComp, _ := velComp.(*VelocityComponent)
	if velocityComp.VX != 0 || velocityComp.VY != 0 {
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

	// Process pending entities
	world.Update(0)

	// Update - companion should move to defensive position
	system.Update(0.1)

	velComp, _ := companion.GetComponent("velocity")
	velocityComp, _ := velComp.(*VelocityComponent)
	if velocityComp.VX == 0 && velocityComp.VY == 0 {
		t.Error("Companion should move to defensive position near owner")
	}
}

func TestCompanionSystemGatherCommand(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&InventoryComponent{
		Items:     make([]*item.Item, 0),
		MaxItems:  20,
		MaxWeight: 100,
		Gold:      0,
	})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandGather},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&VelocityComponent{})
	companion.AddComponent(NewCompanionInventoryComponent(10, 50.0))

	// Create a nearby item entity
	testItem := &item.Item{
		Name:        "Test Coin",
		Description: "A shiny coin",
		Type:        item.TypeConsumable,
		Rarity:      item.RarityCommon,
		Stats: item.Stats{
			Value:  10,
			Weight: 0.1,
		},
	}
	itemEntity := world.CreateEntity()
	itemEntity.AddComponent(&PositionComponent{X: 150, Y: 100}) // 50 pixels away
	itemEntity.AddComponent(&ItemEntityComponent{Item: testItem})

	// Process pending entities
	world.Update(0)

	// First update - companion should start moving toward item
	system.Update(0.1)

	velComp, _ := companion.GetComponent("velocity")
	velocityComp := velComp.(*VelocityComponent)
	if velocityComp.VX <= 0 {
		t.Error("Companion should move toward item (positive X velocity)")
	}
}

func TestCompanionSystemGatherPickup(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(&InventoryComponent{
		Items:     make([]*item.Item, 0),
		MaxItems:  20,
		MaxWeight: 100,
		Gold:      0,
	})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandGather},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&VelocityComponent{})
	companionInv := NewCompanionInventoryComponent(10, 50.0)
	companion.AddComponent(companionInv)

	// Create an item right next to companion (within pickup range of 32)
	testItem := &item.Item{
		Name:        "Gathered Gem",
		Description: "A precious gem",
		Type:        item.TypeConsumable,
		Rarity:      item.RarityUncommon,
		Stats: item.Stats{
			Value:  50,
			Weight: 0.5,
		},
	}
	itemEntity := world.CreateEntity()
	itemEntity.AddComponent(&PositionComponent{X: 110, Y: 100}) // 10 pixels away (within pickup)
	itemEntity.AddComponent(&ItemEntityComponent{Item: testItem})

	// Process pending entities
	world.Update(0)

	initialItemCount := companionInv.GetItemCount()
	itemEntityID := itemEntity.ID

	// Update - companion should pick up the item
	system.Update(0.1)

	// Process pending entity removals
	world.Update(0)

	// Item should be in companion inventory
	if companionInv.GetItemCount() != initialItemCount+1 {
		t.Errorf("Companion inventory should have 1 item, got %d", companionInv.GetItemCount())
	}

	// Item entity should be removed from world
	if _, exists := world.GetEntity(itemEntityID); exists {
		t.Error("Item entity should be removed from world after pickup")
	}
}

func TestCompanionSystemGatherNoItems(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})

	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandGather},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&VelocityComponent{})

	// Process pending entities
	world.Update(0)

	// Update - no items, companion should stop
	system.Update(0.1)

	velComp, _ := companion.GetComponent("velocity")
	velocityComp := velComp.(*VelocityComponent)
	if velocityComp.VX != 0 || velocityComp.VY != 0 {
		t.Error("Companion should stop when no items nearby")
	}
}

func TestCompanionSystemGatherTransferToOwner(t *testing.T) {
	world := NewWorld()
	system := NewCompanionSystem(world)

	ownerInv := &InventoryComponent{
		Items:     make([]*item.Item, 0),
		MaxItems:  20,
		MaxWeight: 100,
		Gold:      0,
	}
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 100, Y: 100})
	owner.AddComponent(ownerInv)

	// Companion with NO inventory - item should go to owner
	companion := world.CreateEntity()
	companionComp := &CompanionComponent{
		OwnerID:  owner.ID,
		Commands: []CommandType{CommandGather},
	}
	companion.AddComponent(companionComp)
	companion.AddComponent(&PositionComponent{X: 100, Y: 100})
	companion.AddComponent(&VelocityComponent{})
	// Note: No CompanionInventoryComponent

	// Create an item right next to companion
	testItem := &item.Item{
		Name:        "Owner's Item",
		Description: "Goes to owner",
		Type:        item.TypeConsumable,
		Rarity:      item.RarityCommon,
		Stats: item.Stats{
			Value:  5,
			Weight: 0.1,
		},
	}
	itemEntity := world.CreateEntity()
	itemEntity.AddComponent(&PositionComponent{X: 110, Y: 100})
	itemEntity.AddComponent(&ItemEntityComponent{Item: testItem})

	// Process pending entities
	world.Update(0)

	initialOwnerItems := len(ownerInv.Items)
	itemEntityID := itemEntity.ID

	// Update - companion should pick up item for owner
	system.Update(0.1)

	// Process pending entity removals
	world.Update(0)

	// Item should be in owner inventory
	if len(ownerInv.Items) != initialOwnerItems+1 {
		t.Errorf("Owner inventory should have 1 item, got %d", len(ownerInv.Items))
	}

	// Item entity should be removed from world
	if _, exists := world.GetEntity(itemEntityID); exists {
		t.Error("Item entity should be removed from world after pickup")
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
		companion.AddComponent(&HealthComponent{Current: 100, Max: 100})
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		system.Update(0.016)
	}
}
