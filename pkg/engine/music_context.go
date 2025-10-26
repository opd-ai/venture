package engine

import (
	"math"
	"time"
)

// MusicContext represents the current game situation for music selection
// Design: Simple enum pattern for clear, type-safe context management
// Why: Avoids string comparisons and provides compile-time safety
type MusicContext int

const (
	// MusicContextExploration is the default peaceful exploration music
	MusicContextExploration MusicContext = iota
	// MusicContextCombat plays when enemies are nearby
	MusicContextCombat
	// MusicContextBoss plays during boss encounters
	MusicContextBoss
	// MusicContextDanger plays when player health is critical
	MusicContextDanger
	// MusicContextVictory plays after defeating all enemies
	MusicContextVictory
	// MusicContextDeath plays when player dies
	MusicContextDeath
)

// String returns the string representation of the music context
func (mc MusicContext) String() string {
	switch mc {
	case MusicContextExploration:
		return "exploration"
	case MusicContextCombat:
		return "combat"
	case MusicContextBoss:
		return "boss"
	case MusicContextDanger:
		return "danger"
	case MusicContextVictory:
		return "victory"
	case MusicContextDeath:
		return "death"
	default:
		return "unknown"
	}
}

// Priority returns the priority level (higher = more important)
// Design: Boss > Combat > Danger > Exploration > Victory > Death
// Why: Ensures dramatic moments take precedence over ambient states
func (mc MusicContext) Priority() int {
	switch mc {
	case MusicContextBoss:
		return 100
	case MusicContextCombat:
		return 80
	case MusicContextDanger:
		return 60
	case MusicContextExploration:
		return 40
	case MusicContextVictory:
		return 20
	case MusicContextDeath:
		return 10
	default:
		return 0
	}
}

// MusicContextDetector analyzes game state to determine appropriate music context
// Design: Stateless detector with configurable thresholds
// Why: Allows easy testing and tuning without side effects
type MusicContextDetector struct {
	// CombatRadius is the distance to detect enemies (in pixels)
	CombatRadius float64
	// BossAttackThreshold is the minimum attack stat to consider an enemy a boss
	BossAttackThreshold float64
	// DangerHealthPercent is the health percentage threshold for danger music (0.0-1.0)
	DangerHealthPercent float64
}

// NewMusicContextDetector creates a detector with default settings
func NewMusicContextDetector() *MusicContextDetector {
	return &MusicContextDetector{
		CombatRadius:        300.0, // 300 pixel radius for enemy detection
		BossAttackThreshold: 20.0,  // Attack > 20 = boss
		DangerHealthPercent: 0.2,   // <20% health = danger
	}
}

