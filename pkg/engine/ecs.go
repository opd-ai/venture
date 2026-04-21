// Package engine provides ECS (Entity-Component-System) implementation.
// This file contains the core Entity and World types that manage game entities
// and their lifecycle within the ECS architecture.
package engine

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/opd-ai/venture/pkg/companion/learning"
	"github.com/opd-ai/venture/pkg/integration/guild_vehicle"
	"github.com/sirupsen/logrus"
)

// budgetWarn* constants control rate-limiting of per-system frame-budget exceeded warnings.
// Warnings are emitted on the 1st, 10th, 100th, and every 1000th occurrence.
const (
	budgetWarnFirst  = 1
	budgetWarnSecond = 10
	budgetWarnThird  = 100
	budgetWarnPeriod = 1000
)

// Entity represents a game object composed of components.
// Entity is identified by a unique ID and contains a collection of components.
type Entity struct {
	ID         uint64
	Components map[string]Component

	// Fast-path cache for frequently accessed components
	// These eliminate map lookups in hot paths
	position          *PositionComponent
	velocity          *VelocityComponent
	health            *HealthComponent
	collider          *ColliderComponent
	inventory         *InventoryComponent
	stats             *StatsComponent
	animation         *AnimationComponent
	attack            *AttackComponent
	experience        *ExperienceComponent
	sprite            *EbitenSprite               // Cached for render system hot path (~93x faster access)
	rotation          *RotationComponent          // Cached for render and collision hot paths
	visualFeedback    *VisualFeedbackComponent    // Cached for render system hot path (visual effects)
	layer             *LayerComponent             // Cached for collision hot path (layer compatibility checks)
	team              *TeamComponent              // Cached for AI system hot path (enemy detection)
	particleEmitter   *ParticleEmitterComponent   // Cached for render system particle drawing hot path
	dropShadow        *DropShadowComponent        // Cached for render system drop shadow hot path
	weatherTint       *WeatherSpriteTintComponent // Cached for render system tint composition hot path
	creatureGenreTint *CreatureGenreTintComponent // Cached for render system tint composition hot path
	prestigeComp      Component                   // Cached for prestige visual tier checks (Phase 4.2) - stored as Component to avoid import cycle

	// Integration component caches for cross-system hot paths
	companionLearning *learning.CompanionLearningComponent      // Cached for companion AI learning system hot path
	guildVehicleFleet *guild_vehicle.GuildVehicleFleetComponent // Cached for guild vehicle fleet system hot path
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
	e.updateComponentCache(c)
}

// updateComponentCache updates the fast-path cache for hot components.
func (e *Entity) updateComponentCache(c Component) {
	switch c.Type() {
	case "position":
		e.cachePosition(c)
	case "velocity":
		e.cacheVelocity(c)
	case "health":
		e.cacheHealth(c)
	case "collider":
		e.cacheCollider(c)
	case "inventory":
		e.cacheInventory(c)
	case "stats":
		e.cacheStats(c)
	case "animation":
		e.cacheAnimation(c)
	case "attack":
		e.cacheAttack(c)
	case "experience":
		e.cacheExperience(c)
	case "sprite":
		e.cacheSprite(c)
	case "rotation":
		e.cacheRotation(c)
	case "visual_feedback":
		e.cacheVisualFeedback(c)
	case "layer":
		e.cacheLayer(c)
	case "team":
		e.cacheTeam(c)
	case "particle_emitter":
		e.cacheParticleEmitter(c)
	case "drop_shadow":
		e.cacheDropShadow(c)
	case "weather_sprite_tint":
		e.cacheWeatherSpriteTint(c)
	case "creature_genre_tint":
		e.cacheCreatureGenreTint(c)
	case "prestige":
		e.cachePrestige(c)
	case "companion_learning":
		e.cacheCompanionLearning(c)
	case "guild_vehicle_fleet":
		e.cacheGuildVehicleFleet(c)
	}
}

// cachePosition updates the position component cache.
func (e *Entity) cachePosition(c Component) {
	if pos, ok := c.(*PositionComponent); ok {
		e.position = pos
	}
}

// cacheVelocity updates the velocity component cache.
func (e *Entity) cacheVelocity(c Component) {
	if vel, ok := c.(*VelocityComponent); ok {
		e.velocity = vel
	}
}

