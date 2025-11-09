// Package environment provides procedural generation of environmental objects.
package environment

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// PlacementConfig configures room decoration placement.
// Phase 20.1: Natural placement algorithm for 5-10 items per room.
type PlacementConfig struct {
	// RoomWidth and RoomHeight in tiles
	RoomWidth  int
	RoomHeight int

	// Density controls decoration count (0.0-1.0)
	// 0.5 = 5-10 items per room (target for Phase 20.1)
	Density float64

	// GenreID for biome-specific decorations
	GenreID string

	// Seed for deterministic placement
	Seed int64

	// AllowWallDecorations enables wall-mounted items
	AllowWallDecorations bool

	// AllowFloorDecorations enables floor items
	AllowFloorDecorations bool

	// MinSpacing minimum distance between decorations (in tiles)
	MinSpacing int
}

// DefaultPlacementConfig returns default placement configuration.
func DefaultPlacementConfig() PlacementConfig {
	return PlacementConfig{
		RoomWidth:             10,
		RoomHeight:            10,
		Density:               0.5,
		GenreID:               "fantasy",
		Seed:                  0,
		AllowWallDecorations:  true,
		AllowFloorDecorations: true,
		MinSpacing:            2,
	}
}

// PlacedObject represents a decoration placed in a room.
type PlacedObject struct {
	// Object reference
	Object *EnvironmentalObject

	// Position in room (tile coordinates)
	X, Y int

	// PlacementType indicates where object is placed
	PlacementType PlacementType

	// Variation applied to the object
	Variation VisualVariation
}

// PlacementType indicates where an object is placed.
type PlacementType int

const (
	// PlacementFloor indicates floor placement
	PlacementFloor PlacementType = iota
	// PlacementWallNorth indicates north wall placement
	PlacementWallNorth
	// PlacementWallSouth indicates south wall placement
	PlacementWallSouth
	// PlacementWallEast indicates east wall placement
	PlacementWallEast
	// PlacementWallWest indicates west wall placement
	PlacementWallWest
	// PlacementCorner indicates corner placement
	PlacementCorner
	// PlacementCenter indicates center placement (for important items)
	PlacementCenter
)

// String returns the string representation of a placement type.
func (p PlacementType) String() string {
	switch p {
	case PlacementFloor:
		return "Floor"
	case PlacementWallNorth:
		return "WallNorth"
	case PlacementWallSouth:
		return "WallSouth"
	case PlacementWallEast:
		return "WallEast"
	case PlacementWallWest:
		return "WallWest"
	case PlacementCorner:
		return "Corner"
	case PlacementCenter:
		return "Center"
	default:
		return "Unknown"
	}
}

// Placer handles decoration placement in rooms.
type Placer struct {
	generator *Generator
	logger    *logrus.Entry
}

// NewPlacer creates a new decoration placer.
func NewPlacer() *Placer {
	return NewPlacerWithLogger(nil)
}

// NewPlacerWithLogger creates a new decoration placer with a logger.
func NewPlacerWithLogger(logger *logrus.Logger) *Placer {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"component": "decoration-placer",
		})
	}
	return &Placer{
		generator: NewGeneratorWithLogger(logger),
		logger:    logEntry,
	}
}

