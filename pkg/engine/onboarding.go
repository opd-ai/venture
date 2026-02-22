// Package engine provides onboarding flow management for new players.
// This file implements OnboardingManager which coordinates the three tutorial layers:
// CharacterCreationTutorial, EbitenTutorialSystem, and TutorialManager.
package engine

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// OnboardingState represents the current phase of the onboarding sequence.
type OnboardingState int

const (
	// StateCharacterCreation - Player is in character creation with tutorial overlay
	StateCharacterCreation OnboardingState = iota
	// StateInGameTutorial - Player has created character, in-game tutorial active
	StateInGameTutorial
	// StateContextHelp - In-game tutorial complete, context-sensitive help active
	StateContextHelp
	// StateComplete - All tutorials completed
	StateComplete
)

// String returns a human-readable name for the onboarding state.
func (s OnboardingState) String() string {
	switch s {
	case StateCharacterCreation:
		return "CharacterCreation"
	case StateInGameTutorial:
		return "InGameTutorial"
	case StateContextHelp:
		return "ContextHelp"
	case StateComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// OnboardingCompletionCallback is called when an onboarding phase completes.
type OnboardingCompletionCallback func(state OnboardingState)

// OnboardingManager coordinates transitions between tutorial layers.
// It manages the canonical onboarding sequence:
// 1. CharacterCreationTutorial overlays character creation UI
// 2. Character creation completes → player entity spawned
// 3. EbitenTutorialSystem activates for in-game guidance
// 4. TutorialManager provides context-sensitive help on first encounter
type OnboardingManager struct {
	mu               sync.RWMutex
	currentState     OnboardingState
	enabled          bool
	skipped          bool
	playerClass      CharacterClass
	completionCb     OnboardingCompletionCallback
	stateChangeCb    OnboardingCompletionCallback
	logger           *logrus.Entry
	tutorialSystem   *EbitenTutorialSystem
	creationTutorial *CharacterCreationTutorial
}

// NewOnboardingManager creates a new onboarding manager.
func NewOnboardingManager(logger *logrus.Entry) *OnboardingManager {
	return &OnboardingManager{
		currentState: StateCharacterCreation,
		enabled:      true,
		logger:       logger,
	}
}

// SetTutorialSystem connects the in-game tutorial system.
func (om *OnboardingManager) SetTutorialSystem(ts *EbitenTutorialSystem) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.tutorialSystem = ts
}

// SetCreationTutorial connects the character creation tutorial.
func (om *OnboardingManager) SetCreationTutorial(cct *CharacterCreationTutorial) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.creationTutorial = cct
}

// SetCompletionCallback sets the callback called when onboarding fully completes.
func (om *OnboardingManager) SetCompletionCallback(cb OnboardingCompletionCallback) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.completionCb = cb
}

// SetStateChangeCallback sets the callback called on any state transition.
func (om *OnboardingManager) SetStateChangeCallback(cb OnboardingCompletionCallback) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.stateChangeCb = cb
}

// GetState returns the current onboarding state.
func (om *OnboardingManager) GetState() OnboardingState {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.currentState
}

// IsEnabled returns whether onboarding is active.
func (om *OnboardingManager) IsEnabled() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.enabled
}

// IsComplete returns whether onboarding has finished.
func (om *OnboardingManager) IsComplete() bool {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.currentState == StateComplete || om.skipped
}

// SetEnabled enables or disables the entire onboarding flow.
// When disabled, all tutorial layers are disabled.
func (om *OnboardingManager) SetEnabled(enabled bool) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.enabled = enabled

	if !enabled {
		// Disable all connected tutorial systems
		if om.creationTutorial != nil {
			om.creationTutorial.SetEnabled(false)
		}
		if om.tutorialSystem != nil {
			om.tutorialSystem.Enabled = false
			om.tutorialSystem.ShowUI = false
		}
	}

	if om.logger != nil {
		om.logger.WithField("enabled", enabled).Debug("onboarding enabled state changed")
	}
}

// SetPlayerClass stores the selected class for class-aware tutorial content.
// Propagates the class to the in-game tutorial system for class-specific hints.
func (om *OnboardingManager) SetPlayerClass(class CharacterClass) {
	om.mu.Lock()
	defer om.mu.Unlock()
	om.playerClass = class

	// Propagate class to tutorial system for class-aware content (Phase 3.4)
	if om.tutorialSystem != nil {
		om.tutorialSystem.SetPlayerClass(class)
	}

	if om.logger != nil {
		om.logger.WithField("class", class.String()).Debug("player class set for onboarding")
	}
}

