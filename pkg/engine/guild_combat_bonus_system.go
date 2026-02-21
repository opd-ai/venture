// Package engine provides the GuildCombatBonusSystem which bridges guild membership
// with proximity-based combat bonuses for cooperative gameplay.
package engine

import (
	"math"
	"math/rand"

	"github.com/sirupsen/logrus"
)

// GuildCombatBonusSystem connects guild membership with combat bonuses when
// guild members fight near each other. This creates emergent cooperative gameplay
// where grouping up with guildmates provides meaningful combat advantages.
//
// Integration Points:
// - Reads from: GuildComponent (player guild membership and rank)
// - Reads from: PositionComponent (entity positions for proximity checks)
// - Reads from: HealthComponent (to detect combat via damage taken)
// - Modifies: GuildCombatBonusComponent (stores calculated bonuses)
// - Used by: CombatSystem (via GetDamageMultiplier/GetDefenseMultiplier)
//
// Bonus Tiers (based on nearby guild member count):
// - 1 member: +5% attack, +3% defense, +2% crit
// - 2 members: +10% attack, +6% defense, +4% crit
// - 3 members: +15% attack, +9% defense, +6% crit
// - 4+ members: +20% attack, +12% defense, +8% crit (cap)
//
// Rank Bonuses (when Officer/Leader nearby):
// - Officer nearby: +3% additional all stats
// - Leader nearby: +5% additional all stats
//
// Genre Modifiers:
// - Fantasy: Base multipliers (guild honor matters)
// - Scifi: +10% guild bonuses (team coordination rewarded)
// - Horror: -30% guild bonuses (fear isolates even allies)
// - Cyberpunk: +20% guild bonuses (gang loyalty is power)
// - PostApoc: -15% guild bonuses (trust is rare)
type GuildCombatBonusSystem struct {
	world  *World
	rng    *rand.Rand
	logger *logrus.Entry

	// Configuration
	genreID     string
	bonusRange  float64 // Distance for proximity bonus (default 200)
	updateDelay float64 // Seconds between recalculations

	// Timing
	timeSinceUpdate float64

	// Genre multipliers
	genreMultipliers map[string]float64

	// Rank bonus values
	rankBonuses map[string]float64
}

// NewGuildCombatBonusSystem creates a new guild combat bonus system.
func NewGuildCombatBonusSystem(world *World, seed int64) *GuildCombatBonusSystem {
	logger := world.GetLogger()
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithField("system_name", "guild_combat_bonus")
	}

	return &GuildCombatBonusSystem{
		world:           world,
		rng:             rand.New(rand.NewSource(seed)),
		logger:          logEntry,
		bonusRange:      200.0,
		updateDelay:     0.25, // Update 4 times per second for responsiveness
		timeSinceUpdate: 0,
		genreMultipliers: map[string]float64{
			"fantasy":   1.0,
			"scifi":     1.1,
			"horror":    0.7,
			"cyberpunk": 1.2,
			"postapoc":  0.85,
		},
		rankBonuses: map[string]float64{
			"Recruit": 0.0,
			"Member":  0.01,
			"Officer": 0.03,
			"Leader":  0.05,
		},
	}
}

// SetGenre sets the genre for genre-aware bonus scaling.
func (s *GuildCombatBonusSystem) SetGenre(genreID string) {
	s.genreID = genreID
	if s.logger != nil {
		s.logger.WithField("genre", genreID).Debug("genre set for guild combat bonuses")
	}
}

// SetBonusRange configures the proximity range for guild bonuses.
func (s *GuildCombatBonusSystem) SetBonusRange(bonusRange float64) {
	s.bonusRange = bonusRange
}

// Update recalculates guild combat bonuses for all entities with guild membership.
func (s *GuildCombatBonusSystem) Update(entities []*Entity, deltaTime float64) {
	s.timeSinceUpdate += deltaTime
	if s.timeSinceUpdate < s.updateDelay {
		return
	}
	s.timeSinceUpdate = 0

	// Collect guild members with positions for efficient lookups
	guildMembers := s.collectGuildMembers(entities)

	// Process each player entity with guild membership
	for _, entity := range entities {
		if !s.shouldProcess(entity) {
			continue
		}

		s.updateEntityBonuses(entity, guildMembers)
	}
}