// PlaceDecorations generates and places decorations in a room.
func (p *Placer) PlaceDecorations(config PlacementConfig) ([]*PlacedObject, error) {
	p.logDebug("placing decorations", logrus.Fields{
		"roomSize": config.RoomWidth * config.RoomHeight,
		"density":  config.Density,
		"genreID":  config.GenreID,
	})

	rng := rand.New(rand.NewSource(config.Seed))

	// Calculate target decoration count based on room size and density
	roomArea := config.RoomWidth * config.RoomHeight
	targetCount := p.calculateDecorationCount(roomArea, config.Density, rng)

	p.logDebug("calculated decoration count", logrus.Fields{
		"target": targetCount,
		"area":   roomArea,
	})

	// Generate decorations
	placements := make([]*PlacedObject, 0, targetCount)
	occupied := make(map[int]map[int]bool) // Track occupied positions

	// Select decoration types for this room
	decorationPool := p.selectDecorationPool(config.GenreID, rng)

	failureCount := 0
	maxFailures := targetCount * 3 // Allow some failures before giving up

	for len(placements) < targetCount && failureCount < maxFailures {
		// Select decoration type
		subType := decorationPool[rng.Intn(len(decorationPool))]

		// Determine placement location
		placement := p.selectPlacementType(subType, config, rng)
		x, y := p.selectPosition(placement, config, occupied, rng)

		if x < 0 || y < 0 {
			// No valid position found, increment failure counter and skip
			failureCount++
			continue
		}

		// Generate decoration object
		objConfig := Config{
			SubType: subType,
			Width:   32,
			Height:  32,
			GenreID: config.GenreID,
			Seed:    rng.Int63(),
		}

		obj, err := p.generator.Generate(objConfig)
		if err != nil {
			p.logError("failed to generate decoration", err, logrus.Fields{
				"subType": subType,
			})
			failureCount++
			continue
		}

		// Generate visual variation
		variation := GenerateVariation(subType, rng)

		// Create placed object
		placed := &PlacedObject{
			Object:        obj,
			X:             x,
			Y:             y,
			PlacementType: placement,
			Variation:     variation,
		}

		placements = append(placements, placed)

		// Mark position as occupied
		if occupied[x] == nil {
			occupied[x] = make(map[int]bool)
		}
		occupied[x][y] = true

		// Mark surrounding area based on min spacing
		for dx := -config.MinSpacing; dx <= config.MinSpacing; dx++ {
			for dy := -config.MinSpacing; dy <= config.MinSpacing; dy++ {
				nx := x + dx
				ny := y + dy
				if occupied[nx] == nil {
					occupied[nx] = make(map[int]bool)
				}
				occupied[nx][ny] = true
			}
		}
	}

	p.logInfo("decorations placed", logrus.Fields{
		"count":  len(placements),
		"target": targetCount,
	})

	return placements, nil
}

// calculateDecorationCount determines how many decorations to place.
func (p *Placer) calculateDecorationCount(roomArea int, density float64, rng *rand.Rand) int {
	// Base count: 5-10 items for average room (100 tiles)
	baseCount := 5 + rng.Intn(6) // 5-10

	// Scale with room size
	sizeMultiplier := float64(roomArea) / 100.0

	// Apply density
	count := int(float64(baseCount) * sizeMultiplier * density)

	// Ensure minimum and maximum
	if count < 3 {
		count = 3
	}
	if count > 20 {
		count = 20
	}

	return count
}

// selectDecorationPool returns suitable decoration types for a genre.
func (p *Placer) selectDecorationPool(genreID string, rng *rand.Rand) []SubType {
	// Base decorations available in all genres
	base := []SubType{
		SubTypePlant, SubTypeStatue, SubTypePainting, SubTypeBanner,
		SubTypeTorch, SubTypeVase, SubTypeBook, SubTypeBarrel,
		SubTypeCrate, SubTypePillar, SubTypeSconce,
	}

	// Genre-specific additions
	var additions []SubType
	switch genreID {
	case "fantasy":
		additions = []SubType{
			SubTypeCrystal, SubTypeTapestry, SubTypeCandlestick,
			SubTypeMushroom, SubTypeMoss,
		}
	case "scifi":
		additions = []SubType{
			SubTypeCrystal, SubTypeGraffiti, SubTypeWreckage,
		}
	case "horror":
		additions = []SubType{
			SubTypeSkull, SubTypeBloodstain, SubTypeChain,
			SubTypeWeb, SubTypeWallCrack, SubTypeMoss,
		}
	case "cyberpunk":
		additions = []SubType{
			SubTypeGraffiti, SubTypeDebris, SubTypeWreckage,
		}
	case "postapoc":
		additions = []SubType{
			SubTypeDebris, SubTypeWreckage, SubTypeRubble,
			SubTypeGrass, SubTypeWallCrack, SubTypeMoss,
		}
	}

	// Combine base and additions
	pool := append(base, additions...)
	return pool
}