// GetPlayerClass returns the stored player class.
func (om *OnboardingManager) GetPlayerClass() CharacterClass {
	om.mu.RLock()
	defer om.mu.RUnlock()
	return om.playerClass
}

// SkipAll skips the entire onboarding flow.
func (om *OnboardingManager) SkipAll() {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.skipped = true
	om.currentState = StateComplete

	if om.creationTutorial != nil {
		om.creationTutorial.SkipTutorial()
	}
	if om.tutorialSystem != nil {
		om.tutorialSystem.DisableTutorial()
	}

	if om.logger != nil {
		om.logger.Debug("onboarding skipped entirely")
	}
}

// TransitionToInGameTutorial moves from character creation to in-game tutorial.
// This should be called when character creation completes.
func (om *OnboardingManager) TransitionToInGameTutorial() {
	om.mu.Lock()

	if !om.enabled || om.skipped {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateInGameTutorial

	// Activate the in-game tutorial system
	if om.tutorialSystem != nil {
		om.tutorialSystem.Enabled = true
		om.tutorialSystem.ShowUI = true
	}

	// Capture callbacks before unlocking
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call state change callback outside of lock
	if stateChangeCb != nil {
		stateChangeCb(om.currentState)
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"previous_state": previousState.String(),
			"current_state":  om.currentState.String(),
		}).Debug("transitioned to in-game tutorial")
	}
}

// TransitionToContextHelp moves from in-game tutorial to context-sensitive help.
// This should be called when the in-game tutorial completes.
func (om *OnboardingManager) TransitionToContextHelp() {
	om.mu.Lock()

	if !om.enabled || om.skipped {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateContextHelp

	// Capture callbacks before unlocking
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call state change callback outside of lock
	if stateChangeCb != nil {
		stateChangeCb(om.currentState)
	}

	if logger != nil {
		logger.WithFields(logrus.Fields{
			"previous_state": previousState.String(),
			"current_state":  om.currentState.String(),
		}).Debug("transitioned to context-sensitive help")
	}
}

// Complete marks the onboarding flow as finished.
// This should be called when context-sensitive help has been viewed
// or after a timeout period.
func (om *OnboardingManager) Complete() {
	om.mu.Lock()

	if om.currentState == StateComplete {
		om.mu.Unlock()
		return
	}

	previousState := om.currentState
	om.currentState = StateComplete

	// Capture callbacks before unlocking
	completionCb := om.completionCb
	stateChangeCb := om.stateChangeCb
	logger := om.logger

	om.mu.Unlock()

	// Call callbacks outside of lock
	if stateChangeCb != nil {
		stateChangeCb(StateComplete)
	}
	if completionCb != nil {
		completionCb(StateComplete)
	}

	if logger != nil {
		logger.WithField("previous_state", previousState.String()).Debug("onboarding completed")
	}
}

// ExportState serializes the onboarding state for save/load.
func (om *OnboardingManager) ExportState() OnboardingStateData {
	om.mu.RLock()
	defer om.mu.RUnlock()

	return OnboardingStateData{
		CurrentState: int(om.currentState),
		Enabled:      om.enabled,
		Skipped:      om.skipped,
		PlayerClass:  int(om.playerClass),
	}
}

// ImportState restores onboarding state from saved data.
func (om *OnboardingManager) ImportState(data OnboardingStateData) {
	om.mu.Lock()
	defer om.mu.Unlock()

	om.currentState = OnboardingState(data.CurrentState)
	om.enabled = data.Enabled
	om.skipped = data.Skipped
	om.playerClass = CharacterClass(data.PlayerClass)

	// Update connected systems based on loaded state
	if om.currentState >= StateInGameTutorial && om.tutorialSystem != nil {
		om.tutorialSystem.Enabled = om.enabled && om.currentState == StateInGameTutorial
		om.tutorialSystem.ShowUI = om.tutorialSystem.Enabled
	}
	if om.currentState >= StateCharacterCreation && om.creationTutorial != nil {
		// Creation tutorial should be disabled if we're past that phase
		if om.currentState > StateCharacterCreation {
			om.creationTutorial.SetEnabled(false)
		}
	}

	if om.logger != nil {
		om.logger.WithFields(logrus.Fields{
			"state":   om.currentState.String(),
			"enabled": om.enabled,
			"skipped": om.skipped,
		}).Debug("onboarding state imported")
	}
}

// OnboardingStateData holds serializable onboarding state.
type OnboardingStateData struct {
	CurrentState int  `json:"current_state"`
	Enabled      bool `json:"enabled"`
	Skipped      bool `json:"skipped"`
	PlayerClass  int  `json:"player_class"`
}
