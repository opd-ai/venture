// Package vehicle provides vehicle generator tests.
// Tests verify generation, determinism, stat scaling, and validation.
//
// Phase 21.1: Vehicle Foundation
package vehicle

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen"
)

func TestNewVehicleGenerator(t *testing.T) {
	gen := NewVehicleGenerator()
	if gen == nil {
		t.Fatal("NewVehicleGenerator returned nil")
	}

	// Check templates loaded for all genres
	expectedGenres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc", ""}
	for _, genre := range expectedGenres {
		if len(gen.templates[genre]) == 0 {
			t.Errorf("No templates loaded for genre: %s", genre)
		}
	}
}

func TestVehicleGenerator_Generate(t *testing.T) {
	gen := NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      5,
		Difficulty: 0.5,
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	vehicles, ok := result.([]*Vehicle)
	if !ok {
		t.Fatal("Result is not []*Vehicle")
	}

	if len(vehicles) == 0 {
		t.Error("No vehicles generated")
	}

	// Check first vehicle
	vehicle := vehicles[0]
	if vehicle.Name == "" {
		t.Error("Vehicle has no name")
	}
	if vehicle.MaxSpeed <= 0 {
		t.Errorf("Invalid MaxSpeed: %f", vehicle.MaxSpeed)
	}
	if vehicle.GenreID != "fantasy" {
		t.Errorf("Wrong genre: %s", vehicle.GenreID)
	}
}

func TestVehicleGenerator_Determinism(t *testing.T) {
	gen := NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    "scifi",
		Depth:      3,
		Difficulty: 0.7,
	}

	seed := int64(54321)

	// Generate twice with same seed
	result1, _ := gen.Generate(seed, params)
	result2, _ := gen.Generate(seed, params)

	vehicles1 := result1.([]*Vehicle)
	vehicles2 := result2.([]*Vehicle)

	if len(vehicles1) != len(vehicles2) {
		t.Fatal("Different vehicle counts")
	}

	// Compare vehicles
	for i := range vehicles1 {
		v1 := vehicles1[i]
		v2 := vehicles2[i]

		if v1.Name != v2.Name {
			t.Errorf("Vehicle %d: names differ: %s vs %s", i, v1.Name, v2.Name)
		}
		if v1.VehicleType != v2.VehicleType {
			t.Errorf("Vehicle %d: types differ", i)
		}
		if v1.Rarity != v2.Rarity {
			t.Errorf("Vehicle %d: rarity differs", i)
		}
		if v1.MaxSpeed != v2.MaxSpeed {
			t.Errorf("Vehicle %d: MaxSpeed differs: %f vs %f", i, v1.MaxSpeed, v2.MaxSpeed)
		}
		if v1.Color != v2.Color {
			t.Errorf("Vehicle %d: colors differ: 0x%X vs 0x%X", i, v1.Color, v2.Color)
		}
	}
}

func TestVehicleGenerator_AllGenres(t *testing.T) {
	gen := NewVehicleGenerator()
	genres := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	for _, genre := range genres {
		t.Run(genre, func(t *testing.T) {
			params := procgen.GenerationParams{
				GenreID:    genre,
				Depth:      5,
				Difficulty: 0.5,
			}

			result, err := gen.Generate(12345, params)
			if err != nil {
				t.Fatalf("Generate failed for %s: %v", genre, err)
			}

			vehicles := result.([]*Vehicle)
			if len(vehicles) == 0 {
				t.Errorf("No vehicles generated for %s", genre)
			}

			// Validate all vehicles
			for i, vehicle := range vehicles {
				if vehicle.GenreID != genre {
					t.Errorf("Vehicle %d has wrong genre: %s", i, vehicle.GenreID)
				}
			}
		})
	}
}

func TestVehicleGenerator_StatScaling(t *testing.T) {
	gen := NewVehicleGenerator()
	baseParams := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      1,
		Difficulty: 0.5,
	}

	// Generate at depth 1
	result1, _ := gen.Generate(99999, baseParams)
	vehicles1 := result1.([]*Vehicle)

	// Generate at depth 10 (should have higher stats)
	highDepthParams := baseParams
	highDepthParams.Depth = 10
	result2, _ := gen.Generate(99999, highDepthParams)
	vehicles2 := result2.([]*Vehicle)

	// Compare average stats (should be higher at depth 10)
	avgSpeed1 := calculateAvgSpeed(vehicles1)
	avgSpeed2 := calculateAvgSpeed(vehicles2)

	if avgSpeed2 <= avgSpeed1 {
		t.Errorf("Depth 10 vehicles should have higher stats than depth 1: %f vs %f", avgSpeed2, avgSpeed1)
	}
}

func TestVehicleGenerator_CustomCount(t *testing.T) {
	gen := NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      5,
		Difficulty: 0.5,
		Custom: map[string]interface{}{
			"count": 10,
		},
	}

	result, err := gen.Generate(12345, params)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	vehicles := result.([]*Vehicle)
	if len(vehicles) != 10 {
		t.Errorf("Expected 10 vehicles, got %d", len(vehicles))
	}
}

