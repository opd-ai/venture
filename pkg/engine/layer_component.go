// Package engine provides the Layer component for multi-layer terrain support.
// This file implements Phase 11.1 multi-layer terrain functionality.
package engine

// LayerComponent tracks which terrain layer an entity is on.
// Phase 11.1: Enables multi-layer environments with platforms, pits, and layer transitions.
//
// Layer values:
//   - 0: Ground layer (default, floor, corridors)
//   - 1: Water/Pit layer (deep water, pits, chasms)
//   - 2: Platform layer (elevated platforms, bridges)
//
// Entities on different layers don't collide with each other,
// enabling strategic gameplay (jumping down from platforms, crossing bridges over pits).
type LayerComponent struct {
	// CurrentLayer is the layer the entity is currently on (0, 1, or 2)
	CurrentLayer int

	// TargetLayer is the layer the entity is transitioning to (via ramp/stairs)
	// Set to -1 when not transitioning
	TargetLayer int

	// TransitionProgress tracks transition completion (0.0 to 1.0)
	// Only valid when TargetLayer != -1
	TransitionProgress float64

	// CanFly indicates if entity can move between layers freely (e.g., flying enemies)
	CanFly bool

	// CanSwim indicates if entity can enter water layer (0 -> 1 transition)
	CanSwim bool

	// CanClimb indicates if entity can use ramps/stairs (layer transitions)
	CanClimb bool
}

// Type returns the component type identifier.
func (l LayerComponent) Type() string {
	return "layer"
}

// NewLayerComponent creates a layer component with default values.
// Entities start on ground layer (0) with standard movement capabilities.
func NewLayerComponent() LayerComponent {
	return LayerComponent{
		CurrentLayer:       0,  // Ground layer
		TargetLayer:        -1, // Not transitioning
		TransitionProgress: 0.0,
		CanFly:             false,
		CanSwim:            false,
		CanClimb:           true, // Most entities can use ramps/stairs
	}
}

// NewFlyingLayerComponent creates a layer component for flying entities.
// Flying entities can move between any layers freely.
func NewFlyingLayerComponent() LayerComponent {
	return LayerComponent{
		CurrentLayer:       0,
		TargetLayer:        -1,
		TransitionProgress: 0.0,
		CanFly:             true,
		CanSwim:            true,
		CanClimb:           true,
	}
}

// NewSwimmingLayerComponent creates a layer component for aquatic entities.
// Swimming entities can move between ground (0) and water (1) layers.
func NewSwimmingLayerComponent() LayerComponent {
	return LayerComponent{
		CurrentLayer:       1, // Start in water layer
		TargetLayer:        -1,
		TransitionProgress: 0.0,
		CanFly:             false,
		CanSwim:            true,
		CanClimb:           false,
	}
}

// IsTransitioning returns true if the entity is currently moving between layers.
func (l *LayerComponent) IsTransitioning() bool {
	return l.TargetLayer != -1
}

// StartTransition begins a transition to a new layer.
// Should be called when entity enters a ramp or stairs tile.
func (l *LayerComponent) StartTransition(targetLayer int) {
	l.TargetLayer = targetLayer
	l.TransitionProgress = 0.0
}

// UpdateTransition advances the layer transition by the given amount.
// Progress should be based on distance traveled along ramp/stairs.
// Returns true if transition is complete.
func (l *LayerComponent) UpdateTransition(progressDelta float64) bool {
	if !l.IsTransitioning() {
		return false
	}

	l.TransitionProgress += progressDelta
	if l.TransitionProgress >= 1.0 {
		// Complete transition
		l.CurrentLayer = l.TargetLayer
		l.TargetLayer = -1
		l.TransitionProgress = 0.0
		return true
	}
	return false
}

// CancelTransition cancels an in-progress layer transition.
// Returns entity to original layer.
func (l *LayerComponent) CancelTransition() {
	l.TargetLayer = -1
	l.TransitionProgress = 0.0
}

// CanTransitionTo checks if entity can move to the specified layer.
// Considers entity capabilities (flying, swimming, climbing) and layer rules.
func (l *LayerComponent) CanTransitionTo(targetLayer int) bool {
	// Flying entities can go anywhere
	if l.CanFly {
		return true
	}

	// Check layer-specific rules
	switch targetLayer {
	case 0: // Ground layer
		// Can always return to ground from water (if can swim) or platform
		return l.CanSwim || l.CurrentLayer == 2
	case 1: // Water/Pit layer
		// Only if can swim
		return l.CanSwim
	case 2: // Platform layer
		// Need to climb (use ramp/stairs)
		return l.CanClimb
	default:
		return false
	}
}

// GetEffectiveLayer returns the layer to use for collision detection.
// During transitions, returns the layer the entity is moving toward.
func (l *LayerComponent) GetEffectiveLayer() int {
	if l.IsTransitioning() && l.TransitionProgress > 0.5 {
		// More than halfway through transition, use target layer
		return l.TargetLayer
	}
	return l.CurrentLayer
}

// OnSameLayer checks if two entities are on the same effective layer.
func OnSameLayer(l1, l2 *LayerComponent) bool {
	if l1 == nil || l2 == nil {
		// If either entity lacks layer component, assume ground layer (0)
		layer1 := 0
		layer2 := 0
		if l1 != nil {
			layer1 = l1.GetEffectiveLayer()
		}
		if l2 != nil {
			layer2 = l2.GetEffectiveLayer()
		}
		return layer1 == layer2
	}
	return l1.GetEffectiveLayer() == l2.GetEffectiveLayer()
}

// Clone creates a deep copy of the layer component.
func (l *LayerComponent) Clone() LayerComponent {
	return LayerComponent{
		CurrentLayer:       l.CurrentLayer,
		TargetLayer:        l.TargetLayer,
		TransitionProgress: l.TransitionProgress,
		CanFly:             l.CanFly,
		CanSwim:            l.CanSwim,
		CanClimb:           l.CanClimb,
	}
}
