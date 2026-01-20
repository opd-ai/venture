// Package economy interfaces.
// This file defines minimal interfaces for ECS integration (World, Entity),
// enabling loose coupling between economy system and game engine.
// Originally extracted from system.go for better discoverability.
package economy

// World is the minimal interface for ECS world operations needed by the economy system.
// Originally from: system.go
type World interface {
	GetEntities() []Entity
}

// Entity is the minimal interface for ECS entities needed by the economy system.
// Originally from: system.go
type Entity interface {
	HasComponent(componentType string) bool
	GetComponent(componentType string) (interface{}, bool)
}
