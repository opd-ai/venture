package engine

import (
	"github.com/sirupsen/logrus"
)

// ReputationSystem tracks player actions and updates faction standings.
// It processes deeds and applies their impacts to reputation and alignment.
type ReputationSystem struct {
	world  *World
	logger *logrus.Logger
}

// NewReputationSystem creates a new ReputationSystem.
func NewReputationSystem(world *World, logger *logrus.Logger) *ReputationSystem {
	if logger == nil {
		logger = logrus.New()
	}

	return &ReputationSystem{
		world:  world,
		logger: logger,
	}
}

// Update processes reputation changes for entities.
// This is called every frame but only processes significant changes.
func (s *ReputationSystem) Update(deltaTime float64) {
	// Reputation system doesn't need per-frame updates
	// Changes are triggered by game events through RecordAction method
}

// RecordAction records a player action that affects reputation and alignment.
// This method creates a Deed and applies it to the entity's ReputationComponent.
//
// Parameters:
//   - entityID: The entity performing the action
//   - description: Human-readable description of the action
//   - factionImpact: Map of faction names to reputation changes
//   - lawImpact: Change to law axis (-1.0 to +1.0)
//   - goodImpact: Change to good axis (-1.0 to +1.0)
//   - location: Optional location where the action occurred
func (s *ReputationSystem) RecordAction(
	entityID uint64,
	description string,
	factionImpact map[string]float64,
	lawImpact, goodImpact float64,
	location string,
) {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		s.logger.WithField("entityID", entityID).Warn("Entity not found for reputation update")
		return
	}

	// Get or create reputation component
	reputationComp := s.getOrCreateReputationComponent(entity)

	// Create deed using the game clock for deterministic timestamps
	deed := Deed{
		Description:   description,
		Timestamp:     s.world.Clock.Now(),
		FactionImpact: factionImpact,
		LawImpact:     lawImpact,
		GoodImpact:    goodImpact,
		Location:      location,
	}

	// Record the deed (applies impacts automatically)
	reputationComp.RecordDeed(deed)

	s.logger.WithFields(logrus.Fields{
		"entityID":    entityID,
		"description": description,
		"alignment":   reputationComp.Alignment.String(),
	}).Debug("Recorded deed")
}

// RecordKill records the killing of an entity and applies reputation/alignment impacts.
// The impact depends on the victim's faction and whether the kill was justified.
func (s *ReputationSystem) RecordKill(killerID, victimID uint64, justified bool) {
	killer, okKiller := s.world.GetEntity(killerID)
	victim, okVictim := s.world.GetEntity(victimID)

	if !okKiller || killer == nil || !okVictim || victim == nil {
		return
	}

	// Get victim's faction (if any)
	victimFaction := s.getEntityFaction(victim)

	// Determine impacts based on victim type and justification
	factionImpact := make(map[string]float64)
	lawImpact := 0.0
	goodImpact := 0.0

	if victimFaction != "" {
		// Killing faction members affects reputation with that faction
		if justified {
			// Justified kills (defending self, bounty target) have less impact
			factionImpact[victimFaction] = -5.0
			lawImpact = 0.01   // Slight lawful shift (upholding justice)
			goodImpact = -0.01 // Slight evil shift (still took a life)
		} else {
			// Unjustified kills are murder
			factionImpact[victimFaction] = -20.0
			lawImpact = -0.05  // Chaotic shift (breaking rules)
			goodImpact = -0.05 // Evil shift (murder)
		}
	} else {
		// Killing non-faction entities (monsters, wildlife)
		if justified {
			// Self-defense: slightly lawful (defending oneself), slight evil (took a life)
			lawImpact = 0.01
			goodImpact = -0.01
		} else {
			// Unprovoked killing: chaotic (breaking norms) and evil (cruelty)
			lawImpact = -0.05
			goodImpact = -0.05
		}
	}

	description := "Killed entity"
	if justified {
		description += " (justified)"
	}

	s.RecordAction(killerID, description, factionImpact, lawImpact, goodImpact, "")
}

