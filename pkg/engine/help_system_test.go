// Package engine provides unit tests for the help system.
// H-001 FIX: Comprehensive test coverage for HelpSystem (489 lines, was 0% coverage)
package engine

import (
	"testing"

	"github.com/opd-ai/venture/pkg/procgen/item"
)

// TestNewHelpSystem validates help system initialization.
func TestNewHelpSystem(t *testing.T) {
	hs := NewHelpSystem()

	if hs == nil {
		t.Fatal("NewHelpSystem returned nil")
	}

	// Check default state
	if !hs.Enabled {
		t.Error("Expected help system to be enabled by default")
	}
	if hs.Visible {
		t.Error("Expected help system to be hidden initially")
	}
	if hs.CurrentTopic != "" {
		t.Errorf("Expected empty current topic, got %q", hs.CurrentTopic)
	}

	// Check default topics exist
	expectedTopics := []string{"controls", "combat", "inventory", "progression", "world", "multiplayer"}
	for _, topicID := range expectedTopics {
		if _, exists := hs.Topics[topicID]; !exists {
			t.Errorf("Expected default topic %q not found", topicID)
		}
	}

	// Check quick hints exist
	expectedHints := []string{"low_health", "level_up", "inventory_full", "no_mana", 
		"enemy_nearby", "item_dropped", "boss_ahead", "quest_complete", "first_death"}
	for _, hintID := range expectedHints {
		if _, exists := hs.QuickHints[hintID]; !exists {
			t.Errorf("Expected quick hint %q not found", hintID)
		}
	}
}

// TestShowTopic validates showing specific help topics.
func TestShowTopic(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		topicID     string
		wantVisible bool
		wantTopic   string
	}{
		{
			name:        "show existing topic when enabled",
			enabled:     true,
			topicID:     "controls",
			wantVisible: true,
			wantTopic:   "controls",
		},
		{
			name:        "show different topic",
			enabled:     true,
			topicID:     "combat",
			wantVisible: true,
			wantTopic:   "combat",
		},
		{
			name:        "ignore when disabled",
			enabled:     false,
			topicID:     "controls",
			wantVisible: false,
			wantTopic:   "",
		},
		{
			name:        "ignore non-existent topic",
			enabled:     true,
			topicID:     "nonexistent",
			wantVisible: false,
			wantTopic:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := NewHelpSystem()
			hs.Enabled = tt.enabled

			hs.ShowTopic(tt.topicID)

			if hs.Visible != tt.wantVisible {
				t.Errorf("ShowTopic() Visible = %v, want %v", hs.Visible, tt.wantVisible)
			}
			if hs.CurrentTopic != tt.wantTopic {
				t.Errorf("ShowTopic() CurrentTopic = %q, want %q", hs.CurrentTopic, tt.wantTopic)
			}
		})
	}
}

// TestHide validates hiding the help display.
func TestHide(t *testing.T) {
	hs := NewHelpSystem()
	hs.Visible = true
	hs.CurrentTopic = "controls"

	hs.Hide()

	if hs.Visible {
		t.Error("Hide() did not set Visible to false")
	}
	// Note: Hide() doesn't clear CurrentTopic, which is correct for resuming
	if hs.CurrentTopic != "controls" {
		t.Errorf("Hide() unexpectedly changed CurrentTopic to %q", hs.CurrentTopic)
	}
}

// TestToggle validates toggling help visibility.
func TestToggle(t *testing.T) {
	tests := []struct {
		name            string
		initialVisible  bool
		initialTopic    string
		wantVisible     bool
		wantDefaultSet  bool
	}{
		{
			name:           "toggle on from hidden state",
			initialVisible: false,
			initialTopic:   "",
			wantVisible:    true,
			wantDefaultSet: true, // Should set default topic
		},
		{
			name:           "toggle off from visible state",
			initialVisible: true,
			initialTopic:   "combat",
			wantVisible:    false,
			wantDefaultSet: false,
		},
		{
			name:           "toggle on with existing topic",
			initialVisible: false,
			initialTopic:   "inventory",
			wantVisible:    true,
			wantDefaultSet: false, // Should keep existing topic
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := NewHelpSystem()
			hs.Visible = tt.initialVisible
			hs.CurrentTopic = tt.initialTopic

			hs.Toggle()

			if hs.Visible != tt.wantVisible {
				t.Errorf("Toggle() Visible = %v, want %v", hs.Visible, tt.wantVisible)
			}

			if tt.wantDefaultSet && hs.CurrentTopic != "controls" {
				t.Errorf("Toggle() did not set default topic, got %q", hs.CurrentTopic)
			}
			if !tt.wantDefaultSet && hs.CurrentTopic != tt.initialTopic {
				t.Errorf("Toggle() changed topic from %q to %q", tt.initialTopic, hs.CurrentTopic)
			}
		})
	}
}