// collectGuildMembers returns a map of guildID -> slice of member entities.
func (s *GuildCombatBonusSystem) collectGuildMembers(entities []*Entity) map[string][]*Entity {
	members := make(map[string][]*Entity)

	for _, entity := range entities {
		guildComp, ok := entity.GetComponent("guild")
		if !ok {
			continue
		}

		gc := guildComp.(*GuildComponent)
		if gc.GuildID == "" {
			continue
		}

		// Must have position for proximity checks
		if _, ok := entity.GetComponent("position"); !ok {
			continue
		}

		members[gc.GuildID] = append(members[gc.GuildID], entity)
	}

	return members
}

// shouldProcess returns true if entity should have guild combat bonuses calculated.
func (s *GuildCombatBonusSystem) shouldProcess(entity *Entity) bool {
	// Must have guild membership
	guildComp, ok := entity.GetComponent("guild")
	if !ok {
		return false
	}

	gc := guildComp.(*GuildComponent)
	if gc.GuildID == "" {
		return false
	}

	// Must have position
	if _, ok := entity.GetComponent("position"); !ok {
		return false
	}

	return true
}

// updateEntityBonuses calculates and applies guild bonuses for a single entity.
func (s *GuildCombatBonusSystem) updateEntityBonuses(entity *Entity, guildMembers map[string][]*Entity) {
	guildComp, _ := entity.GetComponent("guild")
	gc := guildComp.(*GuildComponent)

	posComp, _ := entity.GetComponent("position")
	pos := posComp.(*PositionComponent)

	// Get or create bonus component
	bonusComp := s.getOrCreateBonusComponent(entity)

	// Clear previous bonuses
	bonusComp.ClearBonuses()

	// Get members of same guild
	members, ok := guildMembers[gc.GuildID]
	if !ok || len(members) <= 1 {
		return // No other guild members
	}

	// Count nearby members and calculate bonuses
	nearbyCount := 0
	maxRankBonus := 0.0

	for _, member := range members {
		if member.ID == entity.ID {
			continue // Skip self
		}

		memberPos, _ := member.GetComponent("position")
		mPos := memberPos.(*PositionComponent)

		dist := s.distance(pos.X, pos.Y, mPos.X, mPos.Y)
		if dist <= s.bonusRange {
			nearbyCount++

			// Check rank for rank bonus
			memberGuildComp, _ := member.GetComponent("guild")
			mGc := memberGuildComp.(*GuildComponent)
			rankBonus := s.rankBonuses[mGc.Rank]
			if rankBonus > maxRankBonus {
				maxRankBonus = rankBonus
			}
		}
	}

	if nearbyCount == 0 {
		return // No nearby guild members
	}

	// Calculate bonuses based on nearby count
	bonusComp.NearbyGuildMemberCount = nearbyCount
	bonusComp.AttackBonus = s.calculateAttackBonus(nearbyCount)
	bonusComp.DefenseBonus = s.calculateDefenseBonus(nearbyCount)
	bonusComp.CritBonus = s.calculateCritBonus(nearbyCount)
	bonusComp.HealingBonus = s.calculateHealingBonus(nearbyCount)
	bonusComp.RankBonus = maxRankBonus

	// Apply genre multiplier to all bonuses
	genreMult := s.getGenreMultiplier()
	bonusComp.AttackBonus *= genreMult
	bonusComp.DefenseBonus *= genreMult
	bonusComp.CritBonus *= genreMult
	bonusComp.HealingBonus *= genreMult

	// Log significant bonus application
	if s.logger != nil && bonusComp.HasSignificantBonus() {
		s.logger.WithFields(logrus.Fields{
			"entityID":     entity.ID,
			"guildID":      gc.GuildID,
			"nearbyCount":  nearbyCount,
			"attackBonus":  bonusComp.AttackBonus,
			"defenseBonus": bonusComp.DefenseBonus,
			"critBonus":    bonusComp.CritBonus,
			"rankBonus":    bonusComp.RankBonus,
			"genre":        s.genreID,
		}).Debug("guild combat bonus applied")
	}
}

// getOrCreateBonusComponent gets existing or creates new bonus component.
func (s *GuildCombatBonusSystem) getOrCreateBonusComponent(entity *Entity) *GuildCombatBonusComponent {
	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if ok {
		return compRaw.(*GuildCombatBonusComponent)
	}

	comp := NewGuildCombatBonusComponent()
	comp.BonusRange = s.bonusRange
	entity.AddComponent(comp)
	return comp
}

