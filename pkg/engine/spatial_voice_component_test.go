package engine

import (
	"math"
	"testing"
)

func TestSpatialVoiceComponent_Type(t *testing.T) {
	c := NewSpatialVoiceComponent()
	if c.Type() != "spatial_voice" {
		t.Errorf("expected type 'spatial_voice', got '%s'", c.Type())
	}
}

func TestNewSpatialVoiceComponent(t *testing.T) {
	c := NewSpatialVoiceComponent()

	if !c.Enabled {
		t.Error("expected Enabled to be true")
	}
	if c.MaxRange != 500.0 {
		t.Errorf("expected MaxRange 500.0, got %f", c.MaxRange)
	}
	if c.MinRange != 50.0 {
		t.Errorf("expected MinRange 50.0, got %f", c.MinRange)
	}
	if c.FalloffCurve != VoiceFalloffLogarithmic {
		t.Errorf("expected FalloffCurve logarithmic, got %s", c.FalloffCurve)
	}
	if c.MinVolume != 0.0 {
		t.Errorf("expected MinVolume 0.0, got %f", c.MinVolume)
	}
	if c.IsAudible {
		t.Error("expected IsAudible to be false initially")
	}
}

func TestNewSpatialVoiceComponentWithRange(t *testing.T) {
	c := NewSpatialVoiceComponentWithRange(100.0, 800.0)

	if c.MinRange != 100.0 {
		t.Errorf("expected MinRange 100.0, got %f", c.MinRange)
	}
	if c.MaxRange != 800.0 {
		t.Errorf("expected MaxRange 800.0, got %f", c.MaxRange)
	}
}

func TestSpatialVoiceComponent_SetRange(t *testing.T) {
	c := NewSpatialVoiceComponent()

	c.SetRange(25.0, 300.0)
	if c.MinRange != 25.0 {
		t.Errorf("expected MinRange 25.0, got %f", c.MinRange)
	}
	if c.MaxRange != 300.0 {
		t.Errorf("expected MaxRange 300.0, got %f", c.MaxRange)
	}

	// Test clamping
	c.SetRange(-10.0, 200.0)
	if c.MinRange != 0 {
		t.Errorf("expected MinRange 0 (clamped), got %f", c.MinRange)
	}

	// Max should be at least min
	c.SetRange(100.0, 50.0)
	if c.MaxRange < c.MinRange {
		t.Error("expected MaxRange >= MinRange")
	}
}

func TestSpatialVoiceComponent_SetFalloffCurve(t *testing.T) {
	c := NewSpatialVoiceComponent()

	c.SetFalloffCurve(VoiceFalloffLinear)
	if c.FalloffCurve != VoiceFalloffLinear {
		t.Errorf("expected FalloffCurve linear, got %s", c.FalloffCurve)
	}

	c.SetFalloffCurve(VoiceFalloffExponential)
	if c.FalloffCurve != VoiceFalloffExponential {
		t.Errorf("expected FalloffCurve exponential, got %s", c.FalloffCurve)
	}
}

func TestSpatialVoiceComponent_SetEnabled(t *testing.T) {
	c := NewSpatialVoiceComponent()

	c.SetEnabled(false)
	if c.Enabled {
		t.Error("expected Enabled to be false")
	}

	c.SetEnabled(true)
	if !c.Enabled {
		t.Error("expected Enabled to be true")
	}
}

func TestSpatialVoiceComponent_UpdatePositions(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffLinear)

	// Source at origin, listener at (50, 0)
	c.UpdatePositions(0, 0, 50, 0)

	if c.CurrentDistance != 50.0 {
		t.Errorf("expected CurrentDistance 50.0, got %f", c.CurrentDistance)
	}
	if !c.IsAudible {
		t.Error("expected IsAudible to be true")
	}
	if c.CurrentVolume != 0.5 {
		t.Errorf("expected CurrentVolume 0.5, got %f", c.CurrentVolume)
	}

	// Source beyond max range
	c.UpdatePositions(0, 0, 150, 0)
	if c.IsAudible {
		t.Error("expected IsAudible to be false beyond max range")
	}
	if c.CurrentVolume != 0 {
		t.Errorf("expected CurrentVolume 0, got %f", c.CurrentVolume)
	}
}

