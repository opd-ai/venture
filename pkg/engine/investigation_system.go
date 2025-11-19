package engine

import (
	"fmt"
	"math"
	"math/rand"

	log "github.com/sirupsen/logrus"
)

// InvestigationSystem processes player investigation actions.
// It handles the examination of the environment to reveal hidden story fragments,
// clues, and environmental details. Players initiate investigations and the system
// scans the nearby area, revealing fragments and marking areas as explored.
type InvestigationSystem struct {
	world          *World
	rng            *rand.Rand      // For detection chance rolls
	fragmentHidden map[uint64]bool // Fragment entity ID → is hidden (revealed through investigation)
}

// NewInvestigationSystem creates a new investigation system.
func NewInvestigationSystem(world *World, seed int64) *InvestigationSystem {
	log.WithFields(log.Fields{
		"system_name": "investigation",
		"seed":        seed,
	}).Debug("Creating investigation system")

	system := &InvestigationSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		fragmentHidden: make(map[uint64]bool),
	}

	log.WithFields(log.Fields{
		"system_name": "investigation",
	}).Debug("Investigation system created")

	return system
}

// Update processes all active investigations and reveals nearby fragments.
func (s *InvestigationSystem) Update(deltaTime float64) {
	// Get all entities with investigation component
	investigators := s.world.GetEntitiesWith("investigation", "position")

	log.WithFields(log.Fields{
		"system_name":        "investigation",
		"delta_time":         deltaTime,
		"investigator_count": len(investigators),
	}).Debug("Starting investigation system update")

	for _, entity := range investigators {
		s.updateInvestigator(entity, deltaTime)
	}

	log.WithFields(log.Fields{
		"system_name":     "investigation",
		"processed_count": len(investigators),
	}).Debug("Investigation system update completed")
}

// updateInvestigator processes a single investigator entity.
func (s *InvestigationSystem) updateInvestigator(entity *Entity, deltaTime float64) {
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"system_name": "investigation",
		"delta_time":  deltaTime,
	}).Debug("Updating investigator entity")

	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Debug("Entity missing investigation component")
		return
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Warn("Invalid investigation component type")
		return
	}

	// Update cooldown timer
	investigation.Update(deltaTime)

	// Check if investigation is active and complete
	if investigation.IsInvestigating && investigation.IsInvestigationComplete() {
		log.WithFields(log.Fields{
			"entity_id":   entity.ID,
			"system_name": "investigation",
		}).Debug("Investigation complete, processing results")
		s.processInvestigation(entity, investigation)
		investigation.StopInvestigation()
		log.WithFields(log.Fields{
			"entity_id":   entity.ID,
			"system_name": "investigation",
		}).Debug("Investigation stopped")
	}
}

// processInvestigation scans the area around the investigator and reveals fragments.
func (s *InvestigationSystem) processInvestigation(entity *Entity, investigation *InvestigationComponent) {
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"system_name": "investigation",
	}).Debug("Processing investigation")

	posComp, ok := entity.GetComponent("position")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "position",
		}).Warn("Investigator missing position component")
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "position",
		}).Warn("Invalid position component type")
		return
	}

	// Get effective investigation radius with skill bonus
	radius := investigation.GetEffectiveRadius()

	// Mark the investigated area
	gridX := int(pos.X / 32.0) // Assuming 32-pixel tiles
	gridY := int(pos.Y / 32.0)
	investigation.MarkAreaDiscovered(gridX, gridY)

	log.WithFields(log.Fields{
		"entity_id": entity.ID,
		"grid_x":    gridX,
		"grid_y":    gridY,
		"radius":    radius,
	}).Debug("Marked area as discovered")

	// Find all fragments within radius
	fragments := s.world.GetEntitiesWith("storyfragment", "position")
	revealedCount := 0

	log.WithFields(log.Fields{
		"entity_id":      entity.ID,
		"fragment_count": len(fragments),
		"radius":         radius,
	}).Debug("Scanning for fragments in investigation radius")

	for _, fragment := range fragments {
		if s.tryRevealFragment(entity, fragment, pos, investigation, radius) {
			revealedCount++
		}
	}

	if revealedCount > 0 {
		log.WithFields(log.Fields{
			"investigator": entity.ID,
			"revealed":     revealedCount,
			"radius":       radius,
		}).Info("Investigation revealed fragments")
	} else {
		log.WithFields(log.Fields{
			"investigator": entity.ID,
			"radius":       radius,
		}).Debug("Investigation revealed no fragments")
	}

	log.WithFields(log.Fields{
		"entity_id":      entity.ID,
		"revealed_count": revealedCount,
	}).Debug("Investigation processing completed")
}

