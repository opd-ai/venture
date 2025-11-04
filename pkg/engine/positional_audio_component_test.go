package engine

import (
	"math"
	"testing"
)

func TestPositionalAudioComponent_Type(t *testing.T) {
	comp := NewPositionalAudioComponent(1000)
	if comp.Type() != "positional_audio" {
		t.Errorf("Expected type 'positional_audio', got '%s'", comp.Type())
	}
}

func TestNewPositionalAudioComponent(t *testing.T) {
	tests := []struct {
		name        string
		maxDistance float64
		wantFalloff float64
		wantFactor  float64
	}{
		{
			name:        "close sound (800px)",
			maxDistance: 800,
			wantFalloff: 2.0,
			wantFactor:  0.3,
		},
		{
			name:        "ambient sound (2000px)",
			maxDistance: 2000,
			wantFalloff: 2.0,
			wantFactor:  0.3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewPositionalAudioComponent(tt.maxDistance)

			if comp.MaxDistance != tt.maxDistance {
				t.Errorf("MaxDistance = %v, want %v", comp.MaxDistance, tt.maxDistance)
			}
			if comp.FalloffExponent != tt.wantFalloff {
				t.Errorf("FalloffExponent = %v, want %v", comp.FalloffExponent, tt.wantFalloff)
			}
			if comp.OcclusionFactor != tt.wantFactor {
				t.Errorf("OcclusionFactor = %v, want %v", comp.OcclusionFactor, tt.wantFactor)
			}
			if !comp.OcclusionEnabled {
				t.Error("OcclusionEnabled should be true by default")
			}
			if comp.VolumeMultiplier != 1.0 {
				t.Errorf("VolumeMultiplier = %v, want 1.0", comp.VolumeMultiplier)
			}
			if comp.StereoPan != 0.0 {
				t.Errorf("StereoPan = %v, want 0.0", comp.StereoPan)
			}
		})
	}
}

func TestReverbComponent_Type(t *testing.T) {
	comp := NewReverbComponent(RoomSizeMedium, RoomMaterialStone)
	if comp.Type() != "reverb" {
		t.Errorf("Expected type 'reverb', got '%s'", comp.Type())
	}
}

func TestNewReverbComponent(t *testing.T) {
	tests := []struct {
		name           string
		roomSize       RoomSize
		material       RoomMaterial
		wantPreDelay   float64
		wantDecay      float64
		wantDampingMin float64
		wantDampingMax float64
	}{
		{
			name:           "small stone room",
			roomSize:       RoomSizeSmall,
			material:       RoomMaterialStone,
			wantPreDelay:   0.01,
			wantDecay:      0.3,
			wantDampingMin: 0.0,
			wantDampingMax: 0.2,
		},
		{
			name:           "medium wood room",
			roomSize:       RoomSizeMedium,
			material:       RoomMaterialWood,
			wantPreDelay:   0.02,
			wantDecay:      0.8,
			wantDampingMin: 0.2,
			wantDampingMax: 0.4,
		},
		{
			name:           "large cloth room",
			roomSize:       RoomSizeLarge,
			material:       RoomMaterialCloth,
			wantPreDelay:   0.04,
			wantDecay:      1.5,
			wantDampingMin: 0.5,
			wantDampingMax: 0.7,
		},
		{
			name:           "huge metal room",
			roomSize:       RoomSizeHuge,
			material:       RoomMaterialMetal,
			wantPreDelay:   0.08,
			wantDecay:      3.0,
			wantDampingMin: 0.0,
			wantDampingMax: 0.1,
		},
		{
			name:           "medium carpet room",
			roomSize:       RoomSizeMedium,
			material:       RoomMaterialCarpet,
			wantPreDelay:   0.02,
			wantDecay:      0.8,
			wantDampingMin: 0.7,
			wantDampingMax: 0.9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewReverbComponent(tt.roomSize, tt.material)

			if comp.RoomSize != tt.roomSize {
				t.Errorf("RoomSize = %v, want %v", comp.RoomSize, tt.roomSize)
			}
			if comp.Material != tt.material {
				t.Errorf("Material = %v, want %v", comp.Material, tt.material)
			}
			if comp.PreDelay != tt.wantPreDelay {
				t.Errorf("PreDelay = %v, want %v", comp.PreDelay, tt.wantPreDelay)
			}
			if comp.DecayTime != tt.wantDecay {
				t.Errorf("DecayTime = %v, want %v", comp.DecayTime, tt.wantDecay)
			}
			if comp.Damping < tt.wantDampingMin || comp.Damping > tt.wantDampingMax {
				t.Errorf("Damping = %v, want range [%v, %v]", comp.Damping, tt.wantDampingMin, tt.wantDampingMax)
			}
			if comp.Amount != 0.5 {
				t.Errorf("Amount = %v, want 0.5", comp.Amount)
			}
			if !comp.Enabled {
				t.Error("Enabled should be true by default")
			}
		})
	}
}

