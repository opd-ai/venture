package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/opd-ai/venture/pkg/procgen/vehicle"
	"github.com/sirupsen/logrus"
)

var (
	genre      = flag.String("genre", "fantasy", "Genre: fantasy, scifi, horror, cyberpunk, or postapoc")
	count      = flag.Int("count", 10, "Number of vehicles to generate")
	depth      = flag.Int("depth", 5, "Depth level (affects stats)")
	difficulty = flag.Float64("difficulty", 0.5, "Difficulty multiplier (0.0-1.0)")
	seed       = flag.Int64("seed", 12345, "Generation seed")
	output     = flag.String("output", "", "Output file (leave empty for console)")
	verbose    = flag.Bool("verbose", false, "Show detailed vehicle information")
)

func main() {
	flag.Parse()

	// Initialize logger
	logger := logging.TestUtilityLogger("vehicletest")
	testLogger := logger.WithFields(logrus.Fields{
		"genre":      *genre,
		"count":      *count,
		"depth":      *depth,
		"difficulty": *difficulty,
		"seed":       *seed,
	})

	testLogger.Info("generating vehicles")

	// Create generator with logger
	gen := vehicle.NewVehicleGeneratorWithLogger(logger)

	// Set up generation parameters
	params := procgen.GenerationParams{
		Difficulty: *difficulty,
		Depth:      *depth,
		GenreID:    *genre,
		Custom: map[string]interface{}{
			"count": *count,
		},
	}

	// Generate vehicles
	genLogger := logging.GeneratorLogger(logger, "vehicle", *seed, *genre)
	genLogger.Debug("starting vehicle generation")

	result, err := gen.Generate(*seed, params)
	if err != nil {
		genLogger.WithError(err).Fatal("generation failed")
	}

	vehicles, ok := result.([]*vehicle.Vehicle)
	if !ok {
		genLogger.Fatal("result is not []*vehicle.Vehicle")
	}

	// Validate
	if err := gen.Validate(vehicles); err != nil {
		genLogger.WithError(err).Fatal("validation failed")
	}

	genLogger.WithField("vehicleCount", len(vehicles)).Info("vehicles generated successfully")

	// Render to string
	rendered := renderVehicles(vehicles, *verbose)

	// Output to file or console
	if *output != "" {
		if err := os.WriteFile(*output, []byte(rendered), 0o644); err != nil {
			testLogger.WithError(err).WithField("outputFile", *output).Fatal("failed to write output file")
		}
		testLogger.WithField("outputFile", *output).Info("vehicles written to file")
	} else {
		fmt.Println(rendered)
	}

	testLogger.Info("vehicle generation test complete")
}

func renderVehicles(vehicles []*vehicle.Vehicle, verbose bool) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Generated %d vehicles:\n\n", len(vehicles)))

	// Group by type
	byType := make(map[vehicle.VehicleType][]*vehicle.Vehicle)
	for _, v := range vehicles {
		byType[v.VehicleType] = append(byType[v.VehicleType], v)
	}

	// Display each type
	types := []vehicle.VehicleType{
		vehicle.TypeMount, vehicle.TypeCart, vehicle.TypeBoat,
		vehicle.TypeGlider, vehicle.TypeMech,
	}

	for _, vType := range types {
		vehiclesOfType := byType[vType]
		if len(vehiclesOfType) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("=== %s (%d) ===\n", vType.String(), len(vehiclesOfType)))

		for _, v := range vehiclesOfType {
			sb.WriteString(formatVehicle(v, verbose))
			sb.WriteString("\n")
		}
	}

	// Summary statistics
	sb.WriteString("\n=== Summary ===\n")
	sb.WriteString(fmt.Sprintf("Total Vehicles: %d\n", len(vehicles)))

	// Count by rarity
	byRarity := make(map[vehicle.Rarity]int)
	for _, v := range vehicles {
		byRarity[v.Rarity]++
	}

	sb.WriteString("By Rarity:\n")
	rarities := []vehicle.Rarity{
		vehicle.RarityCommon, vehicle.RarityUncommon, vehicle.RarityRare,
		vehicle.RarityEpic, vehicle.RarityLegendary,
	}
	for _, r := range rarities {
		if count := byRarity[r]; count > 0 {
			sb.WriteString(fmt.Sprintf("  %s: %d (%.1f%%)\n", r.String(), count, float64(count)/float64(len(vehicles))*100))
		}
	}

	// Average stats
	if len(vehicles) > 0 {
		avgSpeed := 0.0
		avgDurability := 0.0
		for _, v := range vehicles {
			avgSpeed += v.MaxSpeed
			avgDurability += v.MaxDurability
		}
		avgSpeed /= float64(len(vehicles))
		avgDurability /= float64(len(vehicles))

		sb.WriteString(fmt.Sprintf("\nAverage Stats:\n"))
		sb.WriteString(fmt.Sprintf("  Max Speed: %.1f\n", avgSpeed))
		sb.WriteString(fmt.Sprintf("  Durability: %.1f\n", avgDurability))
	}

	return sb.String()
}

func formatVehicle(v *vehicle.Vehicle, verbose bool) string {
	var sb strings.Builder

	// Basic info
	sb.WriteString(fmt.Sprintf("  %s (%s)\n", v.Name, v.Rarity.String()))

	if verbose {
		sb.WriteString(fmt.Sprintf("    Description: %s\n", v.Description))
		sb.WriteString(fmt.Sprintf("    Stats:\n"))
		sb.WriteString(fmt.Sprintf("      Max Speed:      %.1f\n", v.MaxSpeed))
		sb.WriteString(fmt.Sprintf("      Acceleration:   %.1f\n", v.Acceleration))
		sb.WriteString(fmt.Sprintf("      Handling:       %.1f\n", v.Handling))
		sb.WriteString(fmt.Sprintf("      Max Durability: %.1f\n", v.MaxDurability))
		sb.WriteString(fmt.Sprintf("      Fuel Capacity:  %.1f (%s)\n", v.FuelCapacity, v.FuelType))
		sb.WriteString(fmt.Sprintf("      Capacity:       %d passengers\n", v.Capacity))
		sb.WriteString(fmt.Sprintf("      Color:          #%06X\n", v.Color))
		sb.WriteString(fmt.Sprintf("      Depth:          %d\n", v.Depth))
	} else {
		sb.WriteString(fmt.Sprintf("    Speed: %.0f | Durability: %.0f | Fuel: %.0f %s\n",
			v.MaxSpeed, v.MaxDurability, v.FuelCapacity, v.FuelType))
	}

	return sb.String()
}