// DetectContext analyzes entities to determine the appropriate music context
// Design: Single-pass entity scan with position-based proximity checks
// Why: O(n) complexity, minimal allocations, clear priority logic
func (d *MusicContextDetector) DetectContext(entities []*Entity, playerEntity *Entity) MusicContext {
	if playerEntity == nil {
		return MusicContextExploration
	}

	// Check if player is dead
	if playerEntity.HasComponent("dead") {
		return MusicContextDeath
	}

	// Get player position
	playerPosComp, hasPos := playerEntity.GetComponent("position")
	if !hasPos {
		return MusicContextExploration
	}
	playerPos := playerPosComp.(*PositionComponent)

	// Check player health for danger state
	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
		health := healthComp.(*HealthComponent)
		healthPercent := float64(health.Current) / float64(health.Max)
		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
			// Danger state active, but continue checking for combat/boss
			// (boss/combat takes priority over danger)
		}
	}

	// Scan for nearby enemies
	hasBoss := false
	hasCombat := false
	allEnemiesDead := true

	for _, entity := range entities {
		// Skip player entities (have input component)
		if entity.HasComponent("input") {
			continue
		}

		// Skip dead entities
		if entity.HasComponent("dead") {
			continue
		}

		// Check if entity is an enemy (has health + team different from player)
		if !entity.HasComponent("health") {
			continue
		}

		// Check team (enemies are on different team than player)
		playerTeam := 1 // Default player team
		if teamComp, hasTeam := playerEntity.GetComponent("team"); hasTeam {
			playerTeam = teamComp.(*TeamComponent).TeamID
		}

		entityTeam := 0 // Default enemy team
		if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
			entityTeam = teamComp.(*TeamComponent).TeamID
		}

		if entityTeam == playerTeam {
			continue // Same team, not an enemy
		}

		// Found a living enemy
		allEnemiesDead = false

		// Check proximity to player
		if posComp, hasPos := entity.GetComponent("position"); hasPos {
			entityPos := posComp.(*PositionComponent)
			distance := d.calculateDistance(playerPos.X, playerPos.Y, entityPos.X, entityPos.Y)

			if distance <= d.CombatRadius {
				hasCombat = true

				// Check if it's a boss (high attack stat)
				if statsComp, hasStats := entity.GetComponent("stats"); hasStats {
					stats := statsComp.(*StatsComponent)
					if stats.Attack >= d.BossAttackThreshold {
						hasBoss = true
					}
				}
			}
		}
	}

	// Return highest priority context
	if hasBoss {
		return MusicContextBoss
	}
	if hasCombat {
		return MusicContextCombat
	}
	if allEnemiesDead && playerEntity.HasComponent("victory") {
		return MusicContextVictory
	}

	// Check danger state last (lowest priority of active contexts)
	if healthComp, hasHealth := playerEntity.GetComponent("health"); hasHealth {
		health := healthComp.(*HealthComponent)
		healthPercent := float64(health.Current) / float64(health.Max)
		if healthPercent <= d.DangerHealthPercent && healthPercent > 0 {
			return MusicContextDanger
		}
	}

	return MusicContextExploration
}

// calculateDistance computes Euclidean distance between two points
func (d *MusicContextDetector) calculateDistance(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return math.Sqrt(dx*dx + dy*dy)
}

// MusicTransitionManager handles smooth transitions between music contexts
// Design: Time-based cooldown with configurable transition duration
// Why: Prevents jarring rapid context switches while maintaining responsiveness
type MusicTransitionManager struct {
	currentContext       MusicContext
	lastTransitionTime   time.Time
	transitionCooldown   time.Duration
	transitionInProgress bool
}

// NewMusicTransitionManager creates a new transition manager
func NewMusicTransitionManager() *MusicTransitionManager {
	return &MusicTransitionManager{
		currentContext:       MusicContextExploration,
		lastTransitionTime:   time.Now(),
		transitionCooldown:   10 * time.Second, // Minimum 10 seconds between transitions
		transitionInProgress: false,
	}
}

// ShouldTransition determines if a context change should occur
// Design: Priority-based with cooldown override for high-priority contexts
// Why: Boss/death contexts should interrupt immediately, others wait for cooldown
func (m *MusicTransitionManager) ShouldTransition(newContext MusicContext) bool {
	// Same context, no transition needed
	if newContext == m.currentContext {
		return false
	}

	// High-priority contexts (boss, death) can interrupt immediately
	if newContext.Priority() >= MusicContextBoss.Priority() {
		return true
	}

	// Check if cooldown period has elapsed
	timeSinceLastTransition := time.Since(m.lastTransitionTime)
	if timeSinceLastTransition < m.transitionCooldown {
		// Within cooldown, only allow if new context has higher priority
		return newContext.Priority() > m.currentContext.Priority()
	}

	// Cooldown elapsed, allow transition
	return true
}

// BeginTransition marks the start of a transition to a new context
func (m *MusicTransitionManager) BeginTransition(newContext MusicContext) {
	m.currentContext = newContext
	m.lastTransitionTime = time.Now()
	m.transitionInProgress = true
}

// CompleteTransition marks the end of a transition
func (m *MusicTransitionManager) CompleteTransition() {
	m.transitionInProgress = false
}

// GetCurrentContext returns the current music context
func (m *MusicTransitionManager) GetCurrentContext() MusicContext {
	return m.currentContext
}

// IsTransitioning returns whether a transition is in progress
func (m *MusicTransitionManager) IsTransitioning() bool {
	return m.transitionInProgress
}

// SetCooldown updates the minimum transition cooldown duration
func (m *MusicTransitionManager) SetCooldown(cooldown time.Duration) {
	m.transitionCooldown = cooldown
}
