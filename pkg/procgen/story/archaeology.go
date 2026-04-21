package story

import (
	"fmt"
	"math/rand"

	log "github.com/sirupsen/logrus"

	"github.com/opd-ai/venture/pkg/procgen"
)

// String returns the string representation of ArtifactType
func (a ArtifactType) String() string {
	switch a {
	case ArtifactMagical:
		return "Magical"
	case ArtifactTech:
		return "Technology"
	case ArtifactRitual:
		return "Ritual"
	case ArtifactData:
		return "Data"
	case ArtifactPreWar:
		return "Pre-War"
	case ArtifactRelic:
		return "Relic"
	default:
		return "Unknown"
	}
}

// ArchaeologicalSite represents a location with artifacts
type ArchaeologicalSite struct {
	Name          string
	Genre         string
	Location      Vector2
	Era           string     // Which historical era it's from
	Artifacts     []Artifact // Discoverable artifacts
	Danger        float64    // 0.0-1.0, how hazardous excavation is
	Depth         int        // Dungeon depth where site is found
	Discovered    bool       // Has player found this site?
	Excavation    float64    // 0.0-1.0, excavation progress
	Description   string     // Site description
	SpritePattern string     // Visual representation
}

// Artifact represents a single archaeological find
type Artifact struct {
	Name          string
	Type          ArtifactType
	Description   string
	Age           int64   // Years old
	Condition     float64 // 0.0-1.0, how intact it is
	Value         float64 // Monetary/XP value
	PowerLevel    float64 // Magical/tech power (if applicable)
	Curse         string  // Curse description (if cursed)
	LoreText      string  // Historical information
	SpritePattern string  // Visual pattern
	Functional    bool    // Can it still be used?
}

// ArchaeologyGenerator creates genre-specific archaeological sites
type ArchaeologyGenerator struct{}

// NewArchaeologyGenerator creates a new archaeology generator
func NewArchaeologyGenerator() *ArchaeologyGenerator {
	return &ArchaeologyGenerator{}
}

// Generate creates an archaeological site with artifacts
func (g *ArchaeologyGenerator) Generate(seed int64, params procgen.GenerationParams) (interface{}, error) {
	if params.Difficulty < 0 || params.Difficulty > 1.0 {
		log.WithFields(log.Fields{
			"seed":       seed,
			"difficulty": params.Difficulty,
		}).Error("invalid difficulty parameter for archaeology generation")
		return nil, fmt.Errorf("%w, got %.2f", ErrInvalidDifficulty, params.Difficulty)
	}

	log.WithFields(log.Fields{
		"seed":       seed,
		"genre":      params.GenreID,
		"difficulty": params.Difficulty,
		"depth":      params.Depth,
	}).Debug("generating archaeological site")

	rng := rand.New(rand.NewSource(seed))

	// Generate site name
	siteName := g.generateSiteName(rng, params.GenreID, params.Depth)

	// Determine number of artifacts (2-6 based on depth and difficulty)
	numArtifacts := 2 + int(float64(params.Depth)*0.3) + int(params.Difficulty*3)
	if numArtifacts > 6 {
		numArtifacts = 6
	}

	// Generate artifacts
	artifacts := make([]Artifact, numArtifacts)
	for i := 0; i < numArtifacts; i++ {
		artifacts[i] = g.generateArtifact(rng, params.GenreID, params.Depth, i)
	}

	// Determine danger level (deeper = more dangerous)
	danger := 0.2 + float64(params.Depth)*0.1 + params.Difficulty*0.3
	if danger > 1.0 {
		danger = 1.0
	}

	site := &ArchaeologicalSite{
		Name:          siteName,
		Genre:         params.GenreID,
		Location:      Vector2{X: 50 + rng.Float64()*30, Y: 50 + rng.Float64()*30},
		Era:           g.selectEra(rng, params.GenreID),
		Artifacts:     artifacts,
		Danger:        danger,
		Depth:         params.Depth,
		Discovered:    false,
		Excavation:    0.0,
		Description:   g.generateSiteDescription(rng, params.GenreID, siteName),
		SpritePattern: g.generateSiteSpritePattern(params.GenreID, rng),
	}

	log.WithFields(log.Fields{
		"site_name":     siteName,
		"num_artifacts": numArtifacts,
		"danger":        danger,
		"era":           site.Era,
	}).Info("archaeological site generated")

	return site, nil
}

