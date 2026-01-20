package story

// Type Definitions
// This file consolidates all type definitions for the story package

// FragmentType represents different types of story fragments
// Originally from: generator.go
type FragmentType int

// ArtifactType categorizes archaeological finds
// Originally from: archaeology.go
type ArtifactType int

// EventType categorizes historical events
// Originally from: timeline.go
type EventType int

// Vector2 represents a 2D position
// Originally from: generator.go
// Shared by: archaeology.go, generator.go
type Vector2 struct {
	X float64
	Y float64
}
