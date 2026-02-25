package engine

import (
	"testing"
)

// TestNewCharacterCreationTutorial tests tutorial creation with default steps.
func TestNewCharacterCreationTutorial(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	if cct == nil {
		t.Fatal("NewCharacterCreationTutorial returned nil")
	}
	if !cct.Enabled {
		t.Error("Tutorial should be enabled by default")
	}
	if !cct.ShowUI {
		t.Error("ShowUI should be true by default")
	}
	if cct.CurrentStepIdx != 0 {
		t.Errorf("CurrentStepIdx = %d, want 0", cct.CurrentStepIdx)
	}
	if cct.Completed {
		t.Error("Tutorial should not be completed initially")
	}
	if cct.Skipped {
		t.Error("Tutorial should not be skipped initially")
	}

	expectedSteps := []string{"welcome_creation", "name_input", "class_selection", "subclass_selection", "portrait_selection", "confirmation"}
	if len(cct.Steps) != len(expectedSteps) {
		t.Fatalf("Expected %d steps, got %d", len(expectedSteps), len(cct.Steps))
	}
	for i, id := range expectedSteps {
		if cct.Steps[i].ID != id {
			t.Errorf("Step %d: ID = %s, want %s", i, cct.Steps[i].ID, id)
		}
		if cct.Steps[i].Completed {
			t.Errorf("Step %d should not be completed initially", i)
		}
		if cct.Steps[i].Title == "" {
			t.Errorf("Step %d has empty title", i)
		}
		if cct.Steps[i].Description == "" {
			t.Errorf("Step %d has empty description", i)
		}
		if cct.Steps[i].Hint == "" {
			t.Errorf("Step %d has empty hint", i)
		}
	}
}

// TestCharacterCreationTutorial_GetCurrentStep tests current step retrieval.
func TestCharacterCreationTutorial_GetCurrentStep(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	step := cct.GetCurrentStep()
	if step == nil {
		t.Fatal("GetCurrentStep returned nil at start")
	}
	if step.ID != "welcome_creation" {
		t.Errorf("First step ID = %s, want 'welcome_creation'", step.ID)
	}

	// Advance
	cct.CurrentStepIdx = 2
	step = cct.GetCurrentStep()
	if step == nil || step.ID != "class_selection" {
		t.Errorf("Step at index 2 should be 'class_selection'")
	}

	// Beyond end
	cct.CurrentStepIdx = len(cct.Steps)
	step = cct.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil after all steps")
	}

	// Disabled
	cct.CurrentStepIdx = 0
	cct.Enabled = false
	step = cct.GetCurrentStep()
	if step != nil {
		t.Error("GetCurrentStep should return nil when disabled")
	}
}

// TestCharacterCreationTutorial_GetProgress tests progress calculation.
func TestCharacterCreationTutorial_GetProgress(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	tests := []struct {
		name    string
		stepIdx int
		wantMin float64
		wantMax float64
	}{
		{"Start", 0, 0.0, 0.01},
		{"Mid", 3, 0.49, 0.51},      // 3/6 = 0.5
		{"Near end", 5, 0.82, 0.84}, // 5/6 ≈ 0.833
		{"Complete", 6, 1.0, 1.0},   // 6/6 = 1.0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cct.CurrentStepIdx = tt.stepIdx
			progress := cct.GetProgress()
			if progress < tt.wantMin || progress > tt.wantMax {
				t.Errorf("GetProgress() = %v, want between %v and %v", progress, tt.wantMin, tt.wantMax)
			}
		})
	}
}

// TestCharacterCreationTutorial_GetProgress_EmptySteps tests progress with no steps.
func TestCharacterCreationTutorial_GetProgress_EmptySteps(t *testing.T) {
	cct := &CharacterCreationTutorial{
		Steps: []CharacterCreationTutorialStep{},
	}
	if cct.GetProgress() != 1.0 {
		t.Errorf("GetProgress() with empty steps = %v, want 1.0", cct.GetProgress())
	}
}

