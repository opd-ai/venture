package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen/environment"
	"github.com/sirupsen/logrus"
)

var (
	genre      = flag.String("genre", "fantasy", "Genre: fantasy, scifi, horror, cyberpunk, postapoc")
	roomWidth  = flag.Int("width", 10, "Room width in tiles")
	roomHeight = flag.Int("height", 10, "Room height in tiles")
	density    = flag.Float64("density", 0.5, "Decoration density (0.0-1.0)")
	spacing    = flag.Int("spacing", 2, "Minimum spacing between decorations")
	seed       = flag.Int64("seed", 12345, "Generation seed")
	output     = flag.String("output", "decorations.txt", "Output file (use .png for images)")
	subtype    = flag.String("subtype", "", "Specific subtype to test (empty for placement demo)")
	verbose    = flag.Bool("verbose", false, "Show detailed information")
)

var subtypeMap = map[string]environment.SubType{
	"sconce":     environment.SubTypeSconce,
	"wallcrack":  environment.SubTypeWallCrack,
	"bloodstain": environment.SubTypeBloodstain,
	"grass":      environment.SubTypeGrass,
	"mushroom":   environment.SubTypeMushroom,
	"skull":      environment.SubTypeSkull,
	"chain":      environment.SubTypeChain,
	"web":        environment.SubTypeWeb,
	"moss":       environment.SubTypeMoss,
	"graffiti":   environment.SubTypeGraffiti,
}

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("decortest")
	testLogger := logger.WithFields(logrus.Fields{
		"genre":      *genre,
		"roomSize":   fmt.Sprintf("%dx%d", *roomWidth, *roomHeight),
		"density":    *density,
		"spacing":    *spacing,
		"seed":       *seed,
		"outputFile": *output,
	})

	testLogger.Info("starting decoration test")

	if *subtype != "" {
		// Test single decoration type
		testSingleDecoration(logger, testLogger)
	} else {
		// Test room placement
		testRoomPlacement(logger, testLogger)
	}
}

func testSingleDecoration(logger *logrus.Logger, testLogger *logrus.Entry) {
	st, ok := subtypeMap[*subtype]
	if !ok {
		testLogger.WithField("subtype", *subtype).Fatal("unknown subtype")
	}

	testLogger.WithField("subtype", st.String()).Info("generating single decoration")

	gen := environment.NewGeneratorWithLogger(logger)

	config := environment.Config{
		SubType: st,
		Width:   64,
		Height:  64,
		GenreID: *genre,
		Seed:    *seed,
	}

	obj, err := gen.Generate(config)
	if err != nil {
		testLogger.WithError(err).Fatal("generation failed")
	}

	testLogger.WithFields(logrus.Fields{
		"name":       obj.Name,
		"type":       obj.Type.String(),
		"size":       fmt.Sprintf("%dx%d", obj.Width, obj.Height),
		"collidable": obj.Collidable,
		"harmful":    obj.Harmful,
	}).Info("decoration generated")

	// Save sprite if output is PNG
	if filepath.Ext(*output) == ".png" {
		f, err := os.Create(*output)
		if err != nil {
			testLogger.WithError(err).Fatal("failed to create output file")
		}
		defer f.Close()

		if err := png.Encode(f, obj.Sprite); err != nil {
			testLogger.WithError(err).Fatal("failed to encode PNG")
		}

		testLogger.WithField("file", *output).Info("saved decoration sprite")
	} else {
		// Text output
		text := fmt.Sprintf("Decoration: %s\n", obj.Name)
		text += fmt.Sprintf("Type: %s (%s)\n", obj.SubType.String(), obj.Type.String())
		text += fmt.Sprintf("Size: %dx%d pixels\n", obj.Width, obj.Height)
		text += fmt.Sprintf("Collidable: %v\n", obj.Collidable)
		text += fmt.Sprintf("Interactable: %v\n", obj.Interactable)
		text += fmt.Sprintf("Harmful: %v (Damage: %d)\n", obj.Harmful, obj.Damage)
		text += fmt.Sprintf("Genre: %s\n", obj.GenreID)
		text += fmt.Sprintf("Seed: %d\n", obj.Seed)

		if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
			testLogger.WithError(err).Fatal("failed to write output file")
		}

		testLogger.WithField("file", *output).Info("saved decoration info")
		fmt.Println(text)
	}
}

func testRoomPlacement(logger *logrus.Logger, testLogger *logrus.Entry) {
	testLogger.Info("testing room placement")

	placer := environment.NewPlacerWithLogger(logger)

	config := environment.PlacementConfig{
		RoomWidth:             *roomWidth,
		RoomHeight:            *roomHeight,
		Density:               *density,
		GenreID:               *genre,
		Seed:                  *seed,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            *spacing,
	}

	placements, err := placer.PlaceDecorations(config)
	if err != nil {
		testLogger.WithError(err).Fatal("placement failed")
	}

	testLogger.WithField("count", len(placements)).Info("decorations placed")

	// Generate output
	text := fmt.Sprintf("Room Decorations (%dx%d)\n", *roomWidth, *roomHeight)
	text += fmt.Sprintf("Genre: %s, Density: %.2f, Spacing: %d\n", *genre, *density, *spacing)
	text += fmt.Sprintf("Seed: %d\n", *seed)
	text += fmt.Sprintf("Total Decorations: %d\n\n", len(placements))

	// Group by placement type
	byType := make(map[environment.PlacementType][]*environment.PlacedObject)
	for _, p := range placements {
		byType[p.PlacementType] = append(byType[p.PlacementType], p)
	}

	for pType, items := range byType {
		text += fmt.Sprintf("\n%s Placements (%d):\n", pType.String(), len(items))
		for _, p := range items {
			if *verbose {
				text += fmt.Sprintf("  - %s at (%d,%d) [Rotation: %.0f°, Scale: %.2f]\n",
					p.Object.Name, p.X, p.Y, p.Variation.Rotation, p.Variation.Scale)
			} else {
				text += fmt.Sprintf("  - %s at (%d,%d)\n", p.Object.Name, p.X, p.Y)
			}
		}
	}

	// Statistics
	text += fmt.Sprintf("\nStatistics:\n")
	text += fmt.Sprintf("  Room Area: %d tiles\n", *roomWidth**roomHeight)
	text += fmt.Sprintf("  Decorations per 100 tiles: %.1f\n",
		float64(len(placements))*100.0/float64(*roomWidth**roomHeight))

	uniqueTypes := make(map[environment.SubType]int)
	for _, p := range placements {
		uniqueTypes[p.Object.SubType]++
	}
	text += fmt.Sprintf("  Unique decoration types: %d\n", len(uniqueTypes))

	if *verbose {
		text += "\nType Distribution:\n"
		for subType, count := range uniqueTypes {
			text += fmt.Sprintf("  - %s: %d (%.1f%%)\n",
				subType.String(), count,
				float64(count)*100.0/float64(len(placements)))
		}
	}

	// Write output
	if err := os.WriteFile(*output, []byte(text), 0o644); err != nil {
		testLogger.WithError(err).Fatal("failed to write output file")
	}

	testLogger.WithField("file", *output).Info("saved placement results")
	fmt.Println(text)
}
