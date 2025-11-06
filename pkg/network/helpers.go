package network

// NewStateUpdate creates a state update with the specified priority.
func NewStateUpdate(entityID uint64, priority uint8) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: priority,
	}
}

// NewCriticalUpdate creates a critical priority state update.
// Use for death events, revival events, and other game-critical state changes.
func NewCriticalUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityCritical,
	}
}

// NewHighPriorityUpdate creates a high priority state update.
// Use for combat events, damage notifications, and important gameplay changes.
func NewHighPriorityUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityHigh,
	}
}

// NewNormalUpdate creates a normal priority state update.
// Use for regular entity position/velocity updates and general state changes.
func NewNormalUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityNormal,
	}
}

// NewLowPriorityUpdate creates a low priority state update.
// Use for cosmetic changes like animations, particles, and visual effects.
func NewLowPriorityUpdate(entityID uint64) *StateUpdate {
	return &StateUpdate{
		EntityID: entityID,
		Priority: PriorityLow,
	}
}