// selectPlacementType determines where to place a decoration.
func (p *Placer) selectPlacementType(subType SubType, config PlacementConfig, rng *rand.Rand) PlacementType {
	// Wall decorations
	wallTypes := []SubType{
		SubTypePainting, SubTypeBanner, SubTypeTapestry,
		SubTypeTorch, SubTypeSconce, SubTypeWallCrack,
		SubTypeMoss, SubTypeGraffiti,
	}

	for _, wt := range wallTypes {
		if subType == wt && config.AllowWallDecorations {
			// Random wall
			walls := []PlacementType{
				PlacementWallNorth, PlacementWallSouth,
				PlacementWallEast, PlacementWallWest,
			}
			return walls[rng.Intn(len(walls))]
		}
	}

	// Floor decorations
	if config.AllowFloorDecorations {
		// Some items prefer corners
		cornerTypes := []SubType{
			SubTypeStatue, SubTypePlant, SubTypeVase,
			SubTypePillar, SubTypeBarrel, SubTypeCrate,
		}

		for _, ct := range cornerTypes {
			if subType == ct && rng.Float64() < 0.3 {
				return PlacementCorner
			}
		}

		return PlacementFloor
	}

	// Default to floor
	return PlacementFloor
}

// selectPosition finds a valid position for the placement type.
func (p *Placer) selectPosition(placement PlacementType, config PlacementConfig,
	occupied map[int]map[int]bool, rng *rand.Rand,
) (int, int) {
	maxAttempts := 50
	for attempt := 0; attempt < maxAttempts; attempt++ {
		var x, y int

		switch placement {
		case PlacementWallNorth:
			x = 1 + rng.Intn(config.RoomWidth-2)
			y = 1 // North wall

		case PlacementWallSouth:
			x = 1 + rng.Intn(config.RoomWidth-2)
			y = config.RoomHeight - 2 // South wall

		case PlacementWallEast:
			x = config.RoomWidth - 2 // East wall
			y = 1 + rng.Intn(config.RoomHeight-2)

		case PlacementWallWest:
			x = 1 // West wall
			y = 1 + rng.Intn(config.RoomHeight-2)

		case PlacementCorner:
			corners := [][2]int{
				{1, 1},
				{config.RoomWidth - 2, 1},
				{1, config.RoomHeight - 2},
				{config.RoomWidth - 2, config.RoomHeight - 2},
			}
			corner := corners[rng.Intn(len(corners))]
			x, y = corner[0], corner[1]

		case PlacementCenter:
			// Center area (middle third of room)
			x = config.RoomWidth/3 + rng.Intn(config.RoomWidth/3)
			y = config.RoomHeight/3 + rng.Intn(config.RoomHeight/3)

		default: // PlacementFloor
			x = 1 + rng.Intn(config.RoomWidth-2)
			y = 1 + rng.Intn(config.RoomHeight-2)
		}

		// Check if position is available
		if occupied[x] == nil || !occupied[x][y] {
			return x, y
		}
	}

	// No valid position found
	return -1, -1
}

// logDebug logs a debug message if logger and level are configured.
func (p *Placer) logDebug(msg string, fields logrus.Fields) {
	if p.logger != nil && p.logger.Logger.GetLevel() >= logrus.DebugLevel {
		p.logger.WithFields(fields).Debug(msg)
	}
}

// logInfo logs an info message if logger is configured.
func (p *Placer) logInfo(msg string, fields logrus.Fields) {
	if p.logger != nil {
		p.logger.WithFields(fields).Info(msg)
	}
}

// logError logs an error message if logger is configured.
func (p *Placer) logError(msg string, err error, fields ...logrus.Fields) {
	if p.logger != nil {
		entry := p.logger.WithError(err)
		if len(fields) > 0 {
			entry = entry.WithFields(fields[0])
		}
		entry.Error(msg)
	}
}
