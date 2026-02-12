package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

// mockPetHomeProvider implements PetHomeProvider for testing
type mockPetHomeProvider struct {
	companionHomes map[uint64]string             // companionID -> houseID
	loyaltyBonuses map[uint64]map[string]float64 // companionID -> houseID -> bonus
}

func newMockPetHomeProvider() *mockPetHomeProvider {
	return &mockPetHomeProvider{
		companionHomes: make(map[uint64]string),
		loyaltyBonuses: make(map[uint64]map[string]float64),
	}
}

func (m *mockPetHomeProvider) GetCompanionHome(companionID uint64) string {
	return m.companionHomes[companionID]
}

func (m *mockPetHomeProvider) GetLoyaltyBonus(companionID uint64, houseID string) float64 {
	if bonuses, ok := m.loyaltyBonuses[companionID]; ok {
		return bonuses[houseID]
	}
	return 0.0
}

func (m *mockPetHomeProvider) assignCompanionToHouse(companionID uint64, houseID string, bonus float64) {
	m.companionHomes[companionID] = houseID
	if _, ok := m.loyaltyBonuses[companionID]; !ok {
		m.loyaltyBonuses[companionID] = make(map[string]float64)
	}
	m.loyaltyBonuses[companionID][houseID] = bonus
}

// TestCompanionLoyaltySystem_SetPetHomeProvider tests the injection method
func TestCompanionLoyaltySystem_SetPetHomeProvider(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	system := NewCompanionLoyaltySystem(world, logger)
	mockProvider := newMockPetHomeProvider()

	// Should accept provider without error
	system.SetPetHomeProvider(mockProvider)

	if system.petHomeProvider == nil {
		t.Fatal("Expected petHomeProvider to be set, got nil")
	}
}

// TestCompanionLoyaltySystem_PassiveGainWithoutHousing tests base loyalty gain without housing
func TestCompanionLoyaltySystem_PassiveGainWithoutHousing(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create companion near owner
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Update for 60 seconds (one minute) - should trigger passive gain
	system.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	companionComp := companionCompRaw.(*CompanionComponent)

	// Base loyalty gain is 0.5 per minute
	expectedLoyalty := 50.5
	if companionComp.Loyalty != expectedLoyalty {
		t.Errorf("Expected loyalty to be %.1f, got %.1f", expectedLoyalty, companionComp.Loyalty)
	}
}

// TestCompanionLoyaltySystem_PassiveGainWithHousing tests loyalty gain with housing bonuses
func TestCompanionLoyaltySystem_PassiveGainWithHousing(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	tests := []struct {
		name           string
		housingBonus   float64
		initialLoyalty float64
		expectedGain   float64 // base 0.5 + housingBonus
	}{
		{
			name:           "basic bedding",
			housingBonus:   0.2,
			initialLoyalty: 50.0,
			expectedGain:   0.7, // 0.5 base + 0.2 housing
		},
		{
			name:           "quality bedding",
			housingBonus:   0.5,
			initialLoyalty: 60.0,
			expectedGain:   1.0, // 0.5 base + 0.5 housing
		},
		{
			name:           "luxury bedding",
			housingBonus:   1.0,
			initialLoyalty: 70.0,
			expectedGain:   1.5, // 0.5 base + 1.0 housing
		},
		{
			name:           "no housing",
			housingBonus:   0.0,
			initialLoyalty: 40.0,
			expectedGain:   0.5, // 0.5 base only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create companion near owner
			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				CompanionType: CompanionTypePet,
				Loyalty:       tt.initialLoyalty,
				Level:         1,
			})
			companion.AddComponent(&PositionComponent{X: 10, Y: 10})
			world.Update(0.0)

			// Assign companion to house with bonus (if any)
			if tt.housingBonus > 0 {
				mockProvider.assignCompanionToHouse(companion.ID, "house_001", tt.housingBonus)
			}

			// Update for 60 seconds (one minute)
			system.Update(60.0)

			companionCompRaw, _ := companion.GetComponent("companion")
			companionComp := companionCompRaw.(*CompanionComponent)

			expectedLoyalty := tt.initialLoyalty + tt.expectedGain
			if companionComp.Loyalty != expectedLoyalty {
				t.Errorf("Expected loyalty to be %.1f (gain: %.1f), got %.1f",
					expectedLoyalty, tt.expectedGain, companionComp.Loyalty)
			}

			// Cleanup
			world.RemoveEntity(companion.ID)
		})
	}
}

