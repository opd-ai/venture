// Package vehicle provides vehicle-specific physics including suspension and weight transfer.
package vehicle

import (
	"math"
)

// WeightTransferComponent simulates dynamic weight distribution during vehicle maneuvers.
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
	Wheelbase float64 // Front-to-rear distance
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
	PrevVelocityX float64
	PrevVelocityY float64
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
		CenterOfMassHeight: 15.0,  // 15 pixels above axles
		Wheelbase:          32.0,  // 32 pixels front-to-rear
		TrackWidth:         16.0,  // 16 pixels left-to-right
		StaticFrontWeight:  0.5,   // 50% front, 50% rear (balanced)
		FrontLeftWeight:    0.25,  // 25% per wheel
		FrontRightWeight:   0.25,
		RearLeftWeight:     0.25,
		RearRightWeight:    0.25,
	}
}

// Update calculates weight transfer based on current vehicle dynamics.
// velocityX, velocityY: current velocity (pixels/s)
// angularVel: current angular velocity (radians/s)
// deltaTime: time step (seconds)
func (w *WeightTransferComponent) Update(velocityX, velocityY, angularVel, deltaTime float64) {
	if deltaTime <= 0 {
		return
	}

	// Calculate linear acceleration
	w.AccelerationX = (velocityX - w.PrevVelocityX) / deltaTime
	w.AccelerationY = (velocityY - w.PrevVelocityY) / deltaTime

	// Calculate angular acceleration
	w.AngularAccel = (angularVel - w.PrevAngularVel) / deltaTime

	// Store current values for next frame
	w.PrevVelocityX = velocityX
	w.PrevVelocityY = velocityY
	w.PrevAngularVel = angularVel

	// Calculate longitudinal weight transfer (acceleration/braking)
	// Positive acceleration shifts weight rearward, braking shifts forward
	longitudinalTransfer := w.calculateLongitudinalTransfer()

	// Calculate lateral weight transfer (turning)
	// Left turn shifts weight right, right turn shifts weight left
	lateralTransfer := w.calculateLateralTransfer()

	// Apply transfers to wheel weights
	w.applyWeightTransfers(longitudinalTransfer, lateralTransfer)

	// Track magnitude for debugging/visualization
	w.LastTransferMagnitude = math.Sqrt(longitudinalTransfer*longitudinalTransfer + lateralTransfer*lateralTransfer)
}

// calculateLongitudinalTransfer computes front-rear weight shift.
// Returns: percentage of weight to transfer from rear to front (negative = rear-heavy)
func (w *WeightTransferComponent) calculateLongitudinalTransfer() float64 {
	if w.Wheelbase <= 0 {
		return 0.0
	}

	// Calculate acceleration magnitude in vehicle's forward direction
	accelMagnitude := math.Sqrt(w.AccelerationX*w.AccelerationX + w.AccelerationY*w.AccelerationY)

	// Weight transfer formula: ΔW = (m * a * h) / L
	// Where:
	//   m = mass (normalized to 1.0)
	//   a = longitudinal acceleration
	//   h = center of mass height
	//   L = wheelbase
	//
	// Simplified: transfer = (a * h) / L
	
	// Determine if accelerating or braking based on velocity change
	isAccelerating := w.AccelerationX > 0 || w.AccelerationY > 0

	transfer := (accelMagnitude * w.CenterOfMassHeight) / w.Wheelbase

	// Positive transfer = weight shifts forward (braking)
	// Negative transfer = weight shifts rearward (acceleration)
	if isAccelerating {
		transfer = -transfer
	}

	// Clamp to reasonable range [-0.3, 0.3] (max 30% shift)
	if transfer > 0.3 {
		transfer = 0.3
	} else if transfer < -0.3 {
		transfer = -0.3
	}

	return transfer
}

// calculateLateralTransfer computes left-right weight shift during turning.
// Returns: percentage of weight to transfer from inside to outside of turn
func (w *WeightTransferComponent) calculateLateralTransfer() float64 {
	if w.TrackWidth <= 0 {
		return 0.0
	}

	// Lateral acceleration from turning
	// F = m * v² / r, but we approximate with angular acceleration
	lateralAccel := math.Abs(w.AngularAccel) * w.TrackWidth

	// Weight transfer formula similar to longitudinal:
	// ΔW = (a_lateral * h) / T
	// Where T = track width
	transfer := (lateralAccel * w.CenterOfMassHeight) / w.TrackWidth

	// Clamp to reasonable range [0.0, 0.25] (max 25% shift)
	if transfer > 0.25 {
		transfer = 0.25
	}

	return transfer
}

// applyWeightTransfers distributes weight to individual wheels based on transfers.
func (w *WeightTransferComponent) applyWeightTransfers(longitudinal, lateral float64) {
	// Start with static distribution
	frontWeight := w.StaticFrontWeight
	rearWeight := 1.0 - w.StaticFrontWeight

	// Apply longitudinal transfer
	frontWeight += longitudinal
	rearWeight -= longitudinal

	// Ensure valid range
	if frontWeight < 0.1 {
		frontWeight = 0.1
	} else if frontWeight > 0.9 {
		frontWeight = 0.9
	}
	rearWeight = 1.0 - frontWeight

	// Distribute front and rear to left/right wheels
	// Lateral transfer affects left-right balance
	leftRatio := 0.5
	rightRatio := 0.5

	// Determine turn direction from angular acceleration
	// Positive angular accel = left turn = weight shifts right
	// Negative angular accel = right turn = weight shifts left
	if w.AngularAccel > 0 {
		// Left turn - shift weight right
		rightRatio += lateral
		leftRatio -= lateral
	} else if w.AngularAccel < 0 {
		// Right turn - shift weight left
		leftRatio += lateral
		rightRatio -= lateral
	}

	// Ensure valid ratios
	if leftRatio < 0.1 {
		leftRatio = 0.1
	} else if leftRatio > 0.9 {
		leftRatio = 0.9
	}
	rightRatio = 1.0 - leftRatio

	// Final distribution to all four wheels
	w.FrontLeftWeight = frontWeight * leftRatio
	w.FrontRightWeight = frontWeight * rightRatio
	w.RearLeftWeight = rearWeight * leftRatio
	w.RearRightWeight = rearWeight * rightRatio

	// Normalize to ensure sum = 1.0 (handles floating point errors)
	total := w.FrontLeftWeight + w.FrontRightWeight + w.RearLeftWeight + w.RearRightWeight
	if total > 0 {
		w.FrontLeftWeight /= total
		w.FrontRightWeight /= total
		w.RearLeftWeight /= total
		w.RearRightWeight /= total
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
