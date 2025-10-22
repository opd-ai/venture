package engine

import "github.com/opd-ai/venture/pkg/combat"

// HealthComponent tracks an entity's health and maximum health.
type HealthComponent struct {
	Current float64
	Max     float64
}

// Type returns the component type identifier.
func (h *HealthComponent) Type() string {
	return "health"
}

// IsAlive returns true if the entity has health remaining.
func (h *HealthComponent) IsAlive() bool {
	return h.Current > 0
}

// IsDead returns true if the entity has no health remaining.
func (h *HealthComponent) IsDead() bool {
	return h.Current <= 0
}

// Heal increases health by the given amount, capped at max health.
func (h *HealthComponent) Heal(amount float64) {
	h.Current += amount
	if h.Current > h.Max {
		h.Current = h.Max
	}
}

// TakeDamage reduces health by the given amount, minimum 0.
func (h *HealthComponent) TakeDamage(amount float64) {
	h.Current -= amount
	if h.Current < 0 {
		h.Current = 0
	}
}

// StatsComponent contains combat statistics for an entity.
type StatsComponent struct {
	// Base stats
	Attack       float64
	Defense      float64
	MagicPower   float64
	MagicDefense float64

	// Critical stats
	CritChance float64 // 0.0 to 1.0 (e.g., 0.15 = 15% chance)
	CritDamage float64 // Multiplier (e.g., 2.0 = 200% damage)

	// Evasion chance
	Evasion float64 // 0.0 to 1.0

	// Resistances per damage type
	Resistances map[combat.DamageType]float64
}

// Type returns the component type identifier.
func (s *StatsComponent) Type() string {
	return "stats"
}

// NewStatsComponent creates a new StatsComponent with default values.
func NewStatsComponent() *StatsComponent {
	return &StatsComponent{
		Attack:       10.0,
		Defense:      5.0,
		MagicPower:   10.0,
		MagicDefense: 5.0,
		CritChance:   0.05,
		CritDamage:   2.0,
		Evasion:      0.0,
		Resistances:  make(map[combat.DamageType]float64),
	}
}

// GetResistance returns the resistance value for a damage type.
// Returns 0.0 if no resistance is configured.
func (s *StatsComponent) GetResistance(damageType combat.DamageType) float64 {
	if resistance, ok := s.Resistances[damageType]; ok {
		return resistance
	}
	return 0.0
}

// AttackComponent marks an entity as being able to attack.
type AttackComponent struct {
	// Damage amount
	Damage float64

	// Damage type
	DamageType combat.DamageType

	// Attack range (for melee/ranged)
	Range float64

	// Attack cooldown in seconds
	Cooldown float64

	// Time until next attack is ready
	CooldownTimer float64
}

// Type returns the component type identifier.
func (a *AttackComponent) Type() string {
	return "attack"
}

// CanAttack returns true if the attack is ready (cooldown expired).
func (a *AttackComponent) CanAttack() bool {
	return a.CooldownTimer <= 0
}

// ResetCooldown resets the cooldown timer.
func (a *AttackComponent) ResetCooldown() {
	a.CooldownTimer = a.Cooldown
}

// UpdateCooldown updates the cooldown timer by the given delta time.
func (a *AttackComponent) UpdateCooldown(deltaTime float64) {
	if a.CooldownTimer > 0 {
		a.CooldownTimer -= deltaTime
		if a.CooldownTimer < 0 {
			a.CooldownTimer = 0
		}
	}
}

// StatusEffectComponent represents a temporary buff or debuff.
type StatusEffectComponent struct {
	// Effect type (e.g., "poison", "stun", "speed_boost")
	EffectType string

	// Duration remaining in seconds
	Duration float64

	// Effect magnitude (meaning depends on effect type)
	Magnitude float64

	// Tick interval for periodic effects (0 = one-time)
	TickInterval float64

	// Time until next tick
	NextTick float64
}

// Type returns the component type identifier.
func (s *StatusEffectComponent) Type() string {
	return "status_effect"
}

// IsExpired returns true if the effect duration has expired.
func (s *StatusEffectComponent) IsExpired() bool {
	return s.Duration <= 0
}

// Update updates the effect duration and tick timer.
func (s *StatusEffectComponent) Update(deltaTime float64) bool {
	s.Duration -= deltaTime

	if s.TickInterval > 0 {
		s.NextTick -= deltaTime
		if s.NextTick <= 0 {
			s.NextTick = s.TickInterval
			return true // Tick occurred
		}
	}

	return false // No tick
}

// TeamComponent identifies which team an entity belongs to.
type TeamComponent struct {
	// Team ID (e.g., 0 = neutral, 1 = player, 2 = enemy)
	TeamID int
}

// Type returns the component type identifier.
func (t *TeamComponent) Type() string {
	return "team"
}

// IsAlly returns true if the other team is an ally.
func (t *TeamComponent) IsAlly(otherTeam int) bool {
	return t.TeamID == otherTeam
}

// IsEnemy returns true if the other team is an enemy.
func (t *TeamComponent) IsEnemy(otherTeam int) bool {
	return t.TeamID != otherTeam && t.TeamID != 0 && otherTeam != 0
}
