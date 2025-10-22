package engine

// Entity represents a game object composed of components.
// Entities are identified by a unique ID and contain a collection of components.
type Entity struct {
	ID         uint64
	Components map[string]Component
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
}

// GetComponent retrieves a component by type.
func (e *Entity) GetComponent(componentType string) (Component, bool) {
	c, ok := e.Components[componentType]
	return c, ok
}

// RemoveComponent removes a component from this entity.
func (e *Entity) RemoveComponent(componentType string) {
	delete(e.Components, componentType)
}

// HasComponent checks if this entity has a component of the given type.
func (e *Entity) HasComponent(componentType string) bool {
	_, ok := e.Components[componentType]
	return ok
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
}

// NewWorld creates a new game world.
func NewWorld() *World {
	return &World{
		entities:         make(map[uint64]*Entity),
		systems:          make([]System, 0),
		cachedEntityList: make([]*Entity, 0, 256), // Pre-allocate for 256 entities
		entityListDirty:  true,
	}
}

// CreateEntity creates a new entity and adds it to the world.
func (w *World) CreateEntity() *Entity {
	id := w.nextEntityID
	w.nextEntityID++
	entity := NewEntity(id)
	w.entitiesToAdd = append(w.entitiesToAdd, entity)
	return entity
}

// AddEntity adds an existing entity to the world.
func (w *World) AddEntity(entity *Entity) {
	w.entitiesToAdd = append(w.entitiesToAdd, entity)
	w.entityListDirty = true
}

// RemoveEntity marks an entity for removal from the world.
func (w *World) RemoveEntity(entityID uint64) {
	w.entityIDsToRemove = append(w.entityIDsToRemove, entityID)
	w.entityListDirty = true
}

// GetEntity retrieves an entity by ID.
func (w *World) GetEntity(entityID uint64) (*Entity, bool) {
	entity, ok := w.entities[entityID]
	return entity, ok
}

// AddSystem adds a system to the world.
func (w *World) AddSystem(system System) {
	w.systems = append(w.systems, system)
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
func (w *World) GetEntitiesWith(componentTypes ...string) []*Entity {
	result := make([]*Entity, 0)
	for _, entity := range w.entities {
		hasAll := true
		for _, compType := range componentTypes {
			if !entity.HasComponent(compType) {
				hasAll = false
				break
			}
		}
		if hasAll {
			result = append(result, entity)
		}
	}
	return result
}
