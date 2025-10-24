//go:build test
// +build test

package engine

import (
	"testing"
)

// Test stub types and functions for tutorial system
// Note: InputComponent is now defined in components_test_stub.go

type TutorialStep struct {
	ID          string
	Title       string
	Description string
	Objective   string
	Completed   bool
	Condition   func(*World) bool
}

type TutorialSystem struct {
	Enabled         bool
	CurrentStepIdx  int
	Steps           []TutorialStep
	ShowUI          bool
	NotificationMsg string
	NotificationTTL float64
}

func NewTutorialSystem() *TutorialSystem {
	return &TutorialSystem{
		Enabled:        true,
		ShowUI:         true,
		Steps:          createDefaultTutorialSteps(),
		CurrentStepIdx: 0,
	}
}

func createDefaultTutorialSteps() []TutorialStep {
	return []TutorialStep{
		{
			ID:          "welcome",
			Title:       "Welcome",
			Description: "Game start",
			Objective:   "Press any key",
			Completed:   false,
			Condition: func(world *World) bool {
				for _, entity := range world.GetEntities() {
					if comp, ok := entity.GetComponent("input"); ok {
						input := comp.(*InputComponent)
						return input.AnyKeyPressed
					}
				}
				return false
			},
		},
		{
			ID:          "movement",
			Title:       "Movement",
			Description: "Learn to move",
			Objective:   "Move around",
			Completed:   false,
			Condition: func(world *World) bool {
				for _, entity := range world.GetEntities() {
					if comp, ok := entity.GetComponent("position"); ok {
						pos := comp.(*PositionComponent)
						distFromStart := (pos.X-400)*(pos.X-400) + (pos.Y-300)*(pos.Y-300)
						return distFromStart > 2500
					}
				}
				return false
			},
		},
		{
			ID:          "inventory",
			Title:       "Inventory",
			Description: "Manage items",
			Objective:   "Open inventory",
			Completed:   false,
			Condition: func(world *World) bool {
				for _, entity := range world.GetEntities() {
					if comp, ok := entity.GetComponent("inventory"); ok {
						inv := comp.(*InventoryComponent)
						return len(inv.Items) > 0
					}
				}
				return false
			},
		},
		{
			ID:          "skills",
			Title:       "Skills",
			Description: "Level up",
			Objective:   "Reach level 2",
			Completed:   false,
			Condition: func(world *World) bool {
				for _, entity := range world.GetEntities() {
					if comp, ok := entity.GetComponent("experience"); ok {
						exp := comp.(*ExperienceComponent)
						return exp.Level >= 2
					}
				}
				return false
			},
		},
	}
}

func (ts *TutorialSystem) Update(entities []*Entity, deltaTime float64) {
	if !ts.Enabled || ts.CurrentStepIdx >= len(ts.Steps) {
		return
	}

	world := &World{entities: make(map[uint64]*Entity), entityListDirty: true}
	for _, entity := range entities {
		world.entities[entity.ID] = entity
	}

	if ts.NotificationTTL > 0 {
		ts.NotificationTTL -= deltaTime
		if ts.NotificationTTL <= 0 {
			ts.NotificationMsg = ""
		}
	}

	currentStep := &ts.Steps[ts.CurrentStepIdx]
	condResult := currentStep.Condition(world)
	if !currentStep.Completed && condResult {
		currentStep.Completed = true
		ts.CurrentStepIdx++
		if ts.CurrentStepIdx >= len(ts.Steps) {
			ts.Enabled = false
		}
	}
}

func (ts *TutorialSystem) GetCurrentStep() *TutorialStep {
	if !ts.Enabled || ts.CurrentStepIdx >= len(ts.Steps) {
		return nil
	}
	return &ts.Steps[ts.CurrentStepIdx]
}

func (ts *TutorialSystem) GetProgress() float64 {
	if len(ts.Steps) == 0 {
		return 1.0
	}
	return float64(ts.CurrentStepIdx) / float64(len(ts.Steps))
}

func (ts *TutorialSystem) Skip() {
	if ts.Enabled && ts.CurrentStepIdx < len(ts.Steps) {
		ts.Steps[ts.CurrentStepIdx].Completed = true
		ts.CurrentStepIdx++
		if ts.CurrentStepIdx >= len(ts.Steps) {
			ts.Enabled = false
		}
	}
}

