package fluids

import (
	"testing"
)

func TestNewBuoyancyCalculator(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)

	if calc == nil {
		t.Fatal("NewBuoyancyCalculator returned nil")
	}
	if calc.gravity != 9.81 {
		t.Errorf("Gravity = %v, want 9.81", calc.gravity)
	}
}

func TestCalculateBuoyancy_NoFluid(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	calc.CalculateBuoyancy(component, 0.0, FluidWater)

	if component.Submerged != 0.0 {
		t.Errorf("Submerged = %v, want 0.0 with no fluid", component.Submerged)
	}
	if component.BuoyantForce != 0.0 {
		t.Errorf("BuoyantForce = %v, want 0.0 with no fluid", component.BuoyantForce)
	}
	if component.Buoyant {
		t.Error("Entity should not be buoyant with no fluid")
	}
}

func TestCalculateBuoyancy_FullySubmerged(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   500.0, // 500 kg
		Volume: 2.0,   // 2 m³
	}

	calc.CalculateBuoyancy(component, 1.0, FluidWater)

	if component.Submerged != 1.0 {
		t.Errorf("Submerged = %v, want 1.0", component.Submerged)
	}

	// Water density = 1000 kg/m³
	// Buoyant force = 1000 * 2.0 * 9.81 = 19620 N
	expectedForce := 1000.0 * 2.0 * 9.81
	if component.BuoyantForce != expectedForce {
		t.Errorf("BuoyantForce = %v, want %v", component.BuoyantForce, expectedForce)
	}

	// Weight = 500 * 9.81 = 4905 N
	// Buoyant force > weight, so should float
	if !component.Buoyant {
		t.Error("Entity should be buoyant (wood floats in water)")
	}
}

func TestCalculateBuoyancy_PartiallySubmerged(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	calc.CalculateBuoyancy(component, 0.5, FluidWater)

	if component.Submerged != 0.5 {
		t.Errorf("Submerged = %v, want 0.5", component.Submerged)
	}

	// Buoyant force should be half of fully submerged
	expectedForce := 1000.0 * 2.0 * 9.81 * 0.5
	if component.BuoyantForce != expectedForce {
		t.Errorf("BuoyantForce = %v, want %v", component.BuoyantForce, expectedForce)
	}
}

func TestCalculateBuoyancy_SinksInWater(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   8000.0, // 8000 kg (steel density ~7850 kg/m³)
		Volume: 1.0,    // 1 m³
	}

	calc.CalculateBuoyancy(component, 1.0, FluidWater)

	// Weight = 8000 * 9.81 = 78480 N
	// Buoyant force = 1000 * 1.0 * 9.81 = 9810 N
	// Buoyant force < weight, so should sink
	if component.Buoyant {
		t.Error("Steel should sink in water")
	}
}

func TestCalculateBuoyancy_FloatsInLava(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   1000.0, // 1000 kg
		Volume: 1.0,    // 1 m³
	}

	calc.CalculateBuoyancy(component, 1.0, FluidLava)

	// Lava density = 3000 kg/m³
	// Buoyant force = 3000 * 1.0 * 9.81 = 29430 N
	// Weight = 1000 * 9.81 = 9810 N
	// Should float in lava
	if !component.Buoyant {
		t.Error("Entity should float in lava (lava is very dense)")
	}
}

func TestGetNetForce(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:         500.0,
		Volume:       2.0,
		BuoyantForce: 19620.0, // Pre-calculated
	}

	netForce := calc.GetNetForce(component)

	// Net force = Buoyant force - Weight
	// = 19620 - (500 * 9.81) = 19620 - 4905 = 14715 N (upward)
	expectedNet := 19620.0 - 500.0*9.81
	if netForce != expectedNet {
		t.Errorf("Net force = %v, want %v", netForce, expectedNet)
	}
}

func TestBuoyancyCalculator_UpdateDensity(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)

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
			calc.UpdateDensity(component)

			if component.Density != tt.expectedDensity {
				t.Errorf("Density = %v, want %v", component.Density, tt.expectedDensity)
			}
		})
	}
}

func TestBuoyancyCalculator_UpdateDensity_ZeroVolume(t *testing.T) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   100.0,
		Volume: 0.0,
	}
	calc.UpdateDensity(component)

	// Should not panic, density should remain 0
	if component.Density != 0.0 {
		t.Errorf("Density = %v, want 0.0 for zero volume", component.Density)
	}
}

func TestNewSwimmingManager(t *testing.T) {
	mgr := NewSwimmingManager(9.81)

	if mgr == nil {
		t.Fatal("NewSwimmingManager returned nil")
	}
	if mgr.calculator == nil {
		t.Error("Swimming manager calculator is nil")
	}
}

