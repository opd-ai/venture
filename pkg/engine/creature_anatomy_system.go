// Package engine provides the CreatureAnatomySystem which assigns appropriate
// anatomical templates to creature entities based on their name, type, and tags.
// This bridges procgen entity generation with the aerial nonhumanoid sprite templates.
package engine

import (
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/rendering/sprites"
	"github.com/sirupsen/logrus"
)

// AnatomyType represents the body plan used for sprite generation.
type AnatomyType int

const (
	// AnatomyHumanoid is the default bipedal body plan
	AnatomyHumanoid AnatomyType = iota
	// AnatomyQuadruped is four-legged animals (wolves, bears, horses)
	AnatomyQuadruped
	// AnatomySerpentine is elongated bodies (snakes, worms, wyrms)
	AnatomySerpentine
	// AnatomyArachnid is eight-legged creatures (spiders, scorpions)
	AnatomyArachnid
	// AnatomyInsect is six-legged segmented creatures (beetles, ants)
	AnatomyInsect
	// AnatomyFlying is winged creatures (dragons, bats, birds)
	AnatomyFlying
	// AnatomyBlob is amorphous masses (slimes, oozes)
	AnatomyBlob
	// AnatomyMechanical is constructs and robots (golems, androids)
	AnatomyMechanical
	// AnatomyUndead is skeletal/ghostly variants
	AnatomyUndead
	// AnatomyMultiLimbed is eldritch horrors (tentacles, aberrations)
	AnatomyMultiLimbed
)

// String returns the template selector string for this anatomy type.
func (a AnatomyType) String() string {
	switch a {
	case AnatomyHumanoid:
		return "humanoid"
	case AnatomyQuadruped:
		return "quadruped"
	case AnatomySerpentine:
		return "serpentine"
	case AnatomyArachnid:
		return "arachnid"
	case AnatomyInsect:
		return "insect"
	case AnatomyFlying:
		return "flying"
	case AnatomyBlob:
		return "blob"
	case AnatomyMechanical:
		return "mechanical"
	case AnatomyUndead:
		return "undead"
	case AnatomyMultiLimbed:
		return "multi_limbed"
	default:
		return "humanoid"
	}
}

// CreatureAnatomyComponent stores the anatomical type and template selection
// for creature sprite generation.
type CreatureAnatomyComponent struct {
	// AnatomyType determines which body plan template to use
	AnatomyType AnatomyType
	// SubType provides further specialization (e.g., "wolf" vs "bear")
	SubType string
	// SizeModifier scales the anatomy template (0.5 = tiny, 1.0 = medium, 2.0 = huge)
	SizeModifier float64
	// GenreVariant stores genre-specific visual modifications
	GenreVariant string
	// Assigned marks whether anatomy has been determined
	Assigned bool
	// Seed for deterministic variation
	Seed int64
}

// Type returns the component type identifier.
func (c *CreatureAnatomyComponent) Type() string { return "creature_anatomy" }

// NewCreatureAnatomyComponent creates a new unassigned anatomy component.
func NewCreatureAnatomyComponent() *CreatureAnatomyComponent {
	return &CreatureAnatomyComponent{
		AnatomyType:  AnatomyHumanoid,
		SubType:      "",
		SizeModifier: 1.0,
		GenreVariant: "",
		Assigned:     false,
		Seed:         0,
	}
}

// CreatureAnatomySystem assigns anatomy types to creatures based on their
// entity properties, enabling correct nonhumanoid sprite template selection.
type CreatureAnatomySystem struct {
	world    *World
	seed     int64
	genreID  string
	logger   *logrus.Entry
	keywords map[string]AnatomyType
}

// NewCreatureAnatomySystem creates a new creature anatomy system.
func NewCreatureAnatomySystem(world *World, seed int64) *CreatureAnatomySystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "creature_anatomy",
		"seed":        seed,
	})

	sys := &CreatureAnatomySystem{
		world:   world,
		seed:    seed,
		genreID: "",
		logger:  logger,
	}

	// Initialize keyword to anatomy mapping
	sys.keywords = buildKeywordMap()

	logger.Debug("creature anatomy system created")
	return sys
}