// TestCharacterCreationTutorial_SkipTutorial tests skip functionality.
func TestCharacterCreationTutorial_SkipTutorial(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	cct.SkipTutorial()

	if !cct.Skipped {
		t.Error("Tutorial should be marked as skipped")
	}
	if !cct.Completed {
		t.Error("Tutorial should be marked as completed after skip")
	}
	if cct.Enabled {
		t.Error("Tutorial should be disabled after skip")
	}
	if cct.ShowUI {
		t.Error("ShowUI should be false after skip")
	}
	if cct.NotificationMsg == "" {
		t.Error("Skip notification should be set")
	}
}

// TestCharacterCreationTutorial_SetEnabled tests the SetEnabled method.
// This validates the fix for Task 3.3 from PLAN.md (ShowTutorials Setting Not Wired).
func TestCharacterCreationTutorial_SetEnabled(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Verify initial state
	if !cct.Enabled {
		t.Error("Expected tutorial to be enabled by default")
	}
	if !cct.ShowUI {
		t.Error("Expected ShowUI to be true by default")
	}

	// Disable via SetEnabled
	cct.SetEnabled(false)
	if cct.Enabled {
		t.Error("Expected tutorial to be disabled after SetEnabled(false)")
	}
	if cct.ShowUI {
		t.Error("Expected ShowUI to be false after SetEnabled(false)")
	}

	// Re-enable via SetEnabled
	cct.SetEnabled(true)
	if !cct.Enabled {
		t.Error("Expected tutorial to be re-enabled after SetEnabled(true)")
	}
	if !cct.ShowUI {
		t.Error("Expected ShowUI to be true after SetEnabled(true)")
	}
}

// TestCharacterCreationTutorial_Reset tests resetting the tutorial.
func TestCharacterCreationTutorial_Reset(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Modify state
	cct.Enabled = false
	cct.ShowUI = false
	cct.CurrentStepIdx = 3
	cct.Completed = true
	cct.Skipped = true
	cct.NotificationMsg = "test"
	cct.NotificationTTL = 5.0
	for i := range cct.Steps {
		cct.Steps[i].Completed = true
	}

	cct.Reset()

	if !cct.Enabled {
		t.Error("Enabled should be true after Reset")
	}
	if !cct.ShowUI {
		t.Error("ShowUI should be true after Reset")
	}
	if cct.CurrentStepIdx != 0 {
		t.Errorf("CurrentStepIdx = %d, want 0 after Reset", cct.CurrentStepIdx)
	}
	if cct.Completed {
		t.Error("Completed should be false after Reset")
	}
	if cct.Skipped {
		t.Error("Skipped should be false after Reset")
	}
	if cct.NotificationMsg != "" {
		t.Error("NotificationMsg should be empty after Reset")
	}
	for i, step := range cct.Steps {
		if step.Completed {
			t.Errorf("Step %d should not be completed after Reset", i)
		}
	}
}

// TestCharacterCreationTutorial_ExportImportState tests state serialization.
func TestCharacterCreationTutorial_ExportImportState(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Advance through some steps
	cct.Steps[0].Completed = true
	cct.Steps[1].Completed = true
	cct.CurrentStepIdx = 2
	cct.Completed = false

	completed, skipped, completedSteps := cct.ExportState()

	if completed {
		t.Error("Exported completed should be false")
	}
	if skipped {
		t.Error("Exported skipped should be false")
	}
	if len(completedSteps) != 2 {
		t.Errorf("Expected 2 completed steps, got %d", len(completedSteps))
	}
	if !completedSteps["welcome_creation"] {
		t.Error("'welcome_creation' should be in completed steps")
	}
	if !completedSteps["name_input"] {
		t.Error("'name_input' should be in completed steps")
	}

	// Import into fresh tutorial
	cct2 := NewCharacterCreationTutorial()
	cct2.ImportState(completed, skipped, completedSteps)

	if cct2.CurrentStepIdx != 2 {
		t.Errorf("After import, CurrentStepIdx = %d, want 2", cct2.CurrentStepIdx)
	}
	if !cct2.Steps[0].Completed {
		t.Error("Step 0 should be completed after import")
	}
	if !cct2.Steps[1].Completed {
		t.Error("Step 1 should be completed after import")
	}
	if cct2.Steps[2].Completed {
		t.Error("Step 2 should not be completed after import")
	}
}