func TestUpdateSwimming_OnLand(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		Stamina:      50.0,
		MaxStamina:   100.0,
		StaminaRegen: 20.0,
	}
	buoyancy := &BuoyancyComponent{}

	mgr.UpdateSwimming(swimming, buoyancy, 0.0, 0.1)

	if swimming.IsSwimming {
		t.Error("Should not be swimming on land")
	}
	if swimming.Stamina <= 50.0 {
		t.Error("Stamina should regenerate on land")
	}
	if swimming.Drowning {
		t.Error("Should not be drowning on land")
	}
}

func TestUpdateSwimming_InWater(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		Stamina:      100.0,
		MaxStamina:   100.0,
		StaminaDrain: 10.0,
	}
	buoyancy := &BuoyancyComponent{}

	mgr.UpdateSwimming(swimming, buoyancy, 0.8, 0.1)

	if !swimming.IsSwimming {
		t.Error("Should be swimming in water")
	}
	if swimming.Stamina >= 100.0 {
		t.Error("Stamina should drain while swimming")
	}
}

func TestUpdateSwimming_TreadingWater(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		Stamina:       100.0,
		MaxStamina:    100.0,
		StaminaDrain:  10.0,
		TreadingWater: true,
	}
	buoyancy := &BuoyancyComponent{}

	mgr.UpdateSwimming(swimming, buoyancy, 0.8, 0.1)

	// Treading water drains half stamina
	// Expected drain = 10.0 * 0.5 * 0.1 = 0.5
	expectedStamina := 100.0 - 0.5
	if swimming.Stamina != expectedStamina {
		t.Errorf("Stamina = %v, want %v (treading water)", swimming.Stamina, expectedStamina)
	}
}

func TestUpdateSwimming_Drowning(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		Stamina:      0.0,
		MaxStamina:   100.0,
		StaminaDrain: 10.0,
	}
	buoyancy := &BuoyancyComponent{}

	mgr.UpdateSwimming(swimming, buoyancy, 0.8, 0.1)

	if !swimming.Drowning {
		t.Error("Should be drowning with zero stamina")
	}
}

func TestGetSwimSpeedMultiplier_NotSwimming(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		IsSwimming: false,
	}

	mult := mgr.GetSwimSpeedMultiplier(swimming)
	if mult != 1.0 {
		t.Errorf("Speed multiplier = %v, want 1.0 when not swimming", mult)
	}
}

func TestGetSwimSpeedMultiplier_FullStamina(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		IsSwimming: true,
		Stamina:    100.0,
		MaxStamina: 100.0,
		SwimSpeed:  0.5,
	}

	mult := mgr.GetSwimSpeedMultiplier(swimming)

	// With full stamina: 0.5 * (0.5 + 0.5 * 1.0) = 0.5 * 1.0 = 0.5
	expectedMult := 0.5
	if mult != expectedMult {
		t.Errorf("Speed multiplier = %v, want %v", mult, expectedMult)
	}
}

func TestGetSwimSpeedMultiplier_HalfStamina(t *testing.T) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		IsSwimming: true,
		Stamina:    50.0,
		MaxStamina: 100.0,
		SwimSpeed:  0.5,
	}

	mult := mgr.GetSwimSpeedMultiplier(swimming)

	// With half stamina: 0.5 * (0.5 + 0.5 * 0.5) = 0.5 * 0.75 = 0.375
	expectedMult := 0.375
	if mult != expectedMult {
		t.Errorf("Speed multiplier = %v, want %v", mult, expectedMult)
	}
}

func TestGetDrowningDamage(t *testing.T) {
	mgr := NewSwimmingManager(9.81)

	tests := []struct {
		name           string
		drowning       bool
		drowningDamage float64
		expectedDamage float64
	}{
		{"Not drowning", false, 10.0, 0.0},
		{"Drowning", true, 10.0, 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			swimming := &SwimmingComponent{
				Drowning:       tt.drowning,
				DrowningDamage: tt.drowningDamage,
			}

			damage := mgr.GetDrowningDamage(swimming)
			if damage != tt.expectedDamage {
				t.Errorf("Drowning damage = %v, want %v", damage, tt.expectedDamage)
			}
		})
	}
}

func TestNewFloodingManager(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	if mgr == nil {
		t.Fatal("NewFloodingManager returned nil")
	}
	if mgr.simulator == nil {
		t.Error("Flooding manager simulator is nil")
	}
}

