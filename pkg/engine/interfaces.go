package engine

// Component represents a data container attached to an Entity.
// Components should be pure data structures without behavior.
// Originally from: ecs.go
type Component interface {
	// Type returns a unique identifier for this component type
	Type() string
}

// System represents a behavior that operates on entities with specific components.
// Systems should be stateless where possible and operate on entity data.
// Originally from: ecs.go
type System interface {
	// Update is called every frame to update entities managed by this system
	Update(entities []*Entity, deltaTime float64)
}
