// Command transitiontest is a CLI tool for testing Phase 16.2 tile transitions.
// It generates tiles with different transition types to verify the auto-tiling
// and blending system.
package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
	"github.com/sirupsen/logrus"
)

var (
	tileType     = flag.String("type", "wall", "Tile type (floor, wall, door, etc.)")
	width        = flag.Int("width", 32, "Tile width in pixels")
	height       = flag.Int("height", 32, "Tile height in pixels")
	genre        = flag.String("genre", "fantasy", "Genre ID")
	seed         = flag.Int64("seed", 12345, "Random seed")
	transition   = flag.String("transition", "all", "Transition type (all, none, full, n, e, s, w, ne, nw, se, sw, ns, ew, nes, new, nsw, esw, inner_ne, inner_nw, inner_se, inner_sw)")
	blendRadius  = flag.Float64("blend", 0.3, "Blend radius (0.0-1.0)")
	cornerRadius = flag.Float64("corner", 0.25, "Corner radius (0.0-1.0)")
	smoothness   = flag.Float64("smooth", 0.5, "Edge smoothness (0.0-1.0)")
	output       = flag.String("output", "transitions.png", "Output file")
	verbose      = flag.Bool("verbose", false, "Show verbose output")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("transitiontest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}

	logger.Info("Transition Test Tool started")

	// Parse tile type
	var tileTypeEnum tiles.TileType
	switch *tileType {
	case "floor":
		tileTypeEnum = tiles.TileFloor
	case "wall":
		tileTypeEnum = tiles.TileWall
	case "door":
		tileTypeEnum = tiles.TileDoor
	case "corridor":
		tileTypeEnum = tiles.TileCorridor
	case "water":
		tileTypeEnum = tiles.TileWater
	case "lava":
		tileTypeEnum = tiles.TileLava
	case "trap":
		tileTypeEnum = tiles.TileTrap
	case "stairs":
		tileTypeEnum = tiles.TileStairs
	default:
		logger.WithField("type", *tileType).Fatal("unknown tile type")
	}

	// Create generator
	gen := tiles.NewGeneratorWithLogger(logger)

	fmt.Printf("=== Tile Transition Test Tool ===\n\n")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Type:         %s\n", *tileType)
	fmt.Printf("  Size:         %dx%d\n", *width, *height)
	fmt.Printf("  Genre:        %s\n", *genre)
	fmt.Printf("  Seed:         %d\n", *seed)
	fmt.Printf("  Transition:   %s\n", *transition)
	fmt.Printf("  Blend Radius: %.2f\n", *blendRadius)
	fmt.Printf("  Corner Rad:   %.2f\n", *cornerRadius)
	fmt.Printf("  Smoothness:   %.2f\n\n", *smoothness)

	// Determine which transitions to generate
	var transitionsToTest []struct {
		name       string
		transition tiles.TileTransitionType
		neighbors  tiles.TileNeighbors
	}

	if *transition == "all" {
		// Generate all transition types in a grid
		transitionsToTest = []struct {
			name       string
			transition tiles.TileTransitionType
			neighbors  tiles.TileNeighbors
		}{
			{"None", tiles.TransitionNone, tiles.TileNeighbors{}},
			{"Full", tiles.TransitionFull, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true, SW: true}},
			{"North", tiles.TransitionN, tiles.TileNeighbors{N: true}},
			{"East", tiles.TransitionE, tiles.TileNeighbors{E: true}},
			{"South", tiles.TransitionS, tiles.TileNeighbors{S: true}},
			{"West", tiles.TransitionW, tiles.TileNeighbors{W: true}},
			{"Northeast", tiles.TransitionNE, tiles.TileNeighbors{N: true, E: true}},
			{"Northwest", tiles.TransitionNW, tiles.TileNeighbors{N: true, W: true}},
			{"Southeast", tiles.TransitionSE, tiles.TileNeighbors{S: true, E: true}},
			{"Southwest", tiles.TransitionSW, tiles.TileNeighbors{S: true, W: true}},
			{"N-S Corridor", tiles.TransitionNS, tiles.TileNeighbors{N: true, S: true}},
			{"E-W Corridor", tiles.TransitionEW, tiles.TileNeighbors{E: true, W: true}},
			{"T-junction NES", tiles.TransitionNES, tiles.TileNeighbors{N: true, E: true, S: true}},
			{"T-junction NEW", tiles.TransitionNEW, tiles.TileNeighbors{N: true, E: true, W: true}},
			{"T-junction NSW", tiles.TransitionNSW, tiles.TileNeighbors{N: true, S: true, W: true}},
			{"T-junction ESW", tiles.TransitionESW, tiles.TileNeighbors{E: true, S: true, W: true}},
			{"Inner NE", tiles.TransitionInnerNE, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NW: true, SE: true, SW: true}},
			{"Inner NW", tiles.TransitionInnerNW, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, SE: true, SW: true}},
			{"Inner SE", tiles.TransitionInnerSE, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SW: true}},
			{"Inner SW", tiles.TransitionInnerSW, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true}},
		}
	} else {
		// Single transition type
		var transType tiles.TileTransitionType
		var neighbors tiles.TileNeighbors
		
		switch *transition {
		case "none":
			transType = tiles.TransitionNone
		case "full":
			transType = tiles.TransitionFull
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true, SW: true}
		case "n":
			transType = tiles.TransitionN
			neighbors = tiles.TileNeighbors{N: true}
		case "e":
			transType = tiles.TransitionE
			neighbors = tiles.TileNeighbors{E: true}
		case "s":
			transType = tiles.TransitionS
			neighbors = tiles.TileNeighbors{S: true}
		case "w":
			transType = tiles.TransitionW
			neighbors = tiles.TileNeighbors{W: true}
		case "ne":
			transType = tiles.TransitionNE
			neighbors = tiles.TileNeighbors{N: true, E: true}
		case "nw":
			transType = tiles.TransitionNW
			neighbors = tiles.TileNeighbors{N: true, W: true}
		case "se":
			transType = tiles.TransitionSE
			neighbors = tiles.TileNeighbors{S: true, E: true}
		case "sw":
			transType = tiles.TransitionSW
			neighbors = tiles.TileNeighbors{S: true, W: true}
		case "ns":
			transType = tiles.TransitionNS
			neighbors = tiles.TileNeighbors{N: true, S: true}
		case "ew":
			transType = tiles.TransitionEW
			neighbors = tiles.TileNeighbors{E: true, W: true}
		case "nes":
			transType = tiles.TransitionNES
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true}
		case "new":
			transType = tiles.TransitionNEW
			neighbors = tiles.TileNeighbors{N: true, E: true, W: true}
		case "nsw":
			transType = tiles.TransitionNSW
			neighbors = tiles.TileNeighbors{N: true, S: true, W: true}
		case "esw":
			transType = tiles.TransitionESW
			neighbors = tiles.TileNeighbors{E: true, S: true, W: true}
		case "inner_ne":
			transType = tiles.TransitionInnerNE
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true, W: true, NW: true, SE: true, SW: true}
		case "inner_nw":
			transType = tiles.TransitionInnerNW
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, SE: true, SW: true}
		case "inner_se":
			transType = tiles.TransitionInnerSE
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SW: true}
		case "inner_sw":
			transType = tiles.TransitionInnerSW
			neighbors = tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true}
		default:
			logger.WithField("transition", *transition).Fatal("unknown transition type")
		}
		
		transitionsToTest = []struct {
			name       string
			transition tiles.TileTransitionType
			neighbors  tiles.TileNeighbors
		}{
			{*transition, transType, neighbors},
		}
	}

	// Create composite image
	cols := 5
	if len(transitionsToTest) < 5 {
		cols = len(transitionsToTest)
	}
	rows := (len(transitionsToTest) + cols - 1) / cols
	
	compositeWidth := cols * (*width + 4) // 4px spacing between tiles
	compositeHeight := rows * (*height + 24) // 24px spacing for labels
	
	composite := image.NewRGBA(image.Rect(0, 0, compositeWidth, compositeHeight))
	
	// Fill background with dark gray
	for y := 0; y < compositeHeight; y++ {
		for x := 0; x < compositeWidth; x++ {
			composite.Set(x, y, image.Black)
		}
	}

	// Generate each transition
	for i, trans := range transitionsToTest {
		row := i / cols
		col := i % cols
		
		config := tiles.TransitionConfig{
			BaseConfig: tiles.Config{
				Type:    tileTypeEnum,
				Width:   *width,
				Height:  *height,
				GenreID: *genre,
				Seed:    *seed,
				Variant: 0.5,
				Custom:  make(map[string]interface{}),
			},
			Transition:   trans.transition,
			Neighbors:    trans.neighbors,
			BlendRadius:  *blendRadius,
			CornerRadius: *cornerRadius,
			Smoothness:   *smoothness,
		}

		img, err := gen.GenerateWithTransition(config)
		if err != nil {
			logger.WithError(err).WithField("transition", trans.name).Fatal("transition generation failed")
		}

		// Calculate position in composite
		x := col * (*width + 4) + 2
		y := row * (*height + 24) + 20 // Leave space for label
		
		// Draw tile
		draw.Draw(composite, image.Rect(x, y, x+*width, y+*height), img, image.Point{}, draw.Src)

		fmt.Printf("Generated: %-15s (%s)\n", trans.name, trans.transition.String())
	}

	// Save composite image
	f, err := os.Create(*output)
	if err != nil {
		logger.WithError(err).Fatal("failed to create output file")
	}
	defer f.Close()

	err = png.Encode(f, composite)
	if err != nil {
		logger.WithError(err).Fatal("failed to encode PNG")
	}

	fmt.Printf("\nSaved %d transitions to: %s\n", len(transitionsToTest), *output)
	fmt.Printf("Image size: %dx%d\n", compositeWidth, compositeHeight)
	logger.Info("transition generation completed")
}
