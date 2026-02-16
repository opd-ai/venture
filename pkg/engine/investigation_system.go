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
	logger         *log.Entry
}

// NewInvestigationSystem creates a new investigation system.
func NewInvestigationSystem(world *World, seed int64) *InvestigationSystem {
	logger := log.WithFields(log.Fields{
		"system_name": "investigation",
	})

	if logger.Logger.GetLevel() >= log.DebugLevel {
		logger.WithField("seed", seed).Debug("Creating investigation system")
	}

	return &InvestigationSystem{
		world:          world,
		rng:            rand.New(rand.NewSource(seed)),
		fragmentHidden: make(map[uint64]bool),
		logger:         logger,
	}
}

// Update processes all active investigations and reveals nearby fragments.
func (s *InvestigationSystem) Update(deltaTime float64) {
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
		s.logger.WithField("entity_id", entity.ID).Warn("Invalid investigation component type")
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
		s.logger.WithField("entity_id", entity.ID).Warn("Investigator missing position component")
		return
	}

	pos, ok := posComp.(*PositionComponent)
	if !ok {
		s.logger.WithField("entity_id", entity.ID).Warn("Invalid position component type")
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

	if revealedCount > 0 && s.logger.Logger.GetLevel() >= log.InfoLevel {
		s.logger.WithFields(log.Fields{
			"investigator": entity.ID,
			"revealed":     revealedCount,
			"radius":       radius,
		}).Info("Investigation revealed fragments")
	}
}

// tryRevealFragment attempts to reveal a single fragment if within range and detection succeeds.
func (s *InvestigationSystem) tryRevealFragment(investigator, fragment *Entity, investigatorPos *PositionComponent, investigation *InvestigationComponent, radius float64) bool {
	fragPos, ok := s.getFragmentPosition(fragment)
	if !ok {
		return false
	}

	if !s.isFragmentWithinRange(investigatorPos, fragPos, radius) {
		return false
	}

	storyFragComp, ok := s.getStoryFragmentComponent(fragment)
	if !ok {
		return false
	}

	if !s.canRevealFragment(fragment.ID, investigation) {
		return false
	}

	if !s.rollDetectionCheck(investigation) {
		return false
	}

	s.revealFragment(fragment, investigator, investigation)

	if s.logger.Logger.GetLevel() >= log.InfoLevel {
		s.logger.WithFields(log.Fields{
			"investigator": investigator.ID,
			"fragment":     fragment.ID,
			"seriesID":     storyFragComp.SeriesID,
			"sequence":     storyFragComp.SequenceNum,
		}).Info("Fragment revealed through investigation")
	}

	return true
}

// getFragmentPosition retrieves the position component from a fragment entity.
func (s *InvestigationSystem) getFragmentPosition(fragment *Entity) (*PositionComponent, bool) {
	fragPosComp, ok := fragment.GetComponent("position")
	if !ok {
		return nil, false
	}

	fragPos, ok := fragPosComp.(*PositionComponent)
	if !ok {
		return nil, false
	}

	return fragPos, true
}

// isFragmentWithinRange checks if a fragment is within investigation range.
func (s *InvestigationSystem) isFragmentWithinRange(investigatorPos, fragPos *PositionComponent, radius float64) bool {
	dx := investigatorPos.X - fragPos.X
	dy := investigatorPos.Y - fragPos.Y
	distance := math.Sqrt(dx*dx + dy*dy)
	radiusPixels := radius * 32.0

	return distance <= radiusPixels
}

// getStoryFragmentComponent retrieves the story fragment component.
func (s *InvestigationSystem) getStoryFragmentComponent(fragment *Entity) (*StoryFragmentComponent, bool) {
	fragComp, ok := fragment.GetComponent("storyfragment")
	if !ok {
		return nil, false
	}

	storyFragComp, ok := fragComp.(*StoryFragmentComponent)
	if !ok {
		return nil, false
	}

	return storyFragComp, true
}

// canRevealFragment checks if a fragment can be revealed by the investigator.
func (s *InvestigationSystem) canRevealFragment(fragmentID uint64, investigation *InvestigationComponent) bool {
	if !s.IsFragmentHidden(fragmentID) {
		return false
	}

	if investigation.HasRevealedFragment(fragmentID) {
		return false
	}

	return true
}

// rollDetectionCheck performs a detection roll against the investigation chance.
func (s *InvestigationSystem) rollDetectionCheck(investigation *InvestigationComponent) bool {
	detectionChance := investigation.GetDetectionChance()
	roll := s.rng.Float64()
	return roll <= detectionChance
}

// revealFragment marks a fragment as revealed and creates a visual/audio feedback event.
func (s *InvestigationSystem) revealFragment(fragment, investigator *Entity, investigation *InvestigationComponent) {
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
		s.logger.WithField("entity_id", entity.ID).Warn("Invalid investigation component type")
		return false
	}

	started := investigation.StartInvestigation()

	if started && s.logger.Logger.GetLevel() >= log.InfoLevel {
		s.logger.WithField("entity_id", entity.ID).Info("Investigation started")
	}

	return started
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
		s.logger.WithField("percentage", percentageHidden).Warn("Invalid percentage for hiding fragments, must be between 0.0 and 1.0")
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

	if s.logger.Logger.GetLevel() >= log.InfoLevel {
		s.logger.WithFields(log.Fields{
			"totalFragments": len(fragments),
			"hiddenCount":    hiddenCount,
			"percentage":     percentageHidden,
		}).Info("Random fragments hidden for investigation mechanic")
	}
}
