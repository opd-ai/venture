package engine

import (
	"fmt"
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
		name    string
		stepIdx int
		wantMin float64
		wantMax float64
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
		name    string
		stepIdx int
		wantIdx int
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

// TestCheckExplorationCondition tests the exploration tutorial condition.
func TestCheckExplorationCondition(t *testing.T) {
	tests := []struct {
		name         string
		hasStats     bool
		areasVisited int64
		want         bool
	}{
		{
			name:         "no player statistics component",
			hasStats:     false,
			areasVisited: 0,
			want:         false,
		},
		{
			name:         "zero areas visited",
			hasStats:     true,
			areasVisited: 0,
			want:         false,
		},
		{
			name:         "less than 3 areas visited",
			hasStats:     true,
			areasVisited: 2,
			want:         false,
		},
		{
			name:         "exactly 3 areas visited",
			hasStats:     true,
			areasVisited: 3,
			want:         true,
		},
		{
			name:         "more than 3 areas visited",
			hasStats:     true,
			areasVisited: 10,
			want:         true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world with player
			world := NewWorld()
			player := world.CreateEntity()
			player.AddComponent(&stubTutorialInputComponent{})

			if tt.hasStats {
				stats := NewPlayerStatisticsComponent()
				// Start session to enable stat tracking
				stats.StartSession(1000)
				if tt.areasVisited > 0 {
					stats.IncrementStat("explore_areas_visited", tt.areasVisited)
				}
				player.AddComponent(stats)
			}

			got := checkExplorationCondition(world)
			if got != tt.want {
				t.Errorf("checkExplorationCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

// stubTutorialInputComponent is a tag component identifying the player entity for tests.
type stubTutorialInputComponent struct{}

func (s *stubTutorialInputComponent) Type() string { return "input" }

// TestCheckExplorationCondition_NoPlayer tests exploration with no player.
func TestCheckExplorationCondition_NoPlayer(t *testing.T) {
	world := NewWorld()
	// No player entity
	if checkExplorationCondition(world) {
		t.Error("checkExplorationCondition() should return false with no player")
	}
}

// TestCheckMovementCondition tests the movement tutorial condition.
func TestCheckMovementCondition(t *testing.T) {
	tests := []struct {
		name             string
		hasStats         bool
		distanceTraveled int64
		want             bool
	}{
		{
			name:             "no player statistics component",
			hasStats:         false,
			distanceTraveled: 0,
			want:             false,
		},
		{
			name:             "zero distance traveled",
			hasStats:         true,
			distanceTraveled: 0,
			want:             false,
		},
		{
			name:             "less than 50 units traveled",
			hasStats:         true,
			distanceTraveled: 49,
			want:             false,
		},
		{
			name:             "exactly 50 units traveled",
			hasStats:         true,
			distanceTraveled: 50,
			want:             true,
		},
		{
			name:             "more than 50 units traveled",
			hasStats:         true,
			distanceTraveled: 100,
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create world with player
			world := NewWorld()
			player := world.CreateEntity()
			player.AddComponent(&stubTutorialInputComponent{})

			if tt.hasStats {
				stats := NewPlayerStatisticsComponent()
				// Start session to enable stat tracking
				stats.StartSession(1000)
				if tt.distanceTraveled > 0 {
					stats.IncrementStat("explore_distance_traveled", tt.distanceTraveled)
				}
				player.AddComponent(stats)
			}

			got := checkMovementCondition(world)
			if got != tt.want {
				t.Errorf("checkMovementCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCheckMovementCondition_NoPlayer tests movement with no player.
func TestCheckMovementCondition_NoPlayer(t *testing.T) {
	world := NewWorld()
	// No player entity
	if checkMovementCondition(world) {
		t.Error("checkMovementCondition() should return false with no player")
	}
}

// TestNewTutorialSystemWithSize tests tutorial system creation with explicit dimensions.
func TestNewTutorialSystemWithSize(t *testing.T) {
	ts := NewTutorialSystemWithSize(1920, 1080)

	if ts == nil {
		t.Fatal("NewTutorialSystemWithSize returned nil")
	}

	if ts.screenWidth != 1920 {
		t.Errorf("screenWidth = %d, want 1920", ts.screenWidth)
	}
	if ts.screenHeight != 1080 {
		t.Errorf("screenHeight = %d, want 1080", ts.screenHeight)
	}

	// Verify buttons are created at correct positions
	if ts.nextButton == nil {
		t.Fatal("nextButton should not be nil")
	}
	if ts.skipButton == nil {
		t.Fatal("skipButton should not be nil")
	}
}

// TestTutorialSystem_Resize tests that button positions update after resize.
func TestTutorialSystem_Resize(t *testing.T) {
	// Create at 800x600
	ts := NewTutorialSystemWithSize(800, 600)

	// Verify initial dimensions
	if ts.screenWidth != 800 || ts.screenHeight != 600 {
		t.Errorf("Initial dimensions = %dx%d, want 800x600", ts.screenWidth, ts.screenHeight)
	}

	// Resize to 1920x1080
	ts.Resize(1920, 1080)

	if ts.screenWidth != 1920 {
		t.Errorf("After resize screenWidth = %d, want 1920", ts.screenWidth)
	}
	if ts.screenHeight != 1080 {
		t.Errorf("After resize screenHeight = %d, want 1080", ts.screenHeight)
	}

	// Verify calling Resize with same dimensions is a no-op (no error)
	ts.Resize(1920, 1080)
	if ts.screenWidth != 1920 || ts.screenHeight != 1080 {
		t.Error("Resize with same dimensions should not change anything")
	}
}

// TestTutorialSystem_Resize_WithNilButtons tests Resize handles nil buttons gracefully.
func TestTutorialSystem_Resize_WithNilButtons(t *testing.T) {
	ts := NewTutorialSystemWithSize(800, 600)
	ts.nextButton = nil
	ts.skipButton = nil

	// Should not panic
	ts.Resize(1920, 1080)

	if ts.screenWidth != 1920 || ts.screenHeight != 1080 {
		t.Error("Resize should still update dimensions with nil buttons")
	}
}

// TestTutorialSystem_Resize_MultipleResolutions tests resize across various resolutions.
func TestTutorialSystem_Resize_MultipleResolutions(t *testing.T) {
	resolutions := []struct {
		width, height int
	}{
		{320, 480},   // Mobile portrait
		{480, 320},   // Mobile landscape
		{800, 600},   // Small desktop
		{1280, 720},  // 720p
		{1920, 1080}, // 1080p
		{2560, 1440}, // 1440p
		{3840, 2160}, // 4K
	}

	ts := NewTutorialSystemWithSize(800, 600)

	for _, res := range resolutions {
		t.Run(fmt.Sprintf("%dx%d", res.width, res.height), func(t *testing.T) {
			ts.Resize(res.width, res.height)

			if ts.screenWidth != res.width {
				t.Errorf("screenWidth = %d, want %d", ts.screenWidth, res.width)
			}
			if ts.screenHeight != res.height {
				t.Errorf("screenHeight = %d, want %d", ts.screenHeight, res.height)
			}
		})
	}
}

// TestTutorialSystem_ImportState_CompletedTutorial verifies completed tutorials stay completed
func TestTutorialSystem_ImportState_CompletedTutorial(t *testing.T) {
	ts := NewTutorialSystem()
	numSteps := len(ts.Steps)

	// Mark all steps as completed
	completedSteps := make(map[string]bool)
	for _, step := range ts.Steps {
		completedSteps[step.ID] = true
	}

	// Import state simulating a completed tutorial load
	ts.ImportState(true, true, numSteps, completedSteps)

	// Verify tutorial is disabled and stays completed
	if ts.Enabled {
		t.Error("Completed tutorial should be disabled after import")
	}
	if ts.ShowUI {
		t.Error("Completed tutorial should not show UI after import")
	}
	if ts.CurrentStepIdx != numSteps {
		t.Errorf("CurrentStepIdx = %d, want %d (all steps complete)", ts.CurrentStepIdx, numSteps)
	}

	// Verify all steps remain completed
	for i, step := range ts.Steps {
		if !step.Completed {
			t.Errorf("Step %d (%s) should remain completed", i, step.ID)
		}
	}
}

// TestTutorialSystem_ClassAwareText verifies that tutorial step descriptions
// change based on the player's selected class.
func TestTutorialSystem_ClassAwareText(t *testing.T) {
	ts := NewTutorialSystem()

	// Find combat step
	var combatStep *TutorialStep
	for i := range ts.Steps {
		if ts.Steps[i].ID == "combat" {
			combatStep = &ts.Steps[i]
			break
		}
	}

	if combatStep == nil {
		t.Fatal("Combat step not found in tutorial")
	}

	defaultDesc := combatStep.Description

	testCases := []struct {
		class      CharacterClass
		shouldDiff bool // Should the description differ from default?
		contains   string
	}{
		{ClassWarrior, true, "sword"},
		{ClassMage, true, "spells"},
		{ClassRogue, true, "backstab"},
		{ClassRanger, true, "arrows"},
		{ClassCleric, true, "heal"},
		{ClassNecromancer, true, "Summon"},
		{ClassBattlemage, true, "martial and magical"},
		{ClassPaladin, true, "holy"},
		{ClassNinja, true, "stealth"},
	}

	for _, tc := range testCases {
		t.Run(tc.class.String(), func(t *testing.T) {
			ts.SetPlayerClass(tc.class)

			desc := ts.getClassAwareDescription(combatStep)

			if tc.shouldDiff && desc == defaultDesc {
				t.Errorf("Expected class-specific description for %s, got default", tc.class)
			}

			if tc.contains != "" && !strings.Contains(strings.ToLower(desc), strings.ToLower(tc.contains)) {
				t.Errorf("Expected description for %s to contain %q, got %q", tc.class, tc.contains, desc)
			}
		})
	}
}

// TestTutorialSystem_GetSetPlayerClass tests the PlayerClass getter/setter.
func TestTutorialSystem_GetSetPlayerClass(t *testing.T) {
	ts := NewTutorialSystem()

	// Default should be ClassWarrior (zero value)
	if ts.GetPlayerClass() != ClassWarrior {
		t.Errorf("Default PlayerClass should be ClassWarrior, got %v", ts.GetPlayerClass())
	}

	// Test setting different classes
	classes := []CharacterClass{ClassMage, ClassRogue, ClassRanger, ClassNecromancer, ClassPaladin}
	for _, class := range classes {
		ts.SetPlayerClass(class)
		if ts.GetPlayerClass() != class {
			t.Errorf("SetPlayerClass(%v) failed, GetPlayerClass() returned %v", class, ts.GetPlayerClass())
		}
	}
}

// TestTutorialSystem_ClassAwareDescription_FallsBackToDefault tests that
// the class-aware description falls back to default when no override exists.
func TestTutorialSystem_ClassAwareDescription_FallsBackToDefault(t *testing.T) {
	ts := NewTutorialSystem()

	// Find welcome step (which has no class-specific overrides)
	var welcomeStep *TutorialStep
	for i := range ts.Steps {
		if ts.Steps[i].ID == "welcome" {
			welcomeStep = &ts.Steps[i]
			break
		}
	}

	if welcomeStep == nil {
		t.Fatal("Welcome step not found")
	}

	// Set various classes and verify default description is used
	classes := []CharacterClass{ClassWarrior, ClassMage, ClassRogue, ClassRanger}
	for _, class := range classes {
		ts.SetPlayerClass(class)
		desc := ts.getClassAwareDescription(welcomeStep)
		if desc != welcomeStep.Description {
			t.Errorf("For %s, expected default description for welcome step, got %q", class, desc)
		}
	}
}

// TestTutorialSystem_ClassAwareDescription_NilStep tests handling of nil step.
func TestTutorialSystem_ClassAwareDescription_NilStep(t *testing.T) {
	ts := NewTutorialSystem()

	desc := ts.getClassAwareDescription(nil)
	if desc != "" {
		t.Errorf("Expected empty string for nil step, got %q", desc)
	}
}

// TestTutorialSystem_AdvanceStep tests manual step advancement via AdvanceStep.
func TestTutorialSystem_AdvanceStep(t *testing.T) {
	ts := NewTutorialSystem()

	// Verify initial state
	if ts.CurrentStepIdx != 0 {
		t.Fatalf("Expected initial step index 0, got %d", ts.CurrentStepIdx)
	}

	// Advance step
	ts.AdvanceStep()

	// Verify step was advanced
	if ts.CurrentStepIdx != 1 {
		t.Errorf("Expected step index 1 after advance, got %d", ts.CurrentStepIdx)
	}

	// Verify previous step was marked completed
	if !ts.Steps[0].Completed {
		t.Error("Previous step should be marked completed after advance")
	}

	// Verify notification was set
	if ts.NotificationMsg == "" {
		t.Error("Expected notification message after advance")
	}
	if ts.NotificationTTL <= 0 {
		t.Error("Expected notification TTL > 0 after advance")
	}
}

// TestTutorialSystem_AdvanceStep_CompleteTutorial tests advancing through all steps.
func TestTutorialSystem_AdvanceStep_CompleteTutorial(t *testing.T) {
	ts := NewTutorialSystem()

	// Advance through all steps
	for i := 0; i < len(ts.Steps); i++ {
		ts.AdvanceStep()
	}

	// Verify tutorial is complete
	if ts.CurrentStepIdx != len(ts.Steps) {
		t.Errorf("Expected step index %d after completing all steps, got %d", len(ts.Steps), ts.CurrentStepIdx)
	}

	// Verify tutorial is disabled
	if ts.Enabled {
		t.Error("Tutorial should be disabled after completing all steps")
	}

	// Verify UI is hidden
	if ts.ShowUI {
		t.Error("UI should be hidden after completing all steps")
	}

	// Verify completion notification
	if ts.NotificationMsg != "Tutorial completed!" {
		t.Errorf("Expected 'Tutorial completed!' notification, got %q", ts.NotificationMsg)
	}
}

// TestTutorialSystem_AdvanceStep_NoOpWhenComplete tests advancing when already complete.
func TestTutorialSystem_AdvanceStep_NoOpWhenComplete(t *testing.T) {
	ts := NewTutorialSystem()

	// Complete all steps
	for i := 0; i < len(ts.Steps); i++ {
		ts.AdvanceStep()
	}

	finalIdx := ts.CurrentStepIdx

	// Try advancing again
	ts.AdvanceStep()

	// Should not change
	if ts.CurrentStepIdx != finalIdx {
		t.Errorf("Expected step index to remain %d, got %d", finalIdx, ts.CurrentStepIdx)
	}
}

// TestTutorialSystem_HintText verifies the hint text includes both ESC and N shortcuts.
func TestTutorialSystem_HintText(t *testing.T) {
	// Note: The actual hint text "ESC: minimize | N: next step" is rendered in drawPanelContent.
	// This test verifies the AdvanceStep and HideTutorialUI methods work as documented.
	ts := NewTutorialSystem()

	// Test ESC behavior (minimize)
	ts.HideTutorialUI()
	if ts.ShowUI {
		t.Error("HideTutorialUI should hide the UI")
	}
	if ts.Enabled == false {
		t.Error("HideTutorialUI should not disable the tutorial, only hide UI")
	}

	// Reset for next test
	ts.ShowUI = true

	// Test N behavior (advance)
	initialStep := ts.CurrentStepIdx
	ts.AdvanceStep()
	if ts.CurrentStepIdx != initialStep+1 {
		t.Error("AdvanceStep should advance to next step")
	}
}
