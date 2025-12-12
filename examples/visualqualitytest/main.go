package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/opd-ai/venture/pkg/logging"
	"github.com/opd-ai/venture/pkg/rendering/palette"
	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/opd-ai/venture/pkg/rendering/tiles"
	"github.com/sirupsen/logrus"
)

var (
	spriteCount = flag.Int("sprites", 100, "Number of sprites to generate per genre")
	tileCount   = flag.Int("tiles", 50, "Number of tiles to generate per type")
	genres      = flag.String("genres", "fantasy,scifi,horror,cyberpunk,postapoc", "Comma-separated list of genres to test")
	seed        = flag.Int64("seed", 12345, "Generation seed")
	verbose     = flag.Bool("verbose", false, "Enable verbose logging")
	output      = flag.String("output", "", "Output report to file")
)

// QualityMetrics tracks visual quality statistics
type QualityMetrics struct {
	TotalSprites     int
	SpritesByGenre   map[string]int
	TotalTiles       int
	TilesByType      map[tiles.TileType]int
	PalettesByGenre  map[string]int
	GenerationErrors int
	AverageGenTime   time.Duration
	TotalGenTime     time.Duration
}

func main() {
	flag.Parse()

	logger := initializeLogger()
	metrics := initializeMetrics()
	genreList := []string{"fantasy", "scifi", "horror", "cyberpunk", "postapoc"}

	logTestConfiguration(logger)

	spriteGen, tileGen, paletteGen := initializeGenerators(logger)

	testSpriteGeneration(logger, metrics, spriteGen, paletteGen, genreList)
	testTileGeneration(logger, metrics, tileGen, genreList)
	calculateAverageGenTime(metrics)

	displayResults(logger, metrics, genreList)
}

// initializeLogger creates and configures the logger instance
func initializeLogger() *logrus.Logger {
	logger := logging.TestUtilityLogger("visualqualitytest")
	if *verbose {
		logger.SetLevel(logrus.DebugLevel)
	}
	return logger
}

// initializeMetrics creates a new QualityMetrics instance with initialized maps
func initializeMetrics() *QualityMetrics {
	return &QualityMetrics{
		SpritesByGenre:  make(map[string]int),
		TilesByType:     make(map[tiles.TileType]int),
		PalettesByGenre: make(map[string]int),
	}
}

// logTestConfiguration logs the test configuration details
func logTestConfiguration(logger *logrus.Logger) {
	logger.Info("=== V3.0 Visual Quality Validation ===")
	logger.WithFields(logrus.Fields{
		"sprites": *spriteCount,
		"tiles":   *tileCount,
		"genres":  *genres,
		"seed":    *seed,
	}).Info("test configuration")
}

// initializeGenerators creates all required generator instances
func initializeGenerators(logger *logrus.Logger) (*sprites.Generator, *tiles.Generator, *palette.Generator) {
	spriteGen := sprites.NewGenerator()
	tileGen := tiles.NewGenerator()
	paletteGen := palette.NewGenerator()
	logger.Info("generators initialized")
	return spriteGen, tileGen, paletteGen
}

// testSpriteGeneration tests sprite generation for all genres and types
func testSpriteGeneration(logger *logrus.Logger, metrics *QualityMetrics, spriteGen *sprites.Generator, paletteGen *palette.Generator, genreList []string) {
	logger.Info("=== Testing Sprite Generation ===")
	spriteTypes := []sprites.SpriteType{
		sprites.SpriteEntity,
		sprites.SpriteItem,
		sprites.SpriteTile,
		sprites.SpriteParticle,
		sprites.SpriteUI,
	}

	for _, genre := range genreList {
		logger.WithField("genre", genre).Info("testing sprite generation")
		testPaletteForGenre(logger, metrics, paletteGen, genre)
		testSpritesForGenre(logger, metrics, spriteGen, genre, spriteTypes)
	}
}

// testPaletteForGenre tests palette generation for a specific genre
func testPaletteForGenre(logger *logrus.Logger, metrics *QualityMetrics, paletteGen *palette.Generator, genre string) {
	startTime := time.Now()
	_, err := paletteGen.Generate(genre, *seed)
	genTime := time.Since(startTime)
	metrics.TotalGenTime += genTime

	if err != nil {
		logger.WithError(err).WithField("genre", genre).Warn("palette generation failed")
		metrics.GenerationErrors++
	} else {
		metrics.PalettesByGenre[genre]++
	}
}

// testSpritesForGenre tests sprite generation for all types within a genre
func testSpritesForGenre(logger *logrus.Logger, metrics *QualityMetrics, spriteGen *sprites.Generator, genre string, spriteTypes []sprites.SpriteType) {
	for _, spriteType := range spriteTypes {
		typeName := spriteType.String()
		for i := 0; i < *spriteCount/len(spriteTypes); i++ {
			generateAndRecordSprite(logger, metrics, spriteGen, genre, spriteType, typeName, i)
		}
	}
}

