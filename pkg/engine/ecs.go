// Package engine provides ECS (Entity-Component-System) implementation.
// This file contains the core Entity and World types that manage game entities
// and their lifecycle within the ECS architecture.
package engine

import (
	"github.com/sirupsen/logrus"
)

// Entity represents a game object composed of components.
// Entities are identified by a unique ID and contain a collection of components.
type Entity struct {
	ID         uint64
	Components map[string]Component

	// Fast-path cache for frequently accessed components
	// These eliminate map lookups in hot paths
	position  *PositionComponent
	velocity  *VelocityComponent
	health    *HealthComponent
	collider  *ColliderComponent
	inventory *InventoryComponent
	stats     *StatsComponent
}

// NewEntity creates a new entity with the given ID.
func NewEntity(id uint64) *Entity {
	return &Entity{
		ID:         id,
		Components: make(map[string]Component),
	}
}

// AddComponent adds a component to this entity.
func (e *Entity) AddComponent(c Component) {
	e.Components[c.Type()] = c

	// Update fast-path cache for hot components
	switch c.Type() {
	case "position":
		e.position = c.(*PositionComponent)
	case "velocity":
		e.velocity = c.(*VelocityComponent)
	case "health":
		e.health = c.(*HealthComponent)
	case "collider":
		e.collider = c.(*ColliderComponent)
	case "inventory":
		e.inventory = c.(*InventoryComponent)
	case "stats":
		e.stats = c.(*StatsComponent)
	}
}

// AddComponentWithLogger adds a component to this entity with logging.
func (e *Entity) AddComponentWithLogger(c Component, logger *logrus.Entry) {
	e.Components[c.Type()] = c
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entityID":      e.ID,
			"componentType": c.Type(),
		}).Debug("component added")
	}
}

// GetComponent retrieves a component by type.
func (e *Entity) GetComponent(componentType string) (Component, bool) {
	c, ok := e.Components[componentType]
	return c, ok
}

// RemoveComponent removes a component from this entity.
func (e *Entity) RemoveComponent(componentType string) {
	delete(e.Components, componentType)

	// Clear fast-path cache for hot components
	switch componentType {
	case "position":
		e.position = nil
	case "velocity":
		e.velocity = nil
	case "health":
		e.health = nil
	case "collider":
		e.collider = nil
	case "inventory":
		e.inventory = nil
	case "stats":
		e.stats = nil
	}
}

// RemoveComponentWithLogger removes a component from this entity with logging.
func (e *Entity) RemoveComponentWithLogger(componentType string, logger *logrus.Entry) {
	delete(e.Components, componentType)
	if logger != nil {
		logger.WithFields(logrus.Fields{
			"entityID":      e.ID,
			"componentType": componentType,
		}).Debug("component removed")
	}
}

// HasComponent checks if this entity has a component of the given type.
func (e *Entity) HasComponent(componentType string) bool {
	_, ok := e.Components[componentType]
	return ok
}

// Typed component getters for hot path optimization.
// These eliminate map lookups and type assertions in performance-critical loops.

// GetPosition retrieves the PositionComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetPosition() *PositionComponent {
	return e.position
}

// GetVelocity retrieves the VelocityComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetVelocity() *VelocityComponent {
	return e.velocity
}

// GetHealth retrieves the HealthComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetHealth() *HealthComponent {
	return e.health
}

// GetCollider retrieves the ColliderComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetCollider() *ColliderComponent {
	return e.collider
}

// GetInventory retrieves the InventoryComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetInventory() *InventoryComponent {
	return e.inventory
}

// GetStats retrieves the StatsComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetStats() *StatsComponent {
	return e.stats
}

// GetExperience retrieves the ExperienceComponent if present.
func (e *Entity) GetExperience() *ExperienceComponent {
	if comp, ok := e.Components["experience"]; ok {
		return comp.(*ExperienceComponent)
	}
	return nil
}

// GetAttack retrieves the AttackComponent if present.
func (e *Entity) GetAttack() *AttackComponent {
	if comp, ok := e.Components["attack"]; ok {
		return comp.(*AttackComponent)
	}
	return nil
}

// GetAnimation retrieves the AnimationComponent if present.
func (e *Entity) GetAnimation() *AnimationComponent {
	if comp, ok := e.Components["animation"]; ok {
		return comp.(*AnimationComponent)
	}
	return nil
}

// World manages all entities and systems in the game.
type World struct {
	entities          map[uint64]*Entity
	systems           []System
	nextEntityID      uint64
	entitiesToAdd     []*Entity
	entityIDsToRemove []uint64

	// Cached entity list to reduce allocations
	cachedEntityList []*Entity
	entityListDirty  bool

	// Reusable buffer for entity queries to reduce allocations
	queryBuffer []*Entity

	// Query cache: map[component types] -> []*Entity
	queryCache      map[string][]*Entity
	queryCacheDirty map[string]bool

	// Logger for ECS operations
	logger *logrus.Entry
}

// NewWorld creates a new game world.
func NewWorld() *World {
	return NewWorldWithLogger(nil)
}