// cacheHealth updates the health component cache.
func (e *Entity) cacheHealth(c Component) {
	if health, ok := c.(*HealthComponent); ok {
		e.health = health
	}
}

// cacheCollider updates the collider component cache.
func (e *Entity) cacheCollider(c Component) {
	if collider, ok := c.(*ColliderComponent); ok {
		e.collider = collider
	}
}

// cacheInventory updates the inventory component cache.
func (e *Entity) cacheInventory(c Component) {
	if inv, ok := c.(*InventoryComponent); ok {
		e.inventory = inv
	}
}

// cacheStats updates the stats component cache.
func (e *Entity) cacheStats(c Component) {
	if stats, ok := c.(*StatsComponent); ok {
		e.stats = stats
	}
}

// cacheAnimation updates the animation component cache.
func (e *Entity) cacheAnimation(c Component) {
	if anim, ok := c.(*AnimationComponent); ok {
		e.animation = anim
	}
}

// cacheAttack updates the attack component cache.
func (e *Entity) cacheAttack(c Component) {
	if atk, ok := c.(*AttackComponent); ok {
		e.attack = atk
	}
}

// cacheExperience updates the experience component cache.
func (e *Entity) cacheExperience(c Component) {
	if exp, ok := c.(*ExperienceComponent); ok {
		e.experience = exp
	}
}

// cacheSprite updates the sprite component cache.
func (e *Entity) cacheSprite(c Component) {
	if sprite, ok := c.(*EbitenSprite); ok {
		e.sprite = sprite
	}
}

// cacheRotation updates the rotation component cache.
func (e *Entity) cacheRotation(c Component) {
	if rot, ok := c.(*RotationComponent); ok {
		e.rotation = rot
	}
}

// cacheVisualFeedback updates the visual feedback component cache.
func (e *Entity) cacheVisualFeedback(c Component) {
	if vf, ok := c.(*VisualFeedbackComponent); ok {
		e.visualFeedback = vf
	}
}

// cacheLayer updates the layer component cache.
func (e *Entity) cacheLayer(c Component) {
	if layer, ok := c.(*LayerComponent); ok {
		e.layer = layer
	}
}

// cacheTeam updates the team component cache.
func (e *Entity) cacheTeam(c Component) {
	if team, ok := c.(*TeamComponent); ok {
		e.team = team
	}
}

// cacheParticleEmitter updates the particle emitter component cache.
func (e *Entity) cacheParticleEmitter(c Component) {
	if emitter, ok := c.(*ParticleEmitterComponent); ok {
		e.particleEmitter = emitter
	}
}

// cacheWeatherSpriteTint updates the weather sprite tint component cache.
func (e *Entity) cacheWeatherSpriteTint(c Component) {
	if wt, ok := c.(*WeatherSpriteTintComponent); ok {
		e.weatherTint = wt
	}
}

// cacheCreatureGenreTint updates the creature genre tint component cache.
func (e *Entity) cacheCreatureGenreTint(c Component) {
	if ct, ok := c.(*CreatureGenreTintComponent); ok {
		e.creatureGenreTint = ct
	}
}

// cachePrestige updates the prestige component cache.
// Stores as Component interface to avoid circular import with pkg/engine/prestige.
func (e *Entity) cachePrestige(c Component) {
	// Store prestige component directly (type checking happens at runtime via component Type() method)
	e.prestigeComp = c
}

// cacheCompanionLearning updates the companion learning component cache.
func (e *Entity) cacheCompanionLearning(c Component) {
	if cl, ok := c.(*learning.CompanionLearningComponent); ok {
		e.companionLearning = cl
	}
}

// cacheGuildVehicleFleet updates the guild vehicle fleet component cache.
func (e *Entity) cacheGuildVehicleFleet(c Component) {
	if gvf, ok := c.(*guild_vehicle.GuildVehicleFleetComponent); ok {
		e.guildVehicleFleet = gvf
	}
}

