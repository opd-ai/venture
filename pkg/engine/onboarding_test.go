package engine

import (
	"testing"
)

// TestOnboardingManager_Creation tests basic OnboardingManager creation.
func TestOnboardingManager_Creation(t *testing.T) {
	om := NewOnboardingManager(nil)
	if om == nil {
		t.Fatal("NewOnboardingManager returned nil")
	}

	if om.GetState() != StateCharacterCreation {
		t.Errorf("expected initial state StateCharacterCreation, got %s", om.GetState())
	}

	if !om.IsEnabled() {
		t.Error("expected OnboardingManager to be enabled by default")
	}

	if om.IsComplete() {
		t.Error("expected OnboardingManager to not be complete initially")
	}
}

// TestOnboardingState_String tests OnboardingState string representations.
func TestOnboardingState_String(t *testing.T) {
	tests := []struct {
		state    OnboardingState
		expected string
	}{
		{StateCharacterCreation, "CharacterCreation"},
		{StateInGameTutorial, "InGameTutorial"},
		{StateContextHelp, "ContextHelp"},
		{StateComplete, "Complete"},
		{OnboardingState(99), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.state.String(); got != tt.expected {
				t.Errorf("OnboardingState(%d).String() = %s, want %s", tt.state, got, tt.expected)
			}
		})
	}
}

// TestOnboardingManager_StateTransitions tests the canonical state transition flow.
func TestOnboardingManager_StateTransitions(t *testing.T) {
	om := NewOnboardingManager(nil)

	// Start in character creation state
	if om.GetState() != StateCharacterCreation {
		t.Errorf("expected StateCharacterCreation, got %s", om.GetState())
	}

	// Transition to in-game tutorial
	om.TransitionToInGameTutorial()
	if om.GetState() != StateInGameTutorial {
		t.Errorf("expected StateInGameTutorial, got %s", om.GetState())
	}

	// Transition to context help
	om.TransitionToContextHelp()
	if om.GetState() != StateContextHelp {
		t.Errorf("expected StateContextHelp, got %s", om.GetState())
	}

	// Complete onboarding
	om.Complete()
	if om.GetState() != StateComplete {
		t.Errorf("expected StateComplete, got %s", om.GetState())
	}

	if !om.IsComplete() {
		t.Error("expected IsComplete() to return true")
	}
}

// TestOnboardingManager_SkipAll tests skipping the entire onboarding flow.
func TestOnboardingManager_SkipAll(t *testing.T) {
	om := NewOnboardingManager(nil)

	// Create mock tutorial systems
	tutorialSystem := &EbitenTutorialSystem{Enabled: true, ShowUI: true}
	creationTutorial := NewCharacterCreationTutorial()

	om.SetTutorialSystem(tutorialSystem)
	om.SetCreationTutorial(creationTutorial)

	om.SkipAll()

	if !om.IsComplete() {
		t.Error("expected IsComplete() after SkipAll()")
	}

	if om.GetState() != StateComplete {
		t.Errorf("expected StateComplete after SkipAll(), got %s", om.GetState())
	}

	// Tutorial systems should be disabled
	if tutorialSystem.Enabled {
		t.Error("expected tutorialSystem.Enabled = false after SkipAll()")
	}
}

// TestOnboardingManager_SetEnabled tests enabling/disabling onboarding.
func TestOnboardingManager_SetEnabled(t *testing.T) {
	om := NewOnboardingManager(nil)

	tutorialSystem := &EbitenTutorialSystem{Enabled: true, ShowUI: true}
	creationTutorial := NewCharacterCreationTutorial()

	om.SetTutorialSystem(tutorialSystem)
	om.SetCreationTutorial(creationTutorial)

	// Disable onboarding
	om.SetEnabled(false)

	if om.IsEnabled() {
		t.Error("expected IsEnabled() = false after SetEnabled(false)")
	}

	// Connected systems should be disabled
	if tutorialSystem.Enabled {
		t.Error("expected tutorialSystem.Enabled = false when onboarding disabled")
	}
	if creationTutorial.Enabled {
		t.Error("expected creationTutorial.Enabled = false when onboarding disabled")
	}
}