// TestShowQuickHintFor validates displaying context-sensitive hints.
func TestShowQuickHintFor(t *testing.T) {
	tests := []struct {
		name        string
		enabled     bool
		context     string
		wantShowing bool
		wantHint    string
	}{
		{
			name:        "show existing hint when enabled",
			enabled:     true,
			context:     "low_health",
			wantShowing: true,
			wantHint:    "Health low! Find healing or retreat to safety",
		},
		{
			name:        "show different hint",
			enabled:     true,
			context:     "inventory_full",
			wantShowing: true,
			wantHint:    "Inventory full! Press I to manage items",
		},
		{
			name:        "ignore when disabled",
			enabled:     false,
			context:     "low_health",
			wantShowing: false,
			wantHint:    "",
		},
		{
			name:        "ignore non-existent hint",
			enabled:     true,
			context:     "nonexistent",
			wantShowing: false,
			wantHint:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := NewHelpSystem()
			hs.Enabled = tt.enabled

			hs.ShowQuickHintFor(tt.context)

			if hs.ShowQuickHint != tt.wantShowing {
				t.Errorf("ShowQuickHintFor() ShowQuickHint = %v, want %v", hs.ShowQuickHint, tt.wantShowing)
			}
			if hs.CurrentHint != tt.wantHint {
				t.Errorf("ShowQuickHintFor() CurrentHint = %q, want %q", hs.CurrentHint, tt.wantHint)
			}
		})
	}
}

// TestHideQuickHint validates hiding quick hints.
func TestHideQuickHint(t *testing.T) {
	hs := NewHelpSystem()
	hs.ShowQuickHint = true
	hs.CurrentHint = "Some hint"

	hs.HideQuickHint()

	if hs.ShowQuickHint {
		t.Error("HideQuickHint() did not set ShowQuickHint to false")
	}
	if hs.CurrentHint != "" {
		t.Errorf("HideQuickHint() did not clear CurrentHint, got %q", hs.CurrentHint)
	}
}

// TestGetTopicList validates retrieving all topic IDs.
func TestGetTopicList(t *testing.T) {
	hs := NewHelpSystem()

	topics := hs.GetTopicList()

	if len(topics) == 0 {
		t.Fatal("GetTopicList() returned empty list")
	}

	expectedTopics := map[string]bool{
		"controls": true, "combat": true, "inventory": true,
		"progression": true, "world": true, "multiplayer": true,
	}

	for _, topicID := range topics {
		if !expectedTopics[topicID] {
			t.Errorf("GetTopicList() returned unexpected topic %q", topicID)
		}
		delete(expectedTopics, topicID)
	}

	if len(expectedTopics) > 0 {
		t.Errorf("GetTopicList() missing topics: %v", expectedTopics)
	}
}

// TestGetTopic validates retrieving specific topics.
func TestGetTopic(t *testing.T) {
	tests := []struct {
		name       string
		topicID    string
		wantExists bool
		wantTitle  string
	}{
		{
			name:       "get existing topic",
			topicID:    "controls",
			wantExists: true,
			wantTitle:  "Game Controls",
		},
		{
			name:       "get combat topic",
			topicID:    "combat",
			wantExists: true,
			wantTitle:  "Combat Guide",
		},
		{
			name:       "get non-existent topic",
			topicID:    "nonexistent",
			wantExists: false,
			wantTitle:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := NewHelpSystem()

			topic, exists := hs.GetTopic(tt.topicID)

			if exists != tt.wantExists {
				t.Errorf("GetTopic() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists {
				if topic == nil {
					t.Fatal("GetTopic() returned nil topic pointer")
				}
				if topic.Title != tt.wantTitle {
					t.Errorf("GetTopic() Title = %q, want %q", topic.Title, tt.wantTitle)
				}
				if topic.ID != tt.topicID {
					t.Errorf("GetTopic() ID = %q, want %q", topic.ID, tt.topicID)
				}
			}
		})
	}
}