// AddComponentWithLogger adds a component to this entity with logging.
func (e *Entity) AddComponentWithLogger(c Component, logger *logrus.Entry) {
	e.Components[c.Type()] = c
	e.updateComponentCache(c)
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
	case "drop_shadow":
		e.dropShadow = nil
	case "weather_sprite_tint":
		e.weatherTint = nil
	case "creature_genre_tint":
		e.creatureGenreTint = nil
	case "companion_learning":
		e.companionLearning = nil
	case "guild_vehicle_fleet":
		e.guildVehicleFleet = nil
	}
}

// RemoveComponentWithLogger removes a component from this entity with logging.
func (e *Entity) RemoveComponentWithLogger(componentType string, logger *logrus.Entry) {
	e.RemoveComponent(componentType)
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

// cacheDropShadow updates the drop shadow component cache.
func (e *Entity) cacheDropShadow(c Component) {
	if ds, ok := c.(*DropShadowComponent); ok {
		e.dropShadow = ds
	}
}

// GetDropShadow retrieves the DropShadowComponent if present.
// Uses cached pointer for zero-overhead access in render hot path.
func (e *Entity) GetDropShadow() *DropShadowComponent {
	return e.dropShadow
}

// GetWeatherSpriteTint retrieves the WeatherSpriteTintComponent if present.
// Uses cached pointer for zero-overhead access in render hot path (~93x faster than map lookup).
func (e *Entity) GetWeatherSpriteTint() *WeatherSpriteTintComponent {
	return e.weatherTint
}

// GetCreatureGenreTint retrieves the CreatureGenreTintComponent if present.
// Uses cached pointer for zero-overhead access in render hot path (~93x faster than map lookup).
func (e *Entity) GetCreatureGenreTint() *CreatureGenreTintComponent {
	return e.creatureGenreTint
}

// GetCompanionLearning retrieves the CompanionLearningComponent if present.
// Uses cached pointer for zero-overhead access in companion AI learning hot path (~93x faster than map lookup).
func (e *Entity) GetCompanionLearning() *learning.CompanionLearningComponent {
	return e.companionLearning
}

// GetGuildVehicleFleet retrieves the GuildVehicleFleetComponent if present.
// Uses cached pointer for zero-overhead access in guild vehicle fleet hot path (~93x faster than map lookup).
func (e *Entity) GetGuildVehicleFleet() *guild_vehicle.GuildVehicleFleetComponent {
	return e.guildVehicleFleet
}

// GetPrestige retrieves the prestige Component if present.
// Uses cached reference for zero-overhead access in prestige visual tier rendering hot path (~93x faster than map lookup).
// Returns Component interface to avoid circular imports with pkg/engine/prestige.
// Callers should use type assertion to access prestige-specific fields.
func (e *Entity) GetPrestige() Component {
	return e.prestigeComp
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

	// queryComponents maps each cache key to the component types that query requires.
	// Used by selective cache invalidation to avoid dirtying unrelated queries.
	queryComponents map[string][]string

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

	// ModRules provides access to mod-defined rule values
	// Phase 6.3 (PLAN.md): Modding System Integration
	ModRules ModRuleProvider

	// Logger for ECS operations
	logger *logrus.Entry

	// Performance metrics for frame time tracking
	performanceMetrics *PerformanceMetrics

	// systemBudget is the per-system frame time budget. Systems that exceed this
	// budget trigger a rate-limited logrus.Warn. Default: 2 ms. Zero disables enforcement.
	systemBudget time.Duration

	// budgetWarnCount tracks how many times a given system has exceeded its budget
	// this session. Used for rate-limiting budget-exceeded warnings.
	budgetWarnCount map[string]int

	// Cache of system names to avoid per-frame reflection (eliminates 2,640 reflection calls/sec at 60 FPS with 44 systems)
	systemNameCache map[System]string

	// entityRemovalHooks are called on the game-loop goroutine for each entity ID
	// that is drained from entityIDsToRemove during World.Update. Systems that
	// maintain per-entity state (e.g. MovementSystem.visitedCells) register a hook
	// here so that state is reclaimed when the entity is removed.
	// Hooks are invoked without holding any mutex and must not call RemoveEntity.
	entityRemovalHooks []func(entityID uint64)

	// Mutex for thread-safe access to entities and metrics
	mu sync.RWMutex

	// entityMu protects the entity staging buffers (entitiesToAdd,
	// entityIDsToRemove, nextEntityID) which can be written from any
	// goroutine (e.g. server player-join handlers) while Update() reads
	// and drains them on the game-loop goroutine.
	entityMu sync.Mutex
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
		entities:           make(map[uint64]*Entity),
		systems:            make([]System, 0),
		nextEntityID:       1,                       // Start entity IDs at 1 (0 reserved as invalid ID)
		cachedEntityList:   make([]*Entity, 0, 256), // Pre-allocate for 256 entities
		queryBuffer:        make([]*Entity, 0, 256), // Pre-allocate query buffer
		queryCache:         make(map[string][]*Entity),
		queryCacheDirty:    make(map[string]bool),
		queryComponents:    make(map[string][]string), // Selective query invalidation index
		entityListDirty:    true,
		Clock:              NewSimulationClock(0), // Default to deterministic simulation clock
		logger:             logEntry,
		performanceMetrics: NewPerformanceMetrics(), // Initialize performance metrics
		systemNameCache:    make(map[System]string), // Initialize system name cache for reflection avoidance
		systemBudget:       2 * time.Millisecond,   // Default 2 ms per-system frame budget
		budgetWarnCount:    make(map[string]int),
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
	w.entityMu.Lock()
	id := w.nextEntityID
	w.nextEntityID++
	entity := NewEntity(id)
	w.entitiesToAdd = append(w.entitiesToAdd, entity)
	w.entityMu.Unlock()

	if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
		w.logger.WithField("entityID", id).Debug("entity created")
	}

	return entity
}

