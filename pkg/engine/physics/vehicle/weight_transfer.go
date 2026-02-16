// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

// WeightTransferComponent stores dynamic weight distribution data during vehicle maneuvers.
// Weight shifts during acceleration, braking, and turning affect tire grip and handling.
type WeightTransferComponent struct {
	// Current acceleration vector (pixels/s²)
	AccelerationX float64
	AccelerationY float64

	// Angular acceleration (radians/s²)
	AngularAccel float64

	// Center of mass height above wheel axles (pixels)
	CenterOfMassHeight float64

	// Wheelbase dimensions (pixels)
	Wheelbase  float64 // Front-to-rear distance
	TrackWidth float64 // Left-to-right distance

	// Weight distribution percentages [0.0, 1.0]
	// These get modified by acceleration/braking/turning
	FrontLeftWeight  float64
	FrontRightWeight float64
	RearLeftWeight   float64
	RearRightWeight  float64

	// Static (resting) weight distribution
	StaticFrontWeight float64 // Percentage of weight on front axle (0.0-1.0)

	// Previous velocity for acceleration calculation
	PrevVelocityX  float64
	PrevVelocityY  float64
	PrevAngularVel float64

	// Performance tracking
	LastTransferMagnitude float64
}

// Type returns the component type identifier.
func (w *WeightTransferComponent) Type() string {
	return "weight_transfer"
}

// NewWeightTransferComponent creates a weight transfer component with default parameters.
func NewWeightTransferComponent() *WeightTransferComponent {
	return &WeightTransferComponent{
		CenterOfMassHeight: 15.0, // 15 pixels above axles
		Wheelbase:          32.0, // 32 pixels front-to-rear
		TrackWidth:         16.0, // 16 pixels left-to-right
		StaticFrontWeight:  0.5,  // 50% front, 50% rear (balanced)
		FrontLeftWeight:    0.25, // 25% per wheel
		FrontRightWeight:   0.25,
		RearLeftWeight:     0.25,
		RearRightWeight:    0.25,
	}
}

// GetWheelWeights returns the current weight distribution for all wheels.
// Returns: [frontLeft, frontRight, rearLeft, rearRight] as percentages [0.0, 1.0]
func (w *WeightTransferComponent) GetWheelWeights() [4]float64 {
	return [4]float64{
		w.FrontLeftWeight,
		w.FrontRightWeight,
		w.RearLeftWeight,
		w.RearRightWeight,
	}
}

// GetFrontAxleWeight returns the percentage of weight on the front axle.
func (w *WeightTransferComponent) GetFrontAxleWeight() float64 {
	return w.FrontLeftWeight + w.FrontRightWeight
}

// GetRearAxleWeight returns the percentage of weight on the rear axle.
func (w *WeightTransferComponent) GetRearAxleWeight() float64 {
	return w.RearLeftWeight + w.RearRightWeight
}

// GetLeftSideWeight returns the percentage of weight on the left side.
func (w *WeightTransferComponent) GetLeftSideWeight() float64 {
	return w.FrontLeftWeight + w.RearLeftWeight
}

// GetRightSideWeight returns the percentage of weight on the right side.
func (w *WeightTransferComponent) GetRightSideWeight() float64 {
	return w.FrontRightWeight + w.RearRightWeight
}

// GetTransferMagnitude returns the magnitude of the last weight transfer.
// Useful for visual feedback (vehicle leaning in turns, nose-diving during braking).
func (w *WeightTransferComponent) GetTransferMagnitude() float64 {
	return w.LastTransferMagnitude
}

// Reset resets the component to static weight distribution.
func (w *WeightTransferComponent) Reset() {
	w.AccelerationX = 0.0
	w.AccelerationY = 0.0
	w.AngularAccel = 0.0
	w.PrevVelocityX = 0.0
	w.PrevVelocityY = 0.0
	w.PrevAngularVel = 0.0
	w.FrontLeftWeight = 0.25
	w.FrontRightWeight = 0.25
	w.RearLeftWeight = 0.25
	w.RearRightWeight = 0.25
	w.LastTransferMagnitude = 0.0
}
