// Package engine provides core ECS components and systems for the Venture game.
package engine

import (
	"encoding/json"
	"math"
)

// VoiceFalloffCurve defines how voice volume decreases with distance.
type VoiceFalloffCurve string

const (
	// VoiceFalloffLinear decreases volume linearly with distance.
	VoiceFalloffLinear VoiceFalloffCurve = "linear"
	// VoiceFalloffLogarithmic decreases volume logarithmically (more realistic).
	VoiceFalloffLogarithmic VoiceFalloffCurve = "logarithmic"
	// VoiceFalloffExponential decreases volume exponentially (sharp falloff).
	VoiceFalloffExponential VoiceFalloffCurve = "exponential"
)

// SpatialVoiceComponent tracks spatial audio settings for proximity voice chat.
type SpatialVoiceComponent struct {
	// Enabled indicates if spatial audio is active for this entity.
	Enabled bool `json:"enabled"`

	// MaxRange is the maximum audible distance in world units.
	MaxRange float64 `json:"max_range"`

	// MinRange is the distance at which voice is at full volume.
	MinRange float64 `json:"min_range"`

	// FalloffCurve defines how volume decreases with distance.
	FalloffCurve VoiceFalloffCurve `json:"falloff_curve"`

	// MinVolume is the volume at max range (usually 0).
	MinVolume float64 `json:"min_volume"`

	// CurrentVolume is the calculated volume based on distance.
	CurrentVolume float64 `json:"current_volume"`

	// CurrentPan is the stereo pan based on relative position (-1.0 left, 1.0 right).
	CurrentPan float64 `json:"current_pan"`

	// ListenerPosition caches the last listener position for calculations.
	ListenerPositionX float64 `json:"listener_position_x"`
	ListenerPositionY float64 `json:"listener_position_y"`

	// SourcePositionX/Y tracks the voice source position.
	SourcePositionX float64 `json:"source_position_x"`
	SourcePositionY float64 `json:"source_position_y"`

	// CurrentDistance is the calculated distance to the listener.
	CurrentDistance float64 `json:"current_distance"`

	// IsAudible indicates if the source is within hearing range.
	IsAudible bool `json:"is_audible"`
}

// Type returns the component type identifier.
func (c *SpatialVoiceComponent) Type() string {
	return "spatial_voice"
}

// NewSpatialVoiceComponent creates a new spatial voice component with default values.
func NewSpatialVoiceComponent() *SpatialVoiceComponent {
	return &SpatialVoiceComponent{
		Enabled:       true,
		MaxRange:      500.0,
		MinRange:      50.0,
		FalloffCurve:  VoiceFalloffLogarithmic,
		MinVolume:     0.0,
		CurrentVolume: 0.0,
		CurrentPan:    0.0,
		IsAudible:     false,
	}
}

// NewSpatialVoiceComponentWithRange creates a component with custom range settings.
func NewSpatialVoiceComponentWithRange(minRange, maxRange float64) *SpatialVoiceComponent {
	c := NewSpatialVoiceComponent()
	c.MinRange = minRange
	c.MaxRange = maxRange
	return c
}

// SetRange sets the min and max audible range.
func (c *SpatialVoiceComponent) SetRange(minRange, maxRange float64) {
	c.MinRange = math.Max(0, minRange)
	c.MaxRange = math.Max(c.MinRange, maxRange)
}

// SetFalloffCurve sets the volume falloff curve type.
func (c *SpatialVoiceComponent) SetFalloffCurve(curve VoiceFalloffCurve) {
	c.FalloffCurve = curve
}

// SetEnabled enables or disables spatial audio.
func (c *SpatialVoiceComponent) SetEnabled(enabled bool) {
	c.Enabled = enabled
}

// UpdatePositions updates the source and listener positions.
func (c *SpatialVoiceComponent) UpdatePositions(sourceX, sourceY, listenerX, listenerY float64) {
	c.SourcePositionX = sourceX
	c.SourcePositionY = sourceY
	c.ListenerPositionX = listenerX
	c.ListenerPositionY = listenerY

	// Calculate distance
	dx := sourceX - listenerX
	dy := sourceY - listenerY
	c.CurrentDistance = math.Sqrt(dx*dx + dy*dy)

	// Check if audible
	c.IsAudible = c.CurrentDistance <= c.MaxRange

	// Calculate volume based on distance
	c.CurrentVolume = c.calculateVolume()

	// Calculate stereo pan based on relative X position
	c.CurrentPan = c.calculatePan()
}

