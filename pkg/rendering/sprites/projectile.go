// Package sprites provides procedural projectile sprite generation.
// Phase 10.2: Projectile Physics System
package sprites

import (
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// ProjectileType defines the visual type of projectile
type ProjectileType string

const (
	ProjectileArrow    ProjectileType = "arrow"
	ProjectileBolt     ProjectileType = "bolt"
	ProjectileBullet   ProjectileType = "bullet"
	ProjectileMagic    ProjectileType = "magic"
	ProjectileFireball ProjectileType = "fireball"
	ProjectileEnergy   ProjectileType = "energy"
)

// GenerateProjectileSprite creates a simple procedural sprite for a projectile.
// seed: RNG seed for deterministic generation
// projectileType: type of projectile to generate
// genreID: genre for color palette
// size: sprite size (typically 8-16 pixels)
func GenerateProjectileSprite(seed int64, projectileType string, genreID string, size int) *ebiten.Image {
	// Create image
	img := ebiten.NewImage(size, size)
	
	// Get genre-appropriate color palette
	pal := palette.Generate(seed, genreID)
	
	// Generate based on type
	switch ProjectileType(projectileType) {
	case ProjectileArrow:
		drawArrow(img, size, pal.Primary)
	case ProjectileBolt:
		drawBolt(img, size, pal.Primary)
	case ProjectileBullet:
		drawBullet(img, size, pal.Accent)
	case ProjectileMagic:
		drawMagicBolt(img, size, pal.Magic)
	case ProjectileFireball:
		drawFireball(img, size, pal.Highlight)
	case ProjectileEnergy:
		drawEnergyBolt(img, size, pal.Accent)
	default:
		// Default to arrow
		drawArrow(img, size, pal.Primary)
	}
	
	return img
}

// drawArrow draws a simple arrow shape (triangle pointing right)
func drawArrow(img *ebiten.Image, size int, col color.RGBA) {
	// Arrow is a triangle: tip at right, base at left
	// Coordinates: (x, y) with origin at top-left
	
	// Triangle vertices
	tipX := size - 2
	tipY := size / 2
	baseLeftX := 2
	baseTopY := size / 3
	baseBottomY := 2 * size / 3
	
	// Draw filled triangle
	drawFilledTriangle(img, 
		baseLeftX, baseTopY,
		baseLeftX, baseBottomY,
		tipX, tipY,
		col)
	
	// Add darker outline
	outline := darken(col, 0.5)
	drawTriangleOutline(img,
		baseLeftX, baseTopY,
		baseLeftX, baseBottomY,
		tipX, tipY,
		outline)
}

// drawBolt draws a crossbow bolt (longer, thicker arrow)
func drawBolt(img *ebiten.Image, size int, col color.RGBA) {
	// Similar to arrow but with a shaft
	tipX := size - 2
	tipY := size / 2
	shaftWidth := size / 6
	
	// Tip triangle
	drawFilledTriangle(img,
		size/2, tipY-shaftWidth,
		size/2, tipY+shaftWidth,
		tipX, tipY,
		col)
	
	// Shaft rectangle
	drawFilledRect(img, 2, tipY-shaftWidth/2, size/2, shaftWidth, col)
	
	// Fletching (feathers at back)
	fletchCol := lighten(col, 0.3)
	drawFilledTriangle(img,
		2, tipY-shaftWidth*2,
		2, tipY+shaftWidth*2,
		size/4, tipY,
		fletchCol)
}

// drawBullet draws a simple circle (bullet/pellet)
func drawBullet(img *ebiten.Image, size int, col color.RGBA) {
	centerX := size / 2
	centerY := size / 2
	radius := size / 3
	
	drawFilledCircle(img, centerX, centerY, radius, col)
	
	// Add shine effect
	shine := lighten(col, 0.5)
	drawFilledCircle(img, centerX-radius/3, centerY-radius/3, radius/3, shine)
}

// drawMagicBolt draws a magical energy bolt (diamond shape with glow)
func drawMagicBolt(img *ebiten.Image, size int, col color.RGBA) {
	centerX := size / 2
	centerY := size / 2
	halfSize := size / 3
	
	// Draw diamond shape
	drawFilledTriangle(img,
		centerX, centerY-halfSize, // Top
		centerX+halfSize, centerY, // Right
		centerX, centerY+halfSize, // Bottom
		col)
	drawFilledTriangle(img,
		centerX, centerY-halfSize, // Top
		centerX-halfSize, centerY, // Left
		centerX, centerY+halfSize, // Bottom
		col)
	
	// Add glow effect (lighter outer diamond)
	glow := lighten(col, 0.7)
	glowSize := halfSize + 2
	drawTriangleOutline(img,
		centerX, centerY-glowSize,
		centerX+glowSize, centerY,
		centerX, centerY+glowSize,
		glow)
	drawTriangleOutline(img,
		centerX, centerY-glowSize,
		centerX-glowSize, centerY,
		centerX, centerY+glowSize,
		glow)
}

// drawFireball draws a fiery sphere
func drawFireball(img *ebiten.Image, size int, col color.RGBA) {
	centerX := size / 2
	centerY := size / 2
	radius := size / 3
	
	// Outer orange/red glow
	drawFilledCircle(img, centerX, centerY, radius+1, col)
	
	// Inner bright core
	core := lighten(col, 0.8)
	drawFilledCircle(img, centerX, centerY, radius-1, core)
	
	// Bright center
	bright := color.RGBA{255, 255, 200, 255}
	drawFilledCircle(img, centerX, centerY, radius/2, bright)
}

// drawEnergyBolt draws a sci-fi energy bolt (elongated oval with trail)
func drawEnergyBolt(img *ebiten.Image, size int, col color.RGBA) {
	centerY := size / 2
	
	// Main body (elongated oval)
	for x := size / 4; x < 3*size/4; x++ {
		radius := int(float64(size/4) * math.Sin(float64(x-size/4)*math.Pi/float64(size/2)))
		drawFilledCircle(img, x, centerY, radius, col)
	}
	
	// Bright core
	core := lighten(col, 0.8)
	for x := size / 3; x < 2*size/3; x++ {
		radius := int(float64(size/6) * math.Sin(float64(x-size/3)*math.Pi/float64(size/3)))
		drawFilledCircle(img, x, centerY, radius, core)
	}
}

// Helper drawing functions

func drawFilledRect(img *ebiten.Image, x, y, width, height int, col color.RGBA) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	for i := x; i < x+width; i++ {
		for j := y; j < y+height; j++ {
			if i >= 0 && i < w && j >= 0 && j < h {
				img.Set(i, j, col)
			}
		}
	}
}