func TestSpatialVoiceComponent_DistanceCalculation(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)

	tests := []struct {
		name         string
		sourceX      float64
		sourceY      float64
		listenerX    float64
		listenerY    float64
		wantDistance float64
	}{
		{"same position", 0, 0, 0, 0, 0},
		{"horizontal", 100, 0, 0, 0, 100},
		{"vertical", 0, 50, 0, 0, 50},
		{"diagonal", 30, 40, 0, 0, 50}, // 3-4-5 triangle
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.UpdatePositions(tt.sourceX, tt.sourceY, tt.listenerX, tt.listenerY)
			if math.Abs(c.CurrentDistance-tt.wantDistance) > 0.001 {
				t.Errorf("expected distance %f, got %f", tt.wantDistance, c.CurrentDistance)
			}
		})
	}
}

func TestSpatialVoiceComponent_VolumeAtDistance_Linear(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffLinear)

	tests := []struct {
		name       string
		distance   float64
		wantVolume float64
	}{
		{"at source", 0, 1.0},
		{"quarter range", 25, 0.75},
		{"half range", 50, 0.5},
		{"three quarter range", 75, 0.25},
		{"at max range", 100, 0.0},
		{"beyond max range", 150, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume := c.GetVolumeAtDistance(tt.distance)
			if math.Abs(volume-tt.wantVolume) > 0.001 {
				t.Errorf("expected volume %f at distance %f, got %f", tt.wantVolume, tt.distance, volume)
			}
		})
	}
}

func TestSpatialVoiceComponent_VolumeAtDistance_WithMinRange(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(50, 100)
	c.SetFalloffCurve(VoiceFalloffLinear)

	tests := []struct {
		name       string
		distance   float64
		wantVolume float64
	}{
		{"at source", 0, 1.0},
		{"within min range", 25, 1.0},
		{"at min range", 50, 1.0},
		{"midway", 75, 0.5},
		{"at max range", 100, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volume := c.GetVolumeAtDistance(tt.distance)
			if math.Abs(volume-tt.wantVolume) > 0.001 {
				t.Errorf("expected volume %f at distance %f, got %f", tt.wantVolume, tt.distance, volume)
			}
		})
	}
}

func TestSpatialVoiceComponent_Logarithmic(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffLogarithmic)

	// Test that logarithmic falloff works
	c.UpdatePositions(0, 0, 25, 0)
	logVol25 := c.CurrentVolume

	// Volume should be between 0 and 1
	if logVol25 <= 0 || logVol25 >= 1 {
		t.Errorf("logarithmic volume at 25 units should be between 0 and 1, got %f", logVol25)
	}

	// At half distance, volume should be reduced
	c.UpdatePositions(0, 0, 50, 0)
	logVol50 := c.CurrentVolume

	if logVol50 >= logVol25 {
		t.Errorf("expected volume at 50 (%f) < volume at 25 (%f)", logVol50, logVol25)
	}
}

func TestSpatialVoiceComponent_Exponential(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffExponential)

	// Exponential should drop quickly then have a slow tail
	c.UpdatePositions(0, 0, 50, 0)
	expVol50 := c.CurrentVolume

	c.SetFalloffCurve(VoiceFalloffLinear)
	c.UpdatePositions(0, 0, 50, 0)
	linVol50 := c.CurrentVolume

	// At 50% distance, exponential (0.25) should be quieter than linear (0.5)
	if expVol50 >= linVol50 {
		t.Errorf("expected exponential (%f) < linear (%f) at mid range", expVol50, linVol50)
	}
}

func TestSpatialVoiceComponent_Pan(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)

	tests := []struct {
		name      string
		sourceX   float64
		sourceY   float64
		wantPan   float64
		tolerance float64
	}{
		{"center", 0, 0, 0, 0.01},
		{"right", 50, 0, 1.0, 0.01},
		{"left", -50, 0, -1.0, 0.01},
		{"partial right", 25, 0, 0.5, 0.01},
		{"far right clamped", 100, 0, 1.0, 0.01}, // Should clamp to 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c.UpdatePositions(tt.sourceX, tt.sourceY, 0, 0)
			if math.Abs(c.CurrentPan-tt.wantPan) > tt.tolerance {
				t.Errorf("expected pan %f, got %f", tt.wantPan, c.CurrentPan)
			}
		})
	}
}

