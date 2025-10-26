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
	case ShapeEllipse:
		return g.inEllipse(dx, dy, centerX, centerY, config.Smoothing)
	case ShapeCapsule:
		return g.inCapsule(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeBean:
		return g.inBean(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeWedge:
		return g.inWedge(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeShield:
		return g.inShield(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeBlade:
		return g.inBlade(dx, dy, centerX, centerY, config.Rotation, config.Smoothing)
	case ShapeSkull:
		return g.inSkull(dx, dy, centerX, centerY, config.Smoothing)
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

// inEllipse checks if a point is inside an ellipse (oval) shape.
// Ellipse is useful for heads, bodies, and organic shapes with different width/height ratios.
func (g *Generator) inEllipse(dx, dy, centerX, centerY, smoothing float64) bool {
	// Ellipse equation: (x/a)^2 + (y/b)^2 <= 1
	// where a = width/2, b = height/2
	radiusX := centerX
	radiusY := centerY

	// Normalize coordinates to ellipse space
	nx := dx / radiusX
	ny := dy / radiusY

	dist := math.Sqrt(nx*nx + ny*ny)

	// Apply smoothing for anti-aliasing
	edge := 1.0 - smoothing
	if dist < edge {
		return true
	}
	if dist > 1.0 {
		return false
	}
	return (dist-edge)/(1.0-edge) < 0.5
}

// inCapsule checks if a point is inside a capsule (rounded rectangle/pill) shape.
// Capsule is perfect for limbs (arms, legs) as it maintains consistent width with rounded ends.
func (g *Generator) inCapsule(dx, dy, centerX, centerY, rotation, smoothing float64) bool {
	// Rotate point by inverse rotation
	angle := -rotation * math.Pi / 180.0
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := dx*cos - dy*sin
	ry := dx*sin + dy*cos

	// Capsule is a rectangle with semicircular ends
	// Determine if vertical or horizontal based on dimensions
	halfWidth := centerX * 0.3   // 30% of width for capsule width
	halfHeight := centerY * 0.85 // 85% of height for capsule length

	// Check if in main rectangular body
	if math.Abs(rx) <= halfWidth && math.Abs(ry) <= halfHeight {
		return true
	}

	// Check if in top semicircle
	if ry > halfHeight {
		topDist := math.Sqrt(rx*rx + math.Pow(ry-halfHeight, 2))
		return topDist <= halfWidth*(1.0+smoothing)
	}

	// Check if in bottom semicircle
	if ry < -halfHeight {
		bottomDist := math.Sqrt(rx*rx + math.Pow(ry+halfHeight, 2))
		return bottomDist <= halfWidth*(1.0+smoothing)
	}

	return false
}

// inBean checks if a point is inside a bean (kidney bean) shape.
// Bean shape is ideal for torsos, providing natural body curves.
func (g *Generator) inBean(dx, dy, centerX, centerY, rotation, smoothing float64) bool {
	// Rotate point
	angle := -rotation * math.Pi / 180.0
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := dx*cos - dy*sin
	ry := dx*sin + dy*cos

	// Normalize to -1 to 1 range
	nx := rx / centerX
	ny := ry / centerY

	// Bean shape: ellipse with indent on one side
	// Use modified ellipse equation with curvature
	dist := math.Sqrt(nx*nx + ny*ny)

	// Add curvature based on x position (indent on right side)
	curvature := 0.2 * nx * (1.0 - ny*ny) // More indent near center Y
	threshold := 0.9 + curvature

	// Apply smoothing
	edge := threshold - smoothing
	if dist < edge {
		return true
	}
	if dist > threshold {
		return false
	}
	return (dist-edge)/(threshold-edge) < 0.5
}

// inWedge checks if a point is inside a wedge (directional triangle/arrow) shape.
// Wedge is useful for indicating facing direction or arrow shapes.
func (g *Generator) inWedge(dx, dy, centerX, centerY, rotation, smoothing float64) bool {
	// Rotate point
	angle := -rotation * math.Pi / 180.0
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := dx*cos - dy*sin
	ry := dx*sin + dy*cos

	// Normalize
	nx := rx / centerX
	ny := ry / centerY

	// Wedge: isosceles triangle pointing upward
	// Base at bottom, point at top
	// Triangle vertices: (0, -1), (-0.7, 0.5), (0.7, 0.5)

	// Check if below base line
	if ny > 0.5 {
		return false
	}

	// Check if left of left edge: line from (-0.7, 0.5) to (0, -1)
	// Slope: (-1 - 0.5) / (0 - (-0.7)) = -1.5 / 0.7 ≈ -2.14
	leftEdge := -2.14*nx - 1.0
	if ny < leftEdge-smoothing {
		return false
	}

	// Check if right of right edge: line from (0.7, 0.5) to (0, -1)
	// Slope: (-1 - 0.5) / (0 - 0.7) = -1.5 / -0.7 ≈ 2.14
	rightEdge := 2.14*nx - 1.0
	if ny < rightEdge-smoothing {
		return false
	}

	return true
}

// inShield checks if a point is inside a shield shape.
// Shield shape is ideal for defense icons and equipped shields.
func (g *Generator) inShield(dx, dy, centerX, centerY, rotation, smoothing float64) bool {
	// Rotate point
	angle := -rotation * math.Pi / 180.0
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := dx*cos - dy*sin
	ry := dx*sin + dy*cos

	// Normalize
	nx := rx / centerX
	ny := ry / centerY

	// Shield: rounded top, pointed bottom
	// Top half: circle/ellipse
	if ny < 0 {
		// Upper shield: ellipse
		dist := math.Sqrt(nx*nx + ny*ny*1.5*1.5)
		if dist <= 0.8+smoothing {
			return true
		}
	} else {
		// Lower shield: converging to point at bottom
		// Check if inside triangle from (-0.8, 0) to (0.8, 0) to (0, 1.2)
		if ny > 1.2 {
			return false
		}

		// Left edge
		leftEdge := -0.67*nx + 0.0 // Line from (-0.8, 0) to (0, 1.2)
		if ny < leftEdge-smoothing {
			return false
		}

		// Right edge
		rightEdge := 0.67*nx + 0.0
		if ny < rightEdge-smoothing {
			return false
		}

		// Width taper: shield narrows toward bottom
		maxWidth := 0.8 * (1.0 - ny/1.2)
		if math.Abs(nx) > maxWidth+smoothing {
			return false
		}

		return true
	}

	return false
}

// inBlade checks if a point is inside a blade (sword) shape.
// Blade shape is perfect for weapon sprites.
func (g *Generator) inBlade(dx, dy, centerX, centerY, rotation, smoothing float64) bool {
	// Rotate point
	angle := -rotation * math.Pi / 180.0
	cos := math.Cos(angle)
	sin := math.Sin(angle)
	rx := dx*cos - dy*sin
	ry := dx*sin + dy*cos

	// Normalize
	nx := rx / centerX
	ny := ry / centerY

	// Blade: thin rectangle with tapered point
	// Blade body: 70% of length, full width
	// Blade tip: 30% of length, tapers to point

	bladeWidth := 0.15 // Thin blade
	bladeStart := -0.9
	bladeEnd := 0.5
	tipEnd := 1.0

	// Check hilt/handle (bottom)
	if ny < bladeStart {
		// Hilt: slightly wider
		if math.Abs(nx) <= 0.25+smoothing {
			return true
		}
	} else if ny <= bladeEnd {
		// Main blade body
		if math.Abs(nx) <= bladeWidth+smoothing {
			return true
		}
	} else if ny <= tipEnd {
		// Tapered tip
		progress := (ny - bladeEnd) / (tipEnd - bladeEnd)
		taperedWidth := bladeWidth * (1.0 - progress)
		if math.Abs(nx) <= taperedWidth+smoothing {
			return true
		}
	}

	return false
}

// inSkull checks if a point is inside a skull shape.
// Skull shape is useful for head/face detail and undead entities.
func (g *Generator) inSkull(dx, dy, centerX, centerY, smoothing float64) bool {
	// Normalize
	nx := dx / centerX
	ny := dy / centerY

	// Skull: rounded cranium + jaw
	// Upper skull: circle
	crownRadius := 0.7
	crownCenterY := -0.3
	crownDist := math.Sqrt(nx*nx + math.Pow(ny-crownCenterY, 2))
	if crownDist <= crownRadius+smoothing {
		// Check if in eye sockets (negative space)
		leftEyeX := -0.3
		rightEyeX := 0.3
		eyeY := -0.2
		eyeRadius := 0.15

		leftEyeDist := math.Sqrt(math.Pow(nx-leftEyeX, 2) + math.Pow(ny-eyeY, 2))
		rightEyeDist := math.Sqrt(math.Pow(nx-rightEyeX, 2) + math.Pow(ny-eyeY, 2))

		// Exclude eye sockets
		if leftEyeDist < eyeRadius || rightEyeDist < eyeRadius {
			return false
		}

		return true
	}

	// Lower jaw: trapezoid shape
	if ny > 0.2 && ny < 0.7 {
		// Jaw narrows toward bottom
		jawWidth := 0.5 - (ny-0.2)*0.4
		if math.Abs(nx) <= jawWidth+smoothing {
			return true
		}
	}

	return false
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
