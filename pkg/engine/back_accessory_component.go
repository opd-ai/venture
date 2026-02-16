// Package engine provides the BackAccessoryComponent, a pure data component
// that stores which back-worn accessory type an entity has (cape, cloak,
// quiver, backpack, banner, scarf, wing-cape). The BackAccessorySystem
// populates this component based on entity role, genre, and seed.
package engine

// BackAccessoryComponent stores the back accessory visual type for an entity.
// It is pure data with no logic, following ECS conventions.
type BackAccessoryComponent struct {
	// AccessoryType is an integer matching sprites.BackAccessoryType constants.
	AccessoryType int
	// Genre stores the genre used when selecting this accessory.
	Genre string
	// Role stores the entity role used for accessory selection.
	Role string
}

// Type returns the component type identifier.
func (b *BackAccessoryComponent) Type() string {
	return "back_accessory"
}