// buildKeywordMap creates the mapping from name/tag keywords to anatomy types.
func buildKeywordMap() map[string]AnatomyType {
	return map[string]AnatomyType{
		// Quadrupeds
		"wolf": AnatomyQuadruped, "bear": AnatomyQuadruped, "horse": AnatomyQuadruped,
		"dog": AnatomyQuadruped, "cat": AnatomyQuadruped, "lion": AnatomyQuadruped,
		"tiger": AnatomyQuadruped, "boar": AnatomyQuadruped, "deer": AnatomyQuadruped,
		"beast": AnatomyQuadruped, "animal": AnatomyQuadruped, "hound": AnatomyQuadruped,
		"panther": AnatomyQuadruped, "jaguar": AnatomyQuadruped, "leopard": AnatomyQuadruped,
		"rhinoceros": AnatomyQuadruped, "elephant": AnatomyQuadruped,

		// Serpentine
		"snake": AnatomySerpentine, "serpent": AnatomySerpentine, "worm": AnatomySerpentine,
		"wyrm": AnatomySerpentine, "viper": AnatomySerpentine, "cobra": AnatomySerpentine,
		"eel": AnatomySerpentine, "naga": AnatomySerpentine, "basilisk": AnatomySerpentine,

		// Arachnids
		"spider": AnatomyArachnid, "scorpion": AnatomyArachnid, "tarantula": AnatomyArachnid,
		"arachnid": AnatomyArachnid, "tick": AnatomyArachnid,

		// Insects
		"beetle": AnatomyInsect, "ant": AnatomyInsect, "centipede": AnatomyInsect,
		"mantis": AnatomyInsect, "wasp": AnatomyInsect, "moth": AnatomyInsect,
		"bee": AnatomyInsect, "fly": AnatomyInsect, "roach": AnatomyInsect,
		"crawler": AnatomyInsect, "bug": AnatomyInsect, "locust": AnatomyInsect,

		// Flying
		"dragon": AnatomyFlying, "bat": AnatomyFlying, "bird": AnatomyFlying,
		"wyvern": AnatomyFlying, "phoenix": AnatomyFlying, "harpy": AnatomyFlying,
		"gargoyle": AnatomyFlying, "gryphon": AnatomyFlying, "griffin": AnatomyFlying,
		"roc": AnatomyFlying, "pterodactyl": AnatomyFlying,

		// Blobs
		"slime": AnatomyBlob, "ooze": AnatomyBlob, "blob": AnatomyBlob,
		"jelly": AnatomyBlob, "amoeba": AnatomyBlob, "pudding": AnatomyBlob,
		"gelatinous": AnatomyBlob,

		// Mechanical
		"robot": AnatomyMechanical, "golem": AnatomyMechanical, "android": AnatomyMechanical,
		"mech": AnatomyMechanical, "drone": AnatomyMechanical, "bot": AnatomyMechanical,
		"construct": AnatomyMechanical, "cyborg": AnatomyMechanical, "automaton": AnatomyMechanical,
		"sentinel": AnatomyMechanical, "machine": AnatomyMechanical,

		// Undead (skeletal/ghostly variants)
		"skeleton": AnatomyUndead, "ghost": AnatomyUndead, "specter": AnatomyUndead,
		"wraith": AnatomyUndead, "lich": AnatomyUndead, "phantom": AnatomyUndead,
		"shade": AnatomyUndead, "banshee": AnatomyUndead,

		// Multi-limbed horrors
		"kraken": AnatomyMultiLimbed, "octopus": AnatomyMultiLimbed, "squid": AnatomyMultiLimbed,
		"shoggoth": AnatomyMultiLimbed, "abomination": AnatomyMultiLimbed,
		"horror": AnatomyMultiLimbed, "eldritch": AnatomyMultiLimbed,
		"aberration": AnatomyMultiLimbed, "hydra": AnatomyMultiLimbed,
		"chimera": AnatomyMultiLimbed, "tentacle": AnatomyMultiLimbed,
		"thing": AnatomyMultiLimbed, "nightmare": AnatomyMultiLimbed,

		// Humanoid indicators (fallback)
		"orc": AnatomyHumanoid, "goblin": AnatomyHumanoid, "troll": AnatomyHumanoid,
		"ogre": AnatomyHumanoid, "giant": AnatomyHumanoid, "minotaur": AnatomyHumanoid,
		"zombie": AnatomyHumanoid, "ghoul": AnatomyHumanoid, "vampire": AnatomyHumanoid,
		"demon": AnatomyHumanoid, "imp": AnatomyHumanoid, "kobold": AnatomyHumanoid,
		"merchant": AnatomyHumanoid, "guard": AnatomyHumanoid, "priest": AnatomyHumanoid,
		"wizard": AnatomyHumanoid, "warrior": AnatomyHumanoid, "knight": AnatomyHumanoid,
	}
}

// SetGenre sets the genre for genre-specific anatomy variations.
func (s *CreatureAnatomySystem) SetGenre(genreID string) {
	s.genreID = genreID
	s.logger = s.logger.WithField("genre", genreID)
}

// Update processes entities and assigns anatomy types.
func (s *CreatureAnatomySystem) Update(entities []*Entity, deltaTime float64) {
	for _, entity := range entities {
		// Skip entities without sprite component
		if !entity.HasComponent("sprite") {
			continue
		}

		// Check if already has anatomy component
		comp, hasComp := entity.GetComponent("creature_anatomy")
		if !hasComp {
			// Create and assign anatomy component
			anatomy := NewCreatureAnatomyComponent()
			anatomy.Seed = s.seed ^ int64(entity.ID)
			entity.AddComponent(anatomy)
			comp = anatomy
		}

		existing, ok := comp.(*CreatureAnatomyComponent)
		if !ok || existing.Assigned {
			continue
		}

		// Determine anatomy from entity properties
		s.assignAnatomy(existing, entity)
		existing.Assigned = true
	}
}

