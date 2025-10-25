//go:build !test
// +build !test

// Package shapes provides procedural shape generation.
// This file implements shape generators for geometric primitives
// used in sprite and UI rendering.
package shapes

import (
	"image"
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
	case ShapeHexagon:
		return g.inPolygon(dx, dy, centerX, centerY, 6, config.Rotation, config.Smoothing)
	case ShapeOctagon:
		return g.inPolygon(dx, dy, centerX, centerY, 8, config.Rotation, config.Smoothing)
	case ShapeCross:
		return g.inCross(dx, dy, centerX, centerY, config.Smoothing)
	case ShapeHeart:
		return g.inHeart(dx, dy, centerX, centerY, config.Smoothing)
	case ShapeCrescent:
		return g.inCrescent(dx, dy, centerX, centerY, config.InnerRatio, config.Rotation, config.Smoothing)
	case ShapeGear:
		return g.inGear(dx, dy, centerX, centerY, config.Sides, config.InnerRatio, config.Rotation, config.Smoothing)
	case ShapeCrystal:
		return g.inCrystal(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeLightning:
		return g.inLightning(dx, dy, centerX, centerY, config.Seed, config.Smoothing)
	case ShapeWave:
		return g.inWave(dx, dy, centerX, centerY, config.Seed, config.Smoothing)
	case ShapeSpiral:
		return g.inSpiral(dx, dy, centerX, centerY, config.Seed, config.Smoothing)
	case ShapeOrganic:
		return g.inOrganic(dx, dy, centerX, centerY, config.Seed, config.Smoothing)
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

// inCross checks if a point is inside a cross/plus shape.
func (g *Generator) inCross(dx, dy, cx, cy, smoothing float64) bool {
	halfW := cx * 0.8
	halfH := cy * 0.8
	thickness := math.Min(halfW, halfH) * 0.3

	absDx := math.Abs(dx)
	absDy := math.Abs(dy)

	// Vertical bar
	verticalBar := absDx <= thickness && absDy <= halfH
	// Horizontal bar
	horizontalBar := absDy <= thickness && absDx <= halfW

	return verticalBar || horizontalBar
}

// inHeart checks if a point is inside a heart shape.
func (g *Generator) inHeart(dx, dy, cx, cy, smoothing float64) bool {
	// Normalize coordinates
	x := dx / (cx * 0.8)
	y := -dy / (cy * 0.8) // Flip Y to point heart upward

	// Heart equation: (x^2 + y^2 - 1)^3 - x^2*y^3 = 0
	// Simplified for filled heart
	x2 := x * x
	y2 := y * y

	// Adjust Y offset to center the heart
	y = y - 0.3

	// Heart boundary check
	value := math.Pow(x2+y2-1, 3) - x2*math.Pow(y, 3)

	return value <= 0
}

// inCrescent checks if a point is inside a crescent/moon shape.
func (g *Generator) inCrescent(dx, dy, cx, cy, innerRatio, rotation, smoothing float64) bool {
	// Rotate point
	angle := rotation * math.Pi / 180.0
	rx := dx*math.Cos(angle) - dy*math.Sin(angle)
	ry := dx*math.Sin(angle) + dy*math.Cos(angle)

	// Outer circle
	dist := math.Sqrt(rx*rx + ry*ry)
	outerRadius := math.Min(cx, cy) * 0.8

	// Inner circle (offset to create crescent)
	offsetX := outerRadius * innerRatio
	innerDx := rx - offsetX
	innerDist := math.Sqrt(innerDx*innerDx + ry*ry)
	innerRadius := outerRadius * 0.85

	// Inside outer circle but outside inner circle
	return dist <= outerRadius && innerDist >= innerRadius
}

// inGear checks if a point is inside a gear shape.
func (g *Generator) inGear(dx, dy, cx, cy float64, teeth int, innerRatio, rotation, smoothing float64) bool {
	if teeth < 4 {
		teeth = 4
	}

	angle := math.Atan2(dy, dx)
	dist := math.Sqrt(dx*dx + dy*dy)
	outerRadius := math.Min(cx, cy) * 0.8
	innerRadius := outerRadius * innerRatio

	// Calculate tooth pattern
	angleStep := 2 * math.Pi / float64(teeth)
	rotRad := rotation * math.Pi / 180.0
	normalizedAngle := math.Mod(angle-rotRad+math.Pi*2, angleStep)

	// Tooth profile: square wave
	toothHeight := (outerRadius - innerRadius) * 0.3
	toothWidth := angleStep * 0.4

	var targetRadius float64
	if normalizedAngle < toothWidth {
		// On tooth
		targetRadius = outerRadius
	} else {
		// Between teeth
		targetRadius = outerRadius - toothHeight
	}

	return dist <= targetRadius && dist >= innerRadius*0.5
}

// inCrystal checks if a point is inside a crystalline/gem shape.
func (g *Generator) inCrystal(dx, dy, cx, cy, rotation, smoothing float64) bool {
	// Rotate point
	angle := rotation * math.Pi / 180.0
	rx := dx*math.Cos(angle) - dy*math.Sin(angle)
	ry := dx*math.Sin(angle) + dy*math.Cos(angle)

	radius := math.Min(cx, cy) * 0.8

	// Crystal is a combination of hexagon top and triangle bottom
	if ry < 0 {
		// Bottom half: triangle
		slope := radius / (radius * 0.6)
		leftEdge := -slope * (-ry)
		rightEdge := slope * (-ry)
		return rx > leftEdge && rx < rightEdge && ry > -radius*0.8
	}

	// Top half: narrower hexagon
	hexWidth := radius * (1.0 - ry/(radius*1.2))
	return math.Abs(rx) < hexWidth && ry < radius*0.4
}

// inLightning checks if a point is inside a lightning bolt shape.
func (g *Generator) inLightning(dx, dy, cx, cy float64, seed int64, smoothing float64) bool {
	radius := math.Min(cx, cy) * 0.8

	// Lightning bolt is a zigzag pattern
	// Vertical main bolt with horizontal branches

	// Main vertical bolt
	boltWidth := radius * 0.15
	if math.Abs(dx) < boltWidth && math.Abs(dy) < radius {
		return true
	}

	// Zigzag pattern using seed for determinism
	segments := 4
	segmentHeight := radius * 2 / float64(segments)

	for i := 0; i < segments; i++ {
		segmentY := -radius + float64(i)*segmentHeight
		if dy > segmentY && dy < segmentY+segmentHeight {
			// Zigzag offset based on segment
			zigzagOffset := radius * 0.3 * float64((i%2)*2-1)
			if math.Abs(dx-zigzagOffset) < boltWidth*1.5 {
				return true
			}
		}
	}

	return false
}

// inWave checks if a point is inside a wave shape.
func (g *Generator) inWave(dx, dy, cx, cy float64, seed int64, smoothing float64) bool {
	radius := math.Min(cx, cy) * 0.8

	// Sine wave pattern
	frequency := 2.0
	amplitude := radius * 0.4
	thickness := radius * 0.15

	// Calculate wave Y position
	waveY := amplitude * math.Sin(frequency*dx/radius)

	// Check if point is within thickness of wave
	return math.Abs(dy-waveY) < thickness && math.Abs(dx) < radius
}

// inSpiral checks if a point is inside a spiral shape.
func (g *Generator) inSpiral(dx, dy, cx, cy float64, seed int64, smoothing float64) bool {
	angle := math.Atan2(dy, dx)
	dist := math.Sqrt(dx*dx + dy*dy)
	maxRadius := math.Min(cx, cy) * 0.8

	// Archimedean spiral: r = a + b*theta
	turns := 3.0
	spiralRadius := (angle + math.Pi) / (2 * math.Pi * turns) * maxRadius

	thickness := maxRadius * 0.1

	// Check if point is on spiral path
	diff := math.Abs(dist - spiralRadius)
	return diff < thickness && dist < maxRadius
}

// inOrganic checks if a point is inside an organic blob shape.
func (g *Generator) inOrganic(dx, dy, cx, cy float64, seed int64, smoothing float64) bool {
	angle := math.Atan2(dy, dx)
	dist := math.Sqrt(dx*dx + dy*dy)
	baseRadius := math.Min(cx, cy) * 0.7

	// Use seed to create deterministic noise
	// Simple pseudo-random based on angle
	noise := math.Sin(float64(seed)*0.001+angle*5.0) * 0.3
	noise += math.Sin(float64(seed)*0.002+angle*3.0) * 0.2
	noise += math.Sin(float64(seed)*0.003+angle*7.0) * 0.1

	// Modulate radius with noise
	targetRadius := baseRadius * (1.0 + noise)

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