// TestCompanionLoyaltySystem_HousingBonusAccumulation tests loyalty accumulation over time
func TestCompanionLoyaltySystem_HousingBonusAccumulation(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create companion with quality bedding (0.5 bonus)
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Assign companion to house with 0.5 bonus
	mockProvider.assignCompanionToHouse(companion.ID, "luxury_house", 0.5)

	// Update for 5 minutes (5 * 60 = 300 seconds)
	// Expected gain: 5 minutes * (0.5 base + 0.5 housing) = 5.0
	for i := 0; i < 5; i++ {
		system.Update(60.0)
	}

	companionCompRaw, _ := companion.GetComponent("companion")
	companionComp := companionCompRaw.(*CompanionComponent)

	expectedLoyalty := 50.0 + 5.0 // 5 minutes * 1.0 gain per minute
	if companionComp.Loyalty != expectedLoyalty {
		t.Errorf("Expected loyalty to be %.1f after 5 minutes, got %.1f",
			expectedLoyalty, companionComp.Loyalty)
	}
}

// TestCompanionLoyaltySystem_HousingBonusWithMaxLoyalty tests housing bonus doesn't exceed max
func TestCompanionLoyaltySystem_HousingBonusWithMaxLoyalty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create companion near max loyalty
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       99.7, // Near max
		Level:         1,
	})
	companion.AddComponent(&PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Assign companion to luxury house (1.0 bonus)
	mockProvider.assignCompanionToHouse(companion.ID, "luxury_house", 1.0)

	// Update for 60 seconds - would gain 1.5 (0.5 + 1.0), but should cap at 100
	system.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	companionComp := companionCompRaw.(*CompanionComponent)

	// Should cap at 100.0
	if companionComp.Loyalty != 100.0 {
		t.Errorf("Expected loyalty to cap at 100.0, got %.1f", companionComp.Loyalty)
	}
}

// TestCompanionLoyaltySystem_FarFromOwnerNoHousingBonus tests no gain when far from owner
func TestCompanionLoyaltySystem_FarFromOwnerNoHousingBonus(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create companion FAR from owner (outside 100 unit range)
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&PositionComponent{X: 200, Y: 200}) // Far away
	world.Update(0.0)

	// Assign companion to house with bonus
	mockProvider.assignCompanionToHouse(companion.ID, "house_001", 0.5)

	// Update for 60 seconds - should NOT gain loyalty (too far)
	system.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	companionComp := companionCompRaw.(*CompanionComponent)

	// Should remain at initial loyalty
	if companionComp.Loyalty != 50.0 {
		t.Errorf("Expected loyalty to stay at 50.0 (far from owner), got %.1f", companionComp.Loyalty)
	}
}

// TestCompanionLoyaltySystem_HousingBonusChangesMidGame tests changing housing bonuses
func TestCompanionLoyaltySystem_HousingBonusChangesMidGame(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create companion
	companion := world.CreateEntity()
	companion.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion.AddComponent(&PositionComponent{X: 10, Y: 10})
	world.Update(0.0)

	// Start with basic bedding (0.2 bonus)
	mockProvider.assignCompanionToHouse(companion.ID, "basic_house", 0.2)

	// Update for 1 minute - gain 0.7 (0.5 + 0.2)
	system.Update(60.0)

	companionCompRaw, _ := companion.GetComponent("companion")
	companionComp := companionCompRaw.(*CompanionComponent)
	if companionComp.Loyalty != 50.7 {
		t.Errorf("After basic bedding: expected 50.7, got %.1f", companionComp.Loyalty)
	}

	// Upgrade to luxury bedding (1.0 bonus)
	mockProvider.assignCompanionToHouse(companion.ID, "luxury_house", 1.0)

	// Update for 1 minute - gain 1.5 (0.5 + 1.0)
	system.Update(60.0)

	companionCompRaw, _ = companion.GetComponent("companion")
	companionComp = companionCompRaw.(*CompanionComponent)
	expectedLoyalty := 50.7 + 1.5
	if companionComp.Loyalty != expectedLoyalty {
		t.Errorf("After luxury upgrade: expected %.1f, got %.1f", expectedLoyalty, companionComp.Loyalty)
	}
}

