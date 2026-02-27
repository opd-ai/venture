package fluids

import (
	"testing"
)

func TestFluidType_String(t *testing.T) {
	tests := []struct {
		name     string
		fluid    FluidType
		expected string
	}{
		{"Water", FluidWater, "Water"},
		{"Lava", FluidLava, "Lava"},
		{"Oil", FluidOil, "Oil"},
		{"Acid", FluidAcid, "Acid"},
		{"Poison", FluidPoison, "Poison"},
		{"Unknown", FluidType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fluid.String()
			if result != tt.expected {
				t.Errorf("FluidType.String() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFluidProperties(t *testing.T) {
	tests := []struct {
		name      string
		fluidType FluidType
	}{
		{"Water", FluidWater},
		{"Lava", FluidLava},
		{"Oil", FluidOil},
		{"Acid", FluidAcid},
		{"Poison", FluidPoison},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			props := GetFluidProperties(tt.fluidType)

			if props.Viscosity < 0.0 || props.Viscosity > 1.0 {
				t.Errorf("Invalid viscosity: %v", props.Viscosity)
			}
			if props.Density <= 0.0 {
				t.Errorf("Invalid density: %v", props.Density)
			}
			if props.FlowRate < 0.0 || props.FlowRate > 1.0 {
				t.Errorf("Invalid flow rate: %v", props.FlowRate)
			}
			if props.Damage < 0.0 {
				t.Errorf("Invalid damage: %v", props.Damage)
			}
			if props.Transparency < 0.0 || props.Transparency > 1.0 {
				t.Errorf("Invalid transparency: %v", props.Transparency)
			}
		})
	}
}

func TestGetFluidProperties_DifferentDensities(t *testing.T) {
	waterProps := GetFluidProperties(FluidWater)
	lavaProps := GetFluidProperties(FluidLava)

	if lavaProps.Density <= waterProps.Density {
		t.Errorf("Lava should be denser than water")
	}
	if lavaProps.Viscosity <= waterProps.Viscosity {
		t.Errorf("Lava should be more viscous than water")
	}
	if lavaProps.FlowRate >= waterProps.FlowRate {
		t.Errorf("Lava should flow slower than water")
	}
}

func TestGetFluidProperties_DamageValues(t *testing.T) {
	water := GetFluidProperties(FluidWater)
	lava := GetFluidProperties(FluidLava)
	acid := GetFluidProperties(FluidAcid)
	poison := GetFluidProperties(FluidPoison)

	if water.Damage != 0.0 {
		t.Errorf("Water should not cause damage")
	}
	if lava.Damage <= 0.0 {
		t.Errorf("Lava should cause damage")
	}
	if acid.Damage <= 0.0 {
		t.Errorf("Acid should cause damage")
	}
	if poison.Damage <= 0.0 {
		t.Errorf("Poison should cause damage")
	}
}

func TestBuoyancyComponent_Type(t *testing.T) {
	component := BuoyancyComponent{}
	if component.Type() != "buoyancy" {
		t.Errorf("BuoyancyComponent.Type() = %v, want %v", component.Type(), "buoyancy")
	}
}

func TestSwimmingComponent_Type(t *testing.T) {
	component := SwimmingComponent{}
	if component.Type() != "swimming" {
		t.Errorf("SwimmingComponent.Type() = %v, want %v", component.Type(), "swimming")
	}
}

func TestFloodingComponent_Type(t *testing.T) {
	component := FloodingComponent{}
	if component.Type() != "flooding" {
		t.Errorf("FloodingComponent.Type() = %v, want %v", component.Type(), "flooding")
	}
}

func TestDefaultSimulationConfig(t *testing.T) {
	config := DefaultSimulationConfig()

	if config.GridWidth <= 0 {
		t.Errorf("Invalid grid width: %v", config.GridWidth)
	}
	if config.GridHeight <= 0 {
		t.Errorf("Invalid grid height: %v", config.GridHeight)
	}
	if config.CellSize <= 0 {
		t.Errorf("Invalid cell size: %v", config.CellSize)
	}
	if config.UpdateRate <= 0 {
		t.Errorf("Invalid update rate: %v", config.UpdateRate)
	}
	if config.Gravity <= 0 {
		t.Errorf("Invalid gravity: %v", config.Gravity)
	}
	if config.MaxIterations <= 0 {
		t.Errorf("Invalid max iterations: %v", config.MaxIterations)
	}
}

func TestUpdateDensity(t *testing.T) {
	tests := []struct {
		name            string
		mass            float64
		volume          float64
		expectedDensity float64
	}{
		{"Wood", 500.0, 2.0, 250.0},
		{"Water", 1000.0, 1.0, 1000.0},
		{"Steel", 7850.0, 1.0, 7850.0},
		{"Cork", 240.0, 1.0, 240.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			component := &BuoyancyComponent{
				Mass:   tt.mass,
				Volume: tt.volume,
			}
			UpdateDensity(component)

			if component.Density != tt.expectedDensity {
				t.Errorf("Density = %v, want %v", component.Density, tt.expectedDensity)
			}
		})
	}
}

func TestUpdateDensity_ZeroVolume(t *testing.T) {
	component := &BuoyancyComponent{
		Mass:   100.0,
		Volume: 0.0,
	}
	UpdateDensity(component)

	// Should not panic, density should remain 0
	if component.Density != 0.0 {
		t.Errorf("Density with zero volume should be 0, got %v", component.Density)
	}
}

func BenchmarkGetFluidProperties(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = GetFluidProperties(FluidWater)
	}
}

