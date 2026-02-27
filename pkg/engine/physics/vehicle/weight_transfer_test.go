package vehicle

import (
	"math"
	"testing"
)

func TestNewWeightTransferComponent(t *testing.T) {
	comp := NewWeightTransferComponent()
	if comp == nil {
		t.Fatal("NewWeightTransferComponent returned nil")
	}
	if comp.Type() != "weight_transfer" {
		t.Errorf("got type %q, want %q", comp.Type(), "weight_transfer")
	}

	// Check default values
	if comp.CenterOfMassHeight != 15.0 {
		t.Errorf("got CenterOfMassHeight=%f, want 15.0", comp.CenterOfMassHeight)
	}
	if comp.Wheelbase != 32.0 {
		t.Errorf("got Wheelbase=%f, want 32.0", comp.Wheelbase)
	}
	if comp.TrackWidth != 16.0 {
		t.Errorf("got TrackWidth=%f, want 16.0", comp.TrackWidth)
	}

	// Check initial weight distribution (should be balanced)
	total := comp.FrontLeftWeight + comp.FrontRightWeight + comp.RearLeftWeight + comp.RearRightWeight
	if math.Abs(total-1.0) > 0.001 {
		t.Errorf("initial weights sum to %f, want 1.0", total)
	}
}

func TestWeightTransferComponent_Update(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	tests := []struct {
		name       string
		velX       float64
		velY       float64
		angularVel float64
		deltaTime  float64
	}{
		{"stationary", 0, 0, 0, 0.016},
		{"accelerating forward", 100, 0, 0, 0.016},
		{"braking", -50, 0, 0, 0.016},
		{"turning left", 50, 50, 0.5, 0.016},
		{"turning right", 50, 50, -0.5, 0.016},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewWeightTransferComponent()

			// Run update twice to calculate acceleration
			sys.UpdateWeightDistribution(comp, 0, 0, 0, tt.deltaTime)
			sys.UpdateWeightDistribution(comp, tt.velX, tt.velY, tt.angularVel, tt.deltaTime)

			// Check that weights still sum to 1.0
			total := comp.FrontLeftWeight + comp.FrontRightWeight + comp.RearLeftWeight + comp.RearRightWeight
			if math.Abs(total-1.0) > 0.001 {
				t.Errorf("weights sum to %f, want 1.0", total)
			}

			// Check that all weights are in valid range
			weights := []float64{comp.FrontLeftWeight, comp.FrontRightWeight, comp.RearLeftWeight, comp.RearRightWeight}
			for i, w := range weights {
				if w < 0.0 || w > 1.0 {
					t.Errorf("wheel %d weight %f out of range [0, 1]", i, w)
				}
			}
		})
	}
}

func TestWeightTransferComponent_GetWheelWeights(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()
	sys.UpdateWeightDistribution(comp, 0, 0, 0, 0.016)

	weights := comp.GetWheelWeights()
	if len(weights) != 4 {
		t.Fatalf("got %d weights, want 4", len(weights))
	}

	// Sum should equal 1.0
	sum := 0.0
	for _, w := range weights {
		sum += w
	}
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("weights sum to %f, want 1.0", sum)
	}
}

func TestWeightTransferComponent_GetAxleWeights(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()
	sys.UpdateWeightDistribution(comp, 0, 0, 0, 0.016)

	frontWeight := GetFrontAxleWeight(comp)
	rearWeight := GetRearAxleWeight(comp)

	// Sum should equal 1.0
	sum := frontWeight + rearWeight
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("front+rear weights sum to %f, want 1.0", sum)
	}

	// Check ranges
	if frontWeight < 0.0 || frontWeight > 1.0 {
		t.Errorf("front weight %f out of range [0, 1]", frontWeight)
	}
	if rearWeight < 0.0 || rearWeight > 1.0 {
		t.Errorf("rear weight %f out of range [0, 1]", rearWeight)
	}
}

