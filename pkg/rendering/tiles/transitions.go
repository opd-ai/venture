// Package tiles provides Phase 16.2 smooth terrain transitions.
// Phase 16.2 introduced Marching Squares auto-tiling for seamless tile connections.
// This file implements auto-tiling using Marching Squares algorithm,
// gradient blending between floor types, edge softening, and corner rounding.
package tiles

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/opd-ai/venture/pkg/rendering/palette"
)

// TileNeighbors represents the 8-directional neighbors of a tile.
// Used for Marching Squares algorithm and transition detection.
type TileNeighbors struct {
	N  bool // North
	NE bool // Northeast
	E  bool // East
	SE bool // Southeast
	S  bool // South
	SW bool // Southwest
	W  bool // West
	NW bool // Northwest
}

// TileTransitionType represents the type of transition based on neighbors.
// Marching Squares creates 47 unique tile variants (15 base + 32 variations).
type TileTransitionType int

const (
	// TransitionNone - no neighbors, isolated tile
	TransitionNone TileTransitionType = iota
	// TransitionFull - all neighbors present
	TransitionFull

	// Single edge connections (4 types)
	TransitionN // North connection only
	TransitionE // East connection only
	TransitionS // South connection only
	TransitionW // West connection only

	// Opposite edge connections (2 types)
	TransitionNS // North-South corridor
	TransitionEW // East-West corridor

	// Adjacent edge connections (4 types)
	TransitionNE // North-East corner
	TransitionNW // North-West corner
	TransitionSE // South-East corner
	TransitionSW // South-West corner

	// Three edge connections (4 types)
	TransitionNES // North, East, South (T-junction)
	TransitionNEW // North, East, West (T-junction)
	TransitionNSW // North, South, West (T-junction)
	TransitionESW // East, South, West (T-junction)

	// Inner corner variants (4 types)
	TransitionInnerNE // Inner corner Northeast
	TransitionInnerNW // Inner corner Northwest
	TransitionInnerSE // Inner corner Southeast
	TransitionInnerSW // Inner corner Southwest
)

// String returns the string representation of a transition type.
func (t TileTransitionType) String() string {
	switch t {
	case TransitionNone:
		return "none"
	case TransitionFull:
		return "full"
	case TransitionN:
		return "n"
	case TransitionE:
		return "e"
	case TransitionS:
		return "s"
	case TransitionW:
		return "w"
	case TransitionNS:
		return "ns"
	case TransitionEW:
		return "ew"
	case TransitionNE:
		return "ne"
	case TransitionNW:
		return "nw"
	case TransitionSE:
		return "se"
	case TransitionSW:
		return "sw"
	case TransitionNES:
		return "nes"
	case TransitionNEW:
		return "new"
	case TransitionNSW:
		return "nsw"
	case TransitionESW:
		return "esw"
	case TransitionInnerNE:
		return "inner_ne"
	case TransitionInnerNW:
		return "inner_nw"
	case TransitionInnerSE:
		return "inner_se"
	case TransitionInnerSW:
		return "inner_sw"
	default:
		return "unknown"
	}
}

// DetermineTransition analyzes neighbors and returns the appropriate transition type.
// Implements simplified Marching Squares algorithm for tile edge detection.
func DetermineTransition(neighbors TileNeighbors) TileTransitionType {
	n, e, s, w := neighbors.N, neighbors.E, neighbors.S, neighbors.W

	if hasNoConnections(n, e, s, w) {
		return TransitionNone
	}

	if fullType := checkFullConnection(neighbors, n, e, s, w); fullType != TransitionNone {
		return fullType
	}

	if tjType := checkTJunctions(n, e, s, w); tjType != TransitionNone {
		return tjType
	}

	if corrType := checkCorridors(n, e, s, w); corrType != TransitionNone {
		return corrType
	}

	if cornType := checkCorners(n, e, s, w); cornType != TransitionNone {
		return cornType
	}

	if singleType := checkSingleConnections(n, e, s, w); singleType != TransitionNone {
		return singleType
	}

	return TransitionNone
}

// hasNoConnections checks if there are no cardinal direction connections.
func hasNoConnections(n, e, s, w bool) bool {
	return !n && !e && !s && !w
}

// checkFullConnection checks for all four cardinal connections and inner corners.
func checkFullConnection(neighbors TileNeighbors, n, e, s, w bool) TileTransitionType {
	if !n || !e || !s || !w {
		return TransitionNone
	}

	if !neighbors.NE {
		return TransitionInnerNE
	}
	if !neighbors.NW {
		return TransitionInnerNW
	}
	if !neighbors.SE {
		return TransitionInnerSE
	}
	if !neighbors.SW {
		return TransitionInnerSW
	}
	return TransitionFull
}