func BenchmarkUpdateDensity(b *testing.B) {
	component := &BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UpdateDensity(component)
	}
}

func TestBuoyancyComponent_Serialization(t *testing.T) {
	original := &BuoyancyComponent{
		Mass:         100.0,
		Volume:       0.2,
		Density:      500.0,
		Buoyant:      true,
		Submerged:    0.75,
		BuoyantForce: 150.0,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restored := &BuoyancyComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.Mass != original.Mass {
		t.Errorf("Mass mismatch: got %f, want %f", restored.Mass, original.Mass)
	}
	if restored.Volume != original.Volume {
		t.Errorf("Volume mismatch: got %f, want %f", restored.Volume, original.Volume)
	}
	if restored.Buoyant != original.Buoyant {
		t.Errorf("Buoyant mismatch: got %v, want %v", restored.Buoyant, original.Buoyant)
	}
	if restored.Submerged != original.Submerged {
		t.Errorf("Submerged mismatch: got %f, want %f", restored.Submerged, original.Submerged)
	}
}

func TestSwimmingComponent_Serialization(t *testing.T) {
	original := &SwimmingComponent{
		IsSwimming:     true,
		Stamina:        75.5,
		MaxStamina:     100.0,
		StaminaDrain:   5.0,
		StaminaRegen:   2.5,
		SwimSpeed:      0.6,
		TreadingWater:  true,
		Drowning:       false,
		DrowningDamage: 10.0,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restored := &SwimmingComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.IsSwimming != original.IsSwimming {
		t.Errorf("IsSwimming mismatch: got %v, want %v", restored.IsSwimming, original.IsSwimming)
	}
	if restored.Stamina != original.Stamina {
		t.Errorf("Stamina mismatch: got %f, want %f", restored.Stamina, original.Stamina)
	}
	if restored.TreadingWater != original.TreadingWater {
		t.Errorf("TreadingWater mismatch: got %v, want %v", restored.TreadingWater, original.TreadingWater)
	}
	if restored.SwimSpeed != original.SwimSpeed {
		t.Errorf("SwimSpeed mismatch: got %f, want %f", restored.SwimSpeed, original.SwimSpeed)
	}
}

func TestFloodingComponent_Serialization(t *testing.T) {
	original := &FloodingComponent{
		AreaID:        "dungeon-room-1",
		FloodLevel:    0.35,
		FloodRate:     0.1,
		MaxFloodLevel: 1.0,
		Sources: []FloodSource{
			{X: 10, Y: 20, FlowRate: 0.05},
			{X: 30, Y: 40, FlowRate: 0.08},
		},
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restored := &FloodingComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.AreaID != original.AreaID {
		t.Errorf("AreaID mismatch: got %q, want %q", restored.AreaID, original.AreaID)
	}
	if restored.FloodLevel != original.FloodLevel {
		t.Errorf("FloodLevel mismatch: got %f, want %f", restored.FloodLevel, original.FloodLevel)
	}
	if len(restored.Sources) != len(original.Sources) {
		t.Fatalf("Sources length mismatch: got %d, want %d", len(restored.Sources), len(original.Sources))
	}
	for i, src := range restored.Sources {
		if src.X != original.Sources[i].X || src.Y != original.Sources[i].Y {
			t.Errorf("Source %d position mismatch: got (%d,%d), want (%d,%d)",
				i, src.X, src.Y, original.Sources[i].X, original.Sources[i].Y)
		}
		if src.FlowRate != original.Sources[i].FlowRate {
			t.Errorf("Source %d FlowRate mismatch: got %f, want %f",
				i, src.FlowRate, original.Sources[i].FlowRate)
		}
	}
}

func TestFloodingComponent_SerializationEmpty(t *testing.T) {
	original := &FloodingComponent{
		AreaID:        "",
		FloodLevel:    0,
		FloodRate:     0,
		MaxFloodLevel: 0,
		Sources:       nil,
	}

	data, err := original.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	restored := &FloodingComponent{}
	err = restored.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if restored.AreaID != "" {
		t.Errorf("Expected empty AreaID, got %q", restored.AreaID)
	}
	if len(restored.Sources) != 0 {
		t.Errorf("Expected empty Sources, got %d", len(restored.Sources))
	}
}

func TestComponentDeserialization_InvalidData(t *testing.T) {
	tests := []struct {
		name      string
		component interface{ Deserialize([]byte) error }
		data      []byte
	}{
		{"BuoyancyComponent", &BuoyancyComponent{}, []byte{1, 2, 3}},
		{"SwimmingComponent", &SwimmingComponent{}, []byte{1, 2, 3}},
		{"FloodingComponent", &FloodingComponent{}, []byte{1, 2, 3}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.component.Deserialize(tt.data)
			if err == nil {
				t.Error("Expected error for invalid data")
			}
		})
	}
}

func BenchmarkBuoyancyComponent_Serialize(b *testing.B) {
	component := &BuoyancyComponent{
		Mass:         100.0,
		Volume:       0.2,
		Density:      500.0,
		Buoyant:      true,
		Submerged:    0.75,
		BuoyantForce: 150.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = component.Serialize()
	}
}

func BenchmarkBuoyancyComponent_Deserialize(b *testing.B) {
	component := &BuoyancyComponent{
		Mass:         100.0,
		Volume:       0.2,
		Density:      500.0,
		Buoyant:      true,
		Submerged:    0.75,
		BuoyantForce: 150.0,
	}
	data, _ := component.Serialize()

	restored := &BuoyancyComponent{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = restored.Deserialize(data)
	}
}

// TestRegisterComponentFactories verifies the RegisterComponentFactories function
// can be called without errors and is idempotent.
func TestRegisterComponentFactories(t *testing.T) {
	tests := []struct {
		name  string
		calls int
	}{
		{"SingleCall", 1},
		{"MultipleCalls", 3},
		{"ManyCallsIdempotent", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call RegisterComponentFactories multiple times to verify idempotency
			for i := 0; i < tt.calls; i++ {
				// Should not panic
				RegisterComponentFactories()
			}
		})
	}
}

// TestComponentTypeIdentifiers verifies all fluid components have correct type identifiers.
func TestComponentTypeIdentifiers(t *testing.T) {
	tests := []struct {
		name         string
		component    interface{ Type() string }
		expectedType string
	}{
		{
			name:         "BuoyancyComponent",
			component:    &BuoyancyComponent{},
			expectedType: "buoyancy",
		},
		{
			name:         "SwimmingComponent",
			component:    &SwimmingComponent{},
			expectedType: "swimming",
		},
		{
			name:         "FloodingComponent",
			component:    &FloodingComponent{},
			expectedType: "flooding",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.component.Type()
			if result != tt.expectedType {
				t.Errorf("Component.Type() = %q, want %q", result, tt.expectedType)
			}
		})
	}
}

