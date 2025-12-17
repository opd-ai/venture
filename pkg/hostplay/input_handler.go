package hostplay

import (
	"encoding/json"
	"math"

	"github.com/opd-ai/venture/pkg/engine"
	"github.com/sirupsen/logrus"
)

// InputHandler processes player input commands and updates entity states.
type InputHandler struct {
	world     *engine.World
	playerMap map[uint64]*engine.Entity // Maps player IDs to their entities
	moveSpeed float64
	logger    *logrus.Entry
}

// NewInputHandler creates a new input handler.
func NewInputHandler(world *engine.World, logger *logrus.Entry) *InputHandler {
	return &InputHandler{
		world:     world,
		playerMap: make(map[uint64]*engine.Entity),
		moveSpeed: 200.0, // pixels per second
		logger:    logger,
	}
}

// RegisterPlayer registers a player entity for input processing.
func (h *InputHandler) RegisterPlayer(playerID uint64, entity *engine.Entity) {
	h.playerMap[playerID] = entity
	h.logger.WithField("player_id", playerID).Debug("player registered for input")
}

// UnregisterPlayer removes a player from input processing.
func (h *InputHandler) UnregisterPlayer(playerID uint64) {
	delete(h.playerMap, playerID)
	h.logger.WithField("player_id", playerID).Debug("player unregistered from input")
}

// ProcessInputRaw processes raw input data (deserializes JSON first).
func (h *InputHandler) ProcessInputRaw(playerID uint64, inputType string, rawData []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(rawData, &data); err != nil {
		h.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"error":     err,
		}).Warn("failed to deserialize input data")
		return
	}
	h.ProcessInput(playerID, inputType, data)
}

// ProcessInput processes an input command for a player.
func (h *InputHandler) ProcessInput(playerID uint64, inputType string, data map[string]interface{}) {
	entity, exists := h.playerMap[playerID]
	if !exists {
		h.logger.WithField("player_id", playerID).Warn("input received for unknown player")
		return
	}

	switch inputType {
	case "move":
		h.processMovement(entity, data)
	case "attack":
		h.processAttack(entity, data)
	case "use_item":
		h.processItemUse(entity, data)
	case "interact":
		h.processInteraction(entity, data)
	default:
		h.logger.WithFields(logrus.Fields{
			"player_id": playerID,
			"type":      inputType,
		}).Debug("unknown input type")
	}
}

// processMovement updates entity velocity based on movement input.
func (h *InputHandler) processMovement(entity *engine.Entity, data map[string]interface{}) {
	comp, ok := entity.GetComponent("velocity")
	if !ok {
		return
	}

	velocity, ok := comp.(*engine.VelocityComponent)
	if !ok {
		return
	}

	// Extract direction from input data
	dx, dxOk := data["dx"].(float64)
	dy, dyOk := data["dy"].(float64)
	if !dxOk || !dyOk {
		return
	}

	// Normalize and apply speed
	magnitudeSq := dx*dx + dy*dy
	if magnitudeSq > 0 {
		// Use sqrt for proper unit vector normalization
		invMagnitude := 1.0 / math.Sqrt(magnitudeSq)
		velocity.VX = dx * invMagnitude * h.moveSpeed
		velocity.VY = dy * invMagnitude * h.moveSpeed
	} else {
		velocity.VX = 0
		velocity.VY = 0
	}
}

// processAttack triggers an attack action for the entity.
func (h *InputHandler) processAttack(entity *engine.Entity, data map[string]interface{}) {
	// Get aim component if it exists
	aimComp, ok := entity.GetComponent("aim")
	if !ok {
		return
	}

	aim, ok := aimComp.(*engine.AimComponent)
	if !ok {
		return
	}

	// Extract target angle from input
	if angle, ok := data["angle"].(float64); ok {
		aim.AimAngle = angle
	}

	// Trigger attack via combat system
	// The combat system will handle the actual attack logic in its update
	h.logger.WithField("entity_id", entity.ID).Debug("attack triggered")
}

// processItemUse triggers item usage for the entity.
func (h *InputHandler) processItemUse(entity *engine.Entity, data map[string]interface{}) {
	// Extract item slot
	slot, ok := data["slot"].(float64)
	if !ok {
		return
	}

	h.logger.WithFields(logrus.Fields{
		"entity_id": entity.ID,
		"slot":      int(slot),
	}).Debug("item use triggered")

	// Item use will be processed by the item use system
}

// processInteraction triggers interaction for the entity.
func (h *InputHandler) processInteraction(entity *engine.Entity, data map[string]interface{}) {
	h.logger.WithField("entity_id", entity.ID).Debug("interaction triggered")
	// Interaction will be processed by the interaction system
}

// GetPlayerEntity returns the entity for a given player ID.
func (h *InputHandler) GetPlayerEntity(playerID uint64) (*engine.Entity, bool) {
	entity, exists := h.playerMap[playerID]
	return entity, exists
}

// PlayerCount returns the number of registered players.
func (h *InputHandler) PlayerCount() int {
	return len(h.playerMap)
}
