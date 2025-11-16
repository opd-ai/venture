// Package engine provides entity persistence for saving/loading entity state.
// This file implements component serialization for world state persistence.
package engine

import (
	"encoding/json"
	"errors"
	"fmt"
)

// ComponentSerializer defines the interface for components that can be serialized
type ComponentSerializer interface {
	Serialize() ([]byte, error)
	Deserialize(data []byte) error
}

// EntityLifecycleTracker tracks entity spawning, modification, and death
type EntityLifecycleTracker struct {
	spawned  map[uint64]bool // Entities spawned this session
	modified map[uint64]bool // Entities modified since last save
	killed   map[uint64]bool // Entities killed (to prevent respawn)
}

// NewEntityLifecycleTracker creates a new lifecycle tracker
func NewEntityLifecycleTracker() *EntityLifecycleTracker {
	return &EntityLifecycleTracker{
		spawned:  make(map[uint64]bool),
		modified: make(map[uint64]bool),
		killed:   make(map[uint64]bool),
	}
}

// MarkSpawned marks an entity as spawned
func (e *EntityLifecycleTracker) MarkSpawned(entityID uint64) {
	e.spawned[entityID] = true
}

// MarkModified marks an entity as modified
func (e *EntityLifecycleTracker) MarkModified(entityID uint64) {
	e.modified[entityID] = true
}

// MarkKilled marks an entity as killed
func (e *EntityLifecycleTracker) MarkKilled(entityID uint64) {
	e.killed[entityID] = true
	delete(e.spawned, entityID)
	delete(e.modified, entityID)
}

// IsKilled checks if an entity was killed
func (e *EntityLifecycleTracker) IsKilled(entityID uint64) bool {
	return e.killed[entityID]
}

// IsModified checks if an entity was modified
func (e *EntityLifecycleTracker) IsModified(entityID uint64) bool {
	return e.modified[entityID]
}

// GetModifiedEntities returns all modified entity IDs
func (e *EntityLifecycleTracker) GetModifiedEntities() []uint64 {
	result := make([]uint64, 0, len(e.modified))
	for id := range e.modified {
		result = append(result, id)
	}
	return result
}

// ClearModified clears the modified entities list (after save)
func (e *EntityLifecycleTracker) ClearModified() {
	e.modified = make(map[uint64]bool)
}

// RespawnRule determines if an entity should respawn after being killed
type RespawnRule int

const (
	// RespawnNever means the entity never respawns (NPCs, unique items)
	RespawnNever RespawnRule = iota
	// RespawnAlways means the entity respawns after death (monsters)
	RespawnAlways
	// RespawnConditional means respawn depends on world state (bosses)
	RespawnConditional
)

// GetRespawnRule determines the respawn behavior for an entity type
func GetRespawnRule(typeName string) RespawnRule {
	switch typeName {
	case "Monster":
		return RespawnAlways
	case "NPC", "Merchant", "Companion":
		return RespawnNever
	case "Boss":
		return RespawnConditional
	case "Item", "Weapon", "Armor", "Consumable":
		return RespawnNever
	default:
		return RespawnNever
	}
}

// SerializeEntity serializes an entity to EntityState for persistence
func SerializeEntity(entity *Entity) (*EntityState, error) {
	if entity == nil {
		return nil, errors.New("entity is nil")
	}

	state := &EntityState{
		ID:         entity.ID,
		TypeName:   getEntityTypeName(entity),
		Components: make(map[string][]byte),
	}

	// Serialize each component
	for _, component := range entity.Components {
		if serializer, ok := component.(ComponentSerializer); ok {
			data, err := serializer.Serialize()
			if err != nil {
				return nil, fmt.Errorf("failed to serialize component %s: %w", component.Type(), err)
			}
			state.Components[component.Type()] = data
		} else {
			// Fallback to JSON serialization for components without Serialize method
			data, err := json.Marshal(component)
			if err != nil {
				return nil, fmt.Errorf("failed to JSON serialize component %s: %w", component.Type(), err)
			}
			state.Components[component.Type()] = data
		}
	}

	return state, nil
}

// DeserializeEntity creates an entity from EntityState
func DeserializeEntity(state *EntityState, world *World) (*Entity, error) {
	if state == nil {
		return nil, errors.New("entity state is nil")
	}
	if world == nil {
		return nil, errors.New("world is nil")
	}

	// Create new entity (world will assign ID, then we'll update it)
	entity := world.CreateEntity()

	// Update entity ID to match saved state (after world processes additions)
	world.Update(0.0) // Force processing of pending additions
	entity.ID = state.ID

	// Deserialize each component
	for componentType, data := range state.Components {
		component, err := createComponentByType(componentType)
		if err != nil {
			return nil, fmt.Errorf("failed to create component %s: %w", componentType, err)
		}

		if serializer, ok := component.(ComponentSerializer); ok {
			if err := serializer.Deserialize(data); err != nil {
				return nil, fmt.Errorf("failed to deserialize component %s: %w", componentType, err)
			}
		} else {
			// Fallback to JSON deserialization
			if err := json.Unmarshal(data, component); err != nil {
				return nil, fmt.Errorf("failed to JSON deserialize component %s: %w", componentType, err)
			}
		}

		entity.AddComponent(component)
	}

	return entity, nil
}

// getEntityTypeName extracts the entity type name from components
func getEntityTypeName(entity *Entity) string {
	// Check for AI component to determine monster/NPC
	if aiComp, ok := entity.GetComponent("ai"); ok {
		if ai, ok := aiComp.(*AIComponent); ok {
			// Assume attacking/chasing states indicate monsters
			if ai.State == AIStateAttack || ai.State == AIStateChase {
				return "Monster"
			}
			return "NPC"
		}
	}

	// Check for item-related components
	if entity.HasComponent("weapon") {
		return "Weapon"
	}
	if entity.HasComponent("armor") {
		return "Armor"
	}
	if entity.HasComponent("consumable") {
		return "Consumable"
	}

	// Check for companion
	if entity.HasComponent("companion") {
		return "Companion"
	}

	// Default to generic entity
	return "Entity"
}

// createComponentByType creates a component instance by type name
func createComponentByType(componentType string) (Component, error) {
	switch componentType {
	case "position":
		return &PositionComponent{}, nil
	case "velocity":
		return &VelocityComponent{}, nil
	case "health":
		return &HealthComponent{}, nil
	case "stats":
		return &StatsComponent{}, nil
	case "base_stats":
		return &BaseStatsComponent{}, nil
	case "ai":
		return &AIComponent{}, nil
	case "inventory":
		return &InventoryComponent{}, nil
	case "experience":
		return &ExperienceComponent{}, nil
	case "collider":
		return &ColliderComponent{}, nil
	case "animation":
		return &AnimationComponent{}, nil
	case "companion":
		return &CompanionComponent{}, nil
	case "vehicle":
		return &VehicleComponent{}, nil
	case "mount":
		return &MountComponent{}, nil
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}
}

// EntityState represents serializable entity data
// This is defined in pkg/world/persistence.go but duplicated here for engine usage
type EntityState struct {
	ID         uint64            `json:"id"`
	TypeName   string            `json:"type_name"`  // "Monster", "NPC", "Item"
	Components map[string][]byte `json:"components"` // Serialized component data
}
