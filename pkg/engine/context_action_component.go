// Package engine provides the context action component for interactive objects.
// This file implements ContextActionComponent which defines context-sensitive
// interaction types for objects in the world (doors, levers, NPCs, etc.).
//
// Phase 11.3: Context Actions System
// Provides 8 action types: Open, Close, Push, Pull, Activate, Talk, Pickup, Read
package engine

import (
	"fmt"
)

// ContextActionType defines the type of context action available on an entity.
type ContextActionType int

const (
	// ActionNone represents no action available
	ActionNone ContextActionType = iota
	// ActionOpen opens a door, chest, or container
	ActionOpen
	// ActionClose closes a door or container
	ActionClose
	// ActionPush pushes a movable object
	ActionPush
	// ActionPull pulls a movable object
	ActionPull
	// ActionActivate activates a lever, switch, or mechanism
	ActionActivate
	// ActionTalk initiates dialog with an NPC
	ActionTalk
	// ActionPickup picks up an item or object
	ActionPickup
	// ActionRead reads a sign, book, or inscription
	ActionRead
)

// String returns the human-readable name of the action type.
func (a ContextActionType) String() string {
	switch a {
	case ActionNone:
		return "None"
	case ActionOpen:
		return "Open"
	case ActionClose:
		return "Close"
	case ActionPush:
		return "Push"
	case ActionPull:
		return "Pull"
	case ActionActivate:
		return "Activate"
	case ActionTalk:
		return "Talk"
	case ActionPickup:
		return "Pickup"
	case ActionRead:
		return "Read"
	default:
		return "Unknown"
	}
}

// ContextActionComponent defines context-sensitive actions available on an entity.
// When the player is near this entity and presses the interaction key (F),
// the specified action is performed.
type ContextActionComponent struct {
	// ActionType is the type of action available
	ActionType ContextActionType

	// Prompt is the text displayed to the player (e.g., "Press F to Open Door")
	Prompt string

	// InteractionRange is the maximum distance for interaction (in pixels)
	InteractionRange float64

	// Enabled determines if the action is currently available
	Enabled bool

	// OnActivate is called when the action is triggered (optional callback)
	OnActivate func(entity *Entity, world *World) error

	// CustomData holds action-specific data (e.g., dialog text, target position)
	CustomData map[string]interface{}
}

// NewContextActionComponent creates a new context action component with the specified action type.
func NewContextActionComponent(actionType ContextActionType) *ContextActionComponent {
	return &ContextActionComponent{
		ActionType:       actionType,
		Prompt:           fmt.Sprintf("Press F to %s", actionType.String()),
		InteractionRange: 64.0, // Default 2 tiles
		Enabled:          true,
		CustomData:       make(map[string]interface{}),
	}
}

// Type returns the component type identifier.
func (c *ContextActionComponent) Type() string {
	return "contextaction"
}

// SetPrompt sets a custom prompt message for this action.
func (c *ContextActionComponent) SetPrompt(prompt string) {
	c.Prompt = prompt
}

// SetRange sets the interaction range in pixels.
func (c *ContextActionComponent) SetRange(distance float64) {
	c.InteractionRange = distance
}

// SetEnabled enables or disables this action.
func (c *ContextActionComponent) SetEnabled(enabled bool) {
	c.Enabled = enabled
}

// SetCallback sets the callback function invoked when action is triggered.
func (c *ContextActionComponent) SetCallback(callback func(entity *Entity, world *World) error) {
	c.OnActivate = callback
}

// SetData sets custom data for this action.
func (c *ContextActionComponent) SetData(key string, value interface{}) {
	c.CustomData[key] = value
}

// GetData retrieves custom data for this action.
func (c *ContextActionComponent) GetData(key string) (interface{}, bool) {
	val, ok := c.CustomData[key]
	return val, ok
}

// Activate triggers the context action if enabled.
// Returns error if action fails or is disabled.
func (c *ContextActionComponent) Activate(entity *Entity, world *World) error {
	if !c.Enabled {
		return fmt.Errorf("action %s is disabled", c.ActionType.String())
	}

	// Call the custom callback if set
	if c.OnActivate != nil {
		return c.OnActivate(entity, world)
	}

	return nil
}
