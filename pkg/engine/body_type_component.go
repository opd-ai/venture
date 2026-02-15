// Package engine provides the BodyTypeComponent which stores per-entity
// body type state. This is a transient visual component—not persisted in saves.
package engine

// BodyTypeComponent holds the derived body type for an entity, controlling
// how anatomical template proportions are modified during sprite generation.
type BodyTypeComponent struct {
	// BodyType is the index of the assigned body type (0 = Average).
	BodyType int

	// GenreID is the genre used when deriving the body type.
	GenreID string

	// Assigned is true once the system has populated this component.
	Assigned bool

	// Dirty flags the sprite for regeneration when body type changes.
	Dirty bool
}

// Type returns the component type identifier.
func (b *BodyTypeComponent) Type() string {
	return "body_type"
}

// NewBodyTypeComponent creates a component with default (unassigned) values.
func NewBodyTypeComponent() *BodyTypeComponent {
	return &BodyTypeComponent{
		BodyType: 0,
		Assigned: false,
		Dirty:    false,
	}
}
