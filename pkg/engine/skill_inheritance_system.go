package engine

import (
	"math"

	"github.com/sirupsen/logrus"
)

// SkillInheritanceSystem handles companion skill learning from owner.
// Companions observe their owner using skills and gradually learn them
// based on loyalty and proximity.
type SkillInheritanceSystem struct {
	world  *World
	logger *logrus.Entry

	// Track recent skill usage by players
	recentSkillUsage map[uint64]map[string]float64 // playerID -> skillID -> time since use
}

// NewSkillInheritanceSystem creates a new skill inheritance system.
func NewSkillInheritanceSystem(world *World) *SkillInheritanceSystem {
	return NewSkillInheritanceSystemWithLogger(world, nil)
}

// NewSkillInheritanceSystemWithLogger creates a system with a logger.
func NewSkillInheritanceSystemWithLogger(world *World, logger *logrus.Logger) *SkillInheritanceSystem {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "skill_inheritance",
		})
		logEntry.Debug("skill inheritance system created")
	}

	return &SkillInheritanceSystem{
		world:            world,
		logger:           logEntry,
		recentSkillUsage: make(map[uint64]map[string]float64),
	}
}

// Update processes companion skill learning.
func (s *SkillInheritanceSystem) Update(deltaTime float64) {
	if s.world == nil {
		return
	}

	// Decay recent skill usage tracking
	s.decaySkillUsage(deltaTime)

	// Get all companions with skill inheritance
	companions := s.world.GetEntitiesWith("companion", "skillinheritance", "position")

	for _, companion := range companions {
		companionCompRaw, ok := companion.GetComponent("companion")
		if !ok {
			continue
		}
		companionComp := companionCompRaw.(*CompanionComponent)

		skillCompRaw, ok := companion.GetComponent("skillinheritance")
		if !ok {
			continue
		}
		skillComp := skillCompRaw.(*SkillInheritanceComponent)

		posCompRaw, ok := companion.GetComponent("position")
		if !ok {
			continue
		}
		posComp := posCompRaw.(*PositionComponent)

		// Check loyalty requirement
		if companionComp.Loyalty < skillComp.RequiredLoyalty {
			continue
		}

		// Get owner
		owner, ok := s.world.GetEntity(companionComp.OwnerID)
		if !ok || owner == nil {
			continue
		}

		ownerPosRaw, ok := owner.GetComponent("position")
		if !ok {
			continue
		}
		ownerPos := ownerPosRaw.(*PositionComponent)

		// Check if companion is near owner (learning range: 300 units)
		distance := s.distance(posComp, ownerPos)
		if distance > 300.0 {
			continue
		}

		// Process learning from recent skill usage
		s.processLearning(companion, companionComp, skillComp, owner.ID, distance, deltaTime)
	}
}

// processLearning applies skill learning based on recent owner skill usage.
func (s *SkillInheritanceSystem) processLearning(companion *Entity, companionComp *CompanionComponent, skillComp *SkillInheritanceComponent, ownerID uint64, distance float64, deltaTime float64) {
	// Get recent skill usage for owner
	recentSkills, ok := s.recentSkillUsage[ownerID]
	if !ok || len(recentSkills) == 0 {
		return
	}

	// Distance factor: closer = faster learning (1.0 at 0, 0.5 at 300)
	distanceFactor := 1.0 - (distance / 600.0)
	if distanceFactor < 0.5 {
		distanceFactor = 0.5
	}

	// Loyalty factor: higher loyalty = faster learning
	loyaltyFactor := companionComp.Loyalty / 100.0

	// Process each recently used skill
	for skillID, timeSinceUse := range recentSkills {
		// Only learn from skills used very recently (within last 5 seconds)
		if timeSinceUse > 5.0 {
			continue
		}

		// Time factor: fresher = better learning (1.0 at 0s, 0.2 at 5s)
		timeFactor := 1.0 - (timeSinceUse / 5.0)
		if timeFactor < 0.2 {
			timeFactor = 0.2
		}

		// Calculate learning progress for this frame
		learningProgress := skillComp.LearningRate * distanceFactor * loyaltyFactor * timeFactor * deltaTime

		// Apply learning
		fullyLearned := skillComp.AddSkillProgress(skillID, learningProgress)

		if fullyLearned && s.logger != nil {
			s.logger.WithFields(logrus.Fields{
				"companion": companion.ID,
				"skill":     skillID,
				"loyalty":   companionComp.Loyalty,
			}).Info("companion fully learned skill")
		}
	}
}

// RegisterSkillUsage records that a player used a skill.
// This is called by the skill system or combat system when skills are used.
func (s *SkillInheritanceSystem) RegisterSkillUsage(playerID uint64, skillID string) {
	if _, exists := s.recentSkillUsage[playerID]; !exists {
		s.recentSkillUsage[playerID] = make(map[string]float64)
	}

	// Reset the timer for this skill
	s.recentSkillUsage[playerID][skillID] = 0.0
}

// decaySkillUsage increases the time counters and removes old entries.
func (s *SkillInheritanceSystem) decaySkillUsage(deltaTime float64) {
	for playerID, skills := range s.recentSkillUsage {
		for skillID := range skills {
			skills[skillID] += deltaTime

			// Remove skills not used in last 10 seconds
			if skills[skillID] > 10.0 {
				delete(skills, skillID)
			}
		}

		// Remove player entry if no recent skills
		if len(skills) == 0 {
			delete(s.recentSkillUsage, playerID)
		}
	}
}

// distance calculates the Euclidean distance between two positions.
func (s *SkillInheritanceSystem) distance(a, b *PositionComponent) float64 {
	dx := a.X - b.X
	dy := a.Y - b.Y
	return math.Sqrt(dx*dx + dy*dy)
}

// GetCompanionSkills returns the active skills for a companion.
// This can be used by AI or combat systems to determine what abilities the companion can use.
func (s *SkillInheritanceSystem) GetCompanionSkills(companionID uint64) []string {
	if s.world == nil {
		return nil
	}

	companion, ok := s.world.GetEntity(companionID)
	if !ok || companion == nil {
		return nil
	}

	skillCompRaw, ok := companion.GetComponent("skillinheritance")
	if !ok {
		return nil
	}
	skillComp := skillCompRaw.(*SkillInheritanceComponent)

	return skillComp.ActiveSkills
}
