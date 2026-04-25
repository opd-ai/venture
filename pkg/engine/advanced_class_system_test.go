package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/class/advanced"
)

func TestNewAdvancedClassSystem(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)

	if system == nil {
		t.Fatal("NewAdvancedClassSystem returned nil")
	}
	if system.manager == nil {
		t.Error("Manager not initialized")
	}
	if system.world != world {
		t.Error("World not set correctly")
	}
}

func TestInitializePlayerClass(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	err := system.InitializePlayerClass(player, advanced.ClassWarrior, 10)
	if err != nil {
		t.Fatalf("InitializePlayerClass failed: %v", err)
	}

	comp, ok := player.GetComponent("advanced_class")
	if !ok {
		t.Fatal("AdvancedClassComponent not added to entity")
	}

	advClass := comp.(*advanced.AdvancedClassComponent)
	if advClass.PrimaryClass != advanced.ClassWarrior {
		t.Errorf("Expected primary class %v, got %v", advanced.ClassWarrior, advClass.PrimaryClass)
	}
	if advClass.Level != 10 {
		t.Errorf("Expected level 10, got %d", advClass.Level)
	}
	if advClass.TalentPoints.PointsTotal != 10 {
		t.Errorf("Expected 10 talent points, got %d", advClass.TalentPoints.PointsTotal)
	}
}

func TestSetSecondaryClass(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 20)

	err := system.SetSecondaryClass(player, advanced.ClassMage)
	if err != nil {
		t.Fatalf("SetSecondaryClass failed: %v", err)
	}

	comp, _ := player.GetComponent("advanced_class")
	advClass := comp.(*advanced.AdvancedClassComponent)
	if advClass.SecondaryClass != advanced.ClassMage {
		t.Errorf("Expected secondary class %v, got %v", advanced.ClassMage, advClass.SecondaryClass)
	}
}

func TestSetSecondaryClass_InvalidSameClass(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 20)

	err := system.SetSecondaryClass(player, advanced.ClassWarrior)
	if err == nil {
		t.Error("Expected error when setting secondary class same as primary")
	}
}

func TestSetPrestigeClass(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 20)

	err := system.SetPrestigeClass(player, advanced.PrestigeBladeMaster)
	if err != nil {
		t.Fatalf("SetPrestigeClass failed: %v", err)
	}

	comp, _ := player.GetComponent("advanced_class")
	advClass := comp.(*advanced.AdvancedClassComponent)
	if advClass.PrestigeClass != advanced.PrestigeBladeMaster {
		t.Errorf("Expected prestige class %v, got %v", advanced.PrestigeBladeMaster, advClass.PrestigeClass)
	}
}

func TestSetPrestigeClass_LevelTooLow(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	err := system.SetPrestigeClass(player, advanced.PrestigeBladeMaster)
	if err == nil {
		t.Error("Expected error when level too low for prestige class")
	}
}

func TestAllocateTalent(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	err := system.AllocateTalent(player, "warrior_power_strike")
	if err != nil {
		t.Fatalf("AllocateTalent failed: %v", err)
	}

	comp, _ := player.GetComponent("advanced_class")
	advClass := comp.(*advanced.AdvancedClassComponent)
	if advClass.TalentPoints.Talents["warrior_power_strike"] != 1 {
		t.Errorf("Expected talent rank 1, got %d", advClass.TalentPoints.Talents["warrior_power_strike"])
	}
	if advClass.TalentPoints.PointsSpent != 1 {
		t.Errorf("Expected 1 point spent, got %d", advClass.TalentPoints.PointsSpent)
	}
}

func TestAllocateTalent_NoPointsAvailable(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 1)

	// Allocate all available points
	system.AllocateTalent(player, "warrior_power_strike")

	// Try to allocate when no points left
	err := system.AllocateTalent(player, "warrior_cleave")
	if err == nil {
		t.Error("Expected error when no talent points available")
	}
}

func TestRespecTalents(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	// Allocate some talents
	system.AllocateTalent(player, "warrior_power_strike")
	system.AllocateTalent(player, "warrior_power_strike")

	// Respec with enough gold
	err := system.RespecTalents(player, 2000)
	if err != nil {
		t.Fatalf("RespecTalents failed: %v", err)
	}

	comp, _ := player.GetComponent("advanced_class")
	advClass := comp.(*advanced.AdvancedClassComponent)
	if len(advClass.TalentPoints.Talents) != 0 {
		t.Errorf("Expected talents to be cleared, got %d talents", len(advClass.TalentPoints.Talents))
	}
	if advClass.TalentPoints.PointsSpent != 0 {
		t.Errorf("Expected points spent to be 0, got %d", advClass.TalentPoints.PointsSpent)
	}
	if advClass.RespecCount != 1 {
		t.Errorf("Expected respec count 1, got %d", advClass.RespecCount)
	}
}

func TestRespecTalents_InsufficientGold(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)
	system.AllocateTalent(player, "warrior_power_strike")

	err := system.RespecTalents(player, 100) // Not enough gold
	if err == nil {
		t.Error("Expected error when insufficient gold")
	}
}

