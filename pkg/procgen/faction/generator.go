// Package faction provides procedural generation of factions for the game world.
// Factions are generated deterministically from world seeds and include kingdoms,
// guilds, cults, corporations, and gangs with procedurally generated names,
// relationships, and characteristics.
package faction

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/opd-ai/venture/pkg/procgen"
	"github.com/sirupsen/logrus"
)

// logger is the package-level logger for faction generation
var logger = logrus.WithField("package", "procgen/faction")

// Generator procedurally generates factions for a game world
type Generator struct {
	// No state - stateless generator
}

// NewGenerator creates a new faction generator
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate creates a set of factions for the world
func (g *Generator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	// Validate parameters
	if err := g.Validate(params); err != nil {
		logger.WithFields(logrus.Fields{
			"seed":       seed,
			"depth":      params.Depth,
			"difficulty": params.Difficulty,
			"genre":      params.GenreID,
			"error":      err.Error(),
		}).Debug("faction generation validation failed")
		return nil, fmt.Errorf("faction generation validation failed: %w", err)
	}

	rng := rand.New(rand.NewSource(seed))
	genreID := params.GenreID
	if genreID == "" {
		genreID = "fantasy" // Default genre
	}

	// Determine number of factions based on depth (3-7 factions per world)
	numFactions := 3 + int(params.Depth/10)
	if numFactions > 7 {
		numFactions = 7
	}

	factions := make([]*engine.Faction, numFactions)

	// Generate each faction
	for i := 0; i < numFactions; i++ {
		factionType := g.chooseFactionType(rng, genreID)
		faction := &engine.Faction{
			ID:             fmt.Sprintf("faction_%d", i),
			Name:           g.generateFactionName(rng, factionType, genreID),
			Type:           factionType,
			GenreID:        genreID,
			Description:    g.generateDescription(rng, factionType, genreID),
			Relationships:  make(map[string]int),
			TerritoryColor: g.generateColor(rng),
			MemberCount:    100 + rng.Intn(900), // 100-999 members
		}
		factions[i] = faction
	}

	// Generate relationships between factions
	g.generateRelationships(rng, factions)

	return factions, nil
}

// Validate checks if generation parameters are valid
func (g *Generator) Validate(params interface{}) error {
	p, ok := params.(procgen.GenerationParams)
	if !ok {
		logger.WithField("params_type", fmt.Sprintf("%T", params)).Debug("invalid params type for faction validation")
		return fmt.Errorf("invalid params type, expected GenerationParams")
	}

	if p.Depth < 0 {
		logger.WithField("depth", p.Depth).Debug("faction validation failed: negative depth")
		return fmt.Errorf("depth cannot be negative")
	}

	if p.Difficulty < 0.0 || p.Difficulty > 1.0 {
		logger.WithField("difficulty", p.Difficulty).Debug("faction validation failed: difficulty out of range")
		return fmt.Errorf("difficulty must be between 0.0 and 1.0")
	}

	return nil
}

// chooseFactionType selects an appropriate faction type for the genre
func (g *Generator) chooseFactionType(rng *rand.Rand, genreID string) engine.FactionType {
	// Genre-specific faction type weights
	weights := g.getFactionTypeWeights(genreID)

	// Sort faction types for deterministic iteration
	// FactionType is a string type, so lexicographic comparison ensures consistent ordering
	types := make([]engine.FactionType, 0, len(weights))
	for factionType := range weights {
		types = append(types, factionType)
	}
	sort.Slice(types, func(i, j int) bool {
		return types[i] < types[j]
	})

	// Calculate total weight
	total := 0
	for _, weight := range weights {
		total += weight
	}

	// Weighted random selection with deterministic iteration
	roll := rng.Intn(total)
	cumulative := 0

	for _, factionType := range types {
		cumulative += weights[factionType]
		if roll < cumulative {
			return factionType
		}
	}

	return engine.FactionTypeGuild // Fallback
}