// calculateAttackBonus returns attack bonus based on nearby guild member count.
// Scales: 1=5%, 2=10%, 3=15%, 4+=20% (capped)
func (s *GuildCombatBonusSystem) calculateAttackBonus(nearbyCount int) float64 {
	if nearbyCount <= 0 {
		return 0.0
	}
	if nearbyCount >= 4 {
		return 0.20
	}
	return float64(nearbyCount) * 0.05
}

// calculateDefenseBonus returns defense bonus based on nearby guild member count.
// Scales: 1=3%, 2=6%, 3=9%, 4+=12% (capped)
func (s *GuildCombatBonusSystem) calculateDefenseBonus(nearbyCount int) float64 {
	if nearbyCount <= 0 {
		return 0.0
	}
	if nearbyCount >= 4 {
		return 0.12
	}
	return float64(nearbyCount) * 0.03
}

// calculateCritBonus returns crit chance bonus based on nearby guild member count.
// Scales: 1=2%, 2=4%, 3=6%, 4+=8% (capped)
func (s *GuildCombatBonusSystem) calculateCritBonus(nearbyCount int) float64 {
	if nearbyCount <= 0 {
		return 0.0
	}
	if nearbyCount >= 4 {
		return 0.08
	}
	return float64(nearbyCount) * 0.02
}

// calculateHealingBonus returns passive healing rate when near guild members.
// Scales: 1=0.5 HP/s, 2=1.0 HP/s, 3=1.5 HP/s, 4+=2.0 HP/s (capped)
func (s *GuildCombatBonusSystem) calculateHealingBonus(nearbyCount int) float64 {
	if nearbyCount <= 0 {
		return 0.0
	}
	if nearbyCount >= 4 {
		return 2.0
	}
	return float64(nearbyCount) * 0.5
}

// getGenreMultiplier returns the genre-specific multiplier.
func (s *GuildCombatBonusSystem) getGenreMultiplier() float64 {
	mult := s.genreMultipliers[s.genreID]
	if mult == 0 {
		return 1.0
	}
	return mult
}

// distance calculates Euclidean distance between two points.
func (s *GuildCombatBonusSystem) distance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// GetDamageMultiplier returns the damage multiplier for an entity from guild bonuses.
// Used by CombatSystem to apply guild-based damage scaling.
func (s *GuildCombatBonusSystem) GetDamageMultiplier(entity *Entity) float64 {
	if entity == nil {
		return 1.0
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return 1.0
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.GetTotalAttackMultiplier()
}

// GetDefenseMultiplier returns the defense multiplier for an entity from guild bonuses.
// Used by CombatSystem to apply guild-based defense scaling.
func (s *GuildCombatBonusSystem) GetDefenseMultiplier(entity *Entity) float64 {
	if entity == nil {
		return 1.0
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return 1.0
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.GetTotalDefenseMultiplier()
}

// GetCritBonus returns the crit chance bonus for an entity from guild bonuses.
// Used by CombatSystem to apply guild-based critical chance bonus.
func (s *GuildCombatBonusSystem) GetCritBonus(entity *Entity) float64 {
	if entity == nil {
		return 0.0
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return 0.0
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.GetTotalCritBonus()
}

// GetHealingBonus returns the passive healing rate for an entity from guild bonuses.
// Used by regeneration systems to apply guild-based healing.
func (s *GuildCombatBonusSystem) GetHealingBonus(entity *Entity) float64 {
	if entity == nil {
		return 0.0
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return 0.0
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.HealingBonus
}

// GetNearbyGuildMemberCount returns how many guild members are near an entity.
// Useful for UI display showing current guild synergy level.
func (s *GuildCombatBonusSystem) GetNearbyGuildMemberCount(entity *Entity) int {
	if entity == nil {
		return 0
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return 0
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.NearbyGuildMemberCount
}

// HasGuildBonus returns true if entity has active guild combat bonuses.
func (s *GuildCombatBonusSystem) HasGuildBonus(entity *Entity) bool {
	if entity == nil {
		return false
	}

	compRaw, ok := entity.GetComponent("guildcombatbonus")
	if !ok {
		return false
	}

	comp := compRaw.(*GuildCombatBonusComponent)
	return comp.HasSignificantBonus()
}

// GetGenreMultiplier returns the genre-specific guild bonus multiplier.
// Used for UI display and testing.
func (s *GuildCombatBonusSystem) GetGenreMultiplier() float64 {
	return s.getGenreMultiplier()
}