// AddEntity adds an existing entity to the world.
// Selective query cache invalidation happens in processPendingEntityAdditions
// when the entity is actually moved from the pending queue to the active world.
func (w *World) AddEntity(entity *Entity) {
	w.entityMu.Lock()
	w.entitiesToAdd = append(w.entitiesToAdd, entity)
	w.entityMu.Unlock()
	w.entityListDirty = true
	// Selective invalidation deferred to processPendingEntityAdditions.
}

// RemoveEntity marks an entity for removal from the world.
// Selectively invalidates only queries that could contain the entity,
// based on the entity's current component types.
func (w *World) RemoveEntity(entityID uint64) {
	// Look up the entity's component types before queuing removal so we can
	// selectively invalidate only the relevant cached queries. Copy the keys
	// while holding w.mu.RLock() to avoid racing on the world/entity maps when
	// RemoveEntity is called from another goroutine.
	var entityComponents map[string]Component
	w.mu.RLock()
	entity, exists := w.entities[entityID]
	if exists && len(entity.Components) > 0 {
		entityComponents = make(map[string]Component, len(entity.Components))
		for componentType := range entity.Components {
			// Only the component type keys are required for selective cache
			// invalidation; store nil values in the copied map to avoid retaining
			// references to live component instances after unlocking.
			entityComponents[componentType] = nil
		}
	}
	w.mu.RUnlock()

	w.entityMu.Lock()
	w.entityIDsToRemove = append(w.entityIDsToRemove, entityID)
	w.entityMu.Unlock()
	w.entityListDirty = true
	w.invalidateQueryCacheForComponents(entityComponents)

	if w.logger != nil && w.logger.Logger.GetLevel() >= logrus.DebugLevel {
		w.logger.WithField("entityID", entityID).Debug("entity marked for removal")
	}
}

// GetEntity retrieves an entity by ID.
func (w *World) GetEntity(entityID uint64) (*Entity, bool) {
	entity, ok := w.entities[entityID]
	if ok {
		return entity, true
	}
	// Also check pending additions
	for _, e := range w.entitiesToAdd {
		if e.ID == entityID {
			return e, true
		}
	}
	return nil, false
}

// AddSystem adds a system to the world.
// It is safe to call concurrently from multiple goroutines.
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

	w.mu.Lock()
	defer w.mu.Unlock()

	w.systems = append(w.systems, system)

	// Cache system name to avoid per-frame reflection in Update()
	systemName := w.getSystemName(system)
	w.systemNameCache[system] = systemName

	if w.logger != nil {
		w.logger.WithField("system", systemName).Debug("system added")
	}
}