// checkTJunctions checks for three-connection T-junction patterns.
func checkTJunctions(n, e, s, w bool) TileTransitionType {
	if n && e && s {
		return TransitionNES
	}
	if n && e && w {
		return TransitionNEW
	}
	if n && s && w {
		return TransitionNSW
	}
	if e && s && w {
		return TransitionESW
	}
	return TransitionNone
}

// checkCorridors checks for two opposite connection corridor patterns.
func checkCorridors(n, e, s, w bool) TileTransitionType {
	if n && s && !e && !w {
		return TransitionNS
	}
	if e && w && !n && !s {
		return TransitionEW
	}
	return TransitionNone
}

// checkCorners checks for two adjacent connection corner patterns.
func checkCorners(n, e, s, w bool) TileTransitionType {
	if n && e {
		return TransitionNE
	}
	if n && w {
		return TransitionNW
	}
	if s && e {
		return TransitionSE
	}
	if s && w {
		return TransitionSW
	}
	return TransitionNone
}

// checkSingleConnections checks for single cardinal direction connections.
func checkSingleConnections(n, e, s, w bool) TileTransitionType {
	if n {
		return TransitionN
	}
	if e {
		return TransitionE
	}
	if s {
		return TransitionS
	}
	if w {
		return TransitionW
	}
	return TransitionNone
}

// TransitionConfig contains parameters for transition tile generation.
type TransitionConfig struct {
	BaseConfig   Config             // Base tile configuration
	Transition   TileTransitionType // Transition type
	Neighbors    TileNeighbors      // Neighbor information
	BlendRadius  float64            // Radius for edge blending (0.0-1.0)
	CornerRadius float64            // Radius for corner rounding (0.0-1.0)
	Smoothness   float64            // Edge smoothness factor (0.0-1.0)
}

// DefaultTransitionConfig returns a default transition configuration.
func DefaultTransitionConfig() TransitionConfig {
	return TransitionConfig{
		BaseConfig:   DefaultConfig(),
		Transition:   TransitionFull,
		Neighbors:    TileNeighbors{},
		BlendRadius:  0.3,
		CornerRadius: 0.25,
		Smoothness:   0.5,
	}
}

// GenerateWithTransition creates a tile with smooth transitions based on neighbors.
func (g *Generator) GenerateWithTransition(config TransitionConfig) (*image.RGBA, error) {
	if err := config.BaseConfig.Validate(); err != nil {
		return nil, err
	}

	// Generate base tile
	img, err := g.Generate(config.BaseConfig)
	if err != nil {
		return nil, err
	}

	// Apply transitions based on type
	rng := rand.New(rand.NewSource(config.BaseConfig.Seed))
	pal, err := g.paletteGen.Generate(config.BaseConfig.GenreID, config.BaseConfig.Seed)
	if err != nil {
		return nil, err
	}

	g.applyTransitions(img, pal, rng, config)

	return img, nil
}