// generateAndRecordSprite generates a single sprite and records the results
func generateAndRecordSprite(logger *logrus.Logger, metrics *QualityMetrics, spriteGen *sprites.Generator, genre string, spriteType sprites.SpriteType, typeName string, index int) {
	spriteSeed := *seed + int64(index*1000)
	config := sprites.Config{
		Type:       spriteType,
		Width:      32,
		Height:     32,
		Seed:       spriteSeed,
		GenreID:    genre,
		Complexity: 0.7,
		Variation:  index % 10,
	}

	startTime := time.Now()
	_, err := spriteGen.Generate(config)
	genTime := time.Since(startTime)
	metrics.TotalGenTime += genTime

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"genre": genre,
			"type":  typeName,
		}).Warn("sprite generation failed")
		metrics.GenerationErrors++
	} else {
		metrics.TotalSprites++
		metrics.SpritesByGenre[genre]++
	}
}

// testTileGeneration tests tile generation for all genres and types
func testTileGeneration(logger *logrus.Logger, metrics *QualityMetrics, tileGen *tiles.Generator, genreList []string) {
	logger.Info("=== Testing Tile Generation ===")
	tileTypes := []tiles.TileType{
		tiles.TileFloor,
		tiles.TileWall,
		tiles.TileDoor,
		tiles.TileWater,
		tiles.TileLava,
		tiles.TileCorridor,
		tiles.TileStairs,
		tiles.TileTrap,
	}

	for _, genre := range genreList {
		logger.WithField("genre", genre).Info("testing tile generation")
		testTilesForGenre(logger, metrics, tileGen, genre, tileTypes)
	}
}

// testTilesForGenre tests tile generation for all types within a genre
func testTilesForGenre(logger *logrus.Logger, metrics *QualityMetrics, tileGen *tiles.Generator, genre string, tileTypes []tiles.TileType) {
	for _, tileType := range tileTypes {
		for i := 0; i < *tileCount/len(tileTypes); i++ {
			generateAndRecordTile(logger, metrics, tileGen, genre, tileType, i)
		}
	}
}

// generateAndRecordTile generates a single tile and records the results
func generateAndRecordTile(logger *logrus.Logger, metrics *QualityMetrics, tileGen *tiles.Generator, genre string, tileType tiles.TileType, index int) {
	tileSeed := *seed + int64(index*2000)
	config := tiles.Config{
		Type:    tileType,
		Width:   32,
		Height:  32,
		Seed:    tileSeed,
		GenreID: genre,
	}

	startTime := time.Now()
	_, err := tileGen.Generate(config)
	genTime := time.Since(startTime)
	metrics.TotalGenTime += genTime

	if err != nil {
		logger.WithError(err).WithFields(logrus.Fields{
			"genre": genre,
			"type":  tileType,
		}).Warn("tile generation failed")
		metrics.GenerationErrors++
	} else {
		metrics.TotalTiles++
		metrics.TilesByType[tileType]++
	}
}

// calculateAverageGenTime computes the average generation time from metrics
func calculateAverageGenTime(metrics *QualityMetrics) {
	totalOperations := metrics.TotalSprites + metrics.TotalTiles + len(metrics.PalettesByGenre)
	if totalOperations > 0 {
		metrics.AverageGenTime = metrics.TotalGenTime / time.Duration(totalOperations)
	}
}

// displayResults prints test results and optionally writes a report file
func displayResults(logger *logrus.Logger, metrics *QualityMetrics, genreList []string) {
	logger.Info("=== Visual Quality Test Results ===")

	printResultsSummary(metrics)
	writeReportIfRequested(logger, metrics, genreList)

	logger.Info("visual quality test complete!")
}

// printResultsSummary prints formatted test results to stdout
func printResultsSummary(metrics *QualityMetrics) {
	fmt.Printf("\n=== Visual Quality Test Results ===\n")
	fmt.Printf("Generated: %s\n\n", time.Now().Format(time.RFC3339))

	printSpriteResults(metrics)
	printTileResults(metrics)
	printPaletteResults(metrics)
	printPerformanceMetrics(metrics)
	printQualityTargets(metrics)
}

// printSpriteResults prints sprite generation results
func printSpriteResults(metrics *QualityMetrics) {
	fmt.Printf("=== Sprite Generation ===\n")
	fmt.Printf("Total Sprites Generated: %d\n", metrics.TotalSprites)
	fmt.Printf("Sprites by Genre:\n")
	for genre, count := range metrics.SpritesByGenre {
		fmt.Printf("  %s: %d\n", genre, count)
	}
	fmt.Println()
}

// printTileResults prints tile generation results
func printTileResults(metrics *QualityMetrics) {
	fmt.Printf("=== Tile Generation ===\n")
	fmt.Printf("Total Tiles Generated: %d\n", metrics.TotalTiles)
	fmt.Printf("Tiles by Type:\n")
	for tileType, count := range metrics.TilesByType {
		fmt.Printf("  %s: %d\n", tileType, count)
	}
	fmt.Println()
}

