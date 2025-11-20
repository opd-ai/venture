package territory_siege

import (
	"fmt"
	"math/rand"
)

// StructureGenerator generates defensive structures for territories.
type StructureGenerator struct {
	rng *rand.Rand
}

// NewStructureGenerator creates a new structure generator with the given RNG source.
func NewStructureGenerator(source rand.Source) *StructureGenerator {
	return &StructureGenerator{
		rng: rand.New(source),
	}
}

// GenerateStructures creates a set of defensive structures for a zone.
func (sg *StructureGenerator) GenerateStructures(zoneID string, count int) []*DefensiveStructure {
	structures := make([]*DefensiveStructure, 0, count)

	// Always include a keep (guild hall defense)
	keep := sg.generateStructure(zoneID, StructureKeep, 0)
	structures = append(structures, keep)

	// Generate remaining structures with weighted distribution
	wallCount := count / 3                                         // ~33% walls
	towerCount := count / 4                                        // ~25% towers
	gateCount := (count - 1) / 5                                   // ~20% gates (minus keep)
	barrackCount := count - 1 - wallCount - towerCount - gateCount // Rest are barracks

	idx := 1
	for i := 0; i < wallCount; i++ {
		structures = append(structures, sg.generateStructure(zoneID, StructureWall, idx))
		idx++
	}
	for i := 0; i < towerCount; i++ {
		structures = append(structures, sg.generateStructure(zoneID, StructureTower, idx))
		idx++
	}
	for i := 0; i < gateCount; i++ {
		structures = append(structures, sg.generateStructure(zoneID, StructureGate, idx))
		idx++
	}
	for i := 0; i < barrackCount; i++ {
		structures = append(structures, sg.generateStructure(zoneID, StructureBarracks, idx))
		idx++
	}

	return structures
}

// generateStructure creates a single defensive structure.
func (sg *StructureGenerator) generateStructure(zoneID string, structureType StructureType, index int) *DefensiveStructure {
	// Generate structure ID
	structureID := fmt.Sprintf("%s_%s_%d", zoneID, structureType.String(), index)

	// Generate position (random within zone bounds)
	x := sg.rng.Float64() * 1000.0 // Zone assumed to be 1000×1000
	y := sg.rng.Float64() * 1000.0

	// Generate HP based on structure type
	var maxHP int
	switch structureType {
	case StructureWall:
		maxHP = 1000 + sg.rng.Intn(4001) // 1000-5000
	case StructureTower:
		maxHP = 500 + sg.rng.Intn(1001) // 500-1500
	case StructureGate:
		maxHP = 800 + sg.rng.Intn(1201) // 800-2000
	case StructureBarracks:
		maxHP = 300 + sg.rng.Intn(501) // 300-800
	case StructureKeep:
		maxHP = 10000 + sg.rng.Intn(10001) // 10000-20000
	}

	return &DefensiveStructure{
		StructureID:  structureID,
		Type:         structureType,
		X:            x,
		Y:            y,
		MaxHP:        maxHP,
		CurrentHP:    maxHP,
		LastDamageAt: 0,
		IsDestroyed:  false,
	}
}

// GetHPRange returns the min and max HP for a structure type.
func GetHPRange(structureType StructureType) (min, max int) {
	switch structureType {
	case StructureWall:
		return 1000, 5000
	case StructureTower:
		return 500, 1500
	case StructureGate:
		return 800, 2000
	case StructureBarracks:
		return 300, 800
	case StructureKeep:
		return 10000, 20000
	default:
		return 0, 0
	}
}
