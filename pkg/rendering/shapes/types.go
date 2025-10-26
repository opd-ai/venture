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
	default:
		return "unknown"
	}
}

// Shape represents a procedurally generated geometric shape.
type Shape struct {
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
}

// DefaultConfig returns a default shape configuration.
func DefaultConfig() Config {
	return Config{
		Type:       ShapeCircle,
		Width:      32,
		Height:     32,
		Color:      color.RGBA{R: 255, G: 255, B: 255, A: 255},
		Seed:       0,
		Sides:      5,
		InnerRatio: 0.5,
		Rotation:   0,
		Smoothing:  0.1,
	}
}