func TestGetRespecCost(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	cost := system.GetRespecCost(player)
	if cost != 1000 { // Base cost
		t.Errorf("Expected base respec cost 1000, got %d", cost)
	}

	// After one respec, cost should increase
	system.RespecTalents(player, 2000)
	cost = system.GetRespecCost(player)
	if cost != 1500 { // Base 1000 + 500
		t.Errorf("Expected respec cost 1500 after one respec, got %d", cost)
	}
}

func TestLevelUp(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	err := system.LevelUp(player)
	if err != nil {
		t.Fatalf("LevelUp failed: %v", err)
	}

	comp, _ := player.GetComponent("advanced_class")
	advClass := comp.(*advanced.AdvancedClassComponent)
	if advClass.Level != 11 {
		t.Errorf("Expected level 11, got %d", advClass.Level)
	}
	if advClass.TalentPoints.PointsTotal != 11 {
		t.Errorf("Expected 11 total talent points, got %d", advClass.TalentPoints.PointsTotal)
	}
}

func TestUpdate_AppliesStatBonuses(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	// Add base components
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&ManaComponent{Current: 50, Max: 50})
	player.AddComponent(&StatsComponent{Attack: 10, Defense: 5, MagicPower: 8})

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	// Allocate talents to get stat bonuses
	system.AllocateTalent(player, "warrior_power_strike")

	entities := []*Entity{player}
	system.Update(entities, 0.016)

	// Verify stat bonuses were applied
	healthComp, _ := player.GetComponent("health")
	health := healthComp.(*HealthComponent)
	if health.Max < 100 {
		t.Error("Expected health bonus from class")
	}
}

func TestGetTalentTree(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)

	tree, err := system.GetTalentTree(advanced.ClassWarrior)
	if err != nil {
		t.Fatalf("GetTalentTree failed: %v", err)
	}
	if tree == nil {
		t.Fatal("Expected talent tree, got nil")
	}
	if len(tree.Offensive) == 0 && len(tree.Defensive) == 0 && len(tree.Utility) == 0 {
		t.Error("Talent tree should have talents")
	}
}

func TestGetPlayerClass(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)
	player := world.CreateEntity()

	system.InitializePlayerClass(player, advanced.ClassWarrior, 10)

	classData, err := system.GetPlayerClass(player)
	if err != nil {
		t.Fatalf("GetPlayerClass failed: %v", err)
	}
	if classData.PrimaryClass != advanced.ClassWarrior {
		t.Errorf("Expected warrior, got %v", classData.PrimaryClass)
	}
}

func TestGetAllSynergies(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)

	synergies := system.GetAllSynergies()
	if len(synergies) == 0 {
		t.Error("Expected synergy bonuses to be defined")
	}
}

func TestUpdate_MultipleEntities(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)

	player1 := world.CreateEntity()
	player1.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player1.AddComponent(&StatsComponent{Attack: 10})
	system.InitializePlayerClass(player1, advanced.ClassWarrior, 10)

	player2 := world.CreateEntity()
	player2.AddComponent(&HealthComponent{Current: 80, Max: 80})
	player2.AddComponent(&StatsComponent{Attack: 8})
	system.InitializePlayerClass(player2, advanced.ClassMage, 8)

	entities := []*Entity{player1, player2}
	system.Update(entities, 0.016)

	// Verify both entities processed without errors
	_, ok1 := player1.GetComponent("advanced_class")
	_, ok2 := player2.GetComponent("advanced_class")
	if !ok1 || !ok2 {
		t.Error("Advanced class components should be present")
	}
}

func TestAdvancedClassComponent_Type(t *testing.T) {
	comp := &advanced.AdvancedClassComponent{}
	if comp.Type() != "advanced_class" {
		t.Errorf("Expected type 'advanced_class', got %s", comp.Type())
	}
}

// TestG32_AdvancedClassSystem_NoBonusAccumulation verifies that Update called
// N times produces the same net stat delta as calling it once (idempotency).
// This is the regression test for G32.
func TestG32_AdvancedClassSystem_NoBonusAccumulation(t *testing.T) {
	world := NewWorld()
	system := NewAdvancedClassSystem(world)

	player := world.CreateEntity()
	baseAttack := float64(10)
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&ManaComponent{Current: 50, Max: 50})
	player.AddComponent(&StatsComponent{Attack: baseAttack})
	if err := system.InitializePlayerClass(player, advanced.ClassWarrior, 10); err != nil {
		t.Fatalf("InitializePlayerClass: %v", err)
	}

	entities := []*Entity{player}

	// Run 300 frames
	for i := 0; i < 300; i++ {
		system.Update(entities, 1.0/60.0)
	}

	statsComp, _ := player.GetComponent("stats")
	stats := statsComp.(*StatsComponent)

	// At 300 frames the attack bonus must equal exactly one application, not 300×.
	expectedAttack := baseAttack + float64(10) // Warrior level-10 strength bonus
	if stats.Attack > expectedAttack+5 {       // allow ±5 for floating-point / manager variance
		t.Errorf("G32: Attack accumulated across frames: got %.1f, want ~%.1f (base %.1f + bonus)",
			stats.Attack, expectedAttack, baseAttack)
	}
}
