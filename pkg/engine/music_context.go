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

	if playerEntity.HasComponent("dead") {
		return MusicContextDeath
	}

	playerPos, ok := d.validatePlayerState(playerEntity)
	if !ok {
		return MusicContextExploration
	}

	hasBoss, hasCombat, allEnemiesDead := d.scanEnemyThreats(entities, playerEntity, playerPos)
	return d.resolveContext(playerEntity, hasBoss, hasCombat, allEnemiesDead)
}

// validatePlayerState checks if player exists and has position.
// Returns player position and true if valid, nil and false otherwise.
func (d *MusicContextDetector) validatePlayerState(playerEntity *Entity) (*PositionComponent, bool) {
	if playerEntity == nil {
		return nil, false
	}

	playerPosComp, hasPos := playerEntity.GetComponent("position")
	if !hasPos {
		return nil, false
	}

	playerPos, ok := playerPosComp.(*PositionComponent)
	if !ok {
		return nil, false
	}

	return playerPos, true
}

// scanEnemyThreats scans entities for nearby enemies and boss threats.
// Returns hasBoss, hasCombat, and allEnemiesDead flags.
func (d *MusicContextDetector) scanEnemyThreats(entities []*Entity, playerEntity *Entity, playerPos *PositionComponent) (bool, bool, bool) {
	hasBoss := false
	hasCombat := false
	allEnemiesDead := true

	playerTeam := d.getPlayerTeam(playerEntity)

	for _, entity := range entities {
		if !d.isValidEnemy(entity, playerTeam) {
			continue
		}

		allEnemiesDead = false

		if d.isEnemyNearby(entity, playerPos) {
			hasCombat = true
			if d.isBoss(entity) {
				hasBoss = true
			}
		}
	}

	return hasBoss, hasCombat, allEnemiesDead
}

// getPlayerTeam returns the player's team ID (default 1).
func (d *MusicContextDetector) getPlayerTeam(playerEntity *Entity) int {
	playerTeam := 1
	if teamComp, hasTeam := playerEntity.GetComponent("team"); hasTeam {
		if team, ok := teamComp.(*TeamComponent); ok {
			playerTeam = team.TeamID
		}
	}
	return playerTeam
}

// isValidEnemy checks if an entity is a valid living enemy.
func (d *MusicContextDetector) isValidEnemy(entity *Entity, playerTeam int) bool {
	if entity.HasComponent("input") || entity.HasComponent("dead") || !entity.HasComponent("health") {
		return false
	}

	entityTeam := 0
	if teamComp, hasTeam := entity.GetComponent("team"); hasTeam {
		if team, ok := teamComp.(*TeamComponent); ok {
			entityTeam = team.TeamID
		}
	}

	return entityTeam != playerTeam
}

// isEnemyNearby checks if an enemy is within combat radius.
func (d *MusicContextDetector) isEnemyNearby(entity *Entity, playerPos *PositionComponent) bool {
	posComp, hasPos := entity.GetComponent("position")
	if !hasPos {
		return false
	}

	entityPos, ok := posComp.(*PositionComponent)
	if !ok {
		return false
	}

	distance := d.calculateDistance(playerPos.X, playerPos.Y, entityPos.X, entityPos.Y)
	return distance <= d.CombatRadius
}

// isBoss checks if an entity is a boss based on attack threshold.
func (d *MusicContextDetector) isBoss(entity *Entity) bool {
	statsComp, hasStats := entity.GetComponent("stats")
	if !hasStats {
		return false
	}

	stats, ok := statsComp.(*StatsComponent)
	if !ok {
		return false
	}

	return stats.Attack >= d.BossAttackThreshold
}

// resolveContext determines the final music context based on threats and player state.
func (d *MusicContextDetector) resolveContext(playerEntity *Entity, hasBoss, hasCombat, allEnemiesDead bool) MusicContext {
	if hasBoss {
		return MusicContextBoss
	}

	if hasCombat {
		return MusicContextCombat
	}

	if allEnemiesDead && playerEntity.HasComponent("victory") {
		return MusicContextVictory
	}

	if d.isPlayerInDanger(playerEntity) {
		return MusicContextDanger
	}

	return MusicContextExploration
}

// isPlayerInDanger checks if player health is below danger threshold.
func (d *MusicContextDetector) isPlayerInDanger(playerEntity *Entity) bool {
	healthComp, hasHealth := playerEntity.GetComponent("health")
	if !hasHealth {
		return false
	}

	health, ok := healthComp.(*HealthComponent)
	if !ok {
		return false
	}

	healthPercent := float64(health.Current) / float64(health.Max)
	return healthPercent <= d.DangerHealthPercent && healthPercent > 0
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