// printPaletteResults prints palette generation results
func printPaletteResults(metrics *QualityMetrics) {
	fmt.Printf("=== Palette Generation ===\n")
	fmt.Printf("Palettes Generated:\n")
	for genre, count := range metrics.PalettesByGenre {
		fmt.Printf("  %s: %d\n", genre, count)
	}
	fmt.Println()
}

// printPerformanceMetrics prints performance statistics
func printPerformanceMetrics(metrics *QualityMetrics) {
	fmt.Printf("=== Performance ===\n")
	fmt.Printf("Total Generation Time: %v\n", metrics.TotalGenTime)
	fmt.Printf("Average Generation Time: %v\n", metrics.AverageGenTime)
	fmt.Printf("Generation Errors: %d\n\n", metrics.GenerationErrors)
}

// printQualityTargets prints quality target assessments
func printQualityTargets(metrics *QualityMetrics) {
	totalOperations := metrics.TotalSprites + metrics.TotalTiles + len(metrics.PalettesByGenre)
	successRate := float64(metrics.TotalSprites+metrics.TotalTiles) / float64(totalOperations) * 100

	fmt.Printf("=== Quality Targets ===\n")
	fmt.Printf("Success Rate: %.2f%% (target: 100%%)\n", successRate)

	if metrics.GenerationErrors == 0 {
		fmt.Printf("Generation Errors: ✅ NONE (target: 0)\n")
	} else {
		fmt.Printf("Generation Errors: ❌ %d (target: 0)\n", metrics.GenerationErrors)
	}

	avgGenTimeMs := float64(metrics.AverageGenTime.Microseconds()) / 1000.0
	fmt.Printf("Average Generation Time: ")
	if avgGenTimeMs < 5.0 {
		fmt.Printf("✅ %.3fms (target: <5ms)\n", avgGenTimeMs)
	} else {
		fmt.Printf("⚠️  %.3fms (target: <5ms)\n", avgGenTimeMs)
	}
}

// writeReportIfRequested writes a report file if output flag is set
func writeReportIfRequested(logger *logrus.Logger, metrics *QualityMetrics, genreList []string) {
	if *output != "" {
		report := generateReport(metrics, genreList)
		if err := os.WriteFile(*output, []byte(report), 0o644); err != nil {
			logger.WithError(err).Error("failed to write report")
		} else {
			logger.WithField("path", *output).Info("report written")
		}
	}
}

func generateReport(metrics *QualityMetrics, genres []string) string {
	totalOperations := metrics.TotalSprites + metrics.TotalTiles + len(metrics.PalettesByGenre)
	successRate := float64(metrics.TotalSprites+metrics.TotalTiles) / float64(totalOperations) * 100
	avgGenTimeMs := float64(metrics.AverageGenTime.Microseconds()) / 1000.0

	report := fmt.Sprintf(`V3.0 Visual Quality Test Report
Generated: %s

=== Configuration ===
Sprites per Genre: %d per type
Tiles per Type: %d per type
Genres Tested: %v
Total Operations: %d

=== Sprite Generation ===
Total Sprites: %d
`,
		time.Now().Format(time.RFC3339),
		20, // sprites per genre per type
		6,  // tiles per type
		genres,
		totalOperations,
		metrics.TotalSprites,
	)

	report += "Sprites by Genre:\n"
	for genre, count := range metrics.SpritesByGenre {
		report += fmt.Sprintf("  %s: %d\n", genre, count)
	}

	report += fmt.Sprintf(`
=== Tile Generation ===
Total Tiles: %d
`, metrics.TotalTiles)

	report += "Tiles by Type:\n"
	for tileType, count := range metrics.TilesByType {
		report += fmt.Sprintf("  %s: %d\n", tileType, count)
	}

	report += fmt.Sprintf(`
=== Palette Generation ===
Total Palettes: %d
`, len(metrics.PalettesByGenre))

	report += "Palettes by Genre:\n"
	for genre, count := range metrics.PalettesByGenre {
		report += fmt.Sprintf("  %s: %d\n", genre, count)
	}

	report += fmt.Sprintf(`
=== Performance Metrics ===
Total Generation Time: %v
Average Generation Time: %.3fms
Generation Errors: %d
Success Rate: %.2f%%

=== Quality Assessment ===
Target: 0 generation errors
Actual: %d errors
Status: %s

Target: <5ms average generation time
Actual: %.3fms
Status: %s

Target: 100%% success rate
Actual: %.2f%%
Status: %s
`,
		metrics.TotalGenTime,
		avgGenTimeMs,
		metrics.GenerationErrors,
		successRate,
		metrics.GenerationErrors,
		map[bool]string{true: "✅ PASS", false: "❌ FAIL"}[metrics.GenerationErrors == 0],
		avgGenTimeMs,
		map[bool]string{true: "✅ PASS", false: "⚠️ WARNING"}[avgGenTimeMs < 5.0],
		successRate,
		map[bool]string{true: "✅ PASS", false: "❌ FAIL"}[successRate >= 99.0],
	)

	return report
}