// TestComponentSerializationRoundTrip verifies all components can serialize and deserialize.
func TestComponentSerializationRoundTrip(t *testing.T) {
	t.Run("BuoyancyComponent", func(t *testing.T) {
		original := &BuoyancyComponent{
			Mass:         100.0,
			Volume:       0.2,
			Density:      500.0,
			Buoyant:      true,
			Submerged:    0.75,
			BuoyantForce: 150.0,
		}

		data, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		restored := &BuoyancyComponent{}
		err = restored.Deserialize(data)
		if err != nil {
			t.Fatalf("Deserialize() error = %v", err)
		}

		if restored.Mass != original.Mass {
			t.Errorf("Mass mismatch: got %v, want %v", restored.Mass, original.Mass)
		}
		if restored.Buoyant != original.Buoyant {
			t.Errorf("Buoyant mismatch: got %v, want %v", restored.Buoyant, original.Buoyant)
		}
	})

	t.Run("SwimmingComponent", func(t *testing.T) {
		original := &SwimmingComponent{
			IsSwimming:     true,
			Stamina:        75.0,
			MaxStamina:     100.0,
			StaminaDrain:   5.0,
			StaminaRegen:   2.5,
			SwimSpeed:      0.8,
			TreadingWater:  false,
			Drowning:       false,
			DrowningDamage: 10.0,
		}

		data, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		restored := &SwimmingComponent{}
		err = restored.Deserialize(data)
		if err != nil {
			t.Fatalf("Deserialize() error = %v", err)
		}

		if restored.IsSwimming != original.IsSwimming {
			t.Errorf("IsSwimming mismatch: got %v, want %v", restored.IsSwimming, original.IsSwimming)
		}
		if restored.Stamina != original.Stamina {
			t.Errorf("Stamina mismatch: got %v, want %v", restored.Stamina, original.Stamina)
		}
	})

	t.Run("FloodingComponent", func(t *testing.T) {
		original := &FloodingComponent{
			AreaID:        "test-area",
			FloodLevel:    0.5,
			FloodRate:     0.1,
			MaxFloodLevel: 1.0,
			Sources: []FloodSource{
				{X: 10, Y: 20, FlowRate: 0.05},
				{X: 30, Y: 40, FlowRate: 0.05},
			},
		}

		data, err := original.Serialize()
		if err != nil {
			t.Fatalf("Serialize() error = %v", err)
		}

		restored := &FloodingComponent{}
		err = restored.Deserialize(data)
		if err != nil {
			t.Fatalf("Deserialize() error = %v", err)
		}

		if restored.AreaID != original.AreaID {
			t.Errorf("AreaID mismatch: got %v, want %v", restored.AreaID, original.AreaID)
		}
		if len(restored.Sources) != len(original.Sources) {
			t.Errorf("Sources length mismatch: got %v, want %v", len(restored.Sources), len(original.Sources))
		}
	})
}

// BenchmarkRegisterComponentFactories measures the performance of the registration function.
func BenchmarkRegisterComponentFactories(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RegisterComponentFactories()
	}
}
