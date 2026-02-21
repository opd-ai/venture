// Package engine provides the GuildCombatBonusComponent for tracking
// guild proximity-based combat bonuses.
package engine

// GuildCombatBonusComponent stores active guild combat bonuses for an entity.
// These bonuses are calculated based on nearby guild members in combat.
type GuildCombatBonusComponent struct {
	// NearbyGuildMemberCount tracks how many guild members are within bonus range
	NearbyGuildMemberCount int

	// AttackBonus is the current attack multiplier bonus (0.0 to 0.5)
	AttackBonus float64

	// DefenseBonus is the current defense multiplier bonus (0.0 to 0.3)
	DefenseBonus float64

	// CritBonus is the current critical chance bonus (0.0 to 0.15)
	CritBonus float64

	// HealingBonus provides health regen per second when near allies
	HealingBonus float64

	// RankBonus is additional bonus from having high-rank guild members nearby
	RankBonus float64

	// LastUpdateTime tracks when bonuses were last recalculated
	LastUpdateTime float64

	// BonusRange is the distance within which guild members provide bonuses
	BonusRange float64

	// IsInCombat tracks if entity is actively in combat (for conditional bonuses)
	IsInCombat bool
}

// Type returns the component type identifier.
func (g *GuildCombatBonusComponent) Type() string {
	return "guildcombatbonus"
}

// NewGuildCombatBonusComponent creates a new component with default values.
func NewGuildCombatBonusComponent() *GuildCombatBonusComponent {
	return &GuildCombatBonusComponent{
		BonusRange: 200.0, // Default range for guild synergy
	}
}

// GetTotalAttackMultiplier returns the total attack bonus multiplier (1.0 + bonuses).
func (g *GuildCombatBonusComponent) GetTotalAttackMultiplier() float64 {
	return 1.0 + g.AttackBonus + g.RankBonus*0.5
}

// GetTotalDefenseMultiplier returns the total defense bonus multiplier (1.0 + bonuses).
func (g *GuildCombatBonusComponent) GetTotalDefenseMultiplier() float64 {
	return 1.0 + g.DefenseBonus + g.RankBonus*0.3
}

// GetTotalCritBonus returns the total critical chance bonus.
func (g *GuildCombatBonusComponent) GetTotalCritBonus() float64 {
	return g.CritBonus + g.RankBonus*0.05
}

// ClearBonuses resets all bonus values to zero.
func (g *GuildCombatBonusComponent) ClearBonuses() {
	g.NearbyGuildMemberCount = 0
	g.AttackBonus = 0.0
	g.DefenseBonus = 0.0
	g.CritBonus = 0.0
	g.HealingBonus = 0.0
	g.RankBonus = 0.0
}

// HasSignificantBonus returns true if any bonus is above threshold.
func (g *GuildCombatBonusComponent) HasSignificantBonus() bool {
	return g.AttackBonus > 0.01 || g.DefenseBonus > 0.01 || g.CritBonus > 0.01
}
