package engine

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestCompanionLoyaltySystem_New(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel) // Suppress logs in tests

	system := NewCompanionLoyaltySystem(world, logger)

	if system == nil {
		t.Fatal("NewCompanionLoyaltySystem returned nil")
	}
	if system.world != world {
		t.Error("System world not set correctly")
	}
	if len(system.PendingLoyaltyChanges) != 0 {
		t.Error("PendingLoyaltyChanges should be empty initially")
	}
}

func TestCompanionLoyaltySystem_QueueLoyaltyChange(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	change := LoyaltyChange{
		CompanionID: 1,
		Amount:      10.0,
		Reason:      LoyaltyReasonFed,
	}

	system.QueueLoyaltyChange(change)

	if len(system.PendingLoyaltyChanges) != 1 {
		t.Errorf("Expected 1 pending change, got %d", len(system.PendingLoyaltyChanges))
	}

	if system.PendingLoyaltyChanges[0].CompanionID != 1 {
		t.Errorf("Expected CompanionID 1, got %d", system.PendingLoyaltyChanges[0].CompanionID)
	}
}

func TestCompanionLoyaltySystem_ModifyLoyalty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Create a companion entity
	entity := world.CreateEntity()
	entity.AddComponent(&CompanionComponent{
		OwnerID:       0,
		CompanionType: CompanionTypePet,
		Loyalty:       50.0,
		Level:         1,
	})
	entity.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Process pending entity additions
	world.Update(0.0)

	tests := []struct {
		name           string
		initialLoyalty float64
		amount         float64
		reason         LoyaltyChangeReason
		expectedMin    float64
		expectedMax    float64
	}{
		{
			name:           "positive change",
			initialLoyalty: 50.0,
			amount:         10.0,
			reason:         LoyaltyReasonFed,
			expectedMin:    60.0,
			expectedMax:    60.0,
		},
		{
			name:           "negative change",
			initialLoyalty: 50.0,
			amount:         -20.0,
			reason:         LoyaltyReasonDamaged,
			expectedMin:    30.0,
			expectedMax:    30.0,
		},
		{
			name:           "clamped at max",
			initialLoyalty: 95.0,
			amount:         10.0,
			reason:         LoyaltyReasonVictory,
			expectedMin:    100.0,
			expectedMax:    100.0,
		},
		{
			name:           "clamped at min",
			initialLoyalty: 5.0,
			amount:         -10.0,
			reason:         LoyaltyReasonAbandoned,
			expectedMin:    0.0,
			expectedMax:    0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset loyalty
			companionCompRaw, _ := entity.GetComponent("companion")
			companionComp := companionCompRaw.(*CompanionComponent)
			companionComp.Loyalty = tt.initialLoyalty

			system.ModifyLoyalty(entity.ID, tt.amount, tt.reason)
			system.Update(0.0) // Process pending changes

			companionCompRaw, _ = entity.GetComponent("companion")
			companionComp = companionCompRaw.(*CompanionComponent)
			finalLoyalty := companionComp.Loyalty
			if finalLoyalty < tt.expectedMin || finalLoyalty > tt.expectedMax {
				t.Errorf("Expected loyalty between %.1f and %.1f, got %.1f",
					tt.expectedMin, tt.expectedMax, finalLoyalty)
			}
		})
	}
}

func TestCompanionLoyaltySystem_PassiveLoyaltyGain(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	// Create owner entity
	owner := world.CreateEntity()
	owner.AddComponent(&PositionComponent{X: 0, Y: 0})

	// Process pending entity additions
	world.Update(0.0)

	tests := []struct {
		name           string
		companionPos   PositionComponent
		initialLoyalty float64
		updateTime     float64
		shouldIncrease bool
	}{
		{
			name:           "near owner",
			companionPos:   PositionComponent{X: 10, Y: 10},
			initialLoyalty: 50.0,
			updateTime:     60.0, // One minute
			shouldIncrease: true,
		},
		{
			name:           "far from owner",
			companionPos:   PositionComponent{X: 200, Y: 200},
			initialLoyalty: 50.0,
			updateTime:     60.0,
			shouldIncrease: false,
		},
		{
			name:           "at max loyalty",
			companionPos:   PositionComponent{X: 10, Y: 10},
			initialLoyalty: 100.0,
			updateTime:     60.0,
			shouldIncrease: false, // Already at max
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create companion entity
			companion := world.CreateEntity()
			companion.AddComponent(&CompanionComponent{
				OwnerID:       owner.ID,
				CompanionType: CompanionTypePet,
				Loyalty:       tt.initialLoyalty,
				Level:         1,
			})
			companion.AddComponent(&tt.companionPos)

			// Process pending entity additions
			world.Update(0.0)

			system.Update(tt.updateTime)

			companionCompRaw, _ := companion.GetComponent("companion")
			companionComp := companionCompRaw.(*CompanionComponent)
			if tt.shouldIncrease {
				if companionComp.Loyalty <= tt.initialLoyalty {
					t.Errorf("Expected loyalty to increase from %.1f, got %.1f",
						tt.initialLoyalty, companionComp.Loyalty)
				}
			} else {
				if companionComp.Loyalty != tt.initialLoyalty {
					t.Errorf("Expected loyalty to stay at %.1f, got %.1f",
						tt.initialLoyalty, companionComp.Loyalty)
				}
			}

			// Cleanup
			world.RemoveEntity(companion.ID)
		})
	}
}

