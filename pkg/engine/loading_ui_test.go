package engine

import (
	"testing"
)

func TestNewLoadingUI(t *testing.T) {
	ui := NewLoadingUI(800, 600)
	if ui == nil {
		t.Fatal("NewLoadingUI() returned nil")
	}

	if ui.screenWidth != 800 {
		t.Errorf("screenWidth = %d, want 800", ui.screenWidth)
	}

	if ui.screenHeight != 600 {
		t.Errorf("screenHeight = %d, want 600", ui.screenHeight)
	}

	if ui.progress != 0.0 {
		t.Errorf("initial progress = %f, want 0.0", ui.progress)
	}

	if ui.message == "" {
		t.Error("initial message should not be empty")
	}
}

func TestLoadingUI_SetProgress(t *testing.T) {
	ui := NewLoadingUI(800, 600)

	tests := []struct {
		name  string
		input float64
		want  float64
	}{
		{"zero", 0.0, 0.0},
		{"half", 0.5, 0.5},
		{"full", 1.0, 1.0},
		{"negative clamped", -0.5, 0.0},
		{"over 1.0 clamped", 1.5, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ui.SetProgress(tt.input)
			if ui.progress != tt.want {
				t.Errorf("SetProgress(%f) = %f, want %f", tt.input, ui.progress, tt.want)
			}
		})
	}
}

func TestLoadingUI_SetMessage(t *testing.T) {
	ui := NewLoadingUI(800, 600)

	testMessage := "Generating dungeons..."
	ui.SetMessage(testMessage)

	if ui.message != testMessage {
		t.Errorf("SetMessage() = %s, want %s", ui.message, testMessage)
	}
}

func TestLoadingUI_Draw(t *testing.T) {
	t.Skip("Skipping Draw test - requires display")
}