// applyTransitions modifies the tile image to add smooth transitions.
func (g *Generator) applyTransitions(img *image.RGBA, pal *palette.Palette, rng *rand.Rand, config TransitionConfig) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Get transition color (different from base for blending)
	transitionColor := g.pickColor(pal, "floor", rng)

	switch config.Transition {
	case TransitionNone:
		// Isolated tile - round all corners
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, true, true, true, true)

	case TransitionFull:
		// No transitions needed
		return

	case TransitionN:
		// Connection to north only
		g.applyEdgeBlend(img, transitionColor, config, "n")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, false, false, true, true)

	case TransitionE:
		// Connection to east only
		g.applyEdgeBlend(img, transitionColor, config, "e")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, false, true, false, true)

	case TransitionS:
		// Connection to south only
		g.applyEdgeBlend(img, transitionColor, config, "s")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, true, true, false, false)

	case TransitionW:
		// Connection to west only
		g.applyEdgeBlend(img, transitionColor, config, "w")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, true, false, true, false)

	case TransitionNS:
		// North-south corridor
		g.applyEdgeBlend(img, transitionColor, config, "e")
		g.applyEdgeBlend(img, transitionColor, config, "w")

	case TransitionEW:
		// East-west corridor
		g.applyEdgeBlend(img, transitionColor, config, "n")
		g.applyEdgeBlend(img, transitionColor, config, "s")

	case TransitionNE:
		// Northeast corner
		g.applyEdgeBlend(img, transitionColor, config, "s")
		g.applyEdgeBlend(img, transitionColor, config, "w")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, false, false, true, true)

	case TransitionNW:
		// Northwest corner
		g.applyEdgeBlend(img, transitionColor, config, "s")
		g.applyEdgeBlend(img, transitionColor, config, "e")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, false, true, false, true)

	case TransitionSE:
		// Southeast corner
		g.applyEdgeBlend(img, transitionColor, config, "n")
		g.applyEdgeBlend(img, transitionColor, config, "w")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, true, false, true, false)

	case TransitionSW:
		// Southwest corner
		g.applyEdgeBlend(img, transitionColor, config, "n")
		g.applyEdgeBlend(img, transitionColor, config, "e")
		g.applyCornerRounding(img, transitionColor, config.CornerRadius, true, true, false, false)

	case TransitionNES:
		// T-junction (missing west)
		g.applyEdgeBlend(img, transitionColor, config, "w")

	case TransitionNEW:
		// T-junction (missing south)
		g.applyEdgeBlend(img, transitionColor, config, "s")

	case TransitionNSW:
		// T-junction (missing east)
		g.applyEdgeBlend(img, transitionColor, config, "e")

	case TransitionESW:
		// T-junction (missing north)
		g.applyEdgeBlend(img, transitionColor, config, "n")

	case TransitionInnerNE, TransitionInnerNW, TransitionInnerSE, TransitionInnerSW:
		// Inner corners - add subtle concave rounding
		g.applyInnerCorner(img, transitionColor, config)
	}

	// Apply smoothing to all edges
	if config.Smoothness > 0 {
		g.applyEdgeSmoothing(img, config.Smoothness, width, height)
	}
}

// applyEdgeBlend creates a gradient blend at the specified edge.
func (g *Generator) applyEdgeBlend(img *image.RGBA, transitionColor color.Color, config TransitionConfig, edge string) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	blendPixels := int(float64(min(width, height)) * config.BlendRadius)

	if blendPixels < 1 {
		return
	}

	tr, tg, tb, ta := transitionColor.RGBA()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var dist int
			switch edge {
			case "n":
				dist = y
			case "s":
				dist = height - 1 - y
			case "e":
				dist = width - 1 - x
			case "w":
				dist = x
			default:
				continue
			}

			if dist >= blendPixels {
				continue
			}

			// Calculate blend factor (0.0 at edge, 1.0 at blendPixels)
			blend := float64(dist) / float64(blendPixels)
			blend = smoothstep(blend) // Use smoothstep for smooth gradient

			// Get current pixel color
			currentColor := img.At(x, y)
			cr, cg, cb, ca := currentColor.RGBA()

			// Blend colors
			nr := uint8(lerp(float64(tr>>8), float64(cr>>8), blend))
			ng := uint8(lerp(float64(tg>>8), float64(cg>>8), blend))
			nb := uint8(lerp(float64(tb>>8), float64(cb>>8), blend))
			na := uint8(lerp(float64(ta>>8), float64(ca>>8), blend))

			img.Set(x, y, color.RGBA{R: nr, G: ng, B: nb, A: na})
		}
	}
}

// applyCornerRounding rounds corners for organic feel.
// Corners are specified as nw, ne, sw, se (true = round this corner).
func (g *Generator) applyCornerRounding(img *image.RGBA, transitionColor color.Color, radius float64, nw, ne, sw, se bool) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	cornerRadius := int(float64(min(width, height)) * radius)

	if cornerRadius < 1 {
		return
	}

	corners := []struct {
		apply  bool
		cx, cy int
	}{
		{nw, cornerRadius, cornerRadius},                          // Northwest
		{ne, width - 1 - cornerRadius, cornerRadius},              // Northeast
		{sw, cornerRadius, height - 1 - cornerRadius},             // Southwest
		{se, width - 1 - cornerRadius, height - 1 - cornerRadius}, // Southeast
	}

	for _, corner := range corners {
		if !corner.apply {
			continue
		}

		// Round this corner
		for dy := -cornerRadius; dy <= cornerRadius; dy++ {
			for dx := -cornerRadius; dx <= cornerRadius; dx++ {
				x := corner.cx + dx
				y := corner.cy + dy

				if x < 0 || x >= width || y < 0 || y >= height {
					continue
				}

				// Calculate distance from corner center
				dist := math.Sqrt(float64(dx*dx + dy*dy))

				// If outside corner radius, blend toward transition color
				if dist > float64(cornerRadius) {
					// Blend based on distance beyond radius
					excess := dist - float64(cornerRadius)
					blend := math.Min(1.0, excess/float64(cornerRadius))
					blend = smoothstep(blend)

					currentColor := img.At(x, y)
					blendedColor := blendColors(currentColor, transitionColor, blend)
					img.Set(x, y, blendedColor)
				}
			}
		}
	}
}