func (ts *TutorialSystem) SkipAll() {
	ts.Enabled = false
	ts.ShowUI = false
}

func (ts *TutorialSystem) Reset() {
	ts.Enabled = true
	ts.ShowUI = true
	ts.CurrentStepIdx = 0
	ts.NotificationMsg = ""
	ts.NotificationTTL = 0
	for i := range ts.Steps {
		ts.Steps[i].Completed = false
	}
}

// ShowNotification displays a notification message for the specified duration.
func (ts *TutorialSystem) ShowNotification(msg string, duration float64) {
	ts.NotificationMsg = msg
	ts.NotificationTTL = duration
}

// GAP-003 REPAIR: Tutorial state serialization for save/load (test stub)
func (ts *TutorialSystem) ExportState() (enabled, showUI bool, currentStepIdx int, completedSteps map[string]bool) {
	completedSteps = make(map[string]bool)
	for _, step := range ts.Steps {
		if step.Completed {
			completedSteps[step.ID] = true
		}
	}
	return ts.Enabled, ts.ShowUI, ts.CurrentStepIdx, completedSteps
}

func (ts *TutorialSystem) ImportState(enabled, showUI bool, currentStepIdx int, completedSteps map[string]bool) {
	ts.Enabled = enabled
	ts.ShowUI = showUI
	ts.CurrentStepIdx = currentStepIdx

	for i := range ts.Steps {
		stepID := ts.Steps[i].ID
		if completed, ok := completedSteps[stepID]; ok {
			ts.Steps[i].Completed = completed
		}
	}

	// GAP-003 REPAIR: Clamp to valid range
	if ts.CurrentStepIdx >= len(ts.Steps) {
		ts.CurrentStepIdx = len(ts.Steps) - 1
	}
	if ts.CurrentStepIdx < 0 {
		ts.CurrentStepIdx = 0
	}
}

// GAP-006 REPAIR: Public API for querying tutorial state (test stub)
func (ts *TutorialSystem) IsStepCompleted(stepID string) bool {
	for _, step := range ts.Steps {
		if step.ID == stepID {
			return step.Completed
		}
	}
	return false
}

func (ts *TutorialSystem) GetStepByID(stepID string) *TutorialStep {
	for i := range ts.Steps {
		if ts.Steps[i].ID == stepID {
			return &ts.Steps[i]
		}
	}
	return nil
}

func (ts *TutorialSystem) IsActive() bool {
	return ts.Enabled && ts.ShowUI
}

func (ts *TutorialSystem) GetCurrentStepID() string {
	step := ts.GetCurrentStep()
	if step == nil {
		return ""
	}
	return step.ID
}

func (ts *TutorialSystem) GetAllSteps() []TutorialStep {
	steps := make([]TutorialStep, len(ts.Steps))
	copy(steps, ts.Steps)
	return steps
}

func splitWords(str string) []string {
	var words []string
	currentWord := ""

	for _, ch := range str {
		if ch == ' ' {
			if currentWord != "" {
				words = append(words, currentWord)
				currentWord = ""
			}
		} else {
			currentWord += string(ch)
		}
	}

	if currentWord != "" {
		words = append(words, currentWord)
	}

	return words
}

func TestNewTutorialSystem(t *testing.T) {
	ts := NewTutorialSystem()

	if ts == nil {
		t.Fatal("NewTutorialSystem returned nil")
	}

	if !ts.Enabled {
		t.Error("Tutorial should be enabled by default")
	}

	if !ts.ShowUI {
		t.Error("Tutorial UI should be shown by default")
	}

	if len(ts.Steps) == 0 {
		t.Error("Tutorial should have default steps")
	}

	if ts.CurrentStepIdx != 0 {
		t.Error("Tutorial should start at step 0")
	}
}

