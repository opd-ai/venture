package qol

import (
	"testing"
	"time"
)

func TestSortCriteriaString(t *testing.T) {
	tests := []struct {
		criteria SortCriteria
		want     string
	}{
		{SortByType, "Type"},
		{SortByRarity, "Rarity"},
		{SortByName, "Name"},
		{SortByValue, "Value"},
		{SortByQuantity, "Quantity"},
		{SortCriteria(99), "Unknown"},
	}

	for _, tt := range tests {
		if got := tt.criteria.String(); got != tt.want {
			t.Errorf("SortCriteria(%d).String() = %s, want %s", tt.criteria, got, tt.want)
		}
	}
}

func TestDefaultAutoLootConfig(t *testing.T) {
	config := DefaultAutoLootConfig(123)

	if config.CompanionID != 123 {
		t.Errorf("CompanionID = %d, want 123", config.CompanionID)
	}
	if !config.Enabled {
		t.Error("Expected Enabled = true")
	}
	if config.Radius != 7.0 {
		t.Errorf("Radius = %f, want 7.0", config.Radius)
	}
	if config.MinRarity != 0 {
		t.Errorf("MinRarity = %d, want 0", config.MinRarity)
	}
	if config.MaxPerCycle != 10 {
		t.Errorf("MaxPerCycle = %d, want 10", config.MaxPerCycle)
	}
}

func TestGuildInvitationExpiry(t *testing.T) {
	tests := []struct {
		name        string
		expiresAt   time.Time
		wantExpired bool
	}{
		{
			name:        "future expiry",
			expiresAt:   time.Now().Add(24 * time.Hour),
			wantExpired: false,
		},
		{
			name:        "past expiry",
			expiresAt:   time.Now().Add(-24 * time.Hour),
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := &GuildInvitation{
				ExpiresAt: tt.expiresAt,
			}
			if got := inv.IsExpired(); got != tt.wantExpired {
				t.Errorf("IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

func TestGuildInvitationDaysUntilExpiry(t *testing.T) {
	inv := &GuildInvitation{
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}

	days := inv.DaysUntilExpiry()
	if days < 1.9 || days > 2.1 {
		t.Errorf("DaysUntilExpiry() = %f, want ~2.0", days)
	}

	invExpired := &GuildInvitation{
		ExpiresAt: time.Now().Add(-24 * time.Hour),
	}
	if got := invExpired.DaysUntilExpiry(); got != 0 {
		t.Errorf("Expired DaysUntilExpiry() = %f, want 0", got)
	}
}

func TestEstimateArrivalTime(t *testing.T) {
	tests := []struct {
		distance float64
		want     float64
	}{
		{0.0, 0.0},
		{3.0, 3.0},
		{5.0, 5.0},
		{10.0, 5.0}, // Clamped to max 5 seconds
		{-1.0, 0.0}, // Negative distance
	}

	for _, tt := range tests {
		if got := EstimateArrivalTime(tt.distance); got != tt.want {
			t.Errorf("EstimateArrivalTime(%f) = %f, want %f", tt.distance, got, tt.want)
		}
	}
}

func TestQoLComponentType(t *testing.T) {
	comp := QoLComponent{}
	if got := comp.Type(); got != "qol" {
		t.Errorf("Type() = %s, want qol", got)
	}
}

func TestQoLComponentSerializeDeserialize(t *testing.T) {
	tests := []struct {
		name string
		comp QoLComponent
	}{
		{
			name: "empty component",
			comp: QoLComponent{},
		},
		{
			name: "full component",
			comp: QoLComponent{
				PlayerID:        12345,
				AutoLootEnabled: true,
				AutoLootRadius:  7.5,
				CraftQueue:      []*CraftQueueEntry{{RecipeID: "iron_sword", Quantity: 3, Position: 0}},
				SortPreset:      "rarity",
				MountWhistle:    true,
				RecipeTracking:  true,
			},
		},
		{
			name: "with craft queue",
			comp: QoLComponent{
				PlayerID: 999,
				CraftQueue: []*CraftQueueEntry{
					{RecipeID: "recipe1", Quantity: 1, Position: 0},
					{RecipeID: "recipe2", Quantity: 5, Position: 1},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			data, err := tt.comp.Serialize()
			if err != nil {
				t.Fatalf("Serialize() error = %v", err)
			}

			// Deserialize
			var restored QoLComponent
			if err := restored.Deserialize(data); err != nil {
				t.Fatalf("Deserialize() error = %v", err)
			}

			// Verify all fields
			if restored.PlayerID != tt.comp.PlayerID {
				t.Errorf("PlayerID = %d, want %d", restored.PlayerID, tt.comp.PlayerID)
			}
			if restored.AutoLootEnabled != tt.comp.AutoLootEnabled {
				t.Errorf("AutoLootEnabled = %v, want %v", restored.AutoLootEnabled, tt.comp.AutoLootEnabled)
			}
			if restored.AutoLootRadius != tt.comp.AutoLootRadius {
				t.Errorf("AutoLootRadius = %f, want %f", restored.AutoLootRadius, tt.comp.AutoLootRadius)
			}
			if restored.SortPreset != tt.comp.SortPreset {
				t.Errorf("SortPreset = %s, want %s", restored.SortPreset, tt.comp.SortPreset)
			}
			if restored.MountWhistle != tt.comp.MountWhistle {
				t.Errorf("MountWhistle = %v, want %v", restored.MountWhistle, tt.comp.MountWhistle)
			}
			if restored.RecipeTracking != tt.comp.RecipeTracking {
				t.Errorf("RecipeTracking = %v, want %v", restored.RecipeTracking, tt.comp.RecipeTracking)
			}
			if len(restored.CraftQueue) != len(tt.comp.CraftQueue) {
				t.Errorf("CraftQueue length = %d, want %d", len(restored.CraftQueue), len(tt.comp.CraftQueue))
			}
		})
	}
}

func TestQoLComponentDeserializeError(t *testing.T) {
	var comp QoLComponent
	err := comp.Deserialize([]byte("invalid json"))
	if err == nil {
		t.Error("Deserialize() expected error for invalid JSON")
	}
}

func BenchmarkEstimateArrivalTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EstimateArrivalTime(7.5)
	}
}

func BenchmarkDefaultAutoLootConfig(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DefaultAutoLootConfig(uint64(i))
	}
}
