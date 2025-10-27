// Package station provides procedural crafting station generation.
// This file implements deterministic station generators that create crafting stations
// (alchemy tables, forges, workbenches) based on genre and seed.
//
// Design Philosophy:
// - Deterministic: same seed always generates same stations
// - Genre-themed: station names reflect genre aesthetics
// - Balanced: 3 stations per area (one per recipe type)
// - Simple: minimal complexity following SIMPLICITY RULE
package station

import (
	"fmt"
	"math/rand"

	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// StationType represents the type of crafting station.
type StationType int

const (
	// StationAlchemyTable is for potion brewing.
	StationAlchemyTable StationType = iota
	// StationForge is for enchanting equipment.
	StationForge
	// StationWorkbench is for crafting magic items.
	StationWorkbench
)

// String returns the string representation of the station type.
func (s StationType) String() string {
	switch s {
	case StationAlchemyTable:
		return "Alchemy Table"
	case StationForge:
		return "Forge"
	case StationWorkbench:
		return "Workbench"
	default:
		return "Unknown"
	}
}

// StationData represents a generated crafting station.
type StationData struct {
	StationType StationType
	Name        string
	GenreID     string
	Seed        int64
	SpawnX      float64 // Set by spawn system, not generator
	SpawnY      float64 // Set by spawn system, not generator
}

// StationNameTemplate defines naming patterns for stations.
type StationNameTemplate struct {
	Prefix    []string
	Adjective []string
	Noun      []string
}

// StationGenerator generates procedural crafting stations.
type StationGenerator struct {
	nameTemplates map[string]map[StationType]StationNameTemplate
	logger        *logrus.Entry
}

// NewStationGenerator creates a new station generator.
func NewStationGenerator() *StationGenerator {
	return NewStationGeneratorWithLogger(nil)
}

// NewStationGeneratorWithLogger creates a new station generator with a logger.
func NewStationGeneratorWithLogger(logger *logrus.Logger) *StationGenerator {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("generator", "station")
	}

	gen := &StationGenerator{
		nameTemplates: make(map[string]map[StationType]StationNameTemplate),
		logger:        logEntry,
	}

	// Register name templates for all genres
	gen.registerFantasyTemplates()
	gen.registerSciFiTemplates()
	gen.registerHorrorTemplates()
	gen.registerCyberpunkTemplates()
	gen.registerPostApocTemplates()

	// Default templates (fantasy)
	gen.nameTemplates[""] = gen.nameTemplates["fantasy"]

	if logEntry != nil {
		logEntry.Debug("station generator initialized")
	}

	return gen
}

// Generate creates crafting stations based on seed and parameters.
// Always generates exactly 3 stations (one per station type).
// Returns a slice of *StationData.
func (g *StationGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
		g.logger.WithFields(logrus.Fields{
			"seed":    seed,
			"genreID": params.GenreID,
		}).Debug("starting station generation")
	}

	rng := rand.New(rand.NewSource(seed))
	genreID := params.GenreID
	if genreID == "" {
		genreID = "fantasy"
	}

	// Get templates for this genre
	templates, exists := g.nameTemplates[genreID]
	if !exists {
		templates = g.nameTemplates["fantasy"]
	}

	// Generate exactly 3 stations (one per type)
	stations := make([]*StationData, 3)
	stationTypes := []StationType{StationAlchemyTable, StationForge, StationWorkbench}

	for i, stationType := range stationTypes {
		stationSeed := seed + int64(i*100) // Derive unique seed per station
		name := g.generateStationName(rng, templates[stationType])

		stations[i] = &StationData{
			StationType: stationType,
			Name:        name,
			GenreID:     genreID,
			Seed:        stationSeed,
			SpawnX:      0, // Set by spawn system
			SpawnY:      0, // Set by spawn system
		}

		if g.logger != nil && g.logger.Logger.GetLevel() >= logrus.DebugLevel {
			g.logger.WithFields(logrus.Fields{
				"stationType": stationType.String(),
				"name":        name,
				"genreID":     genreID,
			}).Debug("generated station")
		}
	}

	return stations, nil
}

// Validate checks if generated stations meet quality standards.
func (g *StationGenerator) Validate(result interface{}) error {
	stations, ok := result.([]*StationData)
	if !ok {
		return fmt.Errorf("result is not []*StationData")
	}

	if len(stations) != 3 {
		return fmt.Errorf("expected exactly 3 stations, got %d", len(stations))
	}

	// Check that we have one of each type
	typeCount := make(map[StationType]int)
	for _, station := range stations {
		if station == nil {
			return fmt.Errorf("station is nil")
		}

		if station.Name == "" {
			return fmt.Errorf("station has empty name")
		}

		if station.StationType < StationAlchemyTable || station.StationType > StationWorkbench {
			return fmt.Errorf("invalid station type: %d", station.StationType)
		}

		typeCount[station.StationType]++
	}

	// Verify one of each type
	if typeCount[StationAlchemyTable] != 1 || typeCount[StationForge] != 1 || typeCount[StationWorkbench] != 1 {
		return fmt.Errorf("must have exactly one of each station type, got: alchemy=%d, forge=%d, workbench=%d",
			typeCount[StationAlchemyTable], typeCount[StationForge], typeCount[StationWorkbench])
	}

	return nil
}

// generateStationName creates a station name from a template.
func (g *StationGenerator) generateStationName(rng *rand.Rand, template StationNameTemplate) string {
	// Combine prefix + adjective + noun
	// Example: "Ancient Arcane Table" or "Corrupted Forge"

	var name string

	// 50% chance to include prefix
	if len(template.Prefix) > 0 && rng.Float64() < 0.5 {
		prefix := template.Prefix[rng.Intn(len(template.Prefix))]
		name = prefix + " "
	}

	// Always include adjective if available
	if len(template.Adjective) > 0 {
		adj := template.Adjective[rng.Intn(len(template.Adjective))]
		name += adj + " "
	}

	// Always include noun
	noun := template.Noun[rng.Intn(len(template.Noun))]
	name += noun

	return name
}

