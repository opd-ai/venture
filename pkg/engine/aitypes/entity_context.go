// Package aitypes defines minimal shared types for the behavior tree AI subsystem.
package aitypes

// EntityContext is the parameter type for behavior tree Tick methods.
// It is kept as an empty interface so that concrete engine entities (*engine.Entity)
// satisfy it without requiring aitypes to import pkg/engine types (which would
// create a circular dependency). Leaf nodes that need to inspect the entity
// perform a type assertion to *engine.Entity inside their Tick implementation.
// Composite nodes (Sequence, Selector, Parallel …) simply pass the value through
// without inspecting it, requiring no assertion at all.
//
// As the extraction matures, this interface may grow to expose methods that are
// safe to add to aitypes (e.g. EntityID() uint64) without creating cycles.
type EntityContext interface{}