func TestAddFloodSource(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	flooding := &FloodingComponent{}

	mgr.AddFloodSource(flooding, 10, 20, 0.5)

	if len(flooding.Sources) != 1 {
		t.Errorf("Sources count = %v, want 1", len(flooding.Sources))
	}

	source := flooding.Sources[0]
	if source.X != 10 || source.Y != 20 || source.FlowRate != 0.5 {
		t.Errorf("Source = {%d, %d, %v}, want {10, 20, 0.5}", source.X, source.Y, source.FlowRate)
	}
}

func TestRemoveFloodSource(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	flooding := &FloodingComponent{}
	mgr.AddFloodSource(flooding, 10, 20, 0.5)
	mgr.AddFloodSource(flooding, 15, 25, 0.3)

	mgr.RemoveFloodSource(flooding, 10, 20)

	if len(flooding.Sources) != 1 {
		t.Errorf("Sources count = %v, want 1 after removal", len(flooding.Sources))
	}

	if flooding.Sources[0].X != 15 || flooding.Sources[0].Y != 25 {
		t.Error("Wrong source was removed")
	}
}

func TestUpdateFlooding(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 50
	config.GridHeight = 50
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	flooding := &FloodingComponent{
		FloodLevel:    0.0,
		FloodRate:     0.1,
		MaxFloodLevel: 1.0,
	}
	mgr.AddFloodSource(flooding, 25, 25, 0.05)

	mgr.UpdateFlooding(flooding, 0.1)

	if flooding.FloodLevel <= 0.0 {
		t.Error("Flood level should increase")
	}
}

func TestUpdateFlooding_OutOfBounds(t *testing.T) {
	config := DefaultSimulationConfig()
	config.GridWidth = 10
	config.GridHeight = 10
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	flooding := &FloodingComponent{
		FloodLevel:    0.0,
		FloodRate:     0.1,
		MaxFloodLevel: 1.0,
	}
	// Add source at out-of-bounds position; should log warning, not panic
	mgr.AddFloodSource(flooding, 999, 999, 0.05)
	mgr.UpdateFlooding(flooding, 0.1)

	// Flood level should still update from FloodRate even though AddFluid fails
	if flooding.FloodLevel <= 0.0 {
		t.Error("Flood level should increase from FloodRate despite out-of-bounds source")
	}
}

func TestIsFullyFlooded(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	tests := []struct {
		name          string
		floodLevel    float64
		maxFloodLevel float64
		expected      bool
	}{
		{"Empty", 0.0, 1.0, false},
		{"Half full", 0.5, 1.0, false},
		{"Fully flooded", 1.0, 1.0, true},
		{"Over flooded", 1.5, 1.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flooding := &FloodingComponent{
				FloodLevel:    tt.floodLevel,
				MaxFloodLevel: tt.maxFloodLevel,
			}

			result := mgr.IsFullyFlooded(flooding)
			if result != tt.expected {
				t.Errorf("IsFullyFlooded = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFloodPercentage(t *testing.T) {
	config := DefaultSimulationConfig()
	sim := NewSimulator(config)
	mgr := NewFloodingManager(sim)

	tests := []struct {
		name          string
		floodLevel    float64
		maxFloodLevel float64
		expected      float64
	}{
		{"Empty", 0.0, 1.0, 0.0},
		{"Quarter", 0.25, 1.0, 0.25},
		{"Half", 0.5, 1.0, 0.5},
		{"Full", 1.0, 1.0, 1.0},
		{"Zero max", 0.5, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flooding := &FloodingComponent{
				FloodLevel:    tt.floodLevel,
				MaxFloodLevel: tt.maxFloodLevel,
			}

			result := mgr.GetFloodPercentage(flooding)
			if result != tt.expected {
				t.Errorf("GetFloodPercentage = %v, want %v", result, tt.expected)
			}
		})
	}
}

func BenchmarkCalculateBuoyancy(b *testing.B) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.CalculateBuoyancy(component, 1.0, FluidWater)
	}
}

func BenchmarkUpdateSwimming(b *testing.B) {
	mgr := NewSwimmingManager(9.81)
	swimming := &SwimmingComponent{
		Stamina:      100.0,
		MaxStamina:   100.0,
		StaminaDrain: 10.0,
		StaminaRegen: 20.0,
		SwimSpeed:    0.5,
	}
	buoyancy := &BuoyancyComponent{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mgr.UpdateSwimming(swimming, buoyancy, 0.8, 0.016)
	}
}

func BenchmarkGetNetForce(b *testing.B) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:         500.0,
		BuoyantForce: 19620.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = calc.GetNetForce(component)
	}
}

func BenchmarkBuoyancyCalculator_UpdateDensity(b *testing.B) {
	calc := NewBuoyancyCalculator(9.81)
	component := &BuoyancyComponent{
		Mass:   500.0,
		Volume: 2.0,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calc.UpdateDensity(component)
	}
}
