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
