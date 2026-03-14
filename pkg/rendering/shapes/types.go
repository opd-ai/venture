// Package shapes provides shape type definitions.
// This file defines shape types, geometry data, and rendering
// parameters used by the shape generator.
package shapes

import (
	"image/color"
)

// ShapeType represents different geometric primitives.
type ShapeType int

const (
	// ShapeCircle represents a circular shape
	ShapeCircle ShapeType = iota
	// ShapeRectangle represents a rectangular shape
	ShapeRectangle
	// ShapeTriangle represents a triangular shape
	ShapeTriangle
	// ShapePolygon represents a multi-sided polygon
	ShapePolygon
	// ShapeStar represents a star shape
	ShapeStar
	// ShapeRing represents a ring/donut shape
	ShapeRing
	// ShapeHexagon represents a six-sided hexagon
	ShapeHexagon
	// ShapeOctagon represents an eight-sided octagon
	ShapeOctagon
	// ShapeCross represents a cross/plus shape
	ShapeCross
	// ShapeHeart represents a heart shape
	ShapeHeart
	// ShapeCrescent represents a crescent/moon shape
	ShapeCrescent
	// ShapeGear represents a mechanical gear shape
	ShapeGear
	// ShapeCrystal represents a crystalline/gem shape
	ShapeCrystal
	// ShapeLightning represents a lightning bolt shape
	ShapeLightning
	// ShapeWave represents a sine wave shape
	ShapeWave
	// ShapeSpiral represents a spiral/vortex shape
	ShapeSpiral
	// ShapeOrganic represents an organic blob shape
	ShapeOrganic
	// ShapeEllipse represents an oval/ellipse shape (Phase 5.1)
	ShapeEllipse
	// ShapeCapsule represents a rounded rectangle/pill shape (Phase 5.1)
	ShapeCapsule
	// ShapeBean represents a kidney bean/organic body shape (Phase 5.1)
	ShapeBean
	// ShapeWedge represents a directional triangle/arrow (Phase 5.1)
	ShapeWedge
	// ShapeShield represents a shield/defense icon (Phase 5.1)
	ShapeShield
	// ShapeBlade represents a sword/blade shape (Phase 5.1)
	ShapeBlade
	// ShapeSkull represents a skull/head shape (Phase 5.1)
	ShapeSkull
	// ShapeFootprint represents humanoid feet visible from above (Phase 45)
	ShapeFootprint
	// ShapeShoulders represents a shoulder/torso silhouette (Phase 45)
	ShapeShoulders
	// ShapeArmReach represents extended arm positions from above (Phase 45)
	ShapeArmReach
)

// String returns the string representation of a shape type.
func (s ShapeType) String() string {
	switch s {
	case ShapeCircle:
		return "circle"
	case ShapeRectangle:
		return "rectangle"
	case ShapeTriangle:
		return "triangle"
	case ShapePolygon:
		return "polygon"
	case ShapeStar:
		return "star"
	case ShapeRing:
		return "ring"
	case ShapeHexagon:
		return "hexagon"
	case ShapeOctagon:
		return "octagon"
	case ShapeCross:
		return "cross"
	case ShapeHeart:
		return "heart"
	case ShapeCrescent:
		return "crescent"
	case ShapeGear:
		return "gear"
	case ShapeCrystal:
		return "crystal"
	case ShapeLightning:
		return "lightning"
	case ShapeWave:
		return "wave"
	case ShapeSpiral:
		return "spiral"
	case ShapeOrganic:
		return "organic"
	case ShapeEllipse:
		return "ellipse"
	case ShapeCapsule:
		return "capsule"
	case ShapeBean:
		return "bean"
	case ShapeWedge:
		return "wedge"
	case ShapeShield:
		return "shield"
	case ShapeBlade:
		return "blade"
	case ShapeSkull:
		return "skull"
	case ShapeFootprint:
		return "footprint"
	case ShapeShoulders:
		return "shoulders"
	case ShapeArmReach:
		return "armreach"
	default:
		return "unknown"
	}
}

// Shape represents a procedurally generated geometric shape.
type Shape struct {
	// Deprecated: Use Config.Type instead. This field duplicates Config.Type
	// and will be removed in a future version.
	Type   ShapeType
	Width  int
	Height int
	Color  color.Color
	Seed   int64

	// Shape-specific parameters
	Sides      int     // For polygons and stars
	InnerRatio float64 // For rings and stars (0.0-1.0)
	Rotation   float64 // Rotation angle in degrees (0-360)
	Smoothing  float64 // Edge smoothing factor (0.0-1.0)
}

// AntiAliasQuality represents the quality level for anti-aliasing.
type AntiAliasQuality int

const (
	// AntiAliasOff disables anti-aliasing (hard edges)
	AntiAliasOff AntiAliasQuality = iota
	// AntiAliasLow uses 2x super-sampling for anti-aliasing
	AntiAliasLow
	// AntiAliasMedium uses 4x super-sampling for anti-aliasing
	AntiAliasMedium
	// AntiAliasHigh uses 8x super-sampling for anti-aliasing
	AntiAliasHigh
)

// Config holds configuration for shape generation.
type Config struct {
	Type       ShapeType
	Width      int
	Height     int
	Color      color.Color
	Seed       int64
	Sides      int
	InnerRatio float64
	Rotation   float64
	Smoothing  float64
	// AntiAlias enables sub-pixel anti-aliasing for smooth edges (Phase 15.1)
	AntiAlias AntiAliasQuality
	// Quality controls the rendering quality level for performance tuning.
	// When true, uses higher quality anti-aliasing (AntiAliasMedium).
	// When false, uses default anti-aliasing level (AntiAliasLow).
	Quality bool
}

// DefaultConfig returns a default shape configuration.
func DefaultConfig() Config {
	return Config{
		Type:       ShapeCircle,
		Width:      64,
		Height:     64,
		Color:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Seed:       0,
		Sides:      5,
		InnerRatio: 0.5,
		Rotation:   0,
		Smoothing:  0.1,
		AntiAlias:  AntiAliasLow,
		Quality:    false,
	}
}
