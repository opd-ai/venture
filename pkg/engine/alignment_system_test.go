package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// createAlignmentTestWorld creates a test world and flushes entities after creation.
func createAlignmentTestWorld() *World {
	return NewWorld()
}

// createAlignmentTestEntity creates an entity and flushes pending additions.
func createAlignmentTestEntity(world *World) *Entity {
	entity := world.CreateEntity()
	world.FlushPendingEntities()
	return entity
}

func TestNewAlignmentSystem(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	if system == nil {
		t.Fatal("NewAlignmentSystem returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.logger != nil {
		t.Error("expected nil logger when not provided")
	}
}

func TestNewAlignmentSystemWithLogger(t *testing.T) {
	world := createAlignmentTestWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	system := NewAlignmentSystemWithLogger(world, logger)

	if system == nil {
		t.Fatal("NewAlignmentSystemWithLogger returned nil")
	}
	if system.world != world {
		t.Error("world not set correctly")
	}
	if system.logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestNewAlignmentSystemWithNilLogger(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystemWithLogger(world, nil)

	if system == nil {
		t.Fatal("NewAlignmentSystemWithLogger returned nil")
	}
	if system.logger != nil {
		t.Error("expected nil logger when nil provided")
	}
}

func TestAlignmentSystemUpdate(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	// Update should not panic even with no entities
	system.Update(0.016)
}

func TestRecordDeed_NewEntity(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	err := system.RecordDeed(entityID, "Help someone", 0.05, 0.10)
	if err != nil {
		t.Fatalf("RecordDeed failed: %v", err)
	}

	// Verify reputation component was created
	comp, ok := entity.GetComponent("reputation")
	if !ok {
		t.Fatal("reputation component not created")
	}

	repComp := comp.(*ReputationComponent)
	if repComp.Alignment.LawAxis != 0.05 {
		t.Errorf("expected LawAxis 0.05, got %f", repComp.Alignment.LawAxis)
	}
	if repComp.Alignment.GoodAxis != 0.10 {
		t.Errorf("expected GoodAxis 0.10, got %f", repComp.Alignment.GoodAxis)
	}

	// Verify deed was recorded
	if len(repComp.KarmaDeeds) != 1 {
		t.Errorf("expected 1 deed in history, got %d", len(repComp.KarmaDeeds))
	}
	if repComp.KarmaDeeds[0].Description != "Help someone" {
		t.Errorf("expected deed description 'Help someone', got %s", repComp.KarmaDeeds[0].Description)
	}
}

func TestRecordDeed_ExistingComponent(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	// Add pre-existing reputation component
	repComp := NewReputationComponent()
	repComp.Alignment.LawAxis = 0.5
	repComp.Alignment.GoodAxis = 0.5
	entity.AddComponent(repComp)

	err := system.RecordDeed(entityID, "Break law", -0.1, -0.05)
	if err != nil {
		t.Fatalf("RecordDeed failed: %v", err)
	}

	// Verify alignment was adjusted
	if repComp.Alignment.LawAxis != 0.4 {
		t.Errorf("expected LawAxis 0.4, got %f", repComp.Alignment.LawAxis)
	}
	if repComp.Alignment.GoodAxis != 0.45 {
		t.Errorf("expected GoodAxis 0.45, got %f", repComp.Alignment.GoodAxis)
	}
}

func TestRecordDeed_NonExistentEntity(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	// Record deed for non-existent entity
	err := system.RecordDeed(99999, "Test", 0.1, 0.1)
	if err != nil {
		t.Errorf("expected nil error for non-existent entity, got %v", err)
	}
}

func TestRecordDeed_WithLogger(t *testing.T) {
	world := createAlignmentTestWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)
	system := NewAlignmentSystemWithLogger(world, logger)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	// Should log without error
	err := system.RecordDeed(entityID, "Logged deed", 0.05, 0.05)
	if err != nil {
		t.Fatalf("RecordDeed with logger failed: %v", err)
	}
}

func TestGetAlignment_EntityWithComponent(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	repComp := NewReputationComponent()
	repComp.Alignment.LawAxis = 0.75
	repComp.Alignment.GoodAxis = -0.25
	entity.AddComponent(repComp)

	alignment := system.GetAlignment(entityID)

	if alignment.LawAxis != 0.75 {
		t.Errorf("expected LawAxis 0.75, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != -0.25 {
		t.Errorf("expected GoodAxis -0.25, got %f", alignment.GoodAxis)
	}
}

func TestGetAlignment_EntityWithoutComponent(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	alignment := system.GetAlignment(entityID)

	if alignment.LawAxis != 0 {
		t.Errorf("expected default LawAxis 0, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0 {
		t.Errorf("expected default GoodAxis 0, got %f", alignment.GoodAxis)
	}
}

func TestGetAlignment_NonExistentEntity(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	alignment := system.GetAlignment(99999)

	if alignment.LawAxis != 0 {
		t.Errorf("expected default LawAxis 0, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0 {
		t.Errorf("expected default GoodAxis 0, got %f", alignment.GoodAxis)
	}
}

func TestGetAlignmentDescription(t *testing.T) {
	tests := []struct {
		name     string
		lawAxis  float64
		goodAxis float64
		expected string
	}{
		{"True Neutral", 0.0, 0.0, "True Neutral"},
		{"True Neutral (near zero)", 0.1, 0.1, "True Neutral"},
		{"Lawful Good", 0.8, 0.8, "Lawful Good"},
		{"Lawful Evil", 0.8, -0.8, "Lawful Evil"},
		{"Chaotic Good", -0.8, 0.8, "Chaotic Good"},
		{"Chaotic Evil", -0.8, -0.8, "Chaotic Evil"},
		{"Neutral Good", 0.0, 0.8, "Good"},
		{"Neutral Evil", 0.0, -0.8, "Evil"},
		{"Lawful Neutral", 0.8, 0.0, "Lawful"},
		{"Chaotic Neutral", -0.8, 0.0, "Chaotic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := createAlignmentTestWorld()
			system := NewAlignmentSystem(world)

			entity := createAlignmentTestEntity(world)
			entityID := entity.ID

			repComp := NewReputationComponent()
			repComp.Alignment.LawAxis = tt.lawAxis
			repComp.Alignment.GoodAxis = tt.goodAxis
			entity.AddComponent(repComp)

			description := system.GetAlignmentDescription(entityID)

			if description != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, description)
			}
		})
	}
}

func TestGetAlignmentDescription_NonExistentEntity(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	description := system.GetAlignmentDescription(99999)

	if description != "True Neutral" {
		t.Errorf("expected 'True Neutral' for non-existent entity, got '%s'", description)
	}
}

func TestRecordCommonDeed(t *testing.T) {
	tests := []struct {
		deed             string
		expectedLawAxis  float64
		expectedGoodAxis float64
	}{
		{DeedKillInnocent, -0.05, -0.1},
		{DeedKillHostile, -0.01, 0.02},
		{DeedSteal, -0.05, -0.05},
		{DeedHelp, 0.01, 0.05},
		{DeedBreakLaw, -0.08, 0},
		{DeedUpholdLaw, 0.08, 0.02},
		{DeedLie, -0.02, -0.02},
		{DeedTellTruth, 0.02, 0.01},
		{DeedBetray, -0.1, -0.15},
		{DeedHonorAgreement, 0.05, 0.03},
		{DeedDonateToCharity, 0, 0.08},
		{DeedRobPoor, -0.05, -0.12},
		{DeedProtectWeak, 0.03, 0.10},
		{DeedExploitWeak, -0.03, -0.10},
		{DeedFulfillContract, 0.06, 0.02},
		{DeedBreakContract, -0.06, -0.03},
		{DeedSacrificeForGreed, 0, -0.15},
		{DeedSacrificeForOthers, 0.05, 0.20},
	}

	for _, tt := range tests {
		t.Run(tt.deed, func(t *testing.T) {
			world := createAlignmentTestWorld()
			system := NewAlignmentSystem(world)

			entity := createAlignmentTestEntity(world)
			entityID := entity.ID

			err := system.RecordCommonDeed(entityID, tt.deed)
			if err != nil {
				t.Fatalf("RecordCommonDeed failed: %v", err)
			}

			alignment := system.GetAlignment(entityID)

			if alignment.LawAxis != tt.expectedLawAxis {
				t.Errorf("expected LawAxis %f, got %f", tt.expectedLawAxis, alignment.LawAxis)
			}
			if alignment.GoodAxis != tt.expectedGoodAxis {
				t.Errorf("expected GoodAxis %f, got %f", tt.expectedGoodAxis, alignment.GoodAxis)
			}
		})
	}
}

func TestRecordCommonDeed_UnknownDeed(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	err := system.RecordCommonDeed(entityID, "Unknown Deed")
	if err != nil {
		t.Fatalf("RecordCommonDeed failed for unknown deed: %v", err)
	}

	alignment := system.GetAlignment(entityID)

	// Unknown deeds should not change alignment
	if alignment.LawAxis != 0 {
		t.Errorf("expected LawAxis 0 for unknown deed, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis != 0 {
		t.Errorf("expected GoodAxis 0 for unknown deed, got %f", alignment.GoodAxis)
	}
}

func TestAlignmentClamping(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	// Apply massive positive alignment shifts
	for i := 0; i < 100; i++ {
		_ = system.RecordDeed(entityID, "Heroic act", 1.0, 1.0)
	}

	alignment := system.GetAlignment(entityID)

	// Alignment should be clamped to 1.0
	if alignment.LawAxis > 1.0 {
		t.Errorf("LawAxis should be clamped to 1.0, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis > 1.0 {
		t.Errorf("GoodAxis should be clamped to 1.0, got %f", alignment.GoodAxis)
	}
}

func TestAlignmentClampingNegative(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	// Apply massive negative alignment shifts
	for i := 0; i < 100; i++ {
		_ = system.RecordDeed(entityID, "Evil act", -1.0, -1.0)
	}

	alignment := system.GetAlignment(entityID)

	// Alignment should be clamped to -1.0
	if alignment.LawAxis < -1.0 {
		t.Errorf("LawAxis should be clamped to -1.0, got %f", alignment.LawAxis)
	}
	if alignment.GoodAxis < -1.0 {
		t.Errorf("GoodAxis should be clamped to -1.0, got %f", alignment.GoodAxis)
	}
}

func TestDeedHistoryRecording(t *testing.T) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)

	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	// Record multiple deeds
	deeds := []string{DeedHelp, DeedSteal, DeedProtectWeak}
	for _, deed := range deeds {
		_ = system.RecordCommonDeed(entityID, deed)
	}

	comp, _ := entity.GetComponent("reputation")
	repComp := comp.(*ReputationComponent)

	if len(repComp.KarmaDeeds) != 3 {
		t.Errorf("expected 3 deeds in history, got %d", len(repComp.KarmaDeeds))
	}

	// Verify deeds are in order
	for i, deed := range deeds {
		if repComp.KarmaDeeds[i].Description != deed {
			t.Errorf("expected deed %d to be '%s', got '%s'", i, deed, repComp.KarmaDeeds[i].Description)
		}
	}
}

func BenchmarkRecordDeed(b *testing.B) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)
	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.RecordDeed(entityID, "Test", 0.01, 0.01)
	}
}

func BenchmarkGetAlignment(b *testing.B) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)
	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	repComp := NewReputationComponent()
	entity.AddComponent(repComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetAlignment(entityID)
	}
}

func BenchmarkGetAlignmentDescription(b *testing.B) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)
	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	repComp := NewReputationComponent()
	repComp.Alignment.LawAxis = 0.75
	repComp.Alignment.GoodAxis = 0.75
	entity.AddComponent(repComp)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.GetAlignmentDescription(entityID)
	}
}

func BenchmarkRecordCommonDeed(b *testing.B) {
	world := createAlignmentTestWorld()
	system := NewAlignmentSystem(world)
	entity := createAlignmentTestEntity(world)
	entityID := entity.ID

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = system.RecordCommonDeed(entityID, DeedHelp)
	}
}