// tryRevealFragment attempts to reveal a single fragment if within range and detection succeeds.
func (s *InvestigationSystem) tryRevealFragment(investigator, fragment *Entity, investigatorPos *PositionComponent, investigation *InvestigationComponent, radius float64) bool {
	log.WithFields(log.Fields{
		"investigator_id": investigator.ID,
		"fragment_id":     fragment.ID,
	}).Debug("Attempting to reveal fragment")

	// Get fragment position
	fragPosComp, ok := fragment.GetComponent("position")
	if !ok {
		log.WithFields(log.Fields{
			"fragment_id":    fragment.ID,
			"component_type": "position",
		}).Debug("Fragment missing position component")
		return false
	}

	fragPos, ok := fragPosComp.(*PositionComponent)
	if !ok {
		log.WithFields(log.Fields{
			"fragment_id":    fragment.ID,
			"component_type": "position",
		}).Debug("Invalid fragment position component type")
		return false
	}

	// Check distance
	dx := investigatorPos.X - fragPos.X
	dy := investigatorPos.Y - fragPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Convert radius from tiles to pixels (32 pixels per tile)
	radiusPixels := radius * 32.0

	if distance > radiusPixels {
		log.WithFields(log.Fields{
			"investigator_id": investigator.ID,
			"fragment_id":     fragment.ID,
			"distance":        distance,
			"radius_pixels":   radiusPixels,
		}).Debug("Fragment out of investigation range")
		return false // Out of range
	}

	// Get fragment component
	fragComp, ok := fragment.GetComponent("storyfragment")
	if !ok {
		log.WithFields(log.Fields{
			"fragment_id":    fragment.ID,
			"component_type": "storyfragment",
		}).Debug("Fragment missing storyfragment component")
		return false
	}

	storyFragComp, ok := fragComp.(*StoryFragmentComponent)
	if !ok {
		log.WithFields(log.Fields{
			"fragment_id":    fragment.ID,
			"component_type": "storyfragment",
		}).Debug("Invalid storyfragment component type")
		return false
	}

	// Check if fragment is hidden and not yet revealed by this investigator
	isHidden := s.IsFragmentHidden(fragment.ID)
	alreadyRevealed := investigation.HasRevealedFragment(fragment.ID)

	if !isHidden {
		log.WithFields(log.Fields{
			"fragment_id": fragment.ID,
		}).Debug("Fragment is not hidden")
		return false
	}

	if alreadyRevealed {
		log.WithFields(log.Fields{
			"investigator_id": investigator.ID,
			"fragment_id":     fragment.ID,
		}).Debug("Fragment already revealed by this investigator")
		return false
	}

	// Roll for detection
	detectionChance := investigation.GetDetectionChance()
	roll := s.rng.Float64()

	if roll > detectionChance {
		log.WithFields(log.Fields{
			"investigator_id": investigator.ID,
			"fragment_id":     fragment.ID,
			"roll":            roll,
			"chance":          detectionChance,
		}).Debug("Detection roll failed")
		return false // Detection failed
	}

	// Success! Reveal the fragment
	s.revealFragment(fragment, investigator, investigation)

	log.WithFields(log.Fields{
		"investigator": investigator.ID,
		"fragment":     fragment.ID,
		"seriesID":     storyFragComp.SeriesID,
		"sequence":     storyFragComp.SequenceNum,
		"distance":     distance,
		"roll":         roll,
		"chance":       detectionChance,
	}).Info("Fragment revealed through investigation")

	return true
}

// revealFragment marks a fragment as revealed and creates a visual/audio feedback event.
func (s *InvestigationSystem) revealFragment(fragment, investigator *Entity, investigation *InvestigationComponent) {
	log.WithFields(log.Fields{
		"investigator_id": investigator.ID,
		"fragment_id":     fragment.ID,
		"system_name":     "investigation",
	}).Debug("Revealing fragment")

	// Mark fragment as revealed
	s.fragmentHidden[fragment.ID] = false
	investigation.MarkFragmentRevealed(fragment.ID)

	log.WithFields(log.Fields{
		"investigator_id": investigator.ID,
		"fragment_id":     fragment.ID,
		"revealed_total":  len(investigation.RevealedFragments),
	}).Debug("Fragment marked as revealed")

	// The fragment can now be discovered normally via proximity
	// DiscoverySystem will handle the actual discovery XP award
}

// StartInvestigation initiates an investigation action for an entity.
// Returns true if investigation started successfully, false if on cooldown or invalid entity.
func (s *InvestigationSystem) StartInvestigation(entity *Entity) bool {
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"system_name": "investigation",
	}).Debug("Attempting to start investigation")

	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Debug("Entity missing investigation component")
		return false
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Warn("Invalid investigation component type")
		return false
	}

	started := investigation.StartInvestigation()

	if started {
		log.WithFields(log.Fields{
			"entity_id":   entity.ID,
			"system_name": "investigation",
		}).Info("Investigation started")
	} else {
		log.WithFields(log.Fields{
			"entity_id":   entity.ID,
			"system_name": "investigation",
		}).Debug("Investigation failed to start (likely on cooldown)")
	}

	return started
}

