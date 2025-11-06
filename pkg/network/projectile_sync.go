package network

import (
	"sync"
	"time"
)

// ProjectileNetworkSync handles projectile synchronization between client and server.
// Phase 10.2 Week 4: Projectile Multiplayer Synchronization
//
// Server-side: Broadcasts ProjectileSpawnMessage when projectiles are created,
// ProjectileHitMessage when collisions occur, and ProjectileDespawnMessage when
// projectiles are removed.
//
// Client-side: Uses client-side prediction for local player projectiles with
// server reconciliation. Remote player projectiles are rendered based on
// server-authoritative messages with interpolation.
type ProjectileNetworkSync struct {
	// sequenceNumber tracks message ordering
	sequenceNumber uint32
	mu             sync.Mutex

	// serverTime is the current server time (seconds)
	serverTime float64

	// predictedProjectiles tracks client-predicted projectiles awaiting server confirmation
	// Map: local prediction ID -> ProjectileSpawnMessage
	predictedProjectiles map[uint64]ProjectileSpawnMessage

	// confirmedProjectiles tracks server-confirmed projectiles
	// Map: server projectile ID -> ProjectileSpawnMessage
	confirmedProjectiles map[uint64]ProjectileSpawnMessage

	// projectileHistory stores recent projectile states for lag compensation
	// Map: projectile ID -> []timestamped state snapshots
	projectileHistory map[uint64][]ProjectileSnapshot
}

// ProjectileSnapshot represents a projectile's state at a specific time.
// Used for lag compensation and interpolation.
type ProjectileSnapshot struct {
	Timestamp float64
	X         float64
	Y         float64
	VX        float64
	VY        float64
	Age       float64
}

// NewProjectileNetworkSync creates a new projectile network synchronization handler.
func NewProjectileNetworkSync() *ProjectileNetworkSync {
	return &ProjectileNetworkSync{
		sequenceNumber:       0,
		serverTime:           0.0,
		predictedProjectiles: make(map[uint64]ProjectileSpawnMessage),
		confirmedProjectiles: make(map[uint64]ProjectileSpawnMessage),
		projectileHistory:    make(map[uint64][]ProjectileSnapshot),
	}
}

// UpdateServerTime updates the current server time.
// Called by the server's game loop or network clock sync.
func (s *ProjectileNetworkSync) UpdateServerTime(deltaTime float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.serverTime += deltaTime
}

// GetServerTime returns the current server time.
func (s *ProjectileNetworkSync) GetServerTime() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.serverTime
}

// CreateSpawnMessage creates a ProjectileSpawnMessage for server broadcast.
// Called when a projectile entity is spawned on the server.
//
// Parameters match ProjectileComponent fields:
// - projectileID: Unique entity ID for the projectile
// - ownerID: Entity ID of the player/entity that fired the projectile
// - x, y: Spawn position coordinates
// - vx, vy: Velocity components (pixels per second)
// - damage: Base damage on hit
// - speed: Movement speed (magnitude of velocity, redundant but convenient)
// - lifetime: Maximum duration before despawn (seconds)
// - pierce: Number of entities projectile can pass through (0=normal, -1=infinite)
// - bounce: Number of wall bounces remaining
// - explosive: Whether projectile explodes on impact
// - explosionRadius: Area damage radius (pixels) if explosive
// - projectileType: Visual/logical type ("arrow", "bullet", "fireball", etc.)
func (s *ProjectileNetworkSync) CreateSpawnMessage(
	projectileID, ownerID uint64,
	x, y, vx, vy, damage, speed, lifetime float64,
	pierce, bounce int,
	explosive bool,
	explosionRadius float64,
	projectileType string,
) ProjectileSpawnMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequenceNumber++
	msg := ProjectileSpawnMessage{
		ProjectileID:    projectileID,
		OwnerID:         ownerID,
		PositionX:       x,
		PositionY:       y,
		VelocityX:       vx,
		VelocityY:       vy,
		Damage:          damage,
		Speed:           speed,
		Lifetime:        lifetime,
		Pierce:          pierce,
		Bounce:          bounce,
		Explosive:       explosive,
		ExplosionRadius: explosionRadius,
		ProjectileType:  projectileType,
		SpawnTime:       s.serverTime,
		SequenceNumber:  s.sequenceNumber,
	}

	// Store in confirmed projectiles (server-authoritative)
	s.confirmedProjectiles[projectileID] = msg

	// Initialize history for this projectile
	s.projectileHistory[projectileID] = []ProjectileSnapshot{
		{
			Timestamp: s.serverTime,
			X:         x,
			Y:         y,
			VX:        vx,
			VY:        vy,
			Age:       0.0,
		},
	}

	return msg
}