// Validate checks archaeological site quality
func (g *ArchaeologyGenerator) Validate(result interface{}) error {
	site, ok := result.(*ArchaeologicalSite)
	if !ok {
		log.Error("validation failed: result is not an *ArchaeologicalSite")
		return ErrInvalidType
	}

	if site.Name == "" {
		log.Warn("validation failed: site name is empty")
		return ErrEmptySiteName
	}

	if len(site.Artifacts) < 2 {
		log.WithFields(log.Fields{
			"site_name":     site.Name,
			"num_artifacts": len(site.Artifacts),
		}).Warn("validation failed: too few artifacts")
		return fmt.Errorf("%w: %d, minimum 2", ErrTooFewArtifacts, len(site.Artifacts))
	}

	if len(site.Artifacts) > 6 {
		log.WithFields(log.Fields{
			"site_name":     site.Name,
			"num_artifacts": len(site.Artifacts),
		}).Warn("validation failed: too many artifacts")
		return fmt.Errorf("%w: %d, maximum 6", ErrTooManyArtifacts, len(site.Artifacts))
	}

	if site.Danger < 0 || site.Danger > 1.0 {
		log.WithFields(log.Fields{
			"site_name": site.Name,
			"danger":    site.Danger,
		}).Warn("validation failed: invalid danger level")
		return fmt.Errorf("%w, got %.2f", ErrInvalidDanger, site.Danger)
	}

	// Validate all artifacts
	for i, artifact := range site.Artifacts {
		if artifact.Name == "" {
			log.WithFields(log.Fields{
				"site_name":    site.Name,
				"artifact_num": i,
			}).Warn("validation failed: artifact has empty name")
			return fmt.Errorf("%w at artifact %d", ErrEmptyArtifactName, i)
		}
		if artifact.Condition < 0 || artifact.Condition > 1.0 {
			return fmt.Errorf("%w: artifact %d has condition %.2f", ErrArtifactCondition, i, artifact.Condition)
		}
	}

	return nil
}

// generateSiteName creates a genre-appropriate site name
func (g *ArchaeologyGenerator) generateSiteName(rng *rand.Rand, genreID string, depth int) string {
	prefixes := g.getSiteNamePrefixes(genreID)
	suffixes := g.getSiteNameSuffixes(genreID)

	prefix := prefixes[rng.Intn(len(prefixes))]
	suffix := suffixes[rng.Intn(len(suffixes))]

	return fmt.Sprintf("%s %s", prefix, suffix)
}

// Genre-specific site name components
func (g *ArchaeologyGenerator) getSiteNamePrefixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Ancient", "Forgotten", "Lost", "Buried", "Hidden", "Sacred", "Cursed"}
	case "scifi":
		return []string{"Derelict", "Abandoned", "Alien", "Ancient", "Crashed", "Buried", "Orbital"}
	case "horror":
		return []string{"Haunted", "Cursed", "Unholy", "Damned", "Twisted", "Corrupted", "Forbidden"}
	case "cyberpunk":
		return []string{"Obsolete", "Abandoned", "Underground", "Hidden", "Encrypted", "Lost", "Buried"}
	case "postapoc":
		return []string{"Pre-War", "Ruined", "Abandoned", "Collapsed", "Irradiated", "Lost", "Buried"}
	default:
		return []string{"Ancient", "Forgotten", "Lost", "Hidden", "Buried", "Old"}
	}
}

func (g *ArchaeologyGenerator) getSiteNameSuffixes(genreID string) []string {
	switch genreID {
	case "fantasy":
		return []string{"Temple", "Ruins", "Tomb", "Sanctum", "Vault", "Crypt", "Chamber"}
	case "scifi":
		return []string{"Station", "Facility", "Vessel", "Laboratory", "Outpost", "Probe", "Archive"}
	case "horror":
		return []string{"Asylum", "Mansion", "Chapel", "Shrine", "Lair", "Catacombs", "Ossuary"}
	case "cyberpunk":
		return []string{"Server Farm", "Data Center", "Facility", "Archive", "Vault", "Network Node"}
	case "postapoc":
		return []string{"Bunker", "Vault", "Shelter", "Facility", "Complex", "Installation", "Repository"}
	default:
		return []string{"Site", "Ruins", "Structure", "Building", "Complex", "Location"}
	}
}

// generateArtifact creates a single artifact
func (g *ArchaeologyGenerator) generateArtifact(rng *rand.Rand, genreID string, depth, index int) Artifact {
	artifactType := g.selectArtifactType(genreID)

	// Age increases with depth (100-1000 years)
	age := int64(100 + depth*50 + rng.Intn(300))

	// Condition decreases with age (0.3-0.9)
	condition := 0.9 - float64(depth)*0.05 - rng.Float64()*0.2
	if condition < 0.3 {
		condition = 0.3
	}

	// Value scales with depth and condition
	value := 50.0 + float64(depth)*25.0 + condition*100.0

	// Power level for magical/tech artifacts
	powerLevel := 0.0
	functional := false
	if artifactType == ArtifactMagical || artifactType == ArtifactTech {
		powerLevel = float64(depth)*0.2 + rng.Float64()*0.5
		functional = condition > 0.6 && rng.Float64() < 0.7
	}

	// Some artifacts are cursed (20% for horror, 10% for fantasy)
	curse := ""
	if genreID == "horror" && rng.Float64() < 0.2 {
		curse = g.generateCurse(rng, genreID)
	} else if genreID == "fantasy" && rng.Float64() < 0.1 {
		curse = g.generateCurse(rng, genreID)
	}

	return Artifact{
		Name:          g.generateArtifactName(rng, genreID, artifactType),
		Type:          artifactType,
		Description:   g.generateArtifactDescription(rng, genreID, artifactType),
		Age:           age,
		Condition:     condition,
		Value:         value,
		PowerLevel:    powerLevel,
		Curse:         curse,
		LoreText:      g.generateLoreText(rng, genreID, artifactType, age),
		SpritePattern: g.generateArtifactSpritePattern(genreID, artifactType, rng),
		Functional:    functional,
	}
}

