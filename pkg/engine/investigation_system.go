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
	return &InvestigationSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		fragmentHidden: make(map[uint64]bool),
	}
}

// Update processes all active investigations and reveals nearby fragments.
func (s *InvestigationSystem) Update(deltaTime float64) {
	// Get all entities with investigation component
	investigators := s.world.GetEntitiesWith("investigation", "position")

	for _, entity := range investigators {
		s.updateInvestigator(entity, deltaTime)
	}
}

// updateInvestigator processes a single investigator entity.
func (s *InvestigationSystem) updateInvestigator(entity *Entity, deltaTime float64) {
	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		return
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		return
	}

	// Update cooldown timer
	investigation.Update(deltaTime)

	// Check if investigation is active and complete
	if investigation.IsInvestigating && investigation.IsInvestigationComplete() {
		s.processInvestigation(entity, investigation)
		investigation.StopInvestigation()
	}
}

// processInvestigation scans the area around the investigator and reveals fragments.
func (s *InvestigationSystem) processInvestigation(entity *Entity, investigation *InvestigationComponent) {
	posComp, ok := entity.GetComponent("position")
	if !ok {
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		return
	}

	// Get effective investigation radius with skill bonus
	radius := investigation.GetEffectiveRadius()

	// Mark the investigated area
	gridX := int(pos.X / 32.0) // Assuming 32-pixel tiles
	gridY := int(pos.Y / 32.0)
	investigation.MarkAreaDiscovered(gridX, gridY)

	// Find all fragments within radius
	fragments := s.world.GetEntitiesWith("storyfragment", "position")
	revealedCount := 0

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
		}).Debug("Investigation revealed fragments")
	}
}

// tryRevealFragment attempts to reveal a single fragment if within range and detection succeeds.
func (s *InvestigationSystem) tryRevealFragment(investigator, fragment *Entity, investigatorPos *PositionComponent, investigation *InvestigationComponent, radius float64) bool {
	// Get fragment position
	fragPosComp, ok := fragment.GetComponent("position")
	if !ok {
		return false
	}

	fragPos, ok := fragPosComp.(*PositionComponent)
	if !ok {
		return false
	}

	// Check distance
	dx := investigatorPos.X - fragPos.X
	dy := investigatorPos.Y - fragPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)

	// Convert radius from tiles to pixels (32 pixels per tile)
	radiusPixels := radius * 32.0

	if distance > radiusPixels {
		return false // Out of range
	}

	// Get fragment component
	fragComp, ok := fragment.GetComponent("storyfragment")
	if !ok {
		return false
	}

	storyFragComp, ok := fragComp.(*StoryFragmentComponent)
	if !ok {
		return false
	}

	// Check if fragment is hidden and not yet revealed by this investigator
	isHidden := s.IsFragmentHidden(fragment.ID)
	alreadyRevealed := investigation.HasRevealedFragment(fragment.ID)

	if !isHidden || alreadyRevealed {
		return false // Not hidden or already revealed
	}

	// Roll for detection
	detectionChance := investigation.GetDetectionChance()
	roll := s.rng.Float64()

	if roll > detectionChance {
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
	// Mark fragment as revealed
	s.fragmentHidden[fragment.ID] = false
	investigation.MarkFragmentRevealed(fragment.ID)

	// The fragment can now be discovered normally via proximity
	// DiscoverySystem will handle the actual discovery XP award
}

// StartInvestigation initiates an investigation action for an entity.
// Returns true if investigation started successfully, false if on cooldown or invalid entity.
func (s *InvestigationSystem) StartInvestigation(entity *Entity) bool {
	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		return false
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		return false
	}

	return investigation.StartInvestigation()
}

// IsInvestigating returns true if the entity is currently investigating.
func (s *InvestigationSystem) IsInvestigating(entity *Entity) bool {
	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		return false
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		return false
	}

	return investigation.IsInvestigating
}

// SetFragmentHidden marks a fragment as hidden (requires investigation to reveal).
// Should be called during fragment generation or spawning.
func (s *InvestigationSystem) SetFragmentHidden(fragmentID uint64, hidden bool) {
	s.fragmentHidden[fragmentID] = hidden
}

// IsFragmentHidden returns true if the fragment is currently hidden.
func (s *InvestigationSystem) IsFragmentHidden(fragmentID uint64) bool {
	hidden, exists := s.fragmentHidden[fragmentID]
	return exists && hidden
}

// GetInvestigationProgress returns the investigation progress for an entity.
// Returns (total investigations, revealed fragments, error).
func (s *InvestigationSystem) GetInvestigationProgress(entity *Entity) (int, int, error) {
	invComp, ok := entity.GetComponent("investigation")
	if !ok {
		return 0, 0, fmt.Errorf("entity has no investigation component")
	}

	investigation, ok := invComp.(*InvestigationComponent)
	if !ok {
		return 0, 0, fmt.Errorf("invalid investigation component type")
	}

	return investigation.TotalInvestigations, len(investigation.RevealedFragments), nil
}

// HideRandomFragments randomly hides a percentage of fragments requiring investigation.
// percentageHidden is a value between 0.0 and 1.0 (e.g., 0.3 for 30% hidden).
func (s *InvestigationSystem) HideRandomFragments(percentageHidden float64) {
	if percentageHidden < 0.0 || percentageHidden > 1.0 {
		return
	}

	fragments := s.world.GetEntitiesWith("storyfragment")

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
	}).Debug("Random fragments hidden for investigation mechanic")
}
