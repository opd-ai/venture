// Package engine provides the EquipmentCrackOverlayComponent for per-entity
// procedural crack pattern data generated from equipment damage states.
// The render system uses these crack segments to draw wear-and-tear overlays
// on entity sprites, adding visual depth to equipment degradation.
package engine

// CrackSegment represents a single procedural crack line in normalized [0,1] space.
type CrackSegment struct {
	X1, Y1   float64 // Start point
	X2, Y2   float64 // End point
	Width    float64 // Line width (pixels)
	Depth    float64 // Visual depth/darkness (0.0–1.0)
	BranchID int     // Which crack tree this segment belongs to
}

// EquipmentCrackOverlayComponent stores procedural crack pattern data for the
// render pipeline. Segments are in normalized [0,1] coordinate space relative
// to the entity sprite bounds. This is a transient visual component.
type EquipmentCrackOverlayComponent struct {
	// Crack line segments to render as overlay
	Segments []CrackSegment

	// Overall crack intensity driving alpha blending (0.0–1.0)
	Intensity float64

	// Edge roughness offset pixels for sprite boundary distortion
	EdgeRoughness float64

	// Genre-specific crack color (RGB 0.0–1.0)
	ColorR float64
	ColorG float64
	ColorB float64

	// Whether overlay is active
	Enabled bool

	// Cache key to avoid regenerating unchanged patterns
	CacheKey uint64

	// Number of crack trees (branching origins)
	TreeCount int
}

// Type returns the component type identifier.
func (e *EquipmentCrackOverlayComponent) Type() string {
	return "equipment_crack_overlay"
}

// NewEquipmentCrackOverlayComponent creates a crack overlay component with no cracks.
func NewEquipmentCrackOverlayComponent() *EquipmentCrackOverlayComponent {
	return &EquipmentCrackOverlayComponent{
		Segments:      nil,
		Intensity:     0.0,
		EdgeRoughness: 0.0,
		ColorR:        0.2,
		ColorG:        0.2,
		ColorB:        0.2,
		Enabled:       false,
	}
}