func drawFilledCircle(img *ebiten.Image, cx, cy, radius int, col color.RGBA) {
	for x := cx - radius; x <= cx+radius; x++ {
		for y := cy - radius; y <= cy+radius; y++ {
			dx := x - cx
			dy := y - cy
			if dx*dx+dy*dy <= radius*radius {
				img.Set(x, y, col)
			}
		}
	}
}

func drawFilledTriangle(img *ebiten.Image, x1, y1, x2, y2, x3, y3 int, col color.RGBA) {
	// Simple scanline triangle fill
	// Sort vertices by y coordinate
	if y1 > y2 {
		x1, x2 = x2, x1
		y1, y2 = y2, y1
	}
	if y1 > y3 {
		x1, x3 = x3, x1
		y1, y3 = y3, y1
	}
	if y2 > y3 {
		x2, x3 = x3, x2
		y2, y3 = y3, y2
	}
	
	// Draw horizontal lines from y1 to y3
	for y := y1; y <= y3; y++ {
		// Calculate x bounds for this scanline
		var xStart, xEnd int
		
		if y < y2 {
			// Upper half
			if y2-y1 != 0 {
				xStart = x1 + (x2-x1)*(y-y1)/(y2-y1)
			} else {
				xStart = x1
			}
			if y3-y1 != 0 {
				xEnd = x1 + (x3-x1)*(y-y1)/(y3-y1)
			} else {
				xEnd = x1
			}
		} else {
			// Lower half
			if y3-y2 != 0 {
				xStart = x2 + (x3-x2)*(y-y2)/(y3-y2)
			} else {
				xStart = x2
			}
			if y3-y1 != 0 {
				xEnd = x1 + (x3-x1)*(y-y1)/(y3-y1)
			} else {
				xEnd = x1
			}
		}
		
		if xStart > xEnd {
			xStart, xEnd = xEnd, xStart
		}
		
		for x := xStart; x <= xEnd; x++ {
			img.Set(x, y, col)
		}
	}
}

func drawTriangleOutline(img *ebiten.Image, x1, y1, x2, y2, x3, y3 int, col color.RGBA) {
	drawLine(img, x1, y1, x2, y2, col)
	drawLine(img, x2, y2, x3, y3, col)
	drawLine(img, x3, y3, x1, y1, col)
}

func drawLine(img *ebiten.Image, x1, y1, x2, y2 int, col color.RGBA) {
	// Bresenham's line algorithm
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy
	
	for {
		img.Set(x1, y1, col)
		if x1 == x2 && y1 == y2 {
			break
		}
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x1 += sx
		}
		if e2 < dx {
			err += dx
			y1 += sy
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func darken(col color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(col.R) * (1.0 - factor)),
		G: uint8(float64(col.G) * (1.0 - factor)),
		B: uint8(float64(col.B) * (1.0 - factor)),
		A: col.A,
	}
}

func lighten(col color.RGBA, factor float64) color.RGBA {
	return color.RGBA{
		R: uint8(math.Min(255, float64(col.R)+(255-float64(col.R))*factor)),
		G: uint8(math.Min(255, float64(col.G)+(255-float64(col.G))*factor)),
		B: uint8(math.Min(255, float64(col.B)+(255-float64(col.B))*factor)),
		A: col.A,
	}
}
