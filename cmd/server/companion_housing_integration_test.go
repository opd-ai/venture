//go:build !android && !ios
// +build !android,!ios

package main

import (
	"testing"

	"github.com/opd-ai/venture/pkg/engine"
	companionhousing "github.com/opd-ai/venture/pkg/integration/companion_housing"
	"github.com/sirupsen/logrus"
)

// TestCompanionHousingIntegration_ServerWiring verifies PetHomeManager integration
func TestCompanionHousingIntegration_ServerWiring(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	// Initialize the loyalty system
	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)

	// Initialize the pet home manager
	petHomeMgr := companionhousing.NewPetHomeManager()

	// Wire them together (this is what main.go does)
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	// Verify the wiring succeeded (system should accept the provider)
	if companionLoyaltySystem == nil {
		t.Fatal("CompanionLoyaltySystem is nil")
	}

	if petHomeMgr == nil {
		t.Fatal("PetHomeManager is nil")
	}

	// Test data flow: create a companion and assign housing
	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	companion := world.CreateEntity()
	companion.AddComponent(&engine.CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Add bedding and assign companion
	petHomeMgr.AddBedding("house_001", "bedding_001", companionhousing.BeddingBasic)
	err := petHomeMgr.AssignCompanionToBed(companion.ID, "bedding_001")
	if err != nil {
		t.Fatalf("Failed to assign companion to bed: %v", err)
	}

	// Get initial loyalty
	companionCompRaw, _ := companion.GetComponent("companion")
	initialLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	// Update system for 1 minute - should apply housing bonus
	companionLoyaltySystem.Update(60.0)

	// Check that loyalty increased
	companionCompRaw, _ = companion.GetComponent("companion")
	finalLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	if finalLoyalty <= initialLoyalty {
		t.Errorf("Expected loyalty to increase from %.1f, got %.1f", initialLoyalty, finalLoyalty)
	}

	// Basic bedding has 0.05 bonus/day, base is 0.5, so total should be 0.55
	expectedGain := 0.55
	actualGain := finalLoyalty - initialLoyalty
	if actualGain != expectedGain {
		t.Errorf("Expected loyalty gain of %.2f, got %.2f", expectedGain, actualGain)
	}
}