// CreateHitMessage creates a ProjectileHitMessage for server broadcast.
// Called when a projectile collides with an entity or wall.
//
// Parameters:
// - projectileID: ID of the projectile that hit
// - hitEntityID: ID of entity that was hit (0 for walls)
// - hitType: "entity", "wall", or "expire"
// - damageDealt: Actual damage applied (may differ from base damage due to armor, etc.)
// - x, y: Collision position
// - projectileDestroyed: Whether projectile should be removed
// - explosionTriggered: Whether explosion occurred (for explosive projectiles)
// - explosionEntities: IDs of entities damaged by explosion
// - explosionDamages: Damage dealt to each explosion entity (parallel array)
func (s *ProjectileNetworkSync) CreateHitMessage(
	projectileID, hitEntityID uint64,
	hitType string,
	damageDealt, x, y float64,
	projectileDestroyed, explosionTriggered bool,
	explosionEntities []uint64,
	explosionDamages []float64,
) ProjectileHitMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequenceNumber++
	msg := ProjectileHitMessage{
		ProjectileID:        projectileID,
		HitEntityID:         hitEntityID,
		HitType:             hitType,
		DamageDealt:         damageDealt,
		PositionX:           x,
		PositionY:           y,
		ProjectileDestroyed: projectileDestroyed,
		ExplosionTriggered:  explosionTriggered,
		ExplosionEntities:   explosionEntities,
		ExplosionDamages:    explosionDamages,
		HitTime:             s.serverTime,
		SequenceNumber:      s.sequenceNumber,
	}

	// If projectile destroyed, remove from confirmed projectiles
	if projectileDestroyed {
		delete(s.confirmedProjectiles, projectileID)
		// Keep history for lag compensation queries for a short time
		// (will be cleaned up by CleanupOldHistory)
	}

	return msg
}

// CreateDespawnMessage creates a ProjectileDespawnMessage for server broadcast.
// Called when a projectile is removed (expired, hit, out of bounds).
//
// Parameters:
// - projectileID: ID of projectile being despawned
// - reason: "expired", "hit", or "out_of_bounds"
func (s *ProjectileNetworkSync) CreateDespawnMessage(
	projectileID uint64,
	reason string,
) ProjectileDespawnMessage {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sequenceNumber++
	msg := ProjectileDespawnMessage{
		ProjectileID:   projectileID,
		Reason:         reason,
		DespawnTime:    s.serverTime,
		SequenceNumber: s.sequenceNumber,
	}

	// Remove from confirmed projectiles
	delete(s.confirmedProjectiles, projectileID)

	return msg
}

// RecordSnapshot records a projectile's state for lag compensation.
// Called by server's projectile update loop.
//
// Parameters:
// - projectileID: ID of projectile
// - x, y: Current position
// - vx, vy: Current velocity
// - age: Time since spawn (seconds)
func (s *ProjectileNetworkSync) RecordSnapshot(
	projectileID uint64,
	x, y, vx, vy, age float64,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	snapshot := ProjectileSnapshot{
		Timestamp: s.serverTime,
		X:         x,
		Y:         y,
		VX:        vx,
		VY:        vy,
		Age:       age,
	}

	history := s.projectileHistory[projectileID]
	history = append(history, snapshot)

	// Keep only recent history (last 1 second)
	// Supports lag compensation up to 1000ms
	// This value balances memory usage (~40 snapshots/projectile) with lag compensation range
	const maxHistoryDuration = 1.0
	cutoffTime := s.serverTime - maxHistoryDuration
	for len(history) > 0 && history[0].Timestamp < cutoffTime {
		history = history[1:]
	}

	s.projectileHistory[projectileID] = history
}

// GetHistoricalState retrieves a projectile's state at a specific time.
// Used for lag compensation when validating client actions.
//
// Returns the interpolated snapshot at the given timestamp, or nil if not available.
func (s *ProjectileNetworkSync) GetHistoricalState(
	projectileID uint64,
	timestamp float64,
) *ProjectileSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	history := s.projectileHistory[projectileID]
	if len(history) == 0 {
		return nil
	}

	// Find bracketing snapshots
	var before, after *ProjectileSnapshot
	for i := range history {
		if history[i].Timestamp <= timestamp {
			before = &history[i]
		}
		if history[i].Timestamp >= timestamp && after == nil {
			after = &history[i]
		}
	}

	// If exact match or only one boundary, return it
	if before != nil && after != nil && before.Timestamp == after.Timestamp {
		return before
	}
	// If timestamp is outside the available range, return nil
	// Don't extrapolate beyond available data for lag compensation
	if before == nil || after == nil {
		return nil
	}

	// Interpolate between before and after
	t := (timestamp - before.Timestamp) / (after.Timestamp - before.Timestamp)
	return &ProjectileSnapshot{
		Timestamp: timestamp,
		X:         before.X + t*(after.X-before.X),
		Y:         before.Y + t*(after.Y-before.Y),
		VX:        before.VX + t*(after.VX-before.VX),
		VY:        before.VY + t*(after.VY-before.VY),
		Age:       before.Age + t*(after.Age-before.Age),
	}
}