// applyInnerCorner adds concave rounding for inner corners.
func (g *Generator) applyInnerCorner(img *image.RGBA, accentColor color.Color, config TransitionConfig) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	cornerRadius := int(float64(min(width, height)) * config.CornerRadius)

	if cornerRadius < 1 {
		return
	}

	// Determine which corner based on transition type
	var cx, cy int
	switch config.Transition {
	case TransitionInnerNE:
		cx, cy = width-1, 0
	case TransitionInnerNW:
		cx, cy = 0, 0
	case TransitionInnerSE:
		cx, cy = width-1, height-1
	case TransitionInnerSW:
		cx, cy = 0, height-1
	default:
		return
	}

	// Apply concave corner effect
	for dy := 0; dy <= cornerRadius; dy++ {
		for dx := 0; dx <= cornerRadius; dx++ {
			var x, y int
			switch config.Transition {
			case TransitionInnerNE:
				x, y = cx-dx, cy+dy
			case TransitionInnerNW:
				x, y = cx+dx, cy+dy
			case TransitionInnerSE:
				x, y = cx-dx, cy-dy
			case TransitionInnerSW:
				x, y = cx+dx, cy-dy
			}

			if x < 0 || x >= width || y < 0 || y >= height {
				continue
			}

			// Calculate distance from corner
			dist := math.Sqrt(float64(dx*dx + dy*dy))

			// If inside corner radius, darken slightly for concave effect
			if dist <= float64(cornerRadius) {
				blend := 1.0 - (dist / float64(cornerRadius))
				blend = smoothstep(blend) * 0.3 // Subtle effect

				currentColor := img.At(x, y)
				darkenedColor := blendColors(currentColor, accentColor, blend)
				img.Set(x, y, darkenedColor)
			}
		}
	}
}

// applyEdgeSmoothing applies smoothing to all edges for organic feel.
func (g *Generator) applyEdgeSmoothing(img *image.RGBA, smoothness float64, width, height int) {
	if smoothness <= 0 {
		return
	}

	temp := copyImageRGBA(img, width, height)
	edgeWidth := 1

	for y := edgeWidth; y < height-edgeWidth; y++ {
		for x := edgeWidth; x < width-edgeWidth; x++ {
			if !isEdgeBoundaryPixel(x, y, edgeWidth, width, height) {
				continue
			}

			avgColor := sampleNeighborhood(temp, x, y, width, height)
			if avgColor != nil {
				blendedColor := blendColors(temp.At(x, y), *avgColor, smoothness)
				img.Set(x, y, blendedColor)
			}
		}
	}
}

// copyImageRGBA creates a copy of the image for sampling.
func copyImageRGBA(img *image.RGBA, width, height int) *image.RGBA {
	temp := image.NewRGBA(img.Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			temp.Set(x, y, img.At(x, y))
		}
	}
	return temp
}

// isEdgeBoundaryPixel determines if a pixel is at the edge boundary.
func isEdgeBoundaryPixel(x, y, edgeWidth, width, height int) bool {
	return x == edgeWidth || x == width-1-edgeWidth ||
		y == edgeWidth || y == height-1-edgeWidth
}

// sampleNeighborhood samples a 3x3 neighborhood and returns average color.
func sampleNeighborhood(img *image.RGBA, x, y, width, height int) *color.RGBA {
	var r, g, b, a uint32
	count := 0

	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			nx, ny := x+dx, y+dy
			if nx >= 0 && nx < width && ny >= 0 && ny < height {
				pr, pg, pb, pa := img.At(nx, ny).RGBA()
				r += pr
				g += pg
				b += pb
				a += pa
				count++
			}
		}
	}

	if count == 0 {
		return nil
	}

	return &color.RGBA{
		R: uint8((r / uint32(count)) >> 8),
		G: uint8((g / uint32(count)) >> 8),
		B: uint8((b / uint32(count)) >> 8),
		A: uint8((a / uint32(count)) >> 8),
	}
}

// Note: Helper functions (smoothstep, lerp, blendColors, min, max) have been relocated to utils.go