// IsInvestigating returns true if the entity is currently investigating.
func (s *InvestigationSystem) IsInvestigating(entity *Entity) bool {
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"system_name": "investigation",
	}).Debug("Checking if entity is investigating")

	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Debug("Entity missing investigation component")
		return false
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Debug("Invalid investigation component type")
		return false
	}

	isInvestigating := investigation.IsInvestigating

	log.WithFields(log.Fields{
		"entity_id":        entity.ID,
		"is_investigating": isInvestigating,
	}).Debug("Investigation status checked")

	return isInvestigating
}

// SetFragmentHidden marks a fragment as hidden (requires investigation to reveal).
// Should be called during fragment generation or spawning.
func (s *InvestigationSystem) SetFragmentHidden(fragmentID uint64, hidden bool) {
	log.WithFields(log.Fields{
		"fragment_id": fragmentID,
		"hidden":      hidden,
		"system_name": "investigation",
	}).Debug("Setting fragment hidden state")

	s.fragmentHidden[fragmentID] = hidden

	log.WithFields(log.Fields{
		"fragment_id": fragmentID,
		"hidden":      hidden,
	}).Debug("Fragment hidden state updated")
}

// IsFragmentHidden returns true if the fragment is currently hidden.
func (s *InvestigationSystem) IsFragmentHidden(fragmentID uint64) bool {
	log.WithFields(log.Fields{
		"fragment_id": fragmentID,
		"system_name": "investigation",
	}).Debug("Checking if fragment is hidden")

	hidden, exists := s.fragmentHidden[fragmentID]
	result := exists && hidden

	log.WithFields(log.Fields{
		"fragment_id": fragmentID,
		"exists":      exists,
		"hidden":      hidden,
		"result":      result,
	}).Debug("Fragment hidden check completed")

	return result
}

// GetInvestigationProgress returns the investigation progress for an entity.
// Returns (total investigations, revealed fragments, error).
func (s *InvestigationSystem) GetInvestigationProgress(entity *Entity) (int, int, error) {
	log.WithFields(log.Fields{
		"entity_id":   entity.ID,
		"system_name": "investigation",
	}).Debug("Getting investigation progress")

	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Error("Entity has no investigation component")
		return 0, 0, fmt.Errorf("entity has no investigation component")
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		log.WithFields(log.Fields{
			"entity_id":      entity.ID,
			"component_type": "investigation",
		}).Error("Invalid investigation component type")
		return 0, 0, fmt.Errorf("invalid investigation component type")
	}

	totalInvestigations := investigation.TotalInvestigations
	revealedCount := len(investigation.RevealedFragments)

	log.WithFields(log.Fields{
		"entity_id":            entity.ID,
		"total_investigations": totalInvestigations,
		"revealed_fragments":   revealedCount,
	}).Debug("Investigation progress retrieved")

	return totalInvestigations, revealedCount, nil
}

// HideRandomFragments randomly hides a percentage of fragments requiring investigation.
// percentageHidden is a value between 0.0 and 1.0 (e.g., 0.3 for 30% hidden).
func (s *InvestigationSystem) HideRandomFragments(percentageHidden float64) {
	log.WithFields(log.Fields{
		"percentage":  percentageHidden,
		"system_name": "investigation",
	}).Debug("Starting random fragment hiding")

	if percentageHidden < 0.0 || percentageHidden > 1.0 {
		log.WithFields(log.Fields{
			"percentage": percentageHidden,
		}).Warn("Invalid percentage for hiding fragments, must be between 0.0 and 1.0")
		return
	}

	fragments := s.world.GetEntitiesWith("storyfragment")

	log.WithFields(log.Fields{
		"total_fragments": len(fragments),
		"percentage":      percentageHidden,
	}).Debug("Processing fragments for hiding")

	for _, fragment := range fragments {
		roll := s.rng.Float64()
		if roll < percentageHidden {
			s.SetFragmentHidden(fragment.ID, true)
		}
	}

	hiddenCount := 0
	for _, hidden := range s.fragmentHidden {
		if hidden {
			hiddenCount++
		}
	}

	log.WithFields(log.Fields{
		"totalFragments": len(fragments),
		"hiddenCount":    hiddenCount,
		"percentage":     percentageHidden,
	}).Info("Random fragments hidden for investigation mechanic")

	log.WithFields(log.Fields{
		"hidden_count": hiddenCount,
		"system_name":  "investigation",
	}).Debug("Fragment hiding completed")
}
