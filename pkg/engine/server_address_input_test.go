package engine

import (
	"testing"
)

// TestNewServerAddressInput tests input field initialization.
func TestNewServerAddressInput(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	if input == nil {
		t.Fatal("expected input to be created")
	}

	if input.screenWidth != 1280 {
		t.Errorf("expected screenWidth 1280, got %d", input.screenWidth)
	}

	if input.screenHeight != 720 {
		t.Errorf("expected screenHeight 720, got %d", input.screenHeight)
	}

	if input.address != "localhost:8080" {
		t.Errorf("expected default address 'localhost:8080', got '%s'", input.address)
	}

	if input.cursorPos != len("localhost:8080") {
		t.Errorf("expected cursorPos %d, got %d", len("localhost:8080"), input.cursorPos)
	}

	if input.isVisible {
		t.Error("expected input to be initially invisible")
	}

	if input.maxLength != 50 {
		t.Errorf("expected maxLength 50, got %d", input.maxLength)
	}
}

// TestServerAddressInput_ShowHide tests visibility toggling.
func TestServerAddressInput_ShowHide(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	// Initially hidden
	if input.IsVisible() {
		t.Error("expected input to be initially invisible")
	}

	// Show
	input.Show()
	if !input.IsVisible() {
		t.Error("expected input to be visible after Show()")
	}

	// Verify Show() resets to default
	input.SetAddress("custom:1234")
	input.Show()
	if input.address != "localhost:8080" {
		t.Errorf("expected Show() to reset to default address, got '%s'", input.address)
	}

	// Hide
	input.Hide()
	if input.IsVisible() {
		t.Error("expected input to be invisible after Hide()")
	}
}

// TestServerAddressInput_SetCallbacks tests callback registration.
func TestServerAddressInput_SetCallbacks(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	var connectCalled, cancelCalled bool
	var capturedAddress string

	input.SetConnectCallback(func(addr string) {
		connectCalled = true
		capturedAddress = addr
	})
	input.SetCancelCallback(func() {
		cancelCalled = true
	})

	if input.onConnect == nil {
		t.Error("expected onConnect callback to be set")
	}
	if input.onCancel == nil {
		t.Error("expected onCancel callback to be set")
	}

	// Test callbacks work
	input.onConnect("test:1234")
	if !connectCalled {
		t.Error("expected connect callback to be called")
	}
	if capturedAddress != "test:1234" {
		t.Errorf("expected captured address 'test:1234', got '%s'", capturedAddress)
	}

	input.onCancel()
	if !cancelCalled {
		t.Error("expected cancel callback to be called")
	}
}

// TestServerAddressInput_GetAddress tests address retrieval.
func TestServerAddressInput_GetAddress(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	// Default address
	got := input.GetAddress()
	if got != "localhost:8080" {
		t.Errorf("expected 'localhost:8080', got '%s'", got)
	}

	// Custom address
	input.SetAddress("example.com:9999")
	got = input.GetAddress()
	if got != "example.com:9999" {
		t.Errorf("expected 'example.com:9999', got '%s'", got)
	}
}

// TestServerAddressInput_SetAddress tests address setting with length limits.
func TestServerAddressInput_SetAddress(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	tests := []struct {
		name            string
		address         string
		expectedAddress string
		expectedCursor  int
	}{
		{"short address", "host:80", "host:80", 7},
		{"long address", "very-long-hostname.example.com:12345", "very-long-hostname.example.com:12345", 36}, // Fixed: actual length is 36
		{"max length address", "12345678901234567890123456789012345678901234567890", "12345678901234567890123456789012345678901234567890", 50},
		{"too long address", "123456789012345678901234567890123456789012345678901234567890", "localhost:8080", 14}, // Should not change
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input.SetAddress("localhost:8080") // Reset to default
			input.SetAddress(tt.address)

			if input.address != tt.expectedAddress {
				t.Errorf("expected address '%s', got '%s'", tt.expectedAddress, input.address)
			}

			if input.cursorPos != tt.expectedCursor {
				t.Errorf("expected cursor position %d, got %d", tt.expectedCursor, input.cursorPos)
			}
		})
	}
}

// TestServerAddressInput_Update_NotVisible tests that update does nothing when hidden.
func TestServerAddressInput_Update_NotVisible(t *testing.T) {
	input := NewServerAddressInput(1280, 720)
	input.Hide()

	initialAddress := input.address
	initialCursor := input.cursorPos

	// Update should do nothing when hidden
	input.Update()

	if input.address != initialAddress {
		t.Error("expected address to not change when input is hidden")
	}

	if input.cursorPos != initialCursor {
		t.Error("expected cursor position to not change when input is hidden")
	}
}

