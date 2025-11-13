package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// TestReputationSystem_RecordAction tests recording actions
func TestReputationSystem_RecordAction(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel) // Suppress debug logs during tests
	system := NewReputationSystem(world, logger)

	// Create test entity
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition

	// Record an action
	factionImpact := map[string]float64{
		"TestFaction": 10.0,
	}
	system.RecordAction(entity.ID, "Test action", factionImpact, 0.05, 0.1, "Test Location")

	// Verify reputation was updated
	rep := system.GetReputation(entity.ID, "TestFaction")
	if rep != 10.0 {
		t.Errorf("Expected reputation 10.0, got %f", rep)
	}

	// Verify alignment was updated
	alignment := system.GetAlignment(entity.ID)
	if alignment.LawAxis != 0.05 {
		t.Errorf("Expected LawAxis 0.05, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0.1 {
		t.Errorf("Expected GoodAxis 0.1, got %f", alignment.GoodAxis)
	}

	// Verify deed was recorded
	comp, _ := entity.GetComponent("reputation")
	repComp := comp.(*ReputationComponent)
	if len(repComp.KarmaDeeds) != 1 {
		t.Errorf("Expected 1 deed, got %d", len(repComp.KarmaDeeds))
	}
}

// TestReputationSystem_RecordKill tests recording kills
func TestReputationSystem_RecordKill(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	killer := world.CreateEntity()
	world.Update(0)
	victim := world.CreateEntity()
	world.Update(0)

	// Test justified kill
	system.RecordKill(killer.ID, victim.ID, true)

	alignment := system.GetAlignment(killer.ID)
	// Justified kill: slight lawful (+0.01), slight evil (-0.01)
	if alignment.LawAxis != 0.01 {
		t.Errorf("Expected LawAxis 0.01 for justified kill, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != -0.01 {
		t.Errorf("Expected GoodAxis -0.01 for justified kill, got %f", alignment.GoodAxis)
	}

	// Reset and test unjustified kill
	killer2 := world.CreateEntity()
	world.Update(0)
	victim2 := world.CreateEntity()
	world.Update(0)

	system.RecordKill(killer2.ID, victim2.ID, false)

	alignment2 := system.GetAlignment(killer2.ID)
	// Unjustified kill: chaotic (-0.05), evil (-0.05)
	if alignment2.LawAxis != -0.05 {
		t.Errorf("Expected LawAxis -0.05 for unjustified kill, got %f", alignment2.LawAxis)
	}
	if alignment2.GoodAxis != -0.05 {
		t.Errorf("Expected GoodAxis -0.05 for unjustified kill, got %f", alignment2.GoodAxis)
	}
}

// TestReputationSystem_RecordHelp tests recording helpful actions
func TestReputationSystem_RecordHelp(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	helper := world.CreateEntity()
	world.Update(0)
	target := world.CreateEntity()
	world.Update(0)

	system.RecordHelp(helper.ID, target.ID)

	alignment := system.GetAlignment(helper.ID)
	// Helping: slight lawful (+0.01), good (+0.02)
	if alignment.LawAxis != 0.01 {
		t.Errorf("Expected LawAxis 0.01, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0.02 {
		t.Errorf("Expected GoodAxis 0.02, got %f", alignment.GoodAxis)
	}
}

// TestReputationSystem_RecordTheft tests recording theft
func TestReputationSystem_RecordTheft(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	thief := world.CreateEntity()
	world.Update(0)

	// Small theft
	system.RecordTheft(thief.ID, "Guards", 50.0)

	rep := system.GetReputation(thief.ID, "Guards")
	// Base -10.0 + (50/100) = -10.5
	if rep != -10.5 {
		t.Errorf("Expected reputation -10.5, got %f", rep)
	}

	alignment := system.GetAlignment(thief.ID)
	// Theft: chaotic (-0.05), evil (-0.03)
	if alignment.LawAxis != -0.05 {
		t.Errorf("Expected LawAxis -0.05, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != -0.03 {
		t.Errorf("Expected GoodAxis -0.03, got %f", alignment.GoodAxis)
	}

	// Large theft (should cap at -30)
	thief2 := world.CreateEntity()
	world.Update(0)
	system.RecordTheft(thief2.ID, "Merchants", 10000.0)

	rep2 := system.GetReputation(thief2.ID, "Merchants")
	if rep2 != -30.0 {
		t.Errorf("Expected capped reputation -30.0, got %f", rep2)
	}
}

// TestReputationSystem_RecordQuestCompletion tests quest completion
func TestReputationSystem_RecordQuestCompletion(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	player := world.CreateEntity()
	world.Update(0)

	// Easy quest (difficulty 0.5)
	system.RecordQuestCompletion(player.ID, "Adventurers", 0.5)

	rep := system.GetReputation(player.ID, "Adventurers")
	expected := 10.0 + (0.5 * 20.0) // 10 + 10 = 20
	if rep != expected {
		t.Errorf("Expected reputation %f, got %f", expected, rep)
	}

	// Hard quest (difficulty 3.0, should cap at 50)
	player2 := world.CreateEntity()
	world.Update(0)
	system.RecordQuestCompletion(player2.ID, "Warriors", 3.0)

	rep2 := system.GetReputation(player2.ID, "Warriors")
	if rep2 != 50.0 {
		t.Errorf("Expected capped reputation 50.0, got %f", rep2)
	}

	alignment := system.GetAlignment(player2.ID)
	// Quest completion: lawful (+0.02), good (+0.01)
	if alignment.LawAxis != 0.02 {
		t.Errorf("Expected LawAxis 0.02, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0.01 {
		t.Errorf("Expected GoodAxis 0.01, got %f", alignment.GoodAxis)
	}
}

// TestReputationSystem_GetReputation tests reputation retrieval
func TestReputationSystem_GetReputation(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	// Non-existent entity
	rep := system.GetReputation(9999, "AnyFaction")
	if rep != 0.0 {
		t.Errorf("Expected 0.0 for non-existent entity, got %f", rep)
	}

	// Entity without reputation component
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	rep = system.GetReputation(entity.ID, "AnyFaction")
	if rep != 0.0 {
		t.Errorf("Expected 0.0 for entity without reputation, got %f", rep)
	}

	// Entity with reputation
	factionImpact := map[string]float64{"TestFaction": 25.0}
	system.RecordAction(entity.ID, "Test", factionImpact, 0, 0, "")

	rep = system.GetReputation(entity.ID, "TestFaction")
	if rep != 25.0 {
		t.Errorf("Expected 25.0, got %f", rep)
	}
}

// TestReputationSystem_GetAlignment tests alignment retrieval
func TestReputationSystem_GetAlignment(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	// Non-existent entity
	alignment := system.GetAlignment(9999)
	if alignment.LawAxis != 0.0 || alignment.GoodAxis != 0.0 {
		t.Error("Expected neutral alignment for non-existent entity")
	}

	// Entity without reputation component
	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	alignment = system.GetAlignment(entity.ID)
	if alignment.LawAxis != 0.0 || alignment.GoodAxis != 0.0 {
		t.Error("Expected neutral alignment for entity without reputation")
	}

	// Entity with modified alignment
	system.RecordAction(entity.ID, "Test", nil, 0.5, 0.3, "")

	alignment = system.GetAlignment(entity.ID)
	if alignment.LawAxis != 0.5 {
		t.Errorf("Expected LawAxis 0.5, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0.3 {
		t.Errorf("Expected GoodAxis 0.3, got %f", alignment.GoodAxis)
	}
}

// TestReputationSystem_InvalidEntity tests handling of invalid entities
func TestReputationSystem_InvalidEntity(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	// RecordAction with invalid entity (should not panic)
	system.RecordAction(9999, "Test", nil, 0, 0, "")

	// RecordKill with invalid entities (should not panic)
	system.RecordKill(9999, 8888, true)

	validEntity := world.CreateEntity()
	system.RecordKill(validEntity.ID, 9999, true)
	system.RecordKill(9999, validEntity.ID, true)

	// RecordHelp with invalid entities (should not panic)
	system.RecordHelp(9999, 8888)
	system.RecordHelp(validEntity.ID, 9999)
	system.RecordHelp(9999, validEntity.ID)
}

// TestReputationSystem_Update tests the Update method
func TestReputationSystem_Update(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	// Update should not panic (it's event-driven, so it does nothing)
	system.Update(0.016)
}

// TestReputationSystem_CumulativeActions tests multiple actions
func TestReputationSystem_CumulativeActions(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	player := world.CreateEntity()
	world.Update(0)

	// Perform multiple good actions
	for i := 0; i < 5; i++ {
		system.RecordAction(player.ID, "Good deed", map[string]float64{"Villagers": 5.0}, 0.02, 0.03, "")
	}

	rep := system.GetReputation(player.ID, "Villagers")
	if rep != 25.0 {
		t.Errorf("Expected cumulative reputation 25.0 (5x5), got %f", rep)
	}

	alignment := system.GetAlignment(player.ID)
	if alignment.LawAxis != 0.1 {
		t.Errorf("Expected cumulative LawAxis 0.1 (5x0.02), got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0.15 {
		t.Errorf("Expected cumulative GoodAxis 0.15 (5x0.03), got %f", alignment.GoodAxis)
	}
}

// BenchmarkReputationSystem_RecordAction benchmarks action recording
func BenchmarkReputationSystem_RecordAction(b *testing.B) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	factionImpact := map[string]float64{"TestFaction": 10.0}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.RecordAction(entity.ID, "Benchmark action", factionImpact, 0.01, 0.01, "")
	}
}

// BenchmarkReputationSystem_GetReputation benchmarks reputation retrieval
func BenchmarkReputationSystem_GetReputation(b *testing.B) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	system := NewReputationSystem(world, logger)

	entity := world.CreateEntity()
	world.Update(0) // Process entity addition
	system.RecordAction(entity.ID, "Setup", map[string]float64{"TestFaction": 10.0}, 0, 0, "")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.GetReputation(entity.ID, "TestFaction")
	}
}
