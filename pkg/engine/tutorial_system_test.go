package engine

import (
	"strings"
	"testing"
)

// TestNewTutorialSystem tests tutorial system creation
func TestNewTutorialSystem(t *testing.T) {
	ts := NewTutorialSystem()

	if ts == nil {
		t.Fatal("NewTutorialSystem returned nil")
	}

	if !ts.Enabled {
		t.Error("Tutorial should be enabled by default")
	}

	if !ts.ShowUI {
		t.Error("UI should be shown by default")
	}

	if ts.CurrentStepIdx != 0 {
		t.Errorf("CurrentStepIdx = %d, want 0", ts.CurrentStepIdx)
	}

	if len(ts.Steps) == 0 {
		t.Error("Steps should not be empty")
	}

	// Verify default steps are created
	expectedSteps := []string{"welcome", "movement", "combat", "health", "inventory", "skills", "exploration"}
	if len(ts.Steps) != len(expectedSteps) {
		t.Errorf("Expected %d steps, got %d", len(expectedSteps), len(ts.Steps))
	}

	for i, expectedID := range expectedSteps {
		if i >= len(ts.Steps) {
			break
		}
		if ts.Steps[i].ID != expectedID {
			t.Errorf("Step %d: ID = %s, want %s", i, ts.Steps[i].ID, expectedID)
		}
		if ts.Steps[i].Completed {
			t.Errorf("Step %d should not be completed initially", i)
		}
	}
}

// TestTutorialSystem_GetCurrentStep tests current step retrieval
func TestTutorialSystem_GetCurrentStep(t *testing.T) {
	ts := NewTutorialSystem()

	// Get first step
	step := ts.GetCurrentStep()
	if step == nil {
		t.Fatal("GetCurrentStep returned nil at start")
	}
	if step.ID != "welcome" {
		t.Errorf("First step ID = %s, want 'welcome'", step.ID)
	}

	// Advance to next step
	ts.CurrentStepIdx = 1
	step = ts.GetCurrentStep()
	if step == nil {
		t.Fatal("GetCurrentStep returned nil for step 1")
	}
	if step.ID != "movement" {
		t.Errorf("Second step ID = %s, want 'movement'", step.ID)
	}

	// After all steps
	ts.CurrentStepIdx = len(ts.Steps)
	step = ts.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil after all steps")
	}

	// When disabled
	ts.CurrentStepIdx = 0
	ts.Enabled = false
	step = ts.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil when disabled")
	}
}