// TestCompanionLoyaltySystem_MultipleCompanionsWithDifferentHousing tests multiple companions
func TestCompanionLoyaltySystem_MultipleCompanionsWithDifferentHousing(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Setup mock housing provider
	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create 3 companions with different housing
	companion1 := world.CreateEntity()
	companion1.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	companion1.AddComponent(&PositionComponent{X: 10, Y: 10})

	companion2 := world.CreateEntity()
	companion2.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       60.0,
		Level:         1,
	})
	companion2.AddComponent(&PositionComponent{X: 15, Y: 15})

	companion3 := world.CreateEntity()
	companion3.AddComponent(&CompanionComponent{
		OwnerID:       owner.ID,
		CompanionType: CompanionTypePet,
		Loyalty:       70.0,
		Level:         1,
	})
	companion3.AddComponent(&PositionComponent{X: 20, Y: 20})

	world.Update(0.0)

	// Different housing bonuses
	mockProvider.assignCompanionToHouse(companion1.ID, "basic_house", 0.2)
	mockProvider.assignCompanionToHouse(companion2.ID, "quality_house", 0.5)
	mockProvider.assignCompanionToHouse(companion3.ID, "luxury_house", 1.0)

	// Update for 1 minute
	system.Update(60.0)

	// Check each companion
	comp1Raw, _ := companion1.GetComponent("companion")
	comp1 := comp1Raw.(*CompanionComponent)
	expected1 := 50.0 + 0.7 // 0.5 base + 0.2 housing
	if comp1.Loyalty != expected1 {
		t.Errorf("Companion 1: expected %.1f, got %.1f", expected1, comp1.Loyalty)
	}

	comp2Raw, _ := companion2.GetComponent("companion")
	comp2 := comp2Raw.(*CompanionComponent)
	expected2 := 60.0 + 1.0 // 0.5 base + 0.5 housing
	if comp2.Loyalty != expected2 {
		t.Errorf("Companion 2: expected %.1f, got %.1f", expected2, comp2.Loyalty)
	}

	comp3Raw, _ := companion3.GetComponent("companion")
	comp3 := comp3Raw.(*CompanionComponent)
	expected3 := 70.0 + 1.5 // 0.5 base + 1.0 housing
	if comp3.Loyalty != expected3 {
		t.Errorf("Companion 3: expected %.1f, got %.1f", expected3, comp3.Loyalty)
	}
}

// BenchmarkCompanionLoyaltySystem_PassiveGainWithHousing benchmarks housing integration
func BenchmarkCompanionLoyaltySystem_PassiveGainWithHousing(b *testing.B) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	mockProvider := newMockPetHomeProvider()
	system.SetPetHomeProvider(mockProvider)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create 10 companions with housing
	for i := 0; i < 10; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:       owner.ID,
			CompanionType: CompanionTypePet,
			Loyalty:       50.0,
			Level:         1,
		})
		companion.AddComponent(&PositionComponent{X: float64(i * 5), Y: float64(i * 5)})
		world.Update(0.0)

		mockProvider.assignCompanionToHouse(companion.ID, "house_001", 0.5)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(60.0)
	}
}

// BenchmarkCompanionLoyaltySystem_PassiveGainWithoutHousing benchmarks base loyalty
func BenchmarkCompanionLoyaltySystem_PassiveGainWithoutHousing(b *testing.B) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Create owner
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})
	world.Update(0.0)

	// Create 10 companions WITHOUT housing
	for i := 0; i < 10; i++ {
		companion := world.CreateEntity()
		companion.AddComponent(&CompanionComponent{
			OwnerID:       owner.ID,
			CompanionType: CompanionTypePet,
			Loyalty:       50.0,
			Level:         1,
		})
		companion.AddComponent(&PositionComponent{X: float64(i * 5), Y: float64(i * 5)})
		world.Update(0.0)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		system.Update(60.0)
	}
}