// TestServerAddressInput_CursorBlink tests cursor blink timer.
func TestServerAddressInput_CursorBlink(t *testing.T) {
	input := NewServerAddressInput(1280, 720)
	input.Show()

	// Initially cursor should be visible
	if !input.showCursor {
		t.Error("expected cursor to be initially visible")
	}

	// After 15 frames, cursor should toggle
	for i := 0; i < 15; i++ {
		input.blinkTimer++
	}
	if input.blinkTimer != 15 {
		t.Errorf("expected blinkTimer to be 15, got %d", input.blinkTimer)
	}

	// Simulate update logic
	if input.blinkTimer >= 15 {
		input.blinkTimer = 0
		input.showCursor = !input.showCursor
	}

	if input.showCursor {
		t.Error("expected cursor to be hidden after blink")
	}
}

// TestServerAddressInput_ResetCursorBlink tests cursor blink reset.
func TestServerAddressInput_ResetCursorBlink(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	// Set up state where cursor is hidden and timer is mid-blink
	input.blinkTimer = 10
	input.showCursor = false

	// Reset cursor blink
	input.resetCursorBlink()

	if input.blinkTimer != 0 {
		t.Errorf("expected blinkTimer to be reset to 0, got %d", input.blinkTimer)
	}

	if !input.showCursor {
		t.Error("expected cursor to be shown after reset")
	}
}

// TestServerAddressInput_Draw tests rendering (basic non-nil check).
func TestServerAddressInput_Draw(t *testing.T) {
	input := NewServerAddressInput(1280, 720)
	input.Show()

	// Draw with nil screen should not panic
	input.Draw(nil)

	// Note: Cannot create ebiten.Image in tests without Ebiten runtime
	// This test verifies the nil check works correctly
}

// TestServerAddressInput_IntegrationWithGame tests input integration with game state.
func TestServerAddressInput_IntegrationWithGame(t *testing.T) {
	game := NewEbitenGame(1280, 720)

	// Verify input is initialized
	if game.ServerAddressInput == nil {
		t.Fatal("expected ServerAddressInput to be initialized")
	}

	// Verify input is initially invisible
	if game.ServerAddressInput.IsVisible() {
		t.Error("expected input to be initially invisible")
	}

	// Verify callbacks are wired
	if game.ServerAddressInput.onConnect == nil {
		t.Error("expected onConnect callback to be wired")
	}
	if game.ServerAddressInput.onCancel == nil {
		t.Error("expected onCancel callback to be wired")
	}

	// Test state transitions: Main Menu → Multi-Player → Join
	game.handleMainMenuSelection(MainMenuOptionMultiPlayer)
	game.handleMultiplayerMenuJoin()

	if game.StateManager.CurrentState() != AppStateServerAddressInput {
		t.Errorf("expected ServerAddressInput state, got %s", game.StateManager.CurrentState().String())
	}

	// Input should be visible
	if !game.ServerAddressInput.IsVisible() {
		t.Error("expected input to be visible after transition")
	}

	// Test cancel
	game.handleServerAddressCancel()
	if game.StateManager.CurrentState() != AppStateMultiPlayerMenu {
		t.Errorf("expected MultiPlayerMenu state after cancel, got %s", game.StateManager.CurrentState().String())
	}
}

// TestServerAddressInput_DefaultValue tests default address is sensible.
func TestServerAddressInput_DefaultValue(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	// Default should be localhost:8080 (standard dev server address)
	if input.address != "localhost:8080" {
		t.Errorf("expected default 'localhost:8080', got '%s'", input.address)
	}

	// Cursor should be at end of default text
	if input.cursorPos != len("localhost:8080") {
		t.Errorf("expected cursor at position %d, got %d", len("localhost:8080"), input.cursorPos)
	}
}

// TestServerAddressInput_MaxLength tests address length validation.
func TestServerAddressInput_MaxLength(t *testing.T) {
	input := NewServerAddressInput(1280, 720)

	// Max length should be 50 characters
	if input.maxLength != 50 {
		t.Errorf("expected maxLength 50, got %d", input.maxLength)
	}

	// SetAddress should reject addresses longer than maxLength
	longAddress := "12345678901234567890123456789012345678901234567890X" // 51 chars
	input.SetAddress(longAddress)

	// Should not have changed from default
	if input.address != "localhost:8080" {
		t.Errorf("expected SetAddress to reject too-long address, got '%s'", input.address)
	}

	// Exactly maxLength should work
	exactAddress := "12345678901234567890123456789012345678901234567890" // 50 chars
	input.SetAddress(exactAddress)

	if input.address != exactAddress {
		t.Errorf("expected SetAddress to accept max-length address, got '%s'", input.address)
	}
}

// BenchmarkServerAddressInput_Update benchmarks update performance.
func BenchmarkServerAddressInput_Update(b *testing.B) {
	input := NewServerAddressInput(1280, 720)
	input.Show()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.Update()
	}
}

// BenchmarkServerAddressInput_ResetCursorBlink benchmarks cursor blink reset performance.
func BenchmarkServerAddressInput_ResetCursorBlink(b *testing.B) {
	input := NewServerAddressInput(1280, 720)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input.resetCursorBlink()
	}
}