// CleanupOldHistory removes projectile history older than 2 seconds.
// Should be called periodically (e.g., once per second) to prevent memory leaks.
func (s *ProjectileNetworkSync) CleanupOldHistory() {
	s.mu.Lock()
	defer s.mu.Unlock()

	const maxHistoryAge = 2.0 // Keep 2 seconds of history for safety (2x maxHistoryDuration for grace period)
	cutoffTime := s.serverTime - maxHistoryAge

	for projectileID, history := range s.projectileHistory {
		// Remove old snapshots
		for len(history) > 0 && history[0].Timestamp < cutoffTime {
			history = history[1:]
		}

		// If no snapshots remain and projectile not confirmed, remove entirely
		if len(history) == 0 {
			if _, exists := s.confirmedProjectiles[projectileID]; !exists {
				delete(s.projectileHistory, projectileID)
			}
		} else {
			s.projectileHistory[projectileID] = history
		}
	}
}

// PredictProjectile records a client-predicted projectile awaiting server confirmation.
// Client-side only. Called when local player fires a projectile.
//
// The prediction ID should be a locally-generated unique ID (e.g., client timestamp).
// When server confirms, client reconciles the prediction with server-authoritative ID.
func (s *ProjectileNetworkSync) PredictProjectile(predictionID uint64, msg ProjectileSpawnMessage) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.predictedProjectiles[predictionID] = msg
}

// ConfirmPrediction confirms a client-predicted projectile with server ID.
// Client-side only. Called when server's ProjectileSpawnMessage arrives.
//
// Returns true if a matching prediction was found and removed.
func (s *ProjectileNetworkSync) ConfirmPrediction(predictionID, serverProjectileID uint64, serverMsg ProjectileSpawnMessage) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if predictedMsg, exists := s.predictedProjectiles[predictionID]; exists {
		delete(s.predictedProjectiles, predictionID)
		s.confirmedProjectiles[serverProjectileID] = serverMsg

		// Check for misprediction: if spawn position/velocity differs significantly, correction needed
		// Tolerances chosen based on typical network jitter (10-50ms @ 300-600 px/s projectile speeds)
		const positionTolerance = 10.0 // pixels (allows ~15-30ms prediction error)
		const velocityTolerance = 50.0 // pixels per second (allows minor aim corrections)
		dx := serverMsg.PositionX - predictedMsg.PositionX
		dy := serverMsg.PositionY - predictedMsg.PositionY
		dvx := serverMsg.VelocityX - predictedMsg.VelocityX
		dvy := serverMsg.VelocityY - predictedMsg.VelocityY

		mispredicted := (dx*dx+dy*dy > positionTolerance*positionTolerance) ||
			(dvx*dvx+dvy*dvy > velocityTolerance*velocityTolerance)

		// Return true to indicate successful confirmation
		// Caller can check misprediction by comparing messages if needed
		_ = mispredicted // For now, just detect; future: trigger correction

		return true
	}

	return false
}

// GetConfirmedProjectile retrieves a confirmed projectile's spawn message.
// Used by client to render remote projectiles.
func (s *ProjectileNetworkSync) GetConfirmedProjectile(projectileID uint64) (ProjectileSpawnMessage, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	msg, exists := s.confirmedProjectiles[projectileID]
	return msg, exists
}

// RemoveProjectile removes a projectile from tracking (confirmed or predicted).
// Called when ProjectileDespawnMessage or ProjectileHitMessage (destroyed) is received.
func (s *ProjectileNetworkSync) RemoveProjectile(projectileID uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.confirmedProjectiles, projectileID)
	// Don't immediately delete history; CleanupOldHistory will handle it
}

// GetStats returns synchronization statistics for debugging/monitoring.
func (s *ProjectileNetworkSync) GetStats() ProjectileSyncStats {
	s.mu.Lock()
	defer s.mu.Unlock()

	var totalSnapshots int
	for _, history := range s.projectileHistory {
		totalSnapshots += len(history)
	}

	return ProjectileSyncStats{
		ConfirmedProjectiles: len(s.confirmedProjectiles),
		PredictedProjectiles: len(s.predictedProjectiles),
		HistoryEntries:       len(s.projectileHistory),
		TotalSnapshots:       totalSnapshots,
		ServerTime:           s.serverTime,
	}
}

// ProjectileSyncStats contains synchronization statistics.
type ProjectileSyncStats struct {
	ConfirmedProjectiles int     // Number of server-confirmed projectiles
	PredictedProjectiles int     // Number of client-predicted projectiles awaiting confirmation
	HistoryEntries       int     // Number of projectiles with history
	TotalSnapshots       int     // Total history snapshots across all projectiles
	ServerTime           float64 // Current server time
}

// CleanupTask starts a background goroutine that periodically cleans up old history.
// Server-side only. Should be started once during server initialization.
// Returns a stop channel - send a value to stop the cleanup task.
func (s *ProjectileNetworkSync) CleanupTask() chan<- struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.CleanupOldHistory()
			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}