// registerFantasyTemplates registers fantasy-themed station name templates.
func (g *StationGenerator) registerFantasyTemplates() {
	g.nameTemplates["fantasy"] = map[StationType]StationNameTemplate{
		StationAlchemyTable: {
			Prefix:    []string{"Ancient", "Mystical", "Arcane", "Sacred"},
			Adjective: []string{"Alchemical", "Enchanted", "Magical", "Blessed"},
			Noun:      []string{"Table", "Workbench", "Altar", "Station"},
		},
		StationForge: {
			Prefix:    []string{"Dwarven", "Ancient", "Legendary", "Master"},
			Adjective: []string{"Flaming", "Runic", "Enchanted", "Tempered"},
			Noun:      []string{"Forge", "Anvil", "Smithy", "Hearth"},
		},
		StationWorkbench: {
			Prefix:    []string{"Artisan's", "Master", "Crafting", "Enchanter's"},
			Adjective: []string{"Magical", "Precision", "Fine", "Skilled"},
			Noun:      []string{"Workbench", "Table", "Station", "Bench"},
		},
	}
}

// registerSciFiTemplates registers sci-fi-themed station name templates.
func (g *StationGenerator) registerSciFiTemplates() {
	g.nameTemplates["scifi"] = map[StationType]StationNameTemplate{
		StationAlchemyTable: {
			Prefix:    []string{"Molecular", "Quantum", "Nano", "Bio"},
			Adjective: []string{"Synthesis", "Assembly", "Processing", "Fabrication"},
			Noun:      []string{"Station", "Unit", "Terminal", "Module"},
		},
		StationForge: {
			Prefix:    []string{"Plasma", "Laser", "Fusion", "Energy"},
			Adjective: []string{"Fabrication", "Manufacturing", "Assembly", "Forging"},
			Noun:      []string{"Station", "Unit", "Terminal", "Bay"},
		},
		StationWorkbench: {
			Prefix:    []string{"Tech", "Engineering", "Robotics", "Cybernetics"},
			Adjective: []string{"Assembly", "Modification", "Crafting", "Fabrication"},
			Noun:      []string{"Station", "Terminal", "Workbench", "Bay"},
		},
	}
}

// registerHorrorTemplates registers horror-themed station name templates.
func (g *StationGenerator) registerHorrorTemplates() {
	g.nameTemplates["horror"] = map[StationType]StationNameTemplate{
		StationAlchemyTable: {
			Prefix:    []string{"Cursed", "Forbidden", "Corrupted", "Dark"},
			Adjective: []string{"Necromantic", "Unholy", "Twisted", "Sinister"},
			Noun:      []string{"Table", "Altar", "Station", "Slab"},
		},
		StationForge: {
			Prefix:    []string{"Blood", "Bone", "Shadow", "Cursed"},
			Adjective: []string{"Infernal", "Profane", "Corrupted", "Damned"},
			Noun:      []string{"Forge", "Anvil", "Hearth", "Pit"},
		},
		StationWorkbench: {
			Prefix:    []string{"Mad", "Twisted", "Corrupted", "Diseased"},
			Adjective: []string{"Surgeon's", "Butcher's", "Torturer's", "Anatomist's"},
			Noun:      []string{"Table", "Workbench", "Station", "Bench"},
		},
	}
}

// registerCyberpunkTemplates registers cyberpunk-themed station name templates.
func (g *StationGenerator) registerCyberpunkTemplates() {
	g.nameTemplates["cyberpunk"] = map[StationType]StationNameTemplate{
		StationAlchemyTable: {
			Prefix:    []string{"Synth", "Chem", "Bio", "Neural"},
			Adjective: []string{"Mixing", "Synthesis", "Processing", "Enhancement"},
			Noun:      []string{"Station", "Terminal", "Unit", "Rig"},
		},
		StationForge: {
			Prefix:    []string{"Cyber", "Augment", "Tech", "Mech"},
			Adjective: []string{"Modification", "Enhancement", "Upgrade", "Fabrication"},
			Noun:      []string{"Station", "Terminal", "Bay", "Rig"},
		},
		StationWorkbench: {
			Prefix:    []string{"Hacker's", "Tech", "Street", "Black Market"},
			Adjective: []string{"Assembly", "Modding", "Crafting", "Tuning"},
			Noun:      []string{"Station", "Terminal", "Workbench", "Rig"},
		},
	}
}

// registerPostApocTemplates registers post-apocalyptic-themed station name templates.
func (g *StationGenerator) registerPostApocTemplates() {
	g.nameTemplates["postapoc"] = map[StationType]StationNameTemplate{
		StationAlchemyTable: {
			Prefix:    []string{"Makeshift", "Scavenged", "Salvaged", "Wasteland"},
			Adjective: []string{"Brewing", "Mixing", "Chemistry", "Processing"},
			Noun:      []string{"Station", "Table", "Setup", "Bench"},
		},
		StationForge: {
			Prefix:    []string{"Scrap", "Wasteland", "Salvage", "Survivor's"},
			Adjective: []string{"Welding", "Metalworking", "Forging", "Repair"},
			Noun:      []string{"Station", "Forge", "Workshop", "Bench"},
		},
		StationWorkbench: {
			Prefix:    []string{"Survivor's", "Scavenger's", "Wasteland", "Makeshift"},
			Adjective: []string{"Repair", "Crafting", "Assembly", "Tinkering"},
			Noun:      []string{"Workbench", "Table", "Station", "Setup"},
		},
	}
}