// getFactionTypeWeights returns genre-specific faction type weights
func (g *Generator) getFactionTypeWeights(genreID string) map[engine.FactionType]int {
	switch genreID {
	case "fantasy":
		return map[engine.FactionType]int{
			engine.FactionTypeKingdom:   40,
			engine.FactionTypeGuild:     30,
			engine.FactionTypeCult:      15,
			engine.FactionTypeMerchants: 15,
		}
	case "sci-fi":
		return map[engine.FactionType]int{
			engine.FactionTypeCorporation: 40,
			engine.FactionTypeGuild:       25,
			engine.FactionTypeRebels:      20,
			engine.FactionTypeMerchants:   15,
		}
	case "horror":
		return map[engine.FactionType]int{
			engine.FactionTypeCult:      50,
			engine.FactionTypeGang:      30,
			engine.FactionTypeMerchants: 20,
		}
	case "cyberpunk":
		return map[engine.FactionType]int{
			engine.FactionTypeCorporation: 35,
			engine.FactionTypeGang:        35,
			engine.FactionTypeRebels:      20,
			engine.FactionTypeMerchants:   10,
		}
	case "post-apocalyptic":
		return map[engine.FactionType]int{
			engine.FactionTypeGang:      40,
			engine.FactionTypeRebels:    30,
			engine.FactionTypeMerchants: 20,
			engine.FactionTypeCult:      10,
		}
	default:
		return map[engine.FactionType]int{
			engine.FactionTypeGuild: 100,
		}
	}
}

// generateFactionName creates a genre-appropriate faction name
func (g *Generator) generateFactionName(rng *rand.Rand, factionType engine.FactionType, genreID string) string {
	prefixes := g.getNamePrefixes(genreID, factionType)
	suffixes := g.getNameSuffixes(genreID, factionType)

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return prefix + " " + suffix
}

// getNamePrefixes returns genre and type-specific name prefixes
func (g *Generator) getNamePrefixes(genreID string, factionType engine.FactionType) []string {
	switch genreID {
	case "fantasy":
		switch factionType {
		case engine.FactionTypeKingdom:
			return []string{"Kingdom of", "Realm of", "Empire of", "Dominion of"}
		case engine.FactionTypeGuild:
			return []string{"Guild of", "Order of", "Fellowship of", "Brotherhood of"}
		case engine.FactionTypeCult:
			return []string{"Cult of", "Circle of", "Sect of", "Cabal of"}
		case engine.FactionTypeMerchants:
			return []string{"Trading Company of", "Merchant League of", "Traders of"}
		}
	case "sci-fi":
		switch factionType {
		case engine.FactionTypeCorporation:
			return []string{"Megacorp", "Systems", "Industries", "Technologies"}
		case engine.FactionTypeRebels:
			return []string{"Free", "Liberation", "Resistance", "Alliance"}
		case engine.FactionTypeMerchants:
			return []string{"Trade Consortium", "Merchant Fleet", "Commerce Guild"}
		}
	case "horror":
		switch factionType {
		case engine.FactionTypeCult:
			return []string{"The Cult of", "Disciples of", "Worshippers of", "Followers of"}
		case engine.FactionTypeGang:
			return []string{"The", "Blood", "Dark", "Shadow"}
		}
	case "cyberpunk":
		switch factionType {
		case engine.FactionTypeCorporation:
			return []string{"Megacorp", "Zaibatsu", "Conglomerate", "Syndicate"}
		case engine.FactionTypeGang:
			return []string{"The", "Neon", "Chrome", "Street"}
		case engine.FactionTypeRebels:
			return []string{"Underground", "The Resistance", "Free", "Liberation"}
		}
	case "post-apocalyptic":
		switch factionType {
		case engine.FactionTypeGang:
			return []string{"Wastelanders", "Raiders", "Scavengers", "Survivors"}
		case engine.FactionTypeRebels:
			return []string{"The Resistance", "Free State", "Liberation Army"}
		case engine.FactionTypeMerchants:
			return []string{"Trade Caravan", "Merchant Alliance", "Traders"}
		}
	}

	return []string{"The"} // Fallback
}

// getNameSuffixes returns genre and type-specific name suffixes
func (g *Generator) getNameSuffixes(genreID string, factionType engine.FactionType) []string {
	switch genreID {
	case "fantasy":
		return []string{"Silverwood", "Ironhold", "Stormhaven", "Dragonspire", "Moonshadow", "Starfall"}
	case "sci-fi":
		return []string{"Prime", "Nova", "Zenith", "Nexus", "Omega", "Titan", "Quantum"}
	case "horror":
		return []string{"the Void", "Eternal Night", "Crimson Moon", "Shadows", "the Abyss"}
	case "cyberpunk":
		return []string{"Runners", "Samurai", "Kings", "Ghosts", "Punks", "Hawks"}
	case "post-apocalyptic":
		return []string{"of the Wastes", "Riders", "Collective", "Brotherhood", "Clan"}
	default:
		return []string{"Faction"}
	}
}

