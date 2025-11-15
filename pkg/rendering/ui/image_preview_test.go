package ui

import (
	"testing"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestNewImagePreviewUI(t *testing.T) {
	ui := NewImagePreviewUI(100, 100, 500, 400)

	if ui.X != 100 || ui.Y != 100 {
		t.Errorf("Expected position (100,100), got (%d,%d)", ui.X, ui.Y)
	}
	if ui.Width != 500 || ui.Height != 400 {
		t.Errorf("Expected size (500,400), got (%d,%d)", ui.Width, ui.Height)
	}
	if ui.Active {
		t.Error("Expected preview to start inactive")
	}
	if ui.MaxImageSize != 512 {
		t.Errorf("Expected max image size 512, got %d", ui.MaxImageSize)
	}
	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to default to true")
	}
}

func TestShowThumbnail(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)
	timestamp := time.Now()
	expiry := time.Now().Add(10 * time.Minute)

	ui.ShowThumbnail("Alice", timestamp, thumbnail, 1024, 768, 512000, expiry)

	if !ui.Active {
		t.Error("Expected preview to be active after ShowThumbnail")
	}
	if ui.SenderName != "Alice" {
		t.Errorf("Expected sender 'Alice', got '%s'", ui.SenderName)
	}
	if ui.ThumbnailImage != thumbnail {
		t.Error("Expected thumbnail image to be set")
	}
	if ui.ImageWidth != 1024 || ui.ImageHeight != 768 {
		t.Errorf("Expected image size 1024x768, got %dx%d", ui.ImageWidth, ui.ImageHeight)
	}
	if ui.ImageSize != 512000 {
		t.Errorf("Expected image size 512000 bytes, got %d", ui.ImageSize)
	}
	if ui.ShowFullImage {
		t.Error("Expected ShowFullImage to be false initially")
	}
	if ui.DownloadProgress != 0.0 {
		t.Errorf("Expected download progress 0.0, got %f", ui.DownloadProgress)
	}
}

func TestShowFullImage(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)
	fullImage := ebiten.NewImage(1024, 768)
	expiry := time.Now().Add(10 * time.Minute)

	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, expiry)
	ui.ShowFullImage(fullImage)

	if ui.FullImage != fullImage {
		t.Error("Expected full image to be set")
	}
	if !ui.ShowFullImage {
		t.Error("Expected ShowFullImage to be true")
	}
	if ui.DownloadProgress != 1.0 {
		t.Errorf("Expected download progress 1.0, got %f", ui.DownloadProgress)
	}
}

func TestHide(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)
	expiry := time.Now().Add(10 * time.Minute)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, expiry)

	if !ui.Active {
		t.Error("Expected preview to be active before Hide")
	}

	ui.Hide()

	if ui.Active {
		t.Error("Expected preview to be inactive after Hide")
	}
	if ui.ThumbnailImage != nil {
		t.Error("Expected thumbnail image to be cleared")
	}
	if ui.FullImage != nil {
		t.Error("Expected full image to be cleared")
	}
}

func TestToggleSelection(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to start as true")
	}

	ui.ToggleSelection()
	if ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to be false after first toggle")
	}

	ui.ToggleSelection()
	if !ui.ConfirmSelected {
		t.Error("Expected ConfirmSelected to be true after second toggle")
	}
}

func TestGetSelectedAction(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	action := ui.GetSelectedAction()
	if action != "download" {
		t.Errorf("Expected 'download', got '%s'", action)
	}

	ui.ToggleSelection()
	action = ui.GetSelectedAction()
	if action != "decline" {
		t.Errorf("Expected 'decline', got '%s'", action)
	}
}

func TestUpdateDownloadProgress(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	tests := []struct {
		input    float32
		expected float32
	}{
		{0.0, 0.0},
		{0.5, 0.5},
		{1.0, 1.0},
		{-0.1, 0.0}, // Clamped to 0
		{1.5, 1.0},  // Clamped to 1
	}

	for _, tt := range tests {
		ui.UpdateDownloadProgress(tt.input)
		if ui.DownloadProgress != tt.expected {
			t.Errorf("Expected progress %f, got %f", tt.expected, ui.DownloadProgress)
		}
	}
}

func TestIsExpired(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	// Not expired (future expiry)
	thumbnail := ebiten.NewImage(128, 128)
	futureExpiry := time.Now().Add(1 * time.Hour)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, futureExpiry)

	if ui.IsExpired() {
		t.Error("Expected image not to be expired")
	}

	// Expired (past expiry)
	pastExpiry := time.Now().Add(-1 * time.Second)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, pastExpiry)

	if !ui.IsExpired() {
		t.Error("Expected image to be expired")
	}
}

