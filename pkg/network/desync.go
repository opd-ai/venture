// Package network provides desync detection and recovery for multiplayer.
// This file implements state checksum validation, periodic synchronization,
// conflict resolution, and audit logging to detect and recover from desyncs.
package network

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DesyncType represents the category of desync detected.
type DesyncType string

const (
	DesyncCombat    DesyncType = "combat"    // Server/client disagree on kill attribution
	DesyncInventory DesyncType = "inventory" // Item counts mismatch
	DesyncPosition  DesyncType = "position"  // Player location differs by >1 tile
	DesyncEntity    DesyncType = "entity"    // Entity exists on client but not server
	DesyncQuest     DesyncType = "quest"     // Quest state diverges
	DesyncGuild     DesyncType = "guild"     // Member list differs across servers
	DesyncHousing   DesyncType = "housing"   // Furniture placement mismatch
	DesyncTrade     DesyncType = "trade"     // Ownership transfer fails
	DesyncVehicle   DesyncType = "vehicle"   // Mount/dismount state differs
	DesyncCompanion DesyncType = "companion" // Loyalty/XP values drift
	DesyncChunk     DesyncType = "chunk"     // Terrain modifications lost
	DesyncSkill     DesyncType = "skill"     // Skill points/unlocks mismatch
)

// StateChecksum represents a hash of entity state for comparison.
type StateChecksum struct {
	EntityID  uint64
	Timestamp uint64
	Hash      [32]byte // SHA-256 hash
}

// DesyncEvent represents a detected desync incident.
type DesyncEvent struct {
	Type         DesyncType
	EntityID     uint64
	DetectedAt   time.Time
	ServerHash   [32]byte
	ClientHash   [32]byte
	RecoveryTime time.Duration
	Recovered    bool
	Details      string
}

// DesyncDetector monitors state and detects desyncs between client/server.
type DesyncDetector struct {
	mu                sync.RWMutex
	events            []DesyncEvent
	lastFullSync      time.Time
	fullSyncInterval  time.Duration // Default: 5 minutes
	checksumCache     map[uint64]StateChecksum
	detectionDeadline time.Duration // Default: 30 seconds
	recoveryDeadline  time.Duration // Default: 10 seconds
	hashPool          sync.Pool
}

// NewDesyncDetector creates a new desync detector with default settings.
func NewDesyncDetector() *DesyncDetector {
	return &DesyncDetector{
		events:            make([]DesyncEvent, 0, 100),
		lastFullSync:      time.Now(),
		fullSyncInterval:  5 * time.Minute,
		checksumCache:     make(map[uint64]StateChecksum),
		detectionDeadline: 30 * time.Second,
		recoveryDeadline:  10 * time.Second,
		hashPool: sync.Pool{
			New: func() interface{} {
				return sha256.New()
			},
		},
	}
}

// SetFullSyncInterval configures how often full state sync occurs.
func (d *DesyncDetector) SetFullSyncInterval(interval time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.fullSyncInterval = interval
}

// SetDetectionDeadline configures max time to identify desync.
func (d *DesyncDetector) SetDetectionDeadline(deadline time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.detectionDeadline = deadline
}

// SetRecoveryDeadline configures max time to restore correct state.
func (d *DesyncDetector) SetRecoveryDeadline(deadline time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recoveryDeadline = deadline
}

// ComputeChecksum calculates SHA-256 hash of component data.
func (d *DesyncDetector) ComputeChecksum(entityID, timestamp uint64, components []ComponentData) StateChecksum {
	h := d.hashPool.Get().(hash.Hash)
	defer func() {
		h.Reset()
		d.hashPool.Put(h)
	}()

	// Write entity ID
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], entityID)
	h.Write(buf[:])

	// Write timestamp
	binary.LittleEndian.PutUint64(buf[:], timestamp)
	h.Write(buf[:])

	// Write component data in deterministic order
	for _, comp := range components {
		h.Write([]byte(comp.Type))
		h.Write(comp.Data)
	}

	var hash [32]byte
	copy(hash[:], h.Sum(nil))

	return StateChecksum{
		EntityID:  entityID,
		Timestamp: timestamp,
		Hash:      hash,
	}
}

// ValidateChecksum compares client and server checksums.
// Returns true if checksums match, false if desync detected.
func (d *DesyncDetector) ValidateChecksum(clientChecksum, serverChecksum StateChecksum) bool {
	return clientChecksum.Hash == serverChecksum.Hash
}

// DetectDesync checks for desync and records event if found.
// Returns true if desync detected, false otherwise.
func (d *DesyncDetector) DetectDesync(desyncType DesyncType, entityID uint64, clientHash, serverHash [32]byte, details string) bool {
	if clientHash == serverHash {
		return false
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	event := DesyncEvent{
		Type:       desyncType,
		EntityID:   entityID,
		DetectedAt: time.Now(),
		ServerHash: serverHash,
		ClientHash: clientHash,
		Recovered:  false,
		Details:    details,
	}

	d.events = append(d.events, event)

	logrus.WithFields(logrus.Fields{
		"system_name":   "desync_detector",
		"desync_type":   string(desyncType),
		"entityID":      entityID,
		"details":       details,
		"server_hash":   fmt.Sprintf("%x", serverHash[:8]),
		"client_hash":   fmt.Sprintf("%x", clientHash[:8]),
		"total_desyncs": len(d.events),
	}).Warn("desync detected")

	return true
}

// RecordRecovery marks a desync event as recovered.
func (d *DesyncDetector) RecordRecovery(desyncType DesyncType, entityID uint64, recoveryTime time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Find most recent unrecovered event of this type
	for i := len(d.events) - 1; i >= 0; i-- {
		if d.events[i].Type == desyncType &&
			d.events[i].EntityID == entityID &&
			!d.events[i].Recovered {
			d.events[i].Recovered = true
			d.events[i].RecoveryTime = recoveryTime

			logrus.WithFields(logrus.Fields{
				"system_name":   "desync_detector",
				"desync_type":   string(desyncType),
				"entityID":      entityID,
				"recovery_time": recoveryTime.String(),
			}).Info("desync recovered")
			break
		}
	}
}

// NeedsFullSync returns true if full sync interval has elapsed.
func (d *DesyncDetector) NeedsFullSync() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return time.Since(d.lastFullSync) >= d.fullSyncInterval
}

