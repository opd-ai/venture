package learning

// Constants for companion learning system configuration.
// These values define the bounds and defaults for skill progression,
// personality evolution, and event memory systems.

const (
	// DefaultLearningRate is the default multiplier for XP gain.
	// Range: 0.5 (slow learner) to 2.0 (fast learner)
	DefaultLearningRate = 1.0

	// DefaultMaxEvents is the maximum number of memorable events
	// stored per companion in LRU memory.
	DefaultMaxEvents = 1000

	// DefaultMaxPersonalityChanges is the maximum number of personality
	// changes tracked in history (LRU eviction).
	DefaultMaxPersonalityChanges = 1000

	// SkillDecayRate is the rate at which unused skills decay per second
	// of delta time. Skills not used within 24 hours begin slow decay.
	SkillDecayRate = 0.1

	// SkillXPPerLevel is the base XP required per level.
	// Formula: (Level + 1) * SkillXPPerLevel
	SkillXPPerLevel = 100.0

	// SkillBonusPerLevel is the bonus multiplier added per skill level.
	// Formula: 1.0 + (Level * SkillBonusPerLevel)
	SkillBonusPerLevel = 0.1

	// TraitMinValue is the minimum value for personality traits.
	TraitMinValue = 0.0

	// TraitMaxValue is the maximum value for personality traits.
	TraitMaxValue = 1.0

	// TraitDefaultValue is the initial neutral value for all traits.
	TraitDefaultValue = 0.5

	// TraitBalanceMinSum is the minimum acceptable sum for opposing traits
	// before normalization is triggered.
	TraitBalanceMinSum = 0.8

	// TraitBalanceMaxSum is the maximum acceptable sum for opposing traits
	// before normalization is triggered.
	TraitBalanceMaxSum = 1.2

	// DefaultSkillCost is the default skill point cost when not specified.
	DefaultSkillCost = 1
)

// Combat action XP rewards
const (
	// CombatBaseXP is the base XP awarded for combat actions.
	CombatBaseXP = 10.0

	// CombatVictoryXP is the XP awarded for defeating enemies.
	CombatVictoryXP = 20.0
)

// Personality adjustment deltas for various actions
const (
	// TraitSmallDelta is a minor personality adjustment.
	TraitSmallDelta = 0.01

	// TraitMediumDelta is a moderate personality adjustment.
	TraitMediumDelta = 0.02

	// TraitLargeDelta is a significant personality adjustment.
	TraitLargeDelta = 0.05

	// ExplorationCuriosityDelta is the curiosity increase for exploration.
	ExplorationCuriosityDelta = 0.015

	// ExplorationPracticalDelta is the practical decrease for exploration.
	ExplorationPracticalDelta = -0.005
)