// TestTutorialSystem_GetProgress tests progress calculation
func TestTutorialSystem_GetProgress(t *testing.T) {
	ts := NewTutorialSystem()

	tests := []struct {
		name          string
		stepIdx       int
		wantMin       float64
		wantMax       float64
	}{
		{"Start", 0, 0.0, 0.01},
		{"Mid", 3, 0.42, 0.44},
		{"Near end", 6, 0.85, 0.87},
		{"Complete", 7, 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts.CurrentStepIdx = tt.stepIdx
			progress := ts.GetProgress()
			if progress < tt.wantMin || progress > tt.wantMax {
				t.Errorf("GetProgress() = %v, want between %v and %v", progress, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestTutorialSystem_GetProgress_EmptySteps tests progress with no steps
func TestTutorialSystem_GetProgress_EmptySteps(t *testing.T) {
	ts := &EbitenTutorialSystem{
		Steps: []TutorialStep{},
	}

	progress := ts.GetProgress()
	if progress != 1.0 {
		t.Errorf("GetProgress() with empty steps = %v, want 1.0", progress)
	}
}

// TestTutorialSystem_Skip tests skipping current step
func TestTutorialSystem_Skip(t *testing.T) {
	ts := NewTutorialSystem()

	initialIdx := ts.CurrentStepIdx
	step := &ts.Steps[initialIdx]

	ts.Skip()

	if !step.Completed {
		t.Error("Step should be marked as completed after Skip()")
	}
	if ts.CurrentStepIdx != initialIdx+1 {
		t.Errorf("CurrentStepIdx = %d, want %d", ts.CurrentStepIdx, initialIdx+1)
	}

	// Skip all remaining steps
	for ts.CurrentStepIdx < len(ts.Steps) {
		ts.Skip()
	}

	if ts.Enabled {
		t.Error("Tutorial should be disabled after skipping all steps")
	}
}

// TestTutorialSystem_Skip_WhenDisabled tests skip does nothing when disabled
func TestTutorialSystem_Skip_WhenDisabled(t *testing.T) {
	ts := NewTutorialSystem()
	ts.Enabled = false
	initialIdx := ts.CurrentStepIdx

	ts.Skip()

	if ts.CurrentStepIdx != initialIdx {
		t.Error("Skip should not advance when disabled")
	}
}

// TestTutorialSystem_SkipAll tests skipping entire tutorial
func TestTutorialSystem_SkipAll(t *testing.T) {
	ts := NewTutorialSystem()

	ts.SkipAll()

	if ts.Enabled {
		t.Error("Tutorial should be disabled after SkipAll()")
	}
	if ts.ShowUI {
		t.Error("UI should be hidden after SkipAll()")
	}
}

// TestTutorialSystem_Reset tests resetting tutorial
func TestTutorialSystem_Reset(t *testing.T) {
	ts := NewTutorialSystem()

	// Modify state
	ts.Enabled = false
	ts.ShowUI = false
	ts.CurrentStepIdx = 5
	ts.NotificationMsg = "Test message"
	ts.NotificationTTL = 3.0
	for i := range ts.Steps {
		ts.Steps[i].Completed = true
	}

	ts.Reset()

	if !ts.Enabled {
		t.Error("Tutorial should be enabled after Reset()")
	}
	if !ts.ShowUI {
		t.Error("UI should be shown after Reset()")
	}
	if ts.CurrentStepIdx != 0 {
		t.Errorf("CurrentStepIdx = %d, want 0 after Reset()", ts.CurrentStepIdx)
	}
	if ts.NotificationMsg != "" {
		t.Error("NotificationMsg should be cleared after Reset()")
	}
	if ts.NotificationTTL != 0 {
		t.Error("NotificationTTL should be 0 after Reset()")
	}

	for i, step := range ts.Steps {
		if step.Completed {
			t.Errorf("Step %d should not be completed after Reset()", i)
		}
	}
}

// TestTutorialSystem_IsStepCompleted tests step completion check
func TestTutorialSystem_IsStepCompleted(t *testing.T) {
	ts := NewTutorialSystem()

	// Initially not completed
	if ts.IsStepCompleted("welcome") {
		t.Error("'welcome' step should not be completed initially")
	}

	// Mark as completed
	ts.Steps[0].Completed = true
	if !ts.IsStepCompleted("welcome") {
		t.Error("'welcome' step should be completed after marking")
	}

	// Non-existent step
	if ts.IsStepCompleted("nonexistent") {
		t.Error("Non-existent step should return false")
	}
}

// TestTutorialSystem_GetStepByID tests step retrieval by ID
func TestTutorialSystem_GetStepByID(t *testing.T) {
	ts := NewTutorialSystem()

	// Valid step
	step := ts.GetStepByID("combat")
	if step == nil {
		t.Fatal("GetStepByID('combat') returned nil")
	}
	if step.ID != "combat" {
		t.Errorf("Step ID = %s, want 'combat'", step.ID)
	}
	if !strings.Contains(step.Title, "Combat") {
		t.Errorf("Step title should contain 'Combat', got: %s", step.Title)
	}

	// Non-existent step
	step = ts.GetStepByID("nonexistent")
	if step != nil {
		t.Error("GetStepByID('nonexistent') should return nil")
	}
}

// TestTutorialSystem_IsActive tests active state check
func TestTutorialSystem_IsActive(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
		showUI  bool
		want    bool
	}{
		{"Both enabled", true, true, true},
		{"Disabled, UI shown", false, true, false},
		{"Enabled, UI hidden", true, false, false},
		{"Both disabled", false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTutorialSystem()
			ts.Enabled = tt.enabled
			ts.ShowUI = tt.showUI

			got := ts.IsActive()
			if got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTutorialSystem_GetCurrentStepID tests current step ID retrieval
func TestTutorialSystem_GetCurrentStepID(t *testing.T) {
	ts := NewTutorialSystem()

	// First step
	id := ts.GetCurrentStepID()
	if id != "welcome" {
		t.Errorf("GetCurrentStepID() = %s, want 'welcome'", id)
	}

	// Advance
	ts.CurrentStepIdx = 2
	id = ts.GetCurrentStepID()
	if id != "combat" {
		t.Errorf("GetCurrentStepID() = %s, want 'combat'", id)
	}

	// Complete
	ts.CurrentStepIdx = len(ts.Steps)
	id = ts.GetCurrentStepID()
	if id != "" {
		t.Errorf("GetCurrentStepID() = %s, want empty string when complete", id)
	}
}

// TestTutorialSystem_GetAllSteps tests retrieving all steps
func TestTutorialSystem_GetAllSteps(t *testing.T) {
	ts := NewTutorialSystem()

	steps := ts.GetAllSteps()

	if len(steps) != len(ts.Steps) {
		t.Errorf("GetAllSteps() returned %d steps, want %d", len(steps), len(ts.Steps))
	}

	// Verify it's a copy (modification doesn't affect original)
	steps[0].Completed = true
	if ts.Steps[0].Completed {
		t.Error("Modifying returned steps should not affect original")
	}
}

// TestTutorialSystem_ExportState tests state export
func TestTutorialSystem_ExportState(t *testing.T) {
	ts := NewTutorialSystem()

	// Modify state
	ts.Enabled = false
	ts.ShowUI = true
	ts.CurrentStepIdx = 3
	ts.Steps[0].Completed = true
	ts.Steps[1].Completed = true

	enabled, showUI, currentStepIdx, completedSteps := ts.ExportState()

	if enabled != false {
		t.Error("Exported enabled should be false")
	}
	if showUI != true {
		t.Error("Exported showUI should be true")
	}
	if currentStepIdx != 3 {
		t.Errorf("Exported currentStepIdx = %d, want 3", currentStepIdx)
	}
	if len(completedSteps) != 2 {
		t.Errorf("Exported %d completed steps, want 2", len(completedSteps))
	}
	if !completedSteps["welcome"] {
		t.Error("'welcome' should be in completed steps")
	}
	if !completedSteps["movement"] {
		t.Error("'movement' should be in completed steps")
	}
}

// TestTutorialSystem_ImportState tests state import
func TestTutorialSystem_ImportState(t *testing.T) {
	ts := NewTutorialSystem()

	// Import state
	completedSteps := map[string]bool{
		"welcome":  true,
		"movement": true,
		"combat":   true,
	}

	ts.ImportState(false, true, 3, completedSteps)

	if ts.Enabled != false {
		t.Error("Enabled should be false after import")
	}
	if ts.ShowUI != true {
		t.Error("ShowUI should be true after import")
	}
	if ts.CurrentStepIdx != 3 {
		t.Errorf("CurrentStepIdx = %d, want 3 after import", ts.CurrentStepIdx)
	}
	if !ts.Steps[0].Completed {
		t.Error("'welcome' step should be completed after import")
	}
	if !ts.Steps[1].Completed {
		t.Error("'movement' step should be completed after import")
	}
	if !ts.Steps[2].Completed {
		t.Error("'combat' step should be completed after import")
	}
	if ts.Steps[3].Completed {
		t.Error("'health' step should not be completed after import")
	}
}

// TestTutorialSystem_ImportState_Validation tests import validation
func TestTutorialSystem_ImportState_Validation(t *testing.T) {
	tests := []struct {
		name          string
		stepIdx       int
		wantIdx       int
	}{
		{"Valid index", 3, 3},
		{"Negative index", -5, 0},
		{"Beyond steps", 100, 6}, // Last step index
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := NewTutorialSystem()
			ts.ImportState(true, true, tt.stepIdx, nil)

			if ts.CurrentStepIdx != tt.wantIdx {
				t.Errorf("CurrentStepIdx = %d, want %d", ts.CurrentStepIdx, tt.wantIdx)
			}
		})
	}
}

// TestTutorialSystem_ShowNotification tests notification display
func TestTutorialSystem_ShowNotification(t *testing.T) {
	ts := NewTutorialSystem()

	msg := "Test notification"
	duration := 2.5

	ts.ShowNotification(msg, duration)

	if ts.NotificationMsg != msg {
		t.Errorf("NotificationMsg = %s, want %s", ts.NotificationMsg, msg)
	}
	if ts.NotificationTTL != duration {
		t.Errorf("NotificationTTL = %v, want %v", ts.NotificationTTL, duration)
	}
}

// TestTutorialSystem_Update tests update logic
func TestTutorialSystem_Update(t *testing.T) {
	ts := NewTutorialSystem()

	// Create mock world with player entity
	player := NewEntity(1)
	
	// Create entities slice
	entities := []*Entity{player}

	// Update without completing condition
	ts.Update(entities, 0.016)

	if ts.CurrentStepIdx != 0 {
		t.Error("Step should not advance without condition being met")
	}

	// Manually mark step as meeting condition by completing it
	// Note: We can't easily test the condition functions without full game state
	// but we can test the progression logic
	ts.Steps[0].Completed = true
	
	// Reset for proper progression test
	ts.CurrentStepIdx = 0
	ts.Steps[0].Completed = false
	
	// Create a custom step with simple condition
	ts.Steps = []TutorialStep{
		{
			ID:        "test1",
			Title:     "Test Step 1",
			Completed: false,
			Condition: func(world *World) bool {
				return true // Always complete
			},
		},
		{
			ID:        "test2",
			Title:     "Test Step 2",
			Completed: false,
			Condition: func(world *World) bool {
				return false // Never complete
			},
		},
	}

	// Update should complete first step and advance
	ts.Update(entities, 0.016)

	if !ts.Steps[0].Completed {
		t.Error("First step should be completed")
	}
	if ts.CurrentStepIdx != 1 {
		t.Errorf("CurrentStepIdx = %d, want 1", ts.CurrentStepIdx)
	}
	if ts.NotificationMsg == "" {
		t.Error("Notification should be set after step completion")
	}
	if ts.NotificationTTL <= 0 {
		t.Error("NotificationTTL should be positive after step completion")
	}
}

// TestTutorialSystem_Update_NotificationTTL tests notification timeout
func TestTutorialSystem_Update_NotificationTTL(t *testing.T) {
	ts := NewTutorialSystem()
	
	ts.NotificationMsg = "Test"
	ts.NotificationTTL = 0.1

	entities := []*Entity{}

	// Update multiple times
	for i := 0; i < 10; i++ {
		ts.Update(entities, 0.016)
	}

	if ts.NotificationMsg != "" {
		t.Error("Notification should be cleared after TTL expires")
	}
	if ts.NotificationTTL > 0 {
		t.Error("NotificationTTL should be 0 or negative")
	}
}

// TestTutorialSystem_Update_WhenDisabled tests update does nothing when disabled
func TestTutorialSystem_Update_WhenDisabled(t *testing.T) {
	ts := NewTutorialSystem()
	ts.Enabled = false

	initialIdx := ts.CurrentStepIdx
	entities := []*Entity{}

	ts.Update(entities, 0.016)

	if ts.CurrentStepIdx != initialIdx {
		t.Error("Update should not change state when disabled")
	}
}

// TestTutorialSystem_Update_Complete tests completing all steps
func TestTutorialSystem_Update_Complete(t *testing.T) {
	ts := NewTutorialSystem()

	// Create steps that all complete immediately
	ts.Steps = []TutorialStep{
		{
			ID:        "step1",
			Completed: false,
			Condition: func(world *World) bool { return true },
		},
		{
			ID:        "step2",
			Completed: false,
			Condition: func(world *World) bool { return true },
		},
	}
	ts.CurrentStepIdx = 0
	ts.Enabled = true

	entities := []*Entity{}

	// Update to complete first step
	ts.Update(entities, 0.016)
	if ts.CurrentStepIdx != 1 {
		t.Fatal("Should advance to step 2")
	}

	// Update to complete second step
	ts.Update(entities, 0.016)
	if ts.CurrentStepIdx != 2 {
		t.Error("Should advance past all steps")
	}
	if ts.Enabled {
		t.Error("Tutorial should be disabled after completing all steps")
	}
	if !strings.Contains(ts.NotificationMsg, "Complete") {
		t.Errorf("Final notification should mention 'Complete', got: %s", ts.NotificationMsg)
	}
}

// TestTutorialSystem_SetActive tests SetActive interface method
func TestTutorialSystem_SetActive(t *testing.T) {
	ts := NewTutorialSystem()

	ts.SetActive(false)
	if ts.ShowUI {
		t.Error("ShowUI should be false after SetActive(false)")
	}

	ts.SetActive(true)
	if !ts.ShowUI {
		t.Error("ShowUI should be true after SetActive(true)")
	}
}

// TestTutorialStep_Validation tests that tutorial steps have required fields
func TestTutorialStep_Validation(t *testing.T) {
	ts := NewTutorialSystem()

	for i, step := range ts.Steps {
		if step.ID == "" {
			t.Errorf("Step %d: ID is empty", i)
		}
		if step.Title == "" {
			t.Errorf("Step %d: Title is empty", i)
		}
		if step.Description == "" {
			t.Errorf("Step %d: Description is empty", i)
		}
		if step.Objective == "" {
			t.Errorf("Step %d: Objective is empty", i)
		}
		if step.Condition == nil {
			t.Errorf("Step %d: Condition is nil", i)
		}
	}
}

// TestTutorialSystem_Integration tests full tutorial workflow
func TestTutorialSystem_Integration(t *testing.T) {
	ts := NewTutorialSystem()

	// Verify initial state
	if !ts.IsActive() {
		t.Error("Tutorial should be active initially")
	}

	step := ts.GetCurrentStep()
	if step == nil || step.ID != "welcome" {
		t.Error("Should start with 'welcome' step")
	}

	progress := ts.GetProgress()
	if progress != 0.0 {
		t.Errorf("Initial progress = %v, want 0.0", progress)
	}

	// Skip first step
	ts.Skip()
	if !ts.IsStepCompleted("welcome") {
		t.Error("'welcome' should be completed")
	}

	step = ts.GetCurrentStep()
	if step == nil || step.ID != "movement" {
		t.Error("Should advance to 'movement' step")
	}

	// Export state
	enabled, showUI, stepIdx, completed := ts.ExportState()
	if stepIdx != 1 {
		t.Errorf("Exported stepIdx = %d, want 1", stepIdx)
	}
	if !completed["welcome"] {
		t.Error("'welcome' should be in exported completed steps")
	}

	// Reset
	ts.Reset()
	if ts.GetCurrentStepID() != "welcome" {
		t.Error("Should return to 'welcome' after reset")
	}

	// Import previous state
	ts.ImportState(enabled, showUI, stepIdx, completed)
	if ts.GetCurrentStepID() != "movement" {
		t.Error("Should restore to 'movement' after import")
	}

	// Skip all
	ts.SkipAll()
	if ts.IsActive() {
		t.Error("Tutorial should not be active after SkipAll()")
	}
}

// TestTutorialSystem_StepProgression tests sequential step progression
func TestTutorialSystem_StepProgression(t *testing.T) {
	ts := NewTutorialSystem()

	expectedSteps := []string{"welcome", "movement", "combat", "health", "inventory", "skills", "exploration"}

	for i, expectedID := range expectedSteps {
		step := ts.GetCurrentStep()
		if step == nil {
			t.Fatalf("Step %d: GetCurrentStep returned nil", i)
		}
		if step.ID != expectedID {
			t.Errorf("Step %d: ID = %s, want %s", i, step.ID, expectedID)
		}

		ts.Skip()
	}

	// After skipping all steps
	if ts.GetCurrentStep() != nil {
		t.Error("GetCurrentStep should return nil after all steps")
	}
	if ts.Enabled {
		t.Error("Tutorial should be disabled after completing all steps")
	}
}

// TestSplitWords tests word splitting helper
func TestSplitWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"hello world", []string{"hello", "world"}},
		{"one two three", []string{"one", "two", "three"}},
		{"single", []string{"single"}},
		{"", []string{}},
		{"  spaces  between  ", []string{"spaces", "between"}},
		{"no spaces", []string{"no", "spaces"}},
	}

	for _, tt := range tests {
		got := splitWords(tt.input)
		if len(got) != len(tt.want) {
			t.Errorf("splitWords(%q) returned %d words, want %d", tt.input, len(got), len(tt.want))
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitWords(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
			}
		}
	}
}

// TestTutorialSystem_MultipleResets tests resetting multiple times
func TestTutorialSystem_MultipleResets(t *testing.T) {
	ts := NewTutorialSystem()

	for i := 0; i < 3; i++ {
		// Advance through some steps
		ts.Skip()
		ts.Skip()

		// Reset
		ts.Reset()

		// Verify reset worked
		if ts.CurrentStepIdx != 0 {
			t.Errorf("Reset %d: CurrentStepIdx = %d, want 0", i, ts.CurrentStepIdx)
		}
		if !ts.Enabled {
			t.Errorf("Reset %d: Tutorial should be enabled", i)
		}
		for j, step := range ts.Steps {
			if step.Completed {
				t.Errorf("Reset %d: Step %d should not be completed", i, j)
			}
		}
	}
}