func TestRoomSize_String(t *testing.T) {
	tests := []struct {
		size RoomSize
		want string
	}{
		{RoomSizeSmall, "Small"},
		{RoomSizeMedium, "Medium"},
		{RoomSizeLarge, "Large"},
		{RoomSizeHuge, "Huge"},
		{RoomSize(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.size.String(); got != tt.want {
				t.Errorf("RoomSize.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoomMaterial_String(t *testing.T) {
	tests := []struct {
		material RoomMaterial
		want     string
	}{
		{RoomMaterialStone, "Stone"},
		{RoomMaterialWood, "Wood"},
		{RoomMaterialCloth, "Cloth"},
		{RoomMaterialMetal, "Metal"},
		{RoomMaterialCarpet, "Carpet"},
		{RoomMaterial(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.material.String(); got != tt.want {
				t.Errorf("RoomMaterial.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateDistance(t *testing.T) {
	tests := []struct {
		name string
		x1   float64
		y1   float64
		x2   float64
		y2   float64
		want float64
	}{
		{
			name: "same position",
			x1:   0, y1: 0, x2: 0, y2: 0,
			want: 0,
		},
		{
			name: "horizontal distance",
			x1:   0, y1: 0, x2: 100, y2: 0,
			want: 100,
		},
		{
			name: "vertical distance",
			x1:   0, y1: 0, x2: 0, y2: 100,
			want: 100,
		},
		{
			name: "diagonal distance (3-4-5 triangle)",
			x1:   0, y1: 0, x2: 300, y2: 400,
			want: 500,
		},
		{
			name: "negative coordinates",
			x1:   -50, y1: -50, x2: 50, y2: 50,
			want: math.Sqrt(20000), // sqrt((100)^2 + (100)^2)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateDistance(tt.x1, tt.y1, tt.x2, tt.y2)
			if math.Abs(got-tt.want) > 0.0001 {
				t.Errorf("CalculateDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateVolumeFalloff(t *testing.T) {
	tests := []struct {
		name     string
		distance float64
		maxDist  float64
		falloff  float64
		wantMin  float64
		wantMax  float64
	}{
		{
			name:     "at source (distance 0)",
			distance: 0,
			maxDist:  1000,
			falloff:  2.0,
			wantMin:  0.99,
			wantMax:  1.0,
		},
		{
			name:     "half distance linear falloff",
			distance: 500,
			maxDist:  1000,
			falloff:  1.0,
			wantMin:  0.49,
			wantMax:  0.51,
		},
		{
			name:     "half distance inverse square",
			distance: 500,
			maxDist:  1000,
			falloff:  2.0,
			wantMin:  0.74,
			wantMax:  0.76,
		},
		{
			name:     "at max distance",
			distance: 1000,
			maxDist:  1000,
			falloff:  2.0,
			wantMin:  0.0,
			wantMax:  0.0,
		},
		{
			name:     "beyond max distance",
			distance: 1500,
			maxDist:  1000,
			falloff:  2.0,
			wantMin:  0.0,
			wantMax:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateVolumeFalloff(tt.distance, tt.maxDist, tt.falloff)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateVolumeFalloff() = %v, want range [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateStereoPan(t *testing.T) {
	tests := []struct {
		name      string
		sourceX   float64
		sourceY   float64
		listenerX float64
		listenerY float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:    "same position (center)",
			sourceX: 100, sourceY: 100,
			listenerX: 100, listenerY: 100,
			wantMin: -0.01,
			wantMax: 0.01,
		},
		{
			name:    "source to the right",
			sourceX: 1000, sourceY: 100,
			listenerX: 100, listenerY: 100,
			wantMin: 0.8,
			wantMax: 1.0,
		},
		{
			name:    "source to the left",
			sourceX: 100, sourceY: 100,
			listenerX: 1000, listenerY: 100,
			wantMin: -1.0,
			wantMax: -0.8,
		},
		{
			name:    "source slightly right",
			sourceX: 600, sourceY: 100,
			listenerX: 500, listenerY: 100,
			wantMin: 0.05,
			wantMax: 0.15,
		},
		{
			name:    "extreme right (clamped)",
			sourceX: 5000, sourceY: 100,
			listenerX: 100, listenerY: 100,
			wantMin: 0.99,
			wantMax: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateStereoPan(tt.sourceX, tt.sourceY, tt.listenerX, tt.listenerY)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("CalculateStereoPan() = %v, want range [%v, %v]", got, tt.wantMin, tt.wantMax)
			}
			// Ensure clamping
			if got < -1.0 || got > 1.0 {
				t.Errorf("CalculateStereoPan() = %v, out of valid range [-1.0, 1.0]", got)
			}
		})
	}
}
