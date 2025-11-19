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