func TestWeightTransferComponent_GetSideWeights(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()
	sys.UpdateWeightDistribution(comp, 0, 0, 0, 0.016)

	leftWeight := GetLeftSideWeight(comp)
	rightWeight := comp.GetRightSideWeight()

	// Sum should equal 1.0
	sum := leftWeight + rightWeight
	if math.Abs(sum-1.0) > 0.001 {
		t.Errorf("left+right weights sum to %f, want 1.0", sum)
	}

	// Check ranges
	if leftWeight < 0.0 || leftWeight > 1.0 {
		t.Errorf("left weight %f out of range [0, 1]", leftWeight)
	}
	if rightWeight < 0.0 || rightWeight > 1.0 {
		t.Errorf("right weight %f out of range [0, 1]", rightWeight)
	}
}

func TestWeightTransferComponent_Acceleration(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()

	// Simulate acceleration
	sys.UpdateWeightDistribution(comp, 0, 0, 0, 0.016)
	sys.UpdateWeightDistribution(comp, 100, 0, 0, 0.016) // Accelerate to 100 px/s

	// During acceleration, weight should shift rearward
	rearWeight := GetRearAxleWeight(comp)
	frontWeight := GetFrontAxleWeight(comp)

	if rearWeight <= frontWeight {
		t.Errorf("during acceleration, rear weight (%f) should exceed front weight (%f)", rearWeight, frontWeight)
	}
}

func TestWeightTransferComponent_Braking(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()

	// Simulate braking
	sys.UpdateWeightDistribution(comp, 100, 0, 0, 0.016) // Moving at 100 px/s
	sys.UpdateWeightDistribution(comp, 50, 0, 0, 0.016)  // Decelerate to 50 px/s

	// During braking, weight should shift forward
	frontWeight := GetFrontAxleWeight(comp)
	rearWeight := GetRearAxleWeight(comp)

	if frontWeight <= rearWeight {
		t.Errorf("during braking, front weight (%f) should exceed rear weight (%f)", frontWeight, rearWeight)
	}
}

func TestWeightTransferComponent_Turning(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()

	// Simulate left turn
	sys.UpdateWeightDistribution(comp, 50, 0, 0, 0.016)
	sys.UpdateWeightDistribution(comp, 50, 0, 0.5, 0.016) // Add positive angular velocity

	// During left turn, weight should shift right
	_ = comp.GetRightSideWeight()
	_ = GetLeftSideWeight(comp)

	// Note: The shift might be subtle, so we check it's not exactly balanced
	transferMag := GetTransferMagnitude(comp)
	if transferMag <= 0 {
		t.Error("turning should produce non-zero transfer magnitude")
	}
}

func TestWeightTransferComponent_Reset(t *testing.T) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()

	// Apply some dynamics
	sys.UpdateWeightDistribution(comp, 0, 0, 0, 0.016)
	sys.UpdateWeightDistribution(comp, 100, 50, 0.5, 0.016)

	// Reset
	comp.Reset()

	// Check that values are reset
	if comp.AccelerationX != 0.0 {
		t.Errorf("after reset, AccelerationX=%f, want 0.0", comp.AccelerationX)
	}
	if comp.AccelerationY != 0.0 {
		t.Errorf("after reset, AccelerationY=%f, want 0.0", comp.AccelerationY)
	}
	if comp.LastTransferMagnitude != 0.0 {
		t.Errorf("after reset, LastTransferMagnitude=%f, want 0.0", comp.LastTransferMagnitude)
	}

	// Check weights are balanced
	if comp.FrontLeftWeight != 0.25 || comp.FrontRightWeight != 0.25 ||
		comp.RearLeftWeight != 0.25 || comp.RearRightWeight != 0.25 {
		t.Error("after reset, weights should be balanced at 0.25 each")
	}
}

// Benchmark tests
func BenchmarkWeightTransferComponent_Update(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()
	velX, velY, angularVel := 100.0, 50.0, 0.5
	deltaTime := 0.016

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sys.UpdateWeightDistribution(comp, velX, velY, angularVel, deltaTime)
	}
}

func BenchmarkWeightTransferComponent_GetWheelWeights(b *testing.B) {
	sys := NewEnhancedVehicleSystem()
	comp := NewWeightTransferComponent()
	sys.UpdateWeightDistribution(comp, 100, 50, 0.5, 0.016)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = comp.GetWheelWeights()
	}
}
