package federation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/engine"
)

// TransferPhase represents the current phase of a player transfer
type TransferPhase int

const (
	// TransferPhasePrepare indicates initial state before transfer begins
	TransferPhasePrepare TransferPhase = iota
	// TransferPhaseTransfer indicates active transfer in progress
	TransferPhaseTransfer
	// TransferPhaseConfirm indicates transfer succeeded and is being confirmed
	TransferPhaseConfirm
	// TransferPhaseRollback indicates transfer failed and is being rolled back
	TransferPhaseRollback
)

// String returns a human-readable name for the transfer phase
func (t TransferPhase) String() string {
	switch t {
	case TransferPhasePrepare:
		return "Prepare"
	case TransferPhaseTransfer:
		return "Transfer"
	case TransferPhaseConfirm:
		return "Confirm"
	case TransferPhaseRollback:
		return "Rollback"
	default:
		return "Unknown"
	}
}

// PlayerState represents serialized player data for transfer
type PlayerState struct {
	PlayerID   uint64                 `json:"player_id"`
	Position   *PositionData          `json:"position"`
	Health     *HealthData            `json:"health"`
	Stats      map[string]interface{} `json:"stats"`
	Inventory  []ItemData             `json:"inventory"`
	Quests     []QuestData            `json:"quests"`
	Reputation map[string]float64     `json:"reputation"`
}

// PositionData serializes position component
type PositionData struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// HealthData serializes health component
type HealthData struct {
	Current float64 `json:"current"`
	Max     float64 `json:"max"`
}

// ItemData serializes inventory item
type ItemData struct {
	ItemID   string `json:"item_id"`
	TypeName string `json:"type_name"`
	Quantity int    `json:"quantity"`
}

// QuestData serializes quest state
type QuestData struct {
	QuestID   string `json:"quest_id"`
	Completed bool   `json:"completed"`
	Progress  int    `json:"progress"`
}

// PlayerTransfer represents a player transfer in progress
type PlayerTransfer struct {
	Phase       TransferPhase `json:"phase"`
	PlayerState *PlayerState  `json:"player_state"`
	StateHash   string        `json:"state_hash"`
	TimeoutAt   int64         `json:"timeout_at"` // Unix timestamp in seconds
	OriginID    string        `json:"origin_id"`
	TargetID    string        `json:"target_id"`
}

// TransferManager handles player transfers between servers
type TransferManager struct {
	mu               sync.RWMutex
	activeTransfers  map[uint64]*PlayerTransfer
	transferTimeout  time.Duration
	stateBackups     map[uint64]*PlayerState
	onTransferStart  func(playerID uint64, targetServer string)
	onTransferCommit func(playerID uint64)
	onTransferFail   func(playerID uint64, reason string)
}

// NewTransferManager creates a new transfer manager
func NewTransferManager() *TransferManager {
	return &TransferManager{
		activeTransfers: make(map[uint64]*PlayerTransfer),
		transferTimeout: 60 * time.Second,
		stateBackups:    make(map[uint64]*PlayerState),
	}
}

// SetTransferTimeout sets the timeout for transfers
func (tm *TransferManager) SetTransferTimeout(timeout time.Duration) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.transferTimeout = timeout
}

// SetCallbacks sets the transfer event callbacks
func (tm *TransferManager) SetCallbacks(
	onStart func(uint64, string),
	onCommit func(uint64),
	onFail func(uint64, string),
) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.onTransferStart = onStart
	tm.onTransferCommit = onCommit
	tm.onTransferFail = onFail
}