func TestTutorialSystemProgress(t *testing.T) {
	ts := NewTutorialSystem()
	totalSteps := len(ts.Steps)

	// Progress should be 0 at start
	if ts.GetProgress() != 0.0 {
		t.Errorf("Initial progress should be 0.0, got %f", ts.GetProgress())
	}

	// Simulate completing steps
	for i := 0; i < totalSteps; i++ {
		expectedProgress := float64(i) / float64(totalSteps)
		actualProgress := ts.GetProgress()

		if actualProgress != expectedProgress {
			t.Errorf("Progress at step %d: expected %f, got %f", i, expectedProgress, actualProgress)
		}

		// Complete current step
		ts.Steps[i].Completed = true
		ts.CurrentStepIdx++
	}

	// Progress should be 1.0 when complete
	if ts.GetProgress() != 1.0 {
		t.Errorf("Final progress should be 1.0, got %f", ts.GetProgress())
	}
}

func TestTutorialSystemGetCurrentStep(t *testing.T) {
	ts := NewTutorialSystem()

	// Should return first step initially
	step := ts.GetCurrentStep()
	if step == nil {
		t.Fatal("GetCurrentStep returned nil for first step")
	}

	if step.ID != ts.Steps[0].ID {
		t.Errorf("Expected first step ID %s, got %s", ts.Steps[0].ID, step.ID)
	}

	// Should return nil when complete
	ts.CurrentStepIdx = len(ts.Steps)
	step = ts.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil when tutorial is complete")
	}

	// Should return nil when disabled
	ts.CurrentStepIdx = 0
	ts.Enabled = false
	step = ts.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil when tutorial is disabled")
	}
}

func TestTutorialSystemSkip(t *testing.T) {
	ts := NewTutorialSystem()
	initialStep := ts.CurrentStepIdx

	// Skip should advance to next step
	ts.Skip()

	if ts.CurrentStepIdx != initialStep+1 {
		t.Errorf("Expected step %d after skip, got %d", initialStep+1, ts.CurrentStepIdx)
	}

	if !ts.Steps[initialStep].Completed {
		t.Error("Skipped step should be marked as completed")
	}

	// Skip to end
	for ts.CurrentStepIdx < len(ts.Steps) {
		ts.Skip()
	}

	if ts.Enabled {
		t.Error("Tutorial should be disabled after skipping all steps")
	}
}

func TestTutorialSystemSkipAll(t *testing.T) {
	ts := NewTutorialSystem()

	ts.SkipAll()

	if ts.Enabled {
		t.Error("Tutorial should be disabled after SkipAll")
	}

	if ts.ShowUI {
		t.Error("Tutorial UI should be hidden after SkipAll")
	}
}

func TestTutorialSystemReset(t *testing.T) {
	ts := NewTutorialSystem()

	// Complete some steps
	ts.Steps[0].Completed = true
	ts.Steps[1].Completed = true
	ts.CurrentStepIdx = 2
	ts.Enabled = false
	ts.ShowUI = false

	// Reset
	ts.Reset()

	if !ts.Enabled {
		t.Error("Tutorial should be enabled after reset")
	}

	if !ts.ShowUI {
		t.Error("Tutorial UI should be shown after reset")
	}

	if ts.CurrentStepIdx != 0 {
		t.Error("Tutorial should start at step 0 after reset")
	}

	// All steps should be incomplete
	for i, step := range ts.Steps {
		if step.Completed {
			t.Errorf("Step %d should not be completed after reset", i)
		}
	}
}

func TestTutorialSystemUpdate(t *testing.T) {
	ts := NewTutorialSystem()
	world := NewWorld()

	// Create a simple entity for testing
	entity := NewEntity(1)
	entity.AddComponent(&InputComponent{})
	entity.AddComponent(&PositionComponent{X: 400, Y: 300})
	world.AddEntity(entity)

	entities := world.GetEntities()

	// Update should not advance without condition met
	initialStep := ts.CurrentStepIdx
	ts.Update(entities, 0.016)

	if ts.CurrentStepIdx != initialStep && !ts.Steps[initialStep].Completed {
		t.Error("Tutorial should not advance without condition being met")
	}

	// Disabled tutorial should not update
	ts.Enabled = false
	initialStep = ts.CurrentStepIdx
	ts.Update(entities, 0.016)

	if ts.CurrentStepIdx != initialStep {
		t.Error("Disabled tutorial should not advance")
	}
}