// Update updates all systems with the current entity list.
// When performance instrumentation is enabled, per-system timing is recorded.
func (w *World) Update(deltaTime float64) {
	// Advance game clock for deterministic time tracking
	w.Clock.Advance(deltaTime)

	// Snapshot pending additions and removals under the entity mutex so that
	// concurrent calls to CreateEntity/AddEntity/RemoveEntity from other
	// goroutines do not race with the iteration below.
	w.entityMu.Lock()
	pendingToAdd := w.entitiesToAdd
	pendingToRemove := w.entityIDsToRemove
	// Setting the fields to nil detaches the staging slices.  Any subsequent
	// append(nil, elem) call by another goroutine always allocates a fresh
	// backing array, so there is no risk of the old slice being modified while
	// we iterate over pendingToAdd / pendingToRemove outside the lock.
	w.entitiesToAdd = nil
	w.entityIDsToRemove = nil
	w.entityMu.Unlock()

	// Process pending additions
	if len(pendingToAdd) > 0 {
		for _, entity := range pendingToAdd {
			w.entities[entity.ID] = entity
			// Selectively invalidate only queries whose required component types
			// overlap with this entity's components.
			w.invalidateQueryCacheForComponents(entity.Components)
		}
		w.entityListDirty = true
	}

	// Process pending removals
	if len(pendingToRemove) > 0 {
		for _, id := range pendingToRemove {
			delete(w.entities, id)
			// Notify registered hooks so systems can reclaim per-entity state.
			for _, hook := range w.entityRemovalHooks {
				hook(id)
			}
		}
		w.entityListDirty = true
	}

	// Rebuild cached entity list if needed
	if w.entityListDirty {
		w.rebuildEntityCache()
	}

	// Update all systems with cached list and timing instrumentation.
	// Uses time.Now() only once per system (unavoidable for per-system metrics).
	// RecordSystemTime is lock-free since World.Update() is single-threaded.
	for _, system := range w.systems {
		startTime := time.Now()
		system.Update(w.cachedEntityList, deltaTime)

		elapsed := time.Since(startTime)
		// Use cached system name (eliminates per-frame reflection)
		systemName := w.systemNameCache[system]
		w.performanceMetrics.RecordSystemTime(systemName, elapsed)

		// Per-system frame-budget enforcement: emit a rate-limited warning when a
		// system exceeds its configured budget (default 2 ms). Warnings are emitted
		// on the 1st, 10th, 100th, and every 1000th occurrence to avoid log flooding.
		if w.logger != nil && w.systemBudget > 0 && elapsed > w.systemBudget {
			w.budgetWarnCount[systemName]++
			count := w.budgetWarnCount[systemName]
			if count == budgetWarnFirst || count == budgetWarnSecond ||
				count == budgetWarnThird || count%budgetWarnPeriod == 0 {
				w.logger.WithFields(logrus.Fields{
					"system":       systemName,
					"elapsed_ms":   elapsed.Milliseconds(),
					"budget_ms":    w.systemBudget.Milliseconds(),
					"excess_count": count,
				}).Warn("system exceeded per-frame budget")
			}
		}
	}
}

// SetSystemBudget configures the per-system frame-time budget for World.Update.
// Systems that exceed the budget emit a rate-limited logrus.Warn.
// Set to 0 to disable budget enforcement. Default: 2 ms.
func (w *World) SetSystemBudget(budget time.Duration) {
	w.systemBudget = budget
}

