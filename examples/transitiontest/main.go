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

	logger := logging.TestUtilityLogger("transitiontest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	logger.Info("Transition Test Tool started")

	tileTypeEnum := parseTileType(*tileType, logger)
	gen := tiles.NewGeneratorWithLogger(logger)

	printConfiguration()

	transitionsToTest := buildTransitionsList(*transition, logger)
	composite := createCompositeImage(transitionsToTest, *width, *height)

	generateTransitions(gen, transitionsToTest, composite, tileTypeEnum, *width, *height)
	saveComposite(composite, *output, len(transitionsToTest), logger)

	logger.Info("transition generation completed")
}

// parseTileType converts tile type string to enum.
func parseTileType(tileType string, logger *logrus.Logger) tiles.TileType {
	switch tileType {
	case "floor":
		return tiles.TileFloor
	case "wall":
		return tiles.TileWall
	case "door":
		return tiles.TileDoor
	case "corridor":
		return tiles.TileCorridor
	case "water":
		return tiles.TileWater
	case "lava":
		return tiles.TileLava
	case "trap":
		return tiles.TileTrap
	case "stairs":
		return tiles.TileStairs
	default:
		logger.WithField("type", tileType).Fatal("unknown tile type")
		return tiles.TileFloor
	}
}

// printConfiguration prints the current configuration.
func printConfiguration() {
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
}

// transitionSpec holds transition configuration.
type transitionSpec struct {
	name       string
	transition tiles.TileTransitionType
	neighbors  tiles.TileNeighbors
}

// buildTransitionsList constructs the list of transitions to test.
func buildTransitionsList(transition string, logger *logrus.Logger) []transitionSpec {
	if transition == "all" {
		return buildAllTransitions()
	}
	return buildSingleTransition(transition, logger)
}

// buildAllTransitions returns all transition types.
func buildAllTransitions() []transitionSpec {
	return []transitionSpec{
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
}

// buildSingleTransition returns a single transition based on name.
func buildSingleTransition(transition string, logger *logrus.Logger) []transitionSpec {
	transMap := map[string]transitionSpec{
		"none":     {"none", tiles.TransitionNone, tiles.TileNeighbors{}},
		"full":     {"full", tiles.TransitionFull, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true, SW: true}},
		"n":        {"n", tiles.TransitionN, tiles.TileNeighbors{N: true}},
		"e":        {"e", tiles.TransitionE, tiles.TileNeighbors{E: true}},
		"s":        {"s", tiles.TransitionS, tiles.TileNeighbors{S: true}},
		"w":        {"w", tiles.TransitionW, tiles.TileNeighbors{W: true}},
		"ne":       {"ne", tiles.TransitionNE, tiles.TileNeighbors{N: true, E: true}},
		"nw":       {"nw", tiles.TransitionNW, tiles.TileNeighbors{N: true, W: true}},
		"se":       {"se", tiles.TransitionSE, tiles.TileNeighbors{S: true, E: true}},
		"sw":       {"sw", tiles.TransitionSW, tiles.TileNeighbors{S: true, W: true}},
		"ns":       {"ns", tiles.TransitionNS, tiles.TileNeighbors{N: true, S: true}},
		"ew":       {"ew", tiles.TransitionEW, tiles.TileNeighbors{E: true, W: true}},
		"nes":      {"nes", tiles.TransitionNES, tiles.TileNeighbors{N: true, E: true, S: true}},
		"new":      {"new", tiles.TransitionNEW, tiles.TileNeighbors{N: true, E: true, W: true}},
		"nsw":      {"nsw", tiles.TransitionNSW, tiles.TileNeighbors{N: true, S: true, W: true}},
		"esw":      {"esw", tiles.TransitionESW, tiles.TileNeighbors{E: true, S: true, W: true}},
		"inner_ne": {"inner_ne", tiles.TransitionInnerNE, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NW: true, SE: true, SW: true}},
		"inner_nw": {"inner_nw", tiles.TransitionInnerNW, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, SE: true, SW: true}},
		"inner_se": {"inner_se", tiles.TransitionInnerSE, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SW: true}},
		"inner_sw": {"inner_sw", tiles.TransitionInnerSW, tiles.TileNeighbors{N: true, E: true, S: true, W: true, NE: true, NW: true, SE: true}},
	}

	if spec, ok := transMap[transition]; ok {
		return []transitionSpec{spec}
	}
	logger.WithField("transition", transition).Fatal("unknown transition type")
	return nil
}

// createCompositeImage creates the composite image for all transitions.
func createCompositeImage(transitions []transitionSpec, width, height int) *image.RGBA {
	cols := 5
	if len(transitions) < 5 {
		cols = len(transitions)
	}
	rows := (len(transitions) + cols - 1) / cols

	compositeWidth := cols * (width + 4)
	compositeHeight := rows * (height + 24)
	composite := image.NewRGBA(image.Rect(0, 0, compositeWidth, compositeHeight))

	for y := 0; y < compositeHeight; y++ {
		for x := 0; x < compositeWidth; x++ {
			composite.Set(x, y, image.Black)
		}
	}
	return composite
}

// generateTransitions generates all transitions and draws them to composite.
func generateTransitions(gen *tiles.Generator, transitions []transitionSpec, composite *image.RGBA, tileType tiles.TileType, width, height int) {
	cols := 5
	if len(transitions) < 5 {
		cols = len(transitions)
	}

	for i, trans := range transitions {
		row := i / cols
		col := i % cols

		config := tiles.TransitionConfig{
			BaseConfig: tiles.Config{
				Type:    tileType,
				Width:   width,
				Height:  height,
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
			fmt.Printf("Error generating %s: %v\n", trans.name, err)
			continue
		}

		x := col*(width+4) + 2
		y := row*(height+24) + 20
		draw.Draw(composite, image.Rect(x, y, x+width, y+height), img, image.Point{}, draw.Src)

		fmt.Printf("Generated: %-15s (%s)\n", trans.name, trans.transition.String())
	}
}

// saveComposite saves the composite image to file.
func saveComposite(composite *image.RGBA, outputPath string, count int, logger *logrus.Logger) {
	f, err := os.Create(outputPath)
	if err != nil {
		logger.WithError(err).Fatal("failed to create output file")
	}
	defer f.Close()

	err = png.Encode(f, composite)
	if err != nil {
		logger.WithError(err).Fatal("failed to encode PNG")
	}

	fmt.Printf("\nSaved %d transitions to: %s\n", count, outputPath)
	fmt.Printf("Image size: %dx%d\n", composite.Bounds().Dx(), composite.Bounds().Dy())
}