// TestCharacterCreationTutorial_ImportState_Completed tests importing a completed tutorial.
func TestCharacterCreationTutorial_ImportState_Completed(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	cct.ImportState(true, false, nil)

	if !cct.Completed {
		t.Error("Tutorial should be completed after import")
	}
	if cct.Enabled {
		t.Error("Tutorial should be disabled after importing completed state")
	}
	if cct.ShowUI {
		t.Error("ShowUI should be false after importing completed state")
	}
}

// TestCharacterCreationTutorial_ImportState_Skipped tests importing a skipped tutorial.
func TestCharacterCreationTutorial_ImportState_Skipped(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	cct.ImportState(true, true, nil)

	if !cct.Skipped {
		t.Error("Tutorial should be marked skipped after import")
	}
	if cct.Enabled {
		t.Error("Tutorial should be disabled after importing skipped state")
	}
}

// TestCharacterCreationTutorial_Update_WhenDisabled tests no-op when disabled.
func TestCharacterCreationTutorial_Update_WhenDisabled(t *testing.T) {
	cct := NewCharacterCreationTutorial()
	cct.Enabled = false

	initialIdx := cct.CurrentStepIdx
	cct.Update(0, 0.016)

	if cct.CurrentStepIdx != initialIdx {
		t.Error("Update should not change state when disabled")
	}
}

// TestCharacterCreationTutorial_Update_WhenCompleted tests no-op when completed.
func TestCharacterCreationTutorial_Update_WhenCompleted(t *testing.T) {
	cct := NewCharacterCreationTutorial()
	cct.Completed = true

	initialIdx := cct.CurrentStepIdx
	cct.Update(0, 0.016)

	if cct.CurrentStepIdx != initialIdx {
		t.Error("Update should not change state when completed")
	}
}

// TestCharacterCreationTutorial_Update_StepSync tests that tutorial steps
// advance to match character creation progress.
func TestCharacterCreationTutorial_Update_StepSync(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Start past the welcome step
	cct.CurrentStepIdx = 1
	cct.Steps[0].Completed = true

	// Simulate character creation advancing to class selection (step 1)
	cct.Update(1, 0.016)

	// Tutorial should advance to class_selection step (index 2)
	if cct.CurrentStepIdx != 2 {
		t.Errorf("CurrentStepIdx = %d, want 2 after creation step 1", cct.CurrentStepIdx)
	}
	if !cct.Steps[1].Completed {
		t.Error("name_input step should be completed")
	}
}

// TestCharacterCreationTutorial_Update_NotificationTTL tests notification timeout.
func TestCharacterCreationTutorial_Update_NotificationTTL(t *testing.T) {
	cct := NewCharacterCreationTutorial()
	cct.NotificationMsg = "Test notification"
	cct.NotificationTTL = 0.1

	// Update multiple times to expire
	for i := 0; i < 10; i++ {
		cct.Update(0, 0.016)
	}

	if cct.NotificationMsg != "" {
		t.Error("Notification should be cleared after TTL expires")
	}
}

// TestCharacterCreationTutorial_CompleteTutorial tests completing all steps.
func TestCharacterCreationTutorial_CompleteTutorial(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Skip welcome step manually
	cct.CurrentStepIdx = 1
	cct.Steps[0].Completed = true

	// Simulate progressing through all creation steps (0=name, 1=class, 2=subclass, 3=portrait, 4=confirmation)
	// Tutorial steps are: 0=welcome, 1=name, 2=class, 3=subclass, 4=portrait, 5=confirmation
	for step := 1; step <= 4; step++ {
		cct.Update(step, 0.016)
	}

	// After stepping through creation steps 1-4, the tutorial should be at the
	// confirmation step (index 5). The confirmation step is completed externally
	// when character creation is confirmed (via CompleteTutorial in the game loop).
	if cct.CurrentStepIdx != 5 {
		t.Fatalf("expected CurrentStepIdx=5 before final completion, got %d", cct.CurrentStepIdx)
	}

	// Simulate the game loop marking the tutorial as complete on character creation confirm
	cct.CompleteTutorial()

	if !cct.Completed {
		t.Fatalf("tutorial should be completed: Completed=false, CurrentStepIdx=%d, len(Steps)=%d",
			cct.CurrentStepIdx, len(cct.Steps))
	}
	if !cct.Steps[5].Completed {
		t.Error("confirmation step should not be left incomplete after CompleteTutorial")
	}
}