// TestOnboardingManager_PlayerClass tests storing and retrieving player class.
func TestOnboardingManager_PlayerClass(t *testing.T) {
	om := NewOnboardingManager(nil)

	om.SetPlayerClass(ClassMage)
	if got := om.GetPlayerClass(); got != ClassMage {
		t.Errorf("expected ClassMage, got %s", got.String())
	}

	om.SetPlayerClass(ClassNecromancer)
	if got := om.GetPlayerClass(); got != ClassNecromancer {
		t.Errorf("expected ClassNecromancer, got %s", got.String())
	}
}

// TestOnboardingManager_ExportImportState tests save/load round-trip.
func TestOnboardingManager_ExportImportState(t *testing.T) {
	om := NewOnboardingManager(nil)

	// Set up some state
	om.TransitionToInGameTutorial()
	om.SetPlayerClass(ClassRanger)

	// Export
	exported := om.ExportState()

	// Create new manager and import
	om2 := NewOnboardingManager(nil)
	om2.ImportState(exported)

	if om2.GetState() != StateInGameTutorial {
		t.Errorf("expected StateInGameTutorial after import, got %s", om2.GetState())
	}

	if om2.GetPlayerClass() != ClassRanger {
		t.Errorf("expected ClassRanger after import, got %s", om2.GetPlayerClass().String())
	}
}

// TestOnboardingManager_Callbacks tests state change and completion callbacks.
func TestOnboardingManager_Callbacks(t *testing.T) {
	om := NewOnboardingManager(nil)

	var stateChanges []OnboardingState
	var completionCalled bool

	om.SetStateChangeCallback(func(state OnboardingState) {
		stateChanges = append(stateChanges, state)
	})

	om.SetCompletionCallback(func(state OnboardingState) {
		completionCalled = true
	})

	// Transition through states
	om.TransitionToInGameTutorial()
	om.TransitionToContextHelp()
	om.Complete()

	// Verify callbacks were called
	if len(stateChanges) != 3 {
		t.Errorf("expected 3 state changes, got %d", len(stateChanges))
	}

	if !completionCalled {
		t.Error("expected completion callback to be called")
	}

	expectedStates := []OnboardingState{StateInGameTutorial, StateContextHelp, StateComplete}
	for i, expected := range expectedStates {
		if i < len(stateChanges) && stateChanges[i] != expected {
			t.Errorf("state change %d: expected %s, got %s", i, expected, stateChanges[i])
		}
	}
}

// TestOnboardingManager_TransitionWhenDisabled tests that transitions are no-ops when disabled.
func TestOnboardingManager_TransitionWhenDisabled(t *testing.T) {
	om := NewOnboardingManager(nil)
	om.SetEnabled(false)

	initialState := om.GetState()

	// These should be no-ops
	om.TransitionToInGameTutorial()
	om.TransitionToContextHelp()

	if om.GetState() != initialState {
		t.Errorf("state should not change when disabled: got %s", om.GetState())
	}
}

// TestOnboardingManager_TransitionWhenSkipped tests that transitions are no-ops when skipped.
func TestOnboardingManager_TransitionWhenSkipped(t *testing.T) {
	om := NewOnboardingManager(nil)
	om.SkipAll()

	// These should be no-ops
	om.TransitionToInGameTutorial()
	om.TransitionToContextHelp()

	// Should still be complete
	if om.GetState() != StateComplete {
		t.Errorf("state should remain Complete after skip: got %s", om.GetState())
	}
}

// TestOnboardingManager_CompleteIdempotent tests that Complete() can be called multiple times.
func TestOnboardingManager_CompleteIdempotent(t *testing.T) {
	om := NewOnboardingManager(nil)

	callCount := 0
	om.SetCompletionCallback(func(state OnboardingState) {
		callCount++
	})

	om.Complete()
	om.Complete()
	om.Complete()

	if callCount != 1 {
		t.Errorf("expected completion callback to be called once, called %d times", callCount)
	}
}

// TestOnboardingManager_TutorialSystemActivation tests that tutorial system is activated on transition.
func TestOnboardingManager_TutorialSystemActivation(t *testing.T) {
	om := NewOnboardingManager(nil)

	tutorialSystem := &EbitenTutorialSystem{Enabled: false, ShowUI: false}
	om.SetTutorialSystem(tutorialSystem)

	om.TransitionToInGameTutorial()

	if !tutorialSystem.Enabled {
		t.Error("expected tutorialSystem.Enabled = true after TransitionToInGameTutorial()")
	}
	if !tutorialSystem.ShowUI {
		t.Error("expected tutorialSystem.ShowUI = true after TransitionToInGameTutorial()")
	}
}
