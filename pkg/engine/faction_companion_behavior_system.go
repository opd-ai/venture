package engine

import (
	"math/rand"

	"github.com/sirupsen/logrus"
)

// FactionCompanionBehaviorSystem modifies companion targeting based on owner faction reputation.
// Companions owned by players will not attack allied faction members and prioritize hostile faction enemies.
// This bridges FactionSystem reputation with CompanionAISystem targeting.
type FactionCompanionBehaviorSystem struct {
	world         *World
	factionSystem *FactionSystem
	rng           *rand.Rand
	genreID       string
}

// NewFactionCompanionBehaviorSystem creates a new system bridging faction and companion behavior.
func NewFactionCompanionBehaviorSystem(world *World, seed int64) *FactionCompanionBehaviorSystem {
	logrus.WithFields(logrus.Fields{
		"system_name": "FactionCompanionBehaviorSystem",
		"seed":        seed,
	}).Debug("Creating faction companion behavior system")

	return &FactionCompanionBehaviorSystem{
		world: world,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// SetFactionSystem sets the faction system for reputation lookups.
func (s *FactionCompanionBehaviorSystem) SetFactionSystem(fs *FactionSystem) {
	s.factionSystem = fs
}

// SetGenre sets the genre for genre-aware behavior modifiers.
func (s *FactionCompanionBehaviorSystem) SetGenre(genreID string) {
	s.genreID = genreID
}

// Update processes companion targeting based on owner faction reputation.
func (s *FactionCompanionBehaviorSystem) Update(entities []*Entity, deltaTime float64) {
	if s.factionSystem == nil {
		return
	}

	// Find player entity for faction lookup
	playerEntity := s.findPlayerEntity()
	if playerEntity == nil {
		return
	}

	// Process all companions
	for _, entity := range entities {
		if !entity.HasComponent("companion") {
			continue
		}

		companionCompRaw, ok := entity.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp := companionCompRaw.(*CompanionComponent)

		// Only process companions owned by player
		if companionComp.OwnerID != playerEntity.ID {
			continue
		}

		// Get companion's AI target
		aiCompRaw, ok := entity.GetComponent("ai")
		if !ok {
			continue
		}
		aiComp := aiCompRaw.(*AIComponent)

		// Validate current target
		if aiComp.Target != nil {
			s.validateTarget(aiComp, companionComp)
		}

		// Apply faction-based aggression modifier
		s.applyFactionAggressionModifier(entity, companionComp)
	}
}

// validateTarget clears invalid targets (allied faction members).
func (s *FactionCompanionBehaviorSystem) validateTarget(aiComp *AIComponent, companionComp *CompanionComponent) {
	target := aiComp.Target
	if target == nil {
		return
	}

	// Check if target has faction component
	factionCompRaw, ok := target.GetComponent("faction")
	if !ok {
		return // Target has no faction, allow attack
	}
	factionComp := factionCompRaw.(*FactionComponent)

	// Skip player faction components
	if factionComp.IsPlayerFaction {
		return
	}

	// Get player's reputation with target's faction
	playerRep := s.factionSystem.GetPlayerReputation(factionComp.FactionID)

	// Companion won't attack friendly faction members (reputation >= 51)
	if playerRep >= 51 {
		logrus.WithFields(logrus.Fields{
			"system_name": "FactionCompanionBehaviorSystem",
			"targetID":    target.ID,
			"factionID":   factionComp.FactionID,
			"reputation":  playerRep,
		}).Debug("Companion refusing to attack allied faction member")
		aiComp.Target = nil
		return
	}

	// High loyalty companions also protect neutral faction members
	if companionComp.Loyalty >= 80 && playerRep >= 1 {
		logrus.WithFields(logrus.Fields{
			"system_name": "FactionCompanionBehaviorSystem",
			"targetID":    target.ID,
			"factionID":   factionComp.FactionID,
			"loyalty":     companionComp.Loyalty,
		}).Debug("Loyal companion protecting neutral faction member")
		aiComp.Target = nil
	}
}

// applyFactionAggressionModifier adjusts companion behavior based on faction reputation.
func (s *FactionCompanionBehaviorSystem) applyFactionAggressionModifier(companion *Entity, companionComp *CompanionComponent) {
	// Get companion's position to find nearby enemies
	posCompRaw, ok := companion.GetComponent("position")
	if !ok {
		return
	}
	posComp := posCompRaw.(*PositionComponent)

	// Get current AI state
	aiCompRaw, ok := companion.GetComponent("ai")
	if !ok {
		return
	}
	aiComp := aiCompRaw.(*AIComponent)

	// If already has valid target, skip
	if aiComp.Target != nil {
		return
	}

	// Only aggressive companions seek targets
	if companionComp.Behavior != BehaviorAggressive {
		return
	}

	// Find nearby enemies prioritizing hostile factions
	target := s.findPriorityTarget(posComp, companionComp.Loyalty)
	if target != nil {
		aiComp.Target = target
		logrus.WithFields(logrus.Fields{
			"system_name": "FactionCompanionBehaviorSystem",
			"companionID": companion.ID,
			"targetID":    target.ID,
		}).Debug("Companion acquired faction-priority target")
	}
}

// findPriorityTarget finds the best target based on faction hostility.
func (s *FactionCompanionBehaviorSystem) findPriorityTarget(pos *PositionComponent, loyalty float64) *Entity {
	// Search radius scales with loyalty (loyal companions are more proactive)
	searchRadius := 150.0 + (loyalty * 0.5) // 150-200 range

	entities := s.world.GetEntitiesWith("ai", "position", "faction")
	var bestTarget *Entity
	bestPriority := -1000

	for _, entity := range entities {
		// Skip companions and players
		if entity.HasComponent("companion") || entity.HasComponent("input") {
			continue
		}

		// Check distance
		entityPosRaw, ok := entity.GetComponent("position")
		if !ok {
			continue
		}
		entityPos := entityPosRaw.(*PositionComponent)
		dx := pos.X - entityPos.X
		dy := pos.Y - entityPos.Y
		distSq := dx*dx + dy*dy
		if distSq > searchRadius*searchRadius {
			continue
		}

		// Get faction to determine priority
		factionCompRaw, ok := entity.GetComponent("faction")
		if !ok {
			continue
		}
		factionComp := factionCompRaw.(*FactionComponent)

		if factionComp.IsPlayerFaction {
			continue
		}

		// Calculate priority based on player reputation
		playerRep := s.factionSystem.GetPlayerReputation(factionComp.FactionID)
		priority := s.calculateTargetPriority(playerRep, distSq)

		// Skip friendly targets
		if priority < 0 {
			continue
		}

		if priority > bestPriority {
			bestPriority = priority
			bestTarget = entity
		}
	}

	return bestTarget
}

// calculateTargetPriority determines attack priority based on faction reputation.
// Returns negative value for targets that should not be attacked.
func (s *FactionCompanionBehaviorSystem) calculateTargetPriority(reputation int, distSq float64) int {
	// Base priority from reputation
	var basePriority int

	if reputation <= -50 {
		// Hostile: highest priority (player's enemies)
		basePriority = 100
	} else if reputation <= 0 {
		// Suspicious: medium priority
		basePriority = 50
	} else if reputation <= 50 {
		// Neutral: low priority
		basePriority = 10
	} else {
		// Friendly: don't attack
		return -1
	}

	// Distance bonus: closer enemies have higher priority
	distBonus := int(10000 / (distSq + 100))

	// Genre-specific modifiers
	genreBonus := 0
	switch s.genreID {
	case "horror":
		// Horror: companions are more defensive, reduce priority
		basePriority = basePriority * 7 / 10
	case "cyberpunk":
		// Cyberpunk: companions are more calculated
		genreBonus = distBonus / 2
	case "fantasy":
		// Fantasy: balanced
		genreBonus = 5
	}

	return basePriority + distBonus + genreBonus
}

// findPlayerEntity locates the player entity.
func (s *FactionCompanionBehaviorSystem) findPlayerEntity() *Entity {
	entities := s.world.GetEntitiesWith("input")
	if len(entities) > 0 {
		return entities[0]
	}
	return nil
}