// RecordHelp records helping an entity (healing, giving items, etc.)
func (s *ReputationSystem) RecordHelp(helperID, targetID uint64) {
	helper, okHelper := s.world.GetEntity(helperID)
	target, okTarget := s.world.GetEntity(targetID)

	if !okHelper || helper == nil || !okTarget || target == nil {
		return
	}

	targetFaction := s.getEntityFaction(target)

	factionImpact := make(map[string]float64)
	if targetFaction != "" {
		factionImpact[targetFaction] = 5.0 // Helping faction members improves standing
	}

	s.RecordAction(
		helperID,
		"Helped entity",
		factionImpact,
		0.01, // Slight lawful shift (following social rules)
		0.02, // Good shift (altruistic action)
		"",
	)
}

// RecordTheft records stealing an item, affecting reputation negatively.
func (s *ReputationSystem) RecordTheft(thiefID uint64, victimFaction string, value float64) {
	// Theft impact scales with value
	reputationLoss := -10.0 - (value / 100.0)
	if reputationLoss < -30.0 {
		reputationLoss = -30.0 // Cap at -30
	}

	factionImpact := make(map[string]float64)
	if victimFaction != "" {
		factionImpact[victimFaction] = reputationLoss
	}

	s.RecordAction(
		thiefID,
		"Theft",
		factionImpact,
		-0.05, // Chaotic shift (breaking laws)
		-0.03, // Evil shift (harming others)
		"",
	)
}

// RecordQuestCompletion records completing a quest for a faction.
func (s *ReputationSystem) RecordQuestCompletion(entityID uint64, faction string, difficulty float64) {
	// Reputation gain scales with difficulty
	reputationGain := 10.0 + (difficulty * 20.0)
	if reputationGain > 50.0 {
		reputationGain = 50.0 // Cap at +50
	}

	factionImpact := map[string]float64{
		faction: reputationGain,
	}

	s.RecordAction(
		entityID,
		"Completed quest for "+faction,
		factionImpact,
		0.02, // Lawful shift (honoring agreements)
		0.01, // Good shift (helping others)
		"",
	)
}

// GetReputation returns the reputation of an entity with a faction.
// If the entity has no ReputationComponent, returns 0 (neutral).
func (s *ReputationSystem) GetReputation(entityID uint64, faction string) float64 {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return 0.0
	}

	comp, ok := entity.GetComponent("reputation")
	if !ok || comp == nil {
		return 0.0
	}

	reputationComp, ok := comp.(*ReputationComponent)
	if !ok {
		return 0.0
	}

	return reputationComp.GetReputation(faction)
}

// GetAlignment returns the alignment of an entity.
// If the entity has no ReputationComponent, returns True Neutral (0,0).
func (s *ReputationSystem) GetAlignment(entityID uint64) Alignment {
	entity, ok := s.world.GetEntity(entityID)
	if !ok || entity == nil {
		return Alignment{LawAxis: 0.0, GoodAxis: 0.0}
	}

	comp, ok := entity.GetComponent("reputation")
	if !ok || comp == nil {
		return Alignment{LawAxis: 0.0, GoodAxis: 0.0}
	}

	reputationComp, ok := comp.(*ReputationComponent)
	if !ok {
		return Alignment{LawAxis: 0.0, GoodAxis: 0.0}
	}

	return reputationComp.Alignment
}

// getOrCreateReputationComponent gets or creates a ReputationComponent for an entity.
func (s *ReputationSystem) getOrCreateReputationComponent(entity *Entity) *ReputationComponent {
	comp, ok := entity.GetComponent("reputation")
	if ok && comp != nil {
		if reputationComp, ok := comp.(*ReputationComponent); ok {
			return reputationComp
		}
	}

	// Create new component
	reputationComp := NewReputationComponent()
	entity.AddComponent(reputationComp)
	return reputationComp
}

// getEntityFaction returns the faction of an entity (if any).
// Returns the FactionID from the entity's FactionComponent, or empty string.
func (s *ReputationSystem) getEntityFaction(entity *Entity) string {
	comp, ok := entity.GetComponent("faction")
	if !ok || comp == nil {
		return ""
	}

	factionComp, ok := comp.(*FactionComponent)
	if !ok {
		return ""
	}

	return factionComp.FactionID
}