func TestSpatialVoiceComponent_DisabledNoVolume(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetEnabled(false)

	c.UpdatePositions(0, 0, 25, 0)

	if c.CurrentVolume != 0 {
		t.Errorf("expected CurrentVolume 0 when disabled, got %f", c.CurrentVolume)
	}
	if c.GetVolumeAtDistance(25) != 0 {
		t.Error("expected GetVolumeAtDistance to return 0 when disabled")
	}
}

func TestSpatialVoiceComponent_IsWithinRange(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)

	c.UpdatePositions(0, 0, 50, 0)
	if !c.IsWithinRange() {
		t.Error("expected IsWithinRange true at 50 units")
	}

	c.UpdatePositions(0, 0, 150, 0)
	if c.IsWithinRange() {
		t.Error("expected IsWithinRange false at 150 units")
	}

	// Disabled
	c.UpdatePositions(0, 0, 50, 0)
	c.SetEnabled(false)
	if c.IsWithinRange() {
		t.Error("expected IsWithinRange false when disabled")
	}
}

func TestSpatialVoiceComponent_GetEffectiveVolume(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffLinear)

	c.UpdatePositions(0, 0, 50, 0) // CurrentVolume = 0.5

	effectiveVol := c.GetEffectiveVolume(0.8)
	expected := 0.8 * 0.5
	if math.Abs(effectiveVol-expected) > 0.001 {
		t.Errorf("expected effective volume %f, got %f", expected, effectiveVol)
	}
}

func TestSpatialVoiceComponent_MinVolume(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)
	c.SetFalloffCurve(VoiceFalloffLinear)
	c.MinVolume = 0.1 // Minimum volume of 10%

	// At max range, should be min volume
	vol := c.GetVolumeAtDistance(100)
	if math.Abs(vol-0.1) > 0.001 {
		t.Errorf("expected min volume 0.1, got %f", vol)
	}

	// At half range, should be between min and max
	vol = c.GetVolumeAtDistance(50)
	// Linear: 0.1 + (1.0 - 0.1) * 0.5 = 0.1 + 0.45 = 0.55
	if math.Abs(vol-0.55) > 0.001 {
		t.Errorf("expected volume 0.55, got %f", vol)
	}
}

func TestSpatialVoiceComponent_Serialize(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(100, 800)
	c.SetFalloffCurve(VoiceFalloffExponential)
	c.UpdatePositions(100, 200, 300, 400)

	data, err := c.Serialize()
	if err != nil {
		t.Fatalf("Serialize failed: %v", err)
	}

	c2 := &SpatialVoiceComponent{}
	err = c2.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialize failed: %v", err)
	}

	if c2.MinRange != c.MinRange {
		t.Errorf("expected MinRange %f, got %f", c.MinRange, c2.MinRange)
	}
	if c2.MaxRange != c.MaxRange {
		t.Errorf("expected MaxRange %f, got %f", c.MaxRange, c2.MaxRange)
	}
	if c2.FalloffCurve != c.FalloffCurve {
		t.Errorf("expected FalloffCurve %s, got %s", c.FalloffCurve, c2.FalloffCurve)
	}
	if c2.SourcePositionX != c.SourcePositionX {
		t.Errorf("expected SourcePositionX %f, got %f", c.SourcePositionX, c2.SourcePositionX)
	}
}

func TestSpatialVoiceComponent_ZeroRangeSpan(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(100, 100) // Min == Max

	// Should return full volume for any distance <= range
	vol := c.GetVolumeAtDistance(100)
	if vol != 1.0 {
		t.Errorf("expected volume 1.0 for zero range span, got %f", vol)
	}

	c.UpdatePositions(0, 0, 100, 0)
	if c.CurrentVolume != 1.0 {
		t.Errorf("expected CurrentVolume 1.0 for zero range span, got %f", c.CurrentVolume)
	}
}

func TestSpatialVoiceComponent_PanZeroDistance(t *testing.T) {
	c := NewSpatialVoiceComponent()
	c.SetRange(0, 100)

	// Same position should have no pan
	c.UpdatePositions(0, 0, 0, 0)
	if c.CurrentPan != 0 {
		t.Errorf("expected CurrentPan 0 at same position, got %f", c.CurrentPan)
	}
}