// selectArtifactType chooses appropriate artifact type for genre
func (g *ArchaeologyGenerator) selectArtifactType(genreID string) ArtifactType {
	switch genreID {
	case "fantasy":
		return ArtifactMagical
	case "scifi":
		return ArtifactTech
	case "horror":
		return ArtifactRitual
	case "cyberpunk":
		return ArtifactData
	case "postapoc":
		return ArtifactPreWar
	default:
		return ArtifactRelic
	}
}

// selectEra chooses an era for the site
func (g *ArchaeologyGenerator) selectEra(rng *rand.Rand, genreID string) string {
	timelineGen := NewTimelineGenerator()
	eras := timelineGen.getEraTemplates(genreID)
	return eras[rng.Intn(len(eras))]
}

// generateArtifactName creates genre-specific artifact names
func (g *ArchaeologyGenerator) generateArtifactName(rng *rand.Rand, genreID string, artifactType ArtifactType) string {
	switch genreID {
	case "fantasy":
		items := []string{"Amulet", "Staff", "Orb", "Tome", "Crown", "Ring", "Sword", "Crystal"}
		prefixes := []string{"Ancient", "Arcane", "Enchanted", "Mystic", "Sacred", "Lost", "Forbidden"}
		return fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], items[rng.Intn(len(items))])

	case "scifi":
		items := []string{"Data Core", "Energy Cell", "Scanner", "Interface", "Drive", "Module", "Device", "Beacon"}
		prefixes := []string{"Alien", "Advanced", "Ancient", "Prototype", "Quantum", "Neural", "Photonic"}
		return fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], items[rng.Intn(len(items))])

	case "horror":
		items := []string{"Idol", "Dagger", "Skull", "Book", "Chalice", "Mask", "Talisman", "Bone"}
		prefixes := []string{"Cursed", "Unholy", "Blood-Soaked", "Corrupted", "Damned", "Twisted", "Dark"}
		return fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], items[rng.Intn(len(items))])

	case "cyberpunk":
		items := []string{"Chip", "Drive", "Crystal", "Deck", "Interface", "Cache", "Node", "Key"}
		prefixes := []string{"Encrypted", "Corporate", "Military", "Prototype", "Black Market", "Hacked", "Classified"}
		return fmt.Sprintf("%s Data %s", prefixes[rng.Intn(len(prefixes))], items[rng.Intn(len(items))])

	case "postapoc":
		items := []string{"Manual", "Tool", "Medicine", "Tech", "Weapon", "Supply Cache", "Blueprint", "Generator"}
		prefixes := []string{"Pre-War", "Military", "Medical", "Industrial", "Civilian", "Government", "Emergency"}
		return fmt.Sprintf("%s %s", prefixes[rng.Intn(len(prefixes))], items[rng.Intn(len(items))])

	default:
		items := []string{"Artifact", "Relic", "Object", "Item", "Treasure"}
		return fmt.Sprintf("Ancient %s", items[rng.Intn(len(items))])
	}
}

// generateArtifactDescription creates description text
func (g *ArchaeologyGenerator) generateArtifactDescription(rng *rand.Rand, genreID string, artifactType ArtifactType) string {
	switch genreID {
	case "fantasy":
		return "A magical artifact from a bygone age, still humming with arcane energy."
	case "scifi":
		return "Advanced technology from an ancient civilization, its purpose unclear."
	case "horror":
		return "An unsettling object that seems to whisper in the darkness."
	case "cyberpunk":
		return "Encrypted data from before the corporate wars, potentially valuable."
	case "postapoc":
		return "A relic from before the fall, surprisingly well-preserved."
	default:
		return "An artifact of historical significance, its origins mysterious."
	}
}

