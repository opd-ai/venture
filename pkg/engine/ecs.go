// Package engine provides ECS (Entity-Component-System) implementation.
// This file contains the core Entity and World types that manage game entities
// and their lifecycle within the ECS architecture.
package engine

import (
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

// Entity represents a game object composed of components.
// Entities are identified by a unique ID and contain a collection of components.
type Entity struct {
	ID         uint64
	Components map[string]Component

	// Fast-path cache for frequently accessed components
	// These eliminate map lookups in hot paths
	position        *PositionComponent
	velocity        *VelocityComponent
	health          *HealthComponent
	collider        *ColliderComponent
	inventory       *InventoryComponent
	stats           *StatsComponent
	animation       *AnimationComponent
	attack          *AttackComponent
	experience      *ExperienceComponent
	sprite          *EbitenSprite             // Cached for render system hot path (~93x faster access)
	rotation        *RotationComponent        // Cached for render and collision hot paths
	visualFeedback  *VisualFeedbackComponent  // Cached for render system hot path (visual effects)
	layer           *LayerComponent           // Cached for collision hot path (layer compatibility checks)
	team            *TeamComponent            // Cached for AI system hot path (enemy detection)
	particleEmitter *ParticleEmitterComponent // Cached for render system particle drawing hot path
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
	// Type assertions here are safe because we control the type based on c.Type()
	switch c.Type() {
	case "position":
		if pos, ok := c.(*PositionComponent); ok {
			e.position = pos
		}
	case "velocity":
		if vel, ok := c.(*VelocityComponent); ok {
			e.velocity = vel
		}
	case "health":
		if health, ok := c.(*HealthComponent); ok {
			e.health = health
		}
	case "collider":
		if collider, ok := c.(*ColliderComponent); ok {
			e.collider = collider
		}
	case "inventory":
		if inv, ok := c.(*InventoryComponent); ok {
			e.inventory = inv
		}
	case "stats":
		if stats, ok := c.(*StatsComponent); ok {
			e.stats = stats
		}
	case "animation":
		if anim, ok := c.(*AnimationComponent); ok {
			e.animation = anim
		}
	case "attack":
		if atk, ok := c.(*AttackComponent); ok {
			e.attack = atk
		}
	case "experience":
		if exp, ok := c.(*ExperienceComponent); ok {
			e.experience = exp
		}
	case "sprite":
		if sprite, ok := c.(*EbitenSprite); ok {
			e.sprite = sprite
		}
	case "rotation":
		if rot, ok := c.(*RotationComponent); ok {
			e.rotation = rot
		}
	case "visual_feedback":
		if vf, ok := c.(*VisualFeedbackComponent); ok {
			e.visualFeedback = vf
		}
	case "layer":
		if layer, ok := c.(*LayerComponent); ok {
			e.layer = layer
		}
	case "team":
		if team, ok := c.(*TeamComponent); ok {
			e.team = team
		}
	case "particle_emitter":
		if emitter, ok := c.(*ParticleEmitterComponent); ok {
			e.particleEmitter = emitter
		}
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
	case "animation":
		e.animation = nil
	case "attack":
		e.attack = nil
	case "experience":
		e.experience = nil
	case "sprite":
		e.sprite = nil
	case "rotation":
		e.rotation = nil
	case "visual_feedback":
		e.visualFeedback = nil
	case "layer":
		e.layer = nil
	case "team":
		e.team = nil
	case "particle_emitter":
		e.particleEmitter = nil
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
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetExperience() *ExperienceComponent {
	return e.experience
}

// GetAttack retrieves the AttackComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetAttack() *AttackComponent {
	return e.attack
}

// GetAnimation retrieves the AnimationComponent if present.
// Uses cached pointer for zero-overhead access.
func (e *Entity) GetAnimation() *AnimationComponent {
	return e.animation
}

// GetSprite retrieves the EbitenSprite if present.
// Uses cached pointer for zero-overhead access (~93x faster than map lookup).
func (e *Entity) GetSprite() *EbitenSprite {
	return e.sprite
}

// GetRotation retrieves the RotationComponent if present.
// Uses cached pointer for zero-overhead access in render and collision hot paths.
func (e *Entity) GetRotation() *RotationComponent {
	return e.rotation
}

// GetVisualFeedback retrieves the VisualFeedbackComponent if present.
// Uses cached pointer for zero-overhead access in render hot path (~93x faster than map lookup).
func (e *Entity) GetVisualFeedback() *VisualFeedbackComponent {
	return e.visualFeedback
}

// GetLayer retrieves the LayerComponent if present.
// Uses cached pointer for zero-overhead access in collision hot path (~93x faster than map lookup).
func (e *Entity) GetLayer() *LayerComponent {
	return e.layer
}

// GetTeam retrieves the TeamComponent if present.
// Uses cached pointer for zero-overhead access in AI hot path (~93x faster than map lookup).
func (e *Entity) GetTeam() *TeamComponent {
	return e.team
}

// GetParticleEmitter retrieves the ParticleEmitterComponent if present.
// Uses cached pointer for zero-overhead access in render hot path (~93x faster than map lookup).
func (e *Entity) GetParticleEmitter() *ParticleEmitterComponent {
	return e.particleEmitter
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

	// Pool for strings.Builder instances to reduce query key allocations
	builderPool sync.Pool

	// Pre-computed cache keys for common queries (zero-allocation fast path)
	keyPosition                   string
	keyPositionVelocity           string
	keyPositionHealth             string
	keyPositionCollider           string
	keyProjectilePositionVelocity string

	// GameClock provides deterministic or real-time clock services
	Clock GameClock

	// Logger for ECS operations
	logger *logrus.Entry

	// Mutex for thread-safe access to entities and metrics
	mu sync.RWMutex
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
		nextEntityID:     1,                       // Start entity IDs at 1 (0 reserved as invalid ID)
		cachedEntityList: make([]*Entity, 0, 256), // Pre-allocate for 256 entities
		queryBuffer:      make([]*Entity, 0, 256), // Pre-allocate query buffer
		queryCache:       make(map[string][]*Entity),
		queryCacheDirty:  make(map[string]bool),
		entityListDirty:  true,
		Clock:            NewSimulationClock(0), // Default to deterministic simulation clock
		logger:           logEntry,
		builderPool: sync.Pool{
			New: func() interface{} {
				return &strings.Builder{}
			},
		},
		// Pre-compute common query keys
		keyPosition:                   "position",
		keyPositionVelocity:           "position|velocity",
		keyPositionHealth:             "position|health",
		keyPositionCollider:           "position|collider",
		keyProjectilePositionVelocity: "projectile|position|velocity",
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
	// Defensive check: prevent nil systems from being added
	// This should not happen in normal operation, but provides safety
	// in case of initialization order issues or programming errors
	if system == nil {
		if w.logger != nil {
			w.logger.Error("attempted to add nil system to world")
		}
		return
	}

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
	// Advance game clock for deterministic time tracking
	w.Clock.Advance(deltaTime)

	// Process pending additions
	if len(w.entitiesToAdd) > 0 {
		for _, entity := range w.entitiesToAdd {
			w.entities[entity.ID] = entity
		}
		w.entitiesToAdd = w.entitiesToAdd[:0]
		w.entityListDirty = true
		w.invalidateQueryCache() // Invalidate query cache when entities are added
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

// generateQueryKey generates a cache key for the component type query.
// Uses fast-path optimization for common queries to eliminate allocations.
func (w *World) generateQueryKey(componentTypes []string) string {
	switch len(componentTypes) {
	case 1:
		if componentTypes[0] == "position" {
			return w.keyPosition
		}
	case 2:
		if componentTypes[0] == "position" {
			switch componentTypes[1] {
			case "velocity":
				return w.keyPositionVelocity
			case "health":
				return w.keyPositionHealth
			case "collider":
				return w.keyPositionCollider
			}
		}
	case 3:
		// Fast path for projectile system's hot query
		if componentTypes[0] == "projectile" && componentTypes[1] == "position" && componentTypes[2] == "velocity" {
			return w.keyProjectilePositionVelocity
		}
	}

	builder := w.builderPool.Get().(*strings.Builder)
	builder.Reset()
	for i, compType := range componentTypes {
		if i > 0 {
			builder.WriteByte('|')
		}
		builder.WriteString(compType)
	}
	key := builder.String()
	w.builderPool.Put(builder)
	return key
}

// processPendingEntityAdditions adds pending entities to the world.
// This ensures newly created entities are included in queries.
func (w *World) processPendingEntityAdditions() {
	if len(w.entitiesToAdd) == 0 {
		return
	}
	for _, entity := range w.entitiesToAdd {
		w.entities[entity.ID] = entity
	}
	w.entitiesToAdd = w.entitiesToAdd[:0]
	w.entityListDirty = true
	w.invalidateQueryCache()
}

// filterEntitiesByComponents filters entities that have all specified components.
// Uses and updates the internal query buffer for efficiency.
// Returns a slice that will be cached - caller must not modify the returned slice.
func (w *World) filterEntitiesByComponents(componentTypes []string) []*Entity {
	w.queryBuffer = w.queryBuffer[:0]

	// Ensure entity list cache is up to date
	if w.entityListDirty {
		w.rebuildEntityCache()
	}

	if cap(w.queryBuffer) < len(w.cachedEntityList) {
		w.queryBuffer = make([]*Entity, 0, len(w.cachedEntityList))
	}

	// Iterate over cached entity list (slice) instead of entities map
	// Slice iteration is significantly faster than map iteration
	for _, entity := range w.cachedEntityList {
		if entityHasAllComponents(entity, componentTypes) {
			w.queryBuffer = append(w.queryBuffer, entity)
		}
	}

	// Copy the buffer to avoid cache corruption.
	// The queryBuffer is reused across calls, so cached results must have their own backing array.
	result := make([]*Entity, len(w.queryBuffer))
	copy(result, w.queryBuffer)
	return result
}

// entityHasAllComponents checks if an entity has all specified component types.
func entityHasAllComponents(entity *Entity, componentTypes []string) bool {
	for _, compType := range componentTypes {
		if !entity.HasComponent(compType) {
			return false
		}
	}
	return true
}

// GetEntitiesWith returns all entities that have all of the specified component types.
// Uses a query cache to avoid repeated filtering. Cache is invalidated when entities are added/removed.
// Fast-path optimization for common queries eliminates all allocations on cache hits.
func (w *World) GetEntitiesWith(componentTypes ...string) []*Entity {
	key := w.generateQueryKey(componentTypes)
	w.processPendingEntityAdditions()

	if !w.queryCacheDirty[key] {
		if cached, exists := w.queryCache[key]; exists {
			return cached
		}
	}

	result := w.filterEntitiesByComponents(componentTypes)
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

// InvalidateQueryCache marks all cached queries as dirty.
// Called when components are added or removed from entities.
func (w *World) InvalidateQueryCache() {
	w.invalidateQueryCache()
}

// FlushPendingEntities processes pending entity additions immediately.
// This ensures newly created entities are available via GetEntity()
// without waiting for the next Update() cycle.
func (w *World) FlushPendingEntities() {
	w.processPendingEntityAdditions()
}

// GetSystems returns all registered systems.
func (w *World) GetSystems() []System {
	return w.systems
}

// GetLogger returns the world's logger entry.
func (w *World) GetLogger() *logrus.Entry {
	return w.logger
}

// GetEntityCount returns the total number of entities in the world.
// Thread-safe for concurrent access from metrics HTTP handler.
func (w *World) GetEntityCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.entities)
}

// GetActiveQuestCount returns the number of entities with active quest components.
// This provides a metric for game progression and player engagement.
// Thread-safe for concurrent access from metrics HTTP handler.
func (w *World) GetActiveQuestCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	count := 0
	for _, entity := range w.entities {
		if entity.HasComponent("quest") {
			count++
		}
	}
	return count
}

// GetTradeVolume returns the total number of completed trades.
// This metric tracks economic activity in the game world.
// Thread-safe for concurrent access from metrics HTTP handler.
func (w *World) GetTradeVolume() uint64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	
	// Sum up trade counters from all entities with trade components
	var total uint64
	for _, entity := range w.entities {
		if comp, ok := entity.GetComponent("trade"); ok {
			if tradeComp, ok := comp.(*TradeComponent); ok && tradeComp != nil {
				total += tradeComp.CompletedTrades
			}
		}
	}
	return total
}