// TestIsActive validates UISystem interface implementation.
func TestIsActive(t *testing.T) {
	hs := NewHelpSystem()

	// Initially not active
	if hs.IsActive() {
		t.Error("IsActive() should return false initially")
	}

	// After showing
	hs.Visible = true
	if !hs.IsActive() {
		t.Error("IsActive() should return true when visible")
	}

	// After hiding
	hs.Visible = false
	if hs.IsActive() {
		t.Error("IsActive() should return false after hiding")
	}
}

// TestSetActive validates UISystem interface implementation.
func TestSetActive(t *testing.T) {
	hs := NewHelpSystem()

	hs.SetActive(true)
	if !hs.Visible {
		t.Error("SetActive(true) did not set Visible to true")
	}

	hs.SetActive(false)
	if hs.Visible {
		t.Error("SetActive(false) did not set Visible to false")
	}
}

// TestUpdateDisabled validates Update behavior when disabled.
func TestUpdateDisabled(t *testing.T) {
	hs := NewHelpSystem()
	hs.Enabled = false
	hs.Visible = true // Should stay true even after Update when disabled

	entities := []*Entity{} // Empty entity list
	hs.Update(entities, 0.016)

	// When disabled, Update should not change state
	if !hs.Visible {
		t.Error("Update() changed Visible when system was disabled")
	}
}

// TestUpdateAutoHints validates automatic hint detection.
func TestUpdateAutoHints(t *testing.T) {
	tests := []struct {
		name         string
		setupEntity  func() *Entity
		wantHintSet  bool
		wantContext  string
	}{
		{
			name: "detect low health",
			setupEntity: func() *Entity {
				e := NewEntity(1)
				e.AddComponent(&EbitenInput{})
				e.AddComponent(&HealthComponent{Current: 20, Max: 100})
				return e
			},
			wantHintSet: true,
			wantContext: "low_health",
		},
		{
			name: "detect full inventory",
			setupEntity: func() *Entity {
				e := NewEntity(2)
				e.AddComponent(&EbitenInput{})
				items := make([]*item.Item, 10)
				for i := range items {
					items[i] = &item.Item{ID: "test", Name: "Test Item"}
				}
				e.AddComponent(&InventoryComponent{Items: items, MaxItems: 10})
				return e
			},
			wantHintSet: true,
			wantContext: "inventory_full",
		},
		{
			name: "no hint for healthy player",
			setupEntity: func() *Entity {
				e := NewEntity(3)
				e.AddComponent(&EbitenInput{})
				e.AddComponent(&HealthComponent{Current: 100, Max: 100})
				return e
			},
			wantHintSet: false,
			wantContext: "",
		},
		{
			name: "no hint for non-player entity",
			setupEntity: func() *Entity {
				e := NewEntity(4)
				e.AddComponent(&HealthComponent{Current: 20, Max: 100})
				return e
			},
			wantHintSet: false,
			wantContext: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hs := NewHelpSystem()
			hs.Enabled = true
			hs.ShowQuickHint = false
			hs.CurrentHint = ""

			entity := tt.setupEntity()
			entities := []*Entity{entity}

			hs.Update(entities, 0.016)

			if tt.wantHintSet {
				if !hs.ShowQuickHint {
					t.Error("Update() did not set ShowQuickHint for detected condition")
				}
				if hs.CurrentHint == "" {
					t.Error("Update() did not set CurrentHint for detected condition")
				}
				// Verify the hint is related to the expected context
				expectedHint := hs.QuickHints[tt.wantContext]
				if hs.CurrentHint != expectedHint {
					t.Errorf("Update() CurrentHint = %q, want hint for context %q", 
						hs.CurrentHint, tt.wantContext)
				}
			} else {
				if hs.ShowQuickHint {
					t.Error("Update() unexpectedly set ShowQuickHint")
				}
			}
		})
	}
}

// TestUpdateDoesNotOverrideHint validates that Update doesn't override existing hints.
func TestUpdateDoesNotOverrideHint(t *testing.T) {
	hs := NewHelpSystem()
	hs.Enabled = true
	hs.ShowQuickHint = true
	hs.CurrentHint = "Important existing hint"

	// Create entity with low health that would normally trigger hint
	entity := NewEntity(1)
	entity.AddComponent(&EbitenInput{})
	entity.AddComponent(&HealthComponent{Current: 20, Max: 100})

	entities := []*Entity{entity}
	hs.Update(entities, 0.016)

	// Should not override existing hint
	if hs.CurrentHint != "Important existing hint" {
		t.Errorf("Update() overrode existing hint: got %q", hs.CurrentHint)
	}
}