// generateDescription creates a description for the faction
func (g *Generator) generateDescription(rng *rand.Rand, factionType engine.FactionType, genreID string) string {
	templates := g.getDescriptionTemplates(factionType, genreID)
	return templates[rng.Intn(len(templates))]
}

// getDescriptionTemplates returns description templates for faction types
func (g *Generator) getDescriptionTemplates(factionType engine.FactionType, genreID string) []string {
	switch factionType {
	case engine.FactionTypeKingdom:
		return []string{
			"A powerful kingdom that rules with an iron fist.",
			"An ancient monarchy with deep traditions.",
			"A prosperous realm known for its skilled warriors.",
		}
	case engine.FactionTypeGuild:
		return []string{
			"A skilled organization of craftsmen and traders.",
			"A secretive guild with ancient knowledge.",
			"A professional association controlling key resources.",
		}
	case engine.FactionTypeCult:
		return []string{
			"A mysterious cult worshipping dark powers.",
			"Fanatic believers seeking forbidden knowledge.",
			"A secretive group with sinister intentions.",
		}
	case engine.FactionTypeCorporation:
		return []string{
			"A megacorporation with vast resources and influence.",
			"A ruthless business empire controlling key industries.",
			"A powerful conglomerate prioritizing profit above all.",
		}
	case engine.FactionTypeGang:
		return []string{
			"A violent gang controlling territory through force.",
			"Criminals and outlaws banding together.",
			"A ruthless organization thriving in chaos.",
		}
	case engine.FactionTypeRebels:
		return []string{
			"Freedom fighters resisting oppressive forces.",
			"Rebels seeking to overthrow the established order.",
			"A resistance movement fighting for their cause.",
		}
	case engine.FactionTypeMerchants:
		return []string{
			"Traders and merchants controlling commerce.",
			"A merchant alliance facilitating trade across regions.",
			"Business-minded individuals seeking profit.",
		}
	}

	return []string{"A faction in the world."}
}

// generateColor creates a random territory color for visualization
func (g *Generator) generateColor(rng *rand.Rand) [4]uint8 {
	return [4]uint8{
		uint8(rng.Intn(256)),
		uint8(rng.Intn(256)),
		uint8(rng.Intn(256)),
		255, // Full alpha
	}
}

// generateRelationships creates relationships between factions
func (g *Generator) generateRelationships(rng *rand.Rand, factions []*engine.Faction) {
	for i := 0; i < len(factions); i++ {
		for j := i + 1; j < len(factions); j++ {
			// Generate relationship value based on faction types
			relationship := g.calculateRelationship(rng, factions[i], factions[j])

			// Set bidirectional relationships
			factions[i].Relationships[factions[j].ID] = relationship
			factions[j].Relationships[factions[i].ID] = relationship
		}
	}
}

// calculateRelationship determines the relationship between two factions
func (g *Generator) calculateRelationship(rng *rand.Rand, faction1, faction2 *engine.Faction) int {
	// Base relationship depends on faction types
	baseRelationship := 0

	// Similar types tend to compete (slight negative)
	if faction1.Type == faction2.Type {
		baseRelationship = -20 + rng.Intn(30) // -20 to +10
	} else {
		// Different types can be neutral to friendly
		baseRelationship = -10 + rng.Intn(40) // -10 to +30
	}

	// Special relationships
	switch {
	case faction1.Type == engine.FactionTypeCult || faction2.Type == engine.FactionTypeCult:
		// Cults tend to be distrusted
		baseRelationship -= 20
	case faction1.Type == engine.FactionTypeMerchants || faction2.Type == engine.FactionTypeMerchants:
		// Merchants tend to be neutral with everyone
		baseRelationship += 10
	case faction1.Type == engine.FactionTypeRebels && faction2.Type == engine.FactionTypeCorporation:
		// Rebels vs Corporations are enemies
		baseRelationship = -75
	case faction1.Type == engine.FactionTypeCorporation && faction2.Type == engine.FactionTypeRebels:
		// Corporations vs Rebels are enemies
		baseRelationship = -75
	}

	// Add randomness
	relationship := baseRelationship + rng.Intn(30) - 15 // ±15 variance

	// Clamp to valid range
	if relationship > 100 {
		relationship = 100
	}
	if relationship < -100 {
		relationship = -100
	}

	return relationship
}
