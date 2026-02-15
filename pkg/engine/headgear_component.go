// Package engine provides the HeadgearComponent, a pure data component
// that stores which headgear type an entity wears. The HeadgearAssignmentSystem
// populates this component based on entity role, genre, and seed.
package engine

// HeadgearComponent stores the headgear visual type for an entity.
// It is pure data with no logic, following ECS conventions.
type HeadgearComponent struct {
	// HeadgearType is an integer matching sprites.HeadgearType constants.
	HeadgearType int
	// Genre stores the genre used when selecting this headgear.
	Genre string
	// Role stores the entity role used for headgear selection.
	Role string
}

// Type returns the component type identifier.
func (h *HeadgearComponent) Type() string {
	return "headgear"
}
