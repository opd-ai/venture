// Package engine provides tests for the context action component.
package engine

import (
	"testing"
)

func TestNewContextActionComponent(t *testing.T) {
	tests := []struct {
		name       string
		actionType ContextActionType
		wantPrompt string
		wantRange  float64
	}{
		{
			name:       "open action",
			actionType: ActionOpen,
			wantPrompt: "Press F to Open",
			wantRange:  64.0,
		},
		{
			name:       "talk action",
			actionType: ActionTalk,
			wantPrompt: "Press F to Talk",
			wantRange:  64.0,
		},
		{
			name:       "activate action",
			actionType: ActionActivate,
			wantPrompt: "Press F to Activate",
			wantRange:  64.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewContextActionComponent(tt.actionType)

			if comp.ActionType != tt.actionType {
				t.Errorf("ActionType = %v, want %v", comp.ActionType, tt.actionType)
			}

			if comp.Prompt != tt.wantPrompt {
				t.Errorf("Prompt = %q, want %q", comp.Prompt, tt.wantPrompt)
			}

			if comp.InteractionRange != tt.wantRange {
				t.Errorf("InteractionRange = %v, want %v", comp.InteractionRange, tt.wantRange)
			}

			if !comp.Enabled {
				t.Error("Expected component to be enabled by default")
			}

			if comp.CustomData == nil {
				t.Error("Expected CustomData map to be initialized")
			}
		})
	}
}

func TestContextActionComponent_Type(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen)
	if got := comp.Type(); got != "contextaction" {
		t.Errorf("Type() = %q, want %q", got, "contextaction")
	}
}

func TestContextActionComponent_SetPrompt(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen)
	customPrompt := "Press E to unlock treasure chest"
	comp.SetPrompt(customPrompt)

	if comp.Prompt != customPrompt {
		t.Errorf("Prompt = %q, want %q", comp.Prompt, customPrompt)
	}
}

func TestContextActionComponent_SetRange(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen)
	newRange := 128.0
	comp.SetRange(newRange)

	if comp.InteractionRange != newRange {
		t.Errorf("InteractionRange = %v, want %v", comp.InteractionRange, newRange)
	}
}

func TestContextActionComponent_SetEnabled(t *testing.T) {
	comp := NewContextActionComponent(ActionOpen)

	comp.SetEnabled(false)
	if comp.Enabled {
		t.Error("Expected component to be disabled")
	}

	comp.SetEnabled(true)
	if !comp.Enabled {
		t.Error("Expected component to be enabled")
	}
}

func TestContextActionComponent_CustomData(t *testing.T) {
	comp := NewContextActionComponent(ActionRead)

	// Set data
	comp.SetData("text", "Ancient inscription on the wall")
	comp.SetData("readCount", 0)

	// Get data
	if text, ok := comp.GetData("text"); !ok {
		t.Error("Expected 'text' key to exist")
	} else if text.(string) != "Ancient inscription on the wall" {
		t.Errorf("text = %v, want %q", text, "Ancient inscription on the wall")
	}

	if count, ok := comp.GetData("readCount"); !ok {
		t.Error("Expected 'readCount' key to exist")
	} else if count.(int) != 0 {
		t.Errorf("readCount = %v, want 0", count)
	}

	// Non-existent key
	if _, ok := comp.GetData("nonexistent"); ok {
		t.Error("Expected 'nonexistent' key to not exist")
	}
}

func TestContextActionComponent_Activate(t *testing.T) {
	world := NewWorld()

	tests := []struct {
		name       string
		enabled    bool
		hasCallback bool
		wantError   bool
	}{
		{
			name:       "enabled with callback",
			enabled:    true,
			hasCallback: true,
			wantError:   false,
		},
		{
			name:       "enabled without callback",
			enabled:    true,
			hasCallback: false,
			wantError:   false,
		},
		{
			name:       "disabled with callback",
			enabled:    false,
			hasCallback: true,
			wantError:   true,
		},
		{
			name:       "disabled without callback",
			enabled:    false,
			hasCallback: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			comp := NewContextActionComponent(ActionOpen)
			comp.SetEnabled(tt.enabled)

			callbackInvoked := false
			if tt.hasCallback {
				comp.SetCallback(func(entity *Entity, world *World) error {
					callbackInvoked = true
					return nil
				})
			}

			entity := world.CreateEntity()
			err := comp.Activate(entity, world)

			if (err != nil) != tt.wantError {
				t.Errorf("Activate() error = %v, wantError %v", err, tt.wantError)
			}

			if tt.enabled && tt.hasCallback && !callbackInvoked {
				t.Error("Expected callback to be invoked")
			}
		})
	}
}

func TestContextActionType_String(t *testing.T) {
	tests := []struct {
		actionType ContextActionType
		want       string
	}{
		{ActionNone, "None"},
		{ActionOpen, "Open"},
		{ActionClose, "Close"},
		{ActionPush, "Push"},
		{ActionPull, "Pull"},
		{ActionActivate, "Activate"},
		{ActionTalk, "Talk"},
		{ActionPickup, "Pickup"},
		{ActionRead, "Read"},
		{ContextActionType(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.actionType.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestContextActionComponent_Integration(t *testing.T) {
	// Test a complete interaction scenario
	world := NewWorld()
	entity := world.CreateEntity()

	// Create a door with open action
	doorAction := NewContextActionComponent(ActionOpen)
	doorAction.SetPrompt("Press F to open the wooden door")
	doorAction.SetRange(64.0)

	doorOpened := false
	doorAction.SetCallback(func(e *Entity, w *World) error {
		doorOpened = true
		// Change to close action after opening
		doorAction.ActionType = ActionClose
		doorAction.SetPrompt("Press F to close the door")
		return nil
	})

	entity.AddComponent(doorAction)

	// Verify initial state
	if doorOpened {
		t.Error("Door should not be opened initially")
	}

	// Activate the action
	if err := doorAction.Activate(entity, world); err != nil {
		t.Errorf("Activate() unexpected error: %v", err)
	}

	// Verify door was opened
	if !doorOpened {
		t.Error("Door should be opened after activation")
	}

	// Verify action changed to close
	if doorAction.ActionType != ActionClose {
		t.Errorf("ActionType = %v, want %v", doorAction.ActionType, ActionClose)
	}
}