func TestVehicleGenerator_Validate(t *testing.T) {
	gen := NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      5,
		Difficulty: 0.5,
	}

	result, _ := gen.Generate(12345, params)

	// Should pass validation
	err := gen.Validate(result)
	if err != nil {
		t.Errorf("Validation failed: %v", err)
	}

	// Test with invalid data
	err = gen.Validate("not a vehicle slice")
	if err == nil {
		t.Error("Should fail validation for wrong type")
	}

	err = gen.Validate([]*Vehicle{})
	if err == nil {
		t.Error("Should fail validation for empty slice")
	}

	// Test with nil vehicle
	invalidVehicles := []*Vehicle{nil}
	err = gen.Validate(invalidVehicles)
	if err == nil {
		t.Error("Should fail validation for nil vehicle")
	}
}

func TestVehicle_ToComponent(t *testing.T) {
	vehicle := &Vehicle{
		Name:          "Test Mount",
		VehicleType:   TypeMount,
		Rarity:        RarityRare,
		MaxSpeed:      200.0,
		Acceleration:  100.0,
		Handling:      3.0,
		MaxDurability: 150.0,
		FuelCapacity:  100.0,
		Capacity:      1,
		FuelType:      "Stamina",
	}

	comp := vehicle.ToComponent()

	if comp == nil {
		t.Fatal("ToComponent returned nil")
	}

	if comp.MaxSpeed != vehicle.MaxSpeed {
		t.Errorf("MaxSpeed mismatch: %f vs %f", comp.MaxSpeed, vehicle.MaxSpeed)
	}

	if comp.Acceleration != vehicle.Acceleration {
		t.Errorf("Acceleration mismatch: %f vs %f", comp.Acceleration, vehicle.Acceleration)
	}

	if comp.Handling != vehicle.Handling {
		t.Errorf("Handling mismatch: %f vs %f", comp.Handling, vehicle.Handling)
	}

	if comp.MaxDurability != vehicle.MaxDurability {
		t.Errorf("MaxDurability mismatch: %f vs %f", comp.MaxDurability, vehicle.MaxDurability)
	}

	if comp.Durability != vehicle.MaxDurability {
		t.Errorf("Durability should start at max: %f vs %f", comp.Durability, vehicle.MaxDurability)
	}

	if comp.FuelAmount != vehicle.FuelCapacity {
		t.Errorf("Fuel should start at capacity: %f vs %f", comp.FuelAmount, vehicle.FuelCapacity)
	}

	if comp.Capacity != vehicle.Capacity {
		t.Errorf("Capacity mismatch: %d vs %d", comp.Capacity, vehicle.Capacity)
	}
}

func TestRarity_GetMultiplier(t *testing.T) {
	tests := []struct {
		rarity     Rarity
		multiplier float64
	}{
		{RarityCommon, 1.0},
		{RarityUncommon, 1.2},
		{RarityRare, 1.5},
		{RarityEpic, 2.0},
		{RarityLegendary, 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.rarity.String(), func(t *testing.T) {
			if m := tt.rarity.GetMultiplier(); m != tt.multiplier {
				t.Errorf("GetMultiplier() = %f, want %f", m, tt.multiplier)
			}
		})
	}
}

func TestVehicleType_String(t *testing.T) {
	tests := []struct {
		vehicleType VehicleType
		expected    string
	}{
		{TypeMount, "Mount"},
		{TypeCart, "Cart"},
		{TypeBoat, "Boat"},
		{TypeGlider, "Glider"},
		{TypeMech, "Mech"},
		{VehicleType(99), "Unknown"},
	}

	for _, tt := range tests {
		if s := tt.vehicleType.String(); s != tt.expected {
			t.Errorf("String() = %s, want %s", s, tt.expected)
		}
	}
}

func TestRarity_String(t *testing.T) {
	tests := []struct {
		rarity   Rarity
		expected string
	}{
		{RarityCommon, "Common"},
		{RarityUncommon, "Uncommon"},
		{RarityRare, "Rare"},
		{RarityEpic, "Epic"},
		{RarityLegendary, "Legendary"},
		{Rarity(99), "Unknown"},
	}

	for _, tt := range tests {
		if s := tt.rarity.String(); s != tt.expected {
			t.Errorf("String() = %s, want %s", s, tt.expected)
		}
	}
}

// Helper function
func calculateAvgSpeed(vehicles []*Vehicle) float64 {
	if len(vehicles) == 0 {
		return 0
	}
	total := 0.0
	for _, v := range vehicles {
		total += v.MaxSpeed
	}
	return total / float64(len(vehicles))
}

// Benchmarks
func BenchmarkVehicleGenerator_Generate(b *testing.B) {
	gen := NewVehicleGenerator()
	params := procgen.GenerationParams{
		GenreID:    "fantasy",
		Depth:      5,
		Difficulty: 0.5,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate(int64(i), params)
	}
}

func BenchmarkVehicle_ToComponent(b *testing.B) {
	vehicle := &Vehicle{
		VehicleType:   TypeMount,
		MaxSpeed:      200.0,
		Acceleration:  100.0,
		Handling:      3.0,
		MaxDurability: 150.0,
		FuelCapacity:  100.0,
		Capacity:      1,
		FuelType:      "Stamina",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = vehicle.ToComponent()
	}
}