func TestGetTimeRemaining(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)

	// Future expiry
	futureExpiry := time.Now().Add(5 * time.Minute)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, futureExpiry)

	remaining := ui.GetTimeRemaining()
	if remaining < 4*time.Minute || remaining > 5*time.Minute {
		t.Errorf("Expected remaining time around 5 minutes, got %v", remaining)
	}

	// Past expiry
	pastExpiry := time.Now().Add(-1 * time.Minute)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, pastExpiry)

	remaining = ui.GetTimeRemaining()
	if remaining != 0 {
		t.Errorf("Expected remaining time 0 for expired image, got %v", remaining)
	}
}

func TestUpdateAutoHideExpired(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)
	pastExpiry := time.Now().Add(-1 * time.Second)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, pastExpiry)

	if !ui.Active {
		t.Error("Expected preview to be active before Update")
	}

	ui.Update(0.016)

	if ui.Active {
		t.Error("Expected preview to auto-hide after expiry")
	}
}

func TestUpdateNoAutoHideFullImage(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	thumbnail := ebiten.NewImage(128, 128)
	fullImage := ebiten.NewImage(1024, 768)
	pastExpiry := time.Now().Add(-1 * time.Second)

	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, pastExpiry)
	ui.ShowFullImage(fullImage)

	ui.Update(0.016)

	// Should not auto-hide when showing full image
	if !ui.Active {
		t.Error("Expected preview to remain active when showing full image, even if expired")
	}
}

func TestImageSizeFormatting(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	tests := []struct {
		sizeBytes int
		// We'll test that it doesn't panic; actual formatting is visual
	}{
		{100},     // 0.1 KB
		{1024},    // 1 KB
		{102400},  // 100 KB
		{512000},  // 500 KB
		{1048576}, // 1 MB (outside limit, but test it anyway)
	}

	thumbnail := ebiten.NewImage(128, 128)
	expiry := time.Now().Add(10 * time.Minute)

	for _, tt := range tests {
		// Just verify it doesn't panic
		ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, tt.sizeBytes, expiry)
		if ui.ImageSize != tt.sizeBytes {
			t.Errorf("Expected image size %d, got %d", tt.sizeBytes, ui.ImageSize)
		}
	}
}

func TestImageDimensions(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	tests := []struct {
		width  int
		height int
	}{
		{128, 128},
		{800, 600},
		{1920, 1080},
		{2048, 2048}, // Max allowed
	}

	thumbnail := ebiten.NewImage(128, 128)
	expiry := time.Now().Add(10 * time.Minute)

	for _, tt := range tests {
		ui.ShowThumbnail("Alice", time.Now(), thumbnail, tt.width, tt.height, 512000, expiry)
		if ui.ImageWidth != tt.width || ui.ImageHeight != tt.height {
			t.Errorf("Expected dimensions %dx%d, got %dx%d", tt.width, tt.height, ui.ImageWidth, ui.ImageHeight)
		}
	}
}

// Test with nil images (edge case)
func TestNilImages(t *testing.T) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	expiry := time.Now().Add(10 * time.Minute)

	// Show thumbnail with nil image
	ui.ShowThumbnail("Alice", time.Now(), nil, 1024, 768, 512000, expiry)

	if ui.ThumbnailImage != nil {
		t.Error("Expected thumbnail image to be nil")
	}

	// Show full image with nil
	ui.ShowFullImage(nil)

	if ui.FullImage != nil {
		t.Error("Expected full image to be nil")
	}
}

// Benchmarks

func BenchmarkShowThumbnail(b *testing.B) {
	ui := NewImagePreviewUI(0, 0, 500, 400)
	thumbnail := ebiten.NewImage(128, 128)
	timestamp := time.Now()
	expiry := time.Now().Add(10 * time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.ShowThumbnail("Alice", timestamp, thumbnail, 1024, 768, 512000, expiry)
	}
}

func BenchmarkShowFullImage(b *testing.B) {
	ui := NewImagePreviewUI(0, 0, 500, 400)
	fullImage := ebiten.NewImage(1024, 768)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.ShowFullImage(fullImage)
	}
}

func BenchmarkUpdate(b *testing.B) {
	ui := NewImagePreviewUI(0, 0, 500, 400)
	thumbnail := ebiten.NewImage(128, 128)
	expiry := time.Now().Add(10 * time.Minute)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, expiry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.Update(0.016)
	}
}

func BenchmarkUpdateDownloadProgress(b *testing.B) {
	ui := NewImagePreviewUI(0, 0, 500, 400)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.UpdateDownloadProgress(0.5)
	}
}

func BenchmarkIsExpired(b *testing.B) {
	ui := NewImagePreviewUI(0, 0, 500, 400)
	thumbnail := ebiten.NewImage(128, 128)
	expiry := time.Now().Add(10 * time.Minute)
	ui.ShowThumbnail("Alice", time.Now(), thumbnail, 1024, 768, 512000, expiry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ui.IsExpired()
	}
}