func TestCompanionLoyaltySystem_Disobedience(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	tests := []struct {
		name             string
		loyalty          float64
		initialBehavior  BehaviorMode
		expectedBehavior BehaviorMode
	}{
		{
			name:             "very low loyalty becomes passive",
			loyalty:          20.0,
			initialBehavior:  BehaviorAggressive,
			expectedBehavior: BehaviorPassive,
		},
		{
			name:             "medium-low loyalty switches to defensive",
			loyalty:          35.0,
			initialBehavior:  BehaviorAggressive,
			expectedBehavior: BehaviorDefensive,
		},
		{
			name:             "high loyalty stays aggressive",
			loyalty:          80.0,
			initialBehavior:  BehaviorAggressive,
			expectedBehavior: BehaviorAggressive,
		},
		{
			name:             "defensive at medium loyalty",
			loyalty:          50.0,
			initialBehavior:  BehaviorDefensive,
			expectedBehavior: BehaviorDefensive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity := world.CreateEntity()
			entity.AddComponent(&CompanionComponent{
				OwnerID:       0,
				CompanionType: CompanionTypePet,
				Loyalty:       tt.loyalty,
				Level:         1,
				Behavior:      tt.initialBehavior,
			})

			// Process pending entity additions
			world.Update(0.0)

			system.Update(0.0)

			companionCompRaw, _ := entity.GetComponent("companion")
			companionComp := companionCompRaw.(*CompanionComponent)
			if companionComp.Behavior != tt.expectedBehavior {
				t.Errorf("Expected behavior %v, got %v", tt.expectedBehavior, companionComp.Behavior)
			}

			world.RemoveEntity(entity.ID)
		})
	}
}

func TestCompanionLoyaltySystem_GetLoyalty(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	entity := world.CreateEntity()
	entity.AddComponent(&CompanionComponent{
		OwnerID:       0,
		CompanionType: CompanionTypePet,
		Loyalty:       75.5,
		Level:         1,
	})

	// Process pending entity additions
	world.Update(0.0)

	loyalty := system.GetLoyalty(entity.ID)
	if loyalty != 75.5 {
		t.Errorf("Expected loyalty 75.5, got %.1f", loyalty)
	}

	// Test non-existent entity
	loyalty = system.GetLoyalty(99999)
	if loyalty != 0.0 {
		t.Errorf("Expected loyalty 0.0 for non-existent entity, got %.1f", loyalty)
	}
}

func TestCompanionLoyaltySystem_GetLoyaltyThreshold(t *testing.T) {
	world := NewWorld()
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)
	system := NewCompanionLoyaltySystem(world, logger)

	tests := []struct {
		loyalty  float64
		expected string
	}{
		{90.0, "Devoted"},
		{80.0, "Devoted"},
		{70.0, "Loyal"},
		{60.0, "Loyal"},
		{50.0, "Neutral"},
		{40.0, "Neutral"},
		{30.0, "Distant"},
		{20.0, "Distant"},
		{10.0, "Rebellious"},
		{0.0, "Rebellious"},
	}

	for _, tt := range tests {
		result := system.GetLoyaltyThreshold(tt.loyalty)
		if result != tt.expected {
			t.Errorf("Loyalty %.1f: expected %s, got %s", tt.loyalty, tt.expected, result)
		}
	}
}

func TestLoyaltyChangeReason_String(t *testing.T) {
	tests := []struct {
		reason   LoyaltyChangeReason
		expected string
	}{
		{LoyaltyReasonFed, "Fed"},
		{LoyaltyReasonHealed, "Healed"},
		{LoyaltyReasonDamaged, "Damaged"},
		{LoyaltyReasonAbandoned, "Abandoned"},
		{LoyaltyReasonTimeTogether, "Time Together"},
		{LoyaltyReasonOwnerDied, "Owner Died"},
		{LoyaltyReasonVictory, "Victory"},
		{LoyaltyChangeReason(999), "Unknown"},
	}

	for _, tt := range tests {
		result := tt.reason.String()
		if result != tt.expected {
			t.Errorf("Expected %s, got %s", tt.expected, result)
		}
	}
}