// calculateVolume calculates the volume based on distance and falloff curve.
func (c *SpatialVoiceComponent) calculateVolume() float64 {
	if !c.Enabled || !c.IsAudible {
		return 0.0
	}

	// Within min range, full volume
	if c.CurrentDistance <= c.MinRange {
		return 1.0
	}

	// Beyond max range, min volume
	if c.CurrentDistance >= c.MaxRange {
		return c.MinVolume
	}

	// Calculate normalized distance (0.0 at min range, 1.0 at max range)
	rangeSpan := c.MaxRange - c.MinRange
	if rangeSpan <= 0 {
		return 1.0
	}
	normalizedDistance := (c.CurrentDistance - c.MinRange) / rangeSpan

	// Apply falloff curve
	var volumeFactor float64
	switch c.FalloffCurve {
	case VoiceFalloffLinear:
		volumeFactor = 1.0 - normalizedDistance
	case VoiceFalloffLogarithmic:
		// Logarithmic falloff (sounds more natural)
		volumeFactor = 1.0 - math.Log10(1+9*normalizedDistance)/math.Log10(10)
	case VoiceFalloffExponential:
		// Exponential falloff (quick drop then slow tail)
		volumeFactor = math.Pow(1.0-normalizedDistance, 2)
	default:
		volumeFactor = 1.0 - normalizedDistance
	}

	// Clamp and apply min volume
	volumeFactor = math.Max(0, math.Min(1, volumeFactor))
	return c.MinVolume + (1.0-c.MinVolume)*volumeFactor
}

// calculatePan calculates the stereo pan based on relative position.
func (c *SpatialVoiceComponent) calculatePan() float64 {
	if !c.Enabled || !c.IsAudible || c.CurrentDistance == 0 {
		return 0.0
	}

	// Calculate relative X position
	dx := c.SourcePositionX - c.ListenerPositionX

	// Normalize pan based on distance (farther = more centered)
	maxPanDistance := c.MaxRange / 2
	if maxPanDistance <= 0 {
		return 0.0
	}

	pan := dx / maxPanDistance
	return math.Max(-1.0, math.Min(1.0, pan))
}

// GetVolumeAtDistance calculates volume for a given distance without updating state.
func (c *SpatialVoiceComponent) GetVolumeAtDistance(distance float64) float64 {
	if !c.Enabled || distance > c.MaxRange {
		return 0.0
	}
	if distance <= c.MinRange {
		return 1.0
	}

	rangeSpan := c.MaxRange - c.MinRange
	if rangeSpan <= 0 {
		return 1.0
	}
	normalizedDistance := (distance - c.MinRange) / rangeSpan

	var volumeFactor float64
	switch c.FalloffCurve {
	case VoiceFalloffLinear:
		volumeFactor = 1.0 - normalizedDistance
	case VoiceFalloffLogarithmic:
		volumeFactor = 1.0 - math.Log10(1+9*normalizedDistance)/math.Log10(10)
	case VoiceFalloffExponential:
		volumeFactor = math.Pow(1.0-normalizedDistance, 2)
	default:
		volumeFactor = 1.0 - normalizedDistance
	}

	volumeFactor = math.Max(0, math.Min(1, volumeFactor))
	return c.MinVolume + (1.0-c.MinVolume)*volumeFactor
}

// IsWithinRange returns true if the current distance is within audible range.
func (c *SpatialVoiceComponent) IsWithinRange() bool {
	return c.Enabled && c.IsAudible
}

// GetEffectiveVolume returns the volume multiplied by any additional factors.
func (c *SpatialVoiceComponent) GetEffectiveVolume(baseVolume float64) float64 {
	return baseVolume * c.CurrentVolume
}

// Serialize converts the component to JSON bytes.
func (c *SpatialVoiceComponent) Serialize() ([]byte, error) {
	return json.Marshal(c)
}

// Deserialize loads the component from JSON bytes.
func (c *SpatialVoiceComponent) Deserialize(data []byte) error {
	return json.Unmarshal(data, c)
}
