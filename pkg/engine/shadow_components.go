// Package engine provides shadow components for the ECS.
// This file defines components for shadow casting, which works in conjunction
// with the lighting system to create depth and atmosphere through realistic
// shadow rendering. Shadows are cast by entities with ShadowComponent when
// illuminated by light sources.
//
// Design Philosophy:
// - Components contain only data, no behavior
// - Shadow calculations happen in ShadowSystem
// - Integration with existing lighting pipeline
// - Performance-conscious with viewport culling
package engine

import (
	"image/color"
)

// ShadowType defines different shadow rendering approaches.
type ShadowType int

const (
	// ShadowTypeHard creates sharp shadow edges (fastest)
	ShadowTypeHard ShadowType = iota
	// ShadowTypeSoft creates soft shadow edges with penumbra (slower)
	ShadowTypeSoft
	// ShadowTypeContact creates contact shadows at ground intersection
	ShadowTypeContact
)

// String returns the string representation of shadow type.
func (s ShadowType) String() string {
	switch s {
	case ShadowTypeHard:
		return "hard"
	case ShadowTypeSoft:
		return "soft"
	case ShadowTypeContact:
		return "contact"
	default:
		return "unknown"
	}
}

// ShadowComponent marks an entity as casting shadows.
// When illuminated by light sources, entities with this component
// will cast shadows based on their shape and the light position.
type ShadowComponent struct {
	// Enabled allows shadows to be toggled on/off without removal
	Enabled bool

	// ShadowType determines the shadow rendering approach
	ShadowType ShadowType

	// Opacity controls shadow darkness (0.0 = invisible, 1.0 = opaque)
	Opacity float64

	// Radius defines the entity's shadow-casting radius (pixels)
	// This should generally match or slightly exceed sprite size
	Radius float64

	// Height affects shadow size/position for contact shadows
	// Higher entities cast larger shadows offset from position
	Height float64

	// CastsShadow determines if this entity blocks light
	CastsShadow bool

	// ReceivesShadow determines if shadows are rendered on this entity
	ReceivesShadow bool

	// SoftEdgeRadius for soft shadows (penumbra size in pixels)
	SoftEdgeRadius float64

	// Color tint for shadow (usually black or dark gray)
	Color color.RGBA
}

// Type returns the component type identifier.
func (s *ShadowComponent) Type() string {
	return "shadow"
}

// NewShadowComponent creates a shadow component with default values.
func NewShadowComponent(radius float64) *ShadowComponent {
	if radius <= 0 {
		radius = 16 // Default to typical sprite size
	}

	return &ShadowComponent{
		Enabled:        true,
		ShadowType:     ShadowTypeHard,
		Opacity:        0.5,
		Radius:         radius,
		Height:         0,
		CastsShadow:    true,
		ReceivesShadow: true,
		SoftEdgeRadius: 4.0,
		Color:          color.RGBA{0, 0, 0, 128}, // Semi-transparent black
	}
}

// NewSoftShadow creates a shadow component with soft edges.
func NewSoftShadow(radius, softEdgeRadius float64) *ShadowComponent {
	shadow := NewShadowComponent(radius)
	shadow.ShadowType = ShadowTypeSoft
	shadow.SoftEdgeRadius = softEdgeRadius
	return shadow
}

// NewContactShadow creates a contact shadow (for entities touching ground).
func NewContactShadow(radius, height float64) *ShadowComponent {
	shadow := NewShadowComponent(radius)
	shadow.ShadowType = ShadowTypeContact
	shadow.Height = height
	shadow.Opacity = 0.3 // Contact shadows are typically lighter
	return shadow
}

// AmbientOcclusionComponent marks an entity as having ambient occlusion.
// This creates subtle darkening in corners and at contact points to enhance
// depth perception. Ambient occlusion is view-independent (unlike shadows).
type AmbientOcclusionComponent struct {
	// Enabled allows AO to be toggled on/off
	Enabled bool

	// Intensity controls how strong the darkening effect is (0.0-1.0)
	Intensity float64

	// Radius defines how far the occlusion extends (pixels)
	Radius float64

	// Samples controls quality vs performance (4-16 typical)
	Samples int

	// CornerDarkening adds extra darkening at convex corners
	CornerDarkening bool

	// CornerAmount controls corner darkening intensity (0.0-1.0)
	CornerAmount float64
}

// Type returns the component type identifier.
func (a *AmbientOcclusionComponent) Type() string {
	return "ambient_occlusion"
}

// NewAmbientOcclusionComponent creates an AO component with default values.
func NewAmbientOcclusionComponent(intensity, radius float64) *AmbientOcclusionComponent {
	if intensity <= 0 {
		intensity = 0.3
	}
	if radius <= 0 {
		radius = 32
	}

	return &AmbientOcclusionComponent{
		Enabled:         true,
		Intensity:       intensity,
		Radius:          radius,
		Samples:         8,
		CornerDarkening: true,
		CornerAmount:    0.5,
	}
}
