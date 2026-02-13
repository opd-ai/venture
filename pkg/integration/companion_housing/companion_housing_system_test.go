package companion_housing

import (
	"testing"
	"time"
)

func TestNewCompanionHousingSystem(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	if system == nil {
		t.Fatal("NewCompanionHousingSystem() returned nil")
	}
	if system.manager != manager {
		t.Error("NewCompanionHousingSystem() did not set manager correctly")
	}
}

func TestCompanionHousingSystem_IsInHouse(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	tests := []struct {
		name         string
		ownerHouseID string
		want         bool
	}{
		{"with house assigned", "house_1", true},
		{"without house assigned", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompanionHousingComponent{OwnerHouseID: tt.ownerHouseID}
			got := system.IsInHouse(c)
			if got != tt.want {
				t.Errorf("IsInHouse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionHousingSystem_HasBedding(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	tests := []struct {
		name      string
		beddingID string
		want      bool
	}{
		{"with bedding", "bed_1", true},
		{"without bedding", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompanionHousingComponent{BeddingID: tt.beddingID}
			got := system.HasBedding(c)
			if got != tt.want {
				t.Errorf("HasBedding() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionHousingSystem_IsTraining(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	tests := []struct {
		name           string
		activeTraining string
		want           bool
	}{
		{"with active training", "training_1", true},
		{"without training", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompanionHousingComponent{ActiveTraining: tt.activeTraining}
			got := system.IsTraining(c)
			if got != tt.want {
				t.Errorf("IsTraining() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionHousingSystem_HasSharedStorage(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	tests := []struct {
		name              string
		sharedChestAccess []string
		want              bool
	}{
		{"with shared storage", []string{"chest_1", "chest_2"}, true},
		{"empty shared storage", []string{}, false},
		{"nil shared storage", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompanionHousingComponent{SharedChestAccess: tt.sharedChestAccess}
			got := system.HasSharedStorage(c)
			if got != tt.want {
				t.Errorf("HasSharedStorage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionHousingSystem_DaysSinceRest(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	baseTime := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name         string
		lastRestTime time.Time
		now          time.Time
		want         float64
	}{
		{"zero rest time", time.Time{}, baseTime, 0.0},
		{"same time", baseTime, baseTime, 0.0},
		{"one day ago", baseTime.Add(-24 * time.Hour), baseTime, 1.0},
		{"half day ago", baseTime.Add(-12 * time.Hour), baseTime, 0.5},
		{"two days ago", baseTime.Add(-48 * time.Hour), baseTime, 2.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CompanionHousingComponent{LastRestTime: tt.lastRestTime}
			got := system.DaysSinceRest(c, tt.now)
			// Use approximate comparison for floating point
			if diff := got - tt.want; diff < -0.0001 || diff > 0.0001 {
				t.Errorf("DaysSinceRest() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompanionHousingSystem_DaysSinceRest_Determinism(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	// Fixed times for determinism
	restTime := time.Date(2026, 1, 18, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 1, 20, 12, 0, 0, 0, time.UTC)

	c := &CompanionHousingComponent{LastRestTime: restTime}

	// Call multiple times - should return same result
	result1 := system.DaysSinceRest(c, now)
	result2 := system.DaysSinceRest(c, now)
	result3 := system.DaysSinceRest(c, now)

	if result1 != result2 || result2 != result3 {
		t.Errorf("DaysSinceRest() not deterministic: %v, %v, %v", result1, result2, result3)
	}

	if result1 != 2.0 {
		t.Errorf("DaysSinceRest() = %v, want 2.0", result1)
	}
}

func TestCompanionHousingSystem_UpdateFromManager(t *testing.T) {
	manager := NewPetHomeManager()
	system := NewCompanionHousingSystem(manager)

	// Set up housing
	manager.AddBedding("house_1", "bed_1", BeddingAdvanced)
	manager.AssignCompanionToBed(100, "bed_1")
	manager.AddTrainingArea("house_1", "training_1", TrainingCombat)
	// Use explicit time for determinism
	sessionTime := time.Date(2026, 2, 13, 14, 30, 0, 0, time.UTC)
	manager.StartTrainingSession(100, "training_1", sessionTime)

	t.Run("assigned companion", func(t *testing.T) {
		c := &CompanionHousingComponent{}
		system.UpdateFromManager(c, 100)

		if c.OwnerHouseID != "house_1" {
			t.Errorf("OwnerHouseID = %s, want house_1", c.OwnerHouseID)
		}
		// BeddingAdvanced = 1.5 * 0.1 = 0.15
		if diff := c.LoyaltyBonus - 0.15; diff < -0.0001 || diff > 0.0001 {
			t.Errorf("LoyaltyBonus = %v, want 0.15", c.LoyaltyBonus)
		}
		// TrainingCombat = 1.5
		if c.TrainingBonus != 1.5 {
			t.Errorf("TrainingBonus = %v, want 1.5", c.TrainingBonus)
		}
	})

	t.Run("unassigned companion", func(t *testing.T) {
		c := &CompanionHousingComponent{
			OwnerHouseID:  "old_house",
			LoyaltyBonus:  0.5,
			TrainingBonus: 2.0,
		}
		system.UpdateFromManager(c, 999) // Companion with no home

		if c.OwnerHouseID != "" {
			t.Errorf("OwnerHouseID = %s, want empty string", c.OwnerHouseID)
		}
		if c.LoyaltyBonus != 0.0 {
			t.Errorf("LoyaltyBonus = %v, want 0.0", c.LoyaltyBonus)
		}
		if c.TrainingBonus != 1.0 {
			t.Errorf("TrainingBonus = %v, want 1.0", c.TrainingBonus)
		}
	})
}

func TestCompanionHousingComponent_Type(t *testing.T) {
	c := &CompanionHousingComponent{}
	got := c.Type()
	want := "companion_housing"

	if got != want {
		t.Errorf("Type() = %s, want %s", got, want)
	}
}

func TestCompanionHousingComponent_SerializeDeserialize(t *testing.T) {
	original := &CompanionHousingComponent{
		OwnerHouseID:      "house_123",
		BeddingID:         "bed_456",
		LastRestTime:      time.Date(2026, 2, 13, 10, 30, 0, 0, time.UTC),
		LoyaltyBonus:      0.15,
		ActiveTraining:    "training_789",
		TrainingBonus:     1.5,
		SharedChestAccess: []string{"chest_1", "chest_2"},
	}

	// Serialize
	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize() error = %v", err)
	}

	// Deserialize into new component
	restored := &CompanionHousingComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize() error = %v", err)
	}

	// Verify all fields match
	if restored.OwnerHouseID != original.OwnerHouseID {
		t.Errorf("OwnerHouseID = %s, want %s", restored.OwnerHouseID, original.OwnerHouseID)
	}
	if restored.BeddingID != original.BeddingID {
		t.Errorf("BeddingID = %s, want %s", restored.BeddingID, original.BeddingID)
	}
	if !restored.LastRestTime.Equal(original.LastRestTime) {
		t.Errorf("LastRestTime = %v, want %v", restored.LastRestTime, original.LastRestTime)
	}
	if restored.LoyaltyBonus != original.LoyaltyBonus {
		t.Errorf("LoyaltyBonus = %v, want %v", restored.LoyaltyBonus, original.LoyaltyBonus)
	}
	if restored.ActiveTraining != original.ActiveTraining {
		t.Errorf("ActiveTraining = %s, want %s", restored.ActiveTraining, original.ActiveTraining)
	}
	if restored.TrainingBonus != original.TrainingBonus {
		t.Errorf("TrainingBonus = %v, want %v", restored.TrainingBonus, original.TrainingBonus)
	}
	if len(restored.SharedChestAccess) != len(original.SharedChestAccess) {
		t.Errorf("SharedChestAccess len = %d, want %d", len(restored.SharedChestAccess), len(original.SharedChestAccess))
	}
	for i, v := range restored.SharedChestAccess {
		if v != original.SharedChestAccess[i] {
			t.Errorf("SharedChestAccess[%d] = %s, want %s", i, v, original.SharedChestAccess[i])
		}
	}
}

func TestCompanionHousingComponent_Deserialize_InvalidJSON(t *testing.T) {
	c := &CompanionHousingComponent{}
	err := c.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() should fail on invalid JSON")
	}
}