// AddEntityRemovalHook registers a function that is called on the game-loop
// goroutine for each entity ID that is drained during World.Update.
// Use this to let systems reclaim per-entity state (e.g. visited-cell maps)
// when an entity is removed. Hooks must not call RemoveEntity.
func (w *World) AddEntityRemovalHook(hook func(entityID uint64)) {
	w.entityRemovalHooks = append(w.entityRemovalHooks, hook)
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
	// Ensure pending entity additions are processed before returning
	w.processPendingEntityAdditions()
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
// Uses selective cache invalidation so only queries relevant to the
// added entities' component types are marked dirty.
func (w *World) processPendingEntityAdditions() {
	w.entityMu.Lock()
	if len(w.entitiesToAdd) == 0 {
		w.entityMu.Unlock()
		return
	}
	pending := w.entitiesToAdd
	w.entitiesToAdd = nil // detach; concurrent appends will use a new backing array
	w.entityMu.Unlock()

	for _, entity := range pending {
		w.entities[entity.ID] = entity
		// Selectively invalidate only queries whose required component types
		// overlap with this entity's components (entity.Components is the component map).
		w.invalidateQueryCacheForComponents(entity.Components)
	}
	w.entityListDirty = true
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
// Records the query's required component types on first call for selective cache invalidation.
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

	// Record the component types for this query key so selective invalidation
	// can avoid dirtying unrelated queries on entity add/remove.
	if _, known := w.queryComponents[key]; !known {
		types := make([]string, len(componentTypes))
		copy(types, componentTypes)
		w.queryComponents[key] = types
	}

	return result
}

// invalidateQueryCacheForComponents marks only cached queries dirty if their
// required component set intersects with the provided component types.
// This reduces cascading O(q) invalidation to O(q × overlap) in the common case.
// Falls back to full invalidation when entityComponents is nil (unknown components).
func (w *World) invalidateQueryCacheForComponents(entityComponents map[string]Component) {
	if entityComponents == nil {
		// Unknown components — invalidate all queries conservatively.
		for key := range w.queryCache {
			w.queryCacheDirty[key] = true
		}
		return
	}

	for key, required := range w.queryComponents {
		for _, compType := range required {
			if _, has := entityComponents[compType]; has {
				w.queryCacheDirty[key] = true
				break
			}
		}
	}
	// Invalidate queries that have not been registered in queryComponents yet
	// (they were built from the pre-computed key fast path before queryComponents was populated).
	for key := range w.queryCache {
		if _, known := w.queryComponents[key]; !known {
			w.queryCacheDirty[key] = true
		}
	}
}

// invalidateQueryCache marks all cached queries as dirty.
// Called when entities are added or removed from the world without known component context.
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

// SetModRules sets the mod rule provider for the world.
// Phase 6.3 (PLAN.md): Modding System Integration
func (w *World) SetModRules(provider ModRuleProvider) {
	w.ModRules = provider
}

// GetModRules returns the mod rule provider, or nil if not set.
// Phase 6.3 (PLAN.md): Modding System Integration
func (w *World) GetModRules() ModRuleProvider {
	return w.ModRules
}

// GetModRuleFloat64 is a convenience method to get a mod rule as float64.
// Returns the default value if no mod provider is set or the rule doesn't exist.
// Phase 6.3 (PLAN.md): Modding System Integration
func (w *World) GetModRuleFloat64(ruleName string, defaultValue float64) float64 {
	if w.ModRules == nil {
		return defaultValue
	}
	return w.ModRules.GetRuleFloat64(ruleName, defaultValue)
}

// GetModRuleBool is a convenience method to get a mod rule as bool.
// Returns the default value if no mod provider is set or the rule doesn't exist.
// Phase 6.3 (PLAN.md): Modding System Integration
func (w *World) GetModRuleBool(ruleName string, defaultValue bool) bool {
	if w.ModRules == nil {
		return defaultValue
	}
	return w.ModRules.GetRuleBool(ruleName, defaultValue)
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

// GetPerformanceMetrics returns the world's performance metrics snapshot.
// This provides access to per-system timing data for performance monitoring.
// Thread-safe for concurrent access from metrics HTTP handler.
func (w *World) GetPerformanceMetrics() *PerformanceMetrics {
	return w.performanceMetrics.GetSnapshot()
}

// getSystemName extracts a readable name from a system instance.
// Uses reflection to get the type name without pointer prefix.
func (w *World) getSystemName(system System) string {
	// Use fmt.Sprintf with %T to get the type name
	typeName := fmt.Sprintf("%T", system)

	// Remove pointer prefix if present
	if len(typeName) > 0 && typeName[0] == '*' {
		typeName = typeName[1:]
	}

	// Extract just the struct name without package path
	// e.g., "engine.MovementSystem" -> "MovementSystem"
	lastDot := -1
	for i := len(typeName) - 1; i >= 0; i-- {
		if typeName[i] == '.' {
			lastDot = i
			break
		}
	}
	if lastDot >= 0 && lastDot < len(typeName)-1 {
		typeName = typeName[lastDot+1:]
	}

	return typeName
}