// generateLoreText creates historical context
func (g *ArchaeologyGenerator) generateLoreText(rng *rand.Rand, genreID string, artifactType ArtifactType, age int64) string {
	templates := []string{
		"This artifact dates back %d years to the %s.",
		"Created during the %s, approximately %d years ago.",
		"Historical records suggest this is %d years old, from the %s.",
	}

	template := templates[rng.Intn(len(templates))]
	timelineGen := NewTimelineGenerator()
	eras := timelineGen.getEraTemplates(genreID)
	era := eras[rng.Intn(len(eras))]

	return fmt.Sprintf(template, age, era, era, age, age, era)
}

// generateCurse creates a curse description
func (g *ArchaeologyGenerator) generateCurse(rng *rand.Rand, genreID string) string {
	curses := []string{
		"Drains vitality from its bearer over time.",
		"Whispers maddening thoughts to those who touch it.",
		"Brings misfortune to anyone who claims it.",
		"Slowly corrupts the mind of its owner.",
		"Attracts malevolent entities from beyond.",
	}
	return curses[rng.Intn(len(curses))]
}

// generateSiteDescription creates site description
func (g *ArchaeologyGenerator) generateSiteDescription(rng *rand.Rand, genreID, siteName string) string {
	templates := []string{
		"The %s appears abandoned for centuries, yet strangely well-preserved.",
		"This %s holds secrets from a forgotten age.",
		"Ancient and mysterious, the %s beckons exploration.",
		"The %s stands as a testament to a civilization long gone.",
	}

	template := templates[rng.Intn(len(templates))]
	return fmt.Sprintf(template, siteName, siteName, siteName, siteName)
}

// generateSiteSpritePattern creates visual pattern for site
func (g *ArchaeologyGenerator) generateSiteSpritePattern(genreID string, rng *rand.Rand) string {
	switch genreID {
	case "fantasy":
		patterns := []string{"ancient_temple", "magical_ruins", "wizard_tower", "sacred_vault"}
		return patterns[rng.Intn(len(patterns))]
	case "scifi":
		patterns := []string{"alien_structure", "crashed_ship", "ancient_tech", "orbital_debris"}
		return patterns[rng.Intn(len(patterns))]
	case "horror":
		patterns := []string{"cursed_shrine", "dark_altar", "twisted_monument", "unholy_ground"}
		return patterns[rng.Intn(len(patterns))]
	case "cyberpunk":
		patterns := []string{"old_server_room", "data_vault", "abandoned_terminal", "hidden_archive"}
		return patterns[rng.Intn(len(patterns))]
	case "postapoc":
		patterns := []string{"collapsed_bunker", "ruined_facility", "buried_vault", "destroyed_complex"}
		return patterns[rng.Intn(len(patterns))]
	default:
		return "archaeological_site"
	}
}

// generateArtifactSpritePattern creates visual pattern for artifact
func (g *ArchaeologyGenerator) generateArtifactSpritePattern(genreID string, artifactType ArtifactType, rng *rand.Rand) string {
	switch artifactType {
	case ArtifactMagical:
		patterns := []string{"glowing_orb", "ancient_staff", "magic_amulet", "enchanted_tome"}
		return patterns[rng.Intn(len(patterns))]
	case ArtifactTech:
		patterns := []string{"alien_device", "data_core", "tech_module", "energy_cell"}
		return patterns[rng.Intn(len(patterns))]
	case ArtifactRitual:
		patterns := []string{"cursed_idol", "ritual_dagger", "dark_tome", "blood_chalice"}
		return patterns[rng.Intn(len(patterns))]
	case ArtifactData:
		patterns := []string{"data_chip", "memory_crystal", "encrypted_drive", "neural_interface"}
		return patterns[rng.Intn(len(patterns))]
	case ArtifactPreWar:
		patterns := []string{"prewar_tech", "old_manual", "survival_kit", "military_equipment"}
		return patterns[rng.Intn(len(patterns))]
	default:
		return "ancient_artifact"
	}
}

// Excavate progresses the excavation of the site
func (s *ArchaeologicalSite) Excavate(amount float64) []Artifact {
	s.Excavation += amount
	if s.Excavation > 1.0 {
		s.Excavation = 1.0
	}

	// Return artifacts that have been uncovered
	threshold := s.Excavation
	uncovered := make([]Artifact, 0)

	for i, artifact := range s.Artifacts {
		// Each artifact requires a certain excavation level
		requiredProgress := float64(i+1) / float64(len(s.Artifacts))
		if threshold >= requiredProgress && !artifact.Functional { // Use functional as "discovered" flag for simplicity
			uncovered = append(uncovered, artifact)
		}
	}

	return uncovered
}

// IsFullyExcavated checks if all artifacts have been found
func (s *ArchaeologicalSite) IsFullyExcavated() bool {
	return s.Excavation >= 1.0
}

// GetExcavationProgress returns current progress as percentage
func (s *ArchaeologicalSite) GetExcavationProgress() float64 {
	return s.Excavation * 100.0
}