// PrepareTransfer initiates a player transfer
func (tm *TransferManager) PrepareTransfer(playerID uint64, world *engine.World, targetServerID string) (*PlayerTransfer, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	// Check if transfer already in progress
	if _, exists := tm.activeTransfers[playerID]; exists {
		return nil, fmt.Errorf("transfer already in progress for player %d", playerID)
	}

	// Serialize player state
	state, err := tm.serializePlayerState(playerID, world)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize player state: %w", err)
	}

	// Create backup
	tm.stateBackups[playerID] = state

	// Compute state hash
	hash, err := tm.computeStateHash(state)
	if err != nil {
		return nil, fmt.Errorf("failed to compute state hash: %w", err)
	}

	// Create transfer
	transfer := &PlayerTransfer{
		Phase:       TransferPhasePrepare,
		PlayerState: state,
		StateHash:   hash,
		TimeoutAt:   time.Now().Add(tm.transferTimeout).Unix(),
		TargetID:    targetServerID,
	}

	tm.activeTransfers[playerID] = transfer

	// Invoke callback
	if tm.onTransferStart != nil {
		tm.onTransferStart(playerID, targetServerID)
	}

	return transfer, nil
}

// BeginTransfer moves transfer to transfer phase
func (tm *TransferManager) BeginTransfer(playerID uint64, originServerID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transfer, exists := tm.activeTransfers[playerID]
	if !exists {
		return fmt.Errorf("no transfer in progress for player %d", playerID)
	}

	if transfer.Phase != TransferPhasePrepare {
		return fmt.Errorf("transfer not in prepare phase")
	}

	transfer.Phase = TransferPhaseTransfer
	transfer.OriginID = originServerID
	return nil
}

// ConfirmTransfer confirms successful transfer
func (tm *TransferManager) ConfirmTransfer(playerID uint64) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transfer, exists := tm.activeTransfers[playerID]
	if !exists {
		return fmt.Errorf("no transfer in progress for player %d", playerID)
	}

	transfer.Phase = TransferPhaseConfirm

	// Clean up
	delete(tm.activeTransfers, playerID)
	delete(tm.stateBackups, playerID)

	// Invoke callback
	if tm.onTransferCommit != nil {
		tm.onTransferCommit(playerID)
	}

	return nil
}

// RollbackTransfer rolls back a failed transfer
func (tm *TransferManager) RollbackTransfer(playerID uint64, reason string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	transfer, exists := tm.activeTransfers[playerID]
	if !exists {
		return fmt.Errorf("no transfer in progress for player %d", playerID)
	}

	transfer.Phase = TransferPhaseRollback

	// Clean up
	delete(tm.activeTransfers, playerID)

	// Backup is kept for restore operation

	// Invoke callback
	if tm.onTransferFail != nil {
		tm.onTransferFail(playerID, reason)
	}

	return nil
}

// RestorePlayerState restores a player from backup
func (tm *TransferManager) RestorePlayerState(playerID uint64, world *engine.World) error {
	tm.mu.RLock()
	backup, exists := tm.stateBackups[playerID]
	tm.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no backup found for player %d", playerID)
	}

	if err := tm.deserializePlayerState(backup, world); err != nil {
		return fmt.Errorf("failed to restore player state: %w", err)
	}

	// Clean up backup after successful restore
	tm.mu.Lock()
	delete(tm.stateBackups, playerID)
	tm.mu.Unlock()

	return nil
}

// GetTransfer retrieves an active transfer
func (tm *TransferManager) GetTransfer(playerID uint64) (*PlayerTransfer, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	transfer, exists := tm.activeTransfers[playerID]
	return transfer, exists
}

// CheckTimeouts checks for expired transfers
func (tm *TransferManager) CheckTimeouts() []uint64 {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now().Unix()
	var expired []uint64

	for playerID, transfer := range tm.activeTransfers {
		if transfer.TimeoutAt < now {
			expired = append(expired, playerID)
		}
	}

	return expired
}

// ValidatePlayerState validates player state for transfer
func (tm *TransferManager) ValidatePlayerState(state *PlayerState) error {
	if state == nil {
		return fmt.Errorf("state is nil")
	}

	if state.PlayerID == 0 {
		return fmt.Errorf("invalid player ID")
	}

	// Validate health bounds
	if state.Health != nil {
		if state.Health.Current < 0 {
			return fmt.Errorf("health current cannot be negative")
		}
		if state.Health.Max <= 0 {
			return fmt.Errorf("health max must be positive")
		}
		// Allow small variance for buffs
		if state.Health.Current > state.Health.Max*1.1 {
			return fmt.Errorf("health current exceeds max by more than 10%%")
		}
	}

	// Validate inventory size
	if len(state.Inventory) > 100 {
		return fmt.Errorf("inventory too large: %d items", len(state.Inventory))
	}

	return nil
}

