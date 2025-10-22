//go:build !test
// +build !test

package shapes

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Generator creates procedural geometric shapes.
type Generator struct{}

// NewGenerator creates a new shape generator.
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a shape image from the configuration.
func (g *Generator) Generate(config Config) (*ebiten.Image, error) {
	img := ebiten.NewImage(config.Width, config.Height)

	// Create shape based on type
	shapeImg := g.generateShape(config)

	// Draw shape to image
	img.DrawImage(shapeImg, nil)

	return img, nil
}

// generateShape creates the shape as an image.
func (g *Generator) generateShape(config Config) *ebiten.Image {
	img := image.NewRGBA(image.Rect(0, 0, config.Width, config.Height))

	centerX := float64(config.Width) / 2.0
	centerY := float64(config.Height) / 2.0

	for y := 0; y < config.Height; y++ {
		for x := 0; x < config.Width; x++ {
			px := float64(x)
			py := float64(y)

			// Calculate distance from center
			dx := px - centerX
			dy := py - centerY

			// Check if pixel is inside shape
			inside := g.isInside(config, dx, dy, centerX, centerY)

			if inside {
				img.Set(x, y, config.Color)
			}
		}
	}

	return ebiten.NewImageFromImage(img)
}

// isInside checks if a point is inside the shape.
func (g *Generator) isInside(config Config, dx, dy, centerX, centerY float64) bool {
	switch config.Type {
	case ShapeCircle:
		return g.inCircle(dx, dy, centerX, centerY, config.Smoothing)
	case ShapeRectangle:
		return g.inRectangle(dx, dy, centerX, centerY, config.Smoothing)
	case ShapeTriangle:
		return g.inTriangle(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapePolygon:
		return g.inPolygon(dx, dy, centerX, centerY, config.Sides, config.Rotation, config.Smoothing)
	case ShapeStar:
		return g.inStar(dx, dy, centerX, centerY, config.Sides, config.InnerRatio, config.Rotation, config.Smoothing)
	case ShapeRing:
		return g.inRing(dx, dy, centerX, centerY, config.InnerRatio, config.Smoothing)
	default:
		return false
	}
}

// inCircle checks if a point is inside a circle.
func (g *Generator) inCircle(dx, dy, cx, cy, smoothing float64) bool {
	dist := math.Sqrt(dx*dx + dy*dy)
	radius := math.Min(cx, cy) * 0.9

	if smoothing == 0 {
		return dist <= radius
	}

	// Smooth edge using smoothstep
	edge := radius * (1.0 - smoothing)
	if dist < edge {
		return true
	}
	if dist > radius {
		return false
	}
	// Smooth transition
	return (dist-edge)/(radius-edge) < 0.5
}

// inRectangle checks if a point is inside a rectangle.
func (g *Generator) inRectangle(dx, dy, cx, cy, smoothing float64) bool {
	halfW := cx * 0.8
	halfH := cy * 0.8

	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	if smoothing == 0 {
		return absDx <= halfW && absDy <= halfH
	}

	// Smooth corners
	edgeW := halfW * (1.0 - smoothing)
	edgeH := halfH * (1.0 - smoothing)

	if absDx < edgeW && absDy < edgeH {
		return true
	}
	if absDx > halfW || absDy > halfH {
		return false
	}

	// Smooth transition at corners
	cornerDx := math.Max(0, absDx-edgeW)
	cornerDy := math.Max(0, absDy-edgeH)
	cornerDist := math.Sqrt(cornerDx*cornerDx + cornerDy*cornerDy)
	cornerRadius := math.Min(halfW-edgeW, halfH-edgeH)

	return cornerDist < cornerRadius*0.5
}

// inTriangle checks if a point is inside a triangle.
func (g *Generator) inTriangle(dx, dy, cx, cy, rotation, smoothing float64) bool {
	// Rotate point
	angle := rotation * math.Pi / 180.0
	rx := dx*math.Cos(angle) - dy*math.Sin(angle)
	ry := dx*math.Sin(angle) + dy*math.Cos(angle)

	// Triangle pointing up
	radius := math.Min(cx, cy) * 0.8
	height := radius * math.Sqrt(3) / 2

	// Simple triangle check
	if ry > height*0.3 {
		return false
	}
	if ry < -height*0.7 {
		return false
	}

	// Check if inside triangle bounds
	slope := height / radius
	leftEdge := -slope * (ry + height*0.7)
	rightEdge := slope * (ry + height*0.7)

	return rx > leftEdge && rx < rightEdge
}

// inPolygon checks if a point is inside a regular polygon.
func (g *Generator) inPolygon(dx, dy, cx, cy float64, sides int, rotation, smoothing float64) bool {
	if sides < 3 {
		sides = 3
	}

	angle := math.Atan2(dy, dx)
	dist := math.Sqrt(dx*dx + dy*dy)
	radius := math.Min(cx, cy) * 0.8

	// Calculate polygon edge distance
	angleStep := 2 * math.Pi / float64(sides)
	rotRad := rotation * math.Pi / 180.0

	// Find closest edge
	normalizedAngle := math.Mod(angle-rotRad+math.Pi*2, angleStep)
	edgeDist := radius * math.Cos(angleStep/2) / math.Cos(normalizedAngle-angleStep/2)

	if smoothing == 0 {
		return dist <= edgeDist
	}

	// Smooth edge
	edge := edgeDist * (1.0 - smoothing)
	if dist < edge {
		return true
	}
	if dist > edgeDist {
		return false
	}
	return (dist-edge)/(edgeDist-edge) < 0.5
}

// inStar checks if a point is inside a star shape.
func (g *Generator) inStar(dx, dy, cx, cy float64, points int, innerRatio, rotation, smoothing float64) bool {
	if points < 3 {
		points = 3
	}

	angle := math.Atan2(dy, dx)
	dist := math.Sqrt(dx*dx + dy*dy)
	outerRadius := math.Min(cx, cy) * 0.8
	innerRadius := outerRadius * innerRatio

	// Calculate star point
	angleStep := math.Pi / float64(points)
	rotRad := rotation * math.Pi / 180.0
	normalizedAngle := math.Mod(angle-rotRad+math.Pi*2, 2*angleStep)

	var targetRadius float64
	if normalizedAngle < angleStep {
		// Outer point
		t := normalizedAngle / angleStep
		targetRadius = outerRadius * (1.0 - 0.5*t*t)
	} else {
		// Inner point
		t := (normalizedAngle - angleStep) / angleStep
		targetRadius = innerRadius + (outerRadius-innerRadius)*t*t
	}

	if smoothing == 0 {
		return dist <= targetRadius
	}

	// Smooth edge
	edge := targetRadius * (1.0 - smoothing)
	if dist < edge {
		return true
	}
	if dist > targetRadius {
		return false
	}
	return (dist-edge)/(targetRadius-edge) < 0.5
}

// inRing checks if a point is inside a ring/donut shape.
func (g *Generator) inRing(dx, dy, cx, cy, innerRatio, smoothing float64) bool {
	dist := math.Sqrt(dx*dx + dy*dy)
	outerRadius := math.Min(cx, cy) * 0.9
	innerRadius := outerRadius * innerRatio

	if smoothing == 0 {
		return dist >= innerRadius && dist <= outerRadius
	}

	// Smooth both edges
	outerEdge := outerRadius * (1.0 - smoothing)
	innerEdge := innerRadius * (1.0 + smoothing)

	if dist < innerEdge || dist > outerRadius {
		return false
	}
	if dist > innerRadius && dist < outerEdge {
		return true
	}

	// Smooth transition
	if dist < innerRadius {
		return (innerRadius-dist)/(innerRadius-innerEdge) < 0.5
	}
	return (dist-outerEdge)/(outerRadius-outerEdge) < 0.5
}