// TestDefaultHelpTopics validates default topics have required fields.
func TestDefaultHelpTopics(t *testing.T) {
	topics := createDefaultHelpTopics()

	requiredTopics := []string{"controls", "combat", "inventory", "progression", "world", "multiplayer"}
	for _, topicID := range requiredTopics {
		topic, exists := topics[topicID]
		if !exists {
			t.Errorf("Required topic %q not found in defaults", topicID)
			continue
		}

		if topic.ID != topicID {
			t.Errorf("Topic %q has mismatched ID %q", topicID, topic.ID)
		}
		if topic.Title == "" {
			t.Errorf("Topic %q has empty Title", topicID)
		}
		if len(topic.Content) == 0 {
			t.Errorf("Topic %q has no Content", topicID)
		}
		// Keys can be empty for some topics (e.g., multiplayer)
	}
}

// TestDefaultQuickHints validates default hints have text.
func TestDefaultQuickHints(t *testing.T) {
	hints := createDefaultQuickHints()

	requiredHints := []string{
		"low_health", "level_up", "inventory_full", "no_mana",
		"enemy_nearby", "item_dropped", "boss_ahead", "quest_complete", "first_death",
	}

	for _, hintID := range requiredHints {
		hint, exists := hints[hintID]
		if !exists {
			t.Errorf("Required hint %q not found in defaults", hintID)
			continue
		}
		if hint == "" {
			t.Errorf("Hint %q has empty text", hintID)
		}
	}
}

// TestUISystemInterface validates EbitenHelpSystem implements UISystem.
func TestUISystemInterface(t *testing.T) {
	var _ UISystem = (*EbitenHelpSystem)(nil)

	hs := NewHelpSystem()

	// Test interface methods
	if hs.IsActive() {
		t.Error("New help system should not be active")
	}

	hs.SetActive(true)
	if !hs.IsActive() {
		t.Error("SetActive(true) should make system active")
	}

	hs.SetActive(false)
	if hs.IsActive() {
		t.Error("SetActive(false) should make system inactive")
	}
}

// TestHelpTopicStructure validates HelpTopic struct fields.
func TestHelpTopicStructure(t *testing.T) {
	topic := HelpTopic{
		ID:      "test",
		Title:   "Test Topic",
		Content: []string{"Line 1", "Line 2"},
		Keys:    []string{"A", "B"},
	}

	if topic.ID != "test" {
		t.Errorf("HelpTopic.ID = %q, want %q", topic.ID, "test")
	}
	if topic.Title != "Test Topic" {
		t.Errorf("HelpTopic.Title = %q, want %q", topic.Title, "Test Topic")
	}
	if len(topic.Content) != 2 {
		t.Errorf("HelpTopic.Content length = %d, want 2", len(topic.Content))
	}
	if len(topic.Keys) != 2 {
		t.Errorf("HelpTopic.Keys length = %d, want 2", len(topic.Keys))
	}
}

// TestEbitenHelpSystemFields validates struct initialization.
func TestEbitenHelpSystemFields(t *testing.T) {
	hs := &EbitenHelpSystem{
		Enabled:       false,
		Visible:       true,
		CurrentTopic:  "test",
		Topics:        map[string]HelpTopic{"test": {ID: "test"}},
		QuickHints:    map[string]string{"test": "hint"},
		ShowQuickHint: true,
		CurrentHint:   "hint text",
	}

	if hs.Enabled {
		t.Error("Field Enabled not set correctly")
	}
	if !hs.Visible {
		t.Error("Field Visible not set correctly")
	}
	if hs.CurrentTopic != "test" {
		t.Error("Field CurrentTopic not set correctly")
	}
	if len(hs.Topics) != 1 {
		t.Error("Field Topics not set correctly")
	}
	if len(hs.QuickHints) != 1 {
		t.Error("Field QuickHints not set correctly")
	}
	if !hs.ShowQuickHint {
		t.Error("Field ShowQuickHint not set correctly")
	}
	if hs.CurrentHint != "hint text" {
		t.Error("Field CurrentHint not set correctly")
	}
}