// serializePlayerState converts player entity to PlayerState
func (tm *TransferManager) serializePlayerState(playerID uint64, world *engine.World) (*PlayerState, error) {
	entity, ok := world.GetEntity(playerID)
	if !ok || entity == nil {
		return nil, fmt.Errorf("player entity not found")
	}

	state := tm.createEmptyPlayerState(playerID)

	tm.serializePosition(entity, state)
	tm.serializeHealth(entity, state)
	tm.serializeInventory(entity, state)

	return state, nil
}

// createEmptyPlayerState creates a new PlayerState with initialized maps.
func (tm *TransferManager) createEmptyPlayerState(playerID uint64) *PlayerState {
	return &PlayerState{
		PlayerID:   playerID,
		Stats:      make(map[string]interface{}),
		Reputation: make(map[string]float64),
	}
}

// serializePosition extracts position component data into PlayerState.
func (tm *TransferManager) serializePosition(entity *engine.Entity, state *PlayerState) {
	if posComp, ok := entity.GetComponent("position"); ok {
		if pos, ok := posComp.(*engine.PositionComponent); ok {
			state.Position = &PositionData{
				X: pos.X,
				Y: pos.Y,
			}
		}
	}
}

// serializeHealth extracts health component data into PlayerState.
func (tm *TransferManager) serializeHealth(entity *engine.Entity, state *PlayerState) {
	if healthComp, ok := entity.GetComponent("health"); ok {
		if health, ok := healthComp.(*engine.HealthComponent); ok {
			state.Health = &HealthData{
				Current: health.Current,
				Max:     health.Max,
			}
		}
	}
}

// serializeInventory extracts inventory component data into PlayerState.
func (tm *TransferManager) serializeInventory(entity *engine.Entity, state *PlayerState) {
	if invComp, ok := entity.GetComponent("inventory"); ok {
		if inv, ok := invComp.(*engine.InventoryComponent); ok {
			state.Inventory = make([]ItemData, 0, len(inv.Items))
			for _, item := range inv.Items {
				state.Inventory = append(state.Inventory, ItemData{
					ItemID:   item.ID,
					TypeName: item.Name,
					Quantity: 1, // Items are stored individually, not stacked
				})
			}
		}
	}
}

// deserializePlayerState restores player entity from PlayerState
func (tm *TransferManager) deserializePlayerState(state *PlayerState, world *engine.World) error {
	entity, ok := world.GetEntity(state.PlayerID)
	if !ok || entity == nil {
		return fmt.Errorf("player entity not found")
	}

	// Restore position
	if state.Position != nil {
		if posComp, ok := entity.GetComponent("position"); ok {
			if pos, ok := posComp.(*engine.PositionComponent); ok {
				pos.X = state.Position.X
				pos.Y = state.Position.Y
			}
		}
	}

	// Restore health
	if state.Health != nil {
		if healthComp, ok := entity.GetComponent("health"); ok {
			if health, ok := healthComp.(*engine.HealthComponent); ok {
				health.Current = state.Health.Current
				health.Max = state.Health.Max
			}
		}
	}

	return nil
}

// computeStateHash computes SHA-256 hash of player state
func (tm *TransferManager) computeStateHash(state *PlayerState) (string, error) {
	data, err := json.Marshal(state)
	if err != nil {
		return "", err
	}

	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:]), nil
}

// VerifyStateHash verifies the integrity of player state
func (tm *TransferManager) VerifyStateHash(state *PlayerState, expectedHash string) error {
	actualHash, err := tm.computeStateHash(state)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %w", err)
	}

	if actualHash != expectedHash {
		return fmt.Errorf("state hash mismatch: expected %s, got %s", expectedHash, actualHash)
	}

	return nil
}
