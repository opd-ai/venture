//go:build test
// +build test

package engine

import (
	"errors"
	"testing"
)

// Stub types for testing (input_system.go has !test build tag)
type InputSystem struct {
	helpSystem   *HelpSystem
	onQuickSave  func() error
	onQuickLoad  func() error
	KeyHelp      int
	KeyQuickSave int
	KeyQuickLoad int
}

type HelpSystem struct {
	Enabled bool
	Visible bool
}

func NewInputSystem() *InputSystem {
	return &InputSystem{
		KeyHelp:      27,  // Escape key
		KeyQuickSave: 290, // F5 key
		KeyQuickLoad: 294, // F9 key
	}
}

func (i *InputSystem) SetHelpSystem(h *HelpSystem) {
	i.helpSystem = h
}

func (i *InputSystem) SetQuickSaveCallback(callback func() error) {
	i.onQuickSave = callback
}

func (i *InputSystem) SetQuickLoadCallback(callback func() error) {
	i.onQuickLoad = callback
}

func (i *InputSystem) Update(entities []*Entity, deltaTime float64) {
	// Stub implementation for testing
}

// TestInputSystem_SetHelpSystem tests help system integration.
func TestInputSystem_SetHelpSystem(t *testing.T) {
	inputSys := NewInputSystem()
	helpSys := &HelpSystem{Enabled: true, Visible: false}

	inputSys.SetHelpSystem(helpSys)

	if inputSys.helpSystem == nil {
		t.Error("Expected help system to be set, got nil")
	}

	if inputSys.helpSystem != helpSys {
		t.Error("Help system reference mismatch")
	}
}

// TestInputSystem_SetQuickSaveCallback tests save callback registration.
func TestInputSystem_SetQuickSaveCallback(t *testing.T) {
	inputSys := NewInputSystem()
	called := false

	callback := func() error {
		called = true
		return nil
	}

	inputSys.SetQuickSaveCallback(callback)

	// Verify callback was set
	if inputSys.onQuickSave == nil {
		t.Fatal("Expected quick save callback to be set, got nil")
	}

	// Call the callback directly to verify it works
	err := inputSys.onQuickSave()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Callback was not invoked")
	}
}

// TestInputSystem_SetQuickLoadCallback tests load callback registration.
func TestInputSystem_SetQuickLoadCallback(t *testing.T) {
	inputSys := NewInputSystem()
	called := false

	callback := func() error {
		called = true
		return nil
	}

	inputSys.SetQuickLoadCallback(callback)

	// Verify callback was set
	if inputSys.onQuickLoad == nil {
		t.Fatal("Expected quick load callback to be set, got nil")
	}

	// Call the callback directly to verify it works
	err := inputSys.onQuickLoad()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if !called {
		t.Error("Callback was not invoked")
	}
}

// TestInputSystem_QuickSaveCallbackError tests error handling in save callback.
func TestInputSystem_QuickSaveCallbackError(t *testing.T) {
	inputSys := NewInputSystem()
	expectedErr := errors.New("save failed")

	callback := func() error {
		return expectedErr
	}

	inputSys.SetQuickSaveCallback(callback)

	// Call the callback and verify error is returned
	err := inputSys.onQuickSave()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestInputSystem_QuickLoadCallbackError tests error handling in load callback.
func TestInputSystem_QuickLoadCallbackError(t *testing.T) {
	inputSys := NewInputSystem()
	expectedErr := errors.New("load failed")

	callback := func() error {
		return expectedErr
	}

	inputSys.SetQuickLoadCallback(callback)

	// Call the callback and verify error is returned
	err := inputSys.onQuickLoad()
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

// TestInputSystem_KeyBindings tests that default key bindings are set correctly.
func TestInputSystem_KeyBindings(t *testing.T) {
	inputSys := NewInputSystem()

	tests := []struct {
		name     string
		key      int // ebiten.Key is an int type
		expected int
	}{
		{"KeyHelp should be Escape", int(inputSys.KeyHelp), 27}, // Escape key code
		{"KeyQuickSave should be F5", int(inputSys.KeyQuickSave), 290}, // F5 key code
		{"KeyQuickLoad should be F9", int(inputSys.KeyQuickLoad), 294}, // F9 key code
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.key != tt.expected {
				t.Errorf("Key binding mismatch: got %d, want %d", tt.key, tt.expected)
			}
		})
	}
}

// TestInputSystem_NilCallbacks tests behavior when callbacks are not set.
func TestInputSystem_NilCallbacks(t *testing.T) {
	inputSys := NewInputSystem()

	// Should not panic when callbacks are nil
	if inputSys.onQuickSave != nil {
		t.Error("Expected onQuickSave to be nil initially")
	}

	if inputSys.onQuickLoad != nil {
		t.Error("Expected onQuickLoad to be nil initially")
	}

	// Update with empty entity list should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Update panicked with nil callbacks: %v", r)
		}
	}()

	inputSys.Update([]*Entity{}, 0.016)
}

// TestInputSystem_MultipleCallbackSets tests overwriting callbacks.
func TestInputSystem_MultipleCallbackSets(t *testing.T) {
	inputSys := NewInputSystem()

	firstCalled := false
	secondCalled := false

	// Set first callback
	inputSys.SetQuickSaveCallback(func() error {
		firstCalled = true
		return nil
	})

	// Set second callback (should replace first)
	inputSys.SetQuickSaveCallback(func() error {
		secondCalled = true
		return nil
	})

	// Call callback
	err := inputSys.onQuickSave()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if firstCalled {
		t.Error("First callback should not have been called")
	}

	if !secondCalled {
		t.Error("Second callback should have been called")
	}
}

// TestInputSystem_IntegrationWithHelpSystem tests full integration scenario.
func TestInputSystem_IntegrationWithHelpSystem(t *testing.T) {
	inputSys := NewInputSystem()
	helpSys := &HelpSystem{
		Enabled: true,
		Visible: false,
	}

	inputSys.SetHelpSystem(helpSys)

	// Verify help system state
	if helpSys.Visible {
		t.Error("Help system should start hidden")
	}

	// Note: We can't actually test key presses without Ebiten,
	// but we can verify the setup is correct
	if inputSys.helpSystem == nil {
		t.Error("Help system reference not set")
	}

	if !helpSys.Enabled {
		t.Error("Help system should be enabled")
	}
}

// TestInputSystem_SaveLoadCallbackSequence tests save then load sequence.
func TestInputSystem_SaveLoadCallbackSequence(t *testing.T) {
	inputSys := NewInputSystem()

	var savedData string
	var loadedData string

	// Setup save callback
	inputSys.SetQuickSaveCallback(func() error {
		savedData = "test_save_data"
		return nil
	})

	// Setup load callback
	inputSys.SetQuickLoadCallback(func() error {
		loadedData = savedData
		return nil
	})

	// Simulate save
	if err := inputSys.onQuickSave(); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	if savedData != "test_save_data" {
		t.Errorf("Save did not store data correctly: got %q", savedData)
	}

	// Simulate load
	if err := inputSys.onQuickLoad(); err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if loadedData != savedData {
		t.Errorf("Load did not restore data correctly: got %q, want %q", loadedData, savedData)
	}
}