// NewWorldWithLogger creates a new game world with a logger.
func NewWorldWithLogger(logger *logrus.Logger) *World {
	var logEntry *logrus.Entry
	if logger != nil {
		logEntry = logger.WithFields(logrus.Fields{
			"system": "ecs",
		})
	}

	w := &World{
		entities:         make(map[uint64]*Entity),
		systems:          make([]System, 0),
		cachedEntityList: make([]*Entity, 0, 256), // Pre-allocate for 256 entities
		queryBuffer:      make([]*Entity, 0, 256), // Pre-allocate query buffer
		queryCache:       make(map[string][]*Entity),
		queryCacheDirty:  make(map[string]bool),
		entityListDirty:  true,
		logger:           logEntry,
	}

	if w.logger != nil {
		w.logger.Debug("world created")
	}

	return w
}

// CreateEntity creates a new entity and adds it to the world.
func (w *World) CreateEntity() *Entity {
	id := w.nextEntityID
	w.nextEntityID++
	entity := NewEntity(id)
	w.entitiesToAdd = append(w.entitiesToAdd, entity)

	if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
		w.logger.WithField("entityID", id).Debug("entity created")
	}

	return entity
}

// AddEntity adds an existing entity to the world.
func (w *World) AddEntity(entity *Entity) {
	w.entitiesToAdd = append(w.entitiesToAdd, entity)
	w.entityListDirty = true
	w.invalidateQueryCache()
}

// RemoveEntity marks an entity for removal from the world.
func (w *World) RemoveEntity(entityID uint64) {
	w.entityIDsToRemove = append(w.entityIDsToRemove, entityID)
	w.entityListDirty = true
	w.invalidateQueryCache()

	if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
		w.logger.WithField("entityID", entityID).Debug("entity marked for removal")
	}
}

// GetEntity retrieves an entity by ID.
func (w *World) GetEntity(entityID uint64) (*Entity, bool) {
	entity, ok := w.entities[entityID]
	return entity, ok
}

// AddSystem adds a system to the world.
func (w *World) AddSystem(system System) {
	w.systems = append(w.systems, system)

	if w.logger != nil {
		// Get system name if available
		systemName := "unknown"
		if named, ok := system.(interface{ Name() string }); ok {
			systemName = named.Name()
		}
		w.logger.WithField("system", systemName).Debug("system added")
	}
}

// Update updates all systems with the current entity list.
func (w *World) Update(deltaTime float64) {
	// Process pending additions
	if len(w.entitiesToAdd) > 0 {
		for _, entity := range w.entitiesToAdd {
			w.entities[entity.ID] = entity
		}
		w.entitiesToAdd = w.entitiesToAdd[:0]
		w.entityListDirty = true
	}

	// Process pending removals
	if len(w.entityIDsToRemove) > 0 {
		for _, id := range w.entityIDsToRemove {
			delete(w.entities, id)
		}
		w.entityIDsToRemove = w.entityIDsToRemove[:0]
		w.entityListDirty = true
	}

	// Rebuild cached entity list if needed
	if w.entityListDirty {
		w.rebuildEntityCache()
	}

	// Update all systems with cached list
	for _, system := range w.systems {
		system.Update(w.cachedEntityList, deltaTime)
	}
}

// rebuildEntityCache rebuilds the cached entity list.
func (w *World) rebuildEntityCache() {
	// Reuse existing slice capacity
	w.cachedEntityList = w.cachedEntityList[:0]

	// Ensure capacity
	if cap(w.cachedEntityList) < len(w.entities) {
		w.cachedEntityList = make([]*Entity, 0, len(w.entities))
	}

	for _, entity := range w.entities {
		w.cachedEntityList = append(w.cachedEntityList, entity)
	}

	w.entityListDirty = false
}

// GetEntities returns all entities in the world.
// Note: Returns the cached list, do not modify.
func (w *World) GetEntities() []*Entity {
	if w.entityListDirty {
		w.rebuildEntityCache()
	}
	return w.cachedEntityList
}

// GetEntitiesWith returns all entities that have all of the specified component types.
// Uses a query cache to avoid repeated filtering. Cache is invalidated when entities are added/removed.
func (w *World) GetEntitiesWith(componentTypes ...string) []*Entity {
	// Generate cache key from component types
	// Use a simple string concatenation with separator
	key := ""
	for i, compType := range componentTypes {
		if i > 0 {
			key += "|"
		}
		key += compType
	}

	// Check if cache is valid
	if !w.queryCacheDirty[key] {
		if cached, exists := w.queryCache[key]; exists {
			return cached
		}
	}

	// Cache miss or dirty - rebuild query result
	// Reuse buffer, reset length to 0
	w.queryBuffer = w.queryBuffer[:0]

	// Ensure capacity
	if cap(w.queryBuffer) < len(w.entities) {
		w.queryBuffer = make([]*Entity, 0, len(w.entities))
	}

	for _, entity := range w.entities {
		hasAll := true
		for _, compType := range componentTypes {
			if !entity.HasComponent(compType) {
				hasAll = false
				break
			}
		}
		if hasAll {
			w.queryBuffer = append(w.queryBuffer, entity)
		}
	}

	// Cache the result (make a copy to avoid queryBuffer reuse issues)
	result := make([]*Entity, len(w.queryBuffer))
	copy(result, w.queryBuffer)
	w.queryCache[key] = result
	w.queryCacheDirty[key] = false

	return result
}

// invalidateQueryCache marks all cached queries as dirty.
// Called when entities are added or removed from the world.
func (w *World) invalidateQueryCache() {
	for key := range w.queryCache {
		w.queryCacheDirty[key] = true
	}
}

// GetSystems returns all registered systems.
func (w *World) GetSystems() []System {
	return w.systems
}

// GetLogger returns the world's logger entry.
func (w *World) GetLogger() *logrus.Entry {
	return w.logger
}