// MarkFullSync updates the last full sync timestamp.
func (d *DesyncDetector) MarkFullSync() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.lastFullSync = time.Now()
}

// GetEvents returns all recorded desync events.
func (d *DesyncDetector) GetEvents() []DesyncEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()
	events := make([]DesyncEvent, len(d.events))
	copy(events, d.events)
	return events
}

// GetEventsByType returns desync events filtered by type.
func (d *DesyncDetector) GetEventsByType(desyncType DesyncType) []DesyncEvent {
	d.mu.RLock()
	defer d.mu.RUnlock()

	filtered := make([]DesyncEvent, 0)
	for _, event := range d.events {
		if event.Type == desyncType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// GetStats returns statistics about detected desyncs.
func (d *DesyncDetector) GetStats() DesyncStats {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := DesyncStats{
		TotalEvents:      len(d.events),
		RecoveredEvents:  0,
		AverageRecovery:  0,
		FalsePositiveEst: 0,
	}

	var totalRecoveryTime time.Duration
	for _, event := range d.events {
		if event.Recovered {
			stats.RecoveredEvents++
			totalRecoveryTime += event.RecoveryTime
		}
	}

	if stats.RecoveredEvents > 0 {
		stats.AverageRecovery = totalRecoveryTime / time.Duration(stats.RecoveredEvents)
	}

	return stats
}

// DesyncStats provides statistics about desync detection.
type DesyncStats struct {
	TotalEvents      int
	RecoveredEvents  int
	AverageRecovery  time.Duration
	FalsePositiveEst float64 // Estimated false positive rate (0.0-1.0)
}

// ClearEvents removes all recorded events (for testing).
func (d *DesyncDetector) ClearEvents() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.events = make([]DesyncEvent, 0, 100)
}

// RollbackRecovery implements client rollback to server state.
type RollbackRecovery struct {
	serverState func(entityID uint64) ([]ComponentData, error)
	applyState  func(entityID uint64, components []ComponentData) error
}

// NewRollbackRecovery creates a rollback recovery strategy.
func NewRollbackRecovery(
	serverState func(uint64) ([]ComponentData, error),
	applyState func(uint64, []ComponentData) error,
) *RollbackRecovery {
	return &RollbackRecovery{
		serverState: serverState,
		applyState:  applyState,
	}
}

// Recover reverts client to server state.
func (r *RollbackRecovery) Recover(event DesyncEvent) error {
	if r.serverState == nil || r.applyState == nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "rollback_recovery",
			"entityID":    event.EntityID,
			"desync_type": string(event.Type),
		}).Error("recovery functions not configured")
		return fmt.Errorf("recovery functions not configured")
	}

	components, err := r.serverState(event.EntityID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "rollback_recovery",
			"entityID":    event.EntityID,
			"desync_type": string(event.Type),
			"error":       err.Error(),
		}).Error("failed to get server state")
		return fmt.Errorf("failed to get server state: %w", err)
	}

	if err := r.applyState(event.EntityID, components); err != nil {
		logrus.WithFields(logrus.Fields{
			"system_name": "rollback_recovery",
			"entityID":    event.EntityID,
			"desync_type": string(event.Type),
			"error":       err.Error(),
		}).Error("failed to apply server state")
		return fmt.Errorf("failed to apply server state: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"system_name":     "rollback_recovery",
		"entityID":        event.EntityID,
		"desync_type":     string(event.Type),
		"component_count": len(components),
	}).Info("rollback recovery completed")

	return nil
}

// PeriodicSyncManager handles full state synchronization.
type PeriodicSyncManager struct {
	detector *DesyncDetector
	syncFunc func() error
	ticker   *time.Ticker
	stopChan chan struct{}
	running  bool
	mu       sync.Mutex
	logger   *logrus.Entry
}

// NewPeriodicSyncManager creates a manager for periodic full sync.
// If logger is nil, a default logger with warn level will be created.
func NewPeriodicSyncManager(detector *DesyncDetector, syncFunc func() error, logger *logrus.Entry) *PeriodicSyncManager {
	if logger == nil {
		log := logrus.New()
		log.SetLevel(logrus.WarnLevel)
		logger = logrus.NewEntry(log)
	}
	return &PeriodicSyncManager{
		detector: detector,
		syncFunc: syncFunc,
		stopChan: make(chan struct{}),
		logger:   logger,
	}
}

// Start begins periodic synchronization.
func (p *PeriodicSyncManager) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return
	}

	p.ticker = time.NewTicker(p.detector.fullSyncInterval)
	p.stopChan = make(chan struct{})
	p.running = true

	go func() {
		for {
			select {
			case <-p.ticker.C:
				if p.syncFunc != nil {
					if err := p.syncFunc(); err != nil {
						p.logger.WithFields(logrus.Fields{
							"system_name": "periodic_sync",
							"error":       err.Error(),
						}).Warn("periodic sync failed")
					}
					p.detector.MarkFullSync()
				}
			case <-p.stopChan:
				return
			}
		}
	}()
}

// Stop halts periodic synchronization.
func (p *PeriodicSyncManager) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return
	}

	if p.ticker != nil {
		p.ticker.Stop()
	}
	close(p.stopChan)
	p.running = false
}
