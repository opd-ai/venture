// Package engine provides the NpcRoleVisualSystem which infers a humanoid
// entity's visual role from its components (class, merchant status, entity name)
// and attaches an NpcRoleVisualComponent. Downstream systems (AnimationSystem,
// DirectionalSpriteSystem) read this component to select role-specific aerial
// anatomy templates, making mages, warriors, merchants, etc. visually distinct.
package engine

import (
	"math/rand"
	"strings"

	"github.com/opd-ai/venture/pkg/class/advanced"
	"github.com/sirupsen/logrus"
)

// NpcRoleVisualSystem scans entities that have sprite components and infers
// their visual role from class, merchant, and naming data.
type NpcRoleVisualSystem struct {
	world  *World
	logger *logrus.Entry
	rng    *rand.Rand

	scanInterval  float64
	timeSinceScan float64
}

// NewNpcRoleVisualSystem creates a new NPC role visual system.
func NewNpcRoleVisualSystem(world *World, seed int64) *NpcRoleVisualSystem {
	logger := logrus.WithFields(logrus.Fields{
		"system_name": "npc_role_visual",
	})
	logger.Debug("npc role visual system created")

	return &NpcRoleVisualSystem{
		world:         world,
		logger:        logger,
		rng:           rand.New(rand.NewSource(seed)),
		scanInterval:  2.0,
		timeSinceScan: 0,
	}
}

// Update scans entities and infers visual roles.
func (s *NpcRoleVisualSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceScan += deltaTime
	if s.timeSinceScan < s.scanInterval {
		return
	}
	s.timeSinceScan = 0

	for _, entity := range entities {
		if entity.GetSprite() == nil {
			continue
		}
		// Skip if already classified
		if _, has := entity.GetComponent("npc_role_visual"); has {
			continue
		}
		// Only classify humanoid-like entities (those with health but not creatures)
		if !s.isHumanoidCandidate(entity) {
			continue
		}

		role := s.inferRole(entity)
		if role != "" {
			entity.AddComponent(NewNpcRoleVisualComponent(role))
		}
	}
}

// isHumanoidCandidate returns true if the entity should get a role classification.
func (s *NpcRoleVisualSystem) isHumanoidCandidate(entity *Entity) bool {
	// Creature visual classified as non-humanoid → skip
	if cvComp, ok := entity.GetComponent("creature_visual"); ok {
		if cv, ok := cvComp.(*CreatureVisualComponent); ok && cv.Form != FormHumanoid {
			return false
		}
	}
	// Must have health (eliminates terrain/decoration entities)
	return entity.HasComponent("health")
}

// inferRole determines the visual role from entity components.
func (s *NpcRoleVisualSystem) inferRole(entity *Entity) string {
	// 1. Merchant component → merchant
	if entity.HasComponent("merchant") {
		return "merchant"
	}

	// 2. Advanced class component → map class to role
	if classComp, ok := entity.GetComponent("advanced_class"); ok {
		if ac, ok := classComp.(*advanced.AdvancedClassComponent); ok {
			if role := s.classToRole(ac.PrimaryClass); role != "" {
				return role
			}
		}
	}

	// 3. Name-based heuristic for NPCs with merchant names
	if role := s.inferRoleFromName(entity); role != "" {
		return role
	}

	// 4. Stats-based heuristic (high magic → mage, high evasion → rogue, etc.)
	return s.inferRoleFromStats(entity)
}

// classToRole maps an advanced class ID to a visual role string.
func (s *NpcRoleVisualSystem) classToRole(classID advanced.ClassID) string {
	switch classID {
	case advanced.ClassMage, advanced.ClassElementalist, advanced.ClassNecromancer, advanced.ClassEnchanter:
		return "mage"
	case advanced.ClassWarrior, advanced.ClassBerserker:
		return "warrior"
	case advanced.ClassKnight, advanced.ClassPaladin:
		return "knight"
	case advanced.ClassRogue, advanced.ClassAssassin, advanced.ClassNinja:
		return "rogue"
	case advanced.ClassRanger:
		return "ranger"
	case advanced.ClassCleric, advanced.ClassBard, advanced.ClassDruid:
		return "priest"
	}
	return ""
}

// inferRoleFromName checks entity display name for role keywords.
func (s *NpcRoleVisualSystem) inferRoleFromName(entity *Entity) string {
	name := s.getEntityDisplayName(entity)
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)

	roleKeywords := []struct {
		role     string
		keywords []string
	}{
		{"merchant", []string{"merchant", "shopkeep", "vendor", "trader", "peddler"}},
		{"mage", []string{"mage", "wizard", "sorcerer", "witch", "warlock", "enchant"}},
		{"warrior", []string{"warrior", "fighter", "gladiator", "berserker"}},
		{"knight", []string{"knight", "paladin", "guard", "sentinel", "templar"}},
		{"rogue", []string{"rogue", "thief", "assassin", "ninja", "shadow"}},
		{"ranger", []string{"ranger", "hunter", "scout", "archer", "tracker"}},
		{"priest", []string{"priest", "cleric", "healer", "druid", "monk", "bard"}},
	}

	for _, rk := range roleKeywords {
		for _, kw := range rk.keywords {
			if strings.Contains(lower, kw) {
				return rk.role
			}
		}
	}
	return ""
}

// getEntityDisplayName attempts to retrieve a display name from the entity.
func (s *NpcRoleVisualSystem) getEntityDisplayName(entity *Entity) string {
	// Try merchant component name
	if merchComp, ok := entity.GetComponent("merchant"); ok {
		if merch, ok := merchComp.(*MerchantComponent); ok && merch.MerchantName != "" {
			return merch.MerchantName
		}
	}
	return ""
}

// inferRoleFromStats guesses a role from combat stat distribution.
func (s *NpcRoleVisualSystem) inferRoleFromStats(entity *Entity) string {
	statsComp, ok := entity.GetComponent("stats")
	if !ok {
		return ""
	}
	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return ""
	}

	// High magic → mage, high attack → warrior, high evasion → rogue, high defense → knight
	if stats.MagicPower > stats.Attack*1.5 {
		return "mage"
	}
	if stats.Evasion > 0.3 {
		return "rogue"
	}
	if stats.BlockChance > 0.3 {
		return "knight"
	}
	if stats.Attack > stats.MagicPower*1.5 {
		return "warrior"
	}
	return ""
}

// GetProcessedCount returns the number of entities that have been role-classified.
func (s *NpcRoleVisualSystem) GetProcessedCount() int {
	if s.world == nil {
		return 0
	}
	count := 0
	for _, entity := range s.world.GetEntities() {
		if entity.HasComponent("npc_role_visual") {
			count++
		}
	}
	return count
}