// TestCharacterCreationTutorial_IsActive tests active state check.
func TestCharacterCreationTutorial_IsActive(t *testing.T) {
	tests := []struct {
		name      string
		enabled   bool
		showUI    bool
		completed bool
		want      bool
	}{
		{"Active", true, true, false, true},
		{"Disabled", false, true, false, false},
		{"UI hidden", true, false, false, false},
		{"Completed", true, true, true, false},
		{"All off", false, false, true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cct := NewCharacterCreationTutorial()
			cct.Enabled = tt.enabled
			cct.ShowUI = tt.showUI
			cct.Completed = tt.completed

			if got := cct.IsActive(); got != tt.want {
				t.Errorf("IsActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestCharacterCreationTutorial_GetStepByID tests step retrieval by ID.
func TestCharacterCreationTutorial_GetStepByID(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	step := cct.GetStepByID("class_selection")
	if step == nil {
		t.Fatal("GetStepByID('class_selection') returned nil")
	}
	if step.ID != "class_selection" {
		t.Errorf("Step ID = %s, want 'class_selection'", step.ID)
	}

	step = cct.GetStepByID("nonexistent")
	if step != nil {
		t.Error("GetStepByID('nonexistent') should return nil")
	}
}

// TestCharacterCreationTutorial_IsStepCompleted tests step completion check.
func TestCharacterCreationTutorial_IsStepCompleted(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	if cct.IsStepCompleted("welcome_creation") {
		t.Error("Step should not be completed initially")
	}

	cct.Steps[0].Completed = true
	if !cct.IsStepCompleted("welcome_creation") {
		t.Error("Step should be completed after marking")
	}

	if cct.IsStepCompleted("nonexistent") {
		t.Error("Non-existent step should return false")
	}
}

// TestTutorialCompletionComponent_Type tests the component type identifier.
func TestTutorialCompletionComponent_Type(t *testing.T) {
	tc := &TutorialCompletionComponent{}
	if tc.Type() != "tutorial_completion" {
		t.Errorf("Type() = %s, want 'tutorial_completion'", tc.Type())
	}
}

// TestShouldShowCreationTutorial tests first-time player detection.
func TestShouldShowCreationTutorial(t *testing.T) {
	tests := []struct {
		name  string
		setup func() *Entity
		want  bool
	}{
		{
			name:  "nil player",
			setup: func() *Entity { return nil },
			want:  true,
		},
		{
			name: "no completion component",
			setup: func() *Entity {
				world := NewWorld()
				return world.CreateEntity()
			},
			want: true,
		},
		{
			name: "tutorial not done",
			setup: func() *Entity {
				world := NewWorld()
				e := world.CreateEntity()
				e.AddComponent(&TutorialCompletionComponent{
					CreationTutorialDone: false,
					CreationTutorialSkip: false,
				})
				return e
			},
			want: true,
		},
		{
			name: "tutorial done",
			setup: func() *Entity {
				world := NewWorld()
				e := world.CreateEntity()
				e.AddComponent(&TutorialCompletionComponent{
					CreationTutorialDone: true,
					CreationTutorialSkip: false,
				})
				return e
			},
			want: false,
		},
		{
			name: "tutorial skipped",
			setup: func() *Entity {
				world := NewWorld()
				e := world.CreateEntity()
				e.AddComponent(&TutorialCompletionComponent{
					CreationTutorialDone: false,
					CreationTutorialSkip: true,
				})
				return e
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			player := tt.setup()
			if got := ShouldShowCreationTutorial(player); got != tt.want {
				t.Errorf("ShouldShowCreationTutorial() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestMarkCreationTutorialComplete tests marking tutorial as complete.
func TestMarkCreationTutorialComplete(t *testing.T) {
	tests := []struct {
		name    string
		skipped bool
	}{
		{"completed normally", false},
		{"skipped", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			world := NewWorld()
			player := world.CreateEntity()

			MarkCreationTutorialComplete(player, tt.skipped)

			comp, ok := player.GetComponent("tutorial_completion")
			if !ok {
				t.Fatal("TutorialCompletionComponent not found after marking")
			}
			tc, ok := comp.(*TutorialCompletionComponent)
			if !ok {
				t.Fatal("Component is not *TutorialCompletionComponent")
			}
			if !tc.CreationTutorialDone {
				t.Error("CreationTutorialDone should be true")
			}
			if tc.CreationTutorialSkip != tt.skipped {
				t.Errorf("CreationTutorialSkip = %v, want %v", tc.CreationTutorialSkip, tt.skipped)
			}

			// Should not show tutorial anymore
			if ShouldShowCreationTutorial(player) {
				t.Error("ShouldShowCreationTutorial should return false after marking complete")
			}
		})
	}
}

// TestMarkCreationTutorialComplete_UpdateExisting tests updating an existing component.
func TestMarkCreationTutorialComplete_UpdateExisting(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()

	// Add initial component
	player.AddComponent(&TutorialCompletionComponent{
		CreationTutorialDone: false,
		CreationTutorialSkip: false,
	})

	// Mark as complete
	MarkCreationTutorialComplete(player, false)

	comp, ok := player.GetComponent("tutorial_completion")
	if !ok {
		t.Fatal("Component not found")
	}
	tc := comp.(*TutorialCompletionComponent)
	if !tc.CreationTutorialDone {
		t.Error("CreationTutorialDone should be true after update")
	}
}

// TestMarkCreationTutorialComplete_NilPlayer tests that nil player is handled safely.
func TestMarkCreationTutorialComplete_NilPlayer(t *testing.T) {
	// Should not panic
	MarkCreationTutorialComplete(nil, false)
}

// TestCharacterCreationTutorial_Integration tests the full tutorial workflow.
func TestCharacterCreationTutorial_Integration(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Verify initial state
	if !cct.IsActive() {
		t.Error("Tutorial should be active initially")
	}
	if cct.GetProgress() != 0.0 {
		t.Errorf("Initial progress = %v, want 0.0", cct.GetProgress())
	}

	step := cct.GetCurrentStep()
	if step == nil || step.ID != "welcome_creation" {
		t.Error("Should start with 'welcome_creation'")
	}

	// Manually advance past welcome
	cct.advanceStep()
	if cct.CurrentStepIdx != 1 {
		t.Errorf("After advancing welcome, step = %d, want 1", cct.CurrentStepIdx)
	}

	// Export state
	completed, skipped, steps := cct.ExportState()
	if completed || skipped {
		t.Error("Should not be completed or skipped yet")
	}
	if !steps["welcome_creation"] {
		t.Error("welcome_creation should be in completed steps")
	}

	// Skip tutorial
	cct.SkipTutorial()
	if !cct.IsComplete() {
		t.Error("Tutorial should be complete after skip")
	}
	if cct.IsActive() {
		t.Error("Tutorial should not be active after skip")
	}

	// Reset
	cct.Reset()
	if !cct.IsActive() {
		t.Error("Tutorial should be active after reset")
	}

	// Import previously exported state
	cct.ImportState(completed, skipped, steps)
	if cct.CurrentStepIdx != 1 {
		t.Errorf("After import, step = %d, want 1", cct.CurrentStepIdx)
	}
}

// TestCharacterCreationTutorial_AllCharacterOptions tests that all character
// creation steps are represented in the tutorial.
func TestCharacterCreationTutorial_AllCharacterOptions(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Verify that all character creation steps have corresponding tutorial steps
	requiredSteps := map[string]bool{
		"name_input":         false, // stepNameInput
		"class_selection":    false, // stepClassSelection
		"subclass_selection": false, // stepSubclassSelection
		"portrait_selection": false, // stepPortraitSelection
		"confirmation":       false, // stepConfirmation
	}

	for _, step := range cct.Steps {
		if _, ok := requiredSteps[step.ID]; ok {
			requiredSteps[step.ID] = true
		}
	}

	for stepID, found := range requiredSteps {
		if !found {
			t.Errorf("Character creation step '%s' not covered by tutorial", stepID)
		}
	}
}

// TestCharacterCreationTutorial_SubclassStep tests that the subclass tutorial step
// exists and syncs correctly with the character creation flow.
func TestCharacterCreationTutorial_SubclassStep(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Verify subclass step exists at correct position (after class_selection)
	subclassStep := cct.GetStepByID("subclass_selection")
	if subclassStep == nil {
		t.Fatal("subclass_selection step not found in tutorial")
	}

	// Verify step index (should be at index 3: welcome=0, name=1, class=2, subclass=3)
	expectedIndex := 3
	for i, step := range cct.Steps {
		if step.ID == "subclass_selection" {
			if i != expectedIndex {
				t.Errorf("subclass_selection at index %d, want %d", i, expectedIndex)
			}
			break
		}
	}

	// Verify step has proper content
	if subclassStep.Title == "" {
		t.Error("subclass_selection step has empty title")
	}
	if subclassStep.Description == "" {
		t.Error("subclass_selection step has empty description")
	}
	if subclassStep.Hint == "" {
		t.Error("subclass_selection step has empty hint")
	}

	// Test synchronization: when character creation reaches subclass step (step 2),
	// tutorial should advance to subclass tutorial step (step 3)
	cct.Reset()
	cct.CurrentStepIdx = 1 // Start at name_input step
	cct.Steps[0].Completed = true

	// Simulate progression to class selection (creation step 1)
	cct.Update(1, 0.016) // creation step 1 -> tutorial step 2 (class_selection)
	if cct.CurrentStepIdx != 2 {
		t.Errorf("After class step, CurrentStepIdx = %d, want 2", cct.CurrentStepIdx)
	}

	// Simulate progression to subclass selection (creation step 2)
	cct.Update(2, 0.016) // creation step 2 -> tutorial step 3 (subclass_selection)
	if cct.CurrentStepIdx != 3 {
		t.Errorf("After subclass step, CurrentStepIdx = %d, want 3", cct.CurrentStepIdx)
	}

	// Verify subclass step is now current
	currentStep := cct.GetCurrentStep()
	if currentStep == nil || currentStep.ID != "subclass_selection" {
		t.Errorf("Current step should be subclass_selection, got %v", currentStep)
	}
}

// TestUpdateOnboardingState tests updating onboarding state on player entity.
// Phase 3.2: Verify onboarding state persistence through TutorialCompletionComponent.
func TestUpdateOnboardingState(t *testing.T) {
	world := NewWorld()
	player := world.CreateEntity()

	// Initially should return defaults
	state, class := GetOnboardingStateFromPlayer(player)
	if state != StateCharacterCreation {
		t.Errorf("Initial state = %v, want StateCharacterCreation", state)
	}
	if class != ClassWarrior {
		t.Errorf("Initial class = %v, want ClassWarrior", class)
	}

	// Update onboarding state
	UpdateOnboardingState(player, StateInGameTutorial, ClassMage)

	// Verify state was stored
	state, class = GetOnboardingStateFromPlayer(player)
	if state != StateInGameTutorial {
		t.Errorf("After update, state = %v, want StateInGameTutorial", state)
	}
	if class != ClassMage {
		t.Errorf("After update, class = %v, want ClassMage", class)
	}

	// Update again to verify modification works
	UpdateOnboardingState(player, StateComplete, ClassRogue)

	state, class = GetOnboardingStateFromPlayer(player)
	if state != StateComplete {
		t.Errorf("After second update, state = %v, want StateComplete", state)
	}
	if class != ClassRogue {
		t.Errorf("After second update, class = %v, want ClassRogue", class)
	}
}

// TestGetOnboardingStateFromPlayer_NilPlayer tests that nil player returns defaults.
func TestGetOnboardingStateFromPlayer_NilPlayer(t *testing.T) {
	state, class := GetOnboardingStateFromPlayer(nil)
	if state != StateCharacterCreation {
		t.Errorf("Nil player state = %v, want StateCharacterCreation", state)
	}
	if class != ClassWarrior {
		t.Errorf("Nil player class = %v, want ClassWarrior", class)
	}
}

// TestUpdateOnboardingState_NilPlayer tests that nil player is handled safely.
func TestUpdateOnboardingState_NilPlayer(t *testing.T) {
	// Should not panic
	UpdateOnboardingState(nil, StateComplete, ClassMage)
}

// TestTutorialCompletionComponent_OnboardingFields tests the new onboarding fields.
func TestTutorialCompletionComponent_OnboardingFields(t *testing.T) {
	tc := &TutorialCompletionComponent{
		CreationTutorialDone: true,
		CreationTutorialSkip: false,
		OnboardingState:      int(StateContextHelp),
		PlayerClass:          int(ClassRanger),
	}

	if tc.OnboardingState != int(StateContextHelp) {
		t.Errorf("OnboardingState = %d, want %d", tc.OnboardingState, int(StateContextHelp))
	}
	if tc.PlayerClass != int(ClassRanger) {
		t.Errorf("PlayerClass = %d, want %d", tc.PlayerClass, int(ClassRanger))
	}
}

// TestCharacterCreationTutorial_OnboardingTransition tests that completing
// character creation triggers the onboarding manager transition.
// Phase 3.2: Verify creation completion -> in-game tutorial transition.
func TestCharacterCreationTutorial_OnboardingTransition(t *testing.T) {
	// Create onboarding manager with tutorial systems
	om := NewOnboardingManager(nil)
	cct := NewCharacterCreationTutorial()
	ts := NewTutorialSystemWithSize(800, 600)

	// Initially disable the in-game tutorial (will be enabled on transition)
	ts.Enabled = false
	ts.ShowUI = false

	// Wire up systems
	om.SetCreationTutorial(cct)
	om.SetTutorialSystem(ts)

	// Verify initial state
	if om.GetState() != StateCharacterCreation {
		t.Errorf("Initial state = %v, want StateCharacterCreation", om.GetState())
	}
	if ts.Enabled {
		t.Error("Tutorial system should be disabled initially")
	}

	// Complete character creation tutorial
	cct.CompleteTutorial()
	if !cct.IsComplete() {
		t.Error("Character creation tutorial should be complete")
	}

	// Set player class and transition (simulating game.go handleCharacterCreation)
	om.SetPlayerClass(ClassMage)
	om.TransitionToInGameTutorial()

	// Verify transition occurred
	if om.GetState() != StateInGameTutorial {
		t.Errorf("After transition, state = %v, want StateInGameTutorial", om.GetState())
	}
	if !ts.Enabled {
		t.Error("Tutorial system should be enabled after transition")
	}
	if !ts.ShowUI {
		t.Error("Tutorial system ShowUI should be true after transition")
	}
	if om.GetPlayerClass() != ClassMage {
		t.Errorf("Player class = %v, want ClassMage", om.GetPlayerClass())
	}
}

// TestCharacterCreationTutorial_WelcomeInputConsumed tests that during the welcome
// step, InputConsumed is always true to prevent the underlying character creation
// from processing input simultaneously.
func TestCharacterCreationTutorial_WelcomeInputConsumed(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// During the welcome step, every Update should set InputConsumed = true
	cct.Update(0, 0.016)

	if !cct.InputConsumed {
		t.Error("InputConsumed should be true during welcome step")
	}

	// Step should still be 0 (welcome) since timer hasn't expired
	if cct.CurrentStepIdx != 0 {
		t.Errorf("CurrentStepIdx = %d, want 0 (welcome should not auto-advance before 3s)", cct.CurrentStepIdx)
	}
}

// TestCharacterCreationTutorial_WelcomeAutoAdvance tests that the welcome step
// auto-advances after 3 seconds.
func TestCharacterCreationTutorial_WelcomeAutoAdvance(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Simulate enough time to auto-advance (3 seconds)
	totalTime := 0.0
	for totalTime < 3.5 {
		cct.Update(0, 0.1)
		totalTime += 0.1
	}

	// Should have advanced past welcome
	if cct.CurrentStepIdx != 1 {
		t.Errorf("After 3.5s, CurrentStepIdx = %d, want 1 (should auto-advance past welcome)", cct.CurrentStepIdx)
	}
}

// TestCharacterCreationTutorial_StepOrderMatchesCreation tests that tutorial steps
// correspond 1:1 with creation steps (offset by 1 for welcome step).
func TestCharacterCreationTutorial_StepOrderMatchesCreation(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Expected mapping:
	// Tutorial index 0 = welcome_creation   (no creation step)
	// Tutorial index 1 = name_input         -> creationStep 0 (stepNameInput)
	// Tutorial index 2 = class_selection    -> creationStep 1 (stepClassSelection)
	// Tutorial index 3 = subclass_selection -> creationStep 2 (stepSubclassSelection)
	// Tutorial index 4 = portrait_selection -> creationStep 3 (stepPortraitSelection)
	// Tutorial index 5 = confirmation       -> creationStep 4 (stepConfirmation)

	expectedIDs := []string{
		"welcome_creation",
		"name_input",
		"class_selection",
		"subclass_selection",
		"portrait_selection",
		"confirmation",
	}

	if len(cct.Steps) != len(expectedIDs) {
		t.Fatalf("Expected %d steps, got %d", len(expectedIDs), len(cct.Steps))
	}

	for i, id := range expectedIDs {
		if cct.Steps[i].ID != id {
			t.Errorf("Step %d: ID = %q, want %q", i, cct.Steps[i].ID, id)
		}
	}
}

// TestCharacterCreationTutorial_FullFlowSync tests the complete synchronization
// between character creation steps and tutorial steps.
func TestCharacterCreationTutorial_FullFlowSync(t *testing.T) {
	cct := NewCharacterCreationTutorial()

	// Advance past welcome (simulate 4s)
	for i := 0; i < 40; i++ {
		cct.Update(0, 0.1)
	}
	if cct.CurrentStepIdx != 1 {
		t.Fatalf("After welcome, CurrentStepIdx = %d, want 1", cct.CurrentStepIdx)
	}

	// Step through creation steps and verify tutorial follows
	creationSteps := []struct {
		creationStep int
		wantIdx      int
		wantID       string
	}{
		{0, 1, "name_input"},         // stay on name_input
		{1, 2, "class_selection"},    // advance to class
		{2, 3, "subclass_selection"}, // advance to subclass
		{3, 4, "portrait_selection"}, // advance to portrait
		{4, 5, "confirmation"},       // advance to confirmation
	}

	for _, tc := range creationSteps {
		cct.Update(tc.creationStep, 0.016)
		if cct.CurrentStepIdx != tc.wantIdx {
			t.Errorf("creationStep=%d: CurrentStepIdx = %d, want %d",
				tc.creationStep, cct.CurrentStepIdx, tc.wantIdx)
		}
		step := cct.GetCurrentStep()
		if step != nil && step.ID != tc.wantID {
			t.Errorf("creationStep=%d: current step ID = %q, want %q",
				tc.creationStep, step.ID, tc.wantID)
		}
	}

	// Completing the tutorial should mark all steps done
	cct.CompleteTutorial()
	if !cct.Completed {
		t.Error("Tutorial should be completed")
	}
	for i, s := range cct.Steps {
		if !s.Completed {
			t.Errorf("Step %d (%s) should be completed", i, s.ID)
		}
	}
}