// assignAnatomy determines the anatomy type from entity name, tags, and components.
func (s *CreatureAnatomySystem) assignAnatomy(anatomy *CreatureAnatomyComponent, entity *Entity) {
	// Collect keywords from entity
	keywords := s.collectKeywords(entity)

	// Find best matching anatomy type
	anatomy.AnatomyType = s.matchAnatomy(keywords, anatomy.Seed)
	anatomy.GenreVariant = s.genreID

	// Determine size modifier from entity
	anatomy.SizeModifier = s.determineSizeModifier(entity)

	// Find subtype for more specific rendering
	anatomy.SubType = s.determineSubType(keywords, anatomy.AnatomyType)

	s.logger.WithFields(logrus.Fields{
		"entity_id":    entity.ID,
		"anatomy_type": anatomy.AnatomyType.String(),
		"sub_type":     anatomy.SubType,
		"size":         anatomy.SizeModifier,
	}).Debug("assigned creature anatomy")
}

// collectKeywords extracts searchable keywords from entity properties.
func (s *CreatureAnatomySystem) collectKeywords(entity *Entity) []string {
	keywords := make([]string, 0, 10)

	// Extract from name component
	if nameComp, ok := entity.GetComponent("name"); ok {
		if nc, ok := nameComp.(*NameComponent); ok && nc.Name != "" {
			// Split name into words and lowercase
			words := strings.Fields(strings.ToLower(nc.Name))
			keywords = append(keywords, words...)
		}
	}

	// Extract from tag component
	if tagComp, ok := entity.GetComponent("tag"); ok {
		if tc, ok := tagComp.(*TagComponent); ok {
			for _, tag := range tc.Tags {
				keywords = append(keywords, strings.ToLower(tag))
			}
		}
	}

	// Check for creature type component
	if ctComp, ok := entity.GetComponent("creature_type"); ok {
		if ct, ok := ctComp.(*CreatureTypeComponent); ok && ct.CreatureType != "" {
			keywords = append(keywords, strings.ToLower(ct.CreatureType))
		}
	}

	return keywords
}

// matchAnatomy finds the best anatomy type match from keywords.
func (s *CreatureAnatomySystem) matchAnatomy(keywords []string, seed int64) AnatomyType {
	// Count matches per anatomy type
	scores := make(map[AnatomyType]int)

	for _, keyword := range keywords {
		if anatomy, found := s.keywords[keyword]; found {
			scores[anatomy]++
		}
	}

	// Find highest scoring type
	bestType := AnatomyHumanoid
	bestScore := 0

	for anatomy, score := range scores {
		if score > bestScore {
			bestScore = score
			bestType = anatomy
		}
	}

	// If no match found, use seed-based fallback for monsters
	if bestScore == 0 {
		// Check if entity seems like a creature vs humanoid NPC
		isCreature := false
		for _, kw := range keywords {
			if kw == "monster" || kw == "boss" || kw == "creature" || kw == "minion" {
				isCreature = true
				break
			}
		}

		if isCreature {
			// Deterministic random fallback for variety
			rng := rand.New(rand.NewSource(seed))
			creatureTypes := []AnatomyType{
				AnatomyQuadruped, AnatomyInsect, AnatomyArachnid,
				AnatomySerpentine, AnatomyMultiLimbed,
			}
			bestType = creatureTypes[rng.Intn(len(creatureTypes))]
		}
	}

	return bestType
}

// determineSizeModifier returns a size multiplier based on entity properties.
func (s *CreatureAnatomySystem) determineSizeModifier(entity *Entity) float64 {
	// Check for creature size proportion component - use WidthScale as proxy
	if sizeComp, ok := entity.GetComponent("creature_size_proportion"); ok {
		if csp, ok := sizeComp.(*CreatureSizeProportionComponent); ok {
			return csp.WidthScale
		}
	}

	// Check for EbitenSprite dimensions to estimate size
	if entity.sprite != nil && entity.sprite.Width > 0 {
		// Normalize to 32x32 standard size
		return entity.sprite.Width / 32.0
	}

	return 1.0
}

// determineSubType finds a specific subtype string for template selection.
func (s *CreatureAnatomySystem) determineSubType(keywords []string, anatomy AnatomyType) string {
	// Look for specific creature names that match the anatomy
	for _, keyword := range keywords {
		if mapped, found := s.keywords[keyword]; found && mapped == anatomy {
			return keyword
		}
	}

	// Return generic subtype
	return anatomy.String()
}

// GetTemplateSelector returns the string used to select the anatomical template.
func (c *CreatureAnatomyComponent) GetTemplateSelector() string {
	if c.SubType != "" {
		return c.SubType
	}
	return c.AnatomyType.String()
}

// CreatureTypeComponent stores the creature classification type.
type CreatureTypeComponent struct {
	CreatureType string
	IsHostile    bool
	IsBoss       bool
}

// Type returns the component type identifier.
func (c *CreatureTypeComponent) Type() string { return "creature_type" }

// GetAerialTemplate returns the appropriate aerial template for this anatomy.
func (c *CreatureAnatomyComponent) GetAerialTemplate(direction sprites.Direction) sprites.AnatomicalTemplate {
	return sprites.SelectNonhumanoidAerialTemplate(c.GetTemplateSelector(), c.GenreVariant, direction)
}