func TestTutorialSystemNotifications(t *testing.T) {
	ts := NewTutorialSystem()

	// Set notification
	ts.NotificationMsg = "Test message"
	ts.NotificationTTL = 1.0

	if ts.NotificationMsg != "Test message" {
		t.Error("Notification message not set correctly")
	}

	// Update to decrease TTL
	world := NewWorld()
	entities := world.GetEntities()

	ts.Update(entities, 0.5)

	if ts.NotificationTTL != 0.5 {
		t.Errorf("Expected TTL 0.5, got %f", ts.NotificationTTL)
	}

	// Update past TTL
	ts.Update(entities, 1.0)

	if ts.NotificationMsg != "" {
		t.Error("Notification should be cleared after TTL expires")
	}

	if ts.NotificationTTL > 0 {
		t.Error("TTL should be 0 or negative after expiration")
	}
}

func TestTutorialSystemSteps(t *testing.T) {
	steps := createDefaultTutorialSteps()

	if len(steps) == 0 {
		t.Fatal("No default tutorial steps created")
	}

	// Verify all steps have required fields
	for i, step := range steps {
		if step.ID == "" {
			t.Errorf("Step %d has empty ID", i)
		}

		if step.Title == "" {
			t.Errorf("Step %d (%s) has empty title", i, step.ID)
		}

		if step.Description == "" {
			t.Errorf("Step %d (%s) has empty description", i, step.ID)
		}

		if step.Objective == "" {
			t.Errorf("Step %d (%s) has empty objective", i, step.ID)
		}

		if step.Condition == nil {
			t.Errorf("Step %d (%s) has nil condition", i, step.ID)
		}

		if step.Completed {
			t.Errorf("Step %d (%s) should not be completed by default", i, step.ID)
		}
	}
}

func TestTutorialSystemStepConditions(t *testing.T) {
	ts := NewTutorialSystem()
	world := NewWorld()

	// Create player entity with various components
	player := NewEntity(1)
	player.AddComponent(&InputComponent{}) // Add input component
	player.AddComponent(&PositionComponent{X: 400, Y: 300})
	player.AddComponent(&HealthComponent{Current: 100, Max: 100})
	player.AddComponent(&InventoryComponent{})
	player.AddComponent(&ExperienceComponent{Level: 1})
	world.AddEntity(player)

	// Update world to process added entities
	world.Update(0.016)

	// Debug: check entities
	entities := world.GetEntities()
	if len(entities) == 0 {
		t.Fatal("World has no entities after update")
	}

	// Test movement condition
	movementStep := findStepByID(ts.Steps, "movement")
	if movementStep != nil {
		// Initially should not be complete
		if movementStep.Condition(world) {
			t.Error("Movement step should not be complete at spawn position")
		}

		// Move player far enough
		if posComp, ok := player.GetComponent("position"); ok {
			pos := posComp.(*PositionComponent)
			pos.X = 500 // Moved 100 units from spawn
			pos.Y = 300

			if !movementStep.Condition(world) {
				t.Error("Movement step should be complete after moving")
			}
		}
	}

	// Test skill/level condition
	skillStep := findStepByID(ts.Steps, "skills")
	if skillStep != nil {
		// Initially should not be complete (level 1)
		if skillStep.Condition(world) {
			t.Error("Skills step should not be complete at level 1")
		}

		// Level up
		if expComp, ok := player.GetComponent("experience"); ok {
			exp := expComp.(*ExperienceComponent)
			exp.Level = 2

			if !skillStep.Condition(world) {
				t.Error("Skills step should be complete at level 2")
			}
		}
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"one", []string{"one"}},
		{"", []string{}},
		{"  spaces  around  ", []string{"spaces", "around"}},
		{"multiple   spaces", []string{"multiple", "spaces"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitWords(tt.input)

			if len(got) != len(tt.want) {
				t.Errorf("splitWords(%q) length = %d, want %d", tt.input, len(got), len(tt.want))
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitWords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// Helper function to find a step by ID
func findStepByID(steps []TutorialStep, id string) *TutorialStep {
	for i := range steps {
		if steps[i].ID == id {
			return &steps[i]
		}
	}
	return nil
}