// TestCompanionHousingIntegration_LuxuryBedding tests high-quality housing bonuses
func TestCompanionHousingIntegration_LuxuryBedding(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	companion := world.CreateEntity()
	companion.AddComponent(&engine.CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Luxury bedding provides 1.0 bonus
	petHomeMgr.AddBedding("luxury_house", "luxury_bed", companionhousing.BeddingLuxury)
	err := petHomeMgr.AssignCompanionToBed(companion.ID, "luxury_bed")
	if err != nil {
		t.Fatalf("Failed to assign companion to luxury bed: %v", err)
	}

	// Update for 1 minute
	companionLoyaltySystem.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	finalLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	// Luxury: 0.5 base + 0.2 housing = 0.7 gain
	expectedLoyalty := 50.7
	if finalLoyalty != expectedLoyalty {
		t.Errorf("Expected loyalty %.2f, got %.2f", expectedLoyalty, finalLoyalty)
	}
}

// TestCompanionHousingIntegration_NoHousingFallback tests without housing (base loyalty only)
func TestCompanionHousingIntegration_NoHousingFallback(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	companion := world.CreateEntity()
	companion.AddComponent(&engine.CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// No housing assignment - should use base loyalty only

	// Update for 1 minute
	companionLoyaltySystem.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	finalLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	// Only base gain: 0.5
	expectedLoyalty := 50.5
	if finalLoyalty != expectedLoyalty {
		t.Errorf("Expected base loyalty %.1f, got %.1f", expectedLoyalty, finalLoyalty)
	}
}

// TestCompanionHousingIntegration_MultipleMinutes tests accumulation over time
func TestCompanionHousingIntegration_MultipleMinutes(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	companion := world.CreateEntity()
	companion.AddComponent(&engine.CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Quality bedding: 0.5 bonus
	petHomeMgr.AddBedding("quality_house", "quality_bed", companionhousing.BeddingStandard)
	err := petHomeMgr.AssignCompanionToBed(companion.ID, "quality_bed")
	if err != nil {
		t.Fatalf("Failed to assign companion to bed: %v", err)
	}

	// Update for 5 minutes (5 * 60 = 300 seconds)
	for i := 0; i < 5; i++ {
		companionLoyaltySystem.Update(60.0)
	}

	companionCompRaw, _ := companion.GetComponent("companion")
	finalLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	// 5 minutes * (0.5 base + 0.1 housing) = 3.0 gain
	expectedLoyalty := 53.0
	if finalLoyalty != expectedLoyalty {
		t.Errorf("After 5 minutes: expected %.2f, got %.2f", expectedLoyalty, finalLoyalty)
	}
}

// TestCompanionHousingIntegration_UpgradeBedding tests changing housing mid-game
func TestCompanionHousingIntegration_UpgradeBedding(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	companion := world.CreateEntity()
	companion.AddComponent(&engine.CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: engine.CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&engine.PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Start with basic bedding
	petHomeMgr.AddBedding("house_001", "basic_bed", companionhousing.BeddingBasic)
	petHomeMgr.AssignCompanionToBed(companion.ID, "basic_bed")

	// 1 minute with basic (0.5 + 0.05 = 0.55 gain)
	companionLoyaltySystem.Update(60.0)

	// Upgrade to luxury
	petHomeMgr.UnassignCompanionBed(companion.ID)
	petHomeMgr.AddBedding("luxury_house", "luxury_bed", companionhousing.BeddingLuxury)
	petHomeMgr.AssignCompanionToBed(companion.ID, "luxury_bed")

	// 1 minute with luxury (0.5 + 0.2 = 0.7 gain)
	companionLoyaltySystem.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	finalLoyalty := companionCompRaw.(*engine.CompanionComponent).Loyalty

	// 50.0 + 0.55 (basic) + 0.7 (luxury) = 51.25
	expectedLoyalty := 51.25
	if finalLoyalty != expectedLoyalty {
		t.Errorf("After upgrade: expected %.2f, got %.2f", expectedLoyalty, finalLoyalty)
	}
}

// BenchmarkCompanionHousingIntegration_10Companions benchmarks 10 companions with housing
func BenchmarkCompanionHousingIntegration_10Companions(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create 10 companions with housing
	for i := 0; i < 10; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&engine.CompanionComponent{
			OwnerID:       owner.ID,
			CompanionType: engine.CompanionTypePet,
			Loyalty:       50.0,
			Level:         1,
		})
		companion.AddComponent(&engine.PositionComponent{X: float64(i * 5), Y: float64(i * 5)})
		world.Update(0.0)

		petHomeMgr.AddBedding("house_001", "bed_"+string(rune(i)), companionhousing.BeddingStandard)
		petHomeMgr.AssignCompanionToBed(companion.ID, "bed_"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		companionLoyaltySystem.Update(60.0)
	}
}

// BenchmarkCompanionHousingIntegration_100Companions benchmarks 100 companions
func BenchmarkCompanionHousingIntegration_100Companions(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	world := engine.NewWorld()

	companionLoyaltySystem := engine.NewCompanionLoyaltySystem(world, logger)
	petHomeMgr := companionhousing.NewPetHomeManager()
	companionLoyaltySystem.SetPetHomeProvider(petHomeMgr)

	owner := world.CreateEntity()
	owner.AddComponent(&engine.PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create 100 companions with housing
	for i := 0; i < 100; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&engine.CompanionComponent{
			OwnerID:       owner.ID,
			CompanionType: engine.CompanionTypePet,
			Loyalty:       50.0,
			Level:         1,
		})
		companion.AddComponent(&engine.PositionComponent{X: float64(i % 10 * 5), Y: float64(i / 10 * 5)})
		world.Update(0.0)

		petHomeMgr.AddBedding("house_001", "bed_"+string(rune(i)), companionhousing.BeddingStandard)
		petHomeMgr.AssignCompanionToBed(companion.ID, "bed_"+string(rune(i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		companionLoyaltySystem.Update(60.0)
	}
}
